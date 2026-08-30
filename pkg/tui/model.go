package tui

import (
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type phase int

const (
	phasePicker phase = iota
	phaseRun
	phaseDone
)

// nextMsg tells the model to launch the next queued operation.
type nextMsg struct{}

// opDoneMsg is sent when a streaming (non-privileged) operation finishes.
type opDoneMsg struct {
	idx int
	err error
}

// opTermMsg is sent when a privileged operation that took over the terminal
// finishes (via bubbletea's Exec).
type opTermMsg struct {
	idx int
	err error
}

// Op describes a single long-running operation shown as one row in the TUI.
type Op struct {
	Label string
	Run   func(t *Task) error

	// Privileged marks an operation that needs direct terminal control (e.g. it
	// triggers a sudo/doas password prompt). In the TUI such ops are run with
	// bubbletea's Exec, which temporarily releases the terminal so the password
	// can actually be read, then restores the TUI. Non-privileged ops stream
	// their output straight into the log pane without touching the terminal.
	Privileged bool
}

// opExec adapts an Op's Run function to tea.ExecCommand so a privileged op can
// be executed with the terminal released into cooked mode (required for sudo/
// doas password prompts) while still capturing output into the task pane.
type opExec struct {
	op Op
	t  *Task
}

func (o opExec) SetStdin(io.Reader)  {}
func (o opExec) SetStdout(io.Writer) {}
func (o opExec) SetStderr(io.Writer) {}
func (o opExec) Run() error          { return o.op.Run(o.t) }

// Model is the bubbletea state for the pkgz TUI.
type Model struct {
	title string
	phase phase

	// picker state (used when choices are present)
	prompt  string
	choices []string
	cursor  int

	spinner spinner.Model
	tasks   []*Task
	ops     []Op

	selected int
	showLog  bool
	done     bool
	err      error

	width  int
	height int

	// queue holds op indices still to run.
	queue []int
	// curIdx is the op index currently executing, or -1 when idle.
	curIdx int
}

// NewModel builds the initial model. Tasks are created for each op label. When
// choices is non-empty the model starts in a picker phase; otherwise it runs
// every op in order.
func NewModel(title, prompt string, choices []string, ops []Op) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = accentStyle()

	labels := make([]string, len(ops))
	for i, op := range ops {
		labels[i] = op.Label
	}

	tasks := make([]*Task, len(labels))
	for i, l := range labels {
		tasks[i] = newTask(i, l)
	}

	ph := phaseRun
	if len(choices) > 0 {
		ph = phasePicker
	}

	queue := make([]int, len(ops))
	for i := range ops {
		queue[i] = i
	}
	if ph == phasePicker {
		queue = nil
	}

	return &Model{
		title:   title,
		phase:   ph,
		prompt:  prompt,
		choices: choices,
		spinner: s,
		tasks:   tasks,
		ops:     ops,
		curIdx:  -1,
		queue:   queue,
	}
}

// Tasks exposes the created tasks so tasks can attach program send hooks.
func (m *Model) Tasks() []*Task { return m.tasks }

// AttachSend wires each task's re-render notification to the program. The send
// fires asynchronously (in a goroutine) because tasks may notify from inside
// bubbletea's Update goroutine (e.g. when an op is launched), and calling
// Program.Send synchronously from there would deadlock the event loop.
func (m *Model) AttachSend(p *tea.Program) {
	for _, t := range m.tasks {
		t.send = func(msg tea.Msg) { go p.Send(msg) }
	}
}

func (m *Model) Init() tea.Cmd {
	if m.phase == phasePicker {
		return m.spinner.Tick
	}
	return tea.Batch(m.spinner.Tick, next())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.phase {
		case phasePicker:
			return m.updatePicker(msg)
		case phaseRun:
			return m.updateRun(msg)
		case phaseDone:
			return m.updateDone(msg)
		}
		return m, nil

	case nextMsg:
		return m.runNext()

	case opDoneMsg:
		if msg.idx >= 0 && msg.idx < len(m.tasks) {
			setTaskResult(m.tasks[msg.idx], msg.err)
		}
		if msg.err != nil && m.err == nil {
			m.err = msg.err
		}
		m.curIdx = -1
		return m.runNext()

	case opTermMsg:
		if msg.idx >= 0 && msg.idx < len(m.tasks) {
			setTaskResult(m.tasks[msg.idx], msg.err)
		}
		if msg.err != nil && m.err == nil {
			m.err = msg.err
		}
		m.curIdx = -1
		return m.runNext()

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	default:
		if m.phase == phaseRun {
			return m.updateRunMsg(msg)
		}
		return m, nil
	}
}

func (m *Model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.phase = phaseDone
		return m, tea.Quit
	case tea.KeyUp, tea.KeyLeft:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown, tea.KeyRight:
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case tea.KeyEnter, tea.KeySpace:
		// Exit picker and run only the chosen op.
		m.selected = m.cursor
		m.phase = phaseRun
		m.queue = []int{m.cursor}
		return m, next()
	}

	if msg.String() == "q" {
		m.phase = phaseDone
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) updateRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.phase = phaseDone
		return m, tea.Quit
	case tea.KeyUp, tea.KeyLeft:
		if m.selected > 0 {
			m.selected--
		}
	case tea.KeyDown, tea.KeyRight:
		if m.selected < len(m.tasks)-1 {
			m.selected++
		}
	}

	if msg.String() == "o" {
		m.showLog = !m.showLog
	} else if msg.String() == "q" {
		m.phase = phaseDone
		return m, tea.Quit
	}
	return m, nil
}

// updateDone handles keys on the final summary screen. `o` keeps toggling the
// captured-output log so the user can inspect results; the documented quit keys
// (q, esc, ctrl+c, enter) return to the shell. Any other key also dismisses, as
// the on-screen hint promises "press any key to exit".
func (m *Model) updateDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc, tea.KeyEnter:
		return m, tea.Quit
	}
	switch msg.String() {
	case "o":
		m.showLog = !m.showLog
		return m, nil
	case "q":
		return m, tea.Quit
	default:
		return m, tea.Quit
	}
}

func (m *Model) updateRunMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case taskUpdateMsg:
		if !m.done {
			return m, m.spinner.Tick
		}
	}
	return m, nil
}

// runNext launches the next queued op, or finishes when the queue is empty.
// On completion the model switches to a done phase that keeps the final summary
// on screen until the user dismisses it (so results are never just a blink).
func (m *Model) runNext() (tea.Model, tea.Cmd) {
	if len(m.queue) == 0 {
		m.done = true
		m.phase = phaseDone
		return m, nil
	}

	idx := m.queue[0]
	m.queue = m.queue[1:]
	m.curIdx = idx
	m.tasks[idx].SetStatus("running")

	op := m.ops[idx]
	if op.Privileged {
		// Run with the terminal released so sudo/doas can prompt for a password.
		return m, tea.Exec(opExec{op: op, t: m.tasks[idx]}, func(err error) tea.Msg {
			return opTermMsg{idx: idx, err: err}
		})
	}

	// Non-privileged: stream output into the pane on a background goroutine.
	go func() {
		err := op.Run(m.tasks[idx])
		m.tasks[idx].send(opDoneMsg{idx: idx, err: err})
	}()
	return m, nil
}

func setTaskResult(t *Task, err error) {
	if err != nil {
		t.SetStatus("failed")
	} else {
		t.SetStatus("done")
	}
}

// next returns a Cmd that triggers the next queued operation.
func next() tea.Cmd {
	return func() tea.Msg { return nextMsg{} }
}
