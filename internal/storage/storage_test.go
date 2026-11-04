package storage

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestStorageSaveAndOpen(t *testing.T) {
	tempDir := t.TempDir()
	eng, err := New(tempDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	content := []byte("Hello, Simple Static CDN!")
	relPath := "documents/greeting.txt"

	info, err := eng.Save(relPath, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("failed to save file: %v", err)
	}

	if info.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), info.Size)
	}
	if !strings.HasPrefix(info.ContentType, "text/plain") {
		t.Errorf("expected text/plain content type, got %s", info.ContentType)
	}
	if info.ETag == "" {
		t.Error("expected non-empty ETag")
	}

	file, openedInfo, err := eng.Open(relPath)
	if err != nil {
		t.Fatalf("failed to open saved file: %v", err)
	}
	defer file.Close()

	readData, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !bytes.Equal(readData, content) {
		t.Errorf("content mismatch: got %s, want %s", string(readData), string(content))
	}
	if openedInfo.Size != int64(len(content)) {
		t.Errorf("opened info size mismatch: got %d, want %d", openedInfo.Size, len(content))
	}
}

func TestStoragePathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	eng, err := New(tempDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	maliciousPaths := []string{
		"../secret.txt",
		"../../etc/passwd",
		"subdir/../../outside.txt",
		"/../../root.txt",
		"..\\windows\\system32\\cmd.exe",
	}

	for _, p := range maliciousPaths {
		_, err := eng.Save(p, strings.NewReader("bad"))
		if err != ErrInvalidPath {
			t.Errorf("expected ErrInvalidPath for %q, got %v", p, err)
		}

		_, _, err = eng.Open(p)
		if err != ErrInvalidPath {
			t.Errorf("expected ErrInvalidPath on open for %q, got %v", p, err)
		}
	}
}

func TestStorageListAndDelete(t *testing.T) {
	tempDir := t.TempDir()
	eng, err := New(tempDir)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	_, _ = eng.Save("assets/css/style.css", strings.NewReader("body { margin: 0; }"))
	_, _ = eng.Save("assets/js/main.js", strings.NewReader("console.log('cdn');"))
	_, _ = eng.Save("images/logo.png", strings.NewReader("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"))

	list, err := eng.List("")
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 files, got %d", len(list))
	}

	assetList, err := eng.List("assets/")
	if err != nil {
		t.Fatalf("failed to list filtered: %v", err)
	}
	if len(assetList) != 2 {
		t.Errorf("expected 2 asset files, got %d", len(assetList))
	}

	if err := eng.Delete("assets/css/style.css"); err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	if eng.Exists("assets/css/style.css") {
		t.Error("file should not exist after deletion")
	}

	err = eng.Delete("assets/css/style.css")
	if err != ErrFileNotFound {
		t.Errorf("expected ErrFileNotFound on second delete, got %v", err)
	}
}
