package usecase

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/jobs"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/queue"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Example: Enqueue send email job
type enqueueSendEmail struct {
	queue queue.Queue
}

type EnqueueEmailRequest struct {
	UserID  int64  `json:"user_id" validate:"required"`
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

func NewEnqueueSendEmail(queue queue.Queue) contract.UseCase {
	return &enqueueSendEmail{queue: queue}
}

func (u *enqueueSendEmail) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "enqueueSendEmail.Serve")
	defer span.End()

	lf := logger.NewFields("EnqueueSendEmail").WithTrace(ctx)

	// Parse request
	var req EnqueueEmailRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	lf.Append(logger.Any("user_id", req.UserID))
	lf.Append(logger.Any("to", req.To))

	// Prepare job payload
	payload := jobs.SendEmailPayload{
		UserID:  req.UserID,
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
	}

	// Enqueue job
	err := u.queue.Enqueue(ctx, jobs.JobTypeSendEmail, payload)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to enqueue job", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to enqueue email job")
	}

	logger.Info("Email job enqueued successfully", lf)
	return *appctx.NewResponse().
		WithData(map[string]string{
			"message": "Email job enqueued successfully",
			"status":  "queued",
		})
}

// Example: Enqueue generate report job with delay
type enqueueGenerateReport struct {
	queue queue.Queue
}

type EnqueueReportRequest struct {
	ReportType string `json:"report_type" validate:"required"`
	UserID     int64  `json:"user_id" validate:"required"`
	StartDate  string `json:"start_date" validate:"required"`
	EndDate    string `json:"end_date" validate:"required"`
	DelayMin   int    `json:"delay_minutes"` // Optional: delay in minutes
}

func NewEnqueueGenerateReport(queue queue.Queue) contract.UseCase {
	return &enqueueGenerateReport{queue: queue}
}

func (u *enqueueGenerateReport) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "enqueueGenerateReport.Serve")
	defer span.End()

	lf := logger.NewFields("EnqueueGenerateReport").WithTrace(ctx)

	// Parse request
	var req EnqueueReportRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	lf.Append(logger.Any("report_type", req.ReportType))
	lf.Append(logger.Any("user_id", req.UserID))

	// Parse dates
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	// Prepare job payload
	payload := jobs.GenerateReportPayload{
		ReportType: req.ReportType,
		UserID:     req.UserID,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	// Enqueue job with delay if specified
	var err error
	if req.DelayMin > 0 {
		delay := time.Duration(req.DelayMin) * time.Minute
		lf.Append(logger.Any("delay", delay.String()))
		err = u.queue.EnqueueWithDelay(ctx, jobs.JobTypeGenerateReport, payload, delay)
	} else {
		err = u.queue.Enqueue(ctx, jobs.JobTypeGenerateReport, payload)
	}

	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to enqueue job", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to enqueue report job")
	}

	logger.Info("Report job enqueued successfully", lf)
	return *appctx.NewResponse().
		WithData(map[string]string{
			"message": "Report generation job enqueued",
			"status":  "queued",
		})
}

// Example: Enqueue sync data job
type enqueueSyncData struct {
	queue queue.Queue
}

type EnqueueSyncRequest struct {
	EntityType string `json:"entity_type" validate:"required"`
	EntityID   string `json:"entity_id" validate:"required"`
	Action     string `json:"action" validate:"required,oneof=create update delete"`
}

func NewEnqueueSyncData(queue queue.Queue) contract.UseCase {
	return &enqueueSyncData{queue: queue}
}

func (u *enqueueSyncData) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "enqueueSyncData.Serve")
	defer span.End()

	lf := logger.NewFields("EnqueueSyncData").WithTrace(ctx)

	// Parse request
	var req EnqueueSyncRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	lf.Append(logger.Any("entity_type", req.EntityType))
	lf.Append(logger.Any("entity_id", req.EntityID))
	lf.Append(logger.Any("action", req.Action))

	// Prepare job payload
	payload := jobs.SyncDataPayload{
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Action:     req.Action,
	}

	// Enqueue job with options
	err := u.queue.EnqueueWithOptions(ctx, jobs.JobTypeSyncData, payload, &queue.EnqueueOptions{
		Queue:     "critical", // Use critical queue for sync jobs
		MaxRetry:  5,          // Retry up to 5 times
		Timeout:   30 * time.Second,
		Unique:    true, // Prevent duplicate sync jobs
		UniqueTTL: 5 * time.Minute,
	})

	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to enqueue job", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to enqueue sync job")
	}

	logger.Info("Sync job enqueued successfully", lf)
	return *appctx.NewResponse().
		WithData(map[string]string{
			"message": "Sync job enqueued successfully",
			"status":  "queued",
		})
}
