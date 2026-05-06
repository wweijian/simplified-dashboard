package ui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

func textAreaField(value string, width, lines int, active bool) string {
	contentWidth := max(width-2, 1)
	wrapped := wrapFieldValue(value, contentWidth)
	for len(wrapped) < lines {
		wrapped = append(wrapped, "")
	}
	if len(wrapped) > lines {
		wrapped = wrapped[:lines]
	}
	return borderedField(width, active).Render(strings.Join(wrapped, "\n"))
}

func wrapFieldValue(value string, width int) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{""}
	}

	var lines []string
	var current strings.Builder
	for _, word := range strings.Fields(value) {
		lines = appendWrappedFieldWord(lines, &current, word, width)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func appendWrappedFieldWord(lines []string, current *strings.Builder, word string, width int) []string {
	if lipgloss.Width(word) > width {
		return appendLongFieldWord(lines, current, word, width)
	}

	next := word
	if current.Len() > 0 {
		next = current.String() + " " + word
	}
	if lipgloss.Width(next) <= width {
		appendWord(current, word)
		return lines
	}

	lines = append(lines, current.String())
	current.Reset()
	current.WriteString(word)
	return lines
}

func appendLongFieldWord(lines []string, current *strings.Builder, word string, width int) []string {
	if current.Len() > 0 {
		lines = append(lines, current.String())
		current.Reset()
	}

	var chunk strings.Builder
	for _, r := range word {
		next := chunk.String() + string(r)
		if lipgloss.Width(next) > width {
			lines = append(lines, chunk.String())
			chunk.Reset()
		}
		chunk.WriteRune(r)
	}
	if chunk.Len() > 0 {
		current.WriteString(chunk.String())
	}
	return lines
}

func appendWord(current *strings.Builder, word string) {
	if current.Len() > 0 {
		current.WriteString(" ")
	}
	current.WriteString(word)
}

func (form addTaskForm) appendRunes(runes []rune) addTaskForm {
	for _, r := range runes {
		form = form.appendRune(r)
	}
	return form
}

func (form addTaskForm) appendRune(r rune) addTaskForm {
	switch form.field {
	case addTaskTitle:
		return form.appendTitleRune(r)
	case addTaskDescription:
		return form.appendDescriptionRune(r)
	case addTaskDueDays:
		return form.appendDueDaysRune(r)
	case addTaskPriority:
		return form.appendPriorityRune(r)
	default:
		return form
	}
}

func (form addTaskForm) appendTitleRune(r rune) addTaskForm {
	if len([]rune(form.title)) >= maxTaskTitleLength {
		form.errorText = titleLengthError()
		return form
	}
	form.title += string(r)
	return form
}

func (form addTaskForm) appendDescriptionRune(r rune) addTaskForm {
	if len([]rune(form.description)) >= maxTaskDescriptionChars {
		form.errorText = descriptionLengthError()
		return form
	}
	form.description += string(r)
	return form
}

func (form addTaskForm) appendDueDaysRune(r rune) addTaskForm {
	if unicode.IsDigit(r) || (r == '-' && form.dueDays == "") {
		form.dueDays += string(r)
	} else {
		form.errorText = "Due date accepts numbers only"
	}
	return form
}

func (form addTaskForm) deleteLastRune() addTaskForm {
	switch form.field {
	case addTaskTitle:
		form.title = trimLastRune(form.title)
	case addTaskDescription:
		form.description = trimLastRune(form.description)
	case addTaskDueDays:
		form.dueDays = trimLastRune(form.dueDays)
	}
	return form
}

func (form addTaskForm) clearField() addTaskForm {
	switch form.field {
	case addTaskTitle:
		form.title = ""
	case addTaskDescription:
		form.description = ""
	case addTaskDueDays:
		form.dueDays = ""
	}
	return form
}
