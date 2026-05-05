package ui

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	tasks "simplified-dashboard/internal/tasks"

	bbt "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type addTaskField int

const (
	addTaskTitle addTaskField = iota
	addTaskDueDays
	addTaskPriority
	addTaskDescription
)

const (
	addTaskFieldCount       = 4
	maxTaskTitleLength      = 50
	maxTaskDescriptionChars = 200
	descriptionVisibleLines = 4
	maxTaskDueDays          = 365
)

type addTaskForm struct {
	title       string
	description string
	dueDays     string
	priority    int
	field       addTaskField
	errorText   string
}

func newAddTaskForm() addTaskForm {
	return addTaskForm{
		priority: 1,
	}
}

func (form addTaskForm) Update(msg bbt.KeyMsg) addTaskForm {
	form.errorText = ""

	switch msg.String() {
	case "tab", "down":
		form.field = (form.field + 1) % addTaskFieldCount
		return form
	case "shift+tab", "up":
		form.field = (form.field + addTaskFieldCount - 1) % addTaskFieldCount
		return form
	case "left":
		if form.field == addTaskPriority {
			form.priority = (form.priority + 2) % 3
		}
		return form
	case "right":
		if form.field == addTaskPriority {
			form.priority = (form.priority + 1) % 3
		}
		return form
	case "backspace", "ctrl+h":
		return form.deleteLastRune()
	case "ctrl+u":
		return form.clearField()
	}

	for _, r := range msg.Runes {
		form = form.appendRune(r)
	}
	return form
}

func (form addTaskForm) Submit(now time.Time) (addTaskForm, tasks.CreateTaskInput, bool) {
	title := strings.TrimSpace(form.title)
	if title == "" {
		form.errorText = "Title is required"
		return form, tasks.CreateTaskInput{}, false
	}
	if len([]rune(title)) > maxTaskTitleLength {
		form.errorText = titleLengthError()
		return form, tasks.CreateTaskInput{}, false
	}

	description := strings.TrimSpace(form.description)
	if len([]rune(description)) > maxTaskDescriptionChars {
		form.errorText = descriptionLengthError()
		return form, tasks.CreateTaskInput{}, false
	}

	if form.dueDays == "" {
		form.errorText = "Due date is required"
		return form, tasks.CreateTaskInput{}, false
	}

	days, err := strconv.Atoi(form.dueDays)
	if err != nil || days < 0 {
		form.errorText = "Due date must be a non-negative number of days"
		return form, tasks.CreateTaskInput{}, false
	}
	if days > maxTaskDueDays {
		form.errorText = "Due date must be within 365 days"
		return form, tasks.CreateTaskInput{}, false
	}

	dueDate := now.AddDate(0, 0, days).Format("2006-01-02")
	input := tasks.CreateTaskInput{
		Title:       title,
		Description: sql.NullString{String: description, Valid: description != ""},
		DueDate:     sql.NullString{String: dueDate, Valid: true},
		Priority:    form.priority,
	}
	return form, input, true
}

func (form addTaskForm) View(width int) string {
	modalWidth := addTaskModalWidth(width)
	fieldWidth := max(modalWidth-12, 12)

	lines := []string{
		formTitleStyle().Render("Add task"),
		fieldLabel("Title", form.field == addTaskTitle),
		inputField(form.title, fieldWidth, form.field == addTaskTitle),
		fieldLabel("Due in days", form.field == addTaskDueDays),
		inputField(form.dueDays, fieldWidth, form.field == addTaskDueDays),
		fieldLabel("Priority", form.field == addTaskPriority),
		priorityField(form.priority, fieldWidth, form.field == addTaskPriority),
		fieldLabel("Description", form.field == addTaskDescription),
		textAreaField(form.description, fieldWidth, descriptionVisibleLines, form.field == addTaskDescription),
	}
	if form.errorText != "" {
		lines = append(lines, errorStyle().Render(form.errorText))
	}

	return lipgloss.NewStyle().
		Width(modalWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color.accent).
		Padding(1, 3).
		Render(strings.Join(lines, "\n"))
}

func addTaskModalWidth(width int) int {
	if width < 36 {
		return max(width-8, 20)
	}
	return min(width-8, 76)
}

func formTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(color.accent).
		MarginBottom(1)
}

func fieldLabel(label string, active bool) string {
	style := lipgloss.NewStyle().Foreground(color.text)
	if active {
		style = style.Foreground(color.accent).Bold(true)
	}
	return style.Render(label)
}

