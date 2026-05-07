package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(model Model) string {
	hints := "1-4: jump  tab/shift+tab: panel  finance: left/right month [] category 0 all c compare  tasks: [] filter s sort  habits: space/c/s/x mark  q quit"

	return lipgloss.NewStyle().
		Width(model.width).
		Render(hints)
}
