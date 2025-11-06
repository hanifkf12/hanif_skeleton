package usecase

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/repository"
	"github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
	"github.com/hanifkf12/hanif_skeleton/pkg/cache"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

// Example usecase with cache: Get users with caching
type userWithCache struct {
	userRepo repository.UserRepository
	cache    cache.Cache
}

func NewUserWithCache(userRepo repository.UserRepository, cache cache.Cache) contract.UseCase {
	return &userWithCache{
		userRepo: userRepo,
		cache:    cache,
	}
}

func (u *userWithCache) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "userWithCache.Serve")
	defer span.End()

	lf := logger.NewFields("UserWithCache").WithTrace(ctx)

	// Build cache key
	cacheKey := cache.NewCacheKey("users").Build("list")
	lf.Append(logger.Any("cache_key", cacheKey))

	// Try to get from cache first
	cachedData, err := u.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		lf.Append(logger.Any("cache", "hit"))
		logger.Info("Data retrieved from cache", lf)

		// Note: In production, you'd unmarshal JSON here
		// For now, return cached data as string for demo
		_ = cachedData // Use the cached data here in production
	} else {
		// Cache miss
		lf.Append(logger.Any("cache", "miss"))
		logger.Info("Cache miss, querying database", lf)
	}

	// Get from database
	users, err := u.userRepo.GetUsers(ctx)
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get users", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors(err.Error())
	}

	// Store in cache (5 minutes expiry)
	// Note: In production, marshal to JSON first
	// jsonData, _ := json.Marshal(users)
	// u.cache.Set(ctx, cacheKey, jsonData, 5*time.Minute)
	u.cache.Set(ctx, cacheKey, fmt.Sprintf("%v", users), 5*time.Minute)

	logger.Info("Users retrieved successfully", lf)
	return *appctx.NewResponse().WithData(users)
}

// CacheStats shows cache statistics
type cacheStats struct {
	cache cache.Cache
}

type CacheStatsResponse struct {
	Status string   `json:"status"`
	Keys   []string `json:"keys,omitempty"`
	Count  int      `json:"count"`
}

func NewCacheStats(cache cache.Cache) contract.UseCase {
	return &cacheStats{cache: cache}
}

func (u *cacheStats) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "cacheStats.Serve")
	defer span.End()

	lf := logger.NewFields("CacheStats").WithTrace(ctx)

	// Ping cache
	if err := u.cache.Ping(ctx); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Cache ping failed", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusServiceUnavailable).
			WithErrors("Cache unavailable")
	}

	// Get all keys
	keys, err := u.cache.Keys(ctx, "*")
	if err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to get cache keys", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors(err.Error())
	}

	response := CacheStatsResponse{
		Status: "ok",
		Keys:   keys,
		Count:  len(keys),
	}

	logger.Info("Cache stats retrieved", lf)
	return *appctx.NewResponse().WithData(response)
}

// ClearCache clears all cache
type clearCache struct {
	cache cache.Cache
}

func NewClearCache(cache cache.Cache) contract.UseCase {
	return &clearCache{cache: cache}
}

func (u *clearCache) Serve(data appctx.Data) appctx.Response {
	ctx := data.FiberCtx.UserContext()
	ctx, span := telemetry.StartSpan(ctx, "clearCache.Serve")
	defer span.End()

	lf := logger.NewFields("ClearCache").WithTrace(ctx)

	// Get specific key from query param (optional)
	key := data.FiberCtx.Query("key")

	if key != "" {
		// Delete specific key
		if err := u.cache.Delete(ctx, key); err != nil {
			telemetry.SpanError(ctx, err)
			lf.Append(logger.Any("error", err.Error()))
			logger.Error("Failed to delete cache key", lf)
			return *appctx.NewResponse().
				WithCode(fiber.StatusInternalServerError).
				WithErrors(err.Error())
		}

		lf.Append(logger.Any("key", key))
		logger.Info("Cache key deleted", lf)

		return *appctx.NewResponse().
			WithData(map[string]string{"message": fmt.Sprintf("Key '%s' deleted", key)})
	}

	// Flush all cache
	if err := u.cache.FlushAll(ctx); err != nil {
		telemetry.SpanError(ctx, err)
		lf.Append(logger.Any("error", err.Error()))
		logger.Error("Failed to flush cache", lf)
		return *appctx.NewResponse().
			WithCode(fiber.StatusInternalServerError).
			WithErrors(err.Error())
	}

	logger.Info("Cache flushed successfully", lf)
	return *appctx.NewResponse().
		WithData(map[string]string{"message": "All cache cleared"})
}
