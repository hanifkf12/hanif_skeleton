package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/hanifkf12/hanif_skeleton/internal/projectgen"
	"github.com/spf13/cobra"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := newCommand().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	var opts projectgen.Options
	cmd := &cobra.Command{
		Use:   "skeleton nama_project git_url",
		Short: "Generate a Go project from hanif_skeleton",
		Long: "Generate a new Go project from the remote skeleton template.\n" +
			"git_url is the new project's repository URL; it determines the Go module and Git origin.\n" +
			"Requires Git and access to the template repository. Existing destinations are never overwritten.",
		Example: "  skeleton inventory-api https://github.com/hanifkf12/inventory-api.git\n" +
			"  skeleton billing-api git@github.com:hanifkf12/billing-api.git --output ~/projects --template-ref main\n" +
			"  skeleton demo-api https://github.com/hanifkf12/demo-api.git --no-git",
		Args:          cobra.ExactArgs(2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name, opts.GitURL = args[0], args[1]
			result, err := projectgen.Generate(cmd.Context(), opts)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Project generated: %s\nGo module: %s\nTemplate commit: %s\n", result.Directory, result.ModulePath, result.TemplateCommit)
			if opts.NoGit {
				fmt.Fprintln(out, "Git initialization skipped.")
			} else {
				fmt.Fprintf(out, "Fresh Git repository on main; origin: %s (no commit or push).\n", opts.GitURL)
			}
			fmt.Fprintf(out, "\nNext steps:\n  cd %s\n  cp .env.example .env\n", shellQuote(result.Directory))
			fmt.Fprintln(out, "  # Configure .env and start PostgreSQL and the OpenTelemetry collector (localhost:4317).")
			fmt.Fprintln(out, "  go run main.go db:migrate\n  make run-http")
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.OutputDir, "output", "o", ".", "Parent directory for the new project")
	cmd.Flags().StringVar(&opts.TemplateRef, "template-ref", projectgen.DefaultTemplateRef, "Template branch, tag, or commit to fetch (pin a tag/commit for reproducibility)")
	cmd.Flags().BoolVar(&opts.NoGit, "no-git", false, "Do not initialize Git or set origin in the generated project")
	return cmd
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
