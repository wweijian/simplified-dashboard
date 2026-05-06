package habits

import "time"

func New(store *Store) Model {
	model := Model{store: store}
	model.load()
	return model
}

func (model Model) Create(input CreateHabitInput) Model {
	if err := model.store.Create(input); err != nil {
		model.err = err
		return model
	}

	model.load()
	if len(model.habits) > 0 {
		model.selected = len(model.habits) - 1
	} else {
		model.selected = 0
	}
	return model
}

func (model Model) UpdateHabit(id int64, input UpdateHabitInput) Model {
	if err := model.store.Update(id, input); err != nil {
		model.err = err
		return model
	}

	model.load()
	return model.selectHabit(id)
}

func (model Model) DeleteHabit(id int64) Model {
	if err := model.store.Delete(id); err != nil {
		model.err = err
		return model
	}

	model.load()
	if model.selected >= len(model.habits) && model.selected > 0 {
		model.selected--
	}
	return model
}

func (model Model) SelectedHabit() (Habit, bool) {
	if len(model.habits) == 0 || model.selected < 0 || model.selected >= len(model.habits) {
		return Habit{}, false
	}
	return model.habits[model.selected].Habit, true
}

func (model Model) selectHabit(id int64) Model {
	for i, habit := range model.habits {
		if habit.Habit.ID == id {
			model.selected = i
			break
		}
	}
	return model
}

func (model *Model) load() {
	habits, err := model.store.ListActive()
	if err != nil {
		model.err = err
		return
	}

	today := currentDate()
	from := today.AddDate(0, 0, -365)
	statuses := make([]HabitStatus, 0, len(habits))
	for _, habit := range habits {
		logs, err := model.store.LogsSince(habit.ID, dateString(from))
		if err != nil {
			model.err = err
			return
		}

		statuses = append(statuses, buildHabitStatus(habit, logs, today))
	}

	model.habits = statuses
	if model.selected >= len(model.habits) {
		model.selected = len(model.habits) - 1
	}
	if model.selected < 0 {
		model.selected = 0
	}
	*model = model.clampSelectedDay()
	model.err = nil
}

func buildHabitStatus(habit Habit, logs map[string]HabitLogStatus, today time.Time) HabitStatus {
	todayKey := dateString(today)
	return HabitStatus{
		Habit:     habit,
		Logs:      logs,
		DueToday:  habitDueOn(habit, today),
		DoneToday: logs[todayKey] == LogComplete,
		Streak:    habitStreak(habit, logs, today),
		History:   habitHistory(habit, logs, today),
	}
}
