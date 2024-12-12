package migration

import (
	"fmt"
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
	"log"
)

func MigrateDatabase() {
	cfg, err := config.LoadAllConfigs()

	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to load configuration file: %v", err))
	}

	databasex.DatabaseMigration(cfg)
}
