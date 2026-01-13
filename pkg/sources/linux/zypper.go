package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type ZypperSource struct {
	elevator *utils.Elevator
}

func NewZypperSource(elevator *utils.Elevator) sources.Source {
	return &ZypperSource{elevator: elevator}
}

func (z *ZypperSource) Name() string {
	return "Zypper"
}

func (z *ZypperSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("zypper", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (z *ZypperSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("rpm", "-q", app), nil
}

func (z *ZypperSource) Install(app string) error {
	return z.elevator.RunPrivileged("zypper", "install", "-y", app)
}

func (z *ZypperSource) Remove(app string) error {
	return z.elevator.RunPrivileged("zypper", "remove", "-y", app)
}

func (z *ZypperSource) Update() error {
	return z.elevator.RunPrivileged("sh", "-c", "zypper refresh && zypper update -y")
}

func (z *ZypperSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("zypper", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (z *ZypperSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("rpm", "-qa")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
