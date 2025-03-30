package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOAuthMiddleware_Unauthorized(t *testing.T) {
	handler := oauthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestOAuthMiddleware_Authorized(t *testing.T) {
	handler := oauthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestProxyRequest(t *testing.T) {
	// Create a test server to simulate the target service.
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from target"))
	}))
	defer target.Close()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	proxyRequest(target.URL, w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "Hello from target" {
		t.Errorf("Expected body 'Hello from target', got '%s'", body)
	}
}
