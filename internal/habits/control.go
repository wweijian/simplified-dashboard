package habits

func (model Model) Update(key string) Model {
	if len(model.habits) == 0 {
		return model
	}

	switch key {
	case "up":
		return model.moveSelection(-1)
	case "down":
		return model.moveSelection(1)
	case "left":
		return model.moveDaySelection(1)
	case "right":
		return model.moveDaySelection(-1)
	case "t":
		model.selectedDayOffset = 0
		return model
	case " ":
		return model.cycleSelectedDayStatus()
	case "c":
		return model.setSelectedDayStatus(LogComplete)
	case "s":
		return model.setSelectedDayStatus(LogSkipped)
	case "x":
		return model.setSelectedDayStatus(LogIncomplete)
	case "d":
		return model.deleteSelectedHabit()
	}

	return model
}

func (model Model) moveSelection(distance int) Model {
	nextSelection := model.selected + distance
	if nextSelection >= 0 && nextSelection < len(model.habits) {
		model.selected = nextSelection
	}
	return model.clampSelectedDay()
}

func (model Model) selectedHistoryMaxOffset() int {
	if len(model.habits) == 0 || model.selected < 0 || model.selected >= len(model.habits) {
		return 0
	}
	if len(model.habits[model.selected].History) == 0 {
		return 0
	}
	return len(model.habits[model.selected].History) - 1
}

func (model Model) clampSelectedDay() Model {
	maxOffset := model.selectedHistoryMaxOffset()
	if model.selectedDayOffset > maxOffset {
		model.selectedDayOffset = maxOffset
	}
	if model.selectedDayOffset < 0 {
		model.selectedDayOffset = 0
	}
	return model
}

func (model Model) moveDaySelection(distance int) Model {
	nextSelection := model.selectedDayOffset + distance
	maxOffset := model.selectedHistoryMaxOffset()

	switch {
	case nextSelection < 0:
		model.selectedDayOffset = maxOffset
	case nextSelection > maxOffset:
		model.selectedDayOffset = 0
	default:
		model.selectedDayOffset = nextSelection
	}
	return model
}

func (model Model) cycleSelectedDayStatus() Model {
	return model.setSelectedDayStatus(nextLogStatus(model.selectedDayStatus()))
}

func (model Model) selectedDayStatus() HabitLogStatus {
	if len(model.habits) == 0 || model.selected < 0 || model.selected >= len(model.habits) {
		return LogIncomplete
	}
	if model.selectedDayOffset < 0 || model.selectedDayOffset >= len(model.habits[model.selected].History) {
		return LogIncomplete
	}

	status := model.habits[model.selected].History[model.selectedDayOffset].Status
	if status == "" {
		return LogIncomplete
	}
	return status
}

func (model Model) setSelectedDayStatus(status HabitLogStatus) Model {
	selectedHabit := model.habits[model.selected].Habit
	date := dateString(currentDate().AddDate(0, 0, -model.selectedDayOffset))
	if err := model.store.SetStatus(selectedHabit.ID, date, status); err != nil {
		model.err = err
		return model
	}

	model.load()
	return model.selectHabit(selectedHabit.ID)
}

func nextLogStatus(status HabitLogStatus) HabitLogStatus {
	switch status {
	case LogComplete:
		return LogSkipped
	case LogSkipped:
		return LogIncomplete
	default:
		return LogComplete
	}
}

func (model Model) archiveSelectedHabit() Model {
	selectedHabit := model.habits[model.selected].Habit
	if err := model.store.Archive(selectedHabit.ID); err != nil {
		model.err = err
		return model
	}

	model.load()
	if model.selected >= len(model.habits) && model.selected > 0 {
		model.selected--
	}
	return model
}

func (model Model) deleteSelectedHabit() Model {
	selectedHabit := model.habits[model.selected].Habit
	return model.DeleteHabit(selectedHabit.ID)
}
