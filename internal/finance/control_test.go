package finance

import (
	"database/sql"
	"testing"
	"time"
)

func TestFinanceUpdateCanMoveMonthAndCategory(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")
	seedTransaction(t, sqlDB, "2026-05-02", -30, foodID, "Food")

	restore := stubCurrentMonth(dateMonth(2026, 5))
	defer restore()

	model := New(store)
	model = model.Update("left")
	if model.selectedMonth.Format("2006-01") != "2026-04" {
		t.Fatalf("expected previous month, got %s", model.selectedMonth)
	}

	model = model.Update("]")
	if !model.SelectedCategory().ID.Valid {
		t.Fatalf("expected a category filter, got %#v", model.SelectedCategory())
	}
}

func TestFinanceUpdateCanClearCategoryAndToggleComparison(t *testing.T) {
	store, sqlDB := openTestStore(t)
	seedCategory(t, sqlDB, "Food", "expense")

	restore := stubCurrentMonth(dateMonth(2026, 5))
	defer restore()

	model := New(store)
	model.categoryIndex = 1

	model = model.Update("c")
	if model.showComparison {
		t.Fatalf("expected comparison to be hidden")
	}

	model = model.Update("0")
	if model.categoryIndex != 0 {
		t.Fatalf("expected category filter to be cleared")
	}
}

func TestFinanceCanSelectCreateUpdateAndDeleteTransactions(t *testing.T) {
	store, sqlDB := openTestStore(t)
	foodID := seedCategory(t, sqlDB, "Food", "expense")
	billsID := seedCategory(t, sqlDB, "Bills", "expense")
	seedTransaction(t, sqlDB, "2026-05-02", -30, foodID, "Older")
	newerID := seedTransaction(t, sqlDB, "2026-05-03", -100, billsID, "Newer")

	restore := stubCurrentMonth(dateMonth(2026, 5))
	defer restore()

	model := New(store)
	if selected, ok := model.SelectedTransaction(); !ok || selected.ID != newerID {
		t.Fatalf("expected newest transaction selected, got %#v ok=%v", selected, ok)
	}

	model = model.Update("down")
	if selected, ok := model.SelectedTransaction(); !ok || selected.Description.String != "Older" {
		t.Fatalf("expected older transaction selected, got %#v ok=%v", selected, ok)
	}

	model = model.CreateTransaction(CreateTransactionInput{
		Date:        "2026-05-04",
		Amount:      -12,
		CategoryID:  sqlNullInt64(foodID),
		Description: sqlNullString("Coffee"),
	})
	if selected, ok := model.SelectedTransaction(); !ok || selected.Description.String != "Coffee" {
		t.Fatalf("expected created transaction selected, got %#v ok=%v", selected, ok)
	}
	if model.summary.Expenses != 142 {
		t.Fatalf("expected summary to refresh after create, got %.2f", model.summary.Expenses)
	}

	created, _ := model.SelectedTransaction()
	model = model.UpdateTransaction(created.ID, UpdateTransactionInput{
		Date:        "2026-05-04",
		Amount:      -20,
		CategoryID:  sqlNullInt64(foodID),
		Description: sqlNullString("Coffee plus snack"),
	})
	if selected, ok := model.SelectedTransaction(); !ok || selected.Amount != -20 {
		t.Fatalf("expected updated transaction selected, got %#v ok=%v", selected, ok)
	}

	model = model.Update("d")
	if len(model.transactions) != 2 {
		t.Fatalf("expected 2 transactions after delete, got %d", len(model.transactions))
	}
}

func stubCurrentMonth(month time.Time) func() {
	original := currentMonth
	currentMonth = func() time.Time { return month }
	return func() { currentMonth = original }
}

func sqlNullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}
