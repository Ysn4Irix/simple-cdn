package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidPath  = errors.New("invalid or illegal file path")
	ErrFileNotFound = errors.New("file not found")
)

type FileInfo struct {
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	ETag        string    `json:"etag"`
	ContentType string    `json:"content_type"`
	SHA256      string    `json:"sha256"`
}

type Engine struct {
	baseDir string
}

func New(baseDir string) (*Engine, error) {
	absPath, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storage path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to initialize storage directory: %w", err)
	}

	return &Engine{baseDir: absPath}, nil
}

func (e *Engine) safePath(relPath string) (string, error) {
	if strings.Contains(relPath, "\x00") {
		return "", ErrInvalidPath
	}

	normalized := filepath.ToSlash(relPath)
	parts := strings.Split(normalized, "/")
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == ".." || trimmed == "." {
			return "", ErrInvalidPath
		}
	}

	cleaned := filepath.Clean(filepath.FromSlash(normalized))
	if filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, `\`) || strings.HasPrefix(cleaned, "/") {
		return "", ErrInvalidPath
	}

	target := filepath.Join(e.baseDir, cleaned)
	rel, err := filepath.Rel(e.baseDir, target)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return "", ErrInvalidPath
	}

	return target, nil
}

func (e *Engine) Save(relPath string, src io.Reader) (*FileInfo, error) {
	fullPath, err := e.safePath(relPath)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, ".upload-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpName := tmpFile.Name()

	hasher := sha256.New()
	multiWriter := io.MultiWriter(tmpFile, hasher)

	sniffBuf := make([]byte, 512)
	n, err := io.ReadFull(src, sniffBuf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		tmpFile.Close()
		os.Remove(tmpName)
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	var size int64
	if n > 0 {
		written, wErr := multiWriter.Write(sniffBuf[:n])
		if wErr != nil {
			tmpFile.Close()
			os.Remove(tmpName)
			return nil, fmt.Errorf("failed to write header: %w", wErr)
		}
		size += int64(written)
	}

	copied, copyErr := io.Copy(multiWriter, src)
	if copyErr != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return nil, fmt.Errorf("failed to save file payload: %w", copyErr)
	}
	size += copied

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return nil, fmt.Errorf("failed to sync file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err := os.Rename(tmpName, fullPath); err != nil {
		os.Remove(tmpName)
		return nil, fmt.Errorf("failed to commit file: %w", err)
	}

	stat, statErr := os.Stat(fullPath)
	if statErr != nil {
		return nil, fmt.Errorf("failed to stat committed file: %w", statErr)
	}

	hashSum := hex.EncodeToString(hasher.Sum(nil))
	etag := generateETag(stat.ModTime(), stat.Size())

	contentType := mime.TypeByExtension(filepath.Ext(fullPath))
	if contentType == "" && n > 0 {
		contentType = http.DetectContentType(sniffBuf[:n])
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	cleanRel := strings.TrimPrefix(filepath.ToSlash(fullPath[len(e.baseDir):]), "/")

	return &FileInfo{
		Path:        cleanRel,
		Size:        size,
		ModTime:     stat.ModTime().UTC(),
		ETag:        etag,
		ContentType: contentType,
		SHA256:      hashSum,
	}, nil
}

func (e *Engine) Open(relPath string) (*os.File, *FileInfo, error) {
	fullPath, err := e.safePath(relPath)
	if err != nil {
		return nil, nil, err
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, ErrFileNotFound
		}
		return nil, nil, err
	}

	if stat.IsDir() {
		return nil, nil, ErrFileNotFound
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, err
	}

	contentType := mime.TypeByExtension(filepath.Ext(fullPath))
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := file.ReadAt(buf, 0)
		if n > 0 {
			contentType = http.DetectContentType(buf[:n])
		} else {
			contentType = "application/octet-stream"
		}
	}

	etag := generateETag(stat.ModTime(), stat.Size())
	cleanRel := strings.TrimPrefix(filepath.ToSlash(fullPath[len(e.baseDir):]), "/")

	info := &FileInfo{
		Path:        cleanRel,
		Size:        stat.Size(),
		ModTime:     stat.ModTime().UTC(),
		ETag:        etag,
		ContentType: contentType,
	}

	return file, info, nil
}

func (e *Engine) Delete(relPath string) error {
	fullPath, err := e.safePath(relPath)
	if err != nil {
		return err
	}

	err = os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}
	return nil
}

func (e *Engine) List(prefix string) ([]*FileInfo, error) {
	var files []*FileInfo

	err := filepath.Walk(e.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel := strings.TrimPrefix(filepath.ToSlash(path[len(e.baseDir):]), "/")
		if prefix != "" && !strings.HasPrefix(rel, prefix) {
			return nil
		}

		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		etag := generateETag(info.ModTime(), info.Size())

		files = append(files, &FileInfo{
			Path:        rel,
			Size:        info.Size(),
			ModTime:     info.ModTime().UTC(),
			ETag:        etag,
			ContentType: contentType,
		})
		return nil
	})

	return files, err
}

func generateETag(modTime time.Time, size int64) string {
	return fmt.Sprintf(`"%x-%x"`, modTime.UnixNano(), size)
}

func (e *Engine) Exists(relPath string) bool {
	fullPath, err := e.safePath(relPath)
	if err != nil {
		return false
	}
	stat, err := os.Stat(fullPath)
	return err == nil && !stat.IsDir()
}
