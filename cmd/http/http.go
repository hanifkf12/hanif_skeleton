package http

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hanifkf12/hanif_skeleton/pkg/app"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

func Start() {
	logger.Setup()
	// Ensure logger is properly cleaned up on shutdown
	defer logger.Cleanup()

	cfg, err := config.LoadAllConfigs()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Initialize tracer
	cleanup, err := telemetry.InitTracer("hanif-skeleton")
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer cleanup()

	application := app.InitializeApp(cfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down HTTP server...")
		_ = application.Shutdown()
	}()

	err = application.Run()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
