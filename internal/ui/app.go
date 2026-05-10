package ui

import (
	"context"
	"time"

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
	ModeAddFinanceTransaction
	ModeEditFinanceTransaction
	ModeAddHabit
	ModeEditHabit
)

type Model struct {
	width                       int
	height                      int
	activePanel                 int
	mode                        ViewMode
	addTaskForm                 addTaskForm
	addFinanceForm              addFinanceTransactionForm
	addHabitForm                addHabitForm
	editingTaskID               int64
	editingFinanceTransactionID int64
	editingHabitID              int64
	calendar                    calendar.Model
	tasks                       tasks.Model
	finance                     finance.Model
	habits                      habits.Model
}

func New(database *db.DB) Model {
	return Model{
		activePanel:    int(Calendar),
		mode:           ModeNormal,
		addTaskForm:    newAddTaskForm(),
		addFinanceForm: newAddFinanceTransactionForm(nil, time.Now()),
		addHabitForm:   newAddHabitForm(),
		calendar:       calendar.New().StartLoading(),
		tasks:          tasks.New(tasks.NewStore(database)),
		finance:        finance.New(finance.NewStore(database)),
		habits:         habits.New(habits.NewStore(database)),
	}
}

type calendarLoadedMsg struct {
	result calendar.LoadResult
}

func (model Model) Init() bbt.Cmd {
	return model.loadCalendarCmd()
}

func (model Model) loadCalendarCmd() bbt.Cmd {
	calendarModel := model.calendar
	return func() bbt.Msg {
		return calendarLoadedMsg{result: calendarModel.LoadEvents(context.Background())}
	}
}

func (model Model) View() string {
	if model.width == 0 {
		return ""
	}
	return renderLayout(model)
}
