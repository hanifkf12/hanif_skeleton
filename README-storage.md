# Storage Package Documentation

## Overview

Storage package menyediakan abstraksi unified untuk berbagai storage backend (Local File System, Google Cloud Storage, AWS S3/MinIO). Implementasi mengikuti **Clean Architecture** dengan 1 contract interface dan 3 implementasi.

## Architecture

```
┌─────────────────────────────────────┐
│         UseCase Layer               │
├─────────────────────────────────────┤
│    Storage Interface (Contract)     │  ← Abstraction
├─────────────────────────────────────┤
│  Implementation Layer               │
│  ├─ Local File Storage              │
│  ├─ Google Cloud Storage (GCS)      │
│  └─ AWS S3 / MinIO                  │
└─────────────────────────────────────┘
```

## Storage Interface

```go
type Storage interface {
    Upload(ctx context.Context, path string, reader io.Reader, contentType string) error
    Download(ctx context.Context, path string) (io.ReadCloser, error)
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
    GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)
    List(ctx context.Context, prefix string) ([]FileInfo, error)
    Close() error
}
```

## Implementations

### 1. Local File Storage
- **Driver**: `local`
- **Use Case**: Development, small projects, simple file serving
- **File**: `pkg/storage/local.go`

**Features:**
- ✅ File system based storage
- ✅ Simple HTTP URL generation
- ✅ No external dependencies
- ✅ Fast for local development

### 2. Google Cloud Storage (GCS)
- **Driver**: `gcs`
- **Use Case**: Production apps on GCP, global CDN
- **File**: `pkg/storage/gcs.go`

**Features:**
- ✅ Scalable cloud storage
- ✅ Signed URL generation
- ✅ Global CDN integration
- ✅ IAM & access control

### 3. AWS S3 / MinIO
- **Driver**: `s3` or `minio`
- **Use Case**: Production apps on AWS, self-hosted MinIO
- **File**: `pkg/storage/s3.go`

**Features:**
- ✅ S3 compatible storage
- ✅ Presigned URL generation
- ✅ Works with MinIO (self-hosted S3)
- ✅ Cost-effective for large files

## Configuration

### Environment Variables

Add to `.env` file:

```bash
# Storage Configuration
# Options: local, gcs, s3, minio
STORAGE_DRIVER=local

# Local Storage Config
STORAGE_LOCAL_BASE_PATH=./storage
STORAGE_LOCAL_BASE_URL=http://localhost:9000/files

# Google Cloud Storage Config
STORAGE_GCS_BUCKET=your-gcs-bucket-name

# S3/MinIO Config
STORAGE_S3_ENDPOINT=                    # Optional, for MinIO
STORAGE_S3_REGION=us-east-1
STORAGE_S3_ACCESS_KEY_ID=your-key-id
STORAGE_S3_SECRET_ACCESS_KEY=your-secret-key
STORAGE_S3_BUCKET=your-bucket-name
STORAGE_S3_USE_SSL=true                 # false for local MinIO
```

### Config Struct

File: `pkg/config/storage.go`

```go
type Storage struct {
    Driver string

    // Local storage
    LocalBasePath string
    LocalBaseURL  string

    // GCS
    GCSBucket string

    // S3/MinIO
    S3Endpoint        string
    S3Region          string
    S3AccessKeyID     string
    S3SecretAccessKey string
    S3Bucket          string
    S3UseSSL          bool
}
```

## Bootstrap Registry

File: `internal/bootstrap/storage.go`

```go
// Initialize storage based on configuration
storage := bootstrap.RegistryStorage(cfg)
defer storage.Close()
```

**Registry automatically:**
- ✅ Selects implementation based on `STORAGE_DRIVER`
- ✅ Validates required configuration
- ✅ Provides sensible defaults
- ✅ Logs initialization status

## Usage Examples

### 1. Basic Upload

```go
package usecase

import (
    "bytes"
    "context"
    "github.com/hanifkf12/hanif_skeleton/pkg/storage"
)

func uploadExample(ctx context.Context, store storage.Storage) error {
    content := []byte("Hello, World!")
    reader := bytes.NewReader(content)
    
    err := store.Upload(ctx, "documents/hello.txt", reader, "text/plain")
    return err
}
```

