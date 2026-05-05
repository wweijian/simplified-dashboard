package tasks

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestSummaryViewEmpty(t *testing.T) {
	model := Model{}

	got := model.SummaryView(40, 5, false)

	if got != "No tasks" {
		t.Fatalf("expected %q, got %q", "No tasks", got)
	}
}

func TestSummaryViewAllDone(t *testing.T) {
	model := Model{
		tasks: []Task{
			{Title: "Done", Completed: true},
		},
	}

	got := model.SummaryView(40, 5, false)

	if got != "All done" {
		t.Fatalf("expected %q, got %q", "All done", got)
	}
}

func TestSummaryViewUsesTaskRowShape(t *testing.T) {
	withCurrentDate(t, "2026-05-05")

	model := Model{
		tasks: []Task{
			{
				Title:    "Pay bill",
				Priority: 2,
				DueDate:  sql.NullString{String: "2026-05-05", Valid: true},
			},
		},
	}

	got := model.SummaryView(80, 10, false)

	if !strings.Contains(got, "[ ] !!! today Pay bill") {
		t.Fatalf("expected aligned summary row, got:\n%s", got)
	}
}

func TestExpandedViewShowsSelectedTask(t *testing.T) {
	model := Model{
		selected: 1,
		tasks: []Task{
			{Title: "First", Priority: 1},
			{Title: "Second", Priority: 2},
		},
	}

	got := model.ExpandedView(80, 10)

	if !strings.Contains(got, "  [ ] !!  First") {
		t.Fatalf("expected unselected first task, got:\n%s", got)
	}

	if !strings.Contains(got, "> [ ] !!! Second") {
		t.Fatalf("expected selected second task, got:\n%s", got)
	}
}

func TestExpandedViewShowsDueDate(t *testing.T) {
	withCurrentDate(t, "2026-05-05")

	model := Model{
		tasks: []Task{
			{
				Title:    "Pay bill",
				Priority: 1,
				DueDate:  sql.NullString{String: "2026-05-05", Valid: true},
			},
		},
	}

	got := model.ExpandedView(80, 10)

	if !strings.Contains(got, "today") {
		t.Fatalf("expected relative due date, got:\n%s", got)
	}
}

func TestExpandedViewShowsDueDateInWeeks(t *testing.T) {
	withCurrentDate(t, "2026-05-05")

	model := Model{
		tasks: []Task{
			{
				Title:    "Plan trip",
				Priority: 1,
				DueDate:  sql.NullString{String: "2026-06-02", Valid: true},
			},
		},
	}

	got := model.ExpandedView(80, 10)

	if !strings.Contains(got, "in 4 weeks") {
		t.Fatalf("expected due date in weeks, got:\n%s", got)
	}
}

func TestExpandedViewShowsOverdueDate(t *testing.T) {
	withCurrentDate(t, "2026-05-05")

	model := Model{
		tasks: []Task{
			{
				Title:    "Send invoice",
				Priority: 1,
				DueDate:  sql.NullString{String: "2026-05-02", Valid: true},
			},
		},
	}

	got := model.ExpandedView(80, 10)

	if !strings.Contains(got, "3 days ago") {
		t.Fatalf("expected overdue date, got:\n%s", got)
	}
}

func TestExpandedViewStrikesCompletedTask(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	model := Model{
		tasks: []Task{
			{
				Title:     "Done task",
				Priority:  1,
				Completed: true,
			},
		},
	}

	got := model.ExpandedView(80, 10)

	if !strings.Contains(got, "\x1b[9m") {
		t.Fatalf("expected completed task to use strikethrough styling, got:\n%q", got)
	}
}

func withCurrentDate(t *testing.T, date string) {
	t.Helper()

	originalCurrentDate := currentDate
	parsedDate, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		t.Fatalf("parse current date fixture: %v", err)
	}

	currentDate = func() time.Time {
		return parsedDate
	}

	t.Cleanup(func() {
		currentDate = originalCurrentDate
	})
}
