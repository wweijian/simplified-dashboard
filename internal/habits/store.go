package habits

import (
	"database/sql"
	"fmt"

	"simplified-dashboard/internal/db"
)

type Habit struct {
	ID            int64
	Name          string
	Frequency     string
	FrequencyDays sql.NullString
	Archived      bool
	CreatedAt     string
}

type HabitLog struct {
	ID        int64
	HabitID   int64
	Date      string
	Completed bool
	Status    HabitLogStatus
}

type CreateHabitInput struct {
	Name          string
	Frequency     string
	FrequencyDays sql.NullString
}

type UpdateHabitInput = CreateHabitInput

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

func (store Store) ListActive() ([]Habit, error) {
	rows, err := store.db.Query(`
		SELECT id, name, frequency, frequency_days, archived, created_at
		FROM habits
		WHERE archived = 0
		ORDER BY created_at ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query habits: %w", err)
	}
	defer rows.Close()

	return scanHabits(rows)
}

func scanHabits(rows *sql.Rows) ([]Habit, error) {
	var habits []Habit
	for rows.Next() {
		habit, err := scanHabit(rows)
		if err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate habits: %w", err)
	}
	return habits, nil
}

func scanHabit(rows *sql.Rows) (Habit, error) {
	var habit Habit
	var archived int
	if err := rows.Scan(
		&habit.ID,
		&habit.Name,
		&habit.Frequency,
		&habit.FrequencyDays,
		&archived,
		&habit.CreatedAt,
	); err != nil {
		return Habit{}, fmt.Errorf("scan habit: %w", err)
	}

	habit.Archived = archived == 1
	return habit, nil
}

func (store Store) Create(input CreateHabitInput) error {
	_, err := store.db.Exec(`
		INSERT INTO habits (name, frequency, frequency_days, archived)
		VALUES (?, ?, ?, 0)
	`, input.Name, input.Frequency, input.FrequencyDays)
	if err != nil {
		return fmt.Errorf("failed to create habit: %w", err)
	}
	return nil
}

func (store Store) Archive(id int64) error {
	_, err := store.db.Exec(`
		UPDATE habits
		SET archived = 1
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("failed to archive habit: %w", err)
	}
	return nil
}

func (store Store) Update(id int64, input UpdateHabitInput) error {
	_, err := store.db.Exec(`
		UPDATE habits
		SET name = ?, frequency = ?, frequency_days = ?
		WHERE id = ?
	`, input.Name, input.Frequency, input.FrequencyDays, id)
	if err != nil {
		return fmt.Errorf("failed to update habit: %w", err)
	}
	return nil
}

func (store Store) Delete(id int64) error {
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("begin delete habit: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM habit_logs WHERE habit_id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete habit logs: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM habits WHERE id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete habit: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete habit: %w", err)
	}
	return nil
}
