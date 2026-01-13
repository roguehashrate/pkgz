package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type PacstallSource struct {
	elevator *utils.Elevator
}

func NewPacstallSource(elevator *utils.Elevator) sources.Source {
	return &PacstallSource{elevator: elevator}
}

func (p *PacstallSource) Name() string {
	return "Pacstall"
}

func (p *PacstallSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("pacstall", "-S", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (p *PacstallSource) Installed(app string) (bool, error) {
	// Check if app is in pacstall installed list
	output, err := utils.RunCommand("pacstall", "-L")
	if err != nil {
		return false, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == app {
			return true, nil
		}
	}
	return false, nil
}

func (p *PacstallSource) Install(app string) error {
	return p.elevator.RunPrivileged("pacstall", "-I", app)
}

func (p *PacstallSource) Remove(app string) error {
	return p.elevator.RunPrivileged("pacstall", "-R", app)
}

func (p *PacstallSource) Update() error {
	return p.elevator.RunPrivileged("pacstall", "-Up")
}

func (p *PacstallSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("pacstall", "-S", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (p *PacstallSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("pacstall", "-L")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
