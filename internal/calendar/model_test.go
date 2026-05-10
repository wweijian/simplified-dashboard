package calendar

import (
	"context"
	"strings"
	"testing"
	"time"
)

type fakeClient struct {
	events       map[string][]Event
	calls        int
	lastTimeMin  time.Time
	lastTimeMax  time.Time
	lastCalendar []string
}

func (client *fakeClient) ListEvents(ctx context.Context, calendar CalendarConfig, timeMin, timeMax time.Time) ([]Event, error) {
	client.calls++
	client.lastTimeMin = timeMin
	client.lastTimeMax = timeMax
	client.lastCalendar = append(client.lastCalendar, calendar.ID)
	events := client.events[calendar.ID]
	for i := range events {
		events[i].CalendarID = calendar.ID
		events[i].CalendarLabel = calendar.DisplayLabel()
		events[i].CalendarColor = calendar.Color
	}
	return events, nil
}

type fakeOpener struct {
	urls []string
}

func (opener *fakeOpener) Open(url string) error {
	opener.urls = append(opener.urls, url)
	return nil
}

func TestLoadConfigFromEnvFullPathsAndValidation(t *testing.T) {
	t.Setenv("GOOGLE_CALENDAR_CREDENTIALS_PATH", "~/dashboard/credentials.json")
	t.Setenv("GOOGLE_CALENDAR_TOKEN_PATH", "~/dashboard/calendar_token.json")
	t.Setenv("GOOGLE_CALENDAR_IDS", "primary, work@example.com")
	t.Setenv("GOOGLE_CALENDAR_LABELS", "Main, Work")
	t.Setenv("GOOGLE_CALENDAR_COLORS", "6, 4")

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv returned error: %v", err)
	}
	if config.CredentialsPath == "" || config.TokenPath == "" {
		t.Fatalf("expected full credential and token paths, got %#v", config)
	}
	if config.Calendars[0].DisplayLabel() != "Main" {
		t.Fatalf("expected label Main, got %q", config.Calendars[0].DisplayLabel())
	}
	if config.Calendars[1].ID != "work@example.com" || config.Calendars[1].Color != "4" {
		t.Fatalf("expected second calendar env values, got %#v", config.Calendars[1])
	}

	t.Setenv("GOOGLE_CALENDAR_CREDENTIALS_PATH", "")
	if _, err := LoadConfigFromEnv(); err == nil {
		t.Fatal("expected missing credentials path error")
	}

	t.Setenv("GOOGLE_CALENDAR_CREDENTIALS_PATH", "~/dashboard/credentials.json")
	t.Setenv("GOOGLE_CALENDAR_TOKEN_PATH", "")
	if _, err := LoadConfigFromEnv(); err == nil {
		t.Fatal("expected missing token path error")
	}

	t.Setenv("GOOGLE_CALENDAR_TOKEN_PATH", "~/dashboard/calendar_token.json")
	t.Setenv("GOOGLE_CALENDAR_IDS", "")
	if _, err := LoadConfigFromEnv(); err == nil {
		t.Fatal("expected missing calendar ids error")
	}
}

func TestVisibleRanges(t *testing.T) {
	now := mustTime(t, "2026-05-08T10:00:00+08:00")
	model := NewWithOptions(Options{Now: func() time.Time { return now }})

	start, end := model.visibleRange()
	assertDate(t, start, "2026-05-08")
	assertDate(t, end, "2026-05-09")

	model.viewMode = ViewWeek
	start, end = model.visibleRange()
	assertDate(t, start, "2026-05-04")
	assertDate(t, end, "2026-05-11")

	model.viewMode = ViewMonth
	start, end = model.visibleRange()
	assertDate(t, start, "2026-05-01")
	assertDate(t, end, "2026-06-01")
}

func TestRefreshMergesAndSortsEvents(t *testing.T) {
	now := mustTime(t, "2026-05-08T10:00:00+08:00")
	client := &fakeClient{events: map[string][]Event{
		"work": {
			event("b", "Budget", "2026-05-08T15:00:00+08:00", "2026-05-08T15:30:00+08:00"),
		},
		"home": {
			event("a", "Breakfast", "2026-05-08T08:00:00+08:00", "2026-05-08T08:30:00+08:00"),
		},
	}}
	model := NewWithOptions(Options{
		Config: testConfig(),
		Client: client,
		Now:    func() time.Time { return now },
	})
	model = model.ApplyLoadResult(model.LoadEvents(context.Background()))

	if model.err != nil {
		t.Fatalf("Refresh returned error: %v", model.err)
	}
	if client.calls != 2 {
		t.Fatalf("expected two calendar calls, got %d", client.calls)
	}
	if got := []string{model.events[0].Title, model.events[1].Title}; strings.Join(got, ",") != "Breakfast,Budget" {
		t.Fatalf("events not sorted: %v", got)
	}
	if model.events[0].CalendarLabel != "Home" || model.events[1].CalendarLabel != "Work" {
		t.Fatalf("calendar labels not applied: %#v", model.events)
	}
}

