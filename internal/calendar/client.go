package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GoogleEventClient struct {
	service *gcal.Service
}

func NewGoogleEventClient(ctx context.Context, config Config) (GoogleEventClient, error) {
	credentials, err := os.ReadFile(config.CredentialsPath)
	if err != nil {
		return GoogleEventClient{}, fmt.Errorf("read google credentials: %w", err)
	}

	oauthConfig, err := google.ConfigFromJSON(credentials, gcal.CalendarReadonlyScope)
	if err != nil {
		return GoogleEventClient{}, fmt.Errorf("parse google credentials: %w", err)
	}

	token, err := tokenFromFile(config.TokenPath)
	if err != nil {
		return GoogleEventClient{}, fmt.Errorf("read google token: %w; run go run . --calendar-auth", err)
	}

	httpClient := oauthConfig.Client(ctx, token)
	service, err := gcal.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return GoogleEventClient{}, fmt.Errorf("create google calendar service: %w", err)
	}

	return GoogleEventClient{service: service}, nil
}

func (client GoogleEventClient) ListEvents(ctx context.Context, calendar CalendarConfig, timeMin, timeMax time.Time) ([]Event, error) {
	response, err := client.service.Events.List(calendar.ID).
		Context(ctx).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		OrderBy("startTime").
		MaxResults(2500).
		Do()
	if err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(response.Items))
	for _, item := range response.Items {
		event, ok := googleEventToEvent(calendar, item)
		if ok {
			events = append(events, event)
		}
	}
	return events, nil
}

func googleEventToEvent(calendar CalendarConfig, item *gcal.Event) (Event, bool) {
	if item == nil || item.Status == "cancelled" {
		return Event{}, false
	}

	start, allDay, ok := parseGoogleEventTime(item.Start)
	if !ok {
		return Event{}, false
	}
	end, _, ok := parseGoogleEventTime(item.End)
	if !ok || !end.After(start) {
		if allDay {
			end = start.AddDate(0, 0, 1)
		} else {
			end = start.Add(time.Hour)
		}
	}

	title := item.Summary
	if title == "" {
		title = "(untitled)"
	}

	return Event{
		ID:            item.Id,
		CalendarID:    calendar.ID,
		CalendarLabel: calendar.DisplayLabel(),
		CalendarColor: calendar.Color,
		Title:         title,
		Start:         start,
		End:           end,
		AllDay:        allDay,
		Location:      item.Location,
		Description:   item.Description,
		HTMLLink:      item.HtmlLink,
	}, true
}

func parseGoogleEventTime(value *gcal.EventDateTime) (time.Time, bool, bool) {
	if value == nil {
		return time.Time{}, false, false
	}
	if value.Date != "" {
		parsed, err := time.ParseInLocation("2006-01-02", value.Date, time.Local)
		return parsed, true, err == nil
	}
	if value.DateTime == "" {
		return time.Time{}, false, false
	}
	parsed, err := time.Parse(time.RFC3339, value.DateTime)
	if err != nil {
		return time.Time{}, false, false
	}
	return parsed.In(time.Local), false, true
}

func tokenFromFile(path string) (*oauth2.Token, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token := &oauth2.Token{}
	if err := json.NewDecoder(file).Decode(token); err != nil {
		return nil, err
	}
	return token, nil
}
