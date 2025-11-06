package queue

import (
	"context"
	"encoding/json"
	"time"
)

// Queue is the interface for job queue operations
type Queue interface {
	// Enqueue enqueues a job to be processed immediately
	Enqueue(ctx context.Context, jobType string, payload interface{}) error

	// EnqueueWithDelay enqueues a job to be processed after a delay
	EnqueueWithDelay(ctx context.Context, jobType string, payload interface{}, delay time.Duration) error

	// EnqueueAt enqueues a job to be processed at a specific time
	EnqueueAt(ctx context.Context, jobType string, payload interface{}, processAt time.Time) error

	// EnqueueWithOptions enqueues a job with custom options
	EnqueueWithOptions(ctx context.Context, jobType string, payload interface{}, opts *EnqueueOptions) error

	// Close closes the queue client
	Close() error
}

// EnqueueOptions holds options for enqueuing a job
type EnqueueOptions struct {
	Queue     string        // Queue name (default: "default")
	MaxRetry  int           // Max retry attempts
	Timeout   time.Duration // Job timeout
	Delay     time.Duration // Delay before processing
	ProcessAt time.Time     // Process at specific time
	Unique    bool          // Unique job (prevent duplicates)
	UniqueTTL time.Duration // TTL for unique constraint
}

// JobHandler is the function signature for job handlers
type JobHandler func(ctx context.Context, payload []byte) error

// JobRegistry manages job handlers
type JobRegistry interface {
	// Register registers a job handler
	Register(jobType string, handler JobHandler)

	// Get gets a job handler by type
	Get(jobType string) (JobHandler, bool)
}

// MarshalPayload marshals job payload to JSON
func MarshalPayload(payload interface{}) ([]byte, error) {
	return json.Marshal(payload)
}

// UnmarshalPayload unmarshals job payload from JSON
func UnmarshalPayload(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
