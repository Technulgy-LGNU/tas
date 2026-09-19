package web

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"tas/backend/config"
)

const (
	sessionCookie   = "tas_session"
	loginCookie     = "tas_login"
	sessionLifetime = 8 * time.Hour
	loginLifetime   = 10 * time.Minute
	maxSessions     = 10000
)

var errUnauthorized = errors.New("authentication required")

type User struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

type loginAttempt struct {
	state, verifier, returnTo string
	expires                   time.Time
}

type authSession struct {
	mu                        sync.Mutex // Serialize refreshes, including one-time refresh tokens.
	accessToken, refreshToken string
	tokenExpires, expires     time.Time
	user                      User
	revoked                   bool
}

// Auth keeps credentials server-side. This in-memory store is for a single instance;
// restarting the server signs everyone out.
type Auth struct {
	cfg      config.AuthConfig
	client   *http.Client
	secure   bool
	frontend *url.URL
	mu       sync.Mutex
	logins   map[string]loginAttempt
	sessions map[string]*authSession
}

func NewAuth(cfg config.AuthConfig) (*Auth, error) {
	for name, raw := range map[string]string{"fusionauth_url": cfg.FusionAuthURL, "oauth_redirect_uri": cfg.OAuthRedirectURI, "frontend_url": cfg.FrontendURL} {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
			return nil, fmt.Errorf("auth.%s must be an absolute HTTP(S) URL", name)
		}
		if name != "frontend_url" && (u.RawQuery != "" || u.Fragment != "") {
			return nil, fmt.Errorf("auth.%s must not contain a query or fragment", name)
		}
		if name == "oauth_redirect_uri" && u.Path != "/auth/callback" {
			return nil, errors.New("auth.oauth_redirect_uri must use /auth/callback")
		}
	}
	if cfg.FusionAuthClientId == "" || cfg.FusionAuthSecret == "" || cfg.FusionAuthTenantId == "" {
		return nil, errors.New("auth requires fusionauth_client_id, fusionauth_client_secret and fusionauth_tenant_id")
	}
	cfg.FusionAuthURL = strings.TrimRight(cfg.FusionAuthURL, "/")
	frontend, _ := url.Parse(cfg.FrontendURL)
	callback, _ := url.Parse(cfg.OAuthRedirectURI)
	// Cookies must reach both the callback and the frontend (via Vite in development).
	if frontend.Hostname() != callback.Hostname() || frontend.Scheme != callback.Scheme {
		return nil, errors.New("auth.frontend_url and oauth_redirect_uri must share a hostname and scheme")
	}
	return &Auth{cfg: cfg, frontend: frontend, secure: callback.Scheme == "https",
		client: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		logins: make(map[string]loginAttempt), sessions: make(map[string]*authSession)}, nil
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func (a *Auth) cookie(c fiber.Ctx, name, value string, lifetime time.Duration) {
	maxAge := int(lifetime.Seconds())
	if lifetime < 0 {
		maxAge = -1
	}
	c.Cookie(&fiber.Cookie{Name: name, Value: value, Path: "/", HTTPOnly: true,
		Secure: a.secure, SameSite: "Lax", MaxAge: maxAge, Expires: time.Now().Add(lifetime)})
}

func (a *Auth) page(path string) string {
	u := *a.frontend
	u.RawQuery = ""
	u.Fragment = path
	u.RawFragment = ""
	return u.String()
}

func safeReturnTo(path string) string {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\\\r\n") || strings.HasPrefix(path, "/login") {
		return "/"
	}
	return path
}

// Called with a.mu held. Expiry is absolute and never extended by refreshes.
func (a *Auth) prune() {
	now := time.Now()
	for id, login := range a.logins {
		if !now.Before(login.expires) {
			delete(a.logins, id)
		}
	}
	for id, session := range a.sessions {
		if !now.Before(session.expires) {
			delete(a.sessions, id)
		}
	}
}

