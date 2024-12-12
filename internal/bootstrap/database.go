package bootstrap

import (
	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
	"log"
)

func RegistryDatabase(cfg *config.Config) databasex.Database {
	database, err := databasex.NewMySql(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return database
}
