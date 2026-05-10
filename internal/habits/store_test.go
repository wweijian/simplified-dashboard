package habits

import (
	"database/sql"
	"testing"

	"simplified-dashboard/internal/db"

	_ "modernc.org/sqlite"
)

func openTestStore(t *testing.T) (*Store, *sql.DB) {
	t.Helper()

	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	t.Cleanup(func() {
		sqlDB.Close()
	})

	if _, err := sqlDB.Exec(`
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
	`); err != nil {
		t.Fatalf("create habits tables: %v", err)
	}

	return NewStore(&db.DB{DB: sqlDB}), sqlDB
}

func seedHabit(t *testing.T, store *Store, name string) {
	t.Helper()

	if err := store.Create(CreateHabitInput{
		Name:      name,
		Frequency: string(FrequencyDaily),
	}); err != nil {
		t.Fatalf("create habit: %v", err)
	}
}

func seedHabitCreatedAt(t *testing.T, store *Store, name string, createdAt string) {
	t.Helper()

	if _, err := store.db.Exec(`
		INSERT INTO habits (name, frequency, archived, created_at)
		VALUES (?, ?, 0, ?)
	`, name, string(FrequencyDaily), createdAt); err != nil {
		t.Fatalf("seed habit with created_at: %v", err)
	}
}

func TestCreateAndListActiveHabits(t *testing.T) {
	store, _ := openTestStore(t)

	seedHabit(t, store, "Exercise")

	habits, err := store.ListActive()
	if err != nil {
		t.Fatalf("list active habits: %v", err)
	}

	if len(habits) != 1 {
		t.Fatalf("expected 1 habit, got %d", len(habits))
	}
	if habits[0].Name != "Exercise" || habits[0].Frequency != string(FrequencyDaily) {
		t.Fatalf("unexpected habit: %#v", habits[0])
	}
}

func TestToggleCompletedCreatesAndFlipsTodayLog(t *testing.T) {
	store, _ := openTestStore(t)
	seedHabit(t, store, "Read")

	habits, err := store.ListActive()
	if err != nil {
		t.Fatalf("list habits: %v", err)
	}

	if err := store.ToggleCompleted(habits[0].ID, "2026-05-06"); err != nil {
		t.Fatalf("toggle complete: %v", err)
	}

	logs, err := store.LogsSince(habits[0].ID, "2026-05-01")
	if err != nil {
		t.Fatalf("list logs: %v", err)
	}
	if logs["2026-05-06"] != LogComplete {
		t.Fatalf("expected log to be completed after first toggle")
	}

	if err := store.ToggleCompleted(habits[0].ID, "2026-05-06"); err != nil {
		t.Fatalf("toggle incomplete: %v", err)
	}

	logs, err = store.LogsSince(habits[0].ID, "2026-05-01")
	if err != nil {
		t.Fatalf("list logs after second toggle: %v", err)
	}
	if logs["2026-05-06"] != LogIncomplete {
		t.Fatalf("expected log to be incomplete after second toggle")
	}
}

func TestSetStatusMarksSkipped(t *testing.T) {
	store, _ := openTestStore(t)
	seedHabit(t, store, "Rest")

	habits, err := store.ListActive()
	if err != nil {
		t.Fatalf("list habits: %v", err)
	}

	if err := store.SetStatus(habits[0].ID, "2026-05-06", LogSkipped); err != nil {
		t.Fatalf("set skipped: %v", err)
	}

	logs, err := store.LogsSince(habits[0].ID, "2026-05-01")
	if err != nil {
		t.Fatalf("list logs: %v", err)
	}
	if logs["2026-05-06"] != LogSkipped {
		t.Fatalf("expected skipped log, got %q", logs["2026-05-06"])
	}
}

func TestArchiveHidesHabitFromActiveList(t *testing.T) {
	store, _ := openTestStore(t)
	seedHabit(t, store, "Archive me")

	habits, err := store.ListActive()
	if err != nil {
		t.Fatalf("list habits: %v", err)
	}
	if err := store.Archive(habits[0].ID); err != nil {
		t.Fatalf("archive habit: %v", err)
	}

	habits, err = store.ListActive()
	if err != nil {
		t.Fatalf("list habits after archive: %v", err)
	}
	if len(habits) != 0 {
		t.Fatalf("expected archived habit to be hidden, got %d habits", len(habits))
	}
}

func TestUpdateHabit(t *testing.T) {
	store, _ := openTestStore(t)
	seedHabit(t, store, "Draft")

	habits, err := store.ListActive()
	if err != nil {
		t.Fatalf("list habits: %v", err)
	}

	input := UpdateHabitInput{
		Name:          "Final",
		Frequency:     string(FrequencyCustom),
		FrequencyDays: sql.NullString{String: "[1,3,5]", Valid: true},
	}
	if err := store.Update(habits[0].ID, input); err != nil {
		t.Fatalf("update habit: %v", err)
	}

	habits, err = store.ListActive()
	if err != nil {
		t.Fatalf("list habits after update: %v", err)
	}

	if habits[0].Name != "Final" || habits[0].Frequency != string(FrequencyCustom) || habits[0].FrequencyDays != input.FrequencyDays {
		t.Fatalf("unexpected updated habit: %#v", habits[0])
	}
}

func TestDeleteHabitRemovesHabitAndLogs(t *testing.T) {
	store, sqlDB := openTestStore(t)
	seedHabit(t, store, "Delete me")

	habits, err := store.ListActive()
	if err != nil {
		t.Fatalf("list habits: %v", err)
	}
	if err := store.ToggleCompleted(habits[0].ID, "2026-05-06"); err != nil {
		t.Fatalf("toggle habit: %v", err)
	}

	if err := store.Delete(habits[0].ID); err != nil {
		t.Fatalf("delete habit: %v", err)
	}

	var habitCount int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM habits`).Scan(&habitCount); err != nil {
		t.Fatalf("count habits: %v", err)
	}
	if habitCount != 0 {
		t.Fatalf("expected habit to be deleted, got %d habits", habitCount)
	}

	var logCount int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM habit_logs`).Scan(&logCount); err != nil {
		t.Fatalf("count habit logs: %v", err)
	}
	if logCount != 0 {
		t.Fatalf("expected habit logs to be deleted, got %d logs", logCount)
	}
}
