package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.StorageDir != "./data" {
		t.Fatalf("expected default storage dir './data', got %s", cfg.StorageDir)
	}
	if !cfg.EnableGzip {
		t.Fatal("expected default EnableGzip to be true")
	}
	if cfg.CacheControl != "public, max-age=60" {
		t.Fatalf("expected default CacheControl 'public, max-age=60', got %s", cfg.CacheControl)
	}
}

func TestLoadWithJSONFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	content := []byte(`{
		"port": 9090,
		"storage_dir": "/var/cdn",
		"enable_gzip": false,
		"max_upload_size_mb": 50
	}`)

	if err := os.WriteFile(configFile, content, 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.StorageDir != "/var/cdn" {
		t.Errorf("expected storage dir '/var/cdn', got %s", cfg.StorageDir)
	}
	if cfg.EnableGzip {
		t.Error("expected EnableGzip to be false")
	}
	if cfg.MaxUploadSizeMB != 50 {
		t.Errorf("expected max upload size 50, got %d", cfg.MaxUploadSizeMB)
	}
}

func TestLoadWithEnvOverrides(t *testing.T) {
	os.Setenv("CDN_PORT", "7777")
	os.Setenv("CDN_API_KEY", "custom-api-key")
	os.Setenv("CDN_ALLOWED_REFERRERS", "example.com, test.org")
	defer func() {
		os.Unsetenv("CDN_PORT")
		os.Unsetenv("CDN_API_KEY")
		os.Unsetenv("CDN_ALLOWED_REFERRERS")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	if cfg.Port != 7777 {
		t.Errorf("expected port 7777 from env, got %d", cfg.Port)
	}
	if cfg.APIKey != "custom-api-key" {
		t.Errorf("expected API key 'custom-api-key', got %s", cfg.APIKey)
	}
	if len(cfg.AllowedReferrers) != 2 || cfg.AllowedReferrers[0] != "example.com" || cfg.AllowedReferrers[1] != "test.org" {
		t.Errorf("unexpected allowed referrers: %v", cfg.AllowedReferrers)
	}
}

func TestLoadWithStandardPortEnv(t *testing.T) {
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("expected port 3000 from PORT env, got %d", cfg.Port)
	}
}
