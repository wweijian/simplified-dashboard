package finance

import (
	"database/sql"
	"fmt"
	"time"
)

func (store Store) ListRecentTransactions(limit int) ([]Transaction, error) {
	rows, err := store.db.Query(recentTransactionsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query finance transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

func (store Store) TransactionsForMonth(month time.Time, category CategoryFilter, limit int) ([]Transaction, error) {
	start, end := monthBounds(month)
	rows, err := store.db.Query(
		transactionsForMonthQuery,
		dateString(start),
		dateString(end),
		category.ID.Valid,
		category.ID.Int64,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly finance transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

func (store Store) CreateTransaction(input CreateTransactionInput) (int64, error) {
	result, err := store.db.Exec(`
		INSERT INTO finance_transactions (date, amount, category_id, description)
		VALUES (?, ?, ?, ?)
	`, input.Date, input.Amount, input.CategoryID, input.Description)
	if err != nil {
		return 0, fmt.Errorf("failed to create finance transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read created finance transaction id: %w", err)
	}
	return id, nil
}

func (store Store) UpdateTransaction(id int64, input UpdateTransactionInput) error {
	_, err := store.db.Exec(`
		UPDATE finance_transactions
		SET date = ?, amount = ?, category_id = ?, description = ?
		WHERE id = ?
	`, input.Date, input.Amount, input.CategoryID, input.Description, id)
	if err != nil {
		return fmt.Errorf("failed to update finance transaction: %w", err)
	}
	return nil
}

func (store Store) DeleteTransaction(id int64) error {
	_, err := store.db.Exec(`
		DELETE FROM finance_transactions
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete finance transaction: %w", err)
	}
	return nil
}

const transactionsForMonthQuery = `
	SELECT
		t.id,
		t.date,
		t.amount,
		t.category_id,
		c.name,
		c.type,
		t.description,
		t.created_at
	FROM finance_transactions t
	LEFT JOIN finance_categories c ON c.id = t.category_id
	WHERE t.date >= ? AND t.date < ?
		AND (? = 0 OR t.category_id = ?)
	ORDER BY t.date DESC, t.id DESC
	LIMIT ?
`

const recentTransactionsQuery = `
	SELECT
		t.id,
		t.date,
		t.amount,
		t.category_id,
		c.name,
		c.type,
		t.description,
		t.created_at
	FROM finance_transactions t
	LEFT JOIN finance_categories c ON c.id = t.category_id
	ORDER BY t.date DESC, t.id DESC
	LIMIT ?
`

func scanTransactions(rows *sql.Rows) ([]Transaction, error) {
	var transactions []Transaction
	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate finance transactions: %w", err)
	}
	return transactions, nil
}

func scanTransaction(rows *sql.Rows) (Transaction, error) {
	var transaction Transaction
	err := rows.Scan(
		&transaction.ID,
		&transaction.Date,
		&transaction.Amount,
		&transaction.CategoryID,
		&transaction.CategoryName,
		&transaction.CategoryType,
		&transaction.Description,
		&transaction.CreatedAt,
	)
	if err != nil {
		return Transaction{}, fmt.Errorf("scan finance transaction: %w", err)
	}
	return transaction, nil
}
