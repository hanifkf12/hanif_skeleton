# Google Pub/Sub Implementation Guide

## Arsitektur

Implementasi Google Pub/Sub di project ini mengikuti arsitektur yang sama dengan HTTP handler:

```
┌─────────────────┐
│  Pub/Sub Topic  │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│  Subscription       │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Router/Manager     │  (internal/router/pubsub)
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Handler            │  (internal/handler)
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Consumer UseCase   │  (internal/usecase)
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Repository         │  (internal/repository)
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Database           │
└─────────────────────┘
```

## Components

### 1. Contract (`internal/usecase/contract/pubsub_consumer.go`)

Interface untuk Pub/Sub consumer, mirip dengan `UseCase` interface:

```go
type PubSubConsumer interface {
    Consume(data appctx.PubSubData) appctx.PubSubResponse
}
```

### 2. AppCtx Data Structures

**PubSubData** (`internal/appctx/pubsub_data.go`):
```go
type PubSubData struct {
    Ctx     context.Context
    Message *pubsub.Message
    Cfg     *config.Config
}
```

**PubSubResponse** (`internal/appctx/pubsub_response.go`):
```go
type PubSubResponse struct {
    Success bool
    Error   error
    Message string
}
```

### 3. Handler (`internal/handler/pubsub_handler.go`)

General handler untuk Pub/Sub, mirip dengan `HttpRequest`:

```go
func PubSubHandler(ctx context.Context, msg *pubsub.Message, consumer contract.PubSubConsumer, conf *config.Config) appctx.PubSubResponse
```

### 4. Router (`internal/router/pubsub/router.go`)

Router untuk mendaftarkan dan mengelola subscriptions:

- `RegisterSubscription(config SubscriptionConfig)` - Daftarkan subscription baru
- `Start(ctx context.Context) error` - Mulai consume messages
- `Stop() error` - Stop consumer gracefully

### 5. Command (`cmd/pubsub/pubsub.go`)

Command untuk menjalankan Pub/Sub worker:
```bash
go run main.go pubsub
```

## Cara Menambahkan Consumer Baru

### Step 1: Buat Consumer UseCase

Buat file baru di `internal/usecase/`, contoh `your_event_consumer.go`:

```go
package usecase

import (
    "encoding/json"
    "github.com/hanifkf12/hanif_skeleton/internal/appctx"
    "github.com/hanifkf12/hanif_skeleton/internal/entity"
    "github.com/hanifkf12/hanif_skeleton/internal/repository"
    "github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
    "github.com/hanifkf12/hanif_skeleton/pkg/logger"
    "github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type yourEventConsumer struct {
    repo repository.YourRepository
}

func NewYourEventConsumer(repo repository.YourRepository) contract.PubSubConsumer {
    return &yourEventConsumer{
        repo: repo,
    }
}

func (c *yourEventConsumer) Consume(data appctx.PubSubData) appctx.PubSubResponse {
    ctx, span := telemetry.StartSpan(data.Ctx, "yourEventConsumer.Consume")
    defer span.End()

    lf := logger.NewFields("YourEventConsumer").WithTrace(ctx)
    lf.Append(logger.Any("message_id", data.Message.ID))

    logger.Info("Processing your event message", lf)

    // Parse message data
    var req entity.YourRequest
    if err := json.Unmarshal(data.Message.Data, &req); err != nil {
        telemetry.SpanError(ctx, err)
        lf.Append(logger.Any("error", err.Error()))
        logger.Error("Failed to parse message data", lf)
        return *appctx.NewPubSubResponse().WithError(err)
    }

    // Business logic
    if err := c.repo.DoSomething(ctx, req); err != nil {
        telemetry.SpanError(ctx, err)
        lf.Append(logger.Any("error", err.Error()))
        logger.Error("Failed to process event", lf)
        return *appctx.NewPubSubResponse().WithError(err)
    }

    logger.Info("Event processed successfully", lf)
    return *appctx.NewPubSubResponse().WithMessage("Success")
}
```

### Step 2: Register Subscription di Command

Edit `cmd/pubsub/pubsub.go`:

```go
// Initialize repositories
yourRepository := yourRepo.NewYourRepository(db)

// Create Pub/Sub router
router := pubsubRouter.NewRouter(cfg, client)

// Register your subscription
router.RegisterSubscription(pubsubRouter.SubscriptionConfig{
    SubscriptionID: "your-subscription-id",
    Consumer:       usecase.NewYourEventConsumer(yourRepository),
    MaxConcurrent:  10, // adjust as needed
})
```

### Step 3: Jalankan Worker

```bash
# Set Google Cloud project ID
export GOOGLE_CLOUD_PROJECT=your-project-id

# Run the Pub/Sub worker
go run main.go pubsub
```

## Configuration

### Environment Variables

```bash
# Required
GOOGLE_CLOUD_PROJECT=your-gcp-project-id

# Optional (uses Application Default Credentials by default)
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
```

### Subscription Settings

Dalam `SubscriptionConfig`:

- `SubscriptionID`: ID subscription di Google Cloud Pub/Sub
- `Consumer`: Instance consumer usecase
- `MaxConcurrent`: Maksimal concurrent messages yang diproses (default: 10)

## Message Handling Flow

