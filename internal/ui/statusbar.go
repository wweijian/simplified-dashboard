package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(model Model) string {
	hints := "1-4: jump  tab/shift+tab: switch panel  tasks: a add  e edit  [] filter  s sort  space toggle  d delete  q quit"

	return lipgloss.NewStyle().
		Width(model.width).
		Render(hints)
}
