package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func renderRightPanel(model Model, width, height int) string {
	style := lipgloss.NewStyle().Width(width).Height(height).Padding(0, 1)
	innerWidth := width - 2

	switch model.activePanel {
	case int(Calendar):
		return style.Render(model.calendar.ExpandedView(innerWidth, height))
	case int(Tasks):
		return style.Render(model.tasks.ExpandedView(innerWidth, height))
	case int(Finance):
		return style.Render(model.finance.ExpandedView(innerWidth, height))
	case int(Habits):
		return style.Render(model.habits.ExpandedView(innerWidth, height))
	default:
		panic("ui/rightcolumn.go: unknown number received")
	}
}
