package ui

import (
	"time"

	calendar "simplified-dashboard/internal/calendar"
	"simplified-dashboard/internal/db"
	finance "simplified-dashboard/internal/finance"
	habits "simplified-dashboard/internal/habits"
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
)

type Model struct {
	width         int
	height        int
	activePanel   int
	mode          ViewMode
	addTaskForm   addTaskForm
	editingTaskID int64
	calendar      Panel
	tasks         tasks.Model
	finance       Panel
	habits        Panel
}

func New(database *db.DB) Model {
	return Model{
		activePanel: int(Calendar),
		mode:        ModeNormal,
		addTaskForm: newAddTaskForm(),
		calendar:    calendar.New(),
		tasks:       tasks.New(tasks.NewStore(database)),
		finance:     finance.New(),
		habits:      habits.New(),
	}
}

func (model Model) Init() bbt.Cmd { return nil }

func (model Model) Update(msg bbt.Msg) (bbt.Model, bbt.Cmd) {
	switch msg := msg.(type) {
	case bbt.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
	case bbt.KeyMsg:
		key := msg.String()

		if model.mode == ModeAddTask || model.mode == ModeEditTask {
			switch key {
			case "esc":
				model.mode = ModeNormal
				model.addTaskForm = newAddTaskForm()
				model.editingTaskID = 0
			case "enter":
				form, input, ok := model.addTaskForm.Submit(time.Now())
				model.addTaskForm = form
				if ok {
					if model.mode == ModeEditTask {
						model.tasks = model.tasks.UpdateTask(model.editingTaskID, input)
					} else {
						model.tasks = model.tasks.Create(input)
					}
					model.mode = ModeNormal
					model.addTaskForm = newAddTaskForm()
					model.editingTaskID = 0
				}
			default:
				model.addTaskForm = model.addTaskForm.Update(msg)
			}
			return model, nil
		}

		if key == "a" {
			if model.activePanel == int(Tasks) {
				model.mode = ModeAddTask
				model.addTaskForm = newAddTaskForm()
			}
			return model, nil
		}

		if key == "e" && model.activePanel == int(Tasks) {
			if task, ok := model.tasks.SelectedTask(); ok {
				model.mode = ModeEditTask
				model.editingTaskID = task.ID
				model.addTaskForm = newEditTaskForm(task, time.Now())
			}
			return model, nil
		}

		if handler, ok := keyBindings[key]; ok {
			return handler(model)
		}
		if model.activePanel == int(Tasks) {
			model.tasks = model.tasks.Update(key)
		}
	}
	return model, nil
}

func (model Model) View() string {
	if model.width == 0 {
		return ""
	}
	return renderLayout(model)
}
