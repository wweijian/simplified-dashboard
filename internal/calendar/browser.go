package calendar

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

type SystemOpener struct{}

func (SystemOpener) Open(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("start browser: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("open browser: %w", err)
		}
	case <-time.After(500 * time.Millisecond):
	}
	return nil
}
