package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrMissingSignature = errors.New("missing signature token or expires parameter")
	ErrExpiredSignature = errors.New("signed URL has expired")
	ErrInvalidSignature = errors.New("invalid signature token")
)

type URLSigner struct {
	secretKey []byte
}

func NewURLSigner(secret string) *URLSigner {
	return &URLSigner{secretKey: []byte(secret)}
}

func (s *URLSigner) GenerateToken(path string, expiresAt time.Time) string {
	cleanPath := strings.TrimPrefix(filepathNormalize(path), "files/")
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	payload := fmt.Sprintf("%s:%d", cleanPath, expiresAt.Unix())

	mac := hmac.New(sha256.New, s.secretKey)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *URLSigner) Verify(path string, query url.Values) error {
	token := query.Get("token")
	expiresStr := query.Get("expires")

	if token == "" || expiresStr == "" {
		return ErrMissingSignature
	}

	expiresUnix, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return ErrInvalidSignature
	}

	if time.Now().Unix() > expiresUnix {
		return ErrExpiredSignature
	}

	cleanPath := strings.TrimPrefix(filepathNormalize(path), "files/")
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	expectedToken := s.GenerateToken(cleanPath, time.Unix(expiresUnix, 0))
	if !hmac.Equal([]byte(token), []byte(expectedToken)) {
		return ErrInvalidSignature
	}

	return nil
}

func (s *URLSigner) SignURL(baseURL, path string, ttl time.Duration) string {
	expiresAt := time.Now().Add(ttl)
	token := s.GenerateToken(path, expiresAt)

	cleanBase := strings.TrimRight(baseURL, "/")
	cleanPath := strings.TrimPrefix(filepathNormalize(path), "/")

	return fmt.Sprintf("%s/files/%s?token=%s&expires=%d", cleanBase, cleanPath, token, expiresAt.Unix())
}

func filepathNormalize(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	return strings.TrimPrefix(p, "/")
}
