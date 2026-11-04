package security

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestURLSigner(t *testing.T) {
	signer := NewURLSigner("super-secret-key-12345")
	path := "videos/intro.mp4"

	// 1. Valid signature
	expires := time.Now().Add(10 * time.Minute)
	token := signer.GenerateToken(path, expires)

	query := url.Values{}
	query.Set("token", token)
	query.Set("expires", strconv.FormatInt(expires.Unix(), 10))

	if err := signer.Verify(path, query); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}

	// 2. Expired signature
	expiredTime := time.Now().Add(-1 * time.Minute)
	expiredToken := signer.GenerateToken(path, expiredTime)
	queryExpired := url.Values{}
	queryExpired.Set("token", expiredToken)
	queryExpired.Set("expires", strconv.FormatInt(expiredTime.Unix(), 10))

	if err := signer.Verify(path, queryExpired); err != ErrExpiredSignature {
		t.Errorf("expected ErrExpiredSignature, got %v", err)
	}

	// 3. Tampered path
	if err := signer.Verify("videos/other.mp4", query); err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature for tampered path, got %v", err)
	}

	// 4. Tampered token
	queryTampered := url.Values{}
	queryTampered.Set("token", "invalid-hex-token")
	queryTampered.Set("expires", strconv.FormatInt(expires.Unix(), 10))
	if err := signer.Verify(path, queryTampered); err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestReferrerMiddleware(t *testing.T) {
	allowed := []string{"mysite.com", "*.partner.org"}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := ReferrerMiddleware(allowed, nextHandler)

	// Direct request (no referer) -> allowed
	reqDirect := httptest.NewRequest(http.MethodGet, "/files/image.png", nil)
	recDirect := httptest.NewRecorder()
	handler.ServeHTTP(recDirect, reqDirect)
	if recDirect.Code != http.StatusOK {
		t.Errorf("expected direct request allowed (200), got %d", recDirect.Code)
	}

	// Allowed exact domain
	reqExact := httptest.NewRequest(http.MethodGet, "/files/image.png", nil)
	reqExact.Header.Set("Referer", "https://mysite.com/blog/article")
	recExact := httptest.NewRecorder()
	handler.ServeHTTP(recExact, reqExact)
	if recExact.Code != http.StatusOK {
		t.Errorf("expected exact referrer allowed (200), got %d", recExact.Code)
	}

	// Allowed wildcard subdomain
	reqWildcard := httptest.NewRequest(http.MethodGet, "/files/image.png", nil)
	reqWildcard.Header.Set("Referer", "https://sub.partner.org/dashboard")
	recWildcard := httptest.NewRecorder()
	handler.ServeHTTP(recWildcard, reqWildcard)
	if recWildcard.Code != http.StatusOK {
		t.Errorf("expected wildcard referrer allowed (200), got %d", recWildcard.Code)
	}

	// Disallowed hotlink
	reqBad := httptest.NewRequest(http.MethodGet, "/files/image.png", nil)
	reqBad.Header.Set("Referer", "https://leechsite.com/pirate")
	recBad := httptest.NewRecorder()
	handler.ServeHTTP(recBad, reqBad)
	if recBad.Code != http.StatusForbidden {
		t.Errorf("expected hotlink blocked (403), got %d", recBad.Code)
	}
}

func TestAuthGuard(t *testing.T) {
	expectedKey := "test-secret-api-key"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	guard := AuthGuard(expectedKey, nextHandler)

	// Missing auth header
	reqNoAuth := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
	recNoAuth := httptest.NewRecorder()
	guard.ServeHTTP(recNoAuth, reqNoAuth)
	if recNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", recNoAuth.Code)
	}

	// Valid X-API-Key
	reqAPIKey := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
	reqAPIKey.Header.Set("X-API-Key", expectedKey)
	recAPIKey := httptest.NewRecorder()
	guard.ServeHTTP(recAPIKey, reqAPIKey)
	if recAPIKey.Code != http.StatusOK {
		t.Errorf("expected 200 OK with valid API key, got %d", recAPIKey.Code)
	}

	// Valid Bearer token
	reqBearer := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
	reqBearer.Header.Set("Authorization", "Bearer "+expectedKey)
	recBearer := httptest.NewRecorder()
	guard.ServeHTTP(recBearer, reqBearer)
	if recBearer.Code != http.StatusOK {
		t.Errorf("expected 200 OK with Bearer token, got %d", recBearer.Code)
	}
}
