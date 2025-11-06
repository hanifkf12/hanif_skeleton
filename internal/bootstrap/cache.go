package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/hanifkf12/hanif_skeleton/pkg/cache"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RegistryCache creates and returns a cache instance based on configuration
func RegistryCache(cfg *config.Config) cache.Cache {
	lf := logger.NewFields("RegistryCache")
	lf.Append(logger.Any("driver", cfg.Cache.Driver))

	switch cfg.Cache.Driver {
	case "redis":
		return registryRedisCache(cfg)
	case "memory":
		return registryMemoryCache(cfg)
	default:
		logger.Info("No cache driver specified, using memory cache", lf)
		return registryMemoryCache(cfg)
	}
}

// registryRedisCache creates Redis cache instance
func registryRedisCache(cfg *config.Config) cache.Cache {
	lf := logger.NewFields("RegistryRedisCache")

	// Default values
	host := cfg.Cache.Host
	if host == "" {
		host = "localhost"
	}

	port := cfg.Cache.Port
	if port == 0 {
		port = 6379
	}

	db := cfg.Cache.DB
	if db < 0 {
		db = 0
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	lf.Append(logger.Any("host", host))
	lf.Append(logger.Any("port", port))
	lf.Append(logger.Any("db", db))

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Cache.Password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	logger.Info("Redis cache initialized successfully", lf)

	return cache.NewRedisCache(client)
}

// registryMemoryCache creates in-memory cache instance
func registryMemoryCache(cfg *config.Config) cache.Cache {
	lf := logger.NewFields("RegistryMemoryCache")

	logger.Info("Memory cache initialized successfully", lf)
	logger.Info("⚠️  Memory cache is for development only, use Redis in production", lf)

	return cache.NewMemoryCache()
}
