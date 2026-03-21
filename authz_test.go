// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"

	api "dappco.re/go/core/api"
)

// casbinModel is a minimal RESTful ACL model for testing authorisation.
const casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch(r.obj, p.obj) && r.act == p.act
`

// newTestEnforcer creates a Casbin enforcer from the inline model and adds
// the given policies programmatically. Each policy is a [subject, object, action] triple.
func newTestEnforcer(t *testing.T, policies [][3]string) *casbin.Enforcer {
	t.Helper()

	m, err := model.NewModelFromString(casbinModel)
	if err != nil {
		t.Fatalf("failed to create casbin model: %v", err)
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("failed to create casbin enforcer: %v", err)
	}

	for _, p := range policies {
		if _, err := e.AddPolicy(p[0], p[1], p[2]); err != nil {
			t.Fatalf("failed to add policy %v: %v", p, err)
		}
	}

	return e
}

// setBasicAuth sets the HTTP Basic Authentication header on a request.
func setBasicAuth(req *http.Request, user, pass string) {
	req.SetBasicAuth(user, pass)
}

// ── WithAuthz ─────────────────────────────────────────────────────────────

func TestWithAuthz_Good_AllowsPermittedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	enforcer := newTestEnforcer(t, [][3]string{
		{"alice", "/stub/*", "GET"},
	})

	e, _ := api.New(api.WithAuthz(enforcer))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	setBasicAuth(req, "alice", "secret")

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for permitted request, got %d", w.Code)
	}
}

func TestWithAuthz_Bad_DeniesUnpermittedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Only alice is permitted; bob has no policy entry.
	enforcer := newTestEnforcer(t, [][3]string{
		{"alice", "/stub/*", "GET"},
	})

	e, _ := api.New(api.WithAuthz(enforcer))
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	setBasicAuth(req, "bob", "secret")

	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unpermitted request, got %d", w.Code)
	}
}

func TestWithAuthz_Good_DifferentMethodsEvaluatedSeparately(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// alice can GET but not DELETE.
	enforcer := newTestEnforcer(t, [][3]string{
		{"alice", "/stub/*", "GET"},
	})

	e, _ := api.New(api.WithAuthz(enforcer))
	e.Register(&stubGroup{})

	h := e.Handler()

	// GET should succeed.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	setBasicAuth(req, "alice", "secret")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET, got %d", w.Code)
	}

	// DELETE should be denied (no policy for DELETE).
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/stub/ping", nil)
	setBasicAuth(req, "alice", "secret")
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for DELETE, got %d", w.Code)
	}
}

func TestWithAuthz_Good_CombinesWithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	enforcer := newTestEnforcer(t, [][3]string{
		{"alice", "/stub/*", "GET"},
	})

	e, _ := api.New(
		api.WithRequestID(),
		api.WithAuthz(enforcer),
	)
	e.Register(&stubGroup{})

	h := e.Handler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
	setBasicAuth(req, "alice", "secret")

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Both authz (allowed) and request ID should be active.
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header from WithRequestID")
	}
}

// casbinWildcardModel extends the base model with a matcher that treats
// "*" as a wildcard subject, allowing any authenticated user through.
const casbinWildcardModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (r.sub == p.sub || p.sub == "*") && keyMatch(r.obj, p.obj) && r.act == p.act
`

func TestWithAuthz_Ugly_WildcardPolicyAllowsAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use a model whose matcher treats "*" as a wildcard subject.
	m, err := model.NewModelFromString(casbinWildcardModel)
	if err != nil {
		t.Fatalf("failed to create casbin model: %v", err)
	}

	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("failed to create casbin enforcer: %v", err)
	}

	if _, err := enforcer.AddPolicy("*", "/stub/*", "GET"); err != nil {
		t.Fatalf("failed to add wildcard policy: %v", err)
	}

	e, _ := api.New(api.WithAuthz(enforcer))
	e.Register(&stubGroup{})

	h := e.Handler()

	// Any user should be allowed by the wildcard policy.
	for _, user := range []string{"alice", "bob", "charlie"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/stub/ping", nil)
		setBasicAuth(req, user, "secret")
		h.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for user %q with wildcard policy, got %d", user, w.Code)
		}
	}
}
