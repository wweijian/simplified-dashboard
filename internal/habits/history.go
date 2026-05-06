package habits

import (
	"fmt"
	"strings"
	"time"
)

func habitHistory(habit Habit, logs map[string]HabitLogStatus, today time.Time) []DayStatus {
	dayCount := habitHistoryDayCount(habit, today)
	days := make([]DayStatus, 0, dayCount)
	for offset := 0; offset < dayCount; offset++ {
		day := today.AddDate(0, 0, -offset)
		key := dateString(day)
		due := habitDueOn(habit, day)
		days = append(days, DayStatus{
			Date:   day,
			Due:    due,
			Status: dayLogStatus(logs, key, due),
		})
	}
	return days
}

func habitHistoryDayCount(habit Habit, today time.Time) int {
	createdAt, ok := parseHabitCreatedAt(habit.CreatedAt)
	if !ok {
		return maxHabitHistoryDays
	}

	createdDate := createdAt.In(time.Local).Truncate(24 * time.Hour)
	today = today.In(time.Local).Truncate(24 * time.Hour)
	if createdDate.After(today) {
		return 1
	}

	daysSinceCreation := int(today.Sub(createdDate).Hours()/24) + 1
	return min(max(daysSinceCreation, 1), maxHabitHistoryDays)
}

func dayLogStatus(logs map[string]HabitLogStatus, key string, due bool) HabitLogStatus {
	if status, ok := logs[key]; ok {
		return status
	}
	if due {
		return LogIncomplete
	}
	return ""
}

func historyDayCount(width int) int {
	count := (width - 12) / 4
	return min(max(count, 7), maxHabitHistoryDays)
}

func displayedHistory(history []DayStatus, selectedOffset, count int) []DayStatus {
	if len(history) == 0 {
		return nil
	}
	count = min(count, len(history))
	selectedOffset = min(max(selectedOffset, 0), len(history)-1)

	startOffset := selectedOffset
	endOffset := selectedOffset + count - 1
	if endOffset >= len(history) {
		endOffset = len(history) - 1
		startOffset = max(endOffset-count+1, 0)
	}

	days := make([]DayStatus, 0, count)
	for offset := endOffset; offset >= startOffset; offset-- {
		days = append(days, history[offset])
	}
	return days
}

func selectedHistoryDay(history []DayStatus, selectedOffset int) DayStatus {
	if len(history) == 0 {
		return DayStatus{Date: currentDate(), Due: true, Status: LogIncomplete}
	}

	selectedOffset = min(max(selectedOffset, 0), len(history)-1)
	return history[selectedOffset]
}

func historyDateRow(history []DayStatus) string {
	parts := make([]string, 0, len(history))
	for _, day := range history {
		label := fmt.Sprintf("%-3s", weekdayLabel(int(day.Date.Weekday())))
		parts = append(parts, habitMetaStyle.Render(label))
	}
	return strings.Join(parts, " ")
}

func historyStatusRow(history []DayStatus, selectedDate time.Time) string {
	parts := make([]string, 0, len(history))
	selectedKey := dateString(selectedDate)
	for _, day := range history {
		parts = append(parts, historyStatusCell(day, selectedKey))
	}
	return strings.Join(parts, " ")
}

func historyStatusCell(day DayStatus, selectedKey string) string {
	symbol := dayStatusSymbol(day)
	if dateString(day.Date) == selectedKey {
		return habitSelectedStyle.Render("[" + symbol + "]")
	}
	return " " + styledDayStatusSymbol(day) + " "
}

func dayStatusSymbol(day DayStatus) string {
	switch day.Status {
	case LogComplete:
		return "●"
	case LogSkipped:
		return "–"
	case LogIncomplete:
		return "○"
	default:
		return "·"
	}
}

func styledDayStatusSymbol(day DayStatus) string {
	symbol := dayStatusSymbol(day)
	switch day.Status {
	case LogComplete:
		return habitCompleteStyle.Render(symbol)
	case LogSkipped:
		return habitSkippedStyle.Render(symbol)
	case LogIncomplete:
		return habitIncompleteStyle.Render(symbol)
	default:
		return habitMutedStyle.Render(symbol)
	}
}

func statusLabel(status HabitLogStatus) string {
	switch status {
	case LogComplete:
		return "complete"
	case LogSkipped:
		return "skipped"
	case LogIncomplete:
		return "incomplete"
	default:
		return "not scheduled"
	}
}

func completionStats(history []DayStatus) (completeCount, dueCount int) {
	for _, day := range history {
		if !day.Due && day.Status == "" {
			continue
		}
		if day.Status == LogSkipped {
			continue
		}
		dueCount++
		if day.Status == LogComplete {
			completeCount++
		}
	}
	return completeCount, dueCount
}

func habitStreak(habit Habit, logs map[string]HabitLogStatus, today time.Time) int {
	streak := 0
	for daysBack := 0; daysBack < 366; daysBack++ {
		day := today.AddDate(0, 0, -daysBack)
		if !habitDueOn(habit, day) {
			continue
		}

		switch logs[dateString(day)] {
		case LogComplete:
			streak++
		case LogSkipped:
			continue
		default:
			return streak
		}
	}
	return streak
}
