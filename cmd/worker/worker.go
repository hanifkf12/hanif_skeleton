package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"

	"github.com/hanifkf12/hanif_skeleton/internal/bootstrap"
	"github.com/hanifkf12/hanif_skeleton/internal/jobs"
	userRepo "github.com/hanifkf12/hanif_skeleton/internal/repository/user"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/queue"
)

var WorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start job queue worker",
	Long:  "Start Asynq worker to process background jobs",
	Run:   runWorker,
}

func runWorker(cmd *cobra.Command, args []string) {
	// Setup logger
	logger.Setup()
	defer logger.Cleanup()

	// Load configuration
	cfg, err := config.LoadAllConfigs()
	if err != nil {
		logger.Fatal(err.Error())
	}

	lf := logger.NewFields("Worker")
	logger.Info("Starting job queue worker", lf)

	// Initialize dependencies
	db := bootstrap.RegistryDatabase(cfg, false)
	cache := bootstrap.RegistryCache(cfg)
	httpClient := bootstrap.RegistryHTTPClient(cfg)

	// Initialize repositories
	userRepository := userRepo.NewUserRepository(db)

	// Create job registry
	registry := queue.NewJobRegistry()

	// Register jobs
	lf.Append(logger.Any("registering", "jobs"))
	logger.Info("Registering job handlers", lf)

	// Register send email job
	registry.Register(
		jobs.JobTypeSendEmail,
		jobs.NewSendEmailJob(userRepository, httpClient, cache),
	)

	// Register generate report job
	registry.Register(
		jobs.JobTypeGenerateReport,
		jobs.NewGenerateReportJob(userRepository, cache),
	)

	// Register sync data job
	registry.Register(
		jobs.JobTypeSyncData,
		jobs.NewSyncDataJob(httpClient, cache),
	)

	logger.Info("Job handlers registered", lf)

	// Create Asynq server
	host := cfg.Queue.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Queue.Port
	if port == 0 {
		port = 6379
	}
	redisDB := cfg.Queue.DB
	if redisDB < 0 {
		redisDB = 0
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     addr,
			Password: cfg.Queue.Password,
			DB:       redisDB,
		},
		asynq.Config{
			Concurrency: 10, // Number of concurrent workers
			Queues: map[string]int{
				"critical": 6, // Process 60% of jobs from critical queue
				"default":  3, // Process 30% of jobs from default queue
				"low":      1, // Process 10% of jobs from low queue
			},
		},
	)

	// Create mux (job router)
	mux := asynq.NewServeMux()

	// Create wrapper for handling jobs
	wrapper := queue.NewAsynqServer(registry)

	// Register handler for all job types
	mux.HandleFunc(jobs.JobTypeSendEmail, func(ctx context.Context, task *asynq.Task) error {
		return wrapper.ProcessTask(ctx, task.Type(), task.Payload())
	})

	mux.HandleFunc(jobs.JobTypeGenerateReport, func(ctx context.Context, task *asynq.Task) error {
		return wrapper.ProcessTask(ctx, task.Type(), task.Payload())
	})

	mux.HandleFunc(jobs.JobTypeSyncData, func(ctx context.Context, task *asynq.Task) error {
		return wrapper.ProcessTask(ctx, task.Type(), task.Payload())
	})

	// Start server in goroutine
	go func() {
		lf.Append(logger.Any("redis_addr", addr))
		lf.Append(logger.Any("concurrency", 10))
		logger.Info("Asynq worker started", lf)

		if err := srv.Run(mux); err != nil {
			lf.Append(logger.Any("error", err.Error()))
			logger.Error("Worker error", lf)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...", lf)
	srv.Shutdown()

	logger.Info("Worker stopped", lf)
}
