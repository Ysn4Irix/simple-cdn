package delivery

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simple-cdn/internal/config"
	"simple-cdn/internal/storage"
)

func setupTestDelivery(t *testing.T) (*Handler, *storage.Engine) {
	tempDir := t.TempDir()
	eng, err := storage.New(tempDir)
	if err != nil {
		t.Fatalf("failed to create storage engine: %v", err)
	}

	cfg := config.DefaultConfig()
	handler := NewHandler(cfg, eng)
	return handler, eng
}

func TestDeliveryStandardHeaders(t *testing.T) {
	handler, eng := setupTestDelivery(t)

	content := "body { color: red; }"
	_, err := eng.Save("css/test.css", strings.NewReader(content))
	if err != nil {
		t.Fatalf("failed to save file: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/files/css/test.css", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("ETag") == "" {
		t.Error("expected ETag header")
	}
	if rec.Header().Get("Accept-Ranges") != "bytes" {
		t.Errorf("expected Accept-Ranges: bytes, got %s", rec.Header().Get("Accept-Ranges"))
	}
	if rec.Header().Get("Cache-Control") != handler.cfg.CacheControl {
		t.Errorf("expected Cache-Control %s, got %s", handler.cfg.CacheControl, rec.Header().Get("Cache-Control"))
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected nosniff header, got %s", rec.Header().Get("X-Content-Type-Options"))
	}
	if rec.Body.String() != content {
		t.Errorf("expected body %q, got %q", content, rec.Body.String())
	}
}

func TestDeliveryIfNoneMatch(t *testing.T) {
	handler, eng := setupTestDelivery(t)

	_, _ = eng.Save("doc.txt", strings.NewReader("sample text"))

	req := httptest.NewRequest(http.MethodGet, "/files/doc.txt", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header")
	}

	req304 := httptest.NewRequest(http.MethodGet, "/files/doc.txt", nil)
	req304.Header.Set("If-None-Match", etag)
	rec304 := httptest.NewRecorder()

	handler.ServeHTTP(rec304, req304)

	if rec304.Code != http.StatusNotModified {
		t.Fatalf("expected status 304, got %d", rec304.Code)
	}
	if rec304.Body.Len() > 0 {
		t.Errorf("expected empty body on 304, got %d bytes", rec304.Body.Len())
	}
}

func TestDeliveryByteRange(t *testing.T) {
	handler, eng := setupTestDelivery(t)

	fullText := "0123456789abcdefghijklmnopqrstuvwxyz"
	_, _ = eng.Save("range.txt", strings.NewReader(fullText))

	req := httptest.NewRequest(http.MethodGet, "/files/range.txt", nil)
	req.Header.Set("Range", "bytes=0-9")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusPartialContent {
		t.Fatalf("expected status 206 Partial Content, got %d", rec.Code)
	}

	expectedBody := fullText[0:10]
	if rec.Body.String() != expectedBody {
		t.Errorf("expected partial content %q, got %q", expectedBody, rec.Body.String())
	}

	contentRange := rec.Header().Get("Content-Range")
	if !strings.HasPrefix(contentRange, "bytes 0-9/") {
		t.Errorf("unexpected Content-Range header: %s", contentRange)
	}
}

func TestDeliveryCompression(t *testing.T) {
	handler, eng := setupTestDelivery(t)

	longContent := strings.Repeat("Lorem ipsum dolor sit amet. ", 100)
	_, _ = eng.Save("text.txt", strings.NewReader(longContent))

	wrapped := CompressionMiddleware(true, handler)

	req := httptest.NewRequest(http.MethodGet, "/files/text.txt", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding: gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("failed to read decompressed body: %v", err)
	}

	if string(decompressed) != longContent {
		t.Errorf("decompressed text mismatch: got %d bytes, want %d bytes", len(decompressed), len(longContent))
	}
}
