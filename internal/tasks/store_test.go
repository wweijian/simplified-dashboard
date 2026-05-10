package tasks

import (
	"database/sql"
	"testing"

	"simplified-dashboard/internal/db"

	_ "modernc.org/sqlite"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()

	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	t.Cleanup(func() {
		sqlDB.Close()
	})

	if _, err := sqlDB.Exec(`
		CREATE TABLE tasks (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			title       TEXT NOT NULL,
			description TEXT,
			due_date    TEXT,
			priority    INTEGER DEFAULT 1,
			completed   INTEGER DEFAULT 0,
			created_at  TEXT DEFAULT (datetime('now'))
		);
	`); err != nil {
		t.Fatalf("create tasks table: %v", err)
	}

	return NewStore(&db.DB{DB: sqlDB})
}

func seedTask(t *testing.T, store *Store, title string) {
	t.Helper()

	input := CreateTaskInput{
		Title:    title,
		Priority: 1,
	}

	mustCreateTask(t, store, input)
}

func mustCreateTask(t *testing.T, store *Store, input CreateTaskInput) {
	t.Helper()

	if err := store.Create(input); err != nil {
		t.Fatalf("create task: %v", err)
	}
}

func mustListTasks(t *testing.T, store *Store) []Task {
	t.Helper()

	tasks, err := store.List()
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	return tasks
}

func TestCreateAndListTasks(t *testing.T) {
	store := openTestStore(t)

	input := CreateTaskInput{
		Title:       "Write tests",
		Description: sql.NullString{String: "Describe the work", Valid: true},
		DueDate:     sql.NullString{String: "2026-05-12", Valid: true},
		Priority:    2,
	}

	mustCreateTask(t, store, input)
	tasks := mustListTasks(t, store)

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	assertTaskMatchesInput(t, tasks[0], input)
}

func assertTaskMatchesInput(t *testing.T, task Task, input CreateTaskInput) {
	t.Helper()

	if task.Title != input.Title {
		t.Fatalf("expected title %q, got %q", input.Title, task.Title)
	}

	if task.Completed {
		t.Fatalf("new task should not be completed")
	}

	if task.Description != input.Description {
		t.Fatalf("expected description %#v, got %#v", input.Description, task.Description)
	}

	if task.DueDate != input.DueDate {
		t.Fatalf("expected due date %#v, got %#v", input.DueDate, task.DueDate)
	}

	if task.Priority != input.Priority {
		t.Fatalf("expected priority %d, got %d", input.Priority, task.Priority)
	}
}

func TestToggleCompleted(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "Toggle me")

	tasks := mustListTasks(t, store)

	if err := store.ToggleCompleted(tasks[0].ID); err != nil {
		t.Fatalf("toggle task: %v", err)
	}

	tasks = mustListTasks(t, store)

	if !tasks[0].Completed {
		t.Fatalf("task should be completed after toggle")
	}
}

func TestDeleteTask(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "Delete me")

	tasks := mustListTasks(t, store)

	if err := store.Delete(tasks[0].ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	tasks = mustListTasks(t, store)

	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks after delete, got %d", len(tasks))
	}
}

func TestUpdatePriority(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "Prioritize me")

	tasks := mustListTasks(t, store)

	if err := store.UpdatePriority(tasks[0].ID, 2); err != nil {
		t.Fatalf("update priority: %v", err)
	}

	tasks = mustListTasks(t, store)

	if tasks[0].Priority != 2 {
		t.Fatalf("expected priority 2, got %d", tasks[0].Priority)
	}
}

func TestUpdateTask(t *testing.T) {
	store := openTestStore(t)

	seedTask(t, store, "Draft")
	tasks := mustListTasks(t, store)

	input := UpdateTaskInput{
		Title:       "Final",
		Description: sql.NullString{String: "Updated notes", Valid: true},
		DueDate:     sql.NullString{String: "2026-05-20", Valid: true},
		Priority:    2,
	}

	if err := store.Update(tasks[0].ID, input); err != nil {
		t.Fatalf("update task: %v", err)
	}

	tasks = mustListTasks(t, store)
	assertTaskMatchesInput(t, tasks[0], input)
}

func TestListByPriority(t *testing.T) {
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

	tasks, err := store.ListByPriority()
	if err != nil {
		t.Fatalf("list tasks by priority: %v", err)
	}

	if tasks[0].Title != "Later but high" {
		t.Fatalf("expected high priority task first, got %q", tasks[0].Title)
	}
}
