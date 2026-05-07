package finance

import (
	"fmt"
	"strconv"
	"strings"
)

func transactionRow(transaction Transaction) string {
	return fmt.Sprintf("%-8s    %-30s    %14s    %s",
		displayDate(transaction.Date),
		truncate(transactionDescription(transaction), 30),
		coloredSignedMoney(transaction.Amount),
		transactionCategory(transaction),
	)
}

func transactionDescription(transaction Transaction) string {
	if transaction.Description.Valid && transaction.Description.String != "" {
		return transaction.Description.String
	}
	return "No description"
}

func transactionCategory(transaction Transaction) string {
	if transaction.CategoryName.Valid && transaction.CategoryName.String != "" {
		return transaction.CategoryName.String
	}
	return "Uncategorized"
}

func displayDate(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return date
	}

	month, day, ok := parseMonthDay(parts[1], parts[2])
	if !ok {
		return date
	}
	return fmt.Sprintf("%02d %s", day, monthName(month))
}

func parseMonthDay(monthText, dayText string) (int, int, bool) {
	month, monthErr := strconv.Atoi(monthText)
	day, dayErr := strconv.Atoi(dayText)
	if monthErr != nil || dayErr != nil || month < 1 || month > 12 {
		return 0, 0, false
	}
	return month, day, true
}

func monthName(month int) string {
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	return months[month-1]
}

func truncate(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	if maxLength <= 1 {
		return text[:maxLength]
	}
	return text[:maxLength-1] + "."
}

func limitLines(lines []string, height int) []string {
	if height <= 0 || len(lines) <= height {
		return lines
	}
	return lines[:height]
}
