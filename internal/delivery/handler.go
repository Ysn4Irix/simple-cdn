package delivery

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"simple-cdn/internal/config"
	"simple-cdn/internal/storage"
)

type Handler struct {
	cfg     *config.Config
	storage *storage.Engine

	cacheMu sync.RWMutex
	cache   map[string]*storage.FileInfo
}

func NewHandler(cfg *config.Config, eng *storage.Engine) *Handler {
	return &Handler{
		cfg:     cfg,
		storage: eng,
		cache:   make(map[string]*storage.FileInfo),
	}
}

func (h *Handler) Purge(path string) {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	cleaned := strings.TrimPrefix(strings.TrimSpace(path), "/")
	if cleaned == "" || cleaned == "*" {
		h.cache = make(map[string]*storage.FileInfo)
		return
	}

	delete(h.cache, cleaned)
	for k := range h.cache {
		if strings.HasPrefix(k, cleaned) {
			delete(h.cache, k)
		}
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	relPath := strings.TrimPrefix(r.URL.Path, "/")
	relPath = strings.TrimPrefix(relPath, "files/")
	if relPath == "" {
		http.NotFound(w, r)
		return
	}

	file, info, err := h.storage.Open(relPath)
	if err != nil {
		if errors.Is(err, storage.ErrFileNotFound) || errors.Is(err, storage.ErrInvalidPath) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	h.cacheMu.Lock()
	h.cache[relPath] = info
	h.cacheMu.Unlock()

	header := w.Header()
	header.Set("Accept-Ranges", "bytes")
	header.Set("ETag", info.ETag)
	header.Set("Last-Modified", info.ModTime.UTC().Format(http.TimeFormat))
	header.Set("X-Content-Type-Options", "nosniff")

	if h.cfg.CacheControl != "" {
		header.Set("Cache-Control", h.cfg.CacheControl)
	}

	if info.ContentType != "" {
		header.Set("Content-Type", info.ContentType)
	}

	// Conditional requests: ETag validation
	if match := r.Header.Get("If-None-Match"); match != "" {
		if etagMatch(match, info.ETag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Conditional requests: If-Modified-Since validation
	if sinceHeader := r.Header.Get("If-Modified-Since"); sinceHeader != "" && r.Header.Get("If-None-Match") == "" {
		if t, parseErr := time.Parse(http.TimeFormat, sinceHeader); parseErr == nil {
			if !info.ModTime.Truncate(time.Second).After(t.Truncate(time.Second)) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	http.ServeContent(w, r, info.Path, info.ModTime, file)
}

func etagMatch(clientHeader, targetETag string) bool {
	if clientHeader == "*" {
		return true
	}
	cleanTarget := strings.TrimPrefix(strings.Trim(targetETag, "\""), "W/")
	for _, part := range strings.Split(clientHeader, ",") {
		cleanPart := strings.TrimSpace(part)
		cleanPart = strings.TrimPrefix(cleanPart, "W/")
		cleanPart = strings.Trim(cleanPart, "\"")
		if cleanPart == cleanTarget {
			return true
		}
	}
	return false
}
