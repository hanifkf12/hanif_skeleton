# Job Queue Package Documentation

## Overview

Job Queue package menyediakan sistem background job processing menggunakan **Asynq** (Redis-based). Terintegrasi dengan Clean Architecture, job dapat mengakses **repository**, **HTTP client**, dan **cache** untuk operasi yang kompleks.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│            HTTP/UseCase Layer                       │
│          Enqueue Jobs via Queue                     │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│         Queue Interface (Abstraction)               │
├─────────────────────────────────────────────────────┤
│         Asynq Client (Redis-based)                  │
│         - Enqueue jobs                              │
│         - Schedule jobs                             │
│         - Priority queues                           │
└─────────────────────────────────────────────────────┘
             │
             ▼ (Redis)
             │
┌─────────────────────────────────────────────────────┐
│         Asynq Worker (Background)                   │
│         - Process jobs                              │
│         - Retry failed jobs                         │
│         - Concurrent processing                     │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│         Job Handlers (internal/jobs/)               │
│         - Access Repository                         │
│         - Access HTTP Client                        │
│         - Access Cache                              │
│         - Business Logic                            │
└─────────────────────────────────────────────────────┘
```

## Features

### ✅ Background Processing
- Asynchronous job execution
- Non-blocking operations
- Scalable workers

### ✅ Scheduling
- Immediate execution
- Delayed execution
- Scheduled at specific time

### ✅ Retry Mechanism
- Automatic retry on failure
- Configurable max retries
- Exponential backoff

### ✅ Priority Queues
- critical, default, low queues
- Weighted processing

### ✅ Unique Jobs
- Prevent duplicate jobs
- TTL-based deduplication

### ✅ Clean Architecture
- Jobs can access repositories
- Jobs can call external APIs
- Jobs can use cache

## Queue Interface

```go
type Queue interface {
    Enqueue(ctx, jobType, payload)
    EnqueueWithDelay(ctx, jobType, payload, delay)
    EnqueueAt(ctx, jobType, payload, processAt)
    EnqueueWithOptions(ctx, jobType, payload, opts)
    Close()
}
```

## Configuration

### Environment Variables

Add to `.env`:

```bash
# Queue Configuration
QUEUE_DRIVER=asynq         # Job queue driver
QUEUE_HOST=localhost       # Redis host
QUEUE_PORT=6379           # Redis port
QUEUE_PASSWORD=           # Redis password (optional)
QUEUE_DB=1                # Redis database (separate from cache)
```

### Config Struct

```go
type Queue struct {
    Driver   string // asynq
    Host     string
    Port     int
    Password string
    DB       int
}
```

## Bootstrap Registry

```go
// Initialize queue client
queue := bootstrap.RegistryQueue(cfg)
defer queue.Close()
```

---

## Creating Jobs

### 1. Job Structure

Jobs are stored in `internal/jobs/` directory.

**File:** `internal/jobs/send_email_job.go`

```go
package jobs

import (
    "context"
    "encoding/json"
    "github.com/hanifkf12/hanif_skeleton/pkg/queue"
)

type SendEmailJob struct {
    userRepo   repository.UserRepository
    httpClient httpclient.HTTPClient
    cache      cache.Cache
}

