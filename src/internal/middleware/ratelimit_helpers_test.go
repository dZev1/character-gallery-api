package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestClientIP_UsesForwardedFor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.10:54321"
	r.Header.Set("X-Forwarded-For", "198.51.100.7, 203.0.113.10")

	if got := clientIP(r); got != "198.51.100.7" {
		t.Errorf("clientIP = %q, want first forwarded address", got)
	}
}

func TestClientIP_FallsBackToRemoteAddr(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "198.51.100.7:54321"

	if got := clientIP(r); got != "198.51.100.7" {
		t.Errorf("clientIP = %q, want RemoteAddr without port", got)
	}
}

func TestClientIP_SkipsEmptyForwarded(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "198.51.100.7:54321"
	r.Header.Set("X-Forwarded-For", "  ,  ")

	if got := clientIP(r); got != "198.51.100.7" {
		t.Errorf("clientIP = %q, want RemoteAddr for empty forwarded header", got)
	}
}
