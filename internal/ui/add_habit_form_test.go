package ui

import (
	"strings"
	"testing"

	"simplified-dashboard/internal/habits"

	bbt "github.com/charmbracelet/bubbletea"
)

func TestAddHabitFormSubmitsDailyHabit(t *testing.T) {
	form := newAddHabitForm()
	form.name = "Exercise"

	_, input, ok := form.Submit()
	if !ok {
		t.Fatalf("expected valid habit form to submit")
	}

	if input.Name != "Exercise" || input.Frequency != string(habits.FrequencyDaily) || input.FrequencyDays.Valid {
		t.Fatalf("unexpected habit input: %#v", input)
	}
}

func TestAddHabitFormSubmitsCustomWeekdays(t *testing.T) {
	form := newAddHabitForm()
	form.name = "Run"
	form.weekdays = [7]bool{}
	form.weekdays[1] = true
	form.weekdays[3] = true
	form.weekdays[5] = true

	_, input, ok := form.Submit()
	if !ok {
		t.Fatalf("expected custom habit form to submit")
	}

	if !input.FrequencyDays.Valid || input.FrequencyDays.String != "[1,3,5]" {
		t.Fatalf("expected encoded weekdays [1,3,5], got %#v", input.FrequencyDays)
	}
}

func TestAddHabitFormRejectsCustomHabitWithoutWeekdays(t *testing.T) {
	form := newAddHabitForm()
	form.name = "Rest"
	form.weekdays = [7]bool{}

	form, _, ok := form.Submit()
	if ok {
		t.Fatalf("expected habit without days to be invalid")
	}
	if !strings.Contains(form.errorText, "day") {
		t.Fatalf("expected day validation error, got %q", form.errorText)
	}
}

func TestAddHabitFormTogglesDaysWithNumbersOneThroughSeven(t *testing.T) {
	form := newAddHabitForm()
	form.field = addHabitDays
	form.weekdays = [7]bool{}

	form = form.Update(bbt.KeyMsg{Type: bbt.KeyRunes, Runes: []rune("3")})

	if !form.weekdays[3] {
		t.Fatalf("expected Wednesday to be selected")
	}
}

func TestAddHabitFormTogglesSundayWithSeven(t *testing.T) {
	form := newAddHabitForm()
	form.field = addHabitDays
	form.weekdays = [7]bool{}

	form = form.Update(bbt.KeyMsg{Type: bbt.KeyRunes, Runes: []rune("7")})

	if !form.weekdays[0] {
		t.Fatalf("expected Sunday to be selected")
	}
}

func TestAddHabitFormCyclesDaysWithArrowsAndSpace(t *testing.T) {
	form := newAddHabitForm()
	form.field = addHabitDays
	form.weekdays = [7]bool{}

	form = form.Update(bbt.KeyMsg{Type: bbt.KeyRight})
	form = form.Update(bbt.KeyMsg{Type: bbt.KeySpace})

	if !form.weekdays[2] {
		t.Fatalf("expected Tuesday to be selected after right then space")
	}
}
