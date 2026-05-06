package habits

import "fmt"

type HabitLogStatus string

const (
	LogComplete   HabitLogStatus = "complete"
	LogIncomplete HabitLogStatus = "incomplete"
	LogSkipped    HabitLogStatus = "skipped"
)

func (store Store) LogsSince(habitID int64, date string) (map[string]HabitLogStatus, error) {
	rows, err := store.db.Query(`
		SELECT date, completed, status
		FROM habit_logs
		WHERE habit_id = ? AND date >= ?
		ORDER BY date ASC
	`, habitID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query habit logs: %w", err)
	}
	defer rows.Close()

	logs := make(map[string]HabitLogStatus)
	for rows.Next() {
		date, status, err := scanHabitLogStatus(rows)
		if err != nil {
			return nil, err
		}
		logs[date] = status
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate habit logs: %w", err)
	}
	return logs, nil
}

type logStatusRow interface {
	Scan(dest ...any) error
}

func scanHabitLogStatus(row logStatusRow) (string, HabitLogStatus, error) {
	var date string
	var completed int
	var status string
	if err := row.Scan(&date, &completed, &status); err != nil {
		return "", "", fmt.Errorf("scan habit log: %w", err)
	}
	return date, normalizeLogStatus(status, completed == 1), nil
}

func (store Store) ToggleCompleted(habitID int64, date string) error {
	logs, err := store.LogsSince(habitID, date)
	if err != nil {
		return err
	}

	nextStatus := LogComplete
	if logs[date] == LogComplete {
		nextStatus = LogIncomplete
	}
	return store.SetStatus(habitID, date, nextStatus)
}

func (store Store) SetStatus(habitID int64, date string, status HabitLogStatus) error {
	status = normalizeLogStatus(string(status), false)
	completed := completedValue(status)

	_, err := store.db.Exec(`
		INSERT INTO habit_logs (habit_id, date, completed, status)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(habit_id, date)
		DO UPDATE SET completed = ?, status = ?
	`, habitID, date, completed, status, completed, status)
	if err != nil {
		return fmt.Errorf("failed to set habit log status: %w", err)
	}
	return nil
}

func completedValue(status HabitLogStatus) int {
	if status == LogComplete {
		return 1
	}
	return 0
}

func normalizeLogStatus(status string, completed bool) HabitLogStatus {
	switch HabitLogStatus(status) {
	case LogComplete, LogIncomplete, LogSkipped:
		return HabitLogStatus(status)
	default:
		if completed {
			return LogComplete
		}
		return LogIncomplete
	}
}
