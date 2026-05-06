package habits

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

func habitDueOn(habit Habit, day time.Time) bool {
	switch Frequency(habit.Frequency) {
	case FrequencyDaily:
		return true
	case FrequencyWeekly:
		createdAt, ok := parseHabitCreatedAt(habit.CreatedAt)
		return !ok || day.Weekday() == createdAt.Weekday()
	case FrequencyCustom:
		return customHabitDueOn(habit.FrequencyDays, day)
	default:
		return true
	}
}

func customHabitDueOn(frequencyDays sql.NullString, day time.Time) bool {
	todayWeekday := int(day.Weekday())
	for _, weekday := range customHabitWeekdays(frequencyDays) {
		if weekday == todayWeekday {
			return true
		}
	}
	return false
}

func habitDaysLabel(habit Habit) string {
	switch Frequency(habit.Frequency) {
	case FrequencyDaily:
		return "every day"
	case FrequencyWeekly:
		return weeklyHabitLabel(habit)
	case FrequencyCustom:
		return customHabitDaysLabel(habit.FrequencyDays)
	default:
		return habit.Frequency
	}
}

func weeklyHabitLabel(habit Habit) string {
	createdAt, ok := parseHabitCreatedAt(habit.CreatedAt)
	if !ok {
		return "weekly"
	}
	return weekdayLabel(int(createdAt.Weekday()))
}

func customHabitDaysLabel(frequencyDays sql.NullString) string {
	weekdays := orderedWeekdays(customHabitWeekdays(frequencyDays))
	if len(weekdays) == 0 {
		return "no days"
	}

	labels := make([]string, 0, len(weekdays))
	for _, weekday := range weekdays {
		labels = append(labels, weekdayLabel(weekday))
	}
	return strings.Join(labels, " ")
}

func customHabitWeekdays(frequencyDays sql.NullString) []int {
	if !frequencyDays.Valid || strings.TrimSpace(frequencyDays.String) == "" {
		return nil
	}

	var weekdays []int
	if err := json.Unmarshal([]byte(frequencyDays.String), &weekdays); err != nil {
		return nil
	}
	return weekdays
}

func orderedWeekdays(weekdays []int) []int {
	seen := make(map[int]bool, len(weekdays))
	for _, weekday := range weekdays {
		if weekday >= 0 && weekday <= 6 {
			seen[weekday] = true
		}
	}
	return weekdaysInDisplayOrder(seen)
}

func weekdaysInDisplayOrder(seen map[int]bool) []int {
	order := []int{1, 2, 3, 4, 5, 6, 0}
	ordered := make([]int, 0, len(seen))
	for _, weekday := range order {
		if seen[weekday] {
			ordered = append(ordered, weekday)
		}
	}
	return ordered
}

func weekdayLabel(weekday int) string {
	labels := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if weekday < 0 || weekday >= len(labels) {
		return "?"
	}
	return labels[weekday]
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

func dateString(day time.Time) string {
	return day.Format("2006-01-02")
}
