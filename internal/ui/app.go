package ui

import (
	calendar "simplified-dashboard/internal/calendar"
	"simplified-dashboard/internal/db"
	finance "simplified-dashboard/internal/finance"
	"simplified-dashboard/internal/habits"
	tasks "simplified-dashboard/internal/tasks"

	bbt "github.com/charmbracelet/bubbletea"
)

type Panel interface {
	SummaryView(width, height int, focused bool) string
	ExpandedView(width, height int) string
}

type ViewMode int

const (
	ModeNormal ViewMode = iota
	ModeAddTask
	ModeEditTask
	ModeAddHabit
	ModeEditHabit
)

type Model struct {
	width          int
	height         int
	activePanel    int
	mode           ViewMode
	addTaskForm    addTaskForm
	addHabitForm   addHabitForm
	editingTaskID  int64
	editingHabitID int64
	calendar       Panel
	tasks          tasks.Model
	finance        finance.Model
	habits         habits.Model
}

func New(database *db.DB) Model {
	return Model{
		activePanel:  int(Calendar),
		mode:         ModeNormal,
		addTaskForm:  newAddTaskForm(),
		addHabitForm: newAddHabitForm(),
		calendar:     calendar.New(),
		tasks:        tasks.New(tasks.NewStore(database)),
		finance:      finance.New(finance.NewStore(database)),
		habits:       habits.New(habits.NewStore(database)),
	}
}

func (model Model) Init() bbt.Cmd { return nil }

func (model Model) View() string {
	if model.width == 0 {
		return ""
	}
	return renderLayout(model)
}
