// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// ── AuthentikUser ──────────────────────────────────────────────────────

func TestAuthentikUser_Good(t *testing.T) {
	u := api.AuthentikUser{
		Username:     "alice",
		Email:        "alice@example.com",
		Name:         "Alice Smith",
		UID:          "abc-123",
		Groups:       []string{"editors", "admins"},
		Entitlements: []string{"premium"},
		JWT:          "tok.en.here",
	}

	if u.Username != "alice" {
		t.Fatalf("expected Username=%q, got %q", "alice", u.Username)
	}
	if u.Email != "alice@example.com" {
		t.Fatalf("expected Email=%q, got %q", "alice@example.com", u.Email)
	}
	if u.Name != "Alice Smith" {
		t.Fatalf("expected Name=%q, got %q", "Alice Smith", u.Name)
	}
	if u.UID != "abc-123" {
		t.Fatalf("expected UID=%q, got %q", "abc-123", u.UID)
	}
	if len(u.Groups) != 2 || u.Groups[0] != "editors" {
		t.Fatalf("expected Groups=[editors admins], got %v", u.Groups)
	}
	if len(u.Entitlements) != 1 || u.Entitlements[0] != "premium" {
		t.Fatalf("expected Entitlements=[premium], got %v", u.Entitlements)
	}
	if u.JWT != "tok.en.here" {
		t.Fatalf("expected JWT=%q, got %q", "tok.en.here", u.JWT)
	}
}

func TestAuthentikUserHasGroup_Good(t *testing.T) {
	u := api.AuthentikUser{
		Groups: []string{"editors", "admins"},
	}

	if !u.HasGroup("admins") {
		t.Fatal("expected HasGroup(admins) = true")
	}
	if !u.HasGroup("editors") {
		t.Fatal("expected HasGroup(editors) = true")
	}
}

func TestAuthentikUserHasGroup_Bad_Empty(t *testing.T) {
	u := api.AuthentikUser{}

	if u.HasGroup("admins") {
		t.Fatal("expected HasGroup(admins) = false for empty user")
	}
}

func TestAuthentikConfig_Good(t *testing.T) {
	cfg := api.AuthentikConfig{
		Issuer:       "https://auth.example.com",
		ClientID:     "my-client",
		TrustedProxy: true,
		PublicPaths:  []string{"/public", "/docs"},
	}

	if cfg.Issuer != "https://auth.example.com" {
		t.Fatalf("expected Issuer=%q, got %q", "https://auth.example.com", cfg.Issuer)
	}
	if cfg.ClientID != "my-client" {
		t.Fatalf("expected ClientID=%q, got %q", "my-client", cfg.ClientID)
	}
	if !cfg.TrustedProxy {
		t.Fatal("expected TrustedProxy=true")
	}
	if len(cfg.PublicPaths) != 2 {
		t.Fatalf("expected 2 public paths, got %d", len(cfg.PublicPaths))
	}
}

// ── Forward auth middleware ────────────────────────────────────────────

