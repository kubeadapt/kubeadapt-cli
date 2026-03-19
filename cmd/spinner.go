package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type spinner struct {
	message string
	done    chan struct{}
}

func newSpinner(message string) *spinner {
	return &spinner{message: message, done: make(chan struct{})}
}

func (s *spinner) start() {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return
	}
	go func() {
		i := 0
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))
		for {
			select {
			case <-s.done:
				fmt.Fprint(os.Stderr, "\r\033[K")
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s %s", style.Render(spinnerFrames[i%len(spinnerFrames)]), s.message)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *spinner) stop() {
	close(s.done)
}
