package bsd

import (
	"fmt"
	"os"
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type OpenBsdPortsSource struct {
	elevator *utils.Elevator
}

func NewOpenBsdPortsSource(elevator *utils.Elevator) sources.Source {
	return &OpenBsdPortsSource{elevator: elevator}
}

func (o *OpenBsdPortsSource) Name() string {
	return "OpenBSD Ports"
}

func (o *OpenBsdPortsSource) Available(app string) (bool, error) {
	portsPath := "/usr/ports/" + app
	info, err := os.Stat(portsPath)
	if err != nil {
		return false, nil
	}
	return info.IsDir(), nil
}

func (o *OpenBsdPortsSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("pkg_info", app), nil
}

func (o *OpenBsdPortsSource) Install(app string) error {
	portsPath := "/usr/ports/" + app
	return o.elevator.RunPrivileged("cd", portsPath, "&&", "make", "install", "clean")
}

func (o *OpenBsdPortsSource) Remove(app string) error {
	return o.elevator.RunPrivileged("pkg_delete", app)
}

func (o *OpenBsdPortsSource) Update() error {
	fmt.Println("Update ports tree manually via 'svn update /usr/ports' or git.")
	return nil
}

func (o *OpenBsdPortsSource) Search(app string) (bool, error) {
	portsDir := "/usr/ports"

	// Get all directories in ports tree
	entries, err := os.ReadDir(portsDir)
	if err != nil {
		return false, nil
	}

	appLower := strings.ToLower(app)
	for _, entry := range entries {
		if entry.IsDir() {
			name := strings.ToLower(entry.Name())
			if strings.Contains(name, appLower) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (o *OpenBsdPortsSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("pkg_info")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
