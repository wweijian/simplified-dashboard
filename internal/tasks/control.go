package tasks

func (model Model) Update(key string) Model {
	if key == "s" {
		return model.toggleSortMode()
	}

	if len(model.tasks) == 0 {
		return model
	}

	switch key {
	case "up":
		return model.moveSelection(-1)
	case "down":
		return model.moveSelection(1)
	case " ":
		return model.toggleSelectedTask()
	case "d":
		return model.deleteSelectedTask()
	case "p":
		return model.cycleSelectedTaskPriority()
	}

	return model
}

func (model Model) toggleSortMode() Model {
	if model.sortMode == SortDefault {
		model.sortMode = SortPriority
	} else {
		model.sortMode = SortDefault
	}

	model.load()
	return model
}

func (model Model) moveSelection(distance int) Model {
	nextSelection := model.selected + distance
	if nextSelection >= 0 && nextSelection < len(model.tasks) {
		model.selected = nextSelection
	}
	return model
}

func (model Model) toggleSelectedTask() Model {
	selectedTask := model.selectedTask()
	if err := model.store.ToggleCompleted(selectedTask.ID); err != nil {
		model.err = err
		return model
	}

	model.load()
	return model
}

func (model Model) deleteSelectedTask() Model {
	selectedTask := model.selectedTask()
	if err := model.store.Delete(selectedTask.ID); err != nil {
		model.err = err
		return model
	}

	model.load()
	if model.selected >= len(model.tasks) && model.selected > 0 {
		model.selected--
	}
	return model
}

func (model Model) cycleSelectedTaskPriority() Model {
	selectedTask := model.selectedTask()
	priority := nextPriority(selectedTask.Priority)

	if err := model.store.UpdatePriority(selectedTask.ID, priority); err != nil {
		model.err = err
		return model
	}

	model.load()
	return model.selectTask(selectedTask.ID)
}

func (model Model) selectTask(id int64) Model {
	for i, task := range model.tasks {
		if task.ID == id {
			model.selected = i
			break
		}
	}
	return model
}

func (model Model) selectedTask() Task {
	return model.tasks[model.selected]
}

func nextPriority(priority int) int {
	switch priority {
	case 0:
		return 1
	case 1:
		return 2
	default:
		return 0
	}
}
