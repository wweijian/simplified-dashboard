package tasks

type SortMode int

const (
	SortDefault SortMode = iota
	SortPriority
)

type Model struct {
	store    *Store
	tasks    []Task
	selected int
	sortMode SortMode
	err      error
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
	model.selected = len(model.tasks) - 1
	return model
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
	model.tasks = tasks
	model.err = nil
}
