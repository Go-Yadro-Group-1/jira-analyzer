package service

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
)

var ErrNoClosedIssues = errors.New("no closed issues to build histogram")

const (
	secondsInHour  = 3600
	secondsInDay   = 86400
	secondsInMonth = 30 * secondsInDay
	secondsInYear  = 365 * secondsInDay
)

func durationToBarLabel(seconds int64) string {
	switch {
	case seconds < secondsInDay:
		return fmt.Sprintf("%dh", seconds/secondsInHour)
	case seconds < secondsInMonth:
		return fmt.Sprintf("%dday", seconds/secondsInDay)
	case seconds < secondsInYear:
		return fmt.Sprintf("%dmonth", seconds/secondsInMonth)
	default:
		if y := seconds / secondsInYear; y >= 8 {
			return "8+year"
		} else {
			return fmt.Sprintf("%dyear", y)
		}
	}
}

func durationToBarIndex(seconds int64) int {
	switch {
	case seconds < secondsInDay:
		return int(seconds / secondsInHour)
	case seconds < secondsInMonth:
		return 23 + int(seconds/secondsInDay)
	case seconds < secondsInYear:
		return 53 + int(seconds/secondsInMonth)
	default:
		y := int(seconds / secondsInYear)
		if y >= 8 {
			return 72
		}
		return 64 + y
	}
}

func buildMultiScaleHistogram(durations []int64) ([]HistogramBar, error) {
	if len(durations) == 0 {
		return nil, ErrNoClosedIssues
	}

	const totalBars = 73

	var counts [totalBars]int
	maxIdx := 0

	for _, d := range durations {
		idx := durationToBarIndex(d)
		counts[idx]++
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	bars := make([]HistogramBar, maxIdx+1)
	for i := 0; i <= maxIdx; i++ {
		bars[i] = HistogramBar{Label: durationToBarLabel(barIndexToSeconds(i)), Count: counts[i]}
	}

	return bars, nil
}

func barIndexToSeconds(i int) int64 {
	switch {
	case i < 24:
		return int64(i) * secondsInHour
	case i < 54:
		return int64(i-23) * secondsInDay
	case i < 65:
		return int64(i-53) * secondsInMonth
	case i < 72:
		return int64(i-64) * secondsInYear
	default:
		return 8 * secondsInYear
	}
}

func buildIssuesDurationHistogram(rows []repository.IssueDuration) (IssuesDurationHistogram, error) {
	durations := make([]int64, len(rows))
	for i, r := range rows {
		durations[i] = int64(r.Duration)
	}

	bars, err := buildMultiScaleHistogram(durations)
	if err != nil {
		return IssuesDurationHistogram{}, err
	}

	return IssuesDurationHistogram{Bars: bars}, nil
}

func buildIssuesTimeSpentHistogram(rows []repository.IssueTimeSpent) (IssuesTimeSpentHistogram, error) {
	durations := make([]int64, len(rows))
	for i, r := range rows {
		durations[i] = int64(r.TimeSpent)
	}

	bars, err := buildMultiScaleHistogram(durations)
	if err != nil {
		return IssuesTimeSpentHistogram{}, err
	}

	return IssuesTimeSpentHistogram{Bars: bars}, nil
}

func buildStatusHistograms(rows []repository.StatusTransition) ([]StatusHistogram, error) {
	if len(rows) == 0 {
		return nil, ErrNoClosedIssues
	}

	byIssue := make(map[int][]repository.StatusTransition)
	for _, r := range rows {
		byIssue[r.IssueID] = append(byIssue[r.IssueID], r)
	}

	durationsByStatus := make(map[string][]int64)

	for _, transitions := range byIssue {
		sort.Slice(transitions, func(i, j int) bool {
			return transitions[i].ChangeTime.Before(transitions[j].ChangeTime)
		})

		for i := 0; i+1 < len(transitions); i++ {
			d := transitions[i+1].ChangeTime.Unix() - transitions[i].ChangeTime.Unix()
			if d < 0 {
				d = 0
			}
			durationsByStatus[transitions[i].FromStatus] = append(durationsByStatus[transitions[i].FromStatus], d)
		}
	}

	statuses := make([]string, 0, len(durationsByStatus))
	for s := range durationsByStatus {
		statuses = append(statuses, s)
	}
	sort.Strings(statuses)

	result := make([]StatusHistogram, 0, len(statuses))
	for _, status := range statuses {
		bars, err := buildMultiScaleHistogram(durationsByStatus[status])
		if err != nil {
			continue
		}
		result = append(result, StatusHistogram{Status: status, Bars: bars})
	}

	return result, nil
}

func buildDailyActivityChart(rows []repository.DailyActivity) (DailyActivityChart, error) {
	if len(rows) == 0 {
		return DailyActivityChart{}, nil
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date.Before(rows[j].Date)
	})

	points := make([]DailyActivityPoint, len(rows))
	var cumCreated, cumClosed int

	for i, r := range rows {
		cumCreated += r.Creation
		cumClosed += r.Completion

		points[i] = DailyActivityPoint{
			Date:              r.Date,
			Created:           r.Creation,
			Closed:            r.Completion,
			CumulativeCreated: cumCreated,
			CumulativeClosed:  cumClosed,
		}
	}

	return DailyActivityChart{Points: points}, nil
}

func buildPriorityChart(rows []repository.PriorityStats) PriorityChart {
	bars := make([]PriorityBar, len(rows))
	for i, r := range rows {
		bars[i] = PriorityBar{Priority: r.Priority, Count: r.Count}
	}
	return PriorityChart{Bars: bars}
}
