package habits

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestSummaryViewShowsDueTodayHabits(t *testing.T) {
	withCurrentDate(t, "2026-05-06")
	store, _ := openTestStore(t)
	seedHabit(t, store, "Exercise")

	model := New(store)

	got := model.SummaryView(40, 5, false)
	if !strings.Contains(got, "[ ] Exercise") {
		t.Fatalf("expected due habit in summary, got:\n%s", got)
	}
}

func TestCustomHabitDueOnlyOnConfiguredWeekdays(t *testing.T) {
	withCurrentDate(t, "2026-05-06") // Wednesday

	habit := Habit{
		Name:          "Run",
		Frequency:     string(FrequencyCustom),
		FrequencyDays: sql.NullString{String: "[1,3,5]", Valid: true},
	}
	if !habitDueOn(habit, currentDate()) {
		t.Fatalf("expected custom habit to be due on Wednesday")
	}

	habit.FrequencyDays = sql.NullString{String: "[2,4]", Valid: true}
	if habitDueOn(habit, currentDate()) {
		t.Fatalf("expected custom habit not to be due on Wednesday")
	}
}

func TestModelToggleSelectedHabitUpdatesTodayCompletion(t *testing.T) {
	withCurrentDate(t, "2026-05-06")
	store, _ := openTestStore(t)
	seedHabit(t, store, "Read")

	model := New(store)
	model = model.Update(" ")

	if len(model.habits) != 1 {
		t.Fatalf("expected 1 habit, got %d", len(model.habits))
	}
	if !model.habits[0].DoneToday {
		t.Fatalf("expected selected habit to be completed today")
	}
}

func TestModelSpaceCyclesSelectedDayStatus(t *testing.T) {
	withCurrentDate(t, "2026-05-06")
	store, _ := openTestStore(t)
	seedHabit(t, store, "Read")

	model := New(store)

	model = model.Update(" ")
	if model.habits[0].History[0].Status != LogComplete {
		t.Fatalf("expected first cycle to complete today, got %q", model.habits[0].History[0].Status)
	}

	model = model.Update(" ")
	if model.habits[0].History[0].Status != LogSkipped {
		t.Fatalf("expected second cycle to skip today, got %q", model.habits[0].History[0].Status)
	}

	model = model.Update(" ")
	if model.habits[0].History[0].Status != LogIncomplete {
		t.Fatalf("expected third cycle to mark today incomplete, got %q", model.habits[0].History[0].Status)
	}
}

func TestModelCanMarkPastDaySkipped(t *testing.T) {
	withCurrentDate(t, "2026-05-06")
	store, _ := openTestStore(t)
	seedHabitCreatedAt(t, store, "Read", "2026-05-01")

	model := New(store)
	model = model.Update("left")
	model = model.Update("s")

	if model.selectedDayOffset != 1 {
		t.Fatalf("expected selected day offset 1, got %d", model.selectedDayOffset)
	}
	if model.habits[0].History[1].Status != LogSkipped {
		t.Fatalf("expected yesterday to be skipped, got %q", model.habits[0].History[1].Status)
	}
}

func TestDaySelectionWrapsBetweenTodayAndCreationDate(t *testing.T) {
	withCurrentDate(t, "2026-05-06")
	store, _ := openTestStore(t)
	seedHabitCreatedAt(t, store, "Read", "2026-05-04")

	model := New(store)
	model = model.Update("right")

	if model.selectedDayOffset != 2 {
		t.Fatalf("expected right from today to wrap to creation date offset 2, got %d", model.selectedDayOffset)
	}

	model = model.Update("left")
	if model.selectedDayOffset != 0 {
		t.Fatalf("expected left from creation date to wrap to today, got %d", model.selectedDayOffset)
	}
}

func TestExpandedViewSeparatesHabitTitleAndHistory(t *testing.T) {
	withCurrentDate(t, "2026-05-06")
	store, _ := openTestStore(t)
	seedHabit(t, store, "Exercise")

	model := New(store)
	model = model.Update("c")

	got := model.ExpandedView(80, 24)
	if !strings.Contains(got, "> [x] Exercise") {
		t.Fatalf("expected title line with selected habit, got:\n%s", got)
	}
	if !strings.Contains(got, "every day") || !strings.Contains(got, "Wed May 06: complete") {
		t.Fatalf("expected meta line with schedule and selected status, got:\n%s", got)
	}
	if !strings.Contains(got, "days  ") || !strings.Contains(got, "log   ") || !strings.Contains(got, "[●]") {
		t.Fatalf("expected separate days/log history lines, got:\n%s", got)
	}
}

