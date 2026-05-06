package habits

import "time"

var currentDate = func() time.Time {
	return time.Now().In(time.Local).Truncate(24 * time.Hour)
}

type Frequency string

const (
	FrequencyDaily  Frequency = "daily"
	FrequencyWeekly Frequency = "weekly"
	FrequencyCustom Frequency = "custom"
)

type Model struct {
	store             *Store
	habits            []HabitStatus
	selected          int
	selectedDayOffset int
	err               error
}

type HabitStatus struct {
	Habit     Habit
	Logs      map[string]HabitLogStatus
	DueToday  bool
	DoneToday bool
	Streak    int
	History   []DayStatus
}

type DayStatus struct {
	Date   time.Time
	Due    bool
	Status HabitLogStatus
}

const maxHabitHistoryDays = 30
