package utils

import (
	"os"
	"os/exec"
	"strings"
)

// RunCommand runs a command and returns its combined output
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCommandSilent runs a command and returns true if successful
func RunCommandSilent(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	err := cmd.Run()
	return err == nil
}

// RunCommandWithRedirect runs a command with output redirected (like > /dev/null 2>&1)
func RunCommandWithRedirect(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

// GetCommandOutput runs a command and returns its stdout as lines
func GetCommandOutput(name string, args ...string) ([]string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Filter out empty lines
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

// RunPrivilegedCommand runs a command with privilege escalation
func RunPrivilegedCommand(elevator, cmd string, args ...string) error {
	fullArgs := append([]string{cmd}, args...)
	execCmd := exec.Command(elevator, fullArgs...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}
