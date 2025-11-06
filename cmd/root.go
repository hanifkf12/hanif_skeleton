package cmd

import (
	"log"

	"github.com/hanifkf12/hanif_skeleton/cmd/http"
	"github.com/hanifkf12/hanif_skeleton/cmd/migration"
	"github.com/hanifkf12/hanif_skeleton/cmd/pubsub"
	"github.com/hanifkf12/hanif_skeleton/cmd/worker"
	"github.com/spf13/cobra"
)

func Start() {
	rootCmd := &cobra.Command{}

	migrateCmd := &cobra.Command{
		Use:   "db:migrate",
		Short: "database migration",
		Run: func(c *cobra.Command, args []string) {
			migration.MigrateDatabase()
		},
	}

	migrateCmd.Flags().BoolP("version", "", false, "print version")
	migrateCmd.Flags().StringP("dir", "", "database/migration/", "directory with migration files")
	migrateCmd.Flags().StringP("table", "", "db", "migrations table name")
	migrateCmd.Flags().BoolP("verbose", "", false, "enable verbose mode")
	migrateCmd.Flags().BoolP("guide", "", false, "print help")

	cmd := []*cobra.Command{
		{
			Use:   "http",
			Short: "http server",
			Run: func(cmd *cobra.Command, args []string) {
				http.Start()
			},
		},
		{
			Use:   "pubsub",
			Short: "pubsub worker",
			Run: func(cmd *cobra.Command, args []string) {
				pubsub.Start()
			},
		},
		worker.WorkerCmd,
		migrateCmd,
	}

	rootCmd.AddCommand(cmd...)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
