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
	"tas/internal/database"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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
	if err := a.authenticate(c, false); err != nil {
		return err
	}
	return c.Next()
}

func (a *API) requireAdmin(c *fiber.Ctx) error {
	if err := a.authenticate(c, true); err != nil {
		return err
	}
	return c.Next()
}

func (a *API) authenticate(c *fiber.Ctx, requireRole bool) error {
	if _, ok := c.Locals("member").(*database.Member); ok {
		return nil
	}
	if a.CFG.Auth.DevAllowAdmin {
		member, err := a.ensureDevAdmin()
		if err != nil {
			return err
		}
		c.Locals("member", member)
		return nil
	}
	if a.CFG.Auth.Issuer == "" || a.CFG.Auth.ClientID == "" {
		return fail(fiber.StatusServiceUnavailable, "FusionAuth is not configured")
	}

	header := c.Get(fiber.HeaderAuthorization)
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" || token == header {
		return fail(fiber.StatusUnauthorized, "missing bearer token")
	}
	claims, err := a.validateJWT(c.Context(), token, requireRole)
	if err != nil {
		return fail(fiber.StatusUnauthorized, err.Error())
	}
	member, err := a.ensureMemberFromClaims(claims)
	if err != nil {
		return err
	}
	c.Locals("claims", claims)
	c.Locals("member", member)
	return nil
}

func (a *API) validateJWT(ctx context.Context, token string, requireRole bool) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid token header")
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, errors.New("invalid token header")
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported token algorithm %q", header.Alg)
	}

	key, err := a.publicKeyForKid(ctx, header.Kid)
	if err != nil {
		return nil, err
	}
	signed := []byte(parts[0] + "." + parts[1])
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid token signature")
	}
	hash := sha256.Sum256(signed)
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], signature); err != nil {
		return nil, errors.New("token signature verification failed")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token claims")
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, errors.New("invalid token claims")
	}
	if err := a.validateClaims(claims, requireRole); err != nil {
		return nil, err
	}
	return claims, nil
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

func (a *API) ensureDevAdmin() (*database.Member, error) {
	now := time.Now()
	member := database.Member{
		FusionAuthUserID: "dev-admin",
		Email:            "dev-admin@localhost",
		Name:             "Development Admin",
		Status:           database.MemberStatusApproved,
		LastLoginAt:      &now,
	}
	if err := a.DB.Where(database.Member{FusionAuthUserID: member.FusionAuthUserID}).FirstOrCreate(&member).Error; err != nil {
		return nil, fail(fiber.StatusInternalServerError, "could not create development admin")
	}
	var adminRole database.Role
	if err := preloadRolePermissions(a.DB).Where("name = ?", "admin").First(&adminRole).Error; err == nil {
		if err := a.DB.Model(&member).Association("Roles").Replace(&adminRole); err != nil {
			return nil, fail(fiber.StatusInternalServerError, "could not assign development admin role")
		}
	}
	if err := preloadMemberPermissions(a.DB).First(&member, member.ID).Error; err != nil {
		return nil, fail(fiber.StatusInternalServerError, "could not load development admin")
	}
	return &member, nil
}

func (a *API) ensureMemberFromClaims(claims map[string]any) (*database.Member, error) {
	subject, _ := claims["sub"].(string)
	if strings.TrimSpace(subject) == "" {
		return nil, fail(fiber.StatusUnauthorized, "token subject is missing")
	}

	now := time.Now()
	email := stringClaim(claims, "email", "preferred_username")
	name := stringClaim(claims, "name", "full_name", "preferred_username", "email")
	member := database.Member{}
	err := preloadMemberPermissions(a.DB).Where("fusion_auth_user_id = ?", subject).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		member = database.Member{
			FusionAuthUserID: subject,
			Email:            email,
			Name:             name,
			Status:           database.MemberStatusPending,
			LastLoginAt:      &now,
		}
		if err := a.DB.Create(&member).Error; err != nil {
			return nil, fail(fiber.StatusInternalServerError, "could not create member")
		}
		if err := preloadMemberPermissions(a.DB).First(&member, member.ID).Error; err != nil {
			return nil, fail(fiber.StatusInternalServerError, "could not load member")
		}
		return &member, nil
	}
	if err != nil {
		return nil, fail(fiber.StatusInternalServerError, "could not load member")
	}

	updates := map[string]any{"last_login_at": &now}
	if email != "" {
		updates["email"] = email
	}
	if name != "" {
		updates["name"] = name
	}
	if err := a.DB.Model(&member).Updates(updates).Error; err != nil {
		return nil, fail(fiber.StatusInternalServerError, "could not update member")
	}
	if err := preloadMemberPermissions(a.DB).First(&member, member.ID).Error; err != nil {
		return nil, fail(fiber.StatusInternalServerError, "could not reload member")
	}
	return &member, nil
}

func preloadRolePermissions(db *gorm.DB) *gorm.DB {
	return db.Preload("Permissions")
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

func stringClaim(claims map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := claims[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
