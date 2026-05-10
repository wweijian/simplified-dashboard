package calendar

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type ViewMode int

const (
	ViewDay ViewMode = iota
	ViewWeek
	ViewMonth
)

type Event struct {
	ID            string
	CalendarID    string
	CalendarLabel string
	CalendarColor string
	Title         string
	Start         time.Time
	End           time.Time
	AllDay        bool
	Location      string
	Description   string
	HTMLLink      string
}

type EventClient interface {
	ListEvents(ctx context.Context, calendar CalendarConfig, timeMin, timeMax time.Time) ([]Event, error)
}

type Opener interface {
	Open(url string) error
}

type Options struct {
	Config Config
	Client EventClient
	Opener Opener
	Now    func() time.Time
}

type Model struct {
	config       Config
	client       EventClient
	opener       Opener
	now          func() time.Time
	viewMode     ViewMode
	selectedDate time.Time
	events       []Event
	selected     int
	loading      bool
	err          error
}

type LoadResult struct {
	Events []Event
	Err    error
}

func New() Model {
	config, err := LoadConfigFromEnv()
	model := NewWithOptions(Options{Config: config})
	if err != nil {
		model.err = err
		return model
	}

	client, err := NewGoogleEventClient(context.Background(), config)
	if err != nil {
		model.err = err
		return model
	}
	model.client = client
	return model
}

func NewWithOptions(options Options) Model {
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().In(time.Local) }
	}
	opener := options.Opener
	if opener == nil {
		opener = SystemOpener{}
	}

	return Model{
		config:       options.Config,
		client:       options.Client,
		opener:       opener,
		now:          now,
		viewMode:     ViewDay,
		selectedDate: dayStart(now()),
		selected:     0,
	}
}

func (model Model) StartLoading() Model {
	model.loading = true
	model.err = nil
	return model
}

func (model Model) LoadEvents(ctx context.Context) LoadResult {
	if model.client == nil {
		return LoadResult{Err: fmt.Errorf("google calendar is not configured; set GOOGLE_CALENDAR_IDS and run go run . --calendar-auth")}
	}
	if len(model.config.Calendars) == 0 {
		return LoadResult{Err: fmt.Errorf("no calendars configured; set GOOGLE_CALENDAR_IDS in .env")}
	}

	timeMin, timeMax := model.loadRange()
	var events []Event
	for _, calendar := range model.config.Calendars {
		calendarEvents, err := model.client.ListEvents(ctx, calendar, timeMin, timeMax)
		if err != nil {
			return LoadResult{Err: fmt.Errorf("load %s calendar events: %w", calendar.DisplayLabel(), err)}
		}
		events = append(events, calendarEvents...)
	}

	sortEvents(events)
	return LoadResult{Events: events}
}

func (model Model) ApplyLoadResult(result LoadResult) Model {
	model.loading = false
	if result.Err != nil {
		model.err = result.Err
		return model
	}
	model.events = result.Events
	model.err = nil
	return model.clampSelection()
}

func (model Model) SummaryView(width, height int, focused bool) string {
	return model.summaryView(width, height)
}

func (model Model) ExpandedView(width, height int) string {
	return model.expandedView(width, height)
}

func (model Model) loadRange() (time.Time, time.Time) {
	now := model.now()
	summaryMin := now.Add(-24 * time.Hour)
	summaryMax := now.AddDate(0, 0, 90)
	viewMin, viewMax := model.visibleRange()
	return minTime(summaryMin, viewMin), maxTime(summaryMax, viewMax)
}

func (model Model) visibleRange() (time.Time, time.Time) {
	switch model.viewMode {
	case ViewWeek:
		start := weekStart(model.selectedDate)
		return start, start.AddDate(0, 0, 7)
	case ViewMonth:
		start := monthStart(model.selectedDate)
		return start, start.AddDate(0, 1, 0)
	default:
		start := dayStart(model.selectedDate)
		return start, start.AddDate(0, 0, 1)
	}
}

func (model Model) visibleEvents() []Event {
	start, end := model.visibleRange()
	return eventsInRange(model.events, start, end)
}

func (model Model) selectedEvent() (Event, bool) {
	events := model.visibleEvents()
	if len(events) == 0 || model.selected < 0 || model.selected >= len(events) {
		return Event{}, false
	}
	return events[model.selected], true
}

func (model Model) clampSelection() Model {
	events := model.visibleEvents()
	if model.selected >= len(events) {
		model.selected = len(events) - 1
	}
	if model.selected < 0 {
		model.selected = 0
	}
	return model
}

func sortEvents(events []Event) {
	sort.SliceStable(events, func(i, j int) bool {
		left := events[i]
		right := events[j]
		if !left.Start.Equal(right.Start) {
			return left.Start.Before(right.Start)
		}
		if left.AllDay != right.AllDay {
			return left.AllDay
		}
		if left.CalendarLabel != right.CalendarLabel {
			return left.CalendarLabel < right.CalendarLabel
		}
		return left.Title < right.Title
	})
}

func eventsInRange(events []Event, start, end time.Time) []Event {
	var filtered []Event
	for _, event := range events {
		if event.End.After(start) && event.Start.Before(end) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func dayStart(value time.Time) time.Time {
	value = value.In(time.Local)
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.Local)
}

func weekStart(value time.Time) time.Time {
	start := dayStart(value)
	offset := (int(start.Weekday()) + 6) % 7
	return start.AddDate(0, 0, -offset)
}

func monthStart(value time.Time) time.Time {
	value = value.In(time.Local)
	return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.Local)
}

func minTime(left, right time.Time) time.Time {
	if left.Before(right) {
		return left
	}
	return right
}

func maxTime(left, right time.Time) time.Time {
	if left.After(right) {
		return left
	}
	return right
}
