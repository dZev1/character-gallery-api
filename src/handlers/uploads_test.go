package handlers

import (
	"bytes"
	"dZev1/character-gallery/internal/auth"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// Minimal valid 1x1 PNG file
var samplePNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

func makeMultipartUploadRequest(t *testing.T, category, filename string, data []byte, role string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if category != "" {
		if err := writer.WriteField("category", category); err != nil {
			t.Fatalf("write category field: %v", err)
		}
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write file bytes: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if role != "" {
		ctx := auth.ContextWithClaims(req.Context(), &auth.Claims{
			UserID:   uuid.New(),
			Username: "tester",
			Role:     role,
		})
		req = req.WithContext(ctx)
	}
	return req
}

func TestUploadHandler_AdminUploadAndServe(t *testing.T) {
	dir := t.TempDir()
	h := NewUploadHandler(dir)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/uploads", h.HandleUpload)
	mux.HandleFunc("GET /uploads/{category}/{filename}", h.HandleServeUpload)

	// 1. Admin uploads valid PNG
	req := makeMultipartUploadRequest(t, "enemies", "goblin.png", samplePNG, "admin")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	imageURL := resp["image_url"]
	if imageURL == "" {
		t.Fatalf("expected non-empty image_url")
	}

	// 2. Serve uploaded image publicly (no auth claims)
	getReq := httptest.NewRequest(http.MethodGet, imageURL, nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 serving %s, got %d", imageURL, getRec.Code)
	}
	if !bytes.Equal(getRec.Body.Bytes(), samplePNG) {
		t.Fatalf("served bytes do not match uploaded PNG")
	}
	if getRec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("expected nosniff header")
	}
}

func TestUploadHandler_ForbiddenForNonAdmin(t *testing.T) {
	h := NewUploadHandler(t.TempDir())
	req := makeMultipartUploadRequest(t, "items", "sword.png", samplePNG, "user")
	rec := httptest.NewRecorder()
	h.HandleUpload(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestUploadHandler_RejectsUnsupportedFormat(t *testing.T) {
	h := NewUploadHandler(t.TempDir())
	req := makeMultipartUploadRequest(t, "items", "evil.html", []byte("<script>alert(1)</script>"), "admin")
	rec := httptest.NewRecorder()
	h.HandleUpload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUploadHandler_RejectsInvalidServePath(t *testing.T) {
	h := NewUploadHandler(t.TempDir())
	mux := http.NewServeMux()
	mux.HandleFunc("GET /uploads/{category}/{filename}", h.HandleServeUpload)

	req := httptest.NewRequest(http.MethodGet, "/uploads/enemies/..passwd.png", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
