package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/pkg/response"
	"github.com/google/uuid"
)

// UploadHandler handles HTTP requests for file upload endpoints.
type UploadHandler struct {
	cfg *config.Config
}

// NewUploadHandler creates a new UploadHandler instance.
func NewUploadHandler(cfg *config.Config) *UploadHandler {
	return &UploadHandler{cfg: cfg}
}

// maxUploadSize is the maximum allowed file size (10 MB).
const maxUploadSize = 10 << 20

// allowedMIMETypes lists the MIME types accepted for uploads.
var allowedMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/svg+xml": true,
}

// allowedExtensions lists the file extensions accepted for uploads.
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".svg":  true,
}

// Upload handles POST /api/upload (requires auth).
// It reads a multipart file, uploads it to Supabase Storage, and returns the public URL.
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	// Limit request body size.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.Error(w, http.StatusBadRequest, "FILE_TOO_LARGE", "File size exceeds the 10MB limit")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_FILE", "No file found in request. Use 'file' as the form field name.")
		return
	}
	defer file.Close()

	// Validate file extension.
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		response.Error(w, http.StatusBadRequest, "INVALID_FILE_TYPE", "File type not allowed. Accepted: jpg, jpeg, png, gif, webp, svg")
		return
	}

	// Read file content to detect MIME type.
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		response.Error(w, http.StatusInternalServerError, "READ_ERROR", "Failed to read file")
		return
	}

	mimeType := http.DetectContentType(buf[:n])
	if !allowedMIMETypes[mimeType] {
		// SVG might not be detected by DetectContentType, allow if extension is .svg.
		if ext != ".svg" {
			response.Error(w, http.StatusBadRequest, "INVALID_FILE_TYPE", fmt.Sprintf("MIME type '%s' is not allowed", mimeType))
			return
		}
		mimeType = "image/svg+xml"
	}

	// Seek back to the beginning of the file for upload.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.Error(w, http.StatusInternalServerError, "READ_ERROR", "Failed to process file")
		return
	}

	// Generate a unique file path: {userID}/{timestamp}_{uuid}{ext}
	timestamp := time.Now().Unix()
	uniqueID := uuid.New().String()[:8]
	objectPath := fmt.Sprintf("%s/%d_%s%s", userID.String(), timestamp, uniqueID, ext)

	// Upload to Supabase Storage via REST API.
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		h.cfg.SupabaseURL, h.cfg.StorageBucket, objectPath)

	uploadReq, err := http.NewRequest("PUT", uploadURL, file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Failed to prepare upload request")
		return
	}

	uploadReq.Header.Set("Authorization", "Bearer "+h.cfg.SupabaseServiceKey)
	uploadReq.Header.Set("Content-Type", mimeType)
	uploadReq.Header.Set("x-upsert", "true")
	uploadReq.ContentLength = header.Size

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(uploadReq)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Failed to upload to storage")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		response.Error(w, http.StatusInternalServerError, "UPLOAD_ERROR",
			fmt.Sprintf("Storage upload failed with status %d: %s", resp.StatusCode, string(body)))
		return
	}

	// Construct the public URL.
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		h.cfg.SupabaseURL, h.cfg.StorageBucket, objectPath)

	response.JSON(w, http.StatusOK, map[string]string{
		"url":      publicURL,
		"path":     objectPath,
		"filename": header.Filename,
		"mime_type": mimeType,
	})
}
