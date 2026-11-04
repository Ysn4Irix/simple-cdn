package delivery

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	// Do not compress 304 Not Modified or 204 No Content
	if status == http.StatusNotModified || status == http.StatusNoContent {
		w.ResponseWriter.WriteHeader(status)
		return
	}

	w.Header().Del("Content-Length")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Add("Vary", "Accept-Encoding")
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.writer == nil {
		return w.ResponseWriter.Write(b)
	}
	return w.writer.Write(b)
}

func isCompressible(contentType string) bool {
	ct := strings.ToLower(contentType)
	if strings.HasPrefix(ct, "text/") {
		return true
	}
	compressibleSubstrings := []string{
		"application/json",
		"application/javascript",
		"application/x-javascript",
		"application/xml",
		"application/rss+xml",
		"application/atom+xml",
		"application/wasm",
		"image/svg+xml",
		"font/",
	}
	for _, sub := range compressibleSubstrings {
		if strings.Contains(ct, sub) {
			return true
		}
	}
	return false
}

func CompressionMiddleware(enabled bool, next http.Handler) http.Handler {
	if !enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || r.Header.Get("Range") != "" {
			next.ServeHTTP(w, r)
			return
		}

		gw := gzipWriterPool.Get().(*gzip.Writer)
		gw.Reset(w)
		defer func() {
			_ = gw.Close()
			gzipWriterPool.Put(gw)
		}()

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			writer:         gw,
		}

		next.ServeHTTP(gzw, r)
	})
}
