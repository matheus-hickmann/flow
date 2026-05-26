package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuth_NoHeader_Returns401(t *testing.T) {
	Setup("", "http://dynamodb:8000")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler should not run")
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rr.Code)
	}
}

func TestRequireAuth_ValidDevToken_PutsUserIDInContext(t *testing.T) {
	Setup("", "http://dynamodb:8000")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer dev-matheus")

	var captured string
	RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = UserIDFrom(r.Context())
	})).ServeHTTP(rr, req)

	if captured != "matheus" {
		t.Errorf("expected userId=matheus, got %q", captured)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
}

func TestRequireAuth_DevTokenRejectedInProd(t *testing.T) {
	Setup("prod-secret", "")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer dev-matheus")

	RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler should not run")
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rr.Code)
	}
}

func TestRequireAuth_ValidSignedToken_PutsUserIDInContext(t *testing.T) {
	Setup("test-secret", "")
	token, err := IssueToken("alice")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	var captured string
	RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = UserIDFrom(r.Context())
	})).ServeHTTP(rr, req)

	if captured != "alice" {
		t.Errorf("expected userId=alice, got %q", captured)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
}

func TestRequireAuth_TamperedToken_Returns401(t *testing.T) {
	Setup("test-secret", "")
	token, _ := IssueToken("alice")
	tampered := token[:len(token)-2] + "XX"

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)

	RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler should not run")
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rr.Code)
	}
}
