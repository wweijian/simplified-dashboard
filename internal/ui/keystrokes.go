package ui

import (
	bbt			"github.com/charmbracelet/bubbletea"
)

type keyHandler func(Model) (Model, bbt.Cmd)

var keyBindings = map[string]keyHandler {
	"q":			handleQuit,
	"ctrl+c":		handleQuit,
	"1":			handleJumpToPanel(int(Calendar)),
	"2":			handleJumpToPanel(int(Tasks)),
	"3":			handleJumpToPanel(int(Finance)),
	"4":			handleJumpToPanel(int(Habits)),
	"tab":			handleNextPanel,
	"shift+tab":	handlePrevPanel,
}

func handleQuit(model Model) (Model, bbt.Cmd) {
	return model, bbt.Quit
}

func handleNextPanel(model Model) (Model, bbt.Cmd) {
	model.activePanel = model.activePanel%4 + 1
	return model, nil
}

func handlePrevPanel(model Model) (Model, bbt.Cmd) {
	model.activePanel = (model.activePanel - 2 + 4) % 4 + 1
	return model, nil
}

func handleJumpToPanel(number int) keyHandler {
	return func(model Model) (Model, bbt.Cmd) {
		model.activePanel = number
		return model, nil
	}
}
