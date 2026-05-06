package ui

import (
	"database/sql"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func modalWidth(width int) int {
	if width < 36 {
		return max(width-8, 20)
	}
	return min(width-8, 76)
}

func modalStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color.accent).
		Padding(1, 3)
}

func formTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(color.accent).
		MarginBottom(1)
}

func fieldLabel(label string, active bool) string {
	style := lipgloss.NewStyle().Foreground(color.text)
	if active {
		style = style.Foreground(color.accent).Bold(true)
	}
	return style.Render(label)
}

func borderedField(width int, active bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)
	if active {
		return style.BorderForeground(color.accent)
	}
	return style.BorderForeground(color.text)
}

func inputField(value string, width int, active bool) string {
	if value == "" {
		value = " "
	}
	return borderedField(width, active).Render(fitFieldValue(value, width-2))
}

func errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(color.danger).
		MarginTop(1)
}

func fitFieldValue(value string, width int) string {
	if lipgloss.Width(value) <= width {
		return value
	}

	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width(string(runes))+3 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "..."
}

func trimLastRune(value string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)
	return string(runes[:len(runes)-1])
}

func nullableStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func renderModal(width int, lines []string) string {
	return modalStyle(width).Render(strings.Join(lines, "\n"))
}
