package finance

import (
	"fmt"
	"strings"
)

func (model Model) ExpandedView(width, height int) string {
	if model.err != nil {
		return "Error loading finance:\n" + model.err.Error()
	}

	lines := []string{model.analyticsHeader(), "", model.spendLine()}
	if model.showComparison {
		lines = append(lines, model.comparisonLine())
	}

	lines = append(lines, "", financeDivider(width), "")
	lines = append(lines, model.analyticsChart(width)...)
	lines = append(lines, "", "", financeTitleStyle.Render("Recent Transactions"), "")
	lines = append(lines, model.transactionLines()...)

	return strings.Join(limitLines(lines, height), "\n")
}

func (model Model) analyticsHeader() string {
	category := model.SelectedCategory().Name
	compare := model.selectedMonth.AddDate(0, -1, 0).Format("Jan 2006")
	return financeTitleStyle.Render(model.selectedMonth.Format("Jan 2006")) +
		financeMetaStyle.Render(fmt.Sprintf("      Category: %s      Compare: %s",
			category,
			compare,
		))
}

func (model Model) spendLine() string {
	return fmt.Sprintf("Spend: %s      Income: %s      Net: %s",
		coloredMoney(model.currentSpend()),
		coloredSignedMoney(model.summary.Income),
		coloredSignedMoney(model.summary.Net),
	)
}

func (model Model) comparisonLine() string {
	return financeMetaStyle.Render(fmt.Sprintf("Previous month: %s      Change: %s",
		formatMoney(model.previousSpend()),
		formatSignedMoney(model.currentSpend()-model.previousSpend()),
	))
}

func (model Model) analyticsChart(width int) []string {
	if model.SelectedCategory().ID.Valid {
		return renderBars("Monthly Spending Trend", monthSpendBars(model.monthlyTrend), width)
	}
	return renderBars("Spending by Category", model.categoryBars(), width)
}

func (model Model) categoryBars() []barDatum {
	return categorySpendBars(model.categories, model.categoryBreakdown)
}

func financeDivider(width int) string {
	return financeMetaStyle.Render(strings.Repeat("─", max(width, 0)))
}

func (model Model) transactionLines() []string {
	if len(model.transactions) == 0 {
		return []string{"No transactions for this selection"}
	}

	lines := make([]string, 0, len(model.transactions))
	for _, transaction := range model.transactions {
		lines = append(lines, transactionRow(transaction))
	}
	return lines
}
