// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url" // Note: AX-6 — net/url url.URL fields are structural for reverse-proxy URL rewriting.
	"os"
	"strconv"

	core "dappco.re/go/core"

	coreapi "dappco.re/go/api"
	"github.com/gin-gonic/gin"
)

const providerUpstreamAllowEnv = "CORE_PROVIDER_UPSTREAM_ALLOW"

// ErrProviderUpstreamBlocked marks provider upstream URL rejections by the
// construction-time SSRF guard.
var ErrProviderUpstreamBlocked = errors.New("provider upstream blocked by SSRF guard")

// ProviderUpstreamBlockedError carries the concrete rejection reason for a
// provider upstream URL blocked by the SSRF guard.
type ProviderUpstreamBlockedError struct {
	Upstream string
	Reason   string
	Cause    error
}

func (e *ProviderUpstreamBlockedError) Error() string {
	if e == nil {
		return ErrProviderUpstreamBlocked.Error()
	}
	if e.Reason == "" {
		return ErrProviderUpstreamBlocked.Error()
	}
	return ErrProviderUpstreamBlocked.Error() + ": " + e.Reason
}

func (e *ProviderUpstreamBlockedError) Is(target error) bool {
	return target == ErrProviderUpstreamBlocked
}

func (e *ProviderUpstreamBlockedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

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
	parsed := core.URLParse(cfg.Upstream)
	if !parsed.OK {
		err, ok := parsed.Value.(error)
		if !ok {
			err = core.E("ProxyProvider.New", "invalid upstream URL", nil)
		}
		return &ProxyProvider{
			config: cfg,
			err:    err,
		}
	}

	target, ok := parsed.Value.(*url.URL)
	if !ok || target == nil {
		return &ProxyProvider{
			config: cfg,
			err:    core.E("ProxyProvider.New", "invalid upstream URL result", nil),
		}
	}

	// url.Parse accepts inputs like "127.0.0.1:9901" without error — they
	// parse without a scheme or host, which causes httputil.ReverseProxy to
	// fail silently at runtime. Require both to be present.
	if target.Scheme == "" || target.Host == "" {
		return &ProxyProvider{
			config: cfg,
			err:    core.E("ProxyProvider.New", core.Sprintf("upstream %q must include a scheme and host (e.g. http://127.0.0.1:9901)", cfg.Upstream), nil),
		}
	}

	if err := validateProviderUpstreamURL(cfg.Upstream, target); err != nil {
		return &ProxyProvider{
			config: cfg,
			err:    err,
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Preserve the original Director but strip the base path so the
	// upstream receives clean paths (e.g. /items instead of /api/v1/cool-widget/items).
	defaultDirector := proxy.Director
	basePath := core.TrimSuffix(cfg.BasePath, "/")

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
	basePath = core.TrimSuffix(core.Trim(basePath), "/")
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
	if core.HasPrefix(path, prefix) {
		trimmed := core.TrimPrefix(path, basePath)
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

func validateProviderUpstreamURL(raw string, target *url.URL) error {
	if target == nil {
		return blockProviderUpstream(raw, "invalid upstream URL result", nil)
	}

	scheme := core.Lower(target.Scheme)
	if scheme != "http" && scheme != "https" {
		return blockProviderUpstream(raw, "only HTTP and HTTPS upstream URLs are permitted", nil)
	}
	if target.User != nil {
		return blockProviderUpstream(raw, "upstream URLs must not include embedded credentials", nil)
	}
	if port := target.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return blockProviderUpstream(raw, "upstream URL port is invalid", err)
		}
	}

	host := target.Hostname()
	if host == "" {
		return blockProviderUpstream(raw, "upstream URL must include a host", nil)
	}

	hostKey := core.TrimSuffix(core.Lower(host), ".")
	if _, ok := providerMetadataHosts[hostKey]; ok {
		return blockProviderUpstream(raw, "metadata host is not permitted: "+host, nil)
	}

	allowCIDRs, err := providerUpstreamAllowCIDRs()
	if err != nil {
		return blockProviderUpstream(raw, "invalid "+providerUpstreamAllowEnv+" entry", err)
	}

	if ip := net.ParseIP(host); ip != nil {
		return validateProviderUpstreamIP(raw, host, ip, allowCIDRs)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return blockProviderUpstream(raw, "DNS resolution failed for "+host, err)
	}
	if len(ips) == 0 {
		return blockProviderUpstream(raw, "DNS resolution returned no IPs for "+host, nil)
	}
	for _, ip := range ips {
		if err := validateProviderUpstreamIP(raw, host, ip, allowCIDRs); err != nil {
			return err
		}
	}

	return nil
}

func validateProviderUpstreamIP(raw, host string, ip net.IP, allowCIDRs []*net.IPNet) error {
	if reason := blockedProviderUpstreamIPReason(ip); reason != "" {
		if providerUpstreamIPAllowed(ip, allowCIDRs) {
			return nil
		}
		return blockProviderUpstream(raw, reason+" for "+host+": "+ip.String(), nil)
	}
	return nil
}

func blockProviderUpstream(raw, reason string, cause error) error {
	return &ProviderUpstreamBlockedError{
		Upstream: raw,
		Reason:   reason,
		Cause:    cause,
	}
}

func providerUpstreamAllowCIDRs() ([]*net.IPNet, error) {
	raw := core.Trim(os.Getenv(providerUpstreamAllowEnv))
	if raw == "" {
		return nil, nil
	}

	parts := core.Split(raw, ",")
	cidrs := make([]*net.IPNet, 0, len(parts))
	for _, part := range parts {
		value := core.Trim(part)
		if value == "" {
			continue
		}
		_, network, err := net.ParseCIDR(value)
		if err != nil {
			return nil, err
		}
		cidrs = append(cidrs, network)
	}
	return cidrs, nil
}

func providerUpstreamIPAllowed(ip net.IP, allowCIDRs []*net.IPNet) bool {
	for _, network := range allowCIDRs {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func blockedProviderUpstreamIPReason(ip net.IP) string {
	if ip == nil {
		return "invalid IP"
	}
	if ip.IsLoopback() {
		return "loopback IP"
	}
	if ip.IsPrivate() {
		return "private IP"
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return "link-local IP"
	}
	if ip.IsUnspecified() {
		return "unspecified IP"
	}
	if ip.IsMulticast() || ip.IsInterfaceLocalMulticast() {
		return "multicast IP"
	}
	if !ip.IsGlobalUnicast() {
		return "non-global-unicast IP"
	}
	for _, network := range providerBlockedCIDRs {
		if network.Contains(ip) {
			return "reserved IP"
		}
	}
	return ""
}

func mustParseProviderCIDRs(values ...string) []*net.IPNet {
	cidrs := make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		_, network, err := net.ParseCIDR(value)
		if err != nil {
			panic(core.Sprintf("provider: invalid upstream SSRF CIDR %q: %v", value, err))
		}
		cidrs = append(cidrs, network)
	}
	return cidrs
}

var providerMetadataHosts = map[string]struct{}{
	"metadata.google.internal": {},
	"metadata.googleapis.com":  {},
	"metadata.azure.com":       {},
	"169.254.169.254":          {},
	"fd00:ec2::254":            {},
	"100.100.100.200":          {},
}

var providerBlockedCIDRs = mustParseProviderCIDRs(
	"0.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/128",
	"::1/128",
	"64:ff9b:1::/48",
	"100::/64",
	"2001::/32",
	"2001:2::/48",
	"2001:db8::/32",
	"2002::/16",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
)