func TestModelCreateHabitSelectsCreatedHabit(t *testing.T) {
	store, _ := openTestStore(t)
	model := New(store)

	model = model.Create(CreateHabitInput{
		Name:      "Journal",
		Frequency: string(FrequencyDaily),
	})

	habit, ok := model.SelectedHabit()
	if !ok {
		t.Fatalf("expected created habit to be selected")
	}
	if habit.Name != "Journal" {
		t.Fatalf("expected selected habit to be Journal, got %q", habit.Name)
	}
}

func TestModelUpdateHabitPreservesSelection(t *testing.T) {
	store, _ := openTestStore(t)
	seedHabit(t, store, "Draft")

	model := New(store)
	selectedHabit, ok := model.SelectedHabit()
	if !ok {
		t.Fatalf("expected selected habit")
	}

	model = model.UpdateHabit(selectedHabit.ID, UpdateHabitInput{
		Name:      "Final",
		Frequency: string(FrequencyWeekly),
	})

	habit, ok := model.SelectedHabit()
	if !ok {
		t.Fatalf("expected selected habit after update")
	}
	if habit.ID != selectedHabit.ID || habit.Name != "Final" {
		t.Fatalf("unexpected selected habit after update: %#v", habit)
	}
}

func TestModelDeleteHabitRemovesHabit(t *testing.T) {
	store, _ := openTestStore(t)
	seedHabit(t, store, "Delete me")

	model := New(store)
	selectedHabit, ok := model.SelectedHabit()
	if !ok {
		t.Fatalf("expected selected habit")
	}

	model = model.DeleteHabit(selectedHabit.ID)

	if _, ok := model.SelectedHabit(); ok {
		t.Fatalf("expected no selected habit after delete")
	}
}

func TestHabitStreakCountsConsecutiveDueCompletions(t *testing.T) {
	habit := Habit{
		Name:      "Meditate",
		Frequency: string(FrequencyDaily),
	}
	logs := map[string]HabitLogStatus{
		"2026-05-04": LogComplete,
		"2026-05-05": LogComplete,
		"2026-05-06": LogComplete,
	}

	streak := habitStreak(habit, logs, mustParseDate(t, "2026-05-06"))
	if streak != 3 {
		t.Fatalf("expected 3-day streak, got %d", streak)
	}
}

func TestHabitStreakSkipsSkippedDaysWithoutBreaking(t *testing.T) {
	habit := Habit{
		Name:      "Meditate",
		Frequency: string(FrequencyDaily),
	}
	logs := map[string]HabitLogStatus{
		"2026-05-04": LogComplete,
		"2026-05-05": LogSkipped,
		"2026-05-06": LogComplete,
	}

	streak := habitStreak(habit, logs, mustParseDate(t, "2026-05-06"))
	if streak != 2 {
		t.Fatalf("expected skipped day not to break 2-day streak, got %d", streak)
	}
}

func TestHabitStreakStopsWhenCustomHabitHasNoDueDays(t *testing.T) {
	habit := Habit{
		Name:          "Rest",
		Frequency:     string(FrequencyCustom),
		FrequencyDays: sql.NullString{String: "[]", Valid: true},
	}

	streak := habitStreak(habit, map[string]HabitLogStatus{}, mustParseDate(t, "2026-05-06"))
	if streak != 0 {
		t.Fatalf("expected empty custom schedule to have 0 streak, got %d", streak)
	}
}

func TestHistoryDayCountCapsAtThirty(t *testing.T) {
	if got := historyDayCount(500); got != 30 {
		t.Fatalf("expected wide view to cap at 30 days, got %d", got)
	}
}

func TestHabitHistoryStopsAtCreationDate(t *testing.T) {
	habit := Habit{
		Name:      "New habit",
		Frequency: string(FrequencyDaily),
		CreatedAt: "2026-05-04",
	}

	history := habitHistory(habit, map[string]HabitLogStatus{}, mustParseDate(t, "2026-05-06"))
	if len(history) != 3 {
		t.Fatalf("expected history from creation date through today, got %d days", len(history))
	}
	if got := dateString(history[len(history)-1].Date); got != "2026-05-04" {
		t.Fatalf("expected oldest history date to be creation date, got %s", got)
	}
}

func withCurrentDate(t *testing.T, date string) {
	t.Helper()

	originalCurrentDate := currentDate
	parsedDate := mustParseDate(t, date)
	currentDate = func() time.Time {
		return parsedDate
	}

	t.Cleanup(func() {
		currentDate = originalCurrentDate
	})
}

func mustParseDate(t *testing.T, date string) time.Time {
	t.Helper()

	parsedDate, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		t.Fatalf("parse date fixture: %v", err)
	}
	return parsedDate
}
