package finance

import (
	"database/sql"
	"fmt"
	"time"
)

func (store Store) MonthlySummary(month time.Time) (MonthlySummary, error) {
	start, end := monthBounds(month)
	summary := MonthlySummary{Month: start}

	err := store.db.QueryRow(monthlySummaryQuery, dateString(start), dateString(end)).
		Scan(&summary.Income, &summary.Expenses, &summary.Net)
	if err != nil {
		return MonthlySummary{}, fmt.Errorf("query monthly finance summary: %w", err)
	}
	return summary, nil
}

const monthlySummaryQuery = `
	SELECT
		COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN amount < 0 THEN -amount ELSE 0 END), 0),
		COALESCE(SUM(amount), 0)
	FROM finance_transactions
	WHERE date >= ? AND date < ?
`

func (store Store) TopExpenseCategories(month time.Time, limit int) ([]CategorySpend, error) {
	start, end := monthBounds(month)
	rows, err := store.db.Query(topExpenseCategoriesQuery, dateString(start), dateString(end), limit)
	if err != nil {
		return nil, fmt.Errorf("query top finance categories: %w", err)
	}
	defer rows.Close()

	return scanCategorySpends(rows)
}

func (store Store) SpendingByCategory(month time.Time) ([]CategorySpend, error) {
	start, end := monthBounds(month)
	rows, err := store.db.Query(spendingByCategoryQuery, dateString(start), dateString(end))
	if err != nil {
		return nil, fmt.Errorf("query finance category spending: %w", err)
	}
	defer rows.Close()

	return scanCategorySpends(rows)
}

const spendingByCategoryQuery = `
	SELECT COALESCE(c.name, 'Uncategorized'), SUM(-t.amount)
	FROM finance_transactions t
	LEFT JOIN finance_categories c ON c.id = t.category_id
	WHERE t.date >= ? AND t.date < ? AND t.amount < 0
	GROUP BY COALESCE(c.name, 'Uncategorized')
	ORDER BY SUM(-t.amount) DESC, COALESCE(c.name, 'Uncategorized') ASC
`

const topExpenseCategoriesQuery = `
	SELECT COALESCE(c.name, 'Uncategorized'), SUM(-t.amount)
	FROM finance_transactions t
	LEFT JOIN finance_categories c ON c.id = t.category_id
	WHERE t.date >= ? AND t.date < ? AND t.amount < 0
	GROUP BY COALESCE(c.name, 'Uncategorized')
	ORDER BY SUM(-t.amount) DESC, COALESCE(c.name, 'Uncategorized') ASC
	LIMIT ?
`

func scanCategorySpends(rows *sql.Rows) ([]CategorySpend, error) {
	var categories []CategorySpend
	for rows.Next() {
		var category CategorySpend
		if err := rows.Scan(&category.Name, &category.Amount); err != nil {
			return nil, fmt.Errorf("scan finance category spend: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate finance category spend: %w", err)
	}
	return categories, nil
}

func (store Store) MonthlySpendTrend(month time.Time, months int, category CategoryFilter) ([]MonthSpend, error) {
	start := monthStart(month).AddDate(0, -(months - 1), 0)
	end := monthStart(month).AddDate(0, 1, 0)
	rows, err := store.db.Query(monthlySpendTrendQuery, dateString(start), dateString(end), category.ID.Valid, category.ID.Int64)
	if err != nil {
		return nil, fmt.Errorf("query monthly spending trend: %w", err)
	}
	defer rows.Close()

	amounts, err := scanMonthlySpend(rows)
	if err != nil {
		return nil, err
	}
	return buildMonthlySpendTrend(start, months, amounts), nil
}

const monthlySpendTrendQuery = `
	SELECT strftime('%Y-%m-01', t.date), SUM(-t.amount)
	FROM finance_transactions t
	WHERE t.date >= ? AND t.date < ? AND t.amount < 0
		AND (? = 0 OR t.category_id = ?)
	GROUP BY strftime('%Y-%m-01', t.date)
	ORDER BY strftime('%Y-%m-01', t.date) ASC
`

func scanMonthlySpend(rows *sql.Rows) (map[string]float64, error) {
	amounts := map[string]float64{}
	for rows.Next() {
		var month string
		var amount float64
		if err := rows.Scan(&month, &amount); err != nil {
			return nil, fmt.Errorf("scan monthly spending trend: %w", err)
		}
		amounts[month] = amount
	}
	return amounts, rows.Err()
}

func buildMonthlySpendTrend(start time.Time, months int, amounts map[string]float64) []MonthSpend {
	trend := make([]MonthSpend, 0, months)
	for i := 0; i < months; i++ {
		month := start.AddDate(0, i, 0)
		trend = append(trend, MonthSpend{Month: month, Amount: amounts[dateString(month)]})
	}
	return trend
}

func monthBounds(month time.Time) (time.Time, time.Time) {
	start := monthStart(month)
	return start, start.AddDate(0, 1, 0)
}

func monthStart(month time.Time) time.Time {
	return time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
}

func dateString(date time.Time) string {
	return date.Format("2006-01-02")
}
