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
		Use:                "db:migrate [command]",
		Short:              "database migration",
		DisableFlagParsing: true,
		Run: func(c *cobra.Command, args []string) {
			migration.MigrateDatabase(args)
		},
	}
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
