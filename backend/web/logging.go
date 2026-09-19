package web

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type requestLogKey struct{}

func requestLogger(ctx context.Context) *slog.Logger {
	if id, ok := ctx.Value(requestLogKey{}).(string); ok {
		return slog.Default().With("request_id", id)
	}
	return slog.Default()
}

// Never include credentials, query parameters, or fragment routes in logs.
func logURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "invalid_url"
	}
	u.User, u.RawQuery, u.Fragment, u.RawFragment = nil, "", "", ""
	u.ForceQuery = false
	return logField(u.String())
}

func logField(raw string) string {
	if len(raw) > 256 {
		return raw[:256] + "…"
	}
	return raw
}

// Keep error classification useful without logging response bodies or URLs
// embedded in net/http errors (which could carry sensitive provider data).
func providerErrorKind(err error) string {
	var dns *net.DNSError
	var certificate *tls.CertificateVerificationError
	var network net.Error
	switch {
	case errors.As(err, &certificate):
		return "tls_certificate"
	case errors.As(err, &dns):
		return "dns"
	case errors.As(err, &network) && network.Timeout():
		return "timeout"
	default:
		return "connection"
	}
}

func requestLogging(c fiber.Ctx) error {
	start := time.Now()
	id := uuid.NewString()
	c.SetContext(context.WithValue(c.Context(), requestLogKey{}, id))
	c.Set("X-Request-ID", id)
	logger := requestLogger(c.Context())
	// Path deliberately excludes OAuth code/state and any other query values.
	attrs := []any{"method", c.Method(), "path", logField(c.Path()),
		"host", logURL("//" + c.Host()), "origin", logURL(c.Get("Origin")),
		"forwarded_proto", logField(c.Get("X-Forwarded-Proto")),
		"forwarded_host", logURL("//" + c.Get("X-Forwarded-Host")),
		"cf_ray", logField(c.Get("CF-Ray"))}
	logger.DebugContext(c.Context(), "http.request_started", attrs...)
	err := c.Next()
	if err != nil {
		err = c.App().ErrorHandler(c, err)
	}
	status := c.Response().StatusCode()
	attrs = append(attrs, "status", status, "duration_ms", time.Since(start).Milliseconds())
	if location := c.GetRespHeader("Location"); location != "" {
		attrs = append(attrs, "redirect", logURL(location))
	}
	level := slog.LevelInfo
	if status >= 500 {
		level = slog.LevelError
	} else if status >= 400 && status != 401 {
		level = slog.LevelWarn
	} else if strings.EqualFold(c.Path(), "/healthcheck") {
		level = slog.LevelDebug
	}
	logger.Log(c.Context(), level, "http.request", attrs...)
	return err
}