type SendEmailPayload struct {
    UserID  int64  `json:"user_id"`
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

func NewSendEmailJob(
    userRepo repository.UserRepository,
    httpClient httpclient.HTTPClient,
    cache cache.Cache,
) queue.JobHandler {
    job := &SendEmailJob{
        userRepo:   userRepo,
        httpClient: httpClient,
        cache:      cache,
    }
    return job.Handle
}

func (j *SendEmailJob) Handle(ctx context.Context, payload []byte) error {
    // Unmarshal payload
    var data SendEmailPayload
    json.Unmarshal(payload, &data)
    
    // Access repository
    users, _ := j.userRepo.GetUsers(ctx)
    
    // Access cache
    cacheKey := cache.NewCacheKey("email").Build(...)
    j.cache.Set(ctx, cacheKey, "sent", 1*time.Hour)
    
    // Call external API
    resp, _ := j.httpClient.Post(ctx, emailAPI, emailData, headers)
    
    return nil
}
```

### 2. Job Types

Define job types in `internal/jobs/types.go`:

```go
const (
    JobTypeSendEmail = "email:send"
    JobTypeGenerateReport = "report:generate"
    JobTypeSyncData = "data:sync"
)
```

---

## Enqueuing Jobs

### 1. Immediate Execution

```go
func (u *usecase) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()
    
    payload := jobs.SendEmailPayload{
        UserID:  123,
        To:      "user@example.com",
        Subject: "Welcome",
        Body:    "Welcome to our platform!",
    }
    
    // Enqueue job
    err := u.queue.Enqueue(ctx, jobs.JobTypeSendEmail, payload)
    if err != nil {
        return *appctx.NewResponse().
            WithCode(fiber.StatusInternalServerError).
            WithErrors("Failed to enqueue job")
    }
    
    return *appctx.NewResponse().
        WithData(map[string]string{"status": "queued"})
}
```

### 2. Delayed Execution

```go
// Process after 5 minutes
delay := 5 * time.Minute
err := queue.EnqueueWithDelay(ctx, jobs.JobTypeSendEmail, payload, delay)
```

### 3. Scheduled Execution

```go
// Process at specific time
processAt := time.Now().Add(24 * time.Hour)
err := queue.EnqueueAt(ctx, jobs.JobTypeSendEmail, payload, processAt)
```

### 4. With Options

```go
err := queue.EnqueueWithOptions(ctx, jobs.JobTypeSyncData, payload, &queue.EnqueueOptions{
    Queue:     "critical",        // Priority queue
    MaxRetry:  5,                // Retry up to 5 times
    Timeout:   30 * time.Second, // Job timeout
    Unique:    true,             // Prevent duplicates
    UniqueTTL: 5 * time.Minute,  // Deduplication window
})
```

---

## Worker Setup

### 1. Register Jobs

**File:** `cmd/worker/worker.go`

```go
func runWorker(cmd *cobra.Command, args []string) {
    // Initialize dependencies
    db := bootstrap.RegistryDatabase(cfg, false)
    cache := bootstrap.RegistryCache(cfg)
    httpClient := bootstrap.RegistryHTTPClient(cfg)
    userRepo := userRepo.NewUserRepository(db)
    
    // Create job registry
    registry := queue.NewJobRegistry()
    
    // Register jobs
    registry.Register(
        jobs.JobTypeSendEmail,
        jobs.NewSendEmailJob(userRepo, httpClient, cache),
    )
    
    registry.Register(
        jobs.JobTypeGenerateReport,
        jobs.NewGenerateReportJob(userRepo, cache),
    )
    
    // Create Asynq server
    srv := asynq.NewServer(redisOpt, asynq.Config{
        Concurrency: 10,
        Queues: map[string]int{
            "critical": 6,  // 60%
            "default":  3,  // 30%
            "low":      1,  // 10%
        },
    })
    
    // Start worker
    srv.Run(mux)
}
```

### 2. Run Worker

```bash
# Start worker
./app worker

# Output:
# [INFO] Starting job queue worker
# [INFO] Registering job handlers
# [INFO] Asynq worker started
```

---

## Example Jobs

### 1. Send Email Job

**Purpose:** Send emails via external email service

**Dependencies:**
- UserRepository - Get user data
- HTTPClient - Call email API
- Cache - Prevent duplicate sends

**File:** `internal/jobs/send_email_job.go`

```go
func (j *SendEmailJob) Handle(ctx context.Context, payload []byte) error {
    var data SendEmailPayload
    json.Unmarshal(payload, &data)
    
    // Check cache (prevent duplicates)
    cacheKey := cache.NewCacheKey("email").Build(...)
    if exists, _ := j.cache.Exists(ctx, cacheKey); exists {
        return nil // Already sent
    }
    
    // Get user from DB
    users, _ := j.userRepo.GetUsers(ctx)
    
    // Call email API
    resp, err := j.httpClient.Post(ctx, emailAPI, emailData, headers)
    if err != nil {
        return err
    }
    
    // Cache result
    j.cache.Set(ctx, cacheKey, "sent", 1*time.Hour)
    
    return nil
}
```

### 2. Generate Report Job

**Purpose:** Generate reports (PDF/CSV/Excel)

**Dependencies:**
- UserRepository - Get data
- Cache - Store result

**File:** `internal/jobs/generate_report_job.go`

```go
func (j *GenerateReportJob) Handle(ctx context.Context, payload []byte) error {
    var data GenerateReportPayload
    json.Unmarshal(payload, &data)
    
    // Get data from DB
    users, _ := j.userRepo.GetUsers(ctx)
    
    // Generate report (simulate)
    time.Sleep(2 * time.Second)
    
    reportData := map[string]interface{}{
        "report_type":  data.ReportType,
        "generated_at": time.Now(),
        "total_users":  len(users),
        "status":       "completed",
    }
    
    // Cache report
    cacheKey := cache.NewCacheKey("report").Build(...)
    reportJSON, _ := json.Marshal(reportData)
    j.cache.Set(ctx, cacheKey, reportJSON, 24*time.Hour)
    
    return nil
}
```

### 3. Sync Data Job

**Purpose:** Sync data with external service

**Dependencies:**
- HTTPClient - Call external API
- Cache - Track sync status

**File:** `internal/jobs/sync_data_job.go`

```go
func (j *SyncDataJob) Handle(ctx context.Context, payload []byte) error {
    var data SyncDataPayload
    json.Unmarshal(payload, &data)
    
    // Prepare sync payload
    syncPayload := map[string]interface{}{
        "entity_type": data.EntityType,
        "entity_id":   data.EntityID,
        "action":      data.Action,
        "timestamp":   time.Now(),
    }
    
    // Call external API
    resp, err := j.httpClient.Post(ctx, syncAPI, syncPayload, headers)
    if err != nil {
        return err
    }
    
    // Update sync status
    j.cache.Set(ctx, syncStatusKey, "synced", 1*time.Hour)
    
    return nil
}
```

---

## Usage in HTTP Endpoints

### Example: Enqueue Email on User Registration

```go
type registerUser struct {
    userRepo repository.UserRepository
    queue    queue.Queue
}

