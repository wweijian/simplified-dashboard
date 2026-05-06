package ui

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"simplified-dashboard/internal/habits"
)

func (form *addHabitForm) validateName(name string) bool {
	if name == "" {
		form.errorText = "Name is required"
		return false
	}
	if len([]rune(name)) > maxHabitNameLength {
		form.errorText = habitNameLengthError()
		return false
	}
	return true
}

func (form *addHabitForm) validateWeekdays(weekdays []int) bool {
	if len(weekdays) == 0 {
		form.errorText = "Choose at least one day"
		return false
	}
	return true
}

func habitInputFromDays(name string, weekdays []int) (habits.CreateHabitInput, error) {
	input := habits.CreateHabitInput{Name: name, Frequency: string(habits.FrequencyDaily)}
	if len(weekdays) == 7 {
		return input, nil
	}

	encoded, err := json.Marshal(weekdays)
	if err != nil {
		return input, fmt.Errorf("could not save days")
	}
	input.Frequency = string(habits.FrequencyCustom)
	input.FrequencyDays = sql.NullString{String: string(encoded), Valid: true}
	return input, nil
}

func habitNameLengthError() string {
	return fmt.Sprintf("Name must be %d characters or fewer", maxHabitNameLength)
}
