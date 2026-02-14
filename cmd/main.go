package main

import (
	"os"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/cli"
)

func main() {
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
