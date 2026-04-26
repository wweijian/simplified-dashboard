package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(model Model) string {
	hints := "1-4: jump  ↑↓: navigate  →/tab: right panel  ←/esc: back  q: quit  ?: help"
	return lipgloss.NewStyle().
		Width(model.width).
		Foreground(color.dim).
		Render(hints)
}