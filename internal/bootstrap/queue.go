package bootstrap

import (
	"fmt"

	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/queue"
)

// RegistryQueue creates and returns a queue instance based on configuration
func RegistryQueue(cfg *config.Config) queue.Queue {
	lf := logger.NewFields("RegistryQueue")
	lf.Append(logger.Any("driver", cfg.Queue.Driver))

	switch cfg.Queue.Driver {
	case "asynq":
		return registryAsynqQueue(cfg)
	default:
		logger.Info("No queue driver specified or unsupported driver", lf)
		return nil
	}
}

// registryAsynqQueue creates Asynq queue instance
func registryAsynqQueue(cfg *config.Config) queue.Queue {
	lf := logger.NewFields("RegistryAsynqQueue")

	// Default values
	host := cfg.Queue.Host
	if host == "" {
		host = "localhost"
	}

	port := cfg.Queue.Port
	if port == 0 {
		port = 6379
	}

	db := cfg.Queue.DB
	if db < 0 {
		db = 0
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	lf.Append(logger.Any("host", host))
	lf.Append(logger.Any("port", port))
	lf.Append(logger.Any("db", db))

	queueClient := queue.NewAsynqClient(addr, cfg.Queue.Password, db)

	logger.Info("Asynq queue initialized successfully", lf)

	return queueClient
}
