package handlers

import (
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/uuidv7"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxUploadImageBytes = 2 << 20 // 2 MiB
	maxMultipartBytes   = maxUploadImageBytes + (64 << 10)
)

var (
	allowedUploadCategories = map[string]bool{
		"items":   true,
		"enemies": true,
		"misc":    true,
	}

	allowedImageMimeToExt = map[string]string{
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/jpeg": ".jpg",
		"image/gif":  ".gif",
	}

	safeUploadFilenameRe = regexp.MustCompile(`^[a-fA-F0-9-]+\.(png|webp|jpg|gif)$`)
)

type UploadHandler struct {
	dir string
}

func NewUploadHandler(dir string) *UploadHandler {
	if strings.TrimSpace(dir) == "" {
		dir = "./uploads"
	}
	return &UploadHandler{dir: dir}
}

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	if claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartBytes)
	if err := r.ParseMultipartForm(maxMultipartBytes); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid upload or file exceeds 2 MiB limit")
		return
	}

	category := strings.TrimSpace(r.FormValue("category"))
	if category == "" {
		category = strings.TrimSpace(r.URL.Query().Get("category"))
	}
	if category == "" {
		category = "misc"
	}
	if !allowedUploadCategories[category] {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Category must be items, enemies, or misc")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing image file field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadImageBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error reading uploaded file")
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Uploaded file is empty")
		return
	}
	if len(data) > maxUploadImageBytes {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "File exceeds 2 MiB limit")
		return
	}

	mimeType := http.DetectContentType(data)
	ext, allowed := allowedImageMimeToExt[mimeType]
	if !allowed {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Unsupported image format (allowed: PNG, WebP, JPEG, GIF)")
		return
	}

	id, err := uuidv7.New()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error generating file id")
		return
	}

	categoryDir := filepath.Join(h.dir, category)
	if err := os.MkdirAll(categoryDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error preparing uploads directory")
		return
	}

	filename := id.String() + ext
	targetPath := filepath.Join(categoryDir, filename)
	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error saving image")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"image_url": "/uploads/" + category + "/" + filename,
	})
}

func (h *UploadHandler) HandleServeUpload(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")
	filename := r.PathValue("filename")

	if !allowedUploadCategories[category] || !safeUploadFilenameRe.MatchString(filename) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Image not found")
		return
	}

	fullPath := filepath.Join(h.dir, category, filename)
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Image not found")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, fullPath)
}
