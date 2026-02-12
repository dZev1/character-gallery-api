package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockAuthStore implements auth.AuthStore for testing
type MockAuthStore struct {
	ValidateFunc       func(keyHash string) (bool, error)
	UpdateLastUsedFunc func(keyHash string) error
}

func (m *MockAuthStore) ValidateAPIKey(keyHash string) (bool, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(keyHash)
	}
	return false, nil
}

func (m *MockAuthStore) UpdateLastUsed(keyHash string) error {
	if m.UpdateLastUsedFunc != nil {
		return m.UpdateLastUsedFunc(keyHash)
	}
	return nil
}

func (m *MockAuthStore) CreateAPIKey(name string) (string, error) {
	return "", nil
}

func TestRequireAPIKey_MissingHeader(t *testing.T) {
	mockStore := &MockAuthStore{}

	handler := RequireAPIKey(mockStore)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	if rec.Body.String() != "Missing API key\n" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRequireAPIKey_InvalidKey(t *testing.T) {
	mockStore := &MockAuthStore{
		ValidateFunc: func(keyHash string) (bool, error) {
			return false, nil
		},
	}

	handler := RequireAPIKey(mockStore)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "invalid_key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	if rec.Body.String() != "Invalid API key\n" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRequireAPIKey_ValidKey(t *testing.T) {
	lastUsedCalled := false

	mockStore := &MockAuthStore{
		ValidateFunc: func(keyHash string) (bool, error) {
			return true, nil
		},
		UpdateLastUsedFunc: func(keyHash string) error {
			lastUsedCalled = true
			return nil
		},
	}

	nextHandlerCalled := false
	handler := RequireAPIKey(mockStore)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "valid_api_key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !nextHandlerCalled {
		t.Fatal("next handler was not called")
	}

	if !lastUsedCalled {
		t.Fatal("UpdateLastUsed was not called")
	}
}

func TestRequireAPIKey_DBError(t *testing.T) {
	mockStore := &MockAuthStore{
		ValidateFunc: func(keyHash string) (bool, error) {
			return false, errors.New("database error")
		},
	}

	handler := RequireAPIKey(mockStore)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "some_key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	if rec.Body.String() != "Error validating API key\n" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRequireAPIKey_HashesKey(t *testing.T) {
	var receivedHash string

	mockStore := &MockAuthStore{
		ValidateFunc: func(keyHash string) (bool, error) {
			receivedHash = keyHash
			return true, nil
		},
	}

	handler := RequireAPIKey(mockStore)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "my_api_key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// The hash should be SHA-256 (64 hex chars), not the raw key
	if len(receivedHash) != 64 {
		t.Fatalf("expected hashed key (64 chars), got length: %d", len(receivedHash))
	}

	if receivedHash == "my_api_key" {
		t.Fatal("key should be hashed, not passed as-is")
	}
}
