package finance

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestSummaryViewShowsSpendAndTopCategories(t *testing.T) {
	model := Model{
		summary: MonthlySummary{
			Month:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
			Expenses: 1234.50,
		},
		topCategories: []CategorySpend{
			{Name: "Food", Amount: 420},
			{Name: "Transport", Amount: 80.25},
		},
		transactions: []Transaction{{ID: 1}},
	}

	got := model.SummaryView(40, 3, false)

	if !strings.Contains(got, "May spend: $1,234.50") {
		t.Fatalf("expected month spend, got:\n%s", got)
	}
	if !strings.Contains(got, "Food: $420.00") || !strings.Contains(got, "Transport: $80.25") {
		t.Fatalf("expected top categories, got:\n%s", got)
	}
}

func TestExpandedViewShowsRecentTransactions(t *testing.T) {
	got := expandedTestModel().ExpandedView(80, 20)

	if !strings.Contains(got, "May 2026      Category: All      Compare: Apr 2026") {
		t.Fatalf("expected analytics header, got:\n%s", got)
	}
	if !strings.Contains(got, "Spend: $100.00") || !strings.Contains(got, "Income: +$5,000.00") {
		t.Fatalf("expected spend line, got:\n%s", got)
	}
	if !strings.Contains(got, "Spending by Category") || !strings.Contains(got, "Food") {
		t.Fatalf("expected category spending chart, got:\n%s", got)
	}
	if !strings.Contains(got, "06 May") || !strings.Contains(got, "Uber Eats") || !strings.Contains(got, "-$24.50") || !strings.Contains(got, "Food") {
		t.Fatalf("expected recent transaction row, got:\n%s", got)
	}
}

func expandedTestModel() Model {
	return Model{
		selectedMonth:     dateMonth(2026, 5),
		showComparison:    true,
		summary:           expandedTestSummary(),
		categories:        []Category{{ID: 2, Name: "Food", Type: "expense"}},
		categoryBreakdown: []CategorySpend{{Name: "Food", Amount: 100}},
		monthlyTrend:      expandedTestTrend(),
		transactions:      expandedTestTransactions(),
	}
}

func expandedTestSummary() MonthlySummary {
	return MonthlySummary{Month: dateMonth(2026, 5), Income: 5000, Expenses: 100, Net: 4900}
}

func expandedTestTrend() []MonthSpend {
	return []MonthSpend{
		{Month: dateMonth(2026, 4), Amount: 80},
		{Month: dateMonth(2026, 5), Amount: 100},
	}
}

func expandedTestTransactions() []Transaction {
	return []Transaction{{
		Date:         "2026-05-06",
		Amount:       -24.50,
		Description:  sql.NullString{String: "Uber Eats", Valid: true},
		CategoryName: sql.NullString{String: "Food", Valid: true},
	}}
}

func TestExpandedViewShowsCategoryTrendWhenFiltered(t *testing.T) {
	model := Model{
		selectedMonth:  time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
		categoryIndex:  1,
		categories:     []Category{{ID: 2, Name: "Food", Type: "expense"}},
		monthlyTrend:   []MonthSpend{{Month: time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local), Amount: 43.40}},
		transactions:   []Transaction{},
		showComparison: false,
	}

	got := model.ExpandedView(80, 20)

	if !strings.Contains(got, "Category: Food") {
		t.Fatalf("expected selected category in header, got:\n%s", got)
	}
	if !strings.Contains(got, "Monthly Spending Trend") || !strings.Contains(got, "May") {
		t.Fatalf("expected monthly trend chart, got:\n%s", got)
	}
}

func TestCategoryChartShowsZeroSpendCategories(t *testing.T) {
	model := Model{
		selectedMonth: dateMonth(2026, 5),
		categories: []Category{
			{ID: 1, Name: "Bills", Type: "expense"},
			{ID: 2, Name: "Food", Type: "expense"},
		},
		categoryBreakdown: []CategorySpend{{Name: "Food", Amount: 43.40}},
	}

	got := model.ExpandedView(90, 20)

	if !strings.Contains(got, "Bills") || !strings.Contains(got, "$0.00") {
		t.Fatalf("expected zero-spend category row, got:\n%s", got)
	}
}

func TestTransactionLinesScrollAroundSelectedTransaction(t *testing.T) {
	model := Model{
		transactions: []Transaction{
			testTransaction("First"),
			testTransaction("Second"),
			testTransaction("Third"),
			testTransaction("Fourth"),
			testTransaction("Fifth"),
		},
		selectedTransaction: 3,
	}

	lines := model.transactionLines(3)

	if len(lines) != 3 {
		t.Fatalf("expected 3 visible transaction lines, got %d", len(lines))
	}
	if strings.Contains(strings.Join(lines, "\n"), "First") {
		t.Fatalf("expected first transaction to be scrolled away, got %q", lines)
	}
	if !strings.Contains(lines[1], ">") || !strings.Contains(lines[1], "Fourth") {
		t.Fatalf("expected selected fourth transaction in middle, got %q", lines)
	}
}

func testTransaction(description string) Transaction {
	return Transaction{
		Date:         "2026-05-06",
		Amount:       -1,
		Description:  sql.NullString{String: description, Valid: true},
		CategoryName: sql.NullString{String: "Food", Valid: true},
	}
}

func TestFormatSignedMoneyRoundsCarry(t *testing.T) {
	got := formatSignedMoney(-9.999)
	if got != "-$10.00" {
		t.Fatalf("expected -$10, got %q", got)
	}
}
