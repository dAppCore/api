// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"io"       // Note: AX-6 - HTTP response bodies are stream boundaries.
	"net/http" // Note: AX-6 - entitlement bridge talks to PHP over HTTP.
	"net/url"  // Note: AX-6 - path escaping is required for feature/workspace URL segments.
	"time"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

const defaultEntitlementBridgeTimeout = 2 * time.Second

// maxEntitlementResponseBytes caps response body reads to defend against
// malformed or hostile upstream services consuming unbounded memory.
const maxEntitlementResponseBytes = 1 << 20 // 1 MiB

// EntitlementBridgeConfig configures the bridge from Go renderers to the
// PHP EntitlementService-backed API.
type EntitlementBridgeConfig struct {
	// BaseURL is the PHP API origin, e.g. "https://app.example.com".
	BaseURL string

	// Token is an optional service token. When empty, request Authorization
	// headers passed to Check/Callback are forwarded instead.
	Token string

	// HTTPClient overrides the default client. When nil, a bounded-timeout
	// client is created so render-time entitlement checks fail closed.
	HTTPClient *http.Client

	// Timeout configures the default client timeout. Zero uses a safe default.
	Timeout time.Duration
}

// EntitlementBridge resolves server-side entitlement gates against the
// authoritative PHP EntitlementService endpoint.
type EntitlementBridge struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewEntitlementBridge creates a bridge that can populate go-html's
// Entitlements callback without importing go-html here.
func NewEntitlementBridge(cfg EntitlementBridgeConfig) *EntitlementBridge {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultEntitlementBridgeTimeout
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	return &EntitlementBridge{
		baseURL: trimTrailingSlash(core.Trim(cfg.BaseURL)),
		token:   core.Trim(cfg.Token),
		client:  client,
	}
}

// Check returns whether feature is allowed for the current workspace. A blank
// workspaceID uses the current-workspace PHP route, resolved from forwarded
// request auth/session headers.
func (b *EntitlementBridge) Check(ctx context.Context, workspaceID, feature string, headers http.Header) (bool, error) {
	const op = "EntitlementBridge.Check"

	if b == nil {
		return false, core.E(op, "nil entitlement bridge", nil)
	}
	if b.client == nil {
		return false, core.E(op, "nil HTTP client", nil)
	}
	if b.baseURL == "" {
		return false, core.E(op, "base URL is required", nil)
	}

	feature = core.Trim(feature)
	if feature == "" {
		return false, core.E(op, "feature is required", nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.checkURL(workspaceID, feature), nil)
	if err != nil {
		return false, core.E(op, "build entitlement request", err)
	}
	req.Header.Set("Accept", "application/json")
	applyEntitlementHeaders(req.Header, headers, b.token, workspaceID)

	resp, err := b.client.Do(req)
	if err != nil {
		return false, core.E(op, "call entitlement service", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxEntitlementResponseBytes))
	if err != nil {
		return false, core.E(op, "read entitlement response", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, core.E(op, core.Sprintf("entitlement service returned %d", resp.StatusCode), nil)
	}

	allowed, ok := entitlementAllowed(data)
	if !ok {
		return false, core.E(op, "entitlement response did not include an allowed flag", nil)
	}
	return allowed, nil
}

// Callback returns the exact callback shape expected by go-html render
// contexts: func(feature string) bool. Errors fail closed by returning false.
func (b *EntitlementBridge) Callback(ctx context.Context, workspaceID string, headers http.Header) func(feature string) bool {
	return func(feature string) bool {
		allowed, err := b.Check(ctx, workspaceID, feature, headers)
		return err == nil && allowed
	}
}

// CallbackForRequest binds the current request context and auth/session
// headers to an entitlement callback suitable for server-side render contexts.
func (b *EntitlementBridge) CallbackForRequest(r *http.Request, workspaceID string) func(feature string) bool {
	if r == nil {
		return b.Callback(context.Background(), workspaceID, nil)
	}
	return b.Callback(r.Context(), workspaceID, r.Header)
}

// CallbackForGin is a convenience wrapper for Gin handlers that render
// go-html components.
func (b *EntitlementBridge) CallbackForGin(c *gin.Context, workspaceID string) func(feature string) bool {
	if c == nil {
		return b.Callback(context.Background(), workspaceID, nil)
	}
	return b.CallbackForRequest(c.Request, workspaceID)
}

func (b *EntitlementBridge) checkURL(workspaceID, feature string) string {
	workspaceID = core.Trim(workspaceID)
	if workspaceID == "" {
		return b.baseURL + "/api/entitlements/check/" + url.PathEscape(feature)
	}
	return b.baseURL + "/api/v1/workspaces/" + url.PathEscape(workspaceID) + "/entitlements/check/" + url.PathEscape(feature)
}

func applyEntitlementHeaders(dst, src http.Header, token, workspaceID string) {
	if token != "" {
		dst.Set("Authorization", "Bearer "+token)
	} else if src != nil {
		if auth := src.Get("Authorization"); auth != "" {
			dst.Set("Authorization", auth)
		}
	}
	hasAuthorization := dst.Get("Authorization") != ""

	for _, name := range []string{"Cookie", "X-Request-ID", "X-Workspace-Id"} {
		if src == nil {
			continue
		}
		if name == "Cookie" && hasAuthorization {
			continue
		}
		if value := src.Get(name); value != "" {
			dst.Set(name, value)
		}
	}

	if workspaceID = core.Trim(workspaceID); workspaceID != "" {
		dst.Set("X-Workspace-Id", workspaceID)
	}
}

func entitlementAllowed(data []byte) (bool, bool) {
	var payload map[string]any
	if result := core.JSONUnmarshal(data, &payload); !result.OK {
		return false, false
	}

	if allowed, ok := boolValue(payload, "allowed"); ok {
		return allowed, true
	}

	entitlement, ok := payload["entitlement"].(map[string]any)
	if !ok {
		return false, false
	}
	for _, key := range []string{"allowed", "can", "enabled", "entitled"} {
		if allowed, ok := boolValue(entitlement, key); ok {
			return allowed, true
		}
	}
	return false, false
}

func boolValue(values map[string]any, key string) (bool, bool) {
	value, ok := values[key]
	if !ok {
		return false, false
	}
	allowed, ok := value.(bool)
	return allowed, ok
}

func trimTrailingSlash(value string) string {
	for core.HasSuffix(value, "/") && value != "/" {
		value = core.TrimSuffix(value, "/")
	}
	return value
}
