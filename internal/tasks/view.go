package tasks

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var currentDate = func() time.Time {
	return time.Now().In(time.Local).Truncate(24 * time.Hour)
}

func (model Model) SummaryView(width, height int, focused bool) string {
	if model.err != nil {
		return "error loading tasks"
	}

	tasks := model.summaryTasks()
	if len(tasks) == 0 {
		return "No tasks"
	}

	var lines []string
	overdueCount := 0
	for _, task := range tasks {
		days, hasDueDate := taskDaysUntilDue(task)
		if task.Completed || !hasDueDate {
			continue
		}

		if days < 0 {
			overdueCount++
		}
	}

	if overdueCount > 0 {
		lines = append(lines, fmt.Sprintf("%d overdue", overdueCount))
	}

	for _, task := range tasks {
		days, hasDueDate := taskDaysUntilDue(task)
		if task.Completed || !hasDueDate || days != 0 {
			continue
		}
		lines = append(lines, taskRow("", task))

		if len(lines) >= height {
			break
		}
	}

	if len(lines) == 0 {
		if hasIncompleteTask(tasks) {
			return "No tasks due today"
		}
		return "All done"
	}

	return strings.Join(lines, "\n")
}

func (model Model) ExpandedView(width, height int) string {
	if model.err != nil {
		return "Error loading tasks:\n" + model.err.Error()
	}

	if len(model.tasks) == 0 {
		if len(model.allTasks) > 0 {
			return fmt.Sprintf("Filter: %s  Sort: %s\n\nNo tasks in this filter", model.FilterLabel(), model.SortLabel())
		}
		return "No tasks yet"
	}

	lines := []string{
		fmt.Sprintf("Filter: %s  Sort: %s", model.FilterLabel(), model.SortLabel()),
		"",
	}
	for i, task := range model.tasks {
		cursor := "  "
		if i == model.selected {
			cursor = "> "
		}
		lines = append(lines, taskRow(cursor, task))
	}
	return strings.Join(lines, "\n")
}

func (model Model) summaryTasks() []Task {
	if len(model.allTasks) > 0 {
		return model.allTasks
	}
	return model.tasks
}

func hasIncompleteTask(tasks []Task) bool {
	for _, task := range tasks {
		if !task.Completed {
			return true
		}
	}
	return false
}

func taskRow(cursor string, task Task) string {
	line := fmt.Sprintf("%s%s %-3s", cursor, checkbox(task), priorityLabel(task.Priority))
	dueText := dueDateText(task)
	if dueText != "" {
		line += " " + dueText
	}
	line += " " + task.Title

	if task.Completed {
		line = lipgloss.NewStyle().Strikethrough(true).Render(line)
	}
	return line
}

func priorityLabel(priority int) string {
	switch priority {
	case 2:
		return "!!!"
	case 0:
		return "!"
	default:
		return "!!"
	}
}

func dueDateText(task Task) string {
	days, ok := taskDaysUntilDue(task)
	if !ok {
		return ""
	}

	return relativeDueDate(days)
}

func daysUntilDue(dueDate time.Time) int {
	return int(dueDate.Sub(currentDate()).Hours() / 24)
}

func relativeDueDate(days int) string {
	switch days {
	case 0:
		return "today"
	case 1:
		return "tomorrow"
	default:
		if days > 0 {
			if days%7 == 0 {
				weeks := days / 7
				return fmt.Sprintf("in %d weeks", weeks)
			}

			return fmt.Sprintf("in %d days", days)
		}

		return fmt.Sprintf("%d days ago", -days)
	}
}

func checkbox(task Task) string {
	if task.Completed {
		return "[x]"
	}
	return "[ ]"
}
