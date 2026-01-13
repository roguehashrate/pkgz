package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type FlatpakSource struct {
	elevator *utils.Elevator
}

func NewFlatpakSource(elevator *utils.Elevator) sources.Source {
	return &FlatpakSource{elevator: elevator}
}

func (f *FlatpakSource) Name() string {
	return "Flatpak"
}

func (f *FlatpakSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("flatpak", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (f *FlatpakSource) Installed(app string) (bool, error) {
	output, err := utils.RunCommand("flatpak", "list")
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (f *FlatpakSource) Install(app string) error {
	// Flatpak install doesn't need privilege elevation
	_, err := utils.RunCommand("flatpak", "install", "-y", app)
	return err
}

func (f *FlatpakSource) Remove(app string) error {
	// Flatpak uninstall doesn't need privilege elevation
	_, err := utils.RunCommand("flatpak", "uninstall", "-y", app)
	return err
}

func (f *FlatpakSource) Update() error {
	// Flatpak update doesn't need privilege elevation
	_, err := utils.RunCommand("flatpak", "update", "-y")
	return err
}

func (f *FlatpakSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("flatpak", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (f *FlatpakSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("flatpak", "list", "--app")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
