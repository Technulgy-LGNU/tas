package web

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"tas/backend/config"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

type authFixture struct {
	app                          *fiber.App
	auth                         *Auth
	provider                     *httptest.Server
	refreshes                    atomic.Int32
	introspections               atomic.Int32
	active                       bool
	application, tenant, subject string
	roles                        []string
	unavailable                  bool
	refreshDenied                bool
	expectedVerifier             string
}

func newFixture(t *testing.T) *authFixture {
	t.Helper()
	f := &authFixture{active: true, application: "client", tenant: "tenant", subject: "user", roles: []string{"admin"}}
	f.provider = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		if r.Form.Get("client_id") != "client" || r.Form.Get("client_secret") != "secret" {
			t.Error("missing client authentication")
		}
		if f.unavailable {
			w.WriteHeader(503)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth2/token":
			if r.Form.Get("grant_type") == "authorization_code" {
				if r.Form.Get("code_verifier") != f.expectedVerifier || r.Form.Get("redirect_uri") != "http://localhost:2005/auth/callback" || r.Form.Get("code") != "code" {
					t.Error("invalid code exchange")
				}
			} else if r.Form.Get("grant_type") == "refresh_token" {
				f.refreshes.Add(1)
				if r.Form.Get("refresh_token") != "refresh" {
					t.Error("wrong refresh token")
				}
				if f.refreshDenied {
					w.WriteHeader(400)
					return
				}
			} else {
				t.Error("unexpected grant")
			}
			_, _ = fmt.Fprint(w, `{"access_token":"access","refresh_token":"rotated","expires_in":3600}`)
		case "/oauth2/introspect":
			f.introspections.Add(1)
			if r.Form.Get("token") != "access" || r.Form.Get("token_type_hint") != "access_token" {
				t.Error("wrong introspection parameters")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"active": f.active, "applicationId": f.application, "tid": f.tenant,
				"sub": f.subject, "roles": f.roles, "exp": time.Now().Add(time.Hour).Unix(), "email": "admin@example.com"})
		default:
			t.Errorf("unexpected endpoint: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(f.provider.Close)
	cfg := &config.Config{Auth: config.AuthConfig{FusionAuthURL: f.provider.URL, FusionAuthClientId: "client", FusionAuthSecret: "secret",
		FusionAuthTenantId: "tenant", OAuthRedirectURI: "http://localhost:2005/auth/callback", FrontendURL: "http://localhost:2005/#/", RequiredRole: "admin"}}
	var err error
	f.auth, err = NewAuth(cfg.Auth)
	if err != nil {
		t.Fatal(err)
	}
	f.app = fiber.New(fiber.Config{Immutable: true})
	f.app.Get("/auth/login", f.auth.Login)
	f.app.Get("/auth/callback", f.auth.Callback)
	private := f.app.Group("/api/v1", f.auth.CSRF)
	private.Post("/auth/logout", f.auth.Logout)
	private.Use(f.auth.RequireAuth)
	private.Get("/auth/me", f.auth.Me)
	private.Post("/private", func(c fiber.Ctx) error { return c.SendStatus(200) })
	return f
}

func (f *authFixture) request(t *testing.T, method, path string, cookies []*http.Cookie, headers map[string]string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, "http://localhost:2005"+path, nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	res, err := f.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}

func cookieNamed(t *testing.T, res *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("missing cookie %s", name)
	return nil
}

func (f *authFixture) begin(t *testing.T) (*http.Cookie, string) {
	res := f.request(t, "GET", "/auth/login?returnTo=%2Freports%3Fpage%3D2", nil, nil)
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("login: %d", res.StatusCode)
	}
	target, _ := url.Parse(res.Header.Get("Location"))
	q := target.Query()
	if q.Get("client_id") != "client" || q.Get("tenantId") != "tenant" || q.Get("code_challenge_method") != "S256" || !strings.Contains(q.Get("scope"), "offline_access") {
		t.Fatalf("wrong authorization parameters: %v", q)
	}
	cookie := cookieNamed(t, res, loginCookie)
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Secure {
		t.Fatal("incorrect development cookie flags")
	}
	f.expectedVerifier = f.auth.logins[cookie.Value].verifier
	hash := sha256.Sum256([]byte(f.expectedVerifier))
	if q.Get("code_challenge") != base64.RawURLEncoding.EncodeToString(hash[:]) {
		t.Fatal("PKCE challenge mismatch")
	}
	return cookie, q.Get("state")
}

