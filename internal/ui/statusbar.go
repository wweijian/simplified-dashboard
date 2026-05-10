package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(model Model) string {
	hints := "1-4: jump  tab/shift+tab: panel  calendar: d/w/m left/right up/down enter/e open a create r refresh  tasks: [] filter s sort  finance: c compare  q quit"

	return lipgloss.NewStyle().
		Width(model.width).
		Render(hints)
}
