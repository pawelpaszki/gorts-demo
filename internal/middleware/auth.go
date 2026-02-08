package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// UserContextKey is the context key for the authenticated user.
	UserContextKey contextKey = "user"
)

// Credentials represents user credentials for authentication.
type Credentials struct {
	Username string
	Password string
}

// User represents an authenticated user.
type User struct {
	Username string
	Role     string
}

// UserStore defines the interface for user authentication.
type UserStore interface {
	Authenticate(username, password string) (*User, bool)
}

// InMemoryUserStore is a simple in-memory user store.
type InMemoryUserStore struct {
	users map[string]Credentials
	roles map[string]string
}

// NewInMemoryUserStore creates a new in-memory user store.
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]Credentials),
		roles: make(map[string]string),
	}
}

// AddUser adds a user to the store.
func (s *InMemoryUserStore) AddUser(username, password, role string) {
	s.users[username] = Credentials{Username: username, Password: password}
	s.roles[username] = role
}

// Authenticate checks if the credentials are valid.
func (s *InMemoryUserStore) Authenticate(username, password string) (*User, bool) {
	creds, exists := s.users[username]
	if !exists {
		return nil, false
	}

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(creds.Password), []byte(password)) != 1 {
		return nil, false
	}

	return &User{
		Username: username,
		Role:     s.roles[username],
	}, true
}

// BasicAuth returns a middleware that requires HTTP Basic Authentication.
func BasicAuth(store UserStore, realm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := parseBasicAuth(r.Header.Get("Authorization"))
			if !ok {
				requireAuth(w, realm)
				return
			}

			user, authenticated := store.Authenticate(username, password)
			if !authenticated {
				requireAuth(w, realm)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns a middleware that requires a specific role.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool)
	for _, role := range roles {
		roleSet[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r.Context())
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !roleSet[user.Role] {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUser retrieves the authenticated user from the context.
func GetUser(ctx context.Context) *User {
	user, ok := ctx.Value(UserContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// parseBasicAuth parses the Authorization header for Basic auth.
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}

	decodedStr := string(decoded)
	colonIdx := strings.IndexByte(decodedStr, ':')
	if colonIdx < 0 {
		return "", "", false
	}

	return decodedStr[:colonIdx], decodedStr[colonIdx+1:], true
}

// requireAuth sends a 401 response requesting authentication.
func requireAuth(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// EncodeBasicAuth encodes username and password for Basic auth header.
func EncodeBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
