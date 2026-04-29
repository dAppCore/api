// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"dappco.re/go/api/internal/stdcompat/fmt"
	"dappco.re/go/api/internal/stdcompat/json"
	"dappco.re/go/api/internal/stdcompat/os"
	"dappco.re/go/api/internal/stdcompat/strings"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	api "dappco.re/go/api"
	"github.com/gin-gonic/gin"
)

// testAuthRoutes provides endpoints for integration testing.
type testAuthRoutes struct{}

func (r *testAuthRoutes) Name() string     { return "authtest" }
func (r *testAuthRoutes) BasePath() string { return "/v1" }

func (r *testAuthRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/public", func(c *gin.Context) {
		c.JSON(200, api.OK("public"))
	})
	rg.GET("/whoami", api.RequireAuth(), func(c *gin.Context) {
		user := api.GetUser(c)
		c.JSON(200, api.OK(user))
	})
	rg.GET("/admin", api.RequireGroup("admins"), func(c *gin.Context) {
		user := api.GetUser(c)
		c.JSON(200, api.OK(user))
	})
}

// getClientCredentialsToken fetches a token from Authentik using
// the client_credentials grant.
func getClientCredentialsToken(t *testing.T, issuer, clientID, clientSecret string) (accessToken, idToken string) {
	t.Helper()

	// Discover token endpoint.
	disc := strings.TrimSuffix(issuer, "/") + "/.well-known/openid-configuration"
	resp, err := http.Get(disc)
	if err != nil {
		t.Fatalf("OIDC discovery failed: %v", err)
	}
	defer resp.Body.Close()

	var config struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		t.Fatalf("decode discovery: %v", err)
	}

	// Request token.
	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"scope":         {"openid email profile entitlements"},
	}
	resp, err = http.PostForm(config.TokenEndpoint, data)
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	if tokenResp.Error != "" {
		t.Fatalf("token error: %s — %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return tokenResp.AccessToken, tokenResp.IDToken
}

func TestAuthentikIntegration(t *testing.T) {
	// Skip unless explicitly enabled — requires live Authentik at auth.lthn.io.
	if os.Getenv("AUTHENTIK_INTEGRATION") != "1" {
		t.Skip("set AUTHENTIK_INTEGRATION=1 to run live Authentik tests")
	}

	issuer := envOr("AUTHENTIK_ISSUER", "https://auth.lthn.io/application/o/core-api/")
	clientID := envOr("AUTHENTIK_CLIENT_ID", "core-api")
	clientSecret := os.Getenv("AUTHENTIK_CLIENT_SECRET")
	if clientSecret == "" {
		t.Fatal("AUTHENTIK_CLIENT_SECRET is required")
	}

	gin.SetMode(gin.TestMode)

	// Fetch a real token from Authentik.
	t.Run("TokenAcquisition", func(t *testing.T) {
		access, id := getClientCredentialsToken(t, issuer, clientID, clientSecret)
		if access == "" {
			t.Fatal("empty access_token")
		}
		if id == "" {
			t.Fatal("empty id_token")
		}
		t.Logf("access_token length: %d", len(access))
		t.Logf("id_token length: %d", len(id))
	})

	// Build the engine with real Authentik config.
	engine, err := api.New(
		api.WithAuthentik(api.AuthentikConfig{
			Issuer:       issuer,
			ClientID:     clientID,
			TrustedProxy: true,
		}),
	)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	engine.Register(&testAuthRoutes{})
	ts := httptest.NewServer(engine.Handler())
	defer ts.Close()

	accessToken, _ := getClientCredentialsToken(t, issuer, clientID, clientSecret)

	t.Run("Health_NoAuth", func(t *testing.T) {
		resp := get(t, ts.URL+"/health", "")
		assertStatus(t, resp, 200)
		body := readBody(t, resp)
		t.Logf("health: %s", body)
	})

	t.Run("Public_NoAuth", func(t *testing.T) {
		resp := get(t, ts.URL+"/v1/public", "")
		assertStatus(t, resp, 200)
		body := readBody(t, resp)
		t.Logf("public: %s", body)
	})

	t.Run("Whoami_NoToken_401", func(t *testing.T) {
		resp := get(t, ts.URL+"/v1/whoami", "")
		assertStatus(t, resp, 401)
	})

	t.Run("Whoami_WithAccessToken", func(t *testing.T) {
		resp := get(t, ts.URL+"/v1/whoami", accessToken)
		assertStatus(t, resp, 200)
		body := readBody(t, resp)
		t.Logf("whoami (access_token): %s", body)

		// Parse response and verify user fields.
		var envelope struct {
			Data api.AuthentikUser `json:"data"`
		}
		if err := json.Unmarshal([]byte(body), &envelope); err != nil {
			t.Fatalf("parse whoami: %v", err)
		}
		if envelope.Data.UID == "" {
			t.Error("expected non-empty UID")
		}
		if !strings.Contains(envelope.Data.Username, "client_credentials") {
			t.Logf("username: %s (service account)", envelope.Data.Username)
		}
	})

	t.Run("Admin_ServiceAccount_403", func(t *testing.T) {
		// Service account has no groups — should get 403.
		resp := get(t, ts.URL+"/v1/admin", accessToken)
		assertStatus(t, resp, 403)
	})

	t.Run("Whoami_ForwardAuthHeaders", func(t *testing.T) {
		// Simulate what Traefik sends after forward auth.
		req, _ := http.NewRequest("GET", ts.URL+"/v1/whoami", nil)
		req.Header.Set("X-authentik-username", "akadmin")
		req.Header.Set("X-authentik-email", "mafiafire@proton.me")
		req.Header.Set("X-authentik-name", "Admin User")
		req.Header.Set("X-authentik-uid", "abc123")
		req.Header.Set("X-authentik-groups", "authentik Admins|admins|developers")
		req.Header.Set("X-authentik-entitlements", "")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()
		assertStatus(t, resp, 200)

		body := readBody(t, resp)
		t.Logf("whoami (forward auth): %s", body)

		var envelope struct {
			Data api.AuthentikUser `json:"data"`
		}
		if err := json.Unmarshal([]byte(body), &envelope); err != nil {
			t.Fatalf("parse: %v", err)
		}
		if envelope.Data.Username != "akadmin" {
			t.Errorf("expected username akadmin, got %s", envelope.Data.Username)
		}
		if !envelope.Data.HasGroup("admins") {
			t.Error("expected admins group")
		}
	})

	t.Run("Admin_ForwardAuth_Admins_200", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/v1/admin", nil)
		req.Header.Set("X-authentik-username", "akadmin")
		req.Header.Set("X-authentik-email", "mafiafire@proton.me")
		req.Header.Set("X-authentik-name", "Admin User")
		req.Header.Set("X-authentik-uid", "abc123")
		req.Header.Set("X-authentik-groups", "authentik Admins|admins|developers")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()
		assertStatus(t, resp, 200)
		t.Logf("admin (forward auth): %s", readBody(t, resp))
	})

	t.Run("InvalidJWT_FailOpen", func(t *testing.T) {
		// Invalid token on a public endpoint — should still work (permissive).
		resp := get(t, ts.URL+"/v1/public", "not-a-real-token")
		assertStatus(t, resp, 200)
	})

	t.Run("InvalidJWT_Protected_401", func(t *testing.T) {
		// Invalid token on a protected endpoint — no user extracted, RequireAuth returns 401.
		resp := get(t, ts.URL+"/v1/whoami", "not-a-real-token")
		assertStatus(t, resp, 401)
	})
}