func (u *registerUser) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()
    
    var req RegisterRequest
    data.FiberCtx.BodyParser(&req)
    
    // Create user in database
    user, err := u.userRepo.CreateUser(ctx, req)
    if err != nil {
        return *appctx.NewResponse().
            WithCode(fiber.StatusInternalServerError).
            WithErrors("Failed to create user")
    }
    
    // Enqueue welcome email (async)
    emailPayload := jobs.SendEmailPayload{
        UserID:  user.ID,
        To:      user.Email,
        Subject: "Welcome!",
        Body:    "Welcome to our platform!",
    }
    
    u.queue.Enqueue(ctx, jobs.JobTypeSendEmail, emailPayload)
    
    // Return immediately (email sent in background)
    return *appctx.NewResponse().WithData(user)
}
```

### Example: Generate Report on Demand

```go
type requestReport struct {
    queue queue.Queue
}

func (u *requestReport) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()
    
    var req ReportRequest
    data.FiberCtx.BodyParser(&req)
    
    // Enqueue report generation with delay
    payload := jobs.GenerateReportPayload{
        ReportType: req.Type,
        UserID:     getUserID(ctx),
        StartDate:  req.StartDate,
        EndDate:    req.EndDate,
    }
    
    // Process in 5 minutes (give time for data to settle)
    u.queue.EnqueueWithDelay(
        ctx,
        jobs.JobTypeGenerateReport,
        payload,
        5*time.Minute,
    )
    
    return *appctx.NewResponse().WithData(map[string]string{
        "message": "Report will be generated in 5 minutes",
        "status":  "queued",
    })
}
```

---

## Priority Queues

### Queue Configuration

```go
asynq.Config{
    Queues: map[string]int{
        "critical": 6,  // 60% of workers
        "default":  3,  // 30% of workers
        "low":      1,  // 10% of workers
    },
}
```

### Enqueue to Specific Queue

```go
// Critical queue (highest priority)
queue.EnqueueWithOptions(ctx, jobType, payload, &queue.EnqueueOptions{
    Queue: "critical",
})

// Default queue
queue.Enqueue(ctx, jobType, payload) // Uses "default"

// Low priority queue
queue.EnqueueWithOptions(ctx, jobType, payload, &queue.EnqueueOptions{
    Queue: "low",
})
```

**Use cases:**
- **critical**: Payment processing, order confirmation
- **default**: Email sending, notifications
- **low**: Report generation, cleanup tasks

---

## Retry Mechanism

### Automatic Retry

Jobs are automatically retried on failure:

```go
queue.EnqueueWithOptions(ctx, jobType, payload, &queue.EnqueueOptions{
    MaxRetry: 5, // Retry up to 5 times
})
```

**Retry schedule (exponential backoff):**
- Attempt 1: Immediately
- Attempt 2: After 30 seconds
- Attempt 3: After 1 minute
- Attempt 4: After 5 minutes
- Attempt 5: After 10 minutes

### Handle Permanent Failures

```go
func (j *job) Handle(ctx context.Context, payload []byte) error {
    // Process job
    err := doWork()
    
    if isPermanentError(err) {
        // Don't retry permanent errors
        logger.Error("Permanent error, not retrying", ...)
        return nil // Return nil to mark as complete
    }
    
    // Transient error - will retry
    return err
}
```

---

## Monitoring

### Logs

Worker automatically logs:

```
[INFO] Job enqueued successfully
  job_type: email:send
  task_id: abc123
  queue: default

[INFO] Processing job
  job_type: email:send
  task_id: abc123

[INFO] Job completed successfully
  job_type: email:send
  duration: 234ms

[ERROR] Job failed
  job_type: email:send
  error: connection timeout
  retry_attempt: 1
```

### Asynq Web UI

Monitor jobs via Asynq web UI:

```bash
# Install asynqmon
go install github.com/hibiken/asynqmon@latest

# Run web UI
asynqmon --redis-addr=localhost:6379

