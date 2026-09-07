package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// Elevator handles privilege escalation
type Elevator struct {
	command string
}

// NewElevator creates a new elevator instance.
func NewElevator() *Elevator {
	return &Elevator{}
}

// SetCommand pins the elevator command (e.g. from config). Auto-detection is
// only used when no explicit command is provided.
func (e *Elevator) SetCommand(command string) {
	if command != "" {
		e.command = command
	}
}

// GetElevatorCommand returns the command to use for privilege escalation,
// preferring an explicitly configured command over auto-detection (doas, then
// sudo). An error is returned when no usable elevator exists.
func (e *Elevator) GetElevatorCommand(configCommand string) (string, error) {
	if e.command != "" {
		return e.command, nil
	}
	if configCommand != "" {
		if !CommandExists(configCommand) {
			return "", fmt.Errorf("configured elevator command %q not found in PATH (check [elevator] in ~/.config/pkgz/config.toml)", configCommand)
		}
		e.command = configCommand
		return e.command, nil
	}

	// Auto-detect: prefer doas, then sudo, then pkexec.
	for _, candidate := range []string{"doas", "sudo", "pkexec"} {
		if CommandExists(candidate) {
			e.command = candidate
			return e.command, nil
		}
	}
	return "", fmt.Errorf("no privilege elevation tool found (doas, sudo, or pkexec); install one of them")
}

// Command returns the resolved elevator command (auto-detecting if needed).
func (e *Elevator) Command() string {
	cmd, _ := e.GetElevatorCommand("")
	return cmd
}

// RunPrivileged runs a command with privilege escalation
func (e *Elevator) RunPrivileged(cmd string, args ...string) error {
	elevator, err := e.GetElevatorCommand("")
	if err != nil {
		return err
	}
	fullArgs := append([]string{cmd}, args...)

	execCmd := exec.Command(elevator, fullArgs...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// RunPrivilegedStreaming runs a command with privilege escalation, forwarding
// each line of its combined stdout+stderr to onLine while keeping stdin
// attached so the elevator (sudo/doas) can prompt for passwords.
func (e *Elevator) RunPrivilegedStreaming(cmd string, args []string, onLine func(string)) error {
	elevator, err := e.GetElevatorCommand("")
	if err != nil {
		return err
	}
	fullArgs := append([]string{cmd}, args...)
	return RunCommandStreaming(elevator, fullArgs, onLine)
}

// CommandExists checks if a command exists in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
