package storage

import (
	"context"
	"io"
	"time"
)

// Storage is the main interface for all storage implementations
type Storage interface {
	// Upload uploads a file to storage
	Upload(ctx context.Context, path string, reader io.Reader, contentType string) error

	// Download downloads a file from storage
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete deletes a file from storage
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, path string) (bool, error)

	// GetURL returns a public or signed URL for a file
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// List lists files in a directory/prefix
	List(ctx context.Context, prefix string) ([]FileInfo, error)

	// Close closes the storage connection
	Close() error
}

// FileInfo represents file metadata
type FileInfo struct {
	Path         string
	Size         int64
	LastModified time.Time
	ContentType  string
}

// UploadOptions provides additional options for upload
type UploadOptions struct {
	ContentType        string
	CacheControl       string
	ContentDisposition string
	Metadata           map[string]string
}
