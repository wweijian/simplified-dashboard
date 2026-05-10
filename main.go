package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	bbt "github.com/charmbracelet/bubbletea"
	"simplified-dashboard/internal/appenv"
	calendar "simplified-dashboard/internal/calendar"
	"simplified-dashboard/internal/db"
	"simplified-dashboard/internal/ui"
)

func main() {
	calendarAuth := flag.Bool("calendar-auth", false, "authorize Google Calendar access")
	flag.Parse()

	if err := appenv.Load(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "error loading .env: %v\n", err)
		os.Exit(1)
	}

	if *calendarAuth {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := calendar.AuthorizeDefault(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "calendar auth error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Google Calendar authorization complete.")
		return
	}

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
