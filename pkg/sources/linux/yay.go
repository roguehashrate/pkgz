package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type YaySource struct {
	elevator *utils.Elevator
}

func NewYaySource(elevator *utils.Elevator) sources.Source {
	return &YaySource{elevator: elevator}
}

func (y *YaySource) Name() string {
	return "Yay (AUR)"
}

func (y *YaySource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("yay", "-Ss", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (y *YaySource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("yay", "-Qm", app), nil
}

func (y *YaySource) Install(app string) error {
	// Yay install doesn't need privilege elevation
	_, err := utils.RunCommand("yay", "-S", "--noconfirm", app)
	return err
}

func (y *YaySource) Remove(app string) error {
	// Yay remove doesn't need privilege elevation
	_, err := utils.RunCommand("yay", "-R", "--noconfirm", app)
	return err
}

func (y *YaySource) Update() error {
	// Yay update doesn't need privilege elevation
	_, err := utils.RunCommand("yay", "-Syu", "--noconfirm")
	return err
}

func (y *YaySource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("yay", "-Ss", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (y *YaySource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("yay", "-Qm")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
