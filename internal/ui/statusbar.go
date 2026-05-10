package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(model Model) string {
	hints := "1-4: jump  tab/shift+tab: panel  calendar: d/w/m arrows enter/e open a create r refresh  tasks: a/e/d [] s  finance: a/e/d arrows [] 0 c r  habits: a/e/d space  q quit"

	return lipgloss.NewStyle().
		Width(model.width).
		Render(hints)
}
