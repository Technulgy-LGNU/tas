package web

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type jwkSet struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

func (a *API) authConfig(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"issuer":        a.CFG.Auth.Issuer,
		"client_id":     a.CFG.Auth.ClientID,
		"required_role": a.CFG.Auth.RequiredRole,
		"dev_allow":     a.CFG.Auth.DevAllowAdmin,
	})
}

func (a *API) requireAuthenticated(c *fiber.Ctx) error {
	return a.requireToken(c, false)
}

func (a *API) requireAdmin(c *fiber.Ctx) error {
	return a.requireToken(c, true)
}

func (a *API) requireToken(c *fiber.Ctx, requireRole bool) error {
	if a.CFG.Auth.DevAllowAdmin {
		return c.Next()
	}
	if a.CFG.Auth.Issuer == "" || a.CFG.Auth.ClientID == "" {
		return fail(fiber.StatusServiceUnavailable, "FusionAuth is not configured")
	}

	header := c.Get(fiber.HeaderAuthorization)
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" || token == header {
		return fail(fiber.StatusUnauthorized, "missing bearer token")
	}
	if err := a.validateJWT(c.Context(), token, requireRole); err != nil {
		return fail(fiber.StatusUnauthorized, err.Error())
	}
	return c.Next()
}

func (a *API) validateJWT(ctx context.Context, token string, requireRole bool) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("invalid token")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return errors.New("invalid token header")
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return errors.New("invalid token header")
	}
	if header.Alg != "RS256" {
		return fmt.Errorf("unsupported token algorithm %q", header.Alg)
	}

	key, err := a.publicKeyForKid(ctx, header.Kid)
	if err != nil {
		return err
	}
	signed := []byte(parts[0] + "." + parts[1])
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return errors.New("invalid token signature")
	}
	hash := sha256.Sum256(signed)
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], signature); err != nil {
		return errors.New("token signature verification failed")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("invalid token claims")
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return errors.New("invalid token claims")
	}
	return a.validateClaims(claims, requireRole)
}

func (a *API) validateClaims(claims map[string]any, requireRole bool) error {
	if iss, _ := claims["iss"].(string); iss != a.CFG.Auth.Issuer {
		return errors.New("invalid token issuer")
	}
	if !claimAudienceContains(claims["aud"], a.CFG.Auth.ClientID) {
		return errors.New("invalid token audience")
	}
	exp, ok := numberClaim(claims["exp"])
	if !ok || time.Unix(int64(exp), 0).Before(time.Now().Add(-30*time.Second)) {
		return errors.New("token is expired")
	}
	if requireRole && a.CFG.Auth.RequiredRole != "" && !rolesContain(claims["roles"], a.CFG.Auth.RequiredRole) {
		return errors.New("required admin role missing")
	}
	return nil
}

func (a *API) publicKeyForKid(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	a.jwksMu.Lock()
	if a.jwksKeys != nil && time.Since(a.jwksFetched) < 15*time.Minute {
		if key := a.jwksKeys[kid]; key != nil {
			a.jwksMu.Unlock()
			return key, nil
		}
	}
	a.jwksMu.Unlock()

	if err := a.refreshJWKS(ctx); err != nil {
		return nil, err
	}

	a.jwksMu.Lock()
	defer a.jwksMu.Unlock()
	if key := a.jwksKeys[kid]; key != nil {
		return key, nil
	}
	return nil, errors.New("token signing key not found")
}

func (a *API) refreshJWKS(ctx context.Context) error {
	urls := []string{
		a.CFG.Auth.Issuer + "/.well-known/jwks.json",
		a.CFG.Auth.Issuer + "/oauth2/jwks",
	}

	var lastErr error
	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("jwks fetch status %d", resp.StatusCode)
			continue
		}
		var set jwkSet
		if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
			lastErr = err
			continue
		}
		keys := make(map[string]*rsa.PublicKey, len(set.Keys))
		for _, rawKey := range set.Keys {
			key, err := rawKey.rsaPublicKey()
			if err == nil && rawKey.Kid != "" {
				keys[rawKey.Kid] = key
			}
		}
		if len(keys) == 0 {
			lastErr = errors.New("jwks contained no usable RSA keys")
			continue
		}
		a.jwksMu.Lock()
		a.jwksKeys = keys
		a.jwksFetched = time.Now()
		a.jwksMu.Unlock()
		return nil
	}
	return lastErr
}

func (k jwkKey) rsaPublicKey() (*rsa.PublicKey, error) {
	if k.Kty != "RSA" {
		return nil, errors.New("not an RSA key")
	}
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func claimAudienceContains(value any, want string) bool {
	switch typed := value.(type) {
	case string:
		return typed == want
	case []any:
		for _, item := range typed {
			if item == want {
				return true
			}
		}
	}
	return false
}

func rolesContain(value any, want string) bool {
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if item == want {
				return true
			}
		}
	case []string:
		for _, item := range typed {
			if item == want {
				return true
			}
		}
	}
	return false
}

func numberClaim(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case json.Number:
		v, err := typed.Float64()
		return v, err == nil
	default:
		return 0, false
	}
}
