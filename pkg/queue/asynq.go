package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hibiken/asynq"
)

// asynqClient implements Queue interface using Asynq
type asynqClient struct {
	client *asynq.Client
}

// NewAsynqClient creates a new Asynq queue client
func NewAsynqClient(redisAddr string, redisPassword string, redisDB int) Queue {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	return &asynqClient{
		client: client,
	}
}

// Enqueue enqueues a job to be processed immediately
func (q *asynqClient) Enqueue(ctx context.Context, jobType string, payload interface{}) error {
	return q.EnqueueWithOptions(ctx, jobType, payload, nil)
}

// EnqueueWithDelay enqueues a job with delay
func (q *asynqClient) EnqueueWithDelay(ctx context.Context, jobType string, payload interface{}, delay time.Duration) error {
	return q.EnqueueWithOptions(ctx, jobType, payload, &EnqueueOptions{
		Delay: delay,
	})
}

// EnqueueAt enqueues a job to be processed at specific time
func (q *asynqClient) EnqueueAt(ctx context.Context, jobType string, payload interface{}, processAt time.Time) error {
	return q.EnqueueWithOptions(ctx, jobType, payload, &EnqueueOptions{
		ProcessAt: processAt,
	})
}

// EnqueueWithOptions enqueues a job with custom options
func (q *asynqClient) EnqueueWithOptions(ctx context.Context, jobType string, payload interface{}, opts *EnqueueOptions) error {
	lf := logger.NewFields("AsynqClient.Enqueue")
	lf.Append(logger.Any("job_type", jobType))

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to marshal job payload", lf)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create task
	task := asynq.NewTask(jobType, payloadBytes)

	// Prepare options
	var taskOpts []asynq.Option

	if opts != nil {
		// Queue name
		if opts.Queue != "" {
			taskOpts = append(taskOpts, asynq.Queue(opts.Queue))
		}

		// Max retry
		if opts.MaxRetry > 0 {
			taskOpts = append(taskOpts, asynq.MaxRetry(opts.MaxRetry))
		}

		// Timeout
		if opts.Timeout > 0 {
			taskOpts = append(taskOpts, asynq.Timeout(opts.Timeout))
		}

		// Process at specific time
		if !opts.ProcessAt.IsZero() {
			taskOpts = append(taskOpts, asynq.ProcessAt(opts.ProcessAt))
			lf.Append(logger.Any("process_at", opts.ProcessAt))
		} else if opts.Delay > 0 {
			// Delay
			taskOpts = append(taskOpts, asynq.ProcessIn(opts.Delay))
			lf.Append(logger.Any("delay", opts.Delay.String()))
		}

		// Unique job
		if opts.Unique {
			ttl := opts.UniqueTTL
			if ttl == 0 {
				ttl = 24 * time.Hour // Default 24 hours
			}
			taskOpts = append(taskOpts, asynq.Unique(ttl))
		}
	}

	// Enqueue task
	info, err := q.client.EnqueueContext(ctx, task, taskOpts...)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to enqueue job", lf)
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	lf.Append(logger.Any("task_id", info.ID))
	lf.Append(logger.Any("queue", info.Queue))
	logger.Info("Job enqueued successfully", lf)

	return nil
}

// Close closes the Asynq client
func (q *asynqClient) Close() error {
	return q.client.Close()
}
