// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	api "dappco.re/go/api"
)

// ── Helpers ─────────────────────────────────────────────────────────────

// sessionTestGroup provides /sess/set and /sess/get endpoints for session tests.
type sessionTestGroup struct{}

func (s *sessionTestGroup) Name() string     { return "sess" }
func (s *sessionTestGroup) BasePath() string { return "/sess" }
func (s *sessionTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/set", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("key", "value")
		session.Save()
		c.JSON(http.StatusOK, api.OK("saved"))
	})
	rg.GET("/get", func(c *gin.Context) {
		session := sessions.Default(c)
		val := session.Get("key")
		c.JSON(http.StatusOK, api.OK(val))
	})
}

// ── WithSessions ────────────────────────────────────────────────────────

func TestWithSessions_Good_SetsSessionCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSessions("session", []byte("test-secret-key!")))
	e.Register(&sessionTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/sess/set", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected Set-Cookie header with name 'session'")
	}
}

func TestWithSessions_Good_SessionPersistsAcrossRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSessions("session", []byte("test-secret-key!")))
	e.Register(&sessionTestGroup{})

	h := e.Handler()

	// First request: set session value.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/sess/set", nil)
	h.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("set: expected 200, got %d", w1.Code)
	}

	// Extract the session cookie from the response.
	var sessionCookie *http.Cookie
	for _, c := range w1.Result().Cookies() {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("set: expected session cookie in response")
	}

	// Second request: get session value, sending the cookie back.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/sess/get", nil)
	req2.AddCookie(sessionCookie)
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w2.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	data, ok := resp.Data.(string)
	if !ok || data != "value" {
		t.Fatalf("expected Data=%q, got %v", "value", resp.Data)
	}
}

func TestWithSessions_Good_EmptySessionReturnsNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(api.WithSessions("session", []byte("test-secret-key!")))
	e.Register(&sessionTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/sess/get", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[any]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if resp.Data != nil {
		t.Fatalf("expected nil Data for empty session, got %v", resp.Data)
	}
}

func TestWithSessions_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e, _ := api.New(
		api.WithSessions("session", []byte("test-secret-key!")),
		api.WithRequestID(),
	)
	e.Register(&sessionTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/sess/set", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Session cookie should be present.
	found := false
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected session cookie")
	}

	// Request ID should also be present.
	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

func TestWithSessions_Ugly_DoubleSessionsDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Applying WithSessions twice should not panic.
	e, err := api.New(
		api.WithSessions("session", []byte("secret-one-here!")),
		api.WithSessions("session", []byte("secret-two-here!")),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e.Register(&sessionTestGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/sess/set", nil)
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
