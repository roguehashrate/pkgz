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
// index). Otherwise every op runs in order. Returns the first operation error
// encountered (nil if all succeeded).
func Run(title, prompt string, choices []string, ops []Op) error {
	labels := make([]string, len(ops))
	for i := range ops {
		labels[i] = ops[i].Label
	}

	picked := make(chan int, 1)
	m := NewModel(title, prompt, choices, labels, picked)
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.AttachSend(p)

	go runOps(p, m, ops, choices, picked)

	final, err := p.Run()
	if err != nil {
		return err
	}
	if mm, ok := final.(*Model); ok {
		return mm.err
	}
	return nil
}

// RunPlain executes the operations with plain sequential console output. Used
// when stdout is not a terminal (pipes, CI, scripts).
func RunPlain(ops []Op) error {
	var firstErr error
	for i, op := range ops {
		fmt.Printf("▶ %s\n", ops[i].Label)
		t := newTask(i, ops[i].Label, nil)
		if err := op.Run(t); err != nil {
			fmt.Printf("  ✗ %v\n", err)
			if firstErr == nil {
				firstErr = err
			}
		} else {
			fmt.Printf("  ✓ done\n")
		}
	}
	return firstErr
}

func runOps(p *tea.Program, m *Model, ops []Op, choices []string, picked chan int) {
	var firstErr error

	if len(choices) > 0 {
		idx, ok := <-picked
		if !ok || idx < 0 || idx >= len(ops) {
			p.Send(allDoneMsg{})
			return
		}
		t := m.Tasks()[idx]
		t.SetStatus("running")
		if err := ops[idx].Run(t); err != nil {
			t.SetStatus("failed")
			firstErr = err
		} else {
			t.SetStatus("done")
		}
		p.Send(opDoneMsg{})
		p.Send(allDoneMsg{err: firstErr})
		return
	}

	for i, op := range ops {
		t := m.Tasks()[i]
		t.SetStatus("running")
		if err := op.Run(t); err != nil {
			t.SetStatus("failed")
			if firstErr == nil {
				firstErr = err
			}
		} else {
			t.SetStatus("done")
		}
		p.Send(opDoneMsg{})
	}
	p.Send(allDoneMsg{err: firstErr})
}
