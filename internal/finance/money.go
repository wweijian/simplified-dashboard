package finance

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func formatMoney(amount float64) string {
	return "$" + formatAbsoluteMoney(amount)
}

func formatSignedMoney(amount float64) string {
	if amount < 0 {
		return "-$" + formatAbsoluteMoney(amount)
	}
	return "+$" + formatAbsoluteMoney(amount)
}

func formatAbsoluteMoney(amount float64) string {
	amount = math.Abs(amount)
	whole := int64(amount)
	cents := int(math.Round((amount - float64(whole)) * 100))
	if cents == 100 {
		whole++
		cents = 0
	}

	return fmt.Sprintf("%s.%02d", commaInt(whole), cents)
}

func coloredMoney(amount float64) string {
	return financeMoneyStyle.Render(formatMoney(amount))
}

func coloredSignedMoney(amount float64) string {
	if amount < 0 {
		return financeNegativeStyle.Render(formatSignedMoney(amount))
	}
	return financeMoneyStyle.Render(formatSignedMoney(amount))
}

func commaInt(value int64) string {
	text := strconv.FormatInt(value, 10)
	if len(text) <= 3 {
		return text
	}

	var parts []string
	for len(text) > 3 {
		parts = append([]string{text[len(text)-3:]}, parts...)
		text = text[:len(text)-3]
	}
	parts = append([]string{text}, parts...)
	return strings.Join(parts, ",")
}
