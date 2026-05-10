package calendar

import (
	"fmt"
	"net/url"
	"time"
)

func (model Model) Update(key string) Model {
	switch key {
	case "d":
		model.viewMode = ViewDay
		return model.clampSelection()
	case "w":
		model.viewMode = ViewWeek
		return model.clampSelection()
	case "m":
		model.viewMode = ViewMonth
		return model.clampSelection()
	case "left":
		return model.moveTimeframe(-1)
	case "right":
		return model.moveTimeframe(1)
	case "up":
		return model.moveSelection(-1)
	case "down":
		return model.moveSelection(1)
	case "r":
		return model.StartLoading()
	case "enter", "e":
		return model.openSelectedEvent()
	case "a":
		return model.openCreateEvent()
	default:
		return model
	}
}

func (model Model) moveTimeframe(distance int) Model {
	switch model.viewMode {
	case ViewWeek:
		model.selectedDate = model.selectedDate.AddDate(0, 0, distance*7)
	case ViewMonth:
		model.selectedDate = model.selectedDate.AddDate(0, distance, 0)
	default:
		model.selectedDate = model.selectedDate.AddDate(0, 0, distance)
	}
	model.selectedDate = dayStart(model.selectedDate)
	model.selected = 0
	return model.StartLoading()
}

func (model Model) moveSelection(distance int) Model {
	events := model.visibleEvents()
	if len(events) == 0 {
		return model
	}
	nextSelection := model.selected + distance
	if nextSelection >= 0 && nextSelection < len(events) {
		model.selected = nextSelection
	}
	return model
}

func (model Model) openSelectedEvent() Model {
	event, ok := model.selectedEvent()
	if !ok || event.HTMLLink == "" {
		return model
	}
	if err := model.opener.Open(event.HTMLLink); err != nil {
		model.err = err
	}
	return model
}

func (model Model) openCreateEvent() Model {
	link := "https://calendar.google.com/calendar/render?action=TEMPLATE"
	if !model.selectedDate.IsZero() {
		date := model.selectedDate.Format("20060102")
		values := url.Values{}
		values.Set("action", "TEMPLATE")
		values.Set("dates", fmt.Sprintf("%s/%s", date, model.selectedDate.AddDate(0, 0, 1).Format("20060102")))
		link = "https://calendar.google.com/calendar/render?" + values.Encode()
	}
	if err := model.opener.Open(link); err != nil {
		model.err = err
	}
	return model
}

func eventDuration(event Event) time.Duration {
	return event.End.Sub(event.Start)
}
