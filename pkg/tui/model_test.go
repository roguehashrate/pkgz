package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func upd(t *testing.T, m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	mm, ok := next.(*Model)
	if !ok {
		t.Fatalf("Update returned %T, want *Model", next)
	}
	return mm, cmd
}

func TestModelSequencePrivilegedThenStreaming(t *testing.T) {
	streamRan := make(chan struct{}, 1)
	ops := []Op{
		{Label: "priv", Privileged: true, Run: func(t *Task) error {
			return nil
		}},
		{Label: "stream", Privileged: false, Run: func(t *Task) error {
			streamRan <- struct{}{}
			return nil
		}},
	}

	m := NewModel("test", "", nil, ops)
	for _, tk := range m.Tasks() {
		tk.send = func(tea.Msg) {}
	}

	// Launch the first (privileged) op -> a tea.Exec command.
	before := m.curIdx
	m, cmd := upd(t, m, nextMsg{})
	if cmd == nil {
		t.Fatal("expected a tea.Exec command for privileged op")
	}
	if m.curIdx == before {
		t.Fatal("expected a privileged op to be in progress")
	}
	if m.Tasks()[0].Status() != "running" {
		t.Fatalf("op0 status = %q, want running", m.Tasks()[0].Status())
	}
	_ = cmd() // resolves to the execMsg (consumed by the real program)

	// Privileged op completes via the tea.Exec callback path.
	m, cmd = upd(t, m, opTermMsg{idx: 0, err: nil})
	if m.Tasks()[0].Status() != "done" {
		t.Fatalf("op0 status = %q, want done", m.Tasks()[0].Status())
	}
	_ = cmd // streaming op returns no cmd; it launches a goroutine

	// Streaming op launched and actually runs on its goroutine.
	if m.Tasks()[1].Status() != "running" {
		t.Fatalf("op1 status = %q, want running", m.Tasks()[1].Status())
	}
	select {
	case <-streamRan:
	case <-time.After(2 * time.Second):
		t.Fatal("expected the streaming op's Run to be invoked")
	}

	// Streaming op completes.
	m, _ = upd(t, m, opDoneMsg{idx: 1, err: nil})
	if m.Tasks()[1].Status() != "done" {
		t.Fatalf("op1 status = %q, want done", m.Tasks()[1].Status())
	}

	// Empty queue -> finish; the model holds on the done screen (no quit yet)
	// so the summary can be seen before returning to the shell.
	m, cmd = upd(t, m, nextMsg{})
	if cmd != nil {
		t.Fatal("expected no command when finishing the empty queue")
	}
	if !m.done || m.phase != phaseDone {
		t.Fatalf("expected done phase, got done=%v phase=%v", m.done, m.phase)
	}

	// A key press dismisses the done screen and quits.
	if _, cmd := upd(t, m, tea.KeyMsg{}); cmd == nil {
		t.Fatal("expected a quit command on key press in done phase")
	}
}

// TestDonePhaseLogToggle verifies that on the completion screen `o` toggles the
// captured-output log (so it must NOT quit), while `q` dismisses and quits.
func TestDonePhaseLogToggle(t *testing.T) {
	m := NewModel("t", "", nil, []Op{{Label: "x"}})
	m.done = true
	m.phase = phaseDone

	// `o` toggles the log open and returns no quit command.
	m, cmd := upd(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	if !m.showLog {
		t.Fatal("expected `o` to open the log on the done screen")
	}
	if cmd != nil {
		t.Fatalf("expected `o` not to quit, got cmd=%v", cmd)
	}

	// Pressing `o` again closes it, still no quit.
	m, cmd = upd(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	if m.showLog {
		t.Fatal("expected second `o` to close the log")
	}
	if cmd != nil {
		t.Fatalf("expected `o` not to quit, got cmd=%v", cmd)
	}

	// `q` still dismisses and quits.
	if _, cmd := upd(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}); cmd == nil {
		t.Fatal("expected `q` to quit on the done screen")
	}
}
