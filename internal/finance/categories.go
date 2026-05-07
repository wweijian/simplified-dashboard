package finance

import (
	"database/sql"
	"fmt"
)

var defaultCategories = []Category{
	{Name: "Income", Type: "income"},
	{Name: "Food", Type: "expense"},
	{Name: "Transport", Type: "expense"},
	{Name: "Groceries", Type: "expense"},
	{Name: "Bills", Type: "expense"},
	{Name: "Subscriptions", Type: "expense"},
	{Name: "Health", Type: "expense"},
	{Name: "Shopping", Type: "expense"},
	{Name: "Travel", Type: "expense"},
	{Name: "Uncategorized", Type: "expense"},
}

func (store Store) EnsureDefaultCategories() error {
	for _, category := range defaultCategories {
		if err := store.CreateCategory(category.Name, category.Type); err != nil {
			return err
		}
	}
	return nil
}

func (store Store) CreateCategory(name, categoryType string) error {
	_, err := store.db.Exec(`
		INSERT OR IGNORE INTO finance_categories (name, type)
		VALUES (?, ?)
	`, name, categoryType)
	if err != nil {
		return fmt.Errorf("failed to create finance category: %w", err)
	}
	return nil
}

func (store Store) ListCategories() ([]Category, error) {
	rows, err := store.db.Query(`
		SELECT id, name, type
		FROM finance_categories
		ORDER BY type ASC, name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query finance categories: %w", err)
	}
	defer rows.Close()

	return scanCategories(rows)
}

func (store Store) ListExpenseCategories() ([]Category, error) {
	rows, err := store.db.Query(`
		SELECT id, name, type
		FROM finance_categories
		WHERE type = 'expense'
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query expense categories: %w", err)
	}
	defer rows.Close()

	return scanCategories(rows)
}

func scanCategories(rows *sql.Rows) ([]Category, error) {
	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Type); err != nil {
			return nil, fmt.Errorf("scan finance category: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate finance categories: %w", err)
	}
	return categories, nil
}
