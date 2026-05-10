package finance

import (
	"database/sql"
	"testing"
	"time"

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

	createFinanceTables(t, sqlDB)

	return NewStore(&db.DB{DB: sqlDB}), sqlDB
}

func createFinanceTables(t *testing.T, sqlDB *sql.DB) {
	t.Helper()

	if _, err := sqlDB.Exec(financeTestSchema); err != nil {
		t.Fatalf("create finance tables: %v", err)
	}
}

const financeTestSchema = `
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
`

func seedCategory(t *testing.T, sqlDB *sql.DB, name, categoryType string) int64 {
	t.Helper()

	result, err := sqlDB.Exec(`
		INSERT INTO finance_categories (name, type)
		VALUES (?, ?)
	`, name, categoryType)
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("category id: %v", err)
	}
	return id
}

func seedTransaction(t *testing.T, sqlDB *sql.DB, date string, amount float64, categoryID int64, description string) int64 {
	t.Helper()

	result, err := sqlDB.Exec(`
		INSERT INTO finance_transactions (date, amount, category_id, description)
		VALUES (?, ?, ?, ?)
	`, date, amount, categoryID, description)
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("transaction id: %v", err)
	}
	return id
}

func TestEnsureDefaultCategoriesIsIdempotent(t *testing.T) {
	store, _ := openTestStore(t)

	if err := store.EnsureDefaultCategories(); err != nil {
		t.Fatalf("seed default categories: %v", err)
	}
	if err := store.EnsureDefaultCategories(); err != nil {
		t.Fatalf("seed default categories again: %v", err)
	}

	categories, err := store.ListCategories()
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	if len(categories) != 10 {
		t.Fatalf("expected 10 default categories, got %d", len(categories))
	}
}

func TestMonthlySummaryUsesSelectedMonth(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")
	incomeID := seedCategory(t, sqlDB, "Income", "income")

	seedTransaction(t, sqlDB, "2026-05-02", -24.50, foodID, "Lunch")
	seedTransaction(t, sqlDB, "2026-05-04", -75, foodID, "Groceries")
	seedTransaction(t, sqlDB, "2026-05-05", 3000, incomeID, "Salary")
	seedTransaction(t, sqlDB, "2026-04-30", -999, foodID, "Old month")

	summary, err := store.MonthlySummary(time.Date(2026, 5, 6, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("monthly summary: %v", err)
	}

	if summary.Income != 3000 || summary.Expenses != 99.50 || summary.Net != 2900.50 {
		t.Fatalf("summary = income %.2f expenses %.2f net %.2f", summary.Income, summary.Expenses, summary.Net)
	}
}

func TestTopExpenseCategories(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")
	transportID := seedCategory(t, sqlDB, "Transport", "expense")
	incomeID := seedCategory(t, sqlDB, "Income", "income")

	seedTransaction(t, sqlDB, "2026-05-02", -24.50, foodID, "Lunch")
	seedTransaction(t, sqlDB, "2026-05-04", -75, foodID, "Groceries")
	seedTransaction(t, sqlDB, "2026-05-05", -12, transportID, "Train")
	seedTransaction(t, sqlDB, "2026-05-06", 3000, incomeID, "Salary")

	categories, err := store.TopExpenseCategories(time.Date(2026, 5, 6, 0, 0, 0, 0, time.Local), 2)
	if err != nil {
		t.Fatalf("top categories: %v", err)
	}

	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}
	if categories[0].Name != "Food" || categories[0].Amount != 99.50 {
		t.Fatalf("expected Food at 99.50 first, got %#v", categories[0])
	}
	if categories[1].Name != "Transport" || categories[1].Amount != 12 {
		t.Fatalf("expected Transport at 12 second, got %#v", categories[1])
	}
}

func TestListRecentTransactionsNewestFirst(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")

	seedTransaction(t, sqlDB, "2026-05-02", -24.50, foodID, "Older")
	seedTransaction(t, sqlDB, "2026-05-04", -75, foodID, "Newer")

	transactions, err := store.ListRecentTransactions(1)
	if err != nil {
		t.Fatalf("recent transactions: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(transactions))
	}
	if !transactions[0].Description.Valid || transactions[0].Description.String != "Newer" {
		t.Fatalf("expected newest transaction first, got %#v", transactions[0])
	}
	if !transactions[0].CategoryName.Valid || transactions[0].CategoryName.String != "Food" {
		t.Fatalf("expected joined category, got %#v", transactions[0].CategoryName)
	}
}

func TestCreateUpdateDeleteTransaction(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")
	billsID := seedCategory(t, sqlDB, "Bills", "expense")

	id, err := store.CreateTransaction(CreateTransactionInput{
		Date:        "2026-05-06",
		Amount:      -24.50,
		CategoryID:  sql.NullInt64{Int64: foodID, Valid: true},
		Description: sql.NullString{String: "Lunch", Valid: true},
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	if err := store.UpdateTransaction(id, UpdateTransactionInput{
		Date:        "2026-05-07",
		Amount:      -100,
		CategoryID:  sql.NullInt64{Int64: billsID, Valid: true},
		Description: sql.NullString{String: "Power", Valid: true},
	}); err != nil {
		t.Fatalf("update transaction: %v", err)
	}

	var date string
	var amount float64
	var categoryID int64
	var description string
	if err := sqlDB.QueryRow(`
		SELECT date, amount, category_id, description
		FROM finance_transactions
		WHERE id = ?
	`, id).Scan(&date, &amount, &categoryID, &description); err != nil {
		t.Fatalf("query transaction: %v", err)
	}
	if date != "2026-05-07" || amount != -100 || categoryID != billsID || description != "Power" {
		t.Fatalf("updated transaction = (%q, %.2f, %d, %q)", date, amount, categoryID, description)
	}

	if err := store.DeleteTransaction(id); err != nil {
		t.Fatalf("delete transaction: %v", err)
	}

	var count int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM finance_transactions`).Scan(&count); err != nil {
		t.Fatalf("count transactions: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected transaction to be deleted, got %d", count)
	}
}
