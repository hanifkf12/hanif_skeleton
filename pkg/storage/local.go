package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// LocalStorage implements Storage interface for local file system
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath, baseURL string) (Storage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// Upload uploads a file to local storage
func (s *LocalStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
	lf := logger.NewFields("LocalStorage.Upload")
	lf.Append(logger.Any("path", path))

	fullPath := filepath.Join(s.basePath, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create directory", lf)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to create file", lf)
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	if _, err := io.Copy(file, reader); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to write file", lf)
		return fmt.Errorf("failed to write file: %w", err)
	}

	logger.Info("File uploaded successfully", lf)
	return nil
}

// Download downloads a file from local storage
func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	lf := logger.NewFields("LocalStorage.Download")
	lf.Append(logger.Any("path", path))

	fullPath := filepath.Join(s.basePath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to open file", lf)
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	logger.Info("File downloaded successfully", lf)
	return file, nil
}

// Delete deletes a file from local storage
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	lf := logger.NewFields("LocalStorage.Delete")
	lf.Append(logger.Any("path", path))

	fullPath := filepath.Join(s.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to delete file", lf)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("File deleted successfully", lf)
	return nil
}

// Exists checks if a file exists
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetURL returns a URL for the file
func (s *LocalStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	// For local storage, expiry is not applicable
	return fmt.Sprintf("%s/%s", s.baseURL, path), nil
}

// List lists files in a directory
func (s *LocalStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	lf := logger.NewFields("LocalStorage.List")
	lf.Append(logger.Any("prefix", prefix))

	fullPath := filepath.Join(s.basePath, prefix)
	var files []FileInfo

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, _ := filepath.Rel(s.basePath, path)
			files = append(files, FileInfo{
				Path:         relPath,
				Size:         info.Size(),
				LastModified: info.ModTime(),
			})
		}
		return nil
	})

	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to list files", lf)
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	logger.Info("Files listed successfully", lf)
	return files, nil
}

// Close closes the storage connection (no-op for local storage)
func (s *LocalStorage) Close() error {
	return nil
}
