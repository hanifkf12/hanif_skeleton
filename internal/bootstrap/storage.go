package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/storage"
)

// RegistryStorage creates and returns a storage instance based on configuration
func RegistryStorage(cfg *config.Config) storage.Storage {
	lf := logger.NewFields("RegistryStorage")
	lf.Append(logger.Any("driver", cfg.Storage.Driver))

	switch cfg.Storage.Driver {
	case "local":
		return registryLocalStorage(cfg)
	case "gcs":
		return registryGCSStorage(cfg)
	case "s3", "minio":
		return registryS3Storage(cfg)
	default:
		logger.Error("Invalid storage driver, using local as default", lf)
		return registryLocalStorage(cfg)
	}
}

// registryLocalStorage creates local file storage
func registryLocalStorage(cfg *config.Config) storage.Storage {
	lf := logger.NewFields("RegistryLocalStorage")

	basePath := cfg.Storage.LocalBasePath
	if basePath == "" {
		basePath = "./storage"
		logger.Info("Using default local storage path: ./storage", lf)
	}

	baseURL := cfg.Storage.LocalBaseURL
	if baseURL == "" {
		baseURL = "http://localhost:9000/files"
		logger.Info("Using default local storage URL", lf)
	}

	lf.Append(logger.Any("base_path", basePath))
	lf.Append(logger.Any("base_url", baseURL))

	store, err := storage.NewLocalStorage(basePath, baseURL)
	if err != nil {
		log.Fatalf("Failed to initialize local storage: %v", err)
	}

	logger.Info("Local storage initialized successfully", lf)
	return store
}

// registryGCSStorage creates Google Cloud Storage
func registryGCSStorage(cfg *config.Config) storage.Storage {
	lf := logger.NewFields("RegistryGCSStorage")

	if cfg.Storage.GCSBucket == "" {
		log.Fatal("GCS bucket name is required")
	}

	lf.Append(logger.Any("bucket", cfg.Storage.GCSBucket))

	ctx := context.Background()
	store, err := storage.NewGCSStorage(ctx, cfg.Storage.GCSBucket)
	if err != nil {
		log.Fatalf("Failed to initialize GCS storage: %v", err)
	}

	logger.Info("GCS storage initialized successfully", lf)
	return store
}

// registryS3Storage creates S3/MinIO storage
func registryS3Storage(cfg *config.Config) storage.Storage {
	lf := logger.NewFields("RegistryS3Storage")

	// Validate required config
	if cfg.Storage.S3Bucket == "" {
		log.Fatal("S3 bucket name is required")
	}
	if cfg.Storage.S3AccessKeyID == "" || cfg.Storage.S3SecretAccessKey == "" {
		log.Fatal("S3 access key ID and secret access key are required")
	}

	// Default values
	region := cfg.Storage.S3Region
	if region == "" {
		region = "us-east-1"
	}

	s3Config := storage.S3Config{
		Endpoint:        cfg.Storage.S3Endpoint,
		Region:          region,
		AccessKeyID:     cfg.Storage.S3AccessKeyID,
		SecretAccessKey: cfg.Storage.S3SecretAccessKey,
		BucketName:      cfg.Storage.S3Bucket,
		UseSSL:          cfg.Storage.S3UseSSL,
	}

	lf.Append(logger.Any("bucket", s3Config.BucketName))
	lf.Append(logger.Any("region", s3Config.Region))
	lf.Append(logger.Any("endpoint", s3Config.Endpoint))

	store, err := storage.NewS3Storage(s3Config)
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}

	storageType := "S3"
	if cfg.Storage.S3Endpoint != "" {
		storageType = "MinIO"
	}

	logger.Info(fmt.Sprintf("%s storage initialized successfully", storageType), lf)
	return store
}
