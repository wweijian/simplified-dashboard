package ui 

import (
	bbt "github.com/charmbracelet/bubbletea"
)

type Model struct {}

func New() Model { return Model{} }

func (m Model) Init() bbt.Cmd { return nil }

func (m Model) Update(msg bbt.Msg) (bbt.Model, bbt.Cmd) {
	if key, ok := msg.(bbt.KeyMsg); ok && key.String() == "q" {
		return m, bbt.Quit
	}
	return m, nil
}

func (m Model) View () string {
	return "dashboard goes here, press q to exit"
}