package utils

// Task is a lightweight progress/reporting hook that long-running operations
// can push status and captured output into. Sources stay TUI-agnostic: the
// orchestrator supplies a Task (e.g. a TUI task), or NoopTask when running in
// a plain/non-TTY context.
type Task interface {
	// SetLabel updates the human-readable label shown for this operation.
	SetLabel(label string)
	// SetStatus updates the short status text (e.g. "running", "done").
	SetStatus(status string)
	// AppendOutput appends a line of captured child output.
	AppendOutput(line string)
}

// NoopTask discards all updates. Used for non-TUI (plain) output paths.
type NoopTask struct{}

func (NoopTask) SetLabel(string)     {}
func (NoopTask) SetStatus(string)    {}
func (NoopTask) AppendOutput(string) {}

var _ Task = NoopTask{}
