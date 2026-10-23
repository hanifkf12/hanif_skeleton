package bootstrap

import (
	"log"

	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
	"github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"
)

func RegistryDatabase(cfg *config.Config) databasex.Database {
	sqlbuilder.SetDefaultDriver(sqlbuilder.DriverPostgreSQL)

	database, err := databasex.NewPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return database
}
