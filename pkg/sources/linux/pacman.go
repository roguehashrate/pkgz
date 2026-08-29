package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type PacmanSource struct {
	elevator *utils.Elevator
}

func NewPacmanSource(elevator *utils.Elevator) sources.Source {
	return &PacmanSource{elevator: elevator}
}

func (p *PacmanSource) Name() string {
	return "Pacman"
}

func (p *PacmanSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("pacman", "-Ss", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (p *PacmanSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("pacman", "-Qn", app), nil
}

func (p *PacmanSource) Install(app string) error {
	return p.elevator.RunPrivileged("pacman", "-S", "--noconfirm", app)
}

func (p *PacmanSource) Remove(app string) error {
	return p.elevator.RunPrivileged("pacman", "-R", "--noconfirm", app)
}

func (p *PacmanSource) Update() error {
	return p.elevator.RunPrivileged("pacman", "-Syu", "--noconfirm")
}

func (p *PacmanSource) ListUpdates() ([]string, error) {
	output, err := utils.RunCommand("pacman", "-Qu")
	if err != nil && strings.TrimSpace(output) == "" {
		return nil, nil
	}

	var updates []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		updates = append(updates, fields[0])
	}
	return updates, nil
}

func (p *PacmanSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("pacman", "-Ss", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (p *PacmanSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("pacman", "-Qn")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
