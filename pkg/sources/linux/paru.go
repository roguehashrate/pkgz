package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type ParuSource struct {
	elevator *utils.Elevator
}

func NewParuSource(elevator *utils.Elevator) sources.Source {
	return &ParuSource{elevator: elevator}
}

func (p *ParuSource) Name() string {
	return "Paru (AUR)"
}

func (p *ParuSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("paru", "-Ss", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (p *ParuSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("paru", "-Qm", app), nil
}

func (p *ParuSource) Install(app string) error {
	// Paru install doesn't need privilege elevation
	_, err := utils.RunCommand("paru", "-S", "--noconfirm", app)
	return err
}

func (p *ParuSource) Remove(app string) error {
	// Paru remove doesn't need privilege elevation
	_, err := utils.RunCommand("paru", "-R", "--noconfirm", app)
	return err
}

func (p *ParuSource) Update() error {
	// Paru update doesn't need privilege elevation
	_, err := utils.RunCommand("paru", "-Syu", "--noconfirm")
	return err
}

func (p *ParuSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("paru", "-Ss", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (p *ParuSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("paru", "-Qm")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
