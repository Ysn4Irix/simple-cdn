package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"simple-cdn/internal/config"
	"simple-cdn/internal/delivery"
	"simple-cdn/internal/security"
	"simple-cdn/internal/storage"
)

type API struct {
	cfg      *config.Config
	storage  *storage.Engine
	delivery *delivery.Handler
	signer   *security.URLSigner
}

func New(cfg *config.Config, eng *storage.Engine, del *delivery.Handler, signer *security.URLSigner) *API {
	return &API{
		cfg:      cfg,
		storage:  eng,
		delivery: del,
		signer:   signer,
	}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/upload", a.handleUpload)
	mux.HandleFunc("/api/files", a.handleFiles)
	mux.HandleFunc("/api/files/", a.handleFileItem)
	mux.HandleFunc("/api/purge", a.handlePurge)
	mux.HandleFunc("/api/sign-url", a.handleSignURL)
}

func (a *API) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (a *API) writeError(w http.ResponseWriter, status int, message string) {
	a.writeJSON(w, status, map[string]string{"error": message})
}

func (a *API) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	maxBytes := a.cfg.MaxUploadSizeMB * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	if err := r.ParseMultipartForm(32 * 1024 * 1024); err != nil {
		a.writeError(w, http.StatusBadRequest, "Failed to parse multipart form or payload too large")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		a.writeError(w, http.StatusBadRequest, "Form field 'file' is required")
		return
	}
	defer file.Close()

	destPath := strings.TrimSpace(r.FormValue("path"))
	if destPath == "" {
		destPath = filepath.Base(header.Filename)
	}

	info, err := a.storage.Save(destPath, file)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidPath) {
			a.writeError(w, http.StatusBadRequest, "Invalid destination path")
			return
		}
		a.writeError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}

	publicURL := fmt.Sprintf("%s/files/%s", a.cfg.BaseURL, info.Path)
	a.delivery.Purge(info.Path)

	a.writeJSON(w, http.StatusCreated, map[string]any{
		"status":       "success",
		"path":         info.Path,
		"url":          publicURL,
		"size":         info.Size,
		"etag":         info.ETag,
		"content_type": info.ContentType,
		"sha256":       info.SHA256,
	})
}

func (a *API) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	prefix := r.URL.Query().Get("prefix")
	list, err := a.storage.List(prefix)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Failed to list files")
		return
	}

	type fileItem struct {
		Path        string    `json:"path"`
		Size        int64     `json:"size"`
		ModTime     time.Time `json:"mod_time"`
		ETag        string    `json:"etag"`
		ContentType string    `json:"content_type"`
		URL         string    `json:"url"`
	}

	items := make([]fileItem, 0, len(list))
	for _, f := range list {
		items = append(items, fileItem{
			Path:        f.Path,
			Size:        f.Size,
			ModTime:     f.ModTime,
			ETag:        f.ETag,
			ContentType: f.ContentType,
			URL:         fmt.Sprintf("%s/files/%s", a.cfg.BaseURL, f.Path),
		})
	}

	a.writeJSON(w, http.StatusOK, map[string]any{
		"count": len(items),
		"files": items,
	})
}

func (a *API) handleFileItem(w http.ResponseWriter, r *http.Request) {
	relPath := strings.TrimPrefix(r.URL.Path, "/api/files/")
	if relPath == "" {
		a.writeError(w, http.StatusBadRequest, "Missing file path")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		err := a.storage.Delete(relPath)
		if err != nil {
			if errors.Is(err, storage.ErrFileNotFound) {
				a.writeError(w, http.StatusNotFound, "File not found")
				return
			}
			if errors.Is(err, storage.ErrInvalidPath) {
				a.writeError(w, http.StatusBadRequest, "Invalid file path")
				return
			}
			a.writeError(w, http.StatusInternalServerError, "Failed to delete file")
			return
		}

		a.delivery.Purge(relPath)
		a.writeJSON(w, http.StatusOK, map[string]string{
			"status":  "deleted",
			"message": "File successfully removed",
			"path":    relPath,
		})
	default:
		a.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (a *API) handlePurge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		a.writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	target := req.Path
	if target == "" {
		target = "*"
	}

	a.delivery.Purge(target)
	a.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "purged",
		"message": "Cache successfully invalidated",
		"target":  target,
	})
}

func (a *API) handleSignURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Path       string `json:"path"`
		TTLSeconds int64  `json:"ttl_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.Path == "" {
		a.writeError(w, http.StatusBadRequest, "Path is required")
		return
	}

	ttl := 1 * time.Hour
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	signedURL := a.signer.SignURL(a.cfg.BaseURL, req.Path, ttl)
	expiresAt := time.Now().Add(ttl).Unix()

	a.writeJSON(w, http.StatusOK, map[string]any{
		"url":        signedURL,
		"expires_at": expiresAt,
		"ttl_sec":    ttl.Seconds(),
	})
}
