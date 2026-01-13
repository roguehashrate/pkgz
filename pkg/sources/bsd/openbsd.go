package bsd

import (
	"fmt"
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type OpenBsdSource struct {
	elevator *utils.Elevator
}

func NewOpenBsdSource(elevator *utils.Elevator) sources.Source {
	return &OpenBsdSource{elevator: elevator}
}

func (o *OpenBsdSource) Name() string {
	return "OpenBSD"
}

func (o *OpenBsdSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("pkg_info", "|", "grep", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (o *OpenBsdSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("pkg_info", app), nil
}

func (o *OpenBsdSource) Install(app string) error {
	return o.elevator.RunPrivileged("pkg_add", app)
}

func (o *OpenBsdSource) Remove(app string) error {
	return o.elevator.RunPrivileged("pkg_delete", app)
}

func (o *OpenBsdSource) Update() error {
	fmt.Println("Please use `syspatch` or ports tree for updates on OpenBSD.")
	return nil
}

func (o *OpenBsdSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("pkg_info", "|", "grep", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (o *OpenBsdSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("pkg_info")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