### 2. File Upload from HTTP Request

See: `internal/usecase/upload_file.go`

```go
type uploadFile struct {
    storage storage.Storage
}

func (u *uploadFile) Serve(data appctx.Data) appctx.Response {
    // Get file from multipart form
    file, err := data.FiberCtx.FormFile("file")
    
    // Upload to storage
    err = u.storage.Upload(ctx, storagePath, reader, contentType)
    
    // Generate URL
    url, _ := u.storage.GetURL(ctx, storagePath, 1*time.Hour)
    
    return response
}
```

### 3. Download File

```go
func downloadExample(ctx context.Context, store storage.Storage) error {
    reader, err := store.Download(ctx, "documents/hello.txt")
    if err != nil {
        return err
    }
    defer reader.Close()
    
    // Read content
    content, err := io.ReadAll(reader)
    return err
}
```

### 4. Check if File Exists

```go
func existsExample(ctx context.Context, store storage.Storage) (bool, error) {
    exists, err := store.Exists(ctx, "documents/hello.txt")
    return exists, err
}
```

### 5. Delete File

```go
func deleteExample(ctx context.Context, store storage.Storage) error {
    err := store.Delete(ctx, "documents/hello.txt")
    return err
}
```

### 6. List Files

```go
func listExample(ctx context.Context, store storage.Storage) ([]storage.FileInfo, error) {
    files, err := store.List(ctx, "documents/")
    
    for _, file := range files {
        fmt.Printf("%s - %d bytes\n", file.Path, file.Size)
    }
    
    return files, err
}
```

### 7. Generate Signed/Presigned URL

```go
func urlExample(ctx context.Context, store storage.Storage) (string, error) {
    // URL valid for 1 hour
    url, err := store.GetURL(ctx, "documents/hello.txt", 1*time.Hour)
    return url, err
}
```

## Integration with Router

File: `internal/router/router.go`

```go
func (rtr *router) Route() {
    db := bootstrap.RegistryDatabase(rtr.cfg, false)
    storage := bootstrap.RegistryStorage(rtr.cfg)
    defer storage.Close()
    
    // File upload endpoint
    uploadFileUseCase := usecase.NewUploadFile(storage)
    rtr.fiber.Post("/upload", rtr.handle(
        handler.HttpRequest,
        uploadFileUseCase,
    ))
}
```

## Setup Instructions

### Local Storage (Development)

1. **Update .env:**
```bash
STORAGE_DRIVER=local
STORAGE_LOCAL_BASE_PATH=./storage
STORAGE_LOCAL_BASE_URL=http://localhost:9000/files
```

2. **Create storage directory:**
```bash
mkdir -p storage/uploads
```

3. **Serve files (optional):**
Add static file serving in Fiber:
```go
app.Static("/files", "./storage")
```

### Google Cloud Storage (Production)

1. **Create GCS bucket:**
```bash
gsutil mb gs://your-bucket-name
```

2. **Set IAM permissions:**
```bash
gsutil iam ch serviceAccount:YOUR_SERVICE_ACCOUNT:roles/storage.objectAdmin gs://your-bucket-name
```

3. **Update .env:**
```bash
STORAGE_DRIVER=gcs
STORAGE_GCS_BUCKET=your-bucket-name
```

4. **Authenticate:**
```bash
# Option 1: Service Account Key
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json

# Option 2: Application Default Credentials (recommended for GCP)
gcloud auth application-default login
```

### AWS S3 (Production)

1. **Create S3 bucket:**
```bash
aws s3 mb s3://your-bucket-name
```

2. **Create IAM user with S3 permissions:**
```bash
aws iam create-user --user-name app-storage
aws iam attach-user-policy --user-name app-storage --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess
aws iam create-access-key --user-name app-storage
```

3. **Update .env:**
```bash
STORAGE_DRIVER=s3
STORAGE_S3_REGION=us-east-1
STORAGE_S3_ACCESS_KEY_ID=YOUR_ACCESS_KEY
STORAGE_S3_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
STORAGE_S3_BUCKET=your-bucket-name
STORAGE_S3_USE_SSL=true
```

