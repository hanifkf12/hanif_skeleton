package usecase

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/storage"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Example usecase for file upload
type uploadFile struct {
	storage storage.Storage
}

func NewUploadFile(storage storage.Storage) contract.UseCase {
	return &uploadFile{storage: storage}
}

func (u *uploadFile) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "uploadFile.Serve")
	defer span.End()

	lf := logger.NewFields("UploadFile").WithTrace(ctx)

	// Get file from multipart form
	file, err := data.FiberCtx.FormFile("file")
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get file from form", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("File is required")
	}

	lf.Append(logger.Any("filename", file.Filename))
	lf.Append(logger.Any("size", file.Size))

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to open uploaded file", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to process file")
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	storagePath := fmt.Sprintf("uploads/%s", filename)

	// Read file content
	content, err := io.ReadAll(src)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to read file", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to read file")
	}

	// Upload to storage
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	err = u.storage.Upload(ctx, storagePath, bytes.NewReader(content), contentType)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to upload file to storage", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to upload file")
	}

	// Generate URL (valid for 1 hour)
	url, err := u.storage.GetURL(ctx, storagePath, 1*time.Hour)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to generate file URL", lf)
		// Continue even if URL generation fails
		url = ""
	}

	response := map[string]interface{}{
		"path":          storagePath,
		"filename":      filename,
		"original_name": file.Filename,
		"size":          file.Size,
		"content_type":  contentType,
		"url":           url,
	}

	logger.Info("File uploaded successfully", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(response)
}
