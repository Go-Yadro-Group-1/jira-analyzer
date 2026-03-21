package service

import (
	"errors"
	"sort"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
)

var ErrNoClosedIssues = errors.New("no closed issues to build histogram")

const (
	secondsInMinute = 60
	secondsInHour   = 3600
	secondsInDay    = 86400
	secondsInMonth  = 30 * secondsInDay
	secondsInYear   = 365 * secondsInDay
)

func chooseUnit(maxDuration int64) DurationUnit {
	switch {
	case maxDuration < secondsInHour:
		return UnitMinute
	case maxDuration < secondsInDay:
		return UnitHour
	case maxDuration < secondsInMonth:
		return UnitDay
	case maxDuration < secondsInYear:
		return UnitMonth
	default:
		return UnitYear
	}
}

func durationToBucket(seconds int64, unit DurationUnit) int {
	switch unit {
	case UnitMinute:
		return int(seconds / secondsInMinute)
	case UnitHour:
		return int(seconds / secondsInHour)
	case UnitDay:
		return int(seconds / secondsInDay)
	case UnitMonth:
		return int(seconds / secondsInMonth)
	case UnitYear:
		bucket := int(seconds / secondsInYear)
		if bucket >= MaxYearBars {
			return MaxYearBars - 1
		}
		return bucket
	default:
		return 0
	}
}

func buildDurationHistogram(durations []int64) (DurationUnit, []int, error) {
	if len(durations) == 0 {
		return 0, nil, ErrNoClosedIssues
	}

	var maxDuration int64
	for _, d := range durations {
		if d > maxDuration {
			maxDuration = d
		}
	}

	unit := chooseUnit(maxDuration)
	size := durationToBucket(maxDuration, unit) + 1

	if unit == UnitYear && size < MaxYearBars {
		size = MaxYearBars
	}

	bars := make([]int, size)
	for _, d := range durations {
		bars[durationToBucket(d, unit)]++
	}

	return unit, bars, nil
}

func buildIssuesDurationHistogram(rows []repository.IssueDuration) (IssuesDurationHistogram, error) {
	durations := make([]int64, len(rows))
	for i, r := range rows {
		durations[i] = r.Duration
	}

	unit, bars, err := buildDurationHistogram(durations)
	if err != nil {
		return IssuesDurationHistogram{}, err
	}

	return IssuesDurationHistogram{Unit: unit, Bars: bars}, nil
}

func buildIssuesTimeSpentHistogram(rows []repository.IssueTimeSpent) (IssuesTimeSpentHistogram, error) {
	durations := make([]int64, len(rows))
	for i, r := range rows {
		durations[i] = r.TimeSpent
	}

	unit, bars, err := buildDurationHistogram(durations)
	if err != nil {
		return IssuesTimeSpentHistogram{}, err
	}

	return IssuesTimeSpentHistogram{Unit: unit, Bars: bars}, nil
}

func buildStatusHistograms(rows []repository.StatusTransition) ([]StatusHistogram, error) {
	if len(rows) == 0 {
		return nil, ErrNoClosedIssues
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
		unit, bars, err := buildDurationHistogram(durationsByStatus[status])
		if err != nil {
			continue
		}

		result = append(result, StatusHistogram{Status: status, Unit: unit, Bars: bars})
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
