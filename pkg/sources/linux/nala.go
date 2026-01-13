package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type NalaSource struct {
	elevator *utils.Elevator
}

func NewNalaSource(elevator *utils.Elevator) sources.Source {
	return &NalaSource{elevator: elevator}
}

func (n *NalaSource) Name() string {
	return "Nala"
}

func (n *NalaSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("nala", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (n *NalaSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("dpkg", "-s", app), nil
}

func (n *NalaSource) Install(app string) error {
	return n.elevator.RunPrivileged("nala", "install", "-y", app)
}

func (n *NalaSource) Remove(app string) error {
	return n.elevator.RunPrivileged("nala", "remove", "-y", app)
}

func (n *NalaSource) Update() error {
	return n.elevator.RunPrivileged("sh", "-c", "nala update && nala upgrade -y")
}

func (n *NalaSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("nala", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (n *NalaSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("dpkg-query", "-f", ".\n", "-W")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
