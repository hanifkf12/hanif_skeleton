package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// S3Storage implements Storage interface for AWS S3 / MinIO
type S3Storage struct {
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucketName string
	region     string
	endpoint   string
}

// S3Config holds S3/MinIO configuration
type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

// NewS3Storage creates a new S3/MinIO storage instance
func NewS3Storage(config S3Config) (Storage, error) {
	// Configure AWS session
	awsConfig := &aws.Config{
		Region:           aws.String(config.Region),
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		S3ForcePathStyle: aws.Bool(true), // Required for MinIO
	}

	// Set endpoint if provided (for MinIO)
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
	}

	// Disable SSL if specified (useful for local MinIO)
	if !config.UseSSL {
		awsConfig.DisableSSL = aws.Bool(true)
	}

	// Create session
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := s3.New(sess)

	return &S3Storage{
		client:     client,
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		bucketName: config.BucketName,
		region:     config.Region,
		endpoint:   config.Endpoint,
	}, nil
}

// Upload uploads a file to S3/MinIO
func (s *S3Storage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
	lf := logger.NewFields("S3Storage.Upload")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("path", path))

	input := &s3manager.UploadInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(path),
		Body:        reader,
		ContentType: aws.String(contentType),
	}

	_, err := s.uploader.UploadWithContext(ctx, input)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to upload file", lf)
		return fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Info("File uploaded successfully to S3", lf)
	return nil
}

// Download downloads a file from S3/MinIO
func (s *S3Storage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	lf := logger.NewFields("S3Storage.Download")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("path", path))

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	}

	result, err := s.client.GetObjectWithContext(ctx, input)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to download file", lf)
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	logger.Info("File downloaded successfully from S3", lf)
	return result.Body, nil
}

// Delete deletes a file from S3/MinIO
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	lf := logger.NewFields("S3Storage.Delete")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("path", path))

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	}

	_, err := s.client.DeleteObjectWithContext(ctx, input)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to delete file", lf)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("File deleted successfully from S3", lf)
	return nil
}

// Exists checks if a file exists in S3/MinIO
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	}

	_, err := s.client.HeadObjectWithContext(ctx, input)
	if err != nil {
		// Check if error is "not found"
		if _, ok := err.(s3.RequestFailure); ok {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetURL returns a presigned URL for the file
func (s *S3Storage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	lf := logger.NewFields("S3Storage.GetURL")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("path", path))
	lf.Append(logger.Any("expiry", expiry.String()))

	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})

	url, err := req.Presign(expiry)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to generate presigned URL", lf)
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	logger.Info("Presigned URL generated successfully", lf)
	return url, nil
}

// List lists files in S3/MinIO with a given prefix
func (s *S3Storage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	lf := logger.NewFields("S3Storage.List")
	lf.Append(logger.Any("bucket", s.bucketName))
	lf.Append(logger.Any("prefix", prefix))

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	}

	var files []FileInfo

	err := s.client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			files = append(files, FileInfo{
				Path:         *obj.Key,
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
			})
		}
		return true
	})

	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to list files", lf)
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	logger.Info("Files listed successfully from S3", lf)
	return files, nil
}

// Close closes the S3 client (no-op)
func (s *S3Storage) Close() error {
	return nil
}