func TestForwardAuthHeaders_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := api.AuthentikConfig{TrustedProxy: true}
	e, _ := api.New(api.WithAuthentik(cfg))

	var gotUser *api.AuthentikUser
	e.Register(&authTestGroup{onRequest: func(c *gin.Context) {
		gotUser = api.GetUser(c)
		c.JSON(http.StatusOK, api.OK("ok"))
	}})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/check", nil)
	req.Header.Set("X-authentik-username", "bob")
	req.Header.Set("X-authentik-email", "bob@example.com")
	req.Header.Set("X-authentik-name", "Bob Jones")
	req.Header.Set("X-authentik-uid", "uid-456")
	req.Header.Set("X-authentik-jwt", "jwt.tok.en")
	req.Header.Set("X-authentik-groups", "staff|admins|ops")
	req.Header.Set("X-authentik-entitlements", "read|write")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUser == nil {
		t.Fatal("expected GetUser to return a user, got nil")
	}
	if gotUser.Username != "bob" {
		t.Fatalf("expected Username=%q, got %q", "bob", gotUser.Username)
	}
	if gotUser.Email != "bob@example.com" {
		t.Fatalf("expected Email=%q, got %q", "bob@example.com", gotUser.Email)
	}
	if gotUser.Name != "Bob Jones" {
		t.Fatalf("expected Name=%q, got %q", "Bob Jones", gotUser.Name)
	}
	if gotUser.UID != "uid-456" {
		t.Fatalf("expected UID=%q, got %q", "uid-456", gotUser.UID)
	}
	if gotUser.JWT != "jwt.tok.en" {
		t.Fatalf("expected JWT=%q, got %q", "jwt.tok.en", gotUser.JWT)
	}
	if len(gotUser.Groups) != 3 {
		t.Fatalf("expected 3 groups, got %d: %v", len(gotUser.Groups), gotUser.Groups)
	}
	if gotUser.Groups[0] != "staff" || gotUser.Groups[1] != "admins" || gotUser.Groups[2] != "ops" {
		t.Fatalf("expected groups [staff admins ops], got %v", gotUser.Groups)
	}
	if len(gotUser.Entitlements) != 2 {
		t.Fatalf("expected 2 entitlements, got %d: %v", len(gotUser.Entitlements), gotUser.Entitlements)
	}
	if gotUser.Entitlements[0] != "read" || gotUser.Entitlements[1] != "write" {
		t.Fatalf("expected entitlements [read write], got %v", gotUser.Entitlements)
	}
}

func TestForwardAuthHeaders_Good_NoHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := api.AuthentikConfig{TrustedProxy: true}
	e, _ := api.New(api.WithAuthentik(cfg))

	var gotUser *api.AuthentikUser
	e.Register(&authTestGroup{onRequest: func(c *gin.Context) {
		gotUser = api.GetUser(c)
		c.JSON(http.StatusOK, api.OK("ok"))
	}})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/check", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUser != nil {
		t.Fatalf("expected GetUser to return nil without headers, got %+v", gotUser)
	}
}

func TestForwardAuthHeaders_Bad_NotTrusted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := api.AuthentikConfig{TrustedProxy: false}
	e, _ := api.New(api.WithAuthentik(cfg))

	var gotUser *api.AuthentikUser
	e.Register(&authTestGroup{onRequest: func(c *gin.Context) {
		gotUser = api.GetUser(c)
		c.JSON(http.StatusOK, api.OK("ok"))
	}})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/check", nil)
	req.Header.Set("X-authentik-username", "mallory")
	req.Header.Set("X-authentik-email", "mallory@evil.com")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUser != nil {
		t.Fatalf("expected GetUser to return nil when TrustedProxy=false, got %+v", gotUser)
	}
}

func TestHealthBypassesAuthentik_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := api.AuthentikConfig{TrustedProxy: true}
	e, _ := api.New(api.WithAuthentik(cfg))

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /health, got %d", w.Code)
	}
}

func TestGetUser_Good_NilContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Engine without WithAuthentik — GetUser should return nil.
	e, _ := api.New()

	var gotUser *api.AuthentikUser
	e.Register(&authTestGroup{onRequest: func(c *gin.Context) {
		gotUser = api.GetUser(c)
		c.JSON(http.StatusOK, api.OK("ok"))
	}})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/check", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUser != nil {
		t.Fatalf("expected GetUser to return nil without middleware, got %+v", gotUser)
	}
}

// ── JWT validation ────────────────────────────────────────────────────

func TestJWTValidation_Bad_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use a fake issuer that won't resolve — JWT validation should fail open.
	cfg := api.AuthentikConfig{
		Issuer:   "https://fake-issuer.invalid",
		ClientID: "test-client",
	}
	e, _ := api.New(api.WithAuthentik(cfg))

	var gotUser *api.AuthentikUser
	e.Register(&authTestGroup{onRequest: func(c *gin.Context) {
		gotUser = api.GetUser(c)
		c.JSON(http.StatusOK, api.OK("ok"))
	}})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/check", nil)
	req.Header.Set("Authorization", "Bearer invalid-jwt-token")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (permissive), got %d", w.Code)
	}
	if gotUser != nil {
		t.Fatalf("expected GetUser to return nil for invalid JWT, got %+v", gotUser)
	}
}

