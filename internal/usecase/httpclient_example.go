package usecase

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/httpclient"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Example: Call external weather API
type weatherService struct {
	httpClient httpclient.HTTPClient
}

type WeatherRequest struct {
	City string `json:"city" validate:"required"`
}

type WeatherResponse struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
	Humidity    int     `json:"humidity"`
}

func NewWeatherService(httpClient httpclient.HTTPClient) contract.UseCase {
	return &weatherService{
		httpClient: httpClient,
	}
}

func (u *weatherService) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "weatherService.Serve")
	defer span.End()

	lf := logger.NewFields("WeatherService").WithTrace(ctx)

	// Parse request
	var req WeatherRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	lf.Append(logger.Any("city", req.City))

	// Call external weather API (example)
	// Note: Replace with actual API endpoint and API key
	url := "https://api.weatherapi.com/v1/current.json?key=YOUR_API_KEY&q=" + req.City

	headers := map[string]string{
		"Accept": "application/json",
	}

	resp, err := u.httpClient.Get(ctx, url, headers)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to fetch weather data", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusServiceUnavailable).
			WithErrors("Failed to fetch weather data from external service")
	}

	if !resp.IsSuccess() {
		lf.Append(logger.Any("status_code", resp.StatusCode))
		logger.Error("Weather API returned error", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusServiceUnavailable).
			WithErrors("Weather service unavailable")
	}

	// Parse response (this is example structure)
	var weatherResp WeatherResponse
	if err := resp.JSON(&weatherResp); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse weather response", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to parse weather data")
	}

	logger.Info("Weather data fetched successfully", lf)
	return *appctx.NewResponse().WithData(weatherResp)
}

// Example: Call payment gateway API
type paymentGateway struct {
	httpClient httpclient.HTTPClient
	apiKey     string
	baseURL    string
}

type PaymentRequest struct {
	Amount      float64 `json:"amount" validate:"required"`
	Currency    string  `json:"currency" validate:"required"`
	Description string  `json:"description"`
}

type PaymentResponse struct {
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	CreatedAt     string  `json:"created_at"`
}

func NewPaymentGateway(httpClient httpclient.HTTPClient, apiKey, baseURL string) contract.UseCase {
	return &paymentGateway{
		httpClient: httpClient,
		apiKey:     apiKey,
		baseURL:    baseURL,
	}
}

func (u *paymentGateway) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "paymentGateway.Serve")
	defer span.End()

	lf := logger.NewFields("PaymentGateway").WithTrace(ctx)

	// Parse request
	var req PaymentRequest
	if err := data.FiberCtx.BodyParser(&req); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Invalid payment request", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusBadRequest).
			WithErrors("Invalid request body")
	}

	lf.Append(logger.Any("amount", req.Amount))
	lf.Append(logger.Any("currency", req.Currency))

	// Prepare request to payment gateway
	url := u.baseURL + "/transactions"

	headers := map[string]string{
		"Authorization": "Bearer " + u.apiKey,
		"Content-Type":  "application/json",
	}

	// Call payment gateway
	resp, err := u.httpClient.Post(ctx, url, req, headers)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to process payment", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusServiceUnavailable).
			WithErrors("Payment service unavailable")
	}

	if !resp.IsSuccess() {
		lf.Append(logger.Any("status_code", resp.StatusCode))
		lf.Append(logger.Any("response", resp.String()))
		logger.Error("Payment gateway returned error", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusPaymentRequired).
			WithErrors("Payment failed")
	}

	// Parse response
	var paymentResp PaymentResponse
	if err := resp.JSON(&paymentResp); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to parse payment response", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors("Failed to parse payment response")
	}

	logger.Info("Payment processed successfully", lf)
	return *appctx.NewResponse().WithData(paymentResp)
}

// Example: Generic 3rd party API call
type thirdPartyAPI struct {
	httpClient httpclient.HTTPClient
}

func NewThirdPartyAPI(httpClient httpclient.HTTPClient) contract.UseCase {
	return &thirdPartyAPI{
		httpClient: httpClient,
	}
}

func (u *thirdPartyAPI) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "thirdPartyAPI.Serve")
	defer span.End()

	lf := logger.NewFields("ThirdPartyAPI").WithTrace(ctx)

	// Example: Call multiple endpoints with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Example 1: GET request
	resp1, err := u.httpClient.Get(ctx, "https://api.example.com/users", nil)
	if err != nil {
		logger.Error("API call 1 failed", lf)
	} else {
		lf.Append(logger.Any("status", resp1.StatusCode))
		logger.Info("API call 1 success", lf)
	}

	// Example 2: POST request
	body := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	resp2, err := u.httpClient.Post(ctx, "https://api.example.com/users", body, map[string]string{
		"Authorization": "Bearer token",
	})
	if err != nil {
		logger.Error("API call 2 failed", lf)
	} else {
		lf.Append(logger.Any("status", resp2.StatusCode))
		logger.Info("API call 2 success", lf)
	}

	return *appctx.NewResponse().WithData(map[string]string{
		"message": "3rd party API calls completed",
	})
}
