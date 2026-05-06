package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type theme struct {
	text   lipgloss.TerminalColor
	accent lipgloss.TerminalColor
	danger lipgloss.TerminalColor
}

var color = theme{
	text:   lipgloss.NoColor{},
	accent: lipgloss.Color("6"),
	danger: lipgloss.Color("1"),
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
	layout := lipgloss.JoinVertical(lipgloss.Left, row, renderStatusBar(model))
	if model.mode == ModeAddTask {
		return overlayCentered(layout, model.addTaskForm.View(model.width), model.width, model.height)
	}
	if model.mode == ModeEditTask {
		return overlayCentered(layout, model.addTaskForm.ViewWithTitle(model.width, "Edit task"), model.width, model.height)
	}
	if model.mode == ModeAddHabit {
		return overlayCentered(layout, model.addHabitForm.View(model.width), model.width, model.height)
	}
	if model.mode == ModeEditHabit {
		return overlayCentered(layout, model.addHabitForm.ViewWithTitle(model.width, "Edit habit"), model.width, model.height)
	}
	return layout
}

func overlayCentered(base, modal string, width, height int) string {
	baseLines := strings.Split(base, "\n")
	modalLines := strings.Split(modal, "\n")

	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	top := max((height-len(modalLines))/2, 0)

	for i, modalLine := range modalLines {
		lineIndex := top + i
		if lineIndex >= len(baseLines) {
			break
		}
		baseLines[lineIndex] = lipgloss.Place(
			width,
			1,
			lipgloss.Center,
			lipgloss.Center,
			modalLine,
		)
	}

	return strings.Join(baseLines, "\n")
}
