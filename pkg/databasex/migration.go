package databasex

import (
	"flag"
	"fmt"
	"log"

	"github.com/hanifkf12/hanif_skeleton/pkg/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func DatabaseMigration(cfg *config.Config, rawArgs []string) {
	fs := flag.NewFlagSet("db:migrate", flag.ContinueOnError)
	dir := fs.String("dir", "database/migration", "directory with migration files")
	table := fs.String("table", "db_migration", "migrations table name")
	verbose := fs.Bool("verbose", false, "enable verbose mode")
	help := fs.Bool("guide", false, "print help")
	version := fs.Bool("version", false, "print version")

	fs.Usage = usage
	if err := fs.Parse(rawArgs); err != nil {
		return
	}

	if *version {
		fmt.Println(goose.VERSION)
		return
	}
	if *help {
		usage()
		return
	}
	if *verbose {
		goose.SetVerbose(true)
	}

	goose.SetTableName(*table)

	args := fs.Args()
	command := "up"
	if len(args) > 0 {
		command = args[0]
	}

	switch command {
	case "help":
		usage()
		return
	case "create":
		var createArgs []string
		if len(args) > 1 {
			createArgs = args[1:]
		}
		if err := goose.Run("create", nil, *dir, createArgs...); err != nil {
			log.Fatalf("goose run: %v", err)
		}
		return
	case "fix":
		if err := goose.Run("fix", nil, *dir); err != nil {
			log.Fatalf("goose run: %v", err)
		}
		return
	}

	var dbDriver string
	var dbConnStr string

	switch cfg.Database.Driver {
	case "postgres", "pgx":
		dbDriver = cfg.Database.Driver
		dbConnStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.Username, cfg.Database.Password, cfg.Database.Name)
	default:
		log.Fatalf("Unsupported database driver: %s", cfg.Database.Driver)
	}

	db, err := goose.OpenDBWithDriver(dbDriver, dbConnStr)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("db migrate: failed to close DB: %v\n", err)
		}
	}()

	arguments := []string{}
	if len(args) > 1 {
		arguments = append(arguments, args[1:]...)
	}

	if err := goose.Run(command, db, *dir, arguments...); err != nil {
		log.Fatalf("db migrate run: %v", err)
	}
}

func usage() {
	fmt.Println(usageCommands)
}

var (
	usageCommands = `
  --dir string     directory with migration files (default "database/migration")
  --guide          print help
  --table string   migrations table name (default "db_migration")
  --verbose        enable verbose mode
  --version        print version

Commands:
    up                   Migrate the DB to the most recent version available
    up-by-one            Migrate the DB up by 1
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations
`
)
