package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type ZypperSource struct {
	elevator *utils.Elevator
}

func NewZypperSource(elevator *utils.Elevator) sources.Source {
	return &ZypperSource{elevator: elevator}
}

func (z *ZypperSource) Name() string {
	return "Zypper"
}

func (z *ZypperSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("zypper", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (z *ZypperSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("rpm", "-q", app), nil
}

func (z *ZypperSource) Install(app string) error {
	return z.elevator.RunPrivileged("zypper", "install", "-y", app)
}

func (z *ZypperSource) Remove(app string) error {
	return z.elevator.RunPrivileged("zypper", "remove", "-y", app)
}

func (z *ZypperSource) Update() error {
	return z.elevator.RunPrivileged("sh", "-c", "zypper refresh && zypper update -y")
}

func (z *ZypperSource) ListUpdates() ([]string, error) {
	output, err := utils.RunCommand("zypper", "list-updates", "--no-refresh")
	if err != nil && strings.TrimSpace(output) == "" {
		return nil, nil
	}

	var updates []string
	for _, line := range strings.Split(output, "\n") {
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		status := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[2])
		if status == "" || name == "" {
			continue
		}
		if status == "S" || strings.HasPrefix(status, "-") {
			continue
		}
		if strings.Contains(status, "+") {
			updates = append(updates, name)
		}
	}
	return updates, nil
}

func (z *ZypperSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("zypper", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (z *ZypperSource) InstalledCount() (int, error) {
	lines, err := utils.GetCommandOutput("rpm", "-qa")
	if err != nil {
		return 0, nil
	}
	return len(lines), nil
}
