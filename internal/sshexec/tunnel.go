package sshexec

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"sshh/internal/model"
)

// RunTunnel starts an SSH tunnel and blocks until it exits.
// Uses -N to skip remote command execution (port-forward only).
// Prints a connected banner only after SSH has been alive for a short window,
// so fast failures (connection refused, auth errors) surface as errors instead.
func RunTunnel(t model.Tunnel) error {
	sshBin, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	args := []string{"-N"}

	switch t.Type {
	case model.TunnelLocal:
		args = append(args, "-L",
			fmt.Sprintf("127.0.0.1:%d:%s:%d", t.LocalPort, t.RemoteHost, t.RemotePort))
	case model.TunnelRemote:
		args = append(args, "-R",
			fmt.Sprintf("%d:127.0.0.1:%d", t.RemotePort, t.LocalPort))
	case model.TunnelDynamic:
		args = append(args, "-D", fmt.Sprintf("127.0.0.1:%d", t.LocalPort))
	}

	if t.SSHPort != 0 && t.SSHPort != 22 {
		args = append(args, "-p", strconv.Itoa(t.SSHPort))
	}
	if t.SSHKey != "" {
		args = append(args, "-i", t.SSHKey)
	}

	// Force SSH to give up after 10 seconds if the host is unreachable.
	// Without this, the OS TCP timeout (60-90s) would apply instead.
	args = append(args, "-o", "ConnectTimeout=10")

	target := t.SSHHost
	if t.SSHUser != "" {
		target = t.SSHUser + "@" + t.SSHHost
	}
	args = append(args, target)

	cmd := exec.Command(sshBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n  Connecting tunnel %q...\n\n", t.Name)

	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for SSH to either exit quickly (failure) or stay alive (success).
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		// SSH exited before the window — connection failed.
		// SSH already printed the error to stderr.
		if err != nil {
			return fmt.Errorf("tunnel failed to connect")
		}
		return nil
	case <-time.After(12 * time.Second):
		// SSH is still alive — tunnel is up.
		fmt.Printf("  Tunnel %q connected\n", t.Name)
		switch t.Type {
		case model.TunnelLocal:
			fmt.Printf("  127.0.0.1:%d  →  %s:%d  (via %s)\n", t.LocalPort, t.RemoteHost, t.RemotePort, target)
		case model.TunnelRemote:
			fmt.Printf("  %s:%d  →  127.0.0.1:%d  (via %s)\n", target, t.RemotePort, t.LocalPort, target)
		case model.TunnelDynamic:
			fmt.Printf("  SOCKS proxy on 127.0.0.1:%d  (via %s)\n", t.LocalPort, target)
		}
		fmt.Printf("  Press Ctrl+C to disconnect\n\n")
	}

	// Block until the tunnel exits (Ctrl+C or server drop).
	err = <-done
	fmt.Printf("  Tunnel %q disconnected.\n\n", t.Name)

	// Ctrl+C / signal termination is normal — don't surface as an error.
	if isInterrupt(err) {
		return nil
	}
	return err
}

// isInterrupt reports whether the error is a normal signal-driven exit.
func isInterrupt(err error) bool {
	if err == nil {
		return false
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}
	// Exit code -1 means killed by a signal; 130 = 128+SIGINT (common shell convention).
	code := exitErr.ExitCode()
	return code == -1 || code == 130
}