func get(t *testing.T, url, bearerToken string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest("GET", url, nil)
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("want status %d, got %d: %s", want, resp.StatusCode, string(b))
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// TestOIDCDiscovery validates that the OIDC discovery endpoint is reachable.
func TestOIDCDiscovery(t *testing.T) {
	if os.Getenv("AUTHENTIK_INTEGRATION") != "1" {
		t.Skip("set AUTHENTIK_INTEGRATION=1 to run live Authentik tests")
	}

	issuer := envOr("AUTHENTIK_ISSUER", "https://auth.lthn.io/application/o/core-api/")
	disc := strings.TrimSuffix(issuer, "/") + "/.well-known/openid-configuration"

	resp, err := http.Get(disc)
	if err != nil {
		t.Fatalf("discovery request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("discovery status: %d", resp.StatusCode)
	}

	var config map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Verify essential fields.
	for _, field := range []string{"issuer", "token_endpoint", "jwks_uri", "authorization_endpoint"} {
		if config[field] == nil {
			t.Errorf("missing field: %s", field)
		}
	}

	if config["issuer"] != issuer {
		t.Errorf("issuer mismatch: got %v, want %s", config["issuer"], issuer)
	}

	// Verify grant types include client_credentials.
	grants, ok := config["grant_types_supported"].([]any)
	if !ok {
		t.Fatal("missing grant_types_supported")
	}
	found := false
	for _, g := range grants {
		if g == "client_credentials" {
			found = true
			break
		}
	}
	if !found {
		t.Error("client_credentials grant not supported")
	}

	fmt.Printf("  OIDC discovery OK — issuer: %s\n", config["issuer"])
	fmt.Printf("  Token endpoint: %s\n", config["token_endpoint"])
	fmt.Printf("  JWKS URI: %s\n", config["jwks_uri"])
}
