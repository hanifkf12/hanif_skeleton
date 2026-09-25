package middleware

import (
	"mime"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
)

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	MaxRequests int
	WindowSize  int // in seconds
}

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

// RateLimit limits requests per IP
// Returns 200 if within limit, 429 if exceeded
func RateLimit(cfg RateLimitConfig) Middleware {
	if cfg.MaxRequests <= 0 {
		cfg.MaxRequests = 1
	}
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = 60
	}

	var mu sync.Mutex
	entries := make(map[string]rateLimitEntry)
	window := time.Duration(cfg.WindowSize) * time.Second
	lastCleanup := time.Now()

	return func(ctx *fiber.Ctx, config *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.RateLimit")

		// Get client IP
		clientIP := ctx.IP()
		lf.Append(logger.Any("client_ip", clientIP))

		now := time.Now()
		mu.Lock()
		if now.Sub(lastCleanup) >= window {
			for ip, candidate := range entries {
				if now.Sub(candidate.windowStart) >= window {
					delete(entries, ip)
				}
			}
			lastCleanup = now
		}
		entry := entries[clientIP]
		if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= window {
			entry = rateLimitEntry{windowStart: now}
		}
		entry.count++
		entries[clientIP] = entry
		mu.Unlock()

		if entry.count > cfg.MaxRequests {
			lf.Append(logger.Any("count", entry.count))
			lf.Append(logger.Any("max", cfg.MaxRequests))
			logger.Error("Rate limit exceeded", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusTooManyRequests).
				WithErrors("Rate limit exceeded")
		}

		logger.Info("Rate limit check passed", lf)
		return *appctx.NewResponse().WithCode(appctx.StatusOK)
	}
}

// ContentTypeValidator validates Content-Type header
// Returns 200 if valid, 415 if invalid
func ContentTypeValidator(allowedTypes []string) Middleware {
	return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.ContentTypeValidator")

		contentType, _, err := mime.ParseMediaType(ctx.Get("Content-Type"))
		lf.Append(logger.Any("content_type", contentType))

		// Check if content type is allowed
		valid := false
		for _, allowedType := range allowedTypes {
			if contentType == allowedType {
				valid = true
				break
			}
		}

		if err != nil || !valid {
			lf.Append(logger.Any("error", "unsupported content type"))
			logger.Error("Content type validation failed", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusUnsupportedMediaType).
				WithErrors("Unsupported content type")
		}

		logger.Info("Content type validation successful", lf)
		return *appctx.NewResponse().WithCode(appctx.StatusOK)
	}
}

// IPWhitelist validates client IP against whitelist
// Returns 200 if in whitelist, 403 if not
func IPWhitelist(allowedIPs []string) Middleware {
	return func(ctx *fiber.Ctx, cfg *config.Config) appctx.Response {
		lf := logger.NewFields("Middleware.IPWhitelist")

		clientIP := ctx.IP()
		lf.Append(logger.Any("client_ip", clientIP))

		// Check if IP is in whitelist
		valid := false
		for _, allowedIP := range allowedIPs {
			if clientIP == allowedIP {
				valid = true
				break
			}
		}

		if !valid {
			lf.Append(logger.Any("error", "IP not in whitelist"))
			logger.Error("IP whitelist check failed", lf)
			return *appctx.NewResponse().
				WithCode(appctx.StatusForbidden).
				WithErrors("Access denied")
		}

		logger.Info("IP whitelist check passed", lf)
		return *appctx.NewResponse().WithCode(appctx.StatusOK)
	}
}
