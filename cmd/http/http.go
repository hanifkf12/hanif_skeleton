package http

import (
	"github.com/hanifkf12/hanif_skeleton/pkg/app"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/logger"
	"log"
)

func Start() {
	logger.Setup()
	cfg, err := config.LoadAllConfigs()
	if err != nil {
		logger.Fatal(err)
	}
	application := app.InitializeApp(cfg)
	err = application.Run()
	if err != nil {
		log.Fatal(err)
	}
}
