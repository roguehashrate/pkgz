package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type ApkSource struct {
	elevator *utils.Elevator
}

func NewApkSource(elevator *utils.Elevator) sources.Source {
	return &ApkSource{elevator: elevator}
}

func (a *ApkSource) Name() string {
	return "Alpine"
}

func (a *ApkSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("apk", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (a *ApkSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("apk", "info", "-e", app), nil
}

func (a *ApkSource) Install(app string) error {
	return a.elevator.RunPrivileged("apk", "add", app)
}

func (a *ApkSource) Remove(app string) error {
	return a.elevator.RunPrivileged("apk", "del", app)
}

func (a *ApkSource) Update() error {
	return a.elevator.RunPrivileged("sh", "-c", "apk update && apk upgrade")
}

func (a *ApkSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("apk", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (a *ApkSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("apk", "info")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
