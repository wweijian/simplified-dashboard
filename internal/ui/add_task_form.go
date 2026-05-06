package ui

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	tasks "simplified-dashboard/internal/tasks"

	bbt "github.com/charmbracelet/bubbletea"
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
	return addTaskForm{priority: 1}
}

func newEditTaskForm(task tasks.Task, now time.Time) addTaskForm {
	return addTaskForm{
		title:       task.Title,
		description: nullableStringValue(task.Description),
		dueDays:     dueDaysFromDate(task.DueDate, now),
		priority:    task.Priority,
	}
}

func (form addTaskForm) Update(msg bbt.KeyMsg) addTaskForm {
	form.errorText = ""

	switch msg.String() {
	case "tab", "down":
		form.field = (form.field + 1) % addTaskFieldCount
	case "shift+tab", "up":
		form.field = (form.field + addTaskFieldCount - 1) % addTaskFieldCount
	case "left":
		form = form.movePriority(-1)
	case "right":
		form = form.movePriority(1)
	case "backspace", "ctrl+h":
		form = form.deleteLastRune()
	case "ctrl+u":
		form = form.clearField()
	default:
		form = form.appendRunes(msg.Runes)
	}
	return form
}

func (form addTaskForm) Submit(now time.Time) (addTaskForm, tasks.CreateTaskInput, bool) {
	title := strings.TrimSpace(form.title)
	if !form.validateTitle(title) {
		return form, tasks.CreateTaskInput{}, false
	}

	description := strings.TrimSpace(form.description)
	if !form.validateDescription(description) {
		return form, tasks.CreateTaskInput{}, false
	}

	days, ok := form.dueDaysValue()
	if !ok {
		return form, tasks.CreateTaskInput{}, false
	}

	return form, tasks.CreateTaskInput{
		Title:       title,
		Description: sql.NullString{String: description, Valid: description != ""},
		DueDate:     sql.NullString{String: now.AddDate(0, 0, days).Format("2006-01-02"), Valid: true},
		Priority:    form.priority,
	}, true
}

func (form *addTaskForm) validateTitle(title string) bool {
	if title == "" {
		form.errorText = "Title is required"
		return false
	}
	if len([]rune(title)) > maxTaskTitleLength {
		form.errorText = titleLengthError()
		return false
	}
	return true
}

func (form *addTaskForm) validateDescription(description string) bool {
	if len([]rune(description)) > maxTaskDescriptionChars {
		form.errorText = descriptionLengthError()
		return false
	}
	return true
}

func (form *addTaskForm) dueDaysValue() (int, bool) {
	if form.dueDays == "" {
		form.errorText = "Due date is required"
		return 0, false
	}

	days, err := strconv.Atoi(form.dueDays)
	if err != nil {
		form.errorText = "Due date must be a number of days"
		return 0, false
	}
	if days > maxTaskDueDays {
		form.errorText = "Due date must be within 365 days"
		return 0, false
	}
	return days, true
}

func (form addTaskForm) View(width int) string {
	return form.ViewWithTitle(width, "Add task")
}

func (form addTaskForm) ViewWithTitle(width int, title string) string {
	modalWidth := modalWidth(width)
	fieldWidth := max(modalWidth-12, 12)

	lines := []string{
		formTitleStyle().Render(title),
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
	return renderModal(modalWidth, lines)
}

func titleLengthError() string {
	return fmt.Sprintf("Title must be %d characters or fewer", maxTaskTitleLength)
}

func descriptionLengthError() string {
	return fmt.Sprintf("Description must be %d characters or fewer", maxTaskDescriptionChars)
}
