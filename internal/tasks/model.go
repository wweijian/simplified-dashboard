package tasks

import "time"

type SortMode int

const (
	SortDefault SortMode = iota
	SortPriority
)

type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterToday
	FilterOverdue
	FilterCompleted
)

type Model struct {
	store      *Store
	allTasks   []Task
	tasks      []Task
	selected   int
	sortMode   SortMode
	filterMode FilterMode
	err        error
}

func New(store *Store) Model {
	model := Model{store: store}
	model.load()
	return model
}

func (model Model) Create(input CreateTaskInput) Model {
	if err := model.store.Create(input); err != nil {
		model.err = err
		return model
	}

	model.load()
	if len(model.tasks) > 0 {
		model.selected = len(model.tasks) - 1
	} else {
		model.selected = 0
	}
	return model
}

func (model Model) UpdateTask(id int64, input UpdateTaskInput) Model {
	if err := model.store.Update(id, input); err != nil {
		model.err = err
		return model
	}

	model.load()
	return model.selectTask(id)
}

func (model Model) SelectedTask() (Task, bool) {
	if len(model.tasks) == 0 || model.selected < 0 || model.selected >= len(model.tasks) {
		return Task{}, false
	}
	return model.tasks[model.selected], true
}

func (model Model) FilterLabel() string {
	switch model.filterMode {
	case FilterToday:
		return "today"
	case FilterOverdue:
		return "overdue"
	case FilterCompleted:
		return "completed"
	default:
		return "all"
	}
}

func (model Model) SortLabel() string {
	if model.sortMode == SortPriority {
		return "priority"
	}
	return "due date"
}

func (model *Model) load() {
	var tasks []Task
	var err error
	if model.sortMode == SortPriority {
		tasks, err = model.store.ListByPriority()
	} else {
		tasks, err = model.store.List()
	}

	if err != nil {
		model.err = err
		return
	}
	model.allTasks = tasks
	model.applyFilter()
	model.err = nil
}

func (model *Model) applyFilter() {
	filtered := make([]Task, 0, len(model.allTasks))
	for _, task := range model.allTasks {
		if model.includesTask(task) {
			filtered = append(filtered, task)
		}
	}

	model.tasks = filtered
	if model.selected >= len(model.tasks) {
		model.selected = len(model.tasks) - 1
	}
	if model.selected < 0 {
		model.selected = 0
	}
}

func (model Model) includesTask(task Task) bool {
	days, hasDueDate := taskDaysUntilDue(task)

	switch model.filterMode {
	case FilterToday:
		return !task.Completed && hasDueDate && days == 0
	case FilterOverdue:
		return !task.Completed && hasDueDate && days < 0
	case FilterCompleted:
		return task.Completed
	default:
		return true
	}
}

func taskDaysUntilDue(task Task) (int, bool) {
	if !task.DueDate.Valid || task.DueDate.String == "" {
		return 0, false
	}

	dueDate, err := time.ParseInLocation("2006-01-02", task.DueDate.String, time.Local)
	if err != nil {
		return 0, false
	}

	return daysUntilDue(dueDate), true
}
