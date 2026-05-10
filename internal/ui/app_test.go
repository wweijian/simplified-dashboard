package ui

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"simplified-dashboard/internal/db"

	bbt "github.com/charmbracelet/bubbletea"
	_ "modernc.org/sqlite"
)

func openTestDatabase(t *testing.T) (*db.DB, *sql.DB) {
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

		CREATE TABLE habits (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			name           TEXT NOT NULL,
			frequency      TEXT NOT NULL,
			frequency_days TEXT,
			archived       INTEGER DEFAULT 0,
			created_at     TEXT DEFAULT (datetime('now'))
		);

		CREATE TABLE habit_logs (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			habit_id   INTEGER NOT NULL REFERENCES habits(id),
			date       TEXT NOT NULL,
			completed  INTEGER DEFAULT 0,
			status     TEXT NOT NULL DEFAULT 'incomplete',
			UNIQUE(habit_id, date)
		);

		CREATE TABLE finance_categories (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL
		);

		CREATE TABLE finance_transactions (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			date        TEXT NOT NULL,
			amount      REAL NOT NULL,
			category_id INTEGER REFERENCES finance_categories(id),
			description TEXT,
			created_at  TEXT DEFAULT (datetime('now'))
		);

	`); err != nil {
		t.Fatalf("create test tables: %v", err)
	}

	return &db.DB{DB: sqlDB}, sqlDB
}

func press(model Model, msg bbt.KeyMsg) Model {
	updated, _ := model.Update(msg)
	return updated.(Model)
}

func pressRune(model Model, value string) Model {
	return press(model, bbt.KeyMsg{
		Type:  bbt.KeyRunes,
		Runes: []rune(value),
	})
}

func pressType(model Model, keyType bbt.KeyType) Model {
	return press(model, bbt.KeyMsg{Type: keyType})
}

func resize(model Model, width, height int) Model {
	updated, _ := model.Update(bbt.WindowSizeMsg{Width: width, Height: height})
	return updated.(Model)
}

func taskCount(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()

	var count int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count); err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	return count
}

func habitCount(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()

	var count int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM habits`).Scan(&count); err != nil {
		t.Fatalf("count habits: %v", err)
	}
	return count
}

func TestAOpensAddTaskModeWithoutCreatingTask(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)

	model = pressRune(model, "a")

	if model.mode != ModeAddTask {
		t.Fatalf("expected add-task mode, got %d", model.mode)
	}

	if taskCount(t, sqlDB) != 0 {
		t.Fatalf("expected no placeholder task, got %d", taskCount(t, sqlDB))
	}
}

func TestAIgnoredOutsideTaskPanel(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Finance)

	model = pressRune(model, "a")

	if model.mode != ModeNormal {
		t.Fatalf("expected normal mode outside task panel, got %d", model.mode)
	}

	if taskCount(t, sqlDB) != 0 {
		t.Fatalf("expected no placeholder task, got %d", taskCount(t, sqlDB))
	}
}

func TestEscCancelsAddTaskModeWithoutCreatingTask(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)

	model = pressRune(model, "a")
	model = pressType(model, bbt.KeyEsc)

	if model.mode != ModeNormal {
		t.Fatalf("expected normal mode, got %d", model.mode)
	}

	if taskCount(t, sqlDB) != 0 {
		t.Fatalf("expected no task after cancel, got %d", taskCount(t, sqlDB))
	}
}

func TestAddTaskModeRendersPopup(t *testing.T) {
	database, _ := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)
	model = resize(model, 80, 24)

	model = pressRune(model, "a")

	got := model.View()
	if !strings.Contains(got, "Add task") || !strings.Contains(got, "Due in days") {
		t.Fatalf("expected add-task popup, got:\n%s", got)
	}
	if !strings.Contains(got, "Description") {
		t.Fatalf("expected add-task popup to include description, got:\n%s", got)
	}
}

func TestAddTaskPopupClearsDashboardTextBehindModalRows(t *testing.T) {
	database, _ := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)
	model = resize(model, 100, 24)

	model = pressRune(model, "a")

	lines := strings.Split(model.View(), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Due in days") && strings.Contains(line, "[3]") {
			t.Fatalf("expected popup row to clear dashboard text, got:\n%s", line)
		}
	}
}

