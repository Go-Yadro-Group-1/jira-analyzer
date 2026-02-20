package cli

import (
	"log"

	"github.com/spf13/cobra"
)

// nolint: gochecknoglobals
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Jira Analyzer service",
	//nolint: revive
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Starting Jira Analyzer service...")

		return nil
	},
}

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(runCmd)
}
