package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type phase int

const (
	phasePicker phase = iota
	phaseRun
	phaseDone
)

// opDoneMsg is sent when a single operation finishes.
type opDoneMsg struct{}

// allDoneMsg signals that every operation has completed.
type allDoneMsg struct{ err error }

// startMsg tells the runner goroutine to begin executing the selected op.
type startMsg struct{ index int }

// Op describes a single long-running operation shown as one row in the TUI.
type Op struct {
	Label string
	Run   func(t *Task) error
}

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

	selected int
	showLog  bool
	done     bool
	err      error

	width  int
	height int

	picked chan int
}

// NewModel builds the initial model. Tasks are created here with a no-op send;
// the caller attaches the real program send before starting execution. When
// choices is non-empty the model starts in a picker phase; otherwise it runs
// every op directly.
func NewModel(title, prompt string, choices []string, labels []string, picked chan int) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = accentStyle()

	tasks := make([]*Task, len(labels))
	for i, l := range labels {
		tasks[i] = newTask(i, l, func(tea.Msg) {})
	}

	ph := phaseRun
	if len(choices) > 0 {
		ph = phasePicker
	}

	return &Model{
		title:   title,
		phase:   ph,
		prompt:  prompt,
		choices: choices,
		spinner: s,
		tasks:   tasks,
		picked:  picked,
	}
}

// Tasks exposes the created tasks so the caller can attach the program send.
func (m *Model) Tasks() []*Task { return m.tasks }

// AttachSend wires each task's re-render notification to the program.
func (m *Model) AttachSend(p *tea.Program) {
	for _, t := range m.tasks {
		t.send = func(msg tea.Msg) { p.Send(msg) }
	}
}

func (m *Model) Init() tea.Cmd {
	if m.phase == phasePicker {
		return nil
	}
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.phase {
		case phasePicker:
			return m.updatePicker(msg)
		case phaseRun:
			return m.updateRun(msg)
		}
		return m, nil

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
		// Exit picker, switch to run phase for the chosen op.
		m.selected = m.cursor
		m.phase = phaseRun
		if m.picked != nil {
			m.picked <- m.cursor
		}
		return m, nil
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
	case opDoneMsg:
		if !m.done {
			return m, m.spinner.Tick
		}
	case allDoneMsg:
		m.done = true
		m.err = msg.err
		m.phase = phaseDone
		return m, tea.Quit
	}
	return m, nil
}
