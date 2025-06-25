// Package server – auth middleware for Pydio Cells OIDC
package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/penwern/curate-preservation-api/pkg/logger"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const userInfoContextKey contextKey = "userInfo"

// UserRole represents a user role in Pydio Cells
type UserRole struct {
	Label string `json:"Label"`
	UUID  string `json:"Uuid"`
}

// UserInfo represents the combined user information from OIDC and Pydio
type UserInfo struct {
	// OIDC fields
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	PreferredName string `json:"preferred_username"`

	// Pydio fields
	Login     string     `json:"Login"`
	UUID      string     `json:"Uuid"`
	GroupPath string     `json:"GroupPath"`
	Roles     []UserRole `json:"Roles"`

	// Additional fields that might be present
	Attributes map[string]interface{} `json:"Attributes,omitempty"`
}

// PydioUserQuery represents the query structure for Pydio user info
type PydioUserQuery struct {
	Queries []PydioQuery `json:"Queries"`
}

// PydioQuery represents the query structure for Pydio user info
type PydioQuery struct {
	UUID string `json:"Uuid"`
}

// PydioUserResponse represents the response from Pydio user info endpoint
type PydioUserResponse struct {
	Users []UserInfo `json:"Users"`
}

// CacheEntry represents a cached user info with expiration
type CacheEntry struct {
	UserInfo  UserInfo
	ExpiresAt time.Time
}

