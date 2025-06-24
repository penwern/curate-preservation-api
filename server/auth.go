// Package server – auth middleware for Cells JWT
package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/penwern/curate-preservation-api/pkg/logger"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const jwtContextKey contextKey = "jwt"

// JWKSCache keeps one remote JWK set per issuer.
type JWKSCache struct {
	m sync.Map // map[string]jwk.Set
}

// Get retrieves a JWK set for the given issuer, fetching it if not cached
func (c *JWKSCache) Get(iss string) (jwk.Set, error) {
	if v, ok := c.m.Load(iss); ok {
		return v.(jwk.Set), nil
	}
	// Dex exposes keys at <issuer>/keys
	ks, err := jwk.Fetch(context.Background(), iss+"/keys",
		jwk.WithHTTPClient(&http.Client{Timeout: 5 * time.Second}))
	if err != nil {
		return nil, err
	}
	c.m.Store(iss, ks)
	return ks, nil
}

var jwksCache JWKSCache

// Auth is a chi-style middleware that validates Cells JWTs.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get("Authorization")
		if !strings.HasPrefix(hdr, "Bearer ") {
			logger.Error("Auth failed: missing bearer token")
			respondWithError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		raw := strings.TrimPrefix(hdr, "Bearer ")
		logger.Debug("Auth: received token length: %d", len(raw))
		logger.Debug("Auth: token starts with: %s", raw[:minInt(50, len(raw))])

		// Check if token has the expected JWT structure (header.payload.signature)
		parts := strings.Split(raw, ".")
		logger.Debug("Auth: token has %d parts (expected 3 for JWT)", len(parts))
		if len(parts) != 3 {
			logger.Error("Auth failed: token doesn't have 3 parts (header.payload.signature), has %d parts", len(parts))
			respondWithError(w, http.StatusUnauthorized, "invalid token format")
			return
		}

		// First parse the token without verification to get the issuer
		tok, err := jwt.ParseString(raw, jwt.WithVerify(false), jwt.WithValidate(false))
		if err != nil {
			logger.Error("Auth failed: invalid token format: %v", err)
			logger.Error("Auth failed: problematic token: %s", raw)
			respondWithError(w, http.StatusUnauthorized, "invalid token format")
			return
		}

		// Get the issuer to fetch the correct key set
		iss := tok.Issuer()
		if iss == "" {
			logger.Error("Auth failed: token missing issuer")
			respondWithError(w, http.StatusUnauthorized, "token missing issuer")
			return
		}

		logger.Debug("Auth: token issuer: %s", iss)

		// Get the key set for this issuer
		keySet, err := jwksCache.Get(iss)
		if err != nil {
			logger.Error("Auth failed: failed to fetch keys from %s: %v", iss, err)
			respondWithError(w, http.StatusUnauthorized, "failed to fetch keys")
			return
		}

		logger.Debug("Auth: successfully fetched key set from issuer")

		// Now verify the token with the key set
		verifiedTok, err := jwt.ParseString(raw, jwt.WithKeySet(keySet))
		if err != nil {
			logger.Error("Auth failed: token verification failed: %v", err)
			respondWithError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		logger.Debug("Auth: token successfully verified")

		// -- optional role check --
		if roles, _ := verifiedTok.Get("roles"); roles != nil {
			logger.Debug("Auth: checking roles: %v", roles)
			if !contains(roles.([]any), "admin") {
				logger.Error("Auth failed: insufficient role (need admin, got: %v)", roles)
				respondWithError(w, http.StatusForbidden, "insufficient role")
				return
			}
			logger.Debug("Auth: admin role verified")
		} else {
			logger.Warn("Auth: no roles claim found in token, allowing access")
		}

		logger.Debug("Auth: authentication successful")
		// token is fine – make it available downstream
		ctx := context.WithValue(r.Context(), jwtContextKey, verifiedTok)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func contains(arr []any, val string) bool {
	for _, v := range arr {
		if v.(string) == val {
			return true
		}
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