func (f *authFixture) session() *http.Cookie {
	f.auth.sessions["session"] = &authSession{accessToken: "access", refreshToken: "refresh", tokenExpires: time.Now().Add(time.Hour), expires: time.Now().Add(time.Hour), user: User{ID: "user"}}
	return &http.Cookie{Name: sessionCookie, Value: "session"}
}

func TestLoginCallbackAndSession(t *testing.T) {
	f := newFixture(t)
	cookie, state := f.begin(t)
	res := f.request(t, "GET", "/auth/callback?code=code&state="+state, []*http.Cookie{cookie}, nil)
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "http://localhost:2005/#/reports?page=2" {
		t.Fatalf("callback: %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	session := cookieNamed(t, res, sessionCookie)
	if !session.HttpOnly || session.Value == "access" || session.Value == cookie.Value {
		t.Fatal("unsafe session cookie")
	}
	res = f.request(t, "GET", "/api/v1/auth/me", []*http.Cookie{session}, nil)
	var body struct {
		User User `json:"user"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 || body.User.ID != "user" || body.User.Email != "admin@example.com" {
		t.Fatalf("me: %d %+v", res.StatusCode, body)
	}
	res = f.request(t, "GET", "/auth/callback?code=code&state="+state, []*http.Cookie{cookie}, nil)
	if !strings.Contains(res.Header.Get("Location"), "invalid_state") {
		t.Fatal("callback replay accepted")
	}
}

func TestCallbackRejectsInvalidAttempts(t *testing.T) {
	for _, kind := range []string{"wrong_state", "missing_cookie", "expired", "provider_error"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			cookie, state := f.begin(t)
			cookies := []*http.Cookie{cookie}
			path := "/auth/callback?code=code&state=" + state
			switch kind {
			case "wrong_state":
				path += "wrong"
			case "missing_cookie":
				cookies = nil
			case "expired":
				a := f.auth.logins[cookie.Value]
				a.expires = time.Now().Add(-time.Minute)
				f.auth.logins[cookie.Value] = a
			case "provider_error":
				path += "&error=access_denied"
			}
			res := f.request(t, "GET", path, cookies, nil)
			if !strings.Contains(res.Header.Get("Location"), "/login?error=") || len(f.auth.sessions) != 0 || f.introspections.Load() != 0 {
				t.Fatal("invalid callback accepted")
			}
		})
	}
}

func TestMiddlewareFailsClosed(t *testing.T) {
	for _, kind := range []string{"no_cookie", "unknown_cookie", "expired", "inactive", "wrong_app", "wrong_tenant", "missing_role", "different_user", "unavailable", "refresh_denied"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			cookie := f.session()
			cookies := []*http.Cookie{cookie}
			want := 401
			switch kind {
			case "no_cookie":
				cookies = nil
			case "unknown_cookie":
				cookie.Value = "unknown"
			case "expired":
				f.auth.sessions["session"].expires = time.Now().Add(-time.Minute)
			case "inactive":
				f.active = false
			case "wrong_app":
				f.application = "another-app"
			case "wrong_tenant":
				f.tenant = "another-tenant"
			case "missing_role":
				f.roles = nil
			case "different_user":
				f.subject = "other-user"
			case "unavailable":
				f.unavailable = true
				want = 503
			case "refresh_denied":
				f.auth.sessions["session"].tokenExpires = time.Now().Add(-time.Minute)
				f.refreshDenied = true
			}
			res := f.request(t, "GET", "/api/v1/auth/me", cookies, nil)
			if res.StatusCode != want {
				t.Fatalf("got %d want %d", res.StatusCode, want)
			}
			if kind == "unavailable" && f.auth.sessions["session"] == nil {
				t.Fatal("temporary outage destroyed session")
			}
		})
	}
}

func TestRefreshIsSerializedAndRotated(t *testing.T) {
	f := newFixture(t)
	cookie := f.session()
	f.auth.sessions["session"].tokenExpires = time.Now().Add(-time.Minute)
	var wg sync.WaitGroup
	for range 6 {
		wg.Go(func() {
			res := f.request(t, "GET", "/api/v1/auth/me", []*http.Cookie{cookie}, nil)
			if res.StatusCode != 200 {
				t.Errorf("refresh: %d", res.StatusCode)
			}
		})
	}
	wg.Wait()
	if f.refreshes.Load() != 1 || f.auth.sessions["session"].refreshToken != "rotated" {
		t.Fatal("refresh was duplicated or token rotation lost")
	}
}

func TestCSRFAndLogout(t *testing.T) {
	f := newFixture(t)
	cookie := f.session()
	for _, headers := range []map[string]string{nil, {"X-TAS-CSRF": "1", "Origin": "https://evil.example"}} {
		res := f.request(t, "POST", "/api/v1/auth/logout", []*http.Cookie{cookie}, headers)
		if res.StatusCode != 403 || f.auth.sessions["session"] == nil {
			t.Fatal("cross-site logout accepted")
		}
	}
	valid := map[string]string{"X-TAS-CSRF": "1", "Origin": "http://localhost:2005"}
	res := f.request(t, "POST", "/api/v1/private", []*http.Cookie{cookie}, valid)
	if res.StatusCode != 200 {
		t.Fatalf("legitimate mutation blocked: %d", res.StatusCode)
	}
	f.unavailable = true
	res = f.request(t, "POST", "/api/v1/auth/logout", []*http.Cookie{cookie}, valid)
	if res.StatusCode != 200 || len(f.auth.sessions) != 0 || cookieNamed(t, res, sessionCookie).MaxAge != -1 {
		t.Fatal("logout failed during outage")
	}
	res = f.request(t, "GET", "/api/v1/auth/me", []*http.Cookie{cookie}, nil)
	if res.StatusCode != 401 {
		t.Fatal("logged-out session reused")
	}
}

func TestSafeReturnTo(t *testing.T) {
	for _, path := range []string{"https://evil.example", "//evil.example", "/\\evil.example", "/login", "/\r\nLocation: x"} {
		if safeReturnTo(path) != "/" {
			t.Fatalf("unsafe redirect accepted: %q", path)
		}
	}
}

func TestAppProtectionAndConfig(t *testing.T) {
	f := newFixture(t)
	cfg := &config.Config{Auth: f.auth.cfg}
	app, err := NewApp(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/auth/me", "/api/v1/future-route"} {
		res, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != 401 || res.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("unprotected route: %s", path)
		}
	}
	cfg.Auth.FusionAuthSecret = ""
	if _, err := NewApp(cfg, nil); err == nil {
		t.Fatal("missing credentials accepted")
	}
	cfg.Auth = f.auth.cfg
	cfg.Auth.FrontendURL = "http://127.0.0.1:5173"
	if _, err := NewAuth(cfg.Auth); err == nil {
		t.Fatal("incompatible cookie hosts accepted")
	}
	cfg.Auth.FrontendURL = "https://localhost:2005/#/"
	cfg.Auth.OAuthRedirectURI = "https://localhost:2005/auth/callback"
	auth, err := NewAuth(cfg.Auth)
	if err != nil || !auth.secure {
		t.Fatal("HTTPS must enable secure cookies")
	}
}
