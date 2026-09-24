package utils

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

// ansiRe strips common ANSI CSI escape sequences (color, cursor moves,
// erase-line) from captured child output so progress lines render cleanly
// instead of as raw escape garbage inside the log pane.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// scanProgress is a bufio.SplitFunc that breaks output on `\r`, `\n` or `\r\n`.
// Package-manager progress meters rewrite the same line with `\r`; splitting on
// it turns each progress frame into its own line instead of one growing blob.
func scanProgress(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// cleanLine strips ANSI escapes and stray carriage returns/trailing whitespace
// from a captured progress line, returning "" for effectively-empty frames.
func cleanLine(s string) string {
	s = ansiRe.ReplaceAllString(s, "")
	s = strings.TrimRight(s, "\r")
	s = strings.TrimSpace(s)
	if s == "\n" {
		return ""
	}
	return s
}

// RunCommandStreaming runs a command, forwarding each line of its combined
// stdout+stderr to onLine, and keeping stdin attached so interactive prompts
// (e.g. sudo/doas passwords) still work. Returns any error the command exits
// with (nil exit code => nil error).
func RunCommandStreaming(name string, args []string, onLine func(string)) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	lineCh := make(chan string, 64)
	var wgdone sync.WaitGroup
	wgdone.Add(1)
	go func() {
		defer wgdone.Done()
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		scanner.Split(scanProgress)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := cleanLine(scanner.Text())
			if line == "" {
				continue
			}
			lineCh <- line
		}
	}()

	go func() {
		wgdone.Wait()
		close(lineCh)
	}()

	for line := range lineCh {
		onLine(line)
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

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
