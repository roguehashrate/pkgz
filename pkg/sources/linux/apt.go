package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type AptSource struct {
	elevator *utils.Elevator
}

func NewAptSource(elevator *utils.Elevator) sources.Source {
	return &AptSource{elevator: elevator}
}

func (a *AptSource) Name() string {
	return "Apt"
}

func (a *AptSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("apt-cache", "search", app)
	if err != nil {
		return false, nil // Return false, not error, for availability
	}
	return strings.Contains(output, app), nil
}

func (a *AptSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("dpkg", "-s", app), nil
}

func (a *AptSource) Install(app string) error {
	return a.elevator.RunPrivileged("apt", "install", "-y", app)
}

func (a *AptSource) Remove(app string) error {
	return a.elevator.RunPrivileged("apt", "remove", "-y", app)
}

func (a *AptSource) Update() error {
	return a.elevator.RunPrivileged("sh", "-c", "apt update && apt upgrade -y")
}

func (a *AptSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("apt-cache", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (a *AptSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("dpkg-query", "-f", ".\n", "-W")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
