package habits

import (
	"fmt"
	"strings"
)

func (model Model) SummaryView(width, height int, focused bool) string {
	if model.err != nil {
		return "error loading habits"
	}
	if len(model.habits) == 0 {
		return "No habits"
	}

	var lines []string
	for _, habit := range model.habits {
		if !habit.DueToday {
			continue
		}

		lines = append(lines, habitSummaryRow(habit))
		if len(lines) >= height {
			break
		}
	}

	if len(lines) == 0 {
		return "No habits due today"
	}
	return strings.Join(lines, "\n")
}

func (model Model) ExpandedView(width, height int) string {
	if model.err != nil {
		return "Error loading habits:\n" + model.err.Error()
	}
	if len(model.habits) == 0 {
		return "No habits yet"
	}

	dayCount := historyDayCount(width)
	lines := []string{"Habits  left/right: day  space: cycle  c/s/x: complete/skip/incomplete  t: today  d: delete", ""}
	for i, habit := range model.habits {
		cursor := "  "
		if i == model.selected {
			cursor = "> "
		}
		lines = append(lines, model.habitExpandedLines(cursor, habit, dayCount)...)
	}

	return strings.Join(lines, "\n")
}

func habitSummaryRow(status HabitStatus) string {
	return fmt.Sprintf("%s %s  %s", habitCheckbox(status.DoneToday), status.Habit.Name, streakLabel(status.Streak))
}

func (model Model) habitExpandedLines(cursor string, status HabitStatus, dayCount int) []string {
	title := habitTitle(cursor, status)
	history := displayedHistory(status.History, model.selectedDayOffset, dayCount)
	selectedDay := selectedHistoryDay(status.History, model.selectedDayOffset)
	completeCount, dueCount := completionStats(status.History)

	meta := habitMeta(status, selectedDay, completeCount, dueCount)
	return []string{
		title,
		meta,
		"    " + habitMetaStyle.Render("days") + "  " + historyDateRow(history),
		"    " + habitMetaStyle.Render("log ") + "  " + historyStatusRow(history, selectedDay.Date),
		"",
	}
}

func habitTitle(cursor string, status HabitStatus) string {
	style := habitHeaderStyle
	if strings.TrimSpace(cursor) == ">" {
		style = habitActiveStyle
	}
	return style.Render(fmt.Sprintf("%s%s %s", cursor, habitCheckbox(status.DoneToday), status.Habit.Name))
}

func habitMeta(status HabitStatus, day DayStatus, completeCount, dueCount int) string {
	return habitMetaStyle.Render(fmt.Sprintf("    %s    %s    %d/%d complete    %s: %s",
		habitDaysLabel(status.Habit),
		streakLabel(status.Streak),
		completeCount,
		dueCount,
		day.Date.Format("Mon Jan 02"),
		statusLabel(day.Status),
	))
}

func habitCheckbox(completed bool) string {
	if completed {
		return "[x]"
	}
	return "[ ]"
}

func streakLabel(streak int) string {
	if streak == 1 {
		return "1-day streak"
	}
	return fmt.Sprintf("%d-day streak", streak)
}
