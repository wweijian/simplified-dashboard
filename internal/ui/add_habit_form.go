package ui

import (
	"strings"

	"simplified-dashboard/internal/habits"

	bbt "github.com/charmbracelet/bubbletea"
)

type addHabitField int

const (
	addHabitName addHabitField = iota
	addHabitDays
)

const (
	addHabitFieldCount = 2
	maxHabitNameLength = 50
)

type addHabitForm struct {
	name      string
	weekdays  [7]bool
	dayCursor int
	field     addHabitField
	errorText string
}

func newAddHabitForm() addHabitForm {
	form := addHabitForm{}
	for i := range form.weekdays {
		form.weekdays[i] = true
	}
	return form
}

func newEditHabitForm(habit habits.Habit) addHabitForm {
	form := newAddHabitForm()
	form.name = habit.Name
	form.weekdays = weekdaysFromHabit(habit)
	return form
}

func (form addHabitForm) Update(msg bbt.KeyMsg) addHabitForm {
	form.errorText = ""

	switch msg.String() {
	case "tab", "down":
		form.field = (form.field + 1) % addHabitFieldCount
	case "shift+tab", "up":
		form.field = (form.field + addHabitFieldCount - 1) % addHabitFieldCount
	case "left":
		form = form.moveDayCursor(-1)
	case "right":
		form = form.moveDayCursor(1)
	case "backspace", "ctrl+h":
		form = form.deleteLastRune()
	case "ctrl+u":
		form = form.clearField()
	case " ":
		form = form.toggleSelectedDay()
	default:
		form = form.appendRunes(msg.Runes)
	}
	return form
}

func (form addHabitForm) Submit() (addHabitForm, habits.CreateHabitInput, bool) {
	name := strings.TrimSpace(form.name)
	if !form.validateName(name) {
		return form, habits.CreateHabitInput{}, false
	}

	weekdays := selectedWeekdays(form.weekdays)
	if !form.validateWeekdays(weekdays) {
		return form, habits.CreateHabitInput{}, false
	}

	input, err := habitInputFromDays(name, weekdays)
	if err != nil {
		form.errorText = err.Error()
		return form, habits.CreateHabitInput{}, false
	}
	return form, input, true
}

func (form addHabitForm) View(width int) string {
	return form.ViewWithTitle(width, "Add habit")
}

func (form addHabitForm) ViewWithTitle(width int, title string) string {
	modalWidth := modalWidth(width)
	fieldWidth := max(modalWidth-12, 12)

	lines := []string{
		formTitleStyle().Render(title),
		fieldLabel("Name", form.field == addHabitName),
		inputField(form.name, fieldWidth, form.field == addHabitName),
		fieldLabel("Days", form.field == addHabitDays),
		habitDaysField(form.weekdays, form.dayCursor, fieldWidth, form.field == addHabitDays),
	}
	if form.errorText != "" {
		lines = append(lines, errorStyle().Render(form.errorText))
	}

	return renderModal(modalWidth, lines)
}

func (form addHabitForm) appendRune(r rune) addHabitForm {
	switch form.field {
	case addHabitName:
		if len([]rune(form.name)) >= maxHabitNameLength {
			form.errorText = habitNameLengthError()
			return form
		}
		form.name += string(r)
	case addHabitDays:
		if weekday, ok := weekdayFromNumber(r); ok {
			form.weekdays[weekday] = !form.weekdays[weekday]
			form.dayCursor = dayCursorForWeekday(weekday)
		} else {
			form.errorText = "Days use 1-7 or Space"
		}
	}
	return form
}

func (form addHabitForm) appendRunes(runes []rune) addHabitForm {
	for _, r := range runes {
		form = form.appendRune(r)
	}
	return form
}

func (form addHabitForm) moveDayCursor(distance int) addHabitForm {
	if form.field == addHabitDays {
		form.dayCursor = (form.dayCursor + distance + len(habitDayPickerOrder)) % len(habitDayPickerOrder)
	}
	return form
}

func (form addHabitForm) deleteLastRune() addHabitForm {
	if form.field == addHabitName {
		form.name = trimLastRune(form.name)
	}
	return form
}

func (form addHabitForm) clearField() addHabitForm {
	if form.field == addHabitName {
		form.name = ""
	}
	return form
}

func (form addHabitForm) toggleSelectedDay() addHabitForm {
	if form.field != addHabitDays {
		return form
	}
	weekday := habitDayPickerOrder[form.dayCursor]
	form.weekdays[weekday] = !form.weekdays[weekday]
	return form
}

func selectedWeekdays(weekdays [7]bool) []int {
	selected := make([]int, 0, len(weekdays))
	for i, enabled := range weekdays {
		if enabled {
			selected = append(selected, i)
		}
	}
	return selected
}
