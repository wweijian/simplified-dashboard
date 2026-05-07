package finance

import (
	"fmt"
	"strings"
)

func (model Model) SummaryView(width, height int, focused bool) string {
	if model.err != nil {
		return "error loading finance"
	}

	lines := []string{
		fmt.Sprintf("%s spend: %s", model.summary.Month.Format("Jan"), formatMoney(model.summary.Expenses)),
	}

	for _, category := range model.topCategories {
		lines = append(lines, fmt.Sprintf("%s: %s", category.Name, formatMoney(category.Amount)))
		if len(lines) >= height {
			break
		}
	}

	if len(lines) == 1 && len(model.transactions) == 0 {
		lines = append(lines, "No transactions")
	}
	return strings.Join(limitLines(lines, height), "\n")
}
