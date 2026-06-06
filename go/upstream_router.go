// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil" // Note: AX-6 — reverse-proxy mechanics are structural; no core primitive.
	"net/url"           // Note: AX-6 — url.Parse is structural for the Rewrite placeholder target.
	"strconv"
	"time"

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
)

const (
	defaultUpstreamRouterPath = "/v1/chat/completions"
	defaultUpstreamCooldown   = 10 * time.Second

	errCodeInvalidRequest      = "invalid_request"
	errCodeInvalidRequestBody  = "invalid_request_body"
	errCodeRoutingRejected     = "routing_rejected"
	errCodeNoUpstream          = "no_upstream_for_key"
	errCodeRequestTooLarge     = "request_too_large"
	errCodeUpstreamUnavailable = "upstream_unavailable"
	errCodeInvalidUpstreamResp = "invalid_upstream_response"
)

type ctxKey int

const (
	poolCtxKey ctxKey = iota
	keyCtxKey
	ginCtxKey
)

// Selector resolves the routing key from the request. body holds the (bounded)
// request body, already read by the handler; it may be empty for bodyless requests.
type Selector func(c *gin.Context, body []byte) (key string, err error)

// RouteFunc inspects the payload after the selector and may override the key or
// reject the request. Returning the same key is a no-op; a non-nil error aborts.
type RouteFunc func(c *gin.Context, key string, body []byte) (newKey string, err error)

// UpstreamRouterOption configures a router built by WithUpstreamRouter.
type UpstreamRouterOption func(*upstreamRouterConfig)

type upstreamRouterConfig struct {
	registry    *UpstreamRegistry
	selector    Selector
	hook        RouteFunc
	paths       []string
	inRaw       []any
	outRaw      []any
	in          []compiledTransformer
	out         []compiledTransformer
	maxAttempts int
	cooldown    time.Duration
	failover    map[int]bool
	transport   http.RoundTripper
}

// routerError carries an HTTP status + envelope code from the transport or
// ModifyResponse to the ReverseProxy ErrorHandler.
type routerError struct {
	status  int
	code    string
	message string
	cause   error
}

func (e *routerError) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e *routerError) Unwrap() error { return e.cause }

// WithSelector overrides the routing-key selector. Default: defaultModelSelector.
func WithSelector(fn Selector) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.selector = fn }
}

// WithRouteHook installs a decision hook to inspect the payload and override/reject.
func WithRouteHook(fn RouteFunc) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.hook = fn }
}

// WithRouterPaths sets the mounted paths (default ["/v1/chat/completions"]).
// Each path forwards its own path + query to the chosen upstream.
func WithRouterPaths(paths ...string) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.paths = paths }
}

// WithUpstreamTransformerIn adds request-body transformers (reuses the existing
// TransformerIn machinery; FieldRenamer etc. work). Operates on the raw body.
func WithUpstreamTransformerIn(t ...any) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.inRaw = append(cfg.inRaw, t...) }
}

// WithUpstreamTransformerOut adds response-body transformers, applied only to
// buffered (non-streaming) responses, on the raw upstream body.
func WithUpstreamTransformerOut(t ...any) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.outRaw = append(cfg.outRaw, t...) }
}

// WithFailover sets the max upstream attempts (default len(pool), each tried once)
// and the cooldown applied to a failed upstream (default 10s).
func WithFailover(maxAttempts int, cooldown time.Duration) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) {
		cfg.maxAttempts = maxAttempts
		if cooldown > 0 {
			cfg.cooldown = cooldown
		}
	}
}

// WithFailoverStatuses overrides which response statuses trigger failover
// (default: all >= 500). Pass e.g. 429 to also fail over on rate-limit responses.
// Passing zero statuses disables status-based failover (transport errors still
// fail over).
func WithFailoverStatuses(statuses ...int) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) {
		cfg.failover = map[int]bool{}
		for _, s := range statuses {
			cfg.failover[s] = true
		}
	}
}

// WithUpstreamTransport sets the base RoundTripper used for dispatch (custom TLS,
// timeouts). Default: a clone of http.DefaultTransport.
func WithUpstreamTransport(rt http.RoundTripper) UpstreamRouterOption {
	return func(cfg *upstreamRouterConfig) { cfg.transport = rt }
}

// defaultFailoverStatuses returns the default failover status set: all >= 500.
func defaultFailoverStatuses() map[int]bool {
	m := map[int]bool{}
	for s := 500; s <= 599; s++ {
		m[s] = true
	}
	return m
}

// defaultModelSelector reads the OpenAI-style "model" field from a JSON body.
func defaultModelSelector(_ *gin.Context, body []byte) (string, error) {
	var probe struct {
		Model string `json:"model"`
	}
	if res := core.JSONUnmarshal(body, &probe); !res.OK {
		return "", core.E("upstream.selector", "request body is not valid JSON", nil)
	}
	if core.Trim(probe.Model) == "" {
		return "", core.E("upstream.selector", "request body has no \"model\" field", nil)
	}
	return probe.Model, nil
}

func poolFromContext(ctx context.Context) ([]Upstream, bool) {
	pool, ok := ctx.Value(poolCtxKey).([]Upstream)
	return pool, ok
}

func keyFromContext(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(keyCtxKey).(string)
	return key, ok
}