1. **Message diterima** dari subscription
2. **Router** meneruskan ke handler dengan consumer yang sesuai
3. **Handler** membuat `PubSubData` dan memanggil `consumer.Consume()`
4. **Consumer** memproses message:
   - Parse message data
   - Validasi
   - Call repository/business logic
   - Return response
5. **Router** meng-Ack/Nack message based on response:
   - `Success = true` → `msg.Ack()` (message dihapus dari queue)
   - `Success = false` → `msg.Nack()` (message akan di-retry)

## Error Handling

### Retryable Errors (Nack)

Jika consumer return error, message akan di-Nack dan Pub/Sub akan retry sesuai dengan subscription retry policy.

```go
return *appctx.NewPubSubResponse().WithError(err)
```

### Non-Retryable Errors (Ack)

Jika ingin Ack message meskipun ada error (untuk mencegah infinite retry):

```go
return *appctx.NewPubSubResponse().
    WithSuccess(true).
    WithMessage("Skipped due to invalid data")
```

## Logging & Tracing

Setiap consumer otomatis mendapat:

- **Trace context** dari OpenTelemetry
- **Structured logging** dengan message ID dan metadata
- **Correlation** antara logs dan traces

```go
lf := logger.NewFields("ConsumerName").WithTrace(ctx)
lf.Append(logger.Any("message_id", data.Message.ID))
lf.Append(logger.Any("custom_field", value))

logger.Info("Message", lf)
```

## Testing

### Unit Test Consumer

```go
func TestYourEventConsumer_Consume(t *testing.T) {
    // Mock repository
    mockRepo := &MockYourRepository{}
    consumer := NewYourEventConsumer(mockRepo)

    // Prepare test data
    messageData, _ := json.Marshal(entity.YourRequest{
        Field: "value",
    })

    msg := &pubsub.Message{
        ID:   "test-message-id",
        Data: messageData,
    }

    data := appctx.PubSubData{
        Ctx:     context.Background(),
        Message: msg,
        Cfg:     &config.Config{},
    }

    // Execute
    resp := consumer.Consume(data)

    // Assert
    assert.True(t, resp.Success)
    assert.Nil(t, resp.Error)
}
```

## Best Practices

### 1. Idempotency

Pastikan consumer idempoten (safe untuk retry):

```go
// Check if already processed
exists, err := c.repo.CheckExists(ctx, req.ID)
if exists {
    logger.Info("Message already processed, skipping", lf)
    return *appctx.NewPubSubResponse().WithSuccess(true)
}
```

### 2. Timeout

Set timeout context untuk long-running operations:

```go
ctx, cancel := context.WithTimeout(data.Ctx, 30*time.Second)
defer cancel()
```

### 3. Dead Letter Queue

Konfigurasi DLQ di Google Cloud untuk messages yang gagal berulang kali:

```bash
gcloud pubsub subscriptions update your-subscription \
    --dead-letter-topic=your-dlq-topic \
    --max-delivery-attempts=5
```

### 4. Batch Processing

Untuk high-throughput, gunakan batch processing atau increase `MaxConcurrent`:

```go
router.RegisterSubscription(pubsubRouter.SubscriptionConfig{
    SubscriptionID: "high-volume-subscription",
    Consumer:       consumer,
    MaxConcurrent:  50, // Increase for more parallel processing
})
```

## Monitoring

### Metrics yang Perlu Dimonitor

1. **Message Processing Rate**
2. **Error Rate**
3. **Message Age** (time between publish dan consume)
4. **Ack/Nack Ratio**
5. **Processing Duration**

### Health Check

Consumer akan log:
- Startup confirmation
- Message received/processed
- Errors dengan full context
- Graceful shutdown

## Deployment

### Docker

Tambahkan service baru di `docker-compose.yml`:

```yaml
pubsub-worker:
  build: .
  command: ["./app", "pubsub"]
  environment:
    - GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT}
    - GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-key.json
  volumes:
    - ./secrets:/secrets
  restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pubsub-worker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: pubsub-worker
  template:
    metadata:
      labels:
        app: pubsub-worker
    spec:
      containers:
      - name: worker
        image: your-image:latest
        command: ["./app", "pubsub"]
        env:
        - name: GOOGLE_CLOUD_PROJECT
          value: "your-project-id"
```

## Troubleshooting

### Messages tidak terproses

1. Check subscription exists: `gcloud pubsub subscriptions list`
2. Check subscription ack deadline
3. Check logs untuk error
4. Verify GOOGLE_CLOUD_PROJECT environment variable

### High retry rate

1. Check consumer logic untuk errors
2. Verify message format
3. Check database connectivity
4. Review retry policy

### Memory issues

1. Reduce `MaxConcurrent`
2. Implement message batching
3. Optimize consumer logic
4. Check for memory leaks

## Examples

Lihat implementasi lengkap di:
- `internal/usecase/user_created_consumer.go` - User creation from Pub/Sub
- `internal/usecase/campaign_created_consumer.go` - Campaign creation from Pub/Sub

## Commands

```bash
# Run HTTP server
go run main.go http

# Run Pub/Sub worker
go run main.go pubsub

# Run database migration
go run main.go db:migrate

# Build binary
go build -o app main.go

# Run binary
./app pubsub
```

