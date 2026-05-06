package ui

import "github.com/charmbracelet/lipgloss"

func priorityField(priority int, width int, active bool) string {
	options := []string{"low", "medium", "high"}
	segments := make([]string, 0, len(options))
	for i, option := range options {
		segments = append(segments, prioritySegment(option, i == priority))
	}
	return borderedField(width, active).Render(lipgloss.JoinHorizontal(lipgloss.Left, segments...))
}

func prioritySegment(option string, selected bool) string {
	style := lipgloss.NewStyle().Padding(0, 2)
	if selected {
		style = style.Bold(true).Foreground(color.accent)
	} else {
		style = style.Foreground(color.text)
	}
	return style.Render(option)
}

func (form addTaskForm) appendPriorityRune(r rune) addTaskForm {
	switch r {
	case '1', 'l', 'L':
		form.priority = 0
	case '2', 'm', 'M':
		form.priority = 1
	case '3', 'h', 'H':
		form.priority = 2
	case ' ', 'p', 'P':
		form.priority = (form.priority + 1) % 3
	}
	return form
}

func (form addTaskForm) movePriority(distance int) addTaskForm {
	if form.field == addTaskPriority {
		form.priority = (form.priority + distance + 3) % 3
	}
	return form
}
