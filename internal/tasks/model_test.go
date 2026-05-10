package tasks

import (
	"database/sql"
	"testing"
)

func TestModelCreateTask(t *testing.T) {
	store := openTestStore(t)
	model := New(store)

	input := CreateTaskInput{
		Title:    "New task",
		Priority: 1,
	}

	model = model.Create(input)

	if len(model.tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(model.tasks))
	}

	if model.tasks[0].Title != "New task" {
		t.Fatalf("expected title %q, got %q", "New task", model.tasks[0].Title)
	}

	if model.selected != 0 {
		t.Fatalf("expected selected index 0, got %d", model.selected)
	}
}

func TestModelSelectionMovement(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "First")
	seedTask(t, store, "Second")

	model := New(store)

	model = model.Update("down")
	if model.selected != 1 {
		t.Fatalf("expected selected index 1, got %d", model.selected)
	}

	model = model.Update("down")
	if model.selected != 1 {
		t.Fatalf("expected selected index to stay at 1, got %d", model.selected)
	}

	model = model.Update("up")
	if model.selected != 0 {
		t.Fatalf("expected selected index 0, got %d", model.selected)
	}

	model = model.Update("up")
	if model.selected != 0 {
		t.Fatalf("expected selected index to stay at 0, got %d", model.selected)
	}
}

func TestModelToggleSelectedTask(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "Toggle me")

	model := New(store)
	model = model.Update(" ")

	if !model.tasks[0].Completed {
		t.Fatalf("expected selected task to be completed")
	}
}

func TestModelDeleteSelectedTask(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "First")
	seedTask(t, store, "Second")

	model := New(store)
	model = model.Update("down")
	model = model.Update("d")

	if len(model.tasks) != 1 {
		t.Fatalf("expected 1 task after delete, got %d", len(model.tasks))
	}

	if model.selected != 0 {
		t.Fatalf("expected selected index to move back to 0, got %d", model.selected)
	}
}

func TestModelCycleSelectedTaskPriority(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "Prioritize me")

	model := New(store)
	model = model.Update("p")

	if model.tasks[0].Priority != 2 {
		t.Fatalf("expected priority 2, got %d", model.tasks[0].Priority)
	}

	model = model.Update("p")

	if model.tasks[0].Priority != 0 {
		t.Fatalf("expected priority 0, got %d", model.tasks[0].Priority)
	}
}

func TestModelToggleSortMode(t *testing.T) {
	store := openTestStore(t)
	model := New(store)

	model = model.Update("s")

	if model.sortMode != SortPriority {
		t.Fatalf("expected priority sort mode, got %d", model.sortMode)
	}

	model = model.Update("s")

	if model.sortMode != SortDefault {
		t.Fatalf("expected default sort mode, got %d", model.sortMode)
	}
}

func TestModelSortsByPriority(t *testing.T) {
	store := openTestStore(t)

	lowPriority := CreateTaskInput{
		Title:    "Soon but low",
		DueDate:  sql.NullString{String: "2026-05-06", Valid: true},
		Priority: 0,
	}

	mustCreateTask(t, store, lowPriority)

	highPriority := CreateTaskInput{
		Title:    "Later but high",
		DueDate:  sql.NullString{String: "2026-05-20", Valid: true},
		Priority: 2,
	}

	mustCreateTask(t, store, highPriority)

	model := New(store)

	if model.tasks[0].Title != "Soon but low" {
		t.Fatalf("expected default sort to use due date first, got %q", model.tasks[0].Title)
	}

	model = model.Update("s")

	if model.tasks[0].Title != "Later but high" {
		t.Fatalf("expected priority sort to put high priority first, got %q", model.tasks[0].Title)
	}
}

func TestModelCyclesFilters(t *testing.T) {
	store := openTestStore(t)
	model := New(store)

	model = model.Update("]")
	if model.filterMode != FilterToday {
		t.Fatalf("expected today filter, got %d", model.filterMode)
	}

	model = model.Update("]")
	if model.filterMode != FilterOverdue {
		t.Fatalf("expected overdue filter, got %d", model.filterMode)
	}

	model = model.Update("]")
	if model.filterMode != FilterCompleted {
		t.Fatalf("expected completed filter, got %d", model.filterMode)
	}

	model = model.Update("]")
	if model.filterMode != FilterAll {
		t.Fatalf("expected all filter, got %d", model.filterMode)
	}

	model = model.Update("[")
	if model.filterMode != FilterCompleted {
		t.Fatalf("expected previous filter to wrap to completed, got %d", model.filterMode)
	}
}

func TestModelFiltersTodayTasks(t *testing.T) {
	withCurrentDate(t, "2026-05-06")
	store := openTestStore(t)

	mustCreateTask(t, store, CreateTaskInput{
		Title:    "Today",
		DueDate:  sql.NullString{String: "2026-05-06", Valid: true},
		Priority: 1,
	})
	mustCreateTask(t, store, CreateTaskInput{
		Title:    "Tomorrow",
		DueDate:  sql.NullString{String: "2026-05-07", Valid: true},
		Priority: 1,
	})

	model := New(store)
	model = model.Update("]")

	if len(model.tasks) != 1 || model.tasks[0].Title != "Today" {
		t.Fatalf("expected only today's task, got %#v", model.tasks)
	}
}

func TestModelFiltersOverdueTasks(t *testing.T) {
	withCurrentDate(t, "2026-05-06")
	store := openTestStore(t)

	mustCreateTask(t, store, CreateTaskInput{
		Title:    "Overdue",
		DueDate:  sql.NullString{String: "2026-05-05", Valid: true},
		Priority: 1,
	})
	mustCreateTask(t, store, CreateTaskInput{
		Title:    "Today",
		DueDate:  sql.NullString{String: "2026-05-06", Valid: true},
		Priority: 1,
	})

	model := New(store)
	model = model.Update("]")
	model = model.Update("]")

	if len(model.tasks) != 1 || model.tasks[0].Title != "Overdue" {
		t.Fatalf("expected only overdue task, got %#v", model.tasks)
	}
}

func TestModelFiltersCompletedTasks(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "Incomplete")
	seedTask(t, store, "Complete")

	tasks := mustListTasks(t, store)
	for _, task := range tasks {
		if task.Title == "Complete" {
			if err := store.ToggleCompleted(task.ID); err != nil {
				t.Fatalf("complete task: %v", err)
			}
			break
		}
	}

	model := New(store)
	model = model.Update("[")

	if len(model.tasks) != 1 || model.tasks[0].Title != "Complete" {
		t.Fatalf("expected only completed task, got %#v", model.tasks)
	}
}

func TestModelUpdateTaskPreservesSelection(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "First")
	seedTask(t, store, "Second")

	model := New(store)
	model = model.Update("down")
	selectedTask := model.selectedTask()

	model = model.UpdateTask(selectedTask.ID, UpdateTaskInput{
		Title:    "Second edited",
		Priority: 2,
	})

	if model.selectedTask().ID != selectedTask.ID {
		t.Fatalf("expected edited task to remain selected")
	}
	if model.selectedTask().Title != "Second edited" {
		t.Fatalf("expected edited task title, got %q", model.selectedTask().Title)
	}
}
