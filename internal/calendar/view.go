package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (model Model) summaryView(width, height int) string {
	if height <= 0 {
		return ""
	}
	if model.loading {
		return renderWrappedLines([]string{"Loading calendar..."}, width, height)
	}
	if model.err != nil {
		return renderWrappedLines([]string{"Calendar unavailable"}, width, height)
	}

	now := model.now()
	events := upcomingEvents(model.events, now)
	if len(events) == 0 {
		return "No upcoming events"
	}

	lines := make([]string, 0, height)
	for _, event := range events {
		lines = appendWrapped(lines, summaryEventRow(event, now, width), width, height)
		if len(lines) >= height {
			break
		}
	}
	return strings.Join(lines, "\n")
}

func (model Model) expandedView(width, height int) string {
	if height <= 0 {
		return ""
	}
	if model.loading {
		return renderWrappedLines([]string{"Loading calendar..."}, width, height)
	}
	if model.err != nil {
		return renderWrappedLines([]string{"Error loading calendar:", model.err.Error()}, width, height)
	}

	var lines []string
	lines = append(lines, model.headerLine(width), "")
	switch model.viewMode {
	case ViewWeek:
		lines = append(lines, model.weekLines(width)...)
	case ViewMonth:
		lines = append(lines, model.monthLines(width)...)
	default:
		lines = append(lines, model.dayLines(width)...)
	}

	return renderWrappedLines(lines, width, height)
}

func (model Model) headerLine(width int) string {
	label := "Day"
	rangeText := model.selectedDate.Format("Mon Jan 02, 2006")
	switch model.viewMode {
	case ViewWeek:
		label = "Week"
		start := weekStart(model.selectedDate)
		end := start.AddDate(0, 0, 6)
		rangeText = fmt.Sprintf("%s - %s", start.Format("Jan 02"), end.Format("Jan 02, 2006"))
	case ViewMonth:
		label = "Month"
		rangeText = monthStart(model.selectedDate).Format("January 2006")
	}

	controls := "  d/w/m view  left/right move  up/down select  enter/e open  a create  r refresh"
	return label + "  " + rangeText + controls
}

func (model Model) dayLines(width int) []string {
	dayStart, dayEnd := model.visibleRange()
	events := eventsInRange(model.events, dayStart, dayEnd)
	if len(events) == 0 {
		return []string{"No events"}
	}

	var lines []string
	allDayEvents, timedEvents := splitAllDay(events)
	for _, event := range allDayEvents {
		lines = append(lines, model.expandedEventRow(event, width))
	}

	var lastEnd time.Time
	for _, event := range timedEvents {
		if !lastEnd.IsZero() && event.Start.After(lastEnd.Add(29*time.Minute)) {
			lines = append(lines, fmt.Sprintf("   %s-%s free", lastEnd.Format("15:04"), event.Start.Format("15:04")))
		}
		lines = append(lines, model.expandedEventRow(event, width))
		if event.End.After(lastEnd) {
			lastEnd = event.End
		}
	}
	return lines
}

func (model Model) weekLines(width int) []string {
	start := weekStart(model.selectedDate)
	var lines []string
	for offset := 0; offset < 7; offset++ {
		day := start.AddDate(0, 0, offset)
		dayEvents := eventsInRange(model.events, day, day.AddDate(0, 0, 1))
		lines = append(lines, day.Format("Mon Jan 02"))
		if len(dayEvents) == 0 {
			lines = append(lines, "   No events")
		}
		for _, event := range dayEvents {
			lines = append(lines, model.expandedEventRow(event, width))
		}
		if offset < 6 {
			lines = append(lines, "")
		}
	}
	return lines
}

func (model Model) monthLines(width int) []string {
	lines := []string{monthGrid(model.events, model.selectedDate, width), ""}

	dayEvents := eventsInRange(model.events, dayStart(model.selectedDate), dayStart(model.selectedDate).AddDate(0, 0, 1))
	lines = append(lines, model.selectedDate.Format("Mon Jan 02"))
	if len(dayEvents) == 0 {
		lines = append(lines, "   No events")
		return lines
	}

	for _, event := range dayEvents {
		lines = append(lines, model.expandedEventRow(event, width))
	}
	return lines
}

func monthGrid(events []Event, selectedDate time.Time, width int) string {
	month := monthStart(selectedDate)
	gridStart := weekStart(month)
	cellWidth := (width - 6) / 7
	if cellWidth < 6 {
		cellWidth = 6
	}
	if cellWidth > 14 {
		cellWidth = 14
	}

	lines := []string{"Mon    Tue    Wed    Thu    Fri    Sat    Sun"}
	for week := 0; week < 6; week++ {
		var cells []string
		for day := 0; day < 7; day++ {
			date := gridStart.AddDate(0, 0, week*7+day)
			count := len(eventsInRange(events, date, date.AddDate(0, 0, 1)))
			label := fmt.Sprintf("%2d", date.Day())
			if date.Month() != month.Month() {
				label = " " + strings.TrimSpace(label)
			}
			if sameDay(date, selectedDate) {
				label = "[" + strings.TrimSpace(label) + "]"
			}
			if count > 0 {
				label = fmt.Sprintf("%s %d", label, count)
			}
			cells = append(cells, fitCell(label, cellWidth))
		}
		lines = append(lines, strings.Join(cells, " "))
	}
	return strings.Join(lines, "\n")
}

