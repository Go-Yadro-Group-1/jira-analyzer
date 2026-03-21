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
	maxYears       = 7
)

func durationToLabel(seconds int64) DurationLabel {
	switch {
	case seconds < secondsInDay:
		return HourLabel(seconds / secondsInHour)
	case seconds < secondsInMonth:
		return DayLabel(seconds / secondsInDay)
	case seconds < secondsInYear:
		return MonthLabel(seconds / secondsInMonth)
	case seconds < int64(maxYears+1)*secondsInYear:
		return YearLabel(seconds / secondsInYear)
	default:
		return LabelOverflowYear
	}
}

func buildDurationHistogram(durations []int64) ([]Bar, error) {
	if len(durations) == 0 {
		return nil, ErrNoClosedIssues
	}

	counts := make(map[DurationLabel]int)
	for _, d := range durations {
		counts[durationToLabel(d)]++
	}

	bars := make([]Bar, 0, len(counts))
	for label, count := range counts {
		bars = append(bars, Bar{Label: label, Height: count})
	}

	sort.Slice(bars, func(i, j int) bool {
		return labelOrder(bars[i].Label) < labelOrder(bars[j].Label)
	})

	return bars, nil
}

// labelOrder возвращает числовое значение метки для сортировки.
func labelOrder(label DurationLabel) int64 {
	var n int64

	s := string(label)

	if label == LabelOverflowYear {
		return int64(maxYears+1) * secondsInYear
	}

	switch {
	case len(s) > 1 && s[len(s)-1] == 'h':
		fmt.Sscanf(s, "%dh", &n)
		return n * secondsInHour
	case len(s) > 3 && s[len(s)-3:] == "day":
		fmt.Sscanf(s, "%dday", &n)
		return n * secondsInDay
	case len(s) > 5 && s[len(s)-5:] == "month":
		fmt.Sscanf(s, "%dmonth", &n)
		return n * secondsInMonth
	case len(s) > 4 && s[len(s)-4:] == "year":
		fmt.Sscanf(s, "%dyear", &n)
		return n * secondsInYear
	}

	return 0
}

func buildIssuesDurationHistogram(rows []repository.IssueDuration) (IssuesDurationHistogram, error) {
	durations := make([]int64, len(rows))
	for i, r := range rows {
		durations[i] = r.Duration
	}

	bars, err := buildDurationHistogram(durations)
	if err != nil {
		return IssuesDurationHistogram{}, err
	}

	return IssuesDurationHistogram{Bars: bars}, nil
}

func buildIssuesTimeSpentHistogram(rows []repository.IssueTimeSpent) (IssuesTimeSpentHistogram, error) {
	durations := make([]int64, len(rows))
	for i, r := range rows {
		durations[i] = r.TimeSpent
	}

	bars, err := buildDurationHistogram(durations)
	if err != nil {
		return IssuesTimeSpentHistogram{}, err
	}

	return IssuesTimeSpentHistogram{Bars: bars}, nil
}

func buildStatusHistograms(rows []repository.StatusTransition) ([]StatusHistogram, error) {
	if len(rows) == 0 {
		return nil, ErrNoClosedIssues
	}

	// группируем переходы по состоянию FromStatus и считаем длительность
	// как разницу между последовательными переходами
	type transitionKey struct {
		status string
	}

	durationsByStatus := make(map[string][]int64)

	for i := 0; i+1 < len(rows); i++ {
		duration := rows[i+1].ChangeTime.Unix() - rows[i].ChangeTime.Unix()
		if duration < 0 {
			duration = 0
		}

		durationsByStatus[rows[i].FromStatus] = append(durationsByStatus[rows[i].FromStatus], duration)
	}

	statuses := make([]string, 0, len(durationsByStatus))
	for s := range durationsByStatus {
		statuses = append(statuses, s)
	}

	sort.Strings(statuses)

	result := make([]StatusHistogram, 0, len(statuses))

	for _, status := range statuses {
		bars, err := buildDurationHistogram(durationsByStatus[status])
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
	bars := make([]Bar, len(rows))
	for i, r := range rows {
		bars[i] = Bar{Label: r.Priority, Height: r.Count}
	}

	return PriorityChart{Bars: bars}
}