### MinIO (Self-Hosted)

1. **Run MinIO:**
```bash
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"
```

2. **Create bucket:**
```bash
# Via MinIO Console: http://localhost:9001
# Or via mc client:
mc alias set myminio http://localhost:9000 minioadmin minioadmin
mc mb myminio/your-bucket-name
```

3. **Update .env:**
```bash
STORAGE_DRIVER=minio
STORAGE_S3_ENDPOINT=http://localhost:9000
STORAGE_S3_REGION=us-east-1
STORAGE_S3_ACCESS_KEY_ID=minioadmin
STORAGE_S3_SECRET_ACCESS_KEY=minioadmin
STORAGE_S3_BUCKET=your-bucket-name
STORAGE_S3_USE_SSL=false
```

## Testing

### Test Upload Endpoint

```bash
# Upload file
curl -X POST http://localhost:9000/upload \
  -F "file=@/path/to/your/file.jpg"

# Response:
# {
#   "code": 200,
#   "data": {
#     "path": "uploads/abc-123-def.jpg",
#     "filename": "abc-123-def.jpg",
#     "original_name": "file.jpg",
#     "size": 123456,
#     "content_type": "image/jpeg",
#     "url": "http://localhost:9000/files/uploads/abc-123-def.jpg"
#   }
# }
```

### Unit Test Storage

```go
func TestUpload(t *testing.T) {
    ctx := context.Background()
    store, _ := storage.NewLocalStorage("./test_storage", "http://test")
    defer os.RemoveAll("./test_storage")
    
    content := []byte("test content")
    err := store.Upload(ctx, "test.txt", bytes.NewReader(content), "text/plain")
    
    assert.NoError(t, err)
    
    exists, _ := store.Exists(ctx, "test.txt")
    assert.True(t, exists)
}
```

## Best Practices

### 1. Use Context

Always pass context for cancellation and timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := storage.Upload(ctx, path, reader, contentType)
```

### 2. Close Reader

Always close ReadCloser from Download:

```go
reader, err := storage.Download(ctx, path)
if err != nil {
    return err
}
defer reader.Close() // Important!
```

### 3. Generate Unique Filenames

```go
import "github.com/google/uuid"

filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
```

### 4. Validate File Types

```go
allowedTypes := map[string]bool{
    "image/jpeg": true,
    "image/png":  true,
    "image/gif":  true,
}

if !allowedTypes[contentType] {
    return errors.New("file type not allowed")
}
```

### 5. Limit File Size

```go
const MaxFileSize = 10 * 1024 * 1024 // 10MB

if file.Size > MaxFileSize {
    return errors.New("file too large")
}
```

### 6. Organize Files by Prefix

```go
// Good structure
uploads/images/2024/01/file.jpg
uploads/documents/2024/01/file.pdf
uploads/avatars/user-123/avatar.png

// Use prefixes in List
images, _ := storage.List(ctx, "uploads/images/")
```

### 7. Clean Up on Error

```go
err := storage.Upload(ctx, path, reader, contentType)
if err != nil {
    return err
}

// If subsequent operation fails, clean up
if dbErr != nil {
    storage.Delete(ctx, path) // Rollback
    return dbErr
}
```

## Error Handling

```go
func handleStorageError(err error) appctx.Response {
    if err == nil {
        return *appctx.NewResponse().WithCode(200)
    }
    
    // Check for specific errors
    if strings.Contains(err.Error(), "not found") {
        return *appctx.NewResponse().
            WithCode(404).
            WithErrors("File not found")
    }
    
    if strings.Contains(err.Error(), "permission denied") {
        return *appctx.NewResponse().
            WithCode(403).
            WithErrors("Permission denied")
    }
    
    // Generic error
    return *appctx.NewResponse().
        WithCode(500).
        WithErrors("Storage error: " + err.Error())
}
```

## Migration Between Storage Providers

### Script to Migrate Files

```go
func migrateStorage(source, dest storage.Storage) error {
    ctx := context.Background()
    
    // List all files from source
    files, err := source.List(ctx, "")
    if err != nil {
        return err
    }
    
    for _, file := range files {
        // Download from source
        reader, err := source.Download(ctx, file.Path)
        if err != nil {
            log.Printf("Failed to download %s: %v", file.Path, err)
            continue
        }
        
        // Upload to destination
        content, _ := io.ReadAll(reader)
        reader.Close()
        
        err = dest.Upload(ctx, file.Path, bytes.NewReader(content), file.ContentType)
        if err != nil {
            log.Printf("Failed to upload %s: %v", file.Path, err)
            continue
        }
        
        log.Printf("Migrated: %s", file.Path)
    }
    
    return nil
}
```

## Performance Tips

### 1. Use Streaming for Large Files

```go
// Good - streaming
reader, _ := http.Get(url)
storage.Upload(ctx, path, reader.Body, contentType)

