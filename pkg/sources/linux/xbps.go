package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type XbpsSource struct {
	elevator *utils.Elevator
}

func NewXbpsSource(elevator *utils.Elevator) sources.Source {
	return &XbpsSource{elevator: elevator}
}

func (x *XbpsSource) Name() string {
	return "XBPS"
}

func (x *XbpsSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("xbps-query", "-Rs", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (x *XbpsSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("xbps-query", "-p", "pkgver", app), nil
}

func (x *XbpsSource) Install(app string) error {
	return x.elevator.RunPrivileged("xbps-install", "-Sy", app)
}

func (x *XbpsSource) Remove(app string) error {
	return x.elevator.RunPrivileged("xbps-remove", "-Ry", app)
}

func (x *XbpsSource) Update() error {
	return x.elevator.RunPrivileged("xbps-install", "-Syu")
}

func (x *XbpsSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("xbps-query", "-Rs", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (x *XbpsSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("xbps-query", "-l")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
