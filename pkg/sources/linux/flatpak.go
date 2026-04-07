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
	appID, err := f.findAppID(app)
	if err != nil {
		return false, nil
	}
	return appID != "", nil
}

func (f *FlatpakSource) findAppID(app string) (string, error) {
	output, err := utils.RunCommand("flatpak", "search", "--columns=application,name", app)
	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")
	appLower := strings.ToLower(app)

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		appID := parts[0]
		appName := parts[1]

		if strings.Contains(strings.ToLower(appName), appLower) ||
			strings.Contains(strings.ToLower(appID), appLower) {
			return appID, nil
		}
	}
	return "", nil
}

func (f *FlatpakSource) Installed(app string) (bool, error) {
	output, err := utils.RunCommand("flatpak", "list", "--user")
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (f *FlatpakSource) Install(app string) error {
	appID, err := f.findAppID(app)
	if err != nil || appID == "" {
		return f.installWithName(app)
	}
	_, err = utils.RunCommand("flatpak", "install", "--user", "-y", "flathub", appID)
	return err
}

func (f *FlatpakSource) installWithName(app string) error {
	_, err := utils.RunCommand("flatpak", "install", "--user", "-y", "flathub", app)
	return err
}

func (f *FlatpakSource) Remove(app string) error {
	_, err := utils.RunCommand("flatpak", "uninstall", "--user", "-y", app)
	return err
}

func (f *FlatpakSource) Update() error {
	_, err := utils.RunCommand("flatpak", "update", "--user", "-y")
	return err
}

func (f *FlatpakSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("flatpak", "search", "--columns=name", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (f *FlatpakSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("flatpak", "list", "--user", "--app")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
