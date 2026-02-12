package main

import (
	"fmt"
	"os"

	"sshh/internal/config"
	"sshh/internal/history"
	"sshh/internal/sshexec"
	"sshh/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Ensure config directory exists.
	dir, err := config.Dir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config dir: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	hist, err := history.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	// Direct connect mode: sshh <name>
	if len(os.Args) > 1 {
		name := os.Args[1]
		_, srv := cfg.FindByName(name)
		if srv == nil {
			fmt.Fprintf(os.Stderr, "Server %q not found\n", name)
			os.Exit(1)
		}
		_ = hist.Record(srv.Name)
		if err := sshexec.Connect(*srv); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// TUI mode.
	m := tui.NewModel(cfg, hist)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If a server was selected, connect after TUI exits.
	if fm, ok := finalModel.(tui.Model); ok && fm.ConnectTo != nil {
		_ = hist.Record(fm.ConnectTo.Name)
		if err := sshexec.Connect(*fm.ConnectTo); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
