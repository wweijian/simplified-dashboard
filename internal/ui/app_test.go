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
	`); err != nil {
		t.Fatalf("create tasks table: %v", err)
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

func TestAOpensAddTaskModeWithoutCreatingTask(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)
	model.activePanel = int(Finance)

	model = pressRune(model, "a")

	if model.mode != ModeAddTask {
		t.Fatalf("expected add-task mode, got %d", model.mode)
	}

	if taskCount(t, sqlDB) != 0 {
		t.Fatalf("expected no placeholder task, got %d", taskCount(t, sqlDB))
	}
}

func TestEscCancelsAddTaskModeWithoutCreatingTask(t *testing.T) {
	database, sqlDB := openTestDatabase(t)
	model := New(database)

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
	model.activePanel = int(Finance)
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
