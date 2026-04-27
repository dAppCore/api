// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"  // Note: AX-6 - OIDC verifier APIs require context.Context; no core primitive.
	"net/http" // Note: AX-6 - structural HTTP status boundary for Gin auth responses; no core primitive.
	"slices"

	core "dappco.re/go/core"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

// AuthentikConfig holds settings for the Authentik forward-auth integration.
//
// Example:
//
//	cfg := api.AuthentikConfig{Issuer: "https://auth.example.com/", ClientID: "core-api"}
type AuthentikConfig struct {
	// Issuer is the OIDC issuer URL (e.g. https://auth.example.com/application/o/my-app/).
	Issuer string

	// ClientID is the OAuth2 client identifier.
	ClientID string

	// TrustedProxy enables reading X-authentik-* headers set by a reverse proxy.
	// When false, headers are ignored to prevent spoofing from untrusted sources.
	TrustedProxy bool

	// PublicPaths lists additional paths that do not require authentication.
	// /health and the configured Swagger UI path are always public.
	PublicPaths []string
}

// AuthentikConfig returns the configured Authentik settings for the engine.
//
// The result snapshots the Engine state at call time and clones slices so
// callers can safely reuse or modify the returned value.
//
// Example:
//
//	cfg := engine.AuthentikConfig()
func (e *Engine) AuthentikConfig() AuthentikConfig {
	if e == nil {
		return AuthentikConfig{}
	}

	return cloneAuthentikConfig(e.authentikConfig)
}

// AuthentikUser represents an authenticated user extracted from Authentik
// forward-auth headers or a validated JWT.
//
// Example:
//
//	user := &api.AuthentikUser{Username: "alice", Groups: []string{"admins"}}
type AuthentikUser struct {
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	Name         string   `json:"name"`
	UID          string   `json:"uid"`
	Groups       []string `json:"groups,omitempty"`
	Entitlements []string `json:"entitlements,omitempty"`
	JWT          string   `json:"-"`
}

// HasGroup reports whether the user belongs to the named group.
//
// Example:
//
//	user.HasGroup("admins")
func (u *AuthentikUser) HasGroup(group string) bool {
	return slices.Contains(u.Groups, group)
}

// authentikUserKey is the Gin context key used to store the authenticated user.
const authentikUserKey = "authentik_user"

// GetUser retrieves the AuthentikUser from the Gin context.
// Returns nil when no user has been set (unauthenticated request or
// middleware not active).
//
// Example:
//
//	user := api.GetUser(c)
func GetUser(c *gin.Context) *AuthentikUser {
	val, exists := c.Get(authentikUserKey)
	if !exists {
		return nil
	}
	user, ok := val.(*AuthentikUser)
	if !ok {
		return nil
	}
	return user
}

// oidcProviderMu guards the provider cache.
var oidcProviderMu core.Mutex

// oidcProviders caches OIDC providers by issuer URL to avoid repeated
// discovery requests.
var oidcProviders = make(map[string]*oidc.Provider)

// getOIDCProvider returns a cached OIDC provider for the given issuer,
// performing discovery on first access.
func getOIDCProvider(ctx context.Context, issuer string) (*oidc.Provider, error) {
	oidcProviderMu.Lock()
	defer oidcProviderMu.Unlock()

	if p, ok := oidcProviders[issuer]; ok {
		return p, nil
	}

	p, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	oidcProviders[issuer] = p
	return p, nil
}

