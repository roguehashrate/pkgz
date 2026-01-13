package utils

import (
	"os"
	"os/exec"
)

// Elevator handles privilege escalation
type Elevator struct {
	command string
}

// NewElevator creates a new elevator instance
func NewElevator() *Elevator {
	return &Elevator{}
}

// GetElevatorCommand returns the command to use for privilege escalation
func (e *Elevator) GetElevatorCommand(configCommand string) string {
	if e.command != "" {
		return e.command
	}

	// Use configured command if available
	if configCommand != "" && configCommand != "sudo" && configCommand != "doas" {
		e.command = configCommand
		return e.command
	}

	// Fallback detection: prefer doas over sudo
	if CommandExists("doas") {
		e.command = "doas"
	} else {
		e.command = "sudo"
	}

	return e.command
}

// RunPrivileged runs a command with privilege escalation
func (e *Elevator) RunPrivileged(cmd string, args ...string) error {
	elevator := e.GetElevatorCommand("")
	fullArgs := append([]string{cmd}, args...)

	execCmd := exec.Command(elevator, fullArgs...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// CommandExists checks if a command exists in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
