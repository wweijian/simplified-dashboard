package ui

import (
	"strings"
	"testing"
	"time"
)

func TestAddTaskFormFieldsDoNotWrapBorders(t *testing.T) {
	view := newAddTaskForm().View(100)

	for _, line := range strings.Split(view, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "─┐" || trimmed == "─┘" {
			t.Fatalf("expected field border not to wrap, got:\n%s", view)
		}
	}
}

func TestAddTaskFormRejectsLongTitle(t *testing.T) {
	form := newAddTaskForm()
	form.title = strings.Repeat("a", maxTaskTitleLength+1)
	form.dueDays = "7"

	form, _, ok := form.Submit(time.Now())
	if ok {
		t.Fatalf("expected long title to be rejected")
	}

	if form.errorText != titleLengthError() {
		t.Fatalf("expected long-title error, got %q", form.errorText)
	}
}

func TestAddTaskFormAcceptsMaxLengthTitle(t *testing.T) {
	form := newAddTaskForm()
	form.title = strings.Repeat("a", maxTaskTitleLength)
	form.dueDays = "7"

	_, _, ok := form.Submit(time.Now())
	if !ok {
		t.Fatalf("expected max-length title to submit")
	}
}

func TestAddTaskFormRejectsLongDescription(t *testing.T) {
	form := newAddTaskForm()
	form.title = "Task"
	form.description = strings.Repeat("a", maxTaskDescriptionChars+1)
	form.dueDays = "7"

	form, _, ok := form.Submit(time.Now())
	if ok {
		t.Fatalf("expected long description to be rejected")
	}

	if form.errorText != descriptionLengthError() {
		t.Fatalf("expected long-description error, got %q", form.errorText)
	}
}

func TestAddTaskFormAcceptsMaxLengthDescription(t *testing.T) {
	form := newAddTaskForm()
	form.title = "Task"
	form.description = strings.Repeat("a", maxTaskDescriptionChars)
	form.dueDays = "7"

	_, input, ok := form.Submit(time.Now())
	if !ok {
		t.Fatalf("expected max-length description to submit")
	}

	if !input.Description.Valid || input.Description.String != form.description {
		t.Fatalf("expected description to be stored, got %#v", input.Description)
	}
}

func TestAddTaskFormStoresBlankDescriptionAsNull(t *testing.T) {
	form := newAddTaskForm()
	form.title = "Task"
	form.dueDays = "7"

	_, input, ok := form.Submit(time.Now())
	if !ok {
		t.Fatalf("expected blank description to submit")
	}

	if input.Description.Valid {
		t.Fatalf("expected blank description to be null, got %#v", input.Description)
	}
}

func TestAddTaskFormDescriptionUsesFourVisibleLines(t *testing.T) {
	field := textAreaField("", 30, descriptionVisibleLines, true)
	lines := strings.Split(field, "\n")

	if len(lines) != descriptionVisibleLines+2 {
		t.Fatalf("expected description field height %d including border, got %d:\n%s",
			descriptionVisibleLines+2, len(lines), field)
	}
}

func TestAddTaskFormDescriptionWraps(t *testing.T) {
	form := newAddTaskForm()
	form.description = "alpha beta gamma delta epsilon zeta"

	view := form.View(42)

	if !strings.Contains(view, "alpha beta gamma") || !strings.Contains(view, "delta epsilon zeta") {
		t.Fatalf("expected wrapped description text, got:\n%s", view)
	}
}

func TestAddTaskFormDescriptionRendersAfterPriority(t *testing.T) {
	view := newAddTaskForm().View(100)

	priorityIndex := strings.Index(view, "Priority")
	descriptionIndex := strings.Index(view, "Description")
	if priorityIndex == -1 || descriptionIndex == -1 {
		t.Fatalf("expected priority and description fields, got:\n%s", view)
	}

	if descriptionIndex < priorityIndex {
		t.Fatalf("expected description after priority, got:\n%s", view)
	}
}

func TestAddTaskFormShowsFullMaxLengthTitle(t *testing.T) {
	form := newAddTaskForm()
	form.title = strings.Repeat("a", maxTaskTitleLength)

	view := form.View(100)

	if !strings.Contains(view, form.title) {
		t.Fatalf("expected max-length title to fit in form, got:\n%s", view)
	}

	if strings.Contains(view, "...") {
		t.Fatalf("expected max-length title not to be truncated, got:\n%s", view)
	}
}

func TestAddTaskFormRejectsDueDateBeyondOneYear(t *testing.T) {
	form := newAddTaskForm()
	form.title = "Plan"
	form.dueDays = "366"

	form, _, ok := form.Submit(time.Now())
	if ok {
		t.Fatalf("expected due date beyond one year to be rejected")
	}

	if form.errorText != "Due date must be within 365 days" {
		t.Fatalf("expected due-date range error, got %q", form.errorText)
	}
}

func TestAddTaskFormAcceptsDueDateAtOneYear(t *testing.T) {
	form := newAddTaskForm()
	form.title = "Plan"
	form.dueDays = "365"

	_, _, ok := form.Submit(time.Now())
	if !ok {
		t.Fatalf("expected due date at one year to submit")
	}
}