func inputField(value string, width int, active bool) string {
	if value == "" {
		value = " "
	}

	style := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)
	if active {
		style = style.BorderForeground(color.accent)
	} else {
		style = style.BorderForeground(color.text)
	}

	return style.Render(fitFieldValue(value, width-2))
}

func textAreaField(value string, width, lines int, active bool) string {
	contentWidth := max(width-2, 1)
	wrapped := wrapFieldValue(value, contentWidth)
	for len(wrapped) < lines {
		wrapped = append(wrapped, "")
	}
	if len(wrapped) > lines {
		wrapped = wrapped[:lines]
	}

	style := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)
	if active {
		style = style.BorderForeground(color.accent)
	} else {
		style = style.BorderForeground(color.text)
	}

	return style.Render(strings.Join(wrapped, "\n"))
}

func priorityField(priority int, width int, active bool) string {
	options := []string{"low", "medium", "high"}
	segments := make([]string, 0, len(options))
	for i, option := range options {
		style := lipgloss.NewStyle().Padding(0, 2)
		if i == priority {
			style = style.Bold(true).Foreground(color.accent)
		} else {
			style = style.Foreground(color.text)
		}
		segments = append(segments, style.Render(option))
	}

	style := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)
	if active {
		style = style.BorderForeground(color.accent)
	} else {
		style = style.BorderForeground(color.text)
	}

	return style.Render(lipgloss.JoinHorizontal(lipgloss.Left, segments...))
}

func errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(color.danger).
		MarginTop(1)
}

func fitFieldValue(value string, width int) string {
	if lipgloss.Width(value) <= width {
		return value
	}

	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width(string(runes))+3 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "..."
}

func wrapFieldValue(value string, width int) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{""}
	}

	var lines []string
	var current strings.Builder
	for _, word := range strings.Fields(value) {
		if lipgloss.Width(word) > width {
			lines = appendWrappedWord(lines, &current, word, width)
			continue
		}

		next := word
		if current.Len() > 0 {
			next = current.String() + " " + word
		}

		if lipgloss.Width(next) > width {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
			continue
		}

		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(word)
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func appendWrappedWord(lines []string, current *strings.Builder, word string, width int) []string {
	if current.Len() > 0 {
		lines = append(lines, current.String())
		current.Reset()
	}

	var chunk strings.Builder
	for _, r := range word {
		next := chunk.String() + string(r)
		if lipgloss.Width(next) > width {
			lines = append(lines, chunk.String())
			chunk.Reset()
		}
		chunk.WriteRune(r)
	}

	if chunk.Len() > 0 {
		current.WriteString(chunk.String())
	}
	return lines
}

func (form addTaskForm) appendRune(r rune) addTaskForm {
	switch form.field {
	case addTaskTitle:
		if len([]rune(form.title)) >= maxTaskTitleLength {
			form.errorText = titleLengthError()
			return form
		}
		form.title += string(r)
	case addTaskDescription:
		if len([]rune(form.description)) >= maxTaskDescriptionChars {
			form.errorText = descriptionLengthError()
			return form
		}
		form.description += string(r)
	case addTaskDueDays:
		if unicode.IsDigit(r) {
			form.dueDays += string(r)
		} else {
			form.errorText = "Due date accepts digits only"
		}
	case addTaskPriority:
		switch r {
		case '1', 'l', 'L':
			form.priority = 0
		case '2', 'm', 'M':
			form.priority = 1
		case '3', 'h', 'H':
			form.priority = 2
		case ' ', 'p', 'P':
			form.priority = (form.priority + 1) % 3
		}
	}
	return form
}

func (form addTaskForm) deleteLastRune() addTaskForm {
	switch form.field {
	case addTaskTitle:
		form.title = trimLastRune(form.title)
	case addTaskDescription:
		form.description = trimLastRune(form.description)
	case addTaskDueDays:
		form.dueDays = trimLastRune(form.dueDays)
	}
	return form
}

func (form addTaskForm) clearField() addTaskForm {
	switch form.field {
	case addTaskTitle:
		form.title = ""
	case addTaskDescription:
		form.description = ""
	case addTaskDueDays:
		form.dueDays = ""
	}
	return form
}

func trimLastRune(value string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)
	return string(runes[:len(runes)-1])
}

func titleLengthError() string {
	return fmt.Sprintf("Title must be %d characters or fewer", maxTaskTitleLength)
}

func descriptionLengthError() string {
	return fmt.Sprintf("Description must be %d characters or fewer", maxTaskDescriptionChars)
}
