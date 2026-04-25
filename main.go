package main

import (
	"fmt"
	"os"

	bbt "github.com/charmbracelet/bubbletea"
	"simplified-dashboard/internal/ui"
	"simplified-dashboard/internal/db"
)

func main() {
	db.Open();
	app := bbt.NewProgram(ui.New(), bbt.WithAltScreen())
	if _, err := app.Run(); err != nil {
		fmt.Fprint(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
