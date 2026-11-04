package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host              string   `json:"host"`
	Port              int      `json:"port"`
	BaseURL           string   `json:"base_url"`
	StorageDir        string   `json:"storage_dir"`
	MaxUploadSizeMB   int64    `json:"max_upload_size_mb"`
	CacheControl      string   `json:"cache_control"`
	EnableGzip        bool     `json:"enable_gzip"`
	APIKey            string   `json:"api_key"`
	SecretKey         string   `json:"secret_key"`
	RequireSignedURLs bool     `json:"require_signed_urls"`
	AllowedReferrers  []string `json:"allowed_referrers"`
	AllowedOrigins    []string `json:"allowed_origins"`
	RateLimitRPS      float64  `json:"rate_limit_rps"`
	RateLimitBurst    int      `json:"rate_limit_burst"`
}

func DefaultConfig() *Config {
	return &Config{
		Host:              "0.0.0.0",
		Port:              8080,
		BaseURL:           "http://localhost:8080",
		StorageDir:        "./data",
		MaxUploadSizeMB:   100,
		CacheControl:      "public, max-age=31536000, immutable",
		EnableGzip:        true,
		APIKey:            "cdn-secret-api-key-change-me",
		SecretKey:         "hmac-signing-secret-change-me",
		RequireSignedURLs: false,
		AllowedReferrers:  []string{},
		AllowedOrigins:    []string{"*"},
		RateLimitRPS:      100,
		RateLimitBurst:    200,
	}
}

func Load(filePath string) (*Config, error) {
	cfg := DefaultConfig()

	if filePath != "" {
		if data, err := os.ReadFile(filePath); err == nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}

	if host := os.Getenv("CDN_HOST"); host != "" {
		cfg.Host = host
	}
	portStr := os.Getenv("CDN_PORT")
	if portStr == "" {
		portStr = os.Getenv("PORT")
	}
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = p
		}
	}
	if baseURL := os.Getenv("CDN_BASE_URL"); baseURL != "" {
		cfg.BaseURL = strings.TrimRight(baseURL, "/")
	}
	if storageDir := os.Getenv("CDN_STORAGE_DIR"); storageDir != "" {
		cfg.StorageDir = storageDir
	}
	if maxUpload := os.Getenv("CDN_MAX_UPLOAD_SIZE_MB"); maxUpload != "" {
		if m, err := strconv.ParseInt(maxUpload, 10, 64); err == nil {
			cfg.MaxUploadSizeMB = m
		}
	}
	if cc := os.Getenv("CDN_CACHE_CONTROL"); cc != "" {
		cfg.CacheControl = cc
	}
	if gzip := os.Getenv("CDN_ENABLE_GZIP"); gzip != "" {
		cfg.EnableGzip = gzip == "true" || gzip == "1"
	}
	if apiKey := os.Getenv("CDN_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if secretKey := os.Getenv("CDN_SECRET_KEY"); secretKey != "" {
		cfg.SecretKey = secretKey
	}
	if reqSigned := os.Getenv("CDN_REQUIRE_SIGNED_URLS"); reqSigned != "" {
		cfg.RequireSignedURLs = reqSigned == "true" || reqSigned == "1"
	}
	if referrers := os.Getenv("CDN_ALLOWED_REFERRERS"); referrers != "" {
		cfg.AllowedReferrers = splitAndTrim(referrers, ",")
	}
	if origins := os.Getenv("CDN_ALLOWED_ORIGINS"); origins != "" {
		cfg.AllowedOrigins = splitAndTrim(origins, ",")
	}
	if rps := os.Getenv("CDN_RATE_LIMIT_RPS"); rps != "" {
		if r, err := strconv.ParseFloat(rps, 64); err == nil {
			cfg.RateLimitRPS = r
		}
	}
	if burst := os.Getenv("CDN_RATE_LIMIT_BURST"); burst != "" {
		if b, err := strconv.Atoi(burst); err == nil {
			cfg.RateLimitBurst = b
		}
	}

	return cfg, nil
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
