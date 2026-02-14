// Package fetcher provides functionality for fetching Jira data.
package fetcher

type Fetcher struct{}

func New() *Fetcher {
	return &Fetcher{}
}
