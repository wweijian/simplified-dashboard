package ui

import (
	calendar 	"simplified-dashboard/internal/calendar"
	finance		"simplified-dashboard/internal/finance"
	tasks		"simplified-dashboard/internal/tasks"
	habits		"simplified-dashboard/internal/habits"
	
	bbt			"github.com/charmbracelet/bubbletea"
)

type Panel interface {
	SummaryView(width, height int, focused bool) string
	ExpandedView(width, height int) string
}

type Model struct {
	width		int
	height		int
	activePanel	int
	calendar	Panel
	tasks		Panel
	finance		Panel
	habits		Panel
}

func New() Model { 
	return Model{
		activePanel:	int(Calendar),
		calendar:		calendar.New(),
		tasks:			tasks.New(),
		finance:		finance.New(),
		habits:			habits.New(),
	}
}

func (model Model) Init() bbt.Cmd { return nil }

func (model Model) Update(msg bbt.Msg) (bbt.Model, bbt.Cmd) {
	switch msg := msg.(type) {
	case bbt.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
	case bbt.KeyMsg:
		if handler, ok := keyBindings[msg.String()]; ok {
			return handler(model);
		}
	}
	return model, nil
}

func (model Model) View () string {
	if model.width == 0 {
		return ""
	}
	return renderLayout(model)
}