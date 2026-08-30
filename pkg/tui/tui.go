package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

var _ utils.Task = (*Task)(nil)

// Run executes the given operations inside the pkgz TUI. The caller must have
// already verified that stdout is a terminal. If choices is non-empty (length
// must equal len(ops)), the TUI starts in a picker phase: the user selects one
// choice, and only the corresponding op is executed (args and labels aligned by
// index). Otherwise every op runs in order. Privileged ops take over the
// terminal while they run (so sudo/doas prompts work) and resume the TUI when
// they finish. Returns the first operation error (nil if all succeeded).
func Run(title, prompt string, choices []string, ops []Op) error {
	opErr, _ := RunAny(title, prompt, choices, ops)
	return opErr
}

// RunAny is like Run but reports the result of the operations (opErr) and any
// failure of the bubbletea program/terminal layer (progErr) separately. Callers
// can use progErr to detect when the TUI could not start (e.g. an unsupported
// or broken terminal) and fall back to plain output instead of silently closing.
func RunAny(title, prompt string, choices []string, ops []Op) (opErr error, progErr error) {
	m := NewModel(title, prompt, choices, ops)
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.AttachSend(p)

	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	if mm, ok := final.(*Model); ok {
		return mm.err, nil
	}
	return nil, nil
}

// RunPlain executes the operations with plain sequential console output. Used
// when stdout is not a terminal (pipes, CI, scripts).
func RunPlain(ops []Op) error {
	var firstErr error
	for i, op := range ops {
		fmt.Printf("▶ %s\n", ops[i].Label)
		t := newTask(i, ops[i].Label)
		if err := op.Run(t); err != nil {
			fmt.Printf("  ✗ %v\n", err)
			if firstErr == nil {
				firstErr = err
			}
		} else {
			fmt.Printf("  ✓ done\n")
		}
		// Surface any detail the operation captured, so non-TTY output is not
		// reduced to just "done".
		for _, line := range t.Lines() {
			fmt.Printf("  %s\n", line)
		}
	}
	return firstErr
}
