package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderLeftColumn(model Model, width, totalHeight int) string {
	base := totalHeight / 4
	extra := totalHeight - base*4
	panels := make([]string, 4)

	for i := range panels {
		h := base
		if i == 3 {
			h += extra
		}
		panels[i] = renderLeftPanel(model, i+1, width, h)
	}
	return lipgloss.JoinVertical(lipgloss.Left, panels...)
}

func renderLeftPanel(model Model, panelNum, width, height int) string {
	isActive := model.activePanel == panelNum

	labelLen := len(panelLabels[panelNum-1])
	dashCount := width - labelLen
	if dashCount < 0 {
		dashCount = 0
	}
	titleStyle := lipgloss.NewStyle().Width(width)
	if isActive {
		titleStyle = titleStyle.Foreground(color.accent).Bold(true)
	}
	title := titleStyle.Render(panelLabels[panelNum-1] + strings.Repeat("─", dashCount))

	var content string
	bodyHeight := height - 1
	switch panelNum {
	case int(Calendar):
		content = model.calendar.SummaryView(width, bodyHeight, isActive)
	case int(Tasks):
		content = model.tasks.SummaryView(width, bodyHeight, isActive)
	case int(Finance):
		content = model.finance.SummaryView(width, bodyHeight, isActive)
	case int(Habits):
		content = model.habits.SummaryView(width, bodyHeight, isActive)
	}

	body := lipgloss.NewStyle().Width(width).Height(bodyHeight).Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, title, body)
}
