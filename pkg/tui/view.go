package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)

func accentStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
}

func statusMark(status string) string {
	switch status {
	case "running":
		return "●"
	case "done":
		return "✓"
	case "failed":
		return "✗"
	default:
		return "○"
	}
}

func statusColor(status string) lipgloss.Color {
	switch status {
	case "running":
		return lipgloss.Color("220")
	case "done":
		return lipgloss.Color("42")
	case "failed":
		return lipgloss.Color("196")
	default:
		return lipgloss.Color("240")
	}
}

// View renders the status list plus the (optional) output pane.
func (m *Model) View() string {
	if m.phase == phasePicker {
		return m.renderPicker()
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	rows := make([]string, 0, len(m.tasks))
	for i, t := range m.tasks {
		rows = append(rows, m.renderRow(i, t))
	}
	b.WriteString(strings.Join(rows, "\n"))

	b.WriteString("\n\n")

	if m.done && m.err != nil {
		b.WriteString(statusBarStyle.Render("Finished with errors.") + "\n")
	} else if m.done {
		b.WriteString(statusBarStyle.Render("All done.") + "\n")
	} else {
		b.WriteString(statusBarStyle.Render("Running..."))
	}

	b.WriteString("\n\n")
	b.WriteString(infoStyle.Render(helpText))

	if m.showLog {
		b.WriteString("\n\n")
		b.WriteString(logBoxStyle.Width(m.logWidth()).Render(m.renderLog()))
	}

	return b.String()
}

func (m *Model) renderRow(idx int, t *Task) string {
	status := t.Status()
	mark := statusMark(status)
	color := statusColor(status)

	row := lipgloss.NewStyle().
		Foreground(color).
		Render(mark + " ")

	if status == "running" {
		row += m.spinner.View() + " "
	} else {
		row += "   "
	}

	label := t.Label()

	if idx == m.selected {
		row += selectedStyle.Render("▸ " + label)
	} else {
		row += lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("  " + label)
	}

	return row
}

func (m *Model) renderLog() string {
	task := m.tasks[m.selected]
	lines := task.Lines()
	if len(lines) == 0 {
		return infoStyle.Render("(no output yet)")
	}
	// Keep only the last N lines that fit the pane.
	maxLines := 20
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return strings.Join(lines, "\n")
}

func (m *Model) logWidth() int {
	if m.width > 4 {
		return m.width - 6
	}
	return 70
}

const helpText = "↑/↓ select task · o toggle log · q quit"

const pickerHelpText = "↑/↓ move · enter select · q quit"

func (m *Model) renderPicker() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	if m.prompt != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(m.prompt))
		b.WriteString("\n\n")
	}

	for i, c := range m.choices {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ "+c) + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("  "+c) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(infoStyle.Render(pickerHelpText))

	return b.String()
}
