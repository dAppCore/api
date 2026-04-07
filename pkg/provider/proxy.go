// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	coreapi "dappco.re/go/core/api"
	"github.com/gin-gonic/gin"
)

// ProxyConfig configures a ProxyProvider that reverse-proxies to an upstream
// process (typically a runtime provider binary listening on 127.0.0.1).
type ProxyConfig struct {
	// Name is the provider identity, e.g. "cool-widget".
	Name string

	// BasePath is the API route prefix, e.g. "/api/v1/cool-widget".
	BasePath string

	// Upstream is the full URL of the upstream process,
	// e.g. "http://127.0.0.1:9901".
	Upstream string

	// Element describes the custom element for GUI rendering.
	// Leave zero-value if the provider has no UI.
	Element ElementSpec

	// SpecFile is the filesystem path to the provider's OpenAPI spec.
	// Used by the Swagger aggregator. Leave empty if none.
	SpecFile string
}

// ProxyProvider reverse-proxies requests to an upstream HTTP process.
// It implements Provider and Renderable so it integrates with the
// service provider framework and GUI discovery.
type ProxyProvider struct {
	config ProxyConfig
	proxy  *httputil.ReverseProxy
	err    error
}

// NewProxy creates a ProxyProvider from the given configuration.
// Invalid upstream URLs do not panic; the provider retains the
// configuration error and responds with a standard 500 envelope when
// mounted. This keeps provider construction safe for callers.
func NewProxy(cfg ProxyConfig) *ProxyProvider {
	target, err := url.Parse(cfg.Upstream)
	if err != nil {
		return &ProxyProvider{
			config: cfg,
			err:    err,
		}
	}

	// url.Parse accepts inputs like "127.0.0.1:9901" without error — they
	// parse without a scheme or host, which causes httputil.ReverseProxy to
	// fail silently at runtime. Require both to be present.
	if target.Scheme == "" || target.Host == "" {
		return &ProxyProvider{
			config: cfg,
			err:    fmt.Errorf("upstream %q must include a scheme and host (e.g. http://127.0.0.1:9901)", cfg.Upstream),
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Preserve the original Director but strip the base path so the
	// upstream receives clean paths (e.g. /items instead of /api/v1/cool-widget/items).
	defaultDirector := proxy.Director
	basePath := strings.TrimSuffix(cfg.BasePath, "/")

	proxy.Director = func(req *http.Request) {
		defaultDirector(req)
		// Strip the base path prefix from the request path.
		req.URL.Path = stripBasePath(req.URL.Path, basePath)
		if req.URL.RawPath != "" {
			req.URL.RawPath = stripBasePath(req.URL.RawPath, basePath)
		}
	}

	return &ProxyProvider{
		config: cfg,
		proxy:  proxy,
	}
}

// Err reports any configuration error detected while constructing the proxy.
// A nil error means the proxy is ready to mount and serve requests.
func (p *ProxyProvider) Err() error {
	if p == nil {
		return nil
	}
	return p.err
}

// stripBasePath removes an exact base path prefix from a request path.
// It only strips when the path matches the base path itself or lives under
// the base path boundary, so "/api" will not accidentally trim "/api-v2".
func stripBasePath(path, basePath string) string {
	basePath = strings.TrimSuffix(strings.TrimSpace(basePath), "/")
	if basePath == "" || basePath == "/" {
		if path == "" {
			return "/"
		}
		return path
	}

	if path == basePath {
		return "/"
	}

	prefix := basePath + "/"
	if strings.HasPrefix(path, prefix) {
		trimmed := strings.TrimPrefix(path, basePath)
		if trimmed == "" {
			return "/"
		}
		return trimmed
	}

	return path
}

// Name returns the provider identity.
func (p *ProxyProvider) Name() string {
	return p.config.Name
}

// BasePath returns the API route prefix.
func (p *ProxyProvider) BasePath() string {
	return p.config.BasePath
}

// RegisterRoutes mounts a catch-all reverse proxy handler on the router group.
func (p *ProxyProvider) RegisterRoutes(rg *gin.RouterGroup) {
	rg.Any("/*path", func(c *gin.Context) {
		if p == nil || p.err != nil || p.proxy == nil {
			details := map[string]any{}
			if p != nil && p.err != nil {
				details["error"] = p.err.Error()
			}
			c.JSON(http.StatusInternalServerError, coreapi.FailWithDetails(
				"invalid_provider_configuration",
				"Provider is misconfigured",
				details,
			))
			return
		}

		// Use the underlying http.ResponseWriter directly. Gin's
		// responseWriter wrapper does not implement http.CloseNotifier,
		// which httputil.ReverseProxy requires for cancellation signalling.
		var w http.ResponseWriter = c.Writer
		if uw, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
			w = uw.Unwrap()
		}
		p.proxy.ServeHTTP(w, c.Request)
	})
}

// Element returns the custom element specification for GUI rendering.
func (p *ProxyProvider) Element() ElementSpec {
	return p.config.Element
}

// SpecFile returns the path to the provider's OpenAPI spec file.
func (p *ProxyProvider) SpecFile() string {
	return p.config.SpecFile
}

// Upstream returns the upstream URL string.
func (p *ProxyProvider) Upstream() string {
	return p.config.Upstream
}
