package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/pkg/cache"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/queue"
)

// GenerateReportJob handles report generation
type GenerateReportJob struct {
	userRepo repository.UserRepository
	cache    cache.Cache
}

// GenerateReportPayload is the payload for generate report job
type GenerateReportPayload struct {
	ReportType string    `json:"report_type"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	UserID     int64     `json:"user_id"`
}

// NewGenerateReportJob creates a new generate report job handler
func NewGenerateReportJob(
	userRepo repository.UserRepository,
	cache cache.Cache,
) queue.JobHandler {
	job := &GenerateReportJob{
		userRepo: userRepo,
		cache:    cache,
	}
	return job.Handle
}

// Handle processes the generate report job
func (j *GenerateReportJob) Handle(ctx context.Context, payload []byte) error {
	lf := logger.NewFields("GenerateReportJob")

	// Unmarshal payload
	var data GenerateReportPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to unmarshal payload", lf)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	lf.Append(logger.Any("report_type", data.ReportType))
	lf.Append(logger.Any("user_id", data.UserID))
	lf.Append(logger.Any("start_date", data.StartDate))
	lf.Append(logger.Any("end_date", data.EndDate))

	logger.Info("Starting report generation", lf)

	// Get users from repository
	users, err := j.userRepo.GetUsers(ctx)
	if err != nil {
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get users", lf)
		return fmt.Errorf("failed to get users: %w", err)
	}

	lf.Append(logger.Any("user_count", len(users)))

	// Simulate report generation (in real app, generate PDF/CSV/etc)
	time.Sleep(2 * time.Second) // Simulate processing

	// Cache the report (in real app, store file path or URL)
	cacheKey := cache.NewCacheKey("report").Build(
		data.ReportType,
		fmt.Sprintf("%d", data.UserID),
	)

	reportData := map[string]interface{}{
		"report_type":  data.ReportType,
		"generated_at": time.Now(),
		"total_users":  len(users),
		"status":       "completed",
	}

	reportJSON, _ := json.Marshal(reportData)
	j.cache.Set(ctx, cacheKey, reportJSON, 24*time.Hour) // Cache for 24 hours

	logger.Info("Report generated successfully", lf)
	return nil
}
