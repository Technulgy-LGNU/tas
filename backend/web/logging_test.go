package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"tas/backend/config"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var out bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &out
}

func TestRequestLogsRedactSecretsAndCorrelate(t *testing.T) {
	out := captureLogs(t)
	app, err := NewApp(&config.Config{Auth: config.AuthConfig{
		FusionAuthURL: "https://auth.example.org", FusionAuthClientId: "client-id-private",
		FusionAuthSecret: "client-secret-private", FusionAuthTenantId: "tenant-id-private",
		FrontendURL: "https://tas.example.org/#/", OAuthRedirectURI: "https://tas.example.org/auth/callback",
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/auth/login?returnTo=/secret-return", "/auth/callback?code=secret-code&state=secret-state", "/api/v1/auth/me?token=secret-query"} {
		req := httptest.NewRequest("GET", "https://tas.example.org"+path, nil)
		req.Header.Set("Cookie", "tas_login=secret-cookie; tas_session=secret-session")
		req.Header.Set("Authorization", "Bearer secret-bearer")
		req.Header.Set("X-Request-ID", "untrusted-request-id")
		req.Header.Set("CF-Ray", "1234-TEST")
		response, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		id := response.Header.Get("X-Request-ID")
		if id == "" || id == "untrusted-request-id" {
			t.Fatal("missing server-generated request ID")
		}
		found := false
		for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
			var entry map[string]any
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				t.Fatal(err)
			}
			if entry["msg"] == "http.request" && entry["request_id"] == id {
				found = true
				if entry["status"] != float64(response.StatusCode) || entry["cf_ray"] != "1234-TEST" {
					t.Fatal("request status/correlation mismatch")
				}
			}
		}
		if !found {
			t.Fatal("response request ID not present in logs")
		}
	}
	for _, secret := range []string{"secret-code", "secret-state", "secret-cookie", "secret-session", "secret-query", "secret-bearer", "secret-return", "client-id-private", "client-secret-private", "tenant-id-private", "code_challenge="} {
		if strings.Contains(out.String(), secret) {
			t.Errorf("sensitive value appeared in logs: %s", secret)
		}
	}
	for _, event := range []string{"auth.login_started", "auth.callback_failed", "auth.session_missing"} {
		if !strings.Contains(out.String(), event) {
			t.Errorf("missing event: %s", event)
		}
	}
}

type diagnosticTransport func(*http.Request) (*http.Response, error)

func (f diagnosticTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestProviderDiagnosticsRedactResponses(t *testing.T) {
	out := captureLogs(t)
	for _, status := range []int{302, 403, 200} {
		a := &Auth{cfg: config.AuthConfig{FusionAuthURL: "https://auth.example.org", FusionAuthSecret: "private-secret"}, client: &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
			Transport: diagnosticTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: status, Header: http.Header{
					"Location":     {"https://user:private-password@access.example.org/login?token=private-token#private-fragment"},
					"Content-Type": {"text/html"},
				}, Body: io.NopCloser(strings.NewReader("<html>private-response-body</html>"))}, nil
			}),
		}}
		var result tokenResponse
		if err := a.post(context.WithValue(context.Background(), requestLogKey{}, "provider-request"), "/oauth2/token", url.Values{"code": {"private-code"}}, &result); err == nil {
			t.Fatal("unexpected successful provider call")
		}
	}
	for _, secret := range []string{"private-secret", "private-password", "private-token", "private-fragment", "private-response-body", "private-code"} {
		if strings.Contains(out.String(), secret) {
			t.Errorf("provider secret appeared in logs: %s", secret)
		}
	}
	for _, expected := range []string{"auth.provider_response", "invalid_json_response", "text/html", "https://access.example.org/login", "provider-request"} {
		if !strings.Contains(out.String(), expected) {
			t.Errorf("missing provider diagnostic: %s", expected)
		}
	}
}

func TestRequestLogUsesFinalErrorStatus(t *testing.T) {
	out := captureLogs(t)
	app := fiber.New()
	app.Use(requestLogging)
	app.Get("/failure", func(c fiber.Ctx) error { return fiber.ErrServiceUnavailable })
	response, err := app.Test(httptest.NewRequest("GET", "/failure", nil))
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != 503 || !strings.Contains(out.String(), `"status":503`) {
		t.Fatal("framework errors must be logged with the final status")
	}
}