func TestBearerAndAuthentikCoexist_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Engine with BOTH bearer auth AND authentik middleware.
	cfg := api.AuthentikConfig{TrustedProxy: true}
	e, _ := api.New(
		api.WithBearerAuth("secret-token"),
		api.WithAuthentik(cfg),
	)

	var gotUser *api.AuthentikUser
	e.Register(&authTestGroup{onRequest: func(c *gin.Context) {
		gotUser = api.GetUser(c)
		c.JSON(http.StatusOK, api.OK("ok"))
	}})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/check", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("X-authentik-username", "carol")
	req.Header.Set("X-authentik-email", "carol@example.com")
	req.Header.Set("X-authentik-name", "Carol White")
	req.Header.Set("X-authentik-uid", "uid-789")
	req.Header.Set("X-authentik-groups", "developers|admins")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUser == nil {
		t.Fatal("expected GetUser to return a user, got nil")
	}
	if gotUser.Username != "carol" {
		t.Fatalf("expected Username=%q, got %q", "carol", gotUser.Username)
	}
	if gotUser.Email != "carol@example.com" {
		t.Fatalf("expected Email=%q, got %q", "carol@example.com", gotUser.Email)
	}
	if len(gotUser.Groups) != 2 || gotUser.Groups[0] != "developers" || gotUser.Groups[1] != "admins" {
		t.Fatalf("expected groups [developers admins], got %v", gotUser.Groups)
	}
}

// ── RequireAuth / RequireGroup ────────────────────────────────────────

func TestRequireAuth_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := api.AuthentikConfig{TrustedProxy: true}
	e, _ := api.New(api.WithAuthentik(cfg))
	e.Register(&protectedGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/protected/data", nil)
	req.Header.Set("X-authentik-username", "alice")
	req.Header.Set("X-authentik-email", "alice@example.com")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequireAuth_Bad_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := api.AuthentikConfig{TrustedProxy: true}
	e, _ := api.New(api.WithAuthentik(cfg))
	e.Register(&protectedGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/protected/data", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `"unauthorised"`) {
		t.Fatalf("expected error code 'unauthorised' in body, got %s", body)
	}
}

func TestRequireAuth_Bad_NoAuthentikMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Engine without WithAuthentik — RequireAuth should still reject.
	e, _ := api.New()
	e.Register(&protectedGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/protected/data", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequireGroup_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := api.AuthentikConfig{TrustedProxy: true}
	e, _ := api.New(api.WithAuthentik(cfg))
	e.Register(&groupRequireGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/admin/panel", nil)
	req.Header.Set("X-authentik-username", "admin-user")
	req.Header.Set("X-authentik-groups", "admins|staff")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequireGroup_Bad_WrongGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := api.AuthentikConfig{TrustedProxy: true}
	e, _ := api.New(api.WithAuthentik(cfg))
	e.Register(&groupRequireGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/admin/panel", nil)
	req.Header.Set("X-authentik-username", "dev-user")
	req.Header.Set("X-authentik-groups", "developers")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `"forbidden"`) {
		t.Fatalf("expected error code 'forbidden' in body, got %s", body)
	}
}

// ── Test helpers ───────────────────────────────────────────────────────

// authTestGroup provides a /v1/check endpoint that calls a custom handler.
type authTestGroup struct {
	onRequest func(c *gin.Context)
}

func (a *authTestGroup) Name() string     { return "auth-test" }
func (a *authTestGroup) BasePath() string { return "/v1" }
func (a *authTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/check", a.onRequest)
}

// protectedGroup provides a /v1/protected/data endpoint guarded by RequireAuth.
type protectedGroup struct{}

func (g *protectedGroup) Name() string     { return "protected" }
func (g *protectedGroup) BasePath() string { return "/v1/protected" }
func (g *protectedGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/data", api.RequireAuth(), func(c *gin.Context) {
		user := api.GetUser(c)
		c.JSON(200, api.OK(user.Username))
	})
}

// groupRequireGroup provides a /v1/admin/panel endpoint guarded by RequireGroup.
type groupRequireGroup struct{}

func (g *groupRequireGroup) Name() string     { return "adminonly" }
func (g *groupRequireGroup) BasePath() string { return "/v1/admin" }
func (g *groupRequireGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/panel", api.RequireGroup("admins"), func(c *gin.Context) {
		c.JSON(200, api.OK("admin panel"))
	})
}
