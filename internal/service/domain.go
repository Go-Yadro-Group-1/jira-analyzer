package service

type IssueDurationHistogram struct {
	ProjectID int
	Buckets   map[string]int
}
