package version

import (
	"github.com/spf13/cobra"
)

// Version use ldflags to set version.
//
//nolint:gochecknoglobals
var Version = "dev"

//nolint:exhaustruct
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of Jira Analyzer",
		Long:  "Print the version of Jira Analyzer.",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println("version of jira-analyzer:", Version)
		},
	}
}
