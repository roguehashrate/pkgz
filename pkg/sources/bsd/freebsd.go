package bsd

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type FreeBsdSource struct {
	elevator *utils.Elevator
}

func NewFreeBsdSource(elevator *utils.Elevator) sources.Source {
	return &FreeBsdSource{elevator: elevator}
}

func (f *FreeBsdSource) Name() string {
	return "FreeBSD"
}

func (f *FreeBsdSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("pkg", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (f *FreeBsdSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("pkg", "info", app), nil
}

func (f *FreeBsdSource) Install(app string) error {
	return f.elevator.RunPrivileged("pkg", "install", "-y", app)
}

func (f *FreeBsdSource) Remove(app string) error {
	return f.elevator.RunPrivileged("pkg", "delete", "-y", app)
}

func (f *FreeBsdSource) Update() error {
	// First update package database
	if err := f.elevator.RunPrivileged("pkg", "update"); err != nil {
		return err
	}
	// Then upgrade packages
	return f.elevator.RunPrivileged("pkg", "upgrade", "-y")
}

func (f *FreeBsdSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("pkg", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (f *FreeBsdSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("pkg", "info")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
