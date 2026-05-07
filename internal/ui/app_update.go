package ui

import (
	"time"

	bbt "github.com/charmbracelet/bubbletea"
)

func (model Model) Update(msg bbt.Msg) (bbt.Model, bbt.Cmd) {
	switch msg := msg.(type) {
	case bbt.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
	case bbt.KeyMsg:
		return model.updateKey(msg)
	}
	return model, nil
}

func (model Model) updateKey(msg bbt.KeyMsg) (Model, bbt.Cmd) {
	key := msg.String()
	if model.isTaskFormMode() {
		return model.updateTaskFormMode(msg), nil
	}
	if model.isHabitFormMode() {
		return model.updateHabitFormMode(msg), nil
	}
	if key == "a" {
		return model.openAddMode(), nil
	}
	if key == "e" {
		return model.openEditMode(), nil
	}
	if handler, ok := keyBindings[key]; ok {
		return handler(model)
	}
	return model.updateActivePanel(key), nil
}

func (model Model) isTaskFormMode() bool {
	return model.mode == ModeAddTask || model.mode == ModeEditTask
}

func (model Model) isHabitFormMode() bool {
	return model.mode == ModeAddHabit || model.mode == ModeEditHabit
}

func (model Model) updateTaskFormMode(msg bbt.KeyMsg) Model {
	switch msg.String() {
	case "esc":
		return model.closeTaskForm()
	case "enter":
		return model.submitTaskForm()
	default:
		model.addTaskForm = model.addTaskForm.Update(msg)
		return model
	}
}

func (model Model) updateHabitFormMode(msg bbt.KeyMsg) Model {
	switch msg.String() {
	case "esc":
		return model.closeHabitForm()
	case "enter":
		return model.submitHabitForm()
	default:
		model.addHabitForm = model.addHabitForm.Update(msg)
		return model
	}
}

func (model Model) submitTaskForm() Model {
	form, input, ok := model.addTaskForm.Submit(time.Now())
	model.addTaskForm = form
	if !ok {
		return model
	}

	if model.mode == ModeEditTask {
		model.tasks = model.tasks.UpdateTask(model.editingTaskID, input)
	} else {
		model.tasks = model.tasks.Create(input)
	}
	return model.closeTaskForm()
}

func (model Model) submitHabitForm() Model {
	form, input, ok := model.addHabitForm.Submit()
	model.addHabitForm = form
	if !ok {
		return model
	}

	if model.mode == ModeEditHabit {
		model.habits = model.habits.UpdateHabit(model.editingHabitID, input)
	} else {
		model.habits = model.habits.Create(input)
	}
	return model.closeHabitForm()
}

func (model Model) closeTaskForm() Model {
	model.mode = ModeNormal
	model.addTaskForm = newAddTaskForm()
	model.editingTaskID = 0
	return model
}

func (model Model) closeHabitForm() Model {
	model.mode = ModeNormal
	model.addHabitForm = newAddHabitForm()
	model.editingHabitID = 0
	return model
}

func (model Model) openAddMode() Model {
	switch model.activePanel {
	case int(Tasks):
		model.mode = ModeAddTask
		model.addTaskForm = newAddTaskForm()
	case int(Habits):
		model.mode = ModeAddHabit
		model.addHabitForm = newAddHabitForm()
	}
	return model
}

func (model Model) openEditMode() Model {
	switch model.activePanel {
	case int(Tasks):
		return model.openTaskEditMode()
	case int(Habits):
		return model.openHabitEditMode()
	default:
		return model
	}
}

func (model Model) openTaskEditMode() Model {
	task, ok := model.tasks.SelectedTask()
	if !ok {
		return model
	}

	model.mode = ModeEditTask
	model.editingTaskID = task.ID
	model.addTaskForm = newEditTaskForm(task, time.Now())
	return model
}

func (model Model) openHabitEditMode() Model {
	habit, ok := model.habits.SelectedHabit()
	if !ok {
		return model
	}

	model.mode = ModeEditHabit
	model.editingHabitID = habit.ID
	model.addHabitForm = newEditHabitForm(habit)
	return model
}

func (model Model) updateActivePanel(key string) Model {
	if model.activePanel == int(Tasks) {
		model.tasks = model.tasks.Update(key)
	}
	if model.activePanel == int(Finance) {
		model.finance = model.finance.Update(key)
	}
	if model.activePanel == int(Habits) {
		model.habits = model.habits.Update(key)
	}
	return model
}