func (a *Auth) Login(c fiber.Ctx) error {
	id, state, verifier := randomToken(), randomToken(), randomToken()
	a.mu.Lock()
	a.prune()
	delete(a.logins, c.Cookies(loginCookie))
	if len(a.logins) >= maxSessions {
		a.mu.Unlock()
		return c.SendStatus(fiber.StatusTooManyRequests)
	}
	a.logins[id] = loginAttempt{state: state, verifier: verifier, returnTo: safeReturnTo(c.Query("returnTo")), expires: time.Now().Add(loginLifetime)}
	a.mu.Unlock()
	a.cookie(c, loginCookie, id, loginLifetime)
	hash := sha256.Sum256([]byte(verifier))
	q := url.Values{"client_id": {a.cfg.FusionAuthClientId}, "tenantId": {a.cfg.FusionAuthTenantId},
		"redirect_uri": {a.cfg.OAuthRedirectURI}, "response_type": {"code"},
		"scope": {"openid email profile offline_access"}, "state": {state},
		"code_challenge": {base64.RawURLEncoding.EncodeToString(hash[:])}, "code_challenge_method": {"S256"}}
	return c.Redirect().To(a.cfg.FusionAuthURL + "/oauth2/authorize?" + q.Encode())
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (a *Auth) post(ctx context.Context, path string, form url.Values, result any) error {
	form.Set("client_id", a.cfg.FusionAuthClientId)
	form.Set("client_secret", a.cfg.FusionAuthSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.FusionAuthURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 400 || res.StatusCode == 401 || res.StatusCode == 403 {
		return errUnauthorized
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("FusionAuth returned HTTP %d", res.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(result)
}

// Introspection verifies the token with FusionAuth; never trust locally decoded JWT claims.
func (a *Auth) inspect(ctx context.Context, token string) (User, time.Time, error) {
	var claims struct {
		Active        bool     `json:"active"`
		ApplicationID string   `json:"applicationId"`
		TenantID      string   `json:"tid"`
		Subject       string   `json:"sub"`
		Email         string   `json:"email"`
		Username      string   `json:"preferred_username"`
		Roles         []string `json:"roles"`
		Exp           int64    `json:"exp"`
	}
	err := a.post(ctx, "/oauth2/introspect", url.Values{"token": {token}, "token_type_hint": {"access_token"}}, &claims)
	if err != nil {
		return User{}, time.Time{}, err
	}
	if !claims.Active || claims.ApplicationID != a.cfg.FusionAuthClientId || claims.TenantID != a.cfg.FusionAuthTenantId || claims.Subject == "" || claims.Exp <= time.Now().Unix() ||
		(a.cfg.RequiredRole != "" && !slices.Contains(claims.Roles, a.cfg.RequiredRole)) {
		return User{}, time.Time{}, errUnauthorized
	}
	return User{ID: claims.Subject, Email: claims.Email, Username: claims.Username, Roles: claims.Roles}, time.Unix(claims.Exp, 0), nil
}

func (a *Auth) Callback(c fiber.Ctx) error {
	a.mu.Lock()
	attempt, ok := a.logins[c.Cookies(loginCookie)]
	delete(a.logins, c.Cookies(loginCookie)) // Each attempt is single-use, including failures.
	a.mu.Unlock()
	a.cookie(c, loginCookie, "", -time.Hour)
	fail := func(reason string) error { return c.Redirect().To(a.page("/login?error=" + reason)) }
	if !ok || !time.Now().Before(attempt.expires) || subtle.ConstantTimeCompare([]byte(c.Query("state")), []byte(attempt.state)) != 1 {
		return fail("invalid_state")
	}
	if c.Query("error") != "" || c.Query("code") == "" {
		return fail("login_failed")
	}
	var tokens tokenResponse
	err := a.post(c.Context(), "/oauth2/token", url.Values{"grant_type": {"authorization_code"}, "code": {c.Query("code")},
		"redirect_uri": {a.cfg.OAuthRedirectURI}, "code_verifier": {attempt.verifier}}, &tokens)
	if err != nil || tokens.AccessToken == "" {
		return fail("login_failed")
	}
	user, expiry, err := a.inspect(c.Context(), tokens.AccessToken)
	if errors.Is(err, errUnauthorized) {
		return fail("access_denied")
	}
	if err != nil {
		return fail("unavailable")
	}
	id := randomToken()
	a.mu.Lock()
	a.prune()
	if len(a.sessions) >= maxSessions {
		a.mu.Unlock()
		return fail("unavailable")
	}
	// Discard any previous session to prevent fixation and stale parallel sessions.
	if old := a.sessions[c.Cookies(sessionCookie)]; old != nil {
		old.mu.Lock()
		old.revoked = true
		old.mu.Unlock()
		delete(a.sessions, c.Cookies(sessionCookie))
	}
	a.sessions[id] = &authSession{accessToken: tokens.AccessToken, refreshToken: tokens.RefreshToken,
		tokenExpires: expiry, expires: time.Now().Add(sessionLifetime), user: user}
	a.mu.Unlock()
	a.cookie(c, sessionCookie, id, sessionLifetime)
	return c.Redirect().To(a.page(attempt.returnTo))
}

// CSRF requires a custom header and an allowed Origin on state-changing requests.
// The frontend uses same-origin requests (including the Vite proxy); no CORS is enabled.
func (a *Auth) CSRF(c fiber.Ctx) error {
	if c.Method() == "GET" || c.Method() == "HEAD" || c.Method() == "OPTIONS" {
		return c.Next()
	}
	callback, _ := url.Parse(a.cfg.OAuthRedirectURI)
	origin := c.Get("Origin")
	if c.Get("X-TAS-CSRF") != "1" || (origin != "" && origin != a.frontend.Scheme+"://"+a.frontend.Host && origin != callback.Scheme+"://"+callback.Host) {
		return c.Status(403).JSON(fiber.Map{"error": "invalid_request_origin"})
	}
	return c.Next()
}

// RequireAuth protects API routes and exposes the verified User through c.Locals("user").
func (a *Auth) RequireAuth(c fiber.Ctx) error {
	id := c.Cookies(sessionCookie)
	a.mu.Lock()
	session := a.sessions[id]
	a.mu.Unlock()
	if session == nil {
		return a.unauthorized(c)
	}
	session.mu.Lock()
	user, err := a.sessionUser(c.Context(), session)
	if errors.Is(err, errUnauthorized) {
		session.revoked = true
	}
	session.mu.Unlock()
	if errors.Is(err, errUnauthorized) {
		a.mu.Lock()
		delete(a.sessions, id)
		a.mu.Unlock()
		return a.unauthorized(c)
	}
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "authentication_unavailable"})
	}
	c.Locals("user", user)
	return c.Next()
}

func (a *Auth) sessionUser(ctx context.Context, s *authSession) (User, error) {
	if s.revoked || !time.Now().Before(s.expires) {
		return User{}, errUnauthorized
	}
	if !time.Now().Before(s.tokenExpires) {
		if s.refreshToken == "" {
			return User{}, errUnauthorized
		}
		var tokens tokenResponse
		err := a.post(ctx, "/oauth2/token", url.Values{"grant_type": {"refresh_token"}, "refresh_token": {s.refreshToken}}, &tokens)
		if err != nil {
			return User{}, err
		}
		if tokens.AccessToken == "" {
			return User{}, errUnauthorized
		}
		s.accessToken = tokens.AccessToken
		if tokens.RefreshToken != "" {
			s.refreshToken = tokens.RefreshToken
		}
	}
	user, expiry, err := a.inspect(ctx, s.accessToken)
	if err != nil {
		return User{}, err
	}
	if user.ID != s.user.ID {
		return User{}, errUnauthorized
	}
	s.user, s.tokenExpires = user, expiry
	return user, nil
}

func (a *Auth) unauthorized(c fiber.Ctx) error {
	a.cookie(c, sessionCookie, "", -time.Hour)
	return c.Status(401).JSON(fiber.Map{"error": "authentication_required"})
}

func (a *Auth) Me(c fiber.Ctx) error { return c.JSON(fiber.Map{"user": c.Locals("user")}) }

// Logout always works locally, even when FusionAuth is unavailable or the session expired.
func (a *Auth) Logout(c fiber.Ctx) error {
	a.mu.Lock()
	if s := a.sessions[c.Cookies(sessionCookie)]; s != nil {
		s.mu.Lock()
		s.revoked = true
		s.accessToken = ""
		s.refreshToken = ""
		s.mu.Unlock()
		delete(a.sessions, c.Cookies(sessionCookie))
	}
	delete(a.logins, c.Cookies(loginCookie))
	a.mu.Unlock()
	a.cookie(c, sessionCookie, "", -time.Hour)
	a.cookie(c, loginCookie, "", -time.Hour)
	q := url.Values{"client_id": {a.cfg.FusionAuthClientId}, "tenantId": {a.cfg.FusionAuthTenantId}}
	return c.JSON(fiber.Map{"logout_url": a.cfg.FusionAuthURL + "/oauth2/logout?" + q.Encode()})
}
