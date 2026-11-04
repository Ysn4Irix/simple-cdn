package security

import (
	"net/http"
	"net/url"
	"strings"
)

func ReferrerMiddleware(allowed []string, next http.Handler) http.Handler {
	if len(allowed) == 0 {
		return next
	}

	allowedMap := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowedMap[strings.ToLower(strings.TrimSpace(a))] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ref := r.Header.Get("Referer")
		if ref == "" {
			// Allow direct requests without referer (e.g. CLI, curl, browser address bar)
			next.ServeHTTP(w, r)
			return
		}

		parsed, err := url.Parse(ref)
		if err != nil {
			http.Error(w, "Forbidden: Invalid Referer", http.StatusForbidden)
			return
		}

		host := strings.ToLower(parsed.Hostname())
		if isHostAllowed(host, allowedMap) {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Forbidden: Hotlinking Prohibited", http.StatusForbidden)
	})
}

func isHostAllowed(host string, allowedMap map[string]struct{}) bool {
	if _, ok := allowedMap[host]; ok {
		return true
	}

	// Wildcard domain matching (*.example.com)
	for allowed := range allowedMap {
		if strings.HasPrefix(allowed, "*.") {
			rootDomain := allowed[2:]
			if host == rootDomain || strings.HasSuffix(host, "."+rootDomain) {
				return true
			}
		}
	}
	return false
}

func CORSMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	origins := "*"
	if len(allowedOrigins) > 0 {
		origins = strings.Join(allowedOrigins, ", ")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origins)
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Range, Content-Type, Authorization, X-API-Key, If-None-Match")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, ETag, Accept-Ranges")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
