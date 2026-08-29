package tui

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// taskUpdateMsg is sent after any task mutation to trigger a re-render.
type taskUpdateMsg struct{}

// Task implements utils.Task and accumulates status + captured output, sending
// a lightweight re-render message to the bubbletea program after each change.
type Task struct {
	id     int
	label  string
	status string

	mu     sync.Mutex
	output []string

	send func(tea.Msg)
}

func newTask(id int, label string, send func(tea.Msg)) *Task {
	if send == nil {
		send = func(tea.Msg) {}
	}
	return &Task{id: id, label: label, status: "pending", send: send}
}

func (t *Task) ID() int { return t.id }

func (t *Task) Label() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.label
}

func (t *Task) Status() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.status
}

// Lines returns a copy of the captured output.
func (t *Task) Lines() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]string, len(t.output))
	copy(out, t.output)
	return out
}

func (t *Task) SetLabel(label string) {
	t.mu.Lock()
	t.label = label
	t.mu.Unlock()
	t.send(taskUpdateMsg{})
}

func (t *Task) SetStatus(status string) {
	t.mu.Lock()
	t.status = status
	t.mu.Unlock()
	t.send(taskUpdateMsg{})
}

func (t *Task) AppendOutput(line string) {
	t.mu.Lock()
	t.output = append(t.output, line)
	t.mu.Unlock()
	t.send(taskUpdateMsg{})
}