func TestValidAddTaskSubmitCreatesTask(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)

	model = pressRune(model, "a")
	model = pressRune(model, "Pay bill")
	model = pressType(model, bbt.KeyTab)
	model = pressRune(model, "7")
	model = pressType(model, bbt.KeyTab)
	model = press(model, bbt.KeyMsg{Type: bbt.KeyRight})
	model = pressType(model, bbt.KeyTab)
	model = pressRune(model, "Monthly utilities")
	model = pressType(model, bbt.KeyEnter)

	if model.mode != ModeNormal {
		t.Fatalf("expected normal mode after submit, got %d", model.mode)
	}

	var title string
	var description string
	var dueDate string
	var priority int
	if err := sqlDB.QueryRow(`SELECT title, description, due_date, priority FROM tasks`).Scan(&title, &description, &dueDate, &priority); err != nil {
		t.Fatalf("query created task: %v", err)
	}

	wantDueDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	if title != "Pay bill" || description != "Monthly utilities" || dueDate != wantDueDate || priority != 2 {
		t.Fatalf("created task = (%q, %q, %q, %d), want (%q, %q, %q, %d)",
			title, description, dueDate, priority, "Pay bill", "Monthly utilities", wantDueDate, 2)
	}
}

func TestEmptyTitleDoesNotSubmit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)

	model = pressRune(model, "a")
	model = pressType(model, bbt.KeyEnter)

	if model.mode != ModeAddTask {
		t.Fatalf("expected add-task mode after invalid submit, got %d", model.mode)
	}

	if taskCount(t, sqlDB) != 0 {
		t.Fatalf("expected no task after invalid submit, got %d", taskCount(t, sqlDB))
	}
}

func TestDueDaysRejectsNonNumericInput(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)

	model = pressRune(model, "a")
	model = pressRune(model, "Pay bill")
	model = pressType(model, bbt.KeyTab)
	model = pressRune(model, "x")
	model = pressType(model, bbt.KeyEnter)

	if model.mode != ModeAddTask {
		t.Fatalf("expected add-task mode after invalid due input, got %d", model.mode)
	}

	if model.addTaskForm.errorText == "" {
		t.Fatalf("expected validation error for non-numeric due input")
	}

	if taskCount(t, sqlDB) != 0 {
		t.Fatalf("expected no task after invalid due input, got %d", taskCount(t, sqlDB))
	}
}

func TestLongTitleDoesNotSubmit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)

	model = pressRune(model, "a")
	model.addTaskForm.title = strings.Repeat("a", maxTaskTitleLength+1)
	model.addTaskForm.dueDays = "7"
	model = pressType(model, bbt.KeyEnter)

	if model.mode != ModeAddTask {
		t.Fatalf("expected add-task mode after long title, got %d", model.mode)
	}

	if taskCount(t, sqlDB) != 0 {
		t.Fatalf("expected no task after long title, got %d", taskCount(t, sqlDB))
	}
}

func TestDueDateBeyondOneYearDoesNotSubmit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Tasks)

	model = pressRune(model, "a")
	model = pressRune(model, "Pay bill")
	model = pressType(model, bbt.KeyTab)
	model = pressRune(model, "366")
	model = pressType(model, bbt.KeyEnter)

	if model.mode != ModeAddTask {
		t.Fatalf("expected add-task mode after distant due date, got %d", model.mode)
	}

	if taskCount(t, sqlDB) != 0 {
		t.Fatalf("expected no task after distant due date, got %d", taskCount(t, sqlDB))
	}
}

func TestTaskPanelKeysDoNotRunWhileAddTaskModeIsOpen(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	if _, err := sqlDB.Exec(`INSERT INTO tasks (title, priority, completed) VALUES ('Keep me', 1, 0)`); err != nil {
		t.Fatalf("seed task: %v", err)
	}

	model := New(database)
	model.activePanel = int(Tasks)

	model = pressRune(model, "a")
	model = pressRune(model, "d")

	if taskCount(t, sqlDB) != 1 {
		t.Fatalf("expected task key to be blocked by modal, got %d tasks", taskCount(t, sqlDB))
	}
}

func TestEOpensEditTaskMode(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	if _, err := sqlDB.Exec(`INSERT INTO tasks (title, due_date, priority, completed) VALUES ('Edit me', '2026-05-06', 1, 0)`); err != nil {
		t.Fatalf("seed task: %v", err)
	}

	model := New(database)
	model.activePanel = int(Tasks)
	model = resize(model, 80, 24)

	model = pressRune(model, "e")

	if model.mode != ModeEditTask {
		t.Fatalf("expected edit-task mode, got %d", model.mode)
	}
	if model.editingTaskID == 0 {
		t.Fatalf("expected editing task id to be set")
	}
	if !strings.Contains(model.View(), "Edit task") {
		t.Fatalf("expected edit-task popup, got:\n%s", model.View())
	}
}