func TestSummaryClipsUpcomingEventsByHeight(t *testing.T) {
	now := mustTime(t, "2026-05-08T10:00:00+08:00")
	model := NewWithOptions(Options{Now: func() time.Time { return now }})
	model.events = []Event{
		event("past", "Past", "2026-05-08T08:00:00+08:00", "2026-05-08T08:30:00+08:00"),
		event("one", "One", "2026-05-08T11:00:00+08:00", "2026-05-08T11:30:00+08:00"),
		event("two", "Two", "2026-05-08T12:00:00+08:00", "2026-05-08T12:30:00+08:00"),
		event("three", "Three", "2026-05-08T13:00:00+08:00", "2026-05-08T13:30:00+08:00"),
	}

	lines := strings.Split(model.SummaryView(80, 2, false), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two summary lines, got %d: %q", len(lines), lines)
	}
	if !strings.Contains(lines[0], "One") || !strings.Contains(lines[1], "Two") {
		t.Fatalf("expected first two upcoming events, got %q", lines)
	}
	if !strings.HasPrefix(lines[0], "Fri May 08 11:00") {
		t.Fatalf("expected summary row to include weekday, date, and time, got %q", lines[0])
	}
}

func TestSummaryWrapsLongEventRows(t *testing.T) {
	now := mustTime(t, "2026-05-08T10:00:00+08:00")
	model := NewWithOptions(Options{Now: func() time.Time { return now }})
	model.events = []Event{
		event("one", "Very long planning conversation", "2026-05-08T11:00:00+08:00", "2026-05-08T11:30:00+08:00"),
	}

	lines := strings.Split(model.SummaryView(18, 4, false), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapped summary, got %q", lines)
	}
	for _, line := range lines {
		if displayWidth(line) > 18 {
			t.Fatalf("line wider than view: %q", line)
		}
		if strings.Contains(line, "…") {
			t.Fatalf("expected wrapping without ellipsis, got %q", lines)
		}
	}
}

func TestLoadingView(t *testing.T) {
	model := NewWithOptions(Options{}).StartLoading()

	if got := model.SummaryView(80, 2, false); got != "Loading calendar..." {
		t.Fatalf("expected loading summary, got %q", got)
	}
	if got := model.ExpandedView(80, 2); got != "Loading calendar..." {
		t.Fatalf("expected loading expanded view, got %q", got)
	}
}

func TestUpdateKeysSwitchMoveRefreshAndOpen(t *testing.T) {
	now := mustTime(t, "2026-05-08T10:00:00+08:00")
	client := &fakeClient{events: map[string][]Event{
		"work": {
			eventWithLink("one", "One", "2026-05-15T11:00:00+08:00", "https://calendar.google.com/event?eid=one"),
			eventWithLink("two", "Two", "2026-05-15T12:00:00+08:00", "https://calendar.google.com/event?eid=two"),
		},
	}}
	opener := &fakeOpener{}
	model := NewWithOptions(Options{
		Config: Config{Calendars: []CalendarConfig{{ID: "work", Label: "Work"}}},
		Client: client,
		Opener: opener,
		Now:    func() time.Time { return now },
	})
	model = model.ApplyLoadResult(model.LoadEvents(context.Background()))

	model = model.Update("w")
	if model.viewMode != ViewWeek {
		t.Fatalf("expected week view")
	}
	model = model.Update("right")
	assertDate(t, model.selectedDate, "2026-05-15")
	if !model.loading {
		t.Fatalf("expected moving timeframe to mark calendar loading")
	}
	model = model.ApplyLoadResult(model.LoadEvents(context.Background()))

	model = model.Update("m")
	if model.viewMode != ViewMonth {
		t.Fatalf("expected month view")
	}
	model = model.Update("d")
	model = model.Update("down")
	model = model.Update("e")
	if len(opener.urls) != 1 || opener.urls[0] != "https://calendar.google.com/event?eid=two" {
		t.Fatalf("expected selected event to open, got %v", opener.urls)
	}

	model = model.Update("a")
	if len(opener.urls) != 2 || !strings.Contains(opener.urls[1], "calendar.google.com/calendar/render") {
		t.Fatalf("expected create URL to open, got %v", opener.urls)
	}
}

func testConfig() Config {
	return Config{Calendars: []CalendarConfig{
		{ID: "work", Label: "Work"},
		{ID: "home", Label: "Home"},
	}}
}

func event(id, title, start, end string) Event {
	return Event{
		ID:    id,
		Title: title,
		Start: mustTime(nil, start),
		End:   mustTime(nil, end),
	}
}

func eventWithLink(id, title, start, link string) Event {
	startTime := mustTime(nil, start)
	return Event{
		ID:       id,
		Title:    title,
		Start:    startTime,
		End:      startTime.Add(30 * time.Minute),
		HTMLLink: link,
	}
}

func mustTime(t *testing.T, value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		if t != nil {
			t.Fatalf("parse time %q: %v", value, err)
		}
		panic(err)
	}
	return parsed.In(time.Local)
}

func assertDate(t *testing.T, value time.Time, expected string) {
	t.Helper()
	if got := value.Format("2006-01-02"); got != expected {
		t.Fatalf("expected date %s, got %s", expected, got)
	}
}
