package usecase

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/crypto"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Example usecase for encrypting data
type encryptData struct {
	crypto crypto.Crypto
}

// Request/Response structures
type EncryptRequest struct {
	Data string `json:"data" validate:"required"`
}

type EncryptResponse struct {
	EncryptedData string `json:"encrypted_data"`
	Hash          string `json:"hash"`
}

type DecryptRequest struct {
	EncryptedData string `json:"encrypted_data" validate:"required"`
}

type DecryptResponse struct {
	Data string `json:"data"`
}

func NewEncryptData(crypto crypto.Crypto) contract.UseCase {
	return &encryptData{crypto: crypto}
}

func (u *encryptData) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "encryptData.Serve")
	defer span.End()

	lf := logger.NewFields("EncryptData").WithTrace(ctx)

	// Parse request
	var req EncryptRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	// Encrypt data
	encrypted, err := u.crypto.Encrypt(req.Data)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to encrypt data", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to encrypt data")
	}

	// Create hash
	hash := u.crypto.Hash(req.Data)

	response := EncryptResponse{
		EncryptedData: encrypted,
		Hash:          hash,
	}

	logger.Info("Data encrypted successfully", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(response)
}

// Decrypt usecase
type decryptData struct {
	crypto crypto.Crypto
}

func NewDecryptData(crypto crypto.Crypto) contract.UseCase {
	return &decryptData{crypto: crypto}
}

func (u *decryptData) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "decryptData.Serve")
	defer span.End()

	lf := logger.NewFields("DecryptData").WithTrace(ctx)

	// Parse request
	var req DecryptRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	// Decrypt data
	decrypted, err := u.crypto.Decrypt(req.EncryptedData)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to decrypt data", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Failed to decrypt data")
	}

	response := DecryptResponse{
		Data: decrypted,
	}

	logger.Info("Data decrypted successfully", lf)
	return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(response)
}
