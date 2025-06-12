package http

import (
	"github.com/hanifkf12/hanif_skeleton/pkg/app"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
	"log"
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
	application.SetupSocket()
	err = application.Run()
	if err != nil {
		log.Fatal(err)
	}
}
