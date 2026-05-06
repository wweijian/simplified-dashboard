package tasks

import (
	"database/sql"
	"fmt"

	"simplified-dashboard/internal/db"
)

type Task struct {
	ID          int64
	Title       string
	Description sql.NullString
	DueDate     sql.NullString
	Priority    int
	Completed   bool
	CreatedAt   string
}

type CreateTaskInput struct {
	Title       string
	Description sql.NullString
	DueDate     sql.NullString
	Priority    int
}

type UpdateTaskInput = CreateTaskInput

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

func (store Store) List() ([]Task, error) {
	rows, err := store.db.Query(
		`SELECT id, title, description, due_date, priority, completed, created_at 
		FROM tasks
		ORDER BY completed ASC, due_date IS NULL ASC, due_date ASC, priority DESC, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (store Store) ListByPriority() ([]Task, error) {
	rows, err := store.db.Query(
		`SELECT id, title, description, due_date, priority, completed, created_at
		FROM tasks
		ORDER BY completed ASC, priority DESC, due_date IS NULL ASC, due_date ASC, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by priority: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

func scanTasks(rows *sql.Rows) ([]Task, error) {
	var tasks []Task
	for rows.Next() {
		var task Task
		var completed int

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.Priority,
			&completed,
			&task.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		task.Completed = completed == 1
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (store Store) Create(input CreateTaskInput) error {
	_, err := store.db.Exec(`
		INSERT INTO tasks (title, description, due_date, priority, completed)
		VALUES (?, ?, ?, ?, 0)
	`, input.Title, input.Description, input.DueDate, input.Priority)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (store Store) Delete(id int64) error {
	_, err := store.db.Exec(`
		DELETE FROM tasks
		WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (store Store) ToggleCompleted(id int64) error {
	_, err := store.db.Exec(`
		UPDATE tasks
		SET completed = CASE completed WHEN 1 THEN 0 ELSE 1 END
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("failed to toggle task completion: %w", err)
	}
	return nil
}

func (store Store) UpdatePriority(id int64, priority int) error {
	_, err := store.db.Exec(`
		UPDATE tasks
		SET priority = ?
		WHERE id = ?
	`, priority, id)
	if err != nil {
		return fmt.Errorf("failed to update task priority: %w", err)
	}
	return nil
}

func (store Store) Update(id int64, input UpdateTaskInput) error {
	_, err := store.db.Exec(`
		UPDATE tasks
		SET title = ?, description = ?, due_date = ?, priority = ?
		WHERE id = ?
	`, input.Title, input.Description, input.DueDate, input.Priority, id)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}