// finalise resolves defaults and compiles transformer pipelines. Returns an
// error if a transformer fails to compile.
func (cfg *upstreamRouterConfig) finalise() error {
	if cfg.selector == nil {
		cfg.selector = defaultModelSelector
	}
	if len(cfg.paths) == 0 {
		cfg.paths = []string{defaultUpstreamRouterPath}
	}
	if cfg.cooldown <= 0 {
		cfg.cooldown = defaultUpstreamCooldown
	}
	if cfg.failover == nil {
		cfg.failover = defaultFailoverStatuses()
	}
	if cfg.transport == nil {
		// Clone so the router owns an isolated connection pool rather than
		// mutating/sharing the process-wide http.DefaultTransport.
		cfg.transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	in, err := compileTransformerPipeline(transformerDirectionIn, cfg.inRaw)
	if err != nil {
		return err
	}
	out, err := compileTransformerPipeline(transformerDirectionOut, cfg.outRaw)
	if err != nil {
		return err
	}
	cfg.in, cfg.out = in, out
	return nil
}

// buildProxy constructs the shared ReverseProxy for the router.
func (cfg *upstreamRouterConfig) buildProxy() *httputil.ReverseProxy {
	balancer := newUpstreamBalancer(cfg.cooldown, time.Now)
	transport := &upstreamTransport{
		base:        cfg.transport,
		balancer:    balancer,
		maxAttempts: cfg.maxAttempts,
		failover:    cfg.failover,
	}
	return &httputil.ReverseProxy{
		Transport:     transport,
		FlushInterval: -1, // stream SSE / chunked responses through immediately
		Rewrite: func(pr *httputil.ProxyRequest) {
			// Placeholder target so the pipeline has a valid URL; the transport
			// overrides scheme/host/path per attempt for the selected upstream.
			if pool, ok := poolFromContext(pr.In.Context()); ok && len(pool) > 0 {
				if target, err := url.Parse(pool[0].URL); err == nil {
					pr.Out.URL.Scheme = target.Scheme
					pr.Out.URL.Host = target.Host
				}
			}
			pr.SetXForwarded()
		},
		ModifyResponse: cfg.modifyResponse,
		ErrorHandler:   cfg.errorHandler,
	}
}

func (cfg *upstreamRouterConfig) modifyResponse(resp *http.Response) error {
	if len(cfg.out) == 0 {
		return nil
	}
	if isEventStream(resp.Header.Get("Content-Type")) {
		return nil // streaming: pass through untransformed
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return &routerError{status: http.StatusBadGateway, code: errCodeInvalidUpstreamResp, message: "could not read upstream response", cause: err}
	}
	c, _ := resp.Request.Context().Value(ginCtxKey).(*gin.Context)
	transformed, err := runTransformerPipeline(c, body, cfg.out)
	if err != nil {
		return &routerError{status: http.StatusBadGateway, code: errCodeInvalidUpstreamResp, message: "response transform failed", cause: err}
	}
	resp.Body = io.NopCloser(bytes.NewReader(transformed))
	resp.ContentLength = int64(len(transformed))
	resp.Header.Set("Content-Length", strconv.Itoa(len(transformed)))
	return nil
}

func (cfg *upstreamRouterConfig) errorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	re := &routerError{status: http.StatusBadGateway, code: errCodeUpstreamUnavailable, message: "upstream request failed"}
	var got *routerError
	if core.As(err, &got) {
		re = got
	}
	slog.Warn("upstream router dispatch failed", "code", re.code, "err", err.Error())
	w.Header().Set("Content-Type", "application/json")
	if re.status == http.StatusServiceUnavailable {
		w.Header().Set("Retry-After", strconv.Itoa(int(cfg.cooldown.Seconds())))
	}
	w.WriteHeader(re.status)
	_ = json.NewEncoder(w).Encode(Fail(re.code, re.message))
}

// handler returns the gin.HandlerFunc mounted at each router path.
func (cfg *upstreamRouterConfig) handler(proxy *httputil.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, ok := readUpstreamBody(c)
		if !ok {
			return
		}

		key, err := cfg.selector(c, body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Fail(errCodeInvalidRequest, err.Error()))
			return
		}
		if cfg.hook != nil {
			newKey, herr := cfg.hook(c, key, body)
			if herr != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, Fail(errCodeRoutingRejected, herr.Error()))
				return
			}
			if core.Trim(newKey) != "" {
				key = newKey
			}
		}

		if len(cfg.in) > 0 {
			body, err = runTransformerPipeline(c, body, cfg.in)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, Fail(errCodeInvalidRequestBody, err.Error()))
				return
			}
		}

		pool, ok := cfg.registry.resolve(key)
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, Fail(errCodeNoUpstream, "no upstream registered for key: "+key))
			return
		}

		bound := body // capture for GetBody closure
		c.Request.Body = io.NopCloser(bytes.NewReader(bound))
		c.Request.ContentLength = int64(len(bound))
		c.Request.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(bound)), nil }

		ctx := context.WithValue(c.Request.Context(), poolCtxKey, pool)
		ctx = context.WithValue(ctx, keyCtxKey, key)
		ctx = context.WithValue(ctx, ginCtxKey, c)
		c.Request = c.Request.WithContext(ctx)

		proxy.ServeHTTP(upstreamResponseWriter(c), c.Request)
	}
}

// upstreamResponseWriter unwraps gin's ResponseWriter to the underlying
// http.ResponseWriter, which httputil.ReverseProxy requires for flush/cancel.
func upstreamResponseWriter(c *gin.Context) http.ResponseWriter {
	var w http.ResponseWriter = c.Writer
	if uw, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
		w = uw.Unwrap()
	}
	return w
}

func readUpstreamBody(c *gin.Context) ([]byte, bool) {
	limited := http.MaxBytesReader(c.Writer, c.Request.Body, maxToolRequestBodyBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, Fail(errCodeRequestTooLarge, "Request body exceeds the maximum allowed size"))
			return nil, false
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, Fail(errCodeInvalidRequest, "Unable to read request body"))
		return nil, false
	}
	return body, true
}

func isEventStream(contentType string) bool {
	return core.HasPrefix(core.Lower(core.Trim(contentType)), "text/event-stream")
}