func (model Model) expandedEventRow(event Event, width int) string {
	cursor := "  "
	if selected, ok := model.selectedEvent(); ok && selected.ID == event.ID && selected.CalendarID == event.CalendarID {
		cursor = "> "
	}

	row := cursor + eventTimeRange(event) + "  " + event.Title
	if event.Location != "" {
		row += "  @ " + event.Location
	}
	label := event.CalendarLabel
	if label != "" {
		row += "  [" + label + "]"
	}
	if !event.AllDay {
		row += "  " + formatDuration(eventDuration(event))
	}
	return row
}

func summaryEventRow(event Event, now time.Time, width int) string {
	row := summaryTimeLabel(event, now) + "  " + event.Title
	label := event.CalendarLabel
	if label != "" {
		row += "  [" + label + "]"
	}
	if event.Location != "" && displayWidth(row)+displayWidth(event.Location)+5 <= width {
		row += "  @ " + event.Location
	}
	return row
}

func upcomingEvents(events []Event, now time.Time) []Event {
	var upcoming []Event
	for _, event := range events {
		if event.End.After(now) {
			upcoming = append(upcoming, event)
		}
	}
	sortEvents(upcoming)
	return upcoming
}

func splitAllDay(events []Event) ([]Event, []Event) {
	var allDayEvents []Event
	var timedEvents []Event
	for _, event := range events {
		if event.AllDay {
			allDayEvents = append(allDayEvents, event)
		} else {
			timedEvents = append(timedEvents, event)
		}
	}
	return allDayEvents, timedEvents
}

func summaryTimeLabel(event Event, now time.Time) string {
	if event.AllDay {
		if sameDay(event.Start, now) || event.Start.Before(now) && event.End.After(now) {
			return event.Start.Format("Mon Jan 02 all-day")
		}
		return event.Start.Format("Mon Jan 02 all-day")
	}
	if event.Start.Before(now) && event.End.After(now) {
		return event.Start.Format("Mon Jan 02 now")
	}
	return event.Start.Format("Mon Jan 02 15:04")
}

func eventTimeRange(event Event) string {
	if event.AllDay {
		return "all-day"
	}
	return event.Start.Format("15:04") + "-" + event.End.Format("15:04")
}

func formatDuration(duration time.Duration) string {
	minutes := int(duration.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	remainder := minutes % 60
	if remainder == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remainder)
}

func renderWrappedLines(lines []string, width, height int) string {
	if height <= 0 {
		return ""
	}
	wrapped := make([]string, 0, height)
	for _, line := range lines {
		wrapped = appendWrapped(wrapped, line, width, height)
		if len(wrapped) >= height {
			break
		}
	}
	return strings.Join(wrapped, "\n")
}

func appendWrapped(lines []string, line string, width, height int) []string {
	for _, wrapped := range wrapLine(line, width) {
		if len(lines) >= height {
			return lines
		}
		lines = append(lines, wrapped)
	}
	return lines
}

func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if displayWidth(line) <= width {
		return splitEmbeddedLines(line)
	}

	var wrapped []string
	for _, segment := range splitEmbeddedLines(line) {
		wrapped = append(wrapped, wrapSegment(segment, width)...)
	}
	return wrapped
}

func splitEmbeddedLines(line string) []string {
	parts := strings.Split(line, "\n")
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}

func wrapSegment(line string, width int) []string {
	if line == "" || displayWidth(line) <= width {
		return []string{line}
	}

	indent := leadingWhitespace(line)
	continuationIndent := indent
	if displayWidth(continuationIndent) >= width {
		continuationIndent = ""
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var wrapped []string
	current := ""
	for _, word := range words {
		if current == "" {
			current = indent + word
			if displayWidth(current) > width {
				chunks := splitLongWord(strings.TrimSpace(current), width)
				wrapped = append(wrapped, chunks[:len(chunks)-1]...)
				current = continuationIndent + chunks[len(chunks)-1]
			}
			continue
		}

		candidate := current + " " + word
		if displayWidth(candidate) <= width {
			current = candidate
			continue
		}

		wrapped = append(wrapped, current)
		current = continuationIndent + word
		if displayWidth(current) > width {
			chunks := splitLongWord(strings.TrimSpace(current), width)
			wrapped = append(wrapped, chunks[:len(chunks)-1]...)
			current = continuationIndent + chunks[len(chunks)-1]
		}
	}
	if current != "" {
		wrapped = append(wrapped, current)
	}
	return wrapped
}

func splitLongWord(word string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	var chunks []string
	var current []rune
	for _, char := range word {
		next := string(append(current, char))
		if len(current) > 0 && displayWidth(next) > width {
			chunks = append(chunks, string(current))
			current = []rune{char}
			continue
		}
		current = append(current, char)
	}
	if len(current) > 0 {
		chunks = append(chunks, string(current))
	}
	return chunks
}

func leadingWhitespace(line string) string {
	var builder strings.Builder
	for _, char := range line {
		if char != ' ' && char != '\t' {
			break
		}
		builder.WriteRune(char)
	}
	return builder.String()
}

func fitCell(value string, width int) string {
	if displayWidth(value) > width {
		return clipLine(value, width)
	}
	return value + strings.Repeat(" ", width-displayWidth(value))
}

func clipLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	var clipped []rune
	for _, char := range line {
		next := string(append(clipped, char))
		if displayWidth(next) > width {
			break
		}
		clipped = append(clipped, char)
	}
	return string(clipped)
}

func displayWidth(value string) int {
	return lipgloss.Width(value)
}

func sameDay(left, right time.Time) bool {
	left = left.In(time.Local)
	right = right.In(time.Local)
	return left.Year() == right.Year() && left.YearDay() == right.YearDay()
}