func TestEditTaskSubmitUpdatesTask(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	if _, err := sqlDB.Exec(`INSERT INTO tasks (title, due_date, priority, completed) VALUES ('Old title', '2026-05-06', 1, 0)`); err != nil {
		t.Fatalf("seed task: %v", err)
	}

	model := New(database)
	model.activePanel = int(Tasks)
	model = pressRune(model, "e")
	model.addTaskForm.title = "New title"
	model.addTaskForm.description = "Updated description"
	model.addTaskForm.dueDays = "5"
	model.addTaskForm.priority = 2
	model = pressType(model, bbt.KeyEnter)

	if model.mode != ModeNormal {
		t.Fatalf("expected normal mode after edit submit, got %d", model.mode)
	}

	var title string
	var description string
	var dueDate string
	var priority int
	if err := sqlDB.QueryRow(`SELECT title, description, due_date, priority FROM tasks`).Scan(&title, &description, &dueDate, &priority); err != nil {
		t.Fatalf("query updated task: %v", err)
	}

	wantDueDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
	if title != "New title" || description != "Updated description" || dueDate != wantDueDate || priority != 2 {
		t.Fatalf("updated task = (%q, %q, %q, %d), want (%q, %q, %q, %d)",
			title, description, dueDate, priority, "New title", "Updated description", wantDueDate, 2)
	}
}

func TestEscCancelsEditTaskModeWithoutUpdatingTask(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	if _, err := sqlDB.Exec(`INSERT INTO tasks (title, due_date, priority, completed) VALUES ('Keep title', '2026-05-06', 1, 0)`); err != nil {
		t.Fatalf("seed task: %v", err)
	}

	model := New(database)
	model.activePanel = int(Tasks)
	model = pressRune(model, "e")
	model.addTaskForm.title = "Discard me"
	model = pressType(model, bbt.KeyEsc)

	if model.mode != ModeNormal {
		t.Fatalf("expected normal mode after edit cancel, got %d", model.mode)
	}

	var title string
	if err := sqlDB.QueryRow(`SELECT title FROM tasks`).Scan(&title); err != nil {
		t.Fatalf("query task title: %v", err)
	}
	if title != "Keep title" {
		t.Fatalf("expected title to be unchanged, got %q", title)
	}
}

func TestHabitPanelSpaceTogglesSelectedHabit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	if _, err := sqlDB.Exec(`INSERT INTO habits (name, frequency, archived) VALUES ('Read', 'daily', 0)`); err != nil {
		t.Fatalf("seed habit: %v", err)
	}

	model := New(database)
	model.activePanel = int(Habits)

	model = press(model, bbt.KeyMsg{Type: bbt.KeySpace})

	var completed int
	if err := sqlDB.QueryRow(`SELECT completed FROM habit_logs WHERE habit_id = 1`).Scan(&completed); err != nil {
		t.Fatalf("query habit log: %v", err)
	}
	if completed != 1 {
		t.Fatalf("expected habit to be completed after space, got %d", completed)
	}
}

func TestAOpensAddHabitModeWithoutCreatingHabit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Habits)

	model = pressRune(model, "a")

	if model.mode != ModeAddHabit {
		t.Fatalf("expected add-habit mode, got %d", model.mode)
	}

	if habitCount(t, sqlDB) != 0 {
		t.Fatalf("expected no placeholder habit, got %d", habitCount(t, sqlDB))
	}
}

func TestAddHabitModeRendersPopup(t *testing.T) {
	database, _ := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Habits)
	model = resize(model, 80, 24)

	model = pressRune(model, "a")

	got := model.View()
	if !strings.Contains(got, "Add habit") || !strings.Contains(got, "Days") {
		t.Fatalf("expected add-habit popup, got:\n%s", got)
	}
}

