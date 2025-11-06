package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/cache"
	"github.com/hanifkf12/hanif_skeleton/pkg/httpclient"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/queue"
)

// SendEmailJob handles sending email notifications
type SendEmailJob struct {
	userRepo   repository.UserRepository
	httpClient httpclient.HTTPClient
	cache      cache.Cache
}

// SendEmailPayload is the payload for send email job
type SendEmailPayload struct {
	UserID  int64  `json:"user_id"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// NewSendEmailJob creates a new send email job handler
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

// Handle processes the send email job
func (j *SendEmailJob) Handle(ctx context.Context, payload []byte) error {
	lf := logger.NewFields("SendEmailJob")

	// Unmarshal payload
	var data SendEmailPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to unmarshal payload", lf)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	lf.Append(logger.Any("user_id", data.UserID))
	lf.Append(logger.Any("to", data.To))
	lf.Append(logger.Any("subject", data.Subject))

	// Check cache first (prevent duplicate sends)
	cacheKey := cache.NewCacheKey("email").Build(fmt.Sprintf("%d", data.UserID), data.Subject)
	exists, _ := j.cache.Exists(ctx, cacheKey)
	if exists {
		logger.Info("Email already sent (cached), skipping", lf)
		return nil
	}

	// Get users from repository (example: verify user exists)
	// Note: In production, you'd have a GetUserByID method
	users, err := j.userRepo.GetUsers(ctx)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get users from database", lf)
		// Don't fail the job, just log the error
		logger.Info("Continuing with email send despite user lookup failure", lf)
	} else if len(users) > 0 {
		lf.Append(logger.Any("user_count", len(users)))
	}

	// Call email service API via HTTP client
	emailPayload := map[string]interface{}{
		"to":      data.To,
		"subject": data.Subject,
		"body":    data.Body,
	}

	headers := map[string]string{
		"X-API-Key": "your-email-service-api-key",
	}

	resp, err := j.httpClient.Post(ctx, "https://api.emailservice.com/send", emailPayload, headers)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to send email via API", lf)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if !resp.IsSuccess() {
		lf.Append(logger.Any("status_code", resp.StatusCode))
		logger.Error("Email service returned error", lf)
		return fmt.Errorf("email service error: status %d", resp.StatusCode)
	}

	// Cache the result to prevent duplicate sends (1 hour)
	j.cache.Set(ctx, cacheKey, "sent", 1*60*60) // 1 hour

	logger.Info("Email sent successfully", lf)
	return nil
}