// UserInfoCache provides thread-safe caching of user information
type UserInfoCache struct {
	cache map[string]CacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewUserInfoCache creates a new user info cache with the specified TTL
func NewUserInfoCache(ttl time.Duration) *UserInfoCache {
	cache := &UserInfoCache{
		cache: make(map[string]CacheEntry),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves user info from cache if valid
func (c *UserInfoCache) Get(token string) (UserInfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[token]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return UserInfo{}, false
	}

	return entry.UserInfo, true
}

// Set stores user info in cache with expiration
func (c *UserInfoCache) Set(token string, userInfo UserInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[token] = CacheEntry{
		UserInfo:  userInfo,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// cleanup removes expired entries from cache
func (c *UserInfoCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for token, entry := range c.cache {
			if now.After(entry.ExpiresAt) {
				delete(c.cache, token)
			}
		}
		c.mutex.Unlock()
	}
}

// Global cache instance
var userInfoCache = NewUserInfoCache(5 * time.Minute)

// parseIPOrCIDR parses an IP address or CIDR range
func parseIPOrCIDR(ipStr string) (*net.IPNet, error) {
	// Check if it's a CIDR range
	if strings.Contains(ipStr, "/") {
		_, ipNet, err := net.ParseCIDR(ipStr)
		return ipNet, err
	}

	// It's a single IP address, convert to /32 or /128
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	var mask net.IPMask
	if ip.To4() != nil {
		// IPv4
		mask = net.CIDRMask(32, 32)
	} else {
		// IPv6
		mask = net.CIDRMask(128, 128)
	}

	return &net.IPNet{IP: ip, Mask: mask}, nil
}

// isIPTrusted checks if the given IP address is in the trusted IPs list
func isIPTrusted(clientIP string, trustedIPs []string) bool {
	if len(trustedIPs) == 0 {
		return false
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		logger.Debug("Auth: failed to parse client IP: %s", clientIP)
		return false
	}

	for _, trustedIP := range trustedIPs {
		ipNet, err := parseIPOrCIDR(trustedIP)
		if err != nil {
			logger.Warn("Auth: failed to parse trusted IP/CIDR '%s': %v", trustedIP, err)
			continue
		}

		if ipNet.Contains(ip) {
			logger.Debug("Auth: client IP %s matches trusted IP/CIDR %s", clientIP, trustedIP)
			return true
		}
	}

	logger.Debug("Auth: client IP %s not found in trusted IPs", clientIP)
	return false
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (most common with proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// getConfig returns configuration URLs for the specified site domain
func getConfig(siteDomain string) (string, string, string) {
	if siteDomain == "" {
		siteDomain = "https://localhost:8080"
	}

	userinfoURL := fmt.Sprintf("%s/oidc/userinfo", siteDomain)
	pydioUserInfoURL := fmt.Sprintf("%s/a/user", siteDomain)

	return siteDomain, userinfoURL, pydioUserInfoURL
}

// validateTokenAndGetUserInfo validates token and retrieves user information using specified domain
func validateTokenAndGetUserInfo(token string, siteDomain string, allowInsecureTLS bool) (*UserInfo, error) {
	logger.Debug("Auth: validating token for domain: %s", siteDomain)

	// Check cache first
	if userInfo, found := userInfoCache.Get(token); found {
		logger.Debug("Auth: using cached user info for user: %s", userInfo.Sub)
		return &userInfo, nil
	}

	logger.Debug("Auth: no cached user info found, fetching from APIs")

	_, userinfoURL, pydioUserInfoURL := getConfig(siteDomain)
	logger.Debug("Auth: using OIDC userinfo URL: %s", userinfoURL)
	logger.Debug("Auth: using Pydio user info URL: %s", pydioUserInfoURL)

	// Step 1: Validate token with OIDC userinfo endpoint
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			// #nosec G402 -- InsecureSkipVerify is configurable via AllowInsecureTLS for development/testing environments
			TLSClientConfig: &tls.Config{InsecureSkipVerify: allowInsecureTLS},
		},
	}

	logger.Debug("Auth: making OIDC userinfo request")
	req, err := http.NewRequest("GET", userinfoURL, nil)
	if err != nil {
		logger.Error("Auth: failed to create userinfo request: %v", err)
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Auth: userinfo request failed: %v", err)
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Auth: failed to close userinfo response body: %v", err)
		}
	}()

	logger.Debug("Auth: OIDC userinfo response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		logger.Error("Auth: userinfo request failed with status: %d", resp.StatusCode)
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var oidcUserInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&oidcUserInfo); err != nil {
		logger.Error("Auth: failed to decode userinfo response: %v", err)
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	logger.Debug("Auth: OIDC user info retrieved for user: %s (email: %s, name: %s)", oidcUserInfo.Sub, oidcUserInfo.Email, oidcUserInfo.Name)

	// Step 2: Get detailed user info from Pydio Cells
	if oidcUserInfo.Sub == "" {
		logger.Error("Auth: user UUID not found in OIDC user info")
		return nil, fmt.Errorf("user UUID not found in OIDC user info")
	}

	logger.Debug("Auth: making Pydio user info request for UUID: %s", oidcUserInfo.Sub)

	pydioQuery := PydioUserQuery{
		Queries: []PydioQuery{{UUID: oidcUserInfo.Sub}},
	}

	queryBytes, err := json.Marshal(pydioQuery)
	if err != nil {
		logger.Error("Auth: failed to marshal Pydio query: %v", err)
		return nil, fmt.Errorf("failed to marshal Pydio query: %w", err)
	}

	logger.Debug("Auth: Pydio query payload: %s", string(queryBytes))

	pydioReq, err := http.NewRequest("POST", pydioUserInfoURL, bytes.NewBuffer(queryBytes))
	if err != nil {
		logger.Error("Auth: failed to create Pydio request: %v", err)
		return nil, fmt.Errorf("failed to create Pydio request: %w", err)
	}
	pydioReq.Header.Set("Authorization", "Bearer "+token)
	pydioReq.Header.Set("Content-Type", "application/json")

	logger.Debug("Auth: making Pydio user info request")
	logger.Debug("Auth: Pydio request headers: %v", pydioReq.Header)

	pydioResp, err := client.Do(pydioReq)
	if err != nil {
		logger.Error("Auth: pydio request failed: %v", err)
		return nil, fmt.Errorf("pydio request failed: %w", err)
	}
	defer func() {
		if err := pydioResp.Body.Close(); err != nil {
			logger.Error("Auth: failed to close Pydio response body: %v", err)
		}
	}()

	logger.Debug("Auth: Pydio user info response status: %d", pydioResp.StatusCode)

	if pydioResp.StatusCode != http.StatusOK {
		logger.Error("Auth: pydio request failed with status: %d", pydioResp.StatusCode)
		return nil, fmt.Errorf("pydio request failed with status: %d", pydioResp.StatusCode)
	}

	var pydioUserInfo PydioUserResponse
	if err := json.NewDecoder(pydioResp.Body).Decode(&pydioUserInfo); err != nil {
		logger.Error("Auth: failed to decode Pydio response: %v", err)
		return nil, fmt.Errorf("failed to decode Pydio response: %w", err)
	}

	logger.Debug("Auth: Pydio user info retrieved, found %d users", len(pydioUserInfo.Users))

	// Combine user info
	if len(pydioUserInfo.Users) == 0 {
		logger.Error("Auth: user not found in Pydio Cells")
		return nil, fmt.Errorf("user not found in Pydio Cells")
	}

	userInfo := pydioUserInfo.Users[0]
	logger.Debug("Auth: Pydio user details - Login: %s, UUID: %s, GroupPath: %s", userInfo.Login, userInfo.UUID, userInfo.GroupPath)

	// Merge OIDC info
	userInfo.Sub = oidcUserInfo.Sub
	userInfo.Email = oidcUserInfo.Email
	userInfo.Name = oidcUserInfo.Name
	userInfo.PreferredName = oidcUserInfo.PreferredName

	logger.Debug("Auth: combined user info - storing in cache")
	// Cache the result
	userInfoCache.Set(token, userInfo)

	logger.Debug("Auth: user validation complete for: %s", userInfo.Sub)
	return &userInfo, nil
}

// TokenRequired creates a middleware that validates tokens using specified domain
func TokenRequired(siteDomain string, trustedIPs []string, allowInsecureTLS bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug("Auth: starting authentication for %s %s", r.Method, r.URL.Path)
			logger.Debug("Auth: site domain: '%s'", siteDomain)

			// Check if the client IP is trusted
			clientIP := getClientIP(r)
			logger.Debug("Auth: client IP: %s", clientIP)

			if isIPTrusted(clientIP, trustedIPs) {
				logger.Info("Auth: allowing trusted IP %s to bypass authentication", clientIP)
				// Create a minimal user info for trusted IPs
				trustedUserInfo := &UserInfo{
					Sub:           "trusted-ip:" + clientIP,
					Email:         "trusted@internal",
					Name:          "Trusted Internal User",
					PreferredName: "trusted",
					Login:         "trusted-ip",
					UUID:          "trusted-ip:" + clientIP,
					GroupPath:     "/trusted",
					Roles:         []UserRole{{Label: "trusted", UUID: "trusted-role"}},
				}

				// Add trusted user info to request context
				ctx := context.WithValue(r.Context(), userInfoContextKey, trustedUserInfo)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Error("Auth failed: missing authorization header")
				respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			logger.Debug("Auth: authorization header present (length: %d)", len(authHeader))

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				logger.Error("Auth failed: invalid Authorization header format: '%s'", authHeader)
				respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format")
				return
			}

			token := parts[1]
			logger.Debug("Auth: extracted bearer token (length: %d)", len(token))

			// Validate token and get user info
			userInfo, err := validateTokenAndGetUserInfo(token, siteDomain, allowInsecureTLS)
			if err != nil {
				logger.Error("Auth failed: %v", err)
				respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			logger.Debug("Auth: token validation successful for user: %s (login: %s)", userInfo.Sub, userInfo.Login)
			logger.Debug("Auth: authentication successful for user: %s, proceeding to handler", userInfo.Sub)

			// Add user info to request context
			ctx := context.WithValue(r.Context(), userInfoContextKey, userInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Auth creates middleware that validates tokens using specified domain
func Auth(siteDomain string, trustedIPs []string, allowInsecureTLS bool) func(http.Handler) http.Handler {
	return TokenRequired(siteDomain, trustedIPs, allowInsecureTLS)
}

// GetUserInfo retrieves user info from request context
func GetUserInfo(r *http.Request) *UserInfo {
	if userInfo, ok := r.Context().Value(userInfoContextKey).(*UserInfo); ok {
		return userInfo
	}
	return nil
}
