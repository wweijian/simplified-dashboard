package finance

import (
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

func stubCurrentMonth(month time.Time) func() {
	original := currentMonth
	currentMonth = func() time.Time { return month }
	return func() { currentMonth = original }
}
