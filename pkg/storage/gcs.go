package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	gcpstorage "cloud.google.com/go/storage"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"google.golang.org/api/iterator"
)

// GCSStorage implements Storage interface for Google Cloud Storage
type GCSStorage struct {
	client     *gcpstorage.Client
	bucketName string
}

// NewGCSStorage creates a new GCS storage instance
func NewGCSStorage(ctx context.Context, bucketName string) (Storage, error) {
	client, err := gcpstorage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// Upload uploads a file to GCS
func (s *GCSStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
	lf := logger.NewFields("GCSStorage.Upload")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("path", path))

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(path)

	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType

	if _, err := io.Copy(writer, reader); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to upload file", lf)
		writer.Close()
		return fmt.Errorf("failed to upload file: %w", err)
	}

	if err := writer.Close(); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to close writer", lf)
		return fmt.Errorf("failed to close writer: %w", err)
	}

	logger.Info("File uploaded successfully to GCS", lf)
	return nil
}

// Download downloads a file from GCS
func (s *GCSStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	lf := logger.NewFields("GCSStorage.Download")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("path", path))

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(path)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to download file", lf)
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	logger.Info("File downloaded successfully from GCS", lf)
	return reader, nil
}

// Delete deletes a file from GCS
func (s *GCSStorage) Delete(ctx context.Context, path string) error {
	lf := logger.NewFields("GCSStorage.Delete")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("path", path))

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(path)

	if err := obj.Delete(ctx); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to delete file", lf)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("File deleted successfully from GCS", lf)
	return nil
}

// Exists checks if a file exists in GCS
func (s *GCSStorage) Exists(ctx context.Context, path string) (bool, error) {
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(path)

	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == gcpstorage.ErrObjectNotExist {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetURL returns a signed URL for the file
func (s *GCSStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	lf := logger.NewFields("GCSStorage.GetURL")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("path", path))
	lf.Append(logger.Any("expiry", expiry.String()))

	opts := &gcpstorage.SignedURLOptions{
		Scheme:  gcpstorage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	}

	url, err := s.client.Bucket(s.bucketName).SignedURL(path, opts)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to generate signed URL", lf)
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	logger.Info("Signed URL generated successfully", lf)
	return url, nil
}

// List lists files in GCS with a given prefix
func (s *GCSStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	lf := logger.NewFields("GCSStorage.List")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("prefix", prefix))

	bucket := s.client.Bucket(s.bucketName)
	query := &gcpstorage.Query{Prefix: prefix}

	var files []FileInfo
	it := bucket.Objects(ctx, query)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			lf.Append(logger.Any("error", err.Error()))
			logger.Error("Failed to list files", lf)
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		files = append(files, FileInfo{
			Path:         attrs.Name,
			Size:         attrs.Size,
			LastModified: attrs.Updated,
			ContentType:  attrs.ContentType,
		})
	}

	logger.Info("Files listed successfully from GCS", lf)
	return files, nil
}

// Close closes the GCS client
func (s *GCSStorage) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
