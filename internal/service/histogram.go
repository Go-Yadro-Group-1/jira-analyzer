package service

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/repository"
)

var ErrNoHistogramData = errors.New("no closed issues to build histogram")

const (
	secondsInHour  = 3600
	secondsInDay   = 86400
	secondsInMonth = 30 * secondsInDay
	secondsInYear  = 365 * secondsInDay
)

const (
	maxYears = 8

	hourBarsCount  = 24
	dayBarsCount   = 30
	monthBarsCount = 11

	dayBarOffset   = hourBarsCount - 1
	monthBarOffset = hourBarsCount + dayBarsCount - 1
	yearBarOffset  = hourBarsCount + dayBarsCount + monthBarsCount - 1
	maxYearIndex   = yearBarOffset + maxYears

	dayZoneEnd   = hourBarsCount + dayBarsCount
	monthZoneEnd = dayZoneEnd + monthBarsCount

	totalBars = hourBarsCount + dayBarsCount + monthBarsCount + maxYears
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
		y := seconds / secondsInYear
		if y >= maxYears {
			return "8+year"
		}

		return fmt.Sprintf("%dyear", y)
	}
}

func durationToBarIndex(seconds int64) int {
	switch {
	case seconds < secondsInDay:
		return int(seconds / secondsInHour)
	case seconds < secondsInMonth:
		return dayBarOffset + int(seconds/secondsInDay)
	case seconds < secondsInYear:
		return monthBarOffset + int(seconds/secondsInMonth)
	default:
		y := int(seconds / secondsInYear)
		if y >= maxYears {
			return maxYearIndex
		}

		return yearBarOffset + y
	}
}

func barIndexToSeconds(idx int) int64 {
	switch {
	case idx < hourBarsCount:
		return int64(idx) * secondsInHour
	case idx < dayZoneEnd:
		return int64(idx-dayBarOffset) * secondsInDay
	case idx < monthZoneEnd:
		return int64(idx-monthBarOffset) * secondsInMonth
	case idx < maxYearIndex:
		return int64(idx-yearBarOffset) * secondsInYear
	default:
		return maxYears * secondsInYear
	}
}

func buildMultiScaleHistogram(durations []int64) ([]HistogramBar, error) {
	if len(durations) == 0 {
		return nil, ErrNoHistogramData
	}

	var counts [totalBars]int

	maxIdx := 0

	for _, duration := range durations {
		idx := durationToBarIndex(duration)

		counts[idx]++

		if idx > maxIdx {
			maxIdx = idx
		}
	}

	bars := make([]HistogramBar, maxIdx+1)

	for idx := 0; idx <= maxIdx; idx++ {
		bars[idx] = HistogramBar{
			Label: durationToBarLabel(barIndexToSeconds(idx)),
			Count: counts[idx],
		}
	}

	return bars, nil
}

func buildIssuesDurationHistogram(
	rows []repository.IssueDuration,
) (IssuesDurationHistogram, error) {
	durations := make([]int64, len(rows))
	for i, row := range rows {
		durations[i] = int64(row.Duration)
	}

	bars, err := buildMultiScaleHistogram(durations)
	if err != nil {
		return IssuesDurationHistogram{}, err
	}

	return IssuesDurationHistogram{Bars: bars}, nil
}

func buildIssuesTimeSpentHistogram(
	rows []repository.IssueTimeSpent,
) (IssuesTimeSpentHistogram, error) {
	durations := make([]int64, len(rows))
	for i, row := range rows {
		durations[i] = int64(row.TimeSpent)
	}

	bars, err := buildMultiScaleHistogram(durations)
	if err != nil {
		return IssuesTimeSpentHistogram{}, err
	}

	return IssuesTimeSpentHistogram{Bars: bars}, nil
}

func buildStatusHistograms(rows []repository.StatusTransition) ([]StatusHistogram, error) {
	if len(rows) == 0 {
		return nil, ErrNoHistogramData
	}

	byIssue := make(map[int][]repository.StatusTransition)
	for _, row := range rows {
		byIssue[row.IssueID] = append(byIssue[row.IssueID], row)
	}

	durationsByStatus := make(map[string][]int64)

	for _, transitions := range byIssue {
		sort.Slice(transitions, func(i, j int) bool {
			return transitions[i].ChangeTime.Before(transitions[j].ChangeTime)
		})

		for i := 0; i+1 < len(transitions); i++ {
			d := transitions[i+1].ChangeTime.Unix() - transitions[i].ChangeTime.Unix()
			d = max(d, 0)
			durationsByStatus[transitions[i].FromStatus] = append(
				durationsByStatus[transitions[i].FromStatus],
				d,
			)
		}
	}

	statuses := make([]string, 0, len(durationsByStatus))
	for status := range durationsByStatus {
		statuses = append(statuses, status)
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
		return DailyActivityChart{Points: nil}, nil
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date.Before(rows[j].Date)
	})

	points := make([]DailyActivityPoint, len(rows))

	var cumCreated, cumClosed int

	for i, row := range rows {
		cumCreated += row.Creation
		cumClosed += row.Completion

		points[i] = DailyActivityPoint{
			Date:              row.Date,
			Created:           row.Creation,
			Closed:            row.Completion,
			CumulativeCreated: cumCreated,
			CumulativeClosed:  cumClosed,
		}
	}

	return DailyActivityChart{Points: points}, nil
}

func buildPriorityChart(rows []repository.PriorityStats) PriorityChart {
	bars := make([]PriorityBar, len(rows))
	for i, row := range rows {
		bars[i] = PriorityBar{Priority: row.Priority, Count: row.Count}
	}

	return PriorityChart{Bars: bars}
}