func TestValidAddHabitSubmitCreatesHabit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Habits)

	model = pressRune(model, "a")
	model = pressRune(model, "Exercise")
	model = pressType(model, bbt.KeyEnter)

	if model.mode != ModeNormal {
		t.Fatalf("expected normal mode after habit submit, got %d", model.mode)
	}

	var name string
	var frequency string
	var frequencyDays sql.NullString
	if err := sqlDB.QueryRow(`SELECT name, frequency, frequency_days FROM habits`).Scan(&name, &frequency, &frequencyDays); err != nil {
		t.Fatalf("query created habit: %v", err)
	}

	if name != "Exercise" || frequency != "daily" || frequencyDays.Valid {
		t.Fatalf("created habit = (%q, %q, %#v), want daily Exercise with no weekdays", name, frequency, frequencyDays)
	}
}

func TestCustomAddHabitSubmitCreatesHabitWithWeekdays(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Habits)

	model = pressRune(model, "a")
	model.addHabitForm.name = "Run"
	model.addHabitForm.weekdays = [7]bool{}
	model.addHabitForm.weekdays[1] = true
	model.addHabitForm.weekdays[3] = true
	model.addHabitForm.weekdays[5] = true
	model = pressType(model, bbt.KeyEnter)

	var frequencyDays string
	if err := sqlDB.QueryRow(`SELECT frequency_days FROM habits`).Scan(&frequencyDays); err != nil {
		t.Fatalf("query created custom habit: %v", err)
	}

	if frequencyDays != "[1,3,5]" {
		t.Fatalf("expected weekdays [1,3,5], got %q", frequencyDays)
	}
}

func TestEOpensEditHabitMode(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	if _, err := sqlDB.Exec(`INSERT INTO habits (name, frequency, archived) VALUES ('Read', 'daily', 0)`); err != nil {
		t.Fatalf("seed habit: %v", err)
	}

	model := New(database)
	model.activePanel = int(Habits)
	model = resize(model, 80, 24)

	model = pressRune(model, "e")

	if model.mode != ModeEditHabit {
		t.Fatalf("expected edit-habit mode, got %d", model.mode)
	}
	if model.editingHabitID == 0 {
		t.Fatalf("expected editing habit id to be set")
	}
	if !strings.Contains(model.View(), "Edit habit") {
		t.Fatalf("expected edit-habit popup, got:\n%s", model.View())
	}
}

func TestEditHabitSubmitUpdatesHabit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	if _, err := sqlDB.Exec(`INSERT INTO habits (name, frequency, archived) VALUES ('Old habit', 'daily', 0)`); err != nil {
		t.Fatalf("seed habit: %v", err)
	}

	model := New(database)
	model.activePanel = int(Habits)
	model = pressRune(model, "e")
	model.addHabitForm.name = "New habit"
	model.addHabitForm.weekdays = [7]bool{}
	model.addHabitForm.weekdays[2] = true
	model.addHabitForm.weekdays[4] = true
	model = pressType(model, bbt.KeyEnter)

	if model.mode != ModeNormal {
		t.Fatalf("expected normal mode after edit submit, got %d", model.mode)
	}

	var name string
	var frequency string
	var frequencyDays string
	if err := sqlDB.QueryRow(`SELECT name, frequency, frequency_days FROM habits`).Scan(&name, &frequency, &frequencyDays); err != nil {
		t.Fatalf("query updated habit: %v", err)
	}

	if name != "New habit" || frequency != "custom" || frequencyDays != "[2,4]" {
		t.Fatalf("updated habit = (%q, %q, %q), want (%q, %q, %q)", name, frequency, frequencyDays, "New habit", "custom", "[2,4]")
	}
}

func TestEscCancelsAddHabitModeWithoutCreatingHabit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Habits)

	model = pressRune(model, "a")
	model = pressType(model, bbt.KeyEsc)

	if model.mode != ModeNormal {
		t.Fatalf("expected normal mode after cancel, got %d", model.mode)
	}
	if habitCount(t, sqlDB) != 0 {
		t.Fatalf("expected no habit after cancel, got %d", habitCount(t, sqlDB))
	}
}

func TestHabitPanelDDeletesSelectedHabit(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	if _, err := sqlDB.Exec(`INSERT INTO habits (name, frequency, archived) VALUES ('Delete me', 'daily', 0)`); err != nil {
		t.Fatalf("seed habit: %v", err)
	}

	model := New(database)
	model.activePanel = int(Habits)

	model = pressRune(model, "d")

	var count int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM habits`).Scan(&count); err != nil {
		t.Fatalf("count habits after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected habit to be deleted, got %d habits", count)
	}
}
