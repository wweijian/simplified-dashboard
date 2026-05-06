package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(model Model) string {
	hints := "1-4: jump  tab/shift+tab: switch panel  a add  e edit  tasks: [] filter s sort  habits: left/right day space cycle c/s/x mark  q quit"

	return lipgloss.NewStyle().
		Width(model.width).
		Render(hints)
}
