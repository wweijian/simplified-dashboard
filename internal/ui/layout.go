package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type theme struct {
	dim    lipgloss.Color
	active lipgloss.Color
	accent lipgloss.Color
}

var color = theme{
	dim:    lipgloss.Color("#343538"),
	active: lipgloss.Color("#1e2040"),
	accent: lipgloss.Color("#7aa2f7"),
}

var panelLabels = []string{
	"[1]─Calendar",
	"[2]─Tasks",
	"[3]─Finance",
	"[4]─Habits",
}

type PanelNumber int

const (
	Calendar PanelNumber = iota + 1
	Tasks
	Finance
	Habits
)

func renderLayout(model Model) string {
	leftWidth, rightWidth := panelWidths(model.width)
	contentHeight := model.height - 1

	left := renderLeftColumn(model, leftWidth, contentHeight)
	divider := strings.Repeat("│\n", contentHeight-1) + "│"
	right := renderRightPanel(model, rightWidth, contentHeight)

	row := lipgloss.JoinHorizontal(lipgloss.Top, left, divider, right)
	return lipgloss.JoinVertical(lipgloss.Left, row, renderStatusBar(model))
}
