package main

import (
	"os"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/internal/cli/client"
	clipmigrate "github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/internal/cli/migrate"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/internal/cli/server"
	"github.com/spf13/cobra"
)

func main() {
	//nolint:exhaustruct
	rootCmd := &cobra.Command{
		Use:   "jira-analyzer",
		Short: "Jira Analyzer",
		Long:  "Jira Analyzer is a tool for analyzing Jira issues.",
	}

	rootCmd.AddCommand(server.NewCommand())
	rootCmd.AddCommand(client.NewCommand())
	rootCmd.AddCommand(clipmigrate.NewCommand())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
