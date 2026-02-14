package main

import (
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/cli"
)

func main() {
	err := cli.Execute()
	if err != nil {
		panic(err)
	}
}
