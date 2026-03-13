package main

import (
	"fmt"
	"os"
	"sync"

	"sshh/internal/config"
	"sshh/internal/history"
	"sshh/internal/sshexec"
	"sshh/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time via -ldflags "-X main.version=v1.2.3"
var version = "dev"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("sshh", version)
		return
	}

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

	// Load all config files concurrently — they are independent.
	var (
		cfg       *config.Config
		hist      *history.History
		tunnelCfg *config.TunnelConfig
		settings  *config.Settings

		cfgErr, histErr, tunnelErr, settingsErr error

		wg sync.WaitGroup
	)
	wg.Add(4)
	go func() { defer wg.Done(); cfg, cfgErr = config.Load() }()
	go func() { defer wg.Done(); hist, histErr = history.Load() }()
	go func() { defer wg.Done(); tunnelCfg, tunnelErr = config.LoadTunnels() }()
	go func() { defer wg.Done(); settings, settingsErr = config.LoadSettings() }()
	wg.Wait()

	for _, e := range []struct {
		label string
		err   error
	}{
		{"loading config", cfgErr},
		{"loading history", histErr},
		{"loading tunnels", tunnelErr},
		{"loading settings", settingsErr},
	} {
		if e.err != nil {
			fmt.Fprintf(os.Stderr, "Error %s: %v\n", e.label, e.err)
			os.Exit(1)
		}
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
	m := tui.NewModel(cfg, tunnelCfg, hist, settings)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm, ok := finalModel.(tui.Model)
	if !ok {
		return
	}

	// If a server was selected, connect after TUI exits.
	if fm.ConnectTo != nil {
		_ = hist.Record(fm.ConnectTo.Name)
		if err := sshexec.Connect(*fm.ConnectTo); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// If a tunnel was selected, run it after TUI exits.
	if fm.RunTunnel != nil {
		if err := sshexec.RunTunnel(*fm.RunTunnel); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
