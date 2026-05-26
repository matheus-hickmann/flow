// Package middleware holds chi-compatible HTTP middlewares.
package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ctxKey int

const userIDKey ctxKey = iota
const devTokenPrefix = "dev-"
const tokenTTL = 24 * time.Hour

var jwtSecret string
var devEndpoint string

// Setup initialises the package with runtime configuration.
// Must be called once at startup before any requests are served.
func Setup(secret, dynEndpoint string) {
	jwtSecret = secret
	devEndpoint = dynEndpoint
}

// IssueToken returns a signed token for the given userId.
// In local dev (devEndpoint set) with no secret, falls back to "dev-{userId}".
func IssueToken(userID string) (string, error) {
	if jwtSecret == "" {
		if devEndpoint != "" {
			return devTokenPrefix + userID, nil
		}
		return "", errors.New("JWT_SECRET is required in production")
	}
	exp := time.Now().Add(tokenTTL).Unix()
	payload := base64urlEncode([]byte(userID + "|" + strconv.FormatInt(exp, 10)))
	sig := computeHMAC(payload, jwtSecret)
	return payload + "." + sig, nil
}

// IsDevMode reports whether the service is running in local dev mode.
func IsDevMode() bool {
	return devEndpoint != ""
}

// RequireAuth blocks requests without a valid token and stores the userId in ctx.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := extractUserID(r)
		if userID == "" {
			writeJSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFrom returns the authenticated userId; empty if missing.
func UserIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

func extractUserID(r *http.Request) string {
	// Cookie takes priority (HttpOnly — invisible to JS)
	if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
		if uid := validateToken(cookie.Value); uid != "" {
			return uid
		}
	}
	// Fall back to Authorization header (for API clients / curl testing)
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return validateToken(strings.TrimSpace(header[7:]))
}

func validateToken(token string) string {
	// Dev tokens are only accepted when running locally
	if devEndpoint != "" && strings.HasPrefix(token, devTokenPrefix) {
		return strings.TrimPrefix(token, devTokenPrefix)
	}
	if jwtSecret == "" {
		return ""
	}
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return ""
	}
	payload, sig := parts[0], parts[1]
	if !hmac.Equal([]byte(computeHMAC(payload, jwtSecret)), []byte(sig)) {
		return ""
	}
	decoded, err := base64urlDecode(payload)
	if err != nil {
		return ""
	}
	pipeIdx := strings.LastIndex(string(decoded), "|")
	if pipeIdx < 0 {
		return ""
	}
	userID := string(decoded[:pipeIdx])
	exp, err := strconv.ParseInt(string(decoded[pipeIdx+1:]), 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return ""
	}
	return userID
}

func computeHMAC(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return base64urlEncode(mac.Sum(nil))
}

func base64urlEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func base64urlDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
