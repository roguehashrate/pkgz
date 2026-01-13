package bsd

import (
	"fmt"
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type FreeBsdPortsSource struct {
	elevator *utils.Elevator
}

func NewFreeBsdPortsSource(elevator *utils.Elevator) sources.Source {
	return &FreeBsdPortsSource{elevator: elevator}
}

func (f *FreeBsdPortsSource) Name() string {
	return "FreeBSD Ports"
}

func (f *FreeBsdPortsSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("make", "-C", "/usr/ports/", "search", "key="+app, "2>/dev/null")
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (f *FreeBsdPortsSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("pkg", "info", app), nil
}

func (f *FreeBsdPortsSource) Install(app string) error {
	portsPath := "/usr/ports/" + app
	return f.elevator.RunPrivileged("make", "-C", portsPath, "install", "clean", "BATCH=yes")
}

func (f *FreeBsdPortsSource) Remove(app string) error {
	return f.elevator.RunPrivileged("pkg", "delete", "-y", app)
}

func (f *FreeBsdPortsSource) Update() error {
	fmt.Println("Update ports tree manually with 'portsnap fetch update' or via git.")
	return nil
}

func (f *FreeBsdPortsSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("make", "-C", "/usr/ports/", "search", "key="+app, "2>/dev/null")
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (f *FreeBsdPortsSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("pkg", "info")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
