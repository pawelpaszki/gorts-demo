package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestUserStore() *InMemoryUserStore {
	store := NewInMemoryUserStore()
	store.AddUser("admin", "secret123", "admin")
	store.AddUser("user", "password", "user")
	store.AddUser("reader", "readonly", "reader")
	return store
}

func TestInMemoryUserStore_Authenticate(t *testing.T) {
	store := newTestUserStore()

	tests := []struct {
		name     string
		username string
		password string
		wantOK   bool
		wantRole string
	}{
		{"valid admin", "admin", "secret123", true, "admin"},
		{"valid user", "user", "password", true, "user"},
		{"wrong password", "admin", "wrong", false, ""},
		{"unknown user", "unknown", "password", false, ""},
		{"empty credentials", "", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, ok := store.Authenticate(tt.username, tt.password)
			if ok != tt.wantOK {
				t.Errorf("Authenticate() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && user.Role != tt.wantRole {
				t.Errorf("Authenticate() role = %v, want %v", user.Role, tt.wantRole)
			}
		})
	}
}

func TestBasicAuth_Success(t *testing.T) {
	store := newTestUserStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			t.Error("Expected user in context")
			return
		}
		w.Write([]byte("Hello, " + user.Username))
	})

	protected := BasicAuth(store, "test")(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", EncodeBasicAuth("admin", "secret123"))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestBasicAuth_NoHeader(t *testing.T) {
	store := newTestUserStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protected := BasicAuth(store, "test")(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Error("Expected WWW-Authenticate header")
	}
}

func TestBasicAuth_InvalidCredentials(t *testing.T) {
	store := newTestUserStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protected := BasicAuth(store, "test")(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", EncodeBasicAuth("admin", "wrongpassword"))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestBasicAuth_MalformedHeader(t *testing.T) {
	store := newTestUserStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protected := BasicAuth(store, "test")(handler)

	tests := []struct {
		name   string
		header string
	}{
		{"empty", ""},
		{"wrong scheme", "Bearer token"},
		{"invalid base64", "Basic !!!invalid!!!"},
		{"no colon", "Basic " + "bm9jb2xvbg=="}, // "nocolon" in base64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()

			protected.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	store := newTestUserStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Admin access granted"))
	})

	protected := BasicAuth(store, "test")(RequireRole("admin")(handler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", EncodeBasicAuth("admin", "secret123"))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	store := newTestUserStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Admin access granted"))
	})

	protected := BasicAuth(store, "test")(RequireRole("admin")(handler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", EncodeBasicAuth("user", "password")) // user role, not admin
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	store := newTestUserStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Access granted"))
	})

	protected := BasicAuth(store, "test")(RequireRole("admin", "user")(handler))

	// Admin should have access
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", EncodeBasicAuth("admin", "secret123"))
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Admin: expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// User should have access
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", EncodeBasicAuth("user", "password"))
	rec = httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("User: expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Reader should NOT have access
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", EncodeBasicAuth("reader", "readonly"))
	rec = httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("Reader: expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestGetUser_NoUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := GetUser(req.Context())
	if user != nil {
		t.Error("Expected nil user when not authenticated")
	}
}

func TestEncodeBasicAuth(t *testing.T) {
	encoded := EncodeBasicAuth("admin", "secret")
	expected := "Basic YWRtaW46c2VjcmV0"
	if encoded != expected {
		t.Errorf("EncodeBasicAuth() = %s, want %s", encoded, expected)
	}
}
