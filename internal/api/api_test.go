package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simple-cdn/internal/config"
	"simple-cdn/internal/delivery"
	"simple-cdn/internal/security"
	"simple-cdn/internal/storage"
)

func setupTestAPI(t *testing.T) (*API, *storage.Engine) {
	tempDir := t.TempDir()
	eng, err := storage.New(tempDir)
	if err != nil {
		t.Fatalf("failed to create storage engine: %v", err)
	}

	cfg := config.DefaultConfig()
	del := delivery.NewHandler(cfg, eng)
	signer := security.NewURLSigner(cfg.SecretKey)
	api := New(cfg, eng, del, signer)

	return api, eng
}

func TestAPIUploadAndList(t *testing.T) {
	apiInstance, _ := setupTestAPI(t)

	// Prepare multipart upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("path", "uploads/test.txt")
	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = io.WriteString(part, "Hello API upload")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	apiInstance.handleUpload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var uploadResp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("failed to decode upload response: %v", err)
	}
	if uploadResp["path"] != "uploads/test.txt" {
		t.Errorf("expected path 'uploads/test.txt', got %v", uploadResp["path"])
	}

	// Test List Files
	reqList := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	recList := httptest.NewRecorder()
	apiInstance.handleFiles(recList, reqList)

	if recList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/files, got %d", recList.Code)
	}

	var listResp struct {
		Count int `json:"count"`
		Files []struct {
			Path string `json:"path"`
		} `json:"files"`
	}
	if err := json.Unmarshal(recList.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	if listResp.Count != 1 || listResp.Files[0].Path != "uploads/test.txt" {
		t.Errorf("unexpected list response: %+v", listResp)
	}
}

func TestAPIDeleteAndPurge(t *testing.T) {
	apiInstance, eng := setupTestAPI(t)

	_, _ = eng.Save("to-delete.txt", strings.NewReader("delete me"))

	// Delete file
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/files/to-delete.txt", nil)
	recDel := httptest.NewRecorder()
	apiInstance.handleFileItem(recDel, reqDel)

	if recDel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on delete, got %d: %s", recDel.Code, recDel.Body.String())
	}

	// Purge cache
	purgePayload := `{"path":"*"}`
	reqPurge := httptest.NewRequest(http.MethodPost, "/api/purge", strings.NewReader(purgePayload))
	recPurge := httptest.NewRecorder()
	apiInstance.handlePurge(recPurge, reqPurge)

	if recPurge.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on purge, got %d: %s", recPurge.Code, recPurge.Body.String())
	}
}

func TestAPISignURL(t *testing.T) {
	apiInstance, _ := setupTestAPI(t)

	payload := `{"path":"media/video.mp4","ttl_seconds":3600}`
	req := httptest.NewRequest(http.MethodPost, "/api/sign-url", strings.NewReader(payload))
	rec := httptest.NewRecorder()

	apiInstance.handleSignURL(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on sign-url, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		URL       string `json:"url"`
		ExpiresAt int64  `json:"expires_at"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !strings.Contains(resp.URL, "token=") || !strings.Contains(resp.URL, "expires=") {
		t.Errorf("unexpected signed URL format: %s", resp.URL)
	}
}