// validateJWT verifies a raw JWT against the configured OIDC issuer and
// extracts user claims on success.
func validateJWT(ctx context.Context, cfg AuthentikConfig, rawToken string) (*AuthentikUser, error) {
	provider, err := getOIDCProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	idToken, err := verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	var claims struct {
		PreferredUsername string   `json:"preferred_username"`
		Email             string   `json:"email"`
		Name              string   `json:"name"`
		Sub               string   `json:"sub"`
		Groups            []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &AuthentikUser{
		Username: claims.PreferredUsername,
		Email:    claims.Email,
		Name:     claims.Name,
		UID:      claims.Sub,
		Groups:   claims.Groups,
		JWT:      rawToken,
	}, nil
}

// authentikMiddleware returns Gin middleware that extracts user identity from
// X-authentik-* headers set by a trusted reverse proxy (e.g. Traefik with
// Authentik forward-auth) or from a JWT in the Authorization header.
//
// The middleware is PERMISSIVE: it populates the context when credentials are
// present but never rejects unauthenticated requests. Downstream handlers
// use GetUser to check authentication.
func authentikMiddleware(cfg AuthentikConfig, publicPaths func() []string) gin.HandlerFunc {
	// Build the set of public paths that skip header extraction entirely.
	public := map[string]bool{
		"/health":  true,
		"/swagger": true,
	}
	for _, p := range cfg.PublicPaths {
		public[p] = true
	}

	return func(c *gin.Context) {
		// Skip public paths.
		path := c.Request.URL.Path
		for p := range public {
			if isPublicPath(path, p) {
				c.Next()
				return
			}
		}
		if publicPaths != nil {
			for _, p := range publicPaths() {
				if isPublicPath(path, p) {
					c.Next()
					return
				}
			}
		}

		// Block 1: Extract user from X-authentik-* forward-auth headers.
		if cfg.TrustedProxy {
			username := c.GetHeader("X-authentik-username")
			if username != "" {
				user := &AuthentikUser{
					Username: username,
					Email:    c.GetHeader("X-authentik-email"),
					Name:     c.GetHeader("X-authentik-name"),
					UID:      c.GetHeader("X-authentik-uid"),
					JWT:      c.GetHeader("X-authentik-jwt"),
				}

				if groups := c.GetHeader("X-authentik-groups"); groups != "" {
					user.Groups = core.Split(groups, "|")
				}
				if ent := c.GetHeader("X-authentik-entitlements"); ent != "" {
					user.Entitlements = core.Split(ent, "|")
				}

				c.Set(authentikUserKey, user)
			}
		}

		// Block 2: Attempt JWT validation for direct API clients.
		// Only when OIDC is configured and no user was extracted from headers.
		if cfg.Issuer != "" && cfg.ClientID != "" && GetUser(c) == nil {
			if auth := c.GetHeader("Authorization"); core.HasPrefix(auth, "Bearer ") {
				rawToken := core.TrimPrefix(auth, "Bearer ")
				if user, err := validateJWT(c.Request.Context(), cfg, rawToken); err == nil {
					c.Set(authentikUserKey, user)
				}
				// On failure: continue without user (fail open / permissive).
			}
		}

		c.Next()
	}
}

func cloneAuthentikConfig(cfg AuthentikConfig) AuthentikConfig {
	out := cfg
	out.Issuer = core.Trim(out.Issuer)
	out.ClientID = core.Trim(out.ClientID)
	out.PublicPaths = normalisePublicPaths(cfg.PublicPaths)
	return out
}

// normalisePublicPaths trims whitespace, ensures a leading slash, and removes
// duplicate entries while preserving the first occurrence of each path.
func normalisePublicPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	out := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))

	for _, path := range paths {
		path = core.Trim(path)
		if path == "" {
			continue
		}
		if !core.HasPrefix(path, "/") {
			path = "/" + path
		}
		for core.HasSuffix(path, "/") && path != "/" {
			path = core.TrimSuffix(path, "/")
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

// RequireAuth is Gin middleware that rejects unauthenticated requests.
// It checks for a user set by the Authentik middleware and returns 401
// when none is present.
//
// Example:
//
//	r.GET("/private", api.RequireAuth(), handler)
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetUser(c) == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				Fail("unauthorised", "Authentication required"))
			return
		}
		c.Next()
	}
}

// RequireGroup is Gin middleware that rejects requests from users who do
// not belong to the specified group. Returns 401 when no user is present
// and 403 when the user lacks the required group membership.
//
// Example:
//
//	r.GET("/admin", api.RequireGroup("admins"), handler)
func RequireGroup(group string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				Fail("unauthorised", "Authentication required"))
			return
		}
		if !user.HasGroup(group) {
			c.AbortWithStatusJSON(http.StatusForbidden,
				Fail("forbidden", "Insufficient permissions"))
			return
		}
		c.Next()
	}
}
