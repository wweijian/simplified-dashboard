package ui

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"simplified-dashboard/internal/habits"
)

func weekdaysFromHabit(habit habits.Habit) [7]bool {
	switch habits.Frequency(habit.Frequency) {
	case habits.FrequencyDaily:
		return allWeekdays()
	case habits.FrequencyWeekly:
		return weekdayFromCreatedAt(habit.CreatedAt)
	default:
		return weekdaysFromNullString(habit.FrequencyDays)
	}
}

func allWeekdays() [7]bool {
	var weekdays [7]bool
	for i := range weekdays {
		weekdays[i] = true
	}
	return weekdays
}

func weekdayFromCreatedAt(value string) [7]bool {
	var weekdays [7]bool
	createdAt, ok := parseHabitCreatedAt(value)
	if ok {
		weekdays[int(createdAt.Weekday())] = true
	}
	return weekdays
}

func weekdaysFromNullString(value sql.NullString) [7]bool {
	var weekdays [7]bool
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return weekdays
	}

	var selected []int
	if err := json.Unmarshal([]byte(value.String), &selected); err != nil {
		return weekdays
	}

	for _, weekday := range selected {
		if weekday >= 0 && weekday < len(weekdays) {
			weekdays[weekday] = true
		}
	}
	return weekdays
}

func parseHabitCreatedAt(value string) (time.Time, bool) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}
