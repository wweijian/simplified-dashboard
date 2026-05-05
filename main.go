package main

import (
	"fmt"
	"os"

	bbt "github.com/charmbracelet/bubbletea"
	"simplified-dashboard/internal/db"
	"simplified-dashboard/internal/ui"
)

func main() {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	app := bbt.NewProgram(ui.New(database), bbt.WithAltScreen())
	if _, err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