# Open browser
http://localhost:8080
```

**Features:**
- View active jobs
- View scheduled jobs
- View failed jobs
- Retry failed jobs
- Delete jobs

---

## Testing

### Mock Queue

```go
type MockQueue struct {
    enqueuedJobs []EnqueuedJob
}

func (m *MockQueue) Enqueue(ctx context.Context, jobType string, payload interface{}) error {
    m.enqueuedJobs = append(m.enqueuedJobs, EnqueuedJob{
        Type:    jobType,
        Payload: payload,
    })
    return nil
}

// Test
func TestRegisterUser(t *testing.T) {
    mockQueue := &MockQueue{}
    usecase := NewRegisterUser(repo, mockQueue)
    
    result := usecase.Serve(data)
    
    assert.Equal(t, 1, len(mockQueue.enqueuedJobs))
    assert.Equal(t, jobs.JobTypeSendEmail, mockQueue.enqueuedJobs[0].Type)
}
```

### Test Job Handler

```go
func TestSendEmailJob(t *testing.T) {
    mockHTTPClient := httpclient.NewMockClient()
    mockCache := cache.NewMemoryCache()
    mockRepo := &MockUserRepository{}
    
    job := NewSendEmailJob(mockRepo, mockHTTPClient, mockCache)
    
    payload := jobs.SendEmailPayload{
        UserID:  123,
        To:      "test@example.com",
        Subject: "Test",
        Body:    "Test body",
    }
    
    payloadBytes, _ := json.Marshal(payload)
    
    err := job(context.Background(), payloadBytes)
    
    assert.NoError(t, err)
}
```

---

## Best Practices

### ✅ DO:
- Keep jobs idempotent (safe to retry)
- Use unique jobs for critical operations
- Log job execution details
- Handle errors gracefully
- Use appropriate queues (critical/default/low)
- Set reasonable timeouts
- Monitor job failures

### ❌ DON'T:
- Don't put large data in payload (use ID and fetch from DB)
- Don't make jobs depend on each other
- Don't use jobs for real-time operations
- Don't ignore job failures
- Don't set infinite retries

---

## Common Patterns

### 1. Email After User Action

```go
// HTTP handler
user := createUser(req)
queue.Enqueue(ctx, jobs.JobTypeSendEmail, emailPayload)
return user
```

### 2. Scheduled Tasks

```go
// Daily report at 8 AM
tomorrow8AM := time.Now().Add(24*time.Hour).Truncate(24*time.Hour).Add(8*time.Hour)
queue.EnqueueAt(ctx, jobs.JobTypeGenerateReport, payload, tomorrow8AM)
```

### 3. Webhook Processing

```go
// Enqueue webhook for async processing
queue.EnqueueWithOptions(ctx, jobs.JobTypeProcessWebhook, webhook, &queue.EnqueueOptions{
    Queue:    "critical",
    MaxRetry: 3,
    Timeout:  10 * time.Second,
})
```

### 4. Data Cleanup

```go
// Daily cleanup at midnight
midnight := time.Now().Add(24*time.Hour).Truncate(24*time.Hour)
queue.EnqueueAt(ctx, jobs.JobTypeCleanupExpired, payload, midnight)
```

---

## Troubleshooting

### Issue: Jobs not processing

**Solution:** Check if worker is running

```bash
./app worker
```

### Issue: Redis connection failed

**Solution:** Verify Redis is running

```bash
redis-cli ping
# Should return: PONG
```

### Issue: Jobs stuck in queue

**Solution:** Check worker logs for errors

```bash
./app worker 2>&1 | grep ERROR
```

---

## Summary

Job Queue package provides:
- ✅ **Background job processing** with Asynq
- ✅ **Redis-based** queue
- ✅ **Scheduling** (immediate, delayed, scheduled)
- ✅ **Retry mechanism** with exponential backoff
- ✅ **Priority queues** (critical, default, low)
- ✅ **Unique jobs** (deduplication)
- ✅ **Clean Architecture** (jobs access repo/cache/http)
- ✅ **Worker command** included
- ✅ **Example jobs** (email, report, sync)

**Files:**
- Interface: `pkg/queue/queue.go`
- Asynq Client: `pkg/queue/asynq.go`
- Registry: `pkg/queue/registry.go`
- Config: `pkg/config/queue.go`
- Bootstrap: `internal/bootstrap/queue.go`
- Jobs: `internal/jobs/*.go`
- Worker: `cmd/worker/worker.go`
- Examples: `internal/usecase/queue_example.go`

**Usage:**
```bash
# Start worker
./app worker

# Enqueue job (via HTTP)
curl -X POST http://localhost:9000/jobs/email \
  -d '{"to":"user@example.com","subject":"Welcome"}'
```

**Your app now has production-ready job queue!** 🚀⚙️📨

