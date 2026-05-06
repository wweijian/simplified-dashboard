package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var habitWeekdayLabels = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
var habitDayPickerOrder = []int{1, 2, 3, 4, 5, 6, 0}

func weekdayFromNumber(r rune) (int, bool) {
	switch r {
	case '1':
		return 1, true
	case '2':
		return 2, true
	case '3':
		return 3, true
	case '4':
		return 4, true
	case '5':
		return 5, true
	case '6':
		return 6, true
	case '7':
		return 0, true
	default:
		return 0, false
	}
}

func dayCursorForWeekday(weekday int) int {
	for i, value := range habitDayPickerOrder {
		if value == weekday {
			return i
		}
	}
	return 0
}

func habitDaysField(weekdays [7]bool, dayCursor int, width int, active bool) string {
	segments := make([]string, 0, len(habitDayPickerOrder))
	for i, weekday := range habitDayPickerOrder {
		segments = append(segments, habitDaySegment(weekdays, weekday, i, dayCursor, active))
	}
	return borderedField(width, active).Render(lipgloss.JoinHorizontal(lipgloss.Left, segments...))
}

func habitDaySegment(weekdays [7]bool, weekday, cursor, dayCursor int, active bool) string {
	display := habitWeekdayLabels[weekday]
	if weekdays[weekday] {
		display = "[" + display + "]"
	}

	style := lipgloss.NewStyle().Padding(0, 1)
	switch {
	case active && cursor == dayCursor:
		style = style.Bold(true).Underline(true).Foreground(color.accent)
	case weekdays[weekday]:
		style = style.Bold(true).Foreground(color.accent)
	default:
		style = style.Foreground(color.text)
	}
	return style.Render(display)
}
