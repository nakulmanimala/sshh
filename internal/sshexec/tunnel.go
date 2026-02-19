package sshexec

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"sshh/internal/model"
)

// RunTunnel replaces the current process with an SSH tunnel command.
// Uses -N to skip remote command execution (port-forward only).
// This function does not return on success.
func RunTunnel(t model.Tunnel) error {
	sshBin, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	args := []string{"ssh", "-N"}

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

	target := t.SSHHost
	if t.SSHUser != "" {
		target = t.SSHUser + "@" + t.SSHHost
	}
	args = append(args, target)

	return syscall.Exec(sshBin, args, os.Environ())
}
