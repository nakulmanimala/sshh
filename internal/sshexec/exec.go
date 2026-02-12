package sshexec

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"sshh/internal/model"
)

// Connect replaces the current process with an ssh connection to the server.
// This function does not return on success.
func Connect(s model.Server) error {
	sshBin, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	args := []string{"ssh"}

	if s.Port != 0 && s.Port != 22 {
		args = append(args, "-p", strconv.Itoa(s.Port))
	}
	if s.Key != "" {
		args = append(args, "-i", s.Key)
	}

	target := s.Host
	if s.User != "" {
		target = s.User + "@" + s.Host
	}
	args = append(args, target)

	// Replace current process with ssh.
	return syscall.Exec(sshBin, args, os.Environ())
}
