package config

// Storage holds storage configuration
type Storage struct {
	Driver string `mapstructure:"STORAGE_DRIVER"` // local, gcs, s3, minio

	// Local storage config
	LocalBasePath string `mapstructure:"STORAGE_LOCAL_BASE_PATH"`
	LocalBaseURL  string `mapstructure:"STORAGE_LOCAL_BASE_URL"`

	// GCS config
	GCSBucket string `mapstructure:"STORAGE_GCS_BUCKET"`

	// S3/MinIO config
	S3Endpoint        string `mapstructure:"STORAGE_S3_ENDPOINT"`
	S3Region          string `mapstructure:"STORAGE_S3_REGION"`
	S3AccessKeyID     string `mapstructure:"STORAGE_S3_ACCESS_KEY_ID"`
	S3SecretAccessKey string `mapstructure:"STORAGE_S3_SECRET_ACCESS_KEY"`
	S3Bucket          string `mapstructure:"STORAGE_S3_BUCKET"`
	S3UseSSL          bool   `mapstructure:"STORAGE_S3_USE_SSL"`
}