// Bad - load to memory
content, _ := io.ReadAll(reader.Body)
storage.Upload(ctx, path, bytes.NewReader(content), contentType)
```

### 2. Concurrent Uploads

```go
func uploadMultiple(files []string) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(files))
    
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            if err := uploadFile(f); err != nil {
                errChan <- err
            }
        }(file)
    }
    
    wg.Wait()
    close(errChan)
    
    // Check errors
    for err := range errChan {
        return err
    }
    return nil
}
```

### 3. Caching URLs

```go
// Cache signed URLs to avoid regenerating
type URLCache struct {
    cache map[string]CachedURL
    mu    sync.RWMutex
}

type CachedURL struct {
    URL     string
    Expires time.Time
}
```

## Monitoring & Logging

All storage operations are logged with:
- ✅ Operation name
- ✅ Storage path
- ✅ Bucket/container name
- ✅ Success/failure status
- ✅ OpenTelemetry trace context

Example log:
```json
{
  "level": "info",
  "event": "LocalStorage.Upload",
  "path": "uploads/file.jpg",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Security Considerations

### 1. Sanitize File Paths

```go
import "path/filepath"

// Prevent directory traversal
safePath := filepath.Clean(userProvidedPath)
if strings.Contains(safePath, "..") {
    return errors.New("invalid path")
}
```

### 2. Signed URL Expiry

```go
// Short expiry for sensitive files
url, _ := storage.GetURL(ctx, path, 5*time.Minute)

// Longer expiry for public assets
url, _ := storage.GetURL(ctx, path, 24*time.Hour)
```

### 3. Access Control

```go
// Check user permission before allowing access
if !user.CanAccessFile(path) {
    return errors.New("unauthorized")
}
```

## Troubleshooting

### Local Storage
- **Issue**: Permission denied
- **Fix**: `chmod -R 755 ./storage`

### GCS
- **Issue**: Authentication failed
- **Fix**: Set `GOOGLE_APPLICATION_CREDENTIALS` or use Application Default Credentials

### S3/MinIO
- **Issue**: SignatureDoesNotMatch
- **Fix**: Verify `ACCESS_KEY_ID` and `SECRET_ACCESS_KEY`

### Common
- **Issue**: Bucket not found
- **Fix**: Create bucket first using cloud console or CLI

## Summary

Storage package provides:
- ✅ **Unified interface** across different storage backends
- ✅ **Easy switching** between Local/GCS/S3/MinIO
- ✅ **Clean Architecture** pattern
- ✅ **Production ready** with logging, tracing, error handling
- ✅ **Well documented** with examples
- ✅ **Bootstrap integration** for easy setup

Choose your storage based on needs:
- **Local**: Development & testing
- **GCS**: Production on Google Cloud
- **S3**: Production on AWS
- **MinIO**: Self-hosted, cost-effective

---

**Files:**
- Contract: `pkg/storage/storage.go`
- Local: `pkg/storage/local.go`
- GCS: `pkg/storage/gcs.go`
- S3/MinIO: `pkg/storage/s3.go`
- Config: `pkg/config/storage.go`
- Bootstrap: `internal/bootstrap/storage.go`
- Example UseCase: `internal/usecase/upload_file.go`

