package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type DnfSource struct {
	elevator *utils.Elevator
}

func NewDnfSource(elevator *utils.Elevator) sources.Source {
	return &DnfSource{elevator: elevator}
}

func (d *DnfSource) Name() string {
	return "DNF"
}

func (d *DnfSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("dnf", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (d *DnfSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("rpm", "-q", app), nil
}

func (d *DnfSource) Install(app string) error {
	return d.elevator.RunPrivileged("dnf", "install", "-y", app)
}

func (d *DnfSource) Remove(app string) error {
	return d.elevator.RunPrivileged("dnf", "remove", "-y", app)
}

func (d *DnfSource) Update() error {
	return d.elevator.RunPrivileged("dnf", "upgrade", "-y")
}

func (d *DnfSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("dnf", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (d *DnfSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("rpm", "-qa")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
