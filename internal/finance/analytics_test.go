package finance

import (
	"database/sql"
	"testing"
	"time"
)

func TestSpendingByCategoryGroupsInvalidCategoryAsUncategorized(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")

	seedTransaction(t, sqlDB, "2026-05-02", -24.50, foodID, "Lunch")
	seedTransaction(t, sqlDB, "2026-05-03", -10, 999, "Mystery")

	spending, err := store.SpendingByCategory(dateMonth(2026, 5))
	if err != nil {
		t.Fatalf("spending by category: %v", err)
	}

	assertCategorySpend(t, spending, "Food", 24.50)
	assertCategorySpend(t, spending, "Uncategorized", 10)
}

func TestMonthlySpendTrendCanFilterByCategory(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")
	billsID := seedCategory(t, sqlDB, "Bills", "expense")

	seedTransaction(t, sqlDB, "2026-04-02", -20, foodID, "April food")
	seedTransaction(t, sqlDB, "2026-05-02", -30, foodID, "May food")
	seedTransaction(t, sqlDB, "2026-05-03", -100, billsID, "May bill")

	trend, err := store.MonthlySpendTrend(dateMonth(2026, 5), 2, categoryFilter(foodID, "Food"))
	if err != nil {
		t.Fatalf("monthly spend trend: %v", err)
	}

	if len(trend) != 2 || trend[0].Amount != 20 || trend[1].Amount != 30 {
		t.Fatalf("expected filtered monthly food trend [20, 30], got %#v", trend)
	}
}

func TestTransactionsForMonthCanFilterByCategory(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")
	billsID := seedCategory(t, sqlDB, "Bills", "expense")

	seedTransaction(t, sqlDB, "2026-05-02", -30, foodID, "Food")
	seedTransaction(t, sqlDB, "2026-05-03", -100, billsID, "Bill")

	transactions, err := store.TransactionsForMonth(dateMonth(2026, 5), categoryFilter(foodID, "Food"), 10)
	if err != nil {
		t.Fatalf("transactions for month: %v", err)
	}

	if len(transactions) != 1 || transactions[0].Description.String != "Food" {
		t.Fatalf("expected only food transaction, got %#v", transactions)
	}
}

func assertCategorySpend(t *testing.T, spending []CategorySpend, name string, amount float64) {
	t.Helper()

	for _, category := range spending {
		if category.Name == name && category.Amount == amount {
			return
		}
	}
	t.Fatalf("expected %s spend %.2f in %#v", name, amount, spending)
}

func categoryFilter(id int64, name string) CategoryFilter {
	return CategoryFilter{ID: sql.NullInt64{Int64: id, Valid: true}, Name: name}
}

func dateMonth(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
}
