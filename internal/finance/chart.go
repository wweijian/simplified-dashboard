package finance

import (
	"fmt"
	"math"
	"strings"
)

type barDatum struct {
	Label string
	Value float64
}

func renderBars(title string, data []barDatum, width int) []string {
	if len(data) == 0 {
		return []string{title, "No spending data"}
	}

	labelWidth := barLabelWidth(data)
	barWidth := max(width-labelWidth-22, 4)
	maxValue := maxBarValue(data)
	lines := []string{financeTitleStyle.Render(title), ""}

	for _, item := range data {
		lines = append(lines, renderBarRow(item, labelWidth, barWidth, maxValue))
	}
	return lines
}

func renderBarRow(item barDatum, labelWidth, barWidth int, maxValue float64) string {
	filled := scaledBarWidth(item.Value, maxValue, barWidth)
	bar := financeBarStyle.Render(paddedBar(filled, barWidth))
	return fmt.Sprintf("%-*s    %s    %s",
		labelWidth,
		truncate(item.Label, labelWidth),
		bar,
		coloredMoney(item.Value),
	)
}

func paddedBar(filled, width int) string {
	return strings.Repeat("█", filled) + strings.Repeat(" ", max(width-filled, 0))
}

func scaledBarWidth(value, maxValue float64, width int) int {
	if maxValue <= 0 || value <= 0 {
		return 0
	}
	return max(1, int(math.Round(value/maxValue*float64(width))))
}

func barLabelWidth(data []barDatum) int {
	width := 8
	for _, item := range data {
		width = max(width, min(len(item.Label), 14))
	}
	return width
}

func maxBarValue(data []barDatum) float64 {
	var maxValue float64
	for _, item := range data {
		if item.Value > maxValue {
			maxValue = item.Value
		}
	}
	return maxValue
}

func categorySpendBars(categories []Category, spending []CategorySpend) []barDatum {
	amounts := categorySpendMap(spending)
	bars := make([]barDatum, 0, len(categories)+len(spending))
	for _, category := range categories {
		bars = append(bars, barDatum{Label: category.Name, Value: amounts[category.Name]})
	}
	return appendExtraSpendBars(bars, amounts)
}

func appendExtraSpendBars(bars []barDatum, amounts map[string]float64) []barDatum {
	for name, amount := range amounts {
		if !hasBarLabel(bars, name) {
			bars = append(bars, barDatum{Label: name, Value: amount})
		}
	}
	return bars
}

func categorySpendMap(spending []CategorySpend) map[string]float64 {
	amounts := map[string]float64{}
	for _, category := range spending {
		amounts[category.Name] = category.Amount
	}
	return amounts
}

func hasBarLabel(bars []barDatum, label string) bool {
	for _, bar := range bars {
		if bar.Label == label {
			return true
		}
	}
	return false
}

func monthSpendBars(months []MonthSpend) []barDatum {
	bars := make([]barDatum, 0, len(months))
	for _, month := range months {
		bars = append(bars, barDatum{Label: month.Month.Format("Jan"), Value: month.Amount})
	}
	return bars
}
