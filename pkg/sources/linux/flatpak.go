package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewFlatpakSource returns a source backed by flatpak (cross-distro).
func NewFlatpakSource(elevator *utils.Elevator) sources.Source {
	c := &commandSource{
		name: "Flatpak",
		available: func(app string) (bool, error) {
			appID, err := flatpakFindAppID(app)
			if err != nil {
				return false, nil
			}
			return appID != "", nil
		},
		installed: func(app string) (bool, error) {
			appID, _ := flatpakInstalledApp(app)
			return appID != "", nil
		},
		install: func(app string) error {
			appID, err := flatpakFindAppID(app)
			if err != nil || appID == "" {
				_, err = utils.RunCommand("flatpak", "install", "--user", "-y", "flathub", app)
				return err
			}
			_, err = utils.RunCommand("flatpak", "install", "--user", "-y", "flathub", appID)
			return err
		},
		remove: func(app string) error {
			appID, scope := flatpakInstalledApp(app)
			if appID == "" {
				appID = app
				scope = "--user"
			}
			_, err := utils.RunCommand("flatpak", "uninstall", scope, "-y", appID)
			return err
		},
		update: func() error {
			_, err := utils.RunCommand("flatpak", "update", "--user", "-y")
			return err
		},
		listUpdates: func() ([]string, error) {
			output, err := utils.RunCommand("flatpak", "remote-ls", "--user", "--updates")
			if err != nil && strings.TrimSpace(output) == "" {
				return nil, nil
			}
			var updates []string
			for _, line := range strings.Split(output, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				line = strings.TrimPrefix(line, "app/")
				line = strings.TrimPrefix(line, "runtime/")
				if idx := strings.Index(line, "/"); idx > 0 {
					line = line[:idx]
				}
				updates = append(updates, line)
			}
			return updates, nil
		},
		search: func(app string) (bool, error) {
			output, err := utils.RunCommand("flatpak", "search", "--columns=name", app)
			if err != nil {
				return false, nil
			}
			return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
		},
		installedCount: countOutput("flatpak", "list", "--user", "--app"),
	}
	return c
}

func flatpakFindAppID(app string) (string, error) {
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

// flatpakInstalledApp returns the App ID and scope (--user or --system) of an
// installed flatpak matching app, or empty strings if it is not installed.
func flatpakInstalledApp(app string) (appID, scope string) {
	appLower := strings.ToLower(app)
	for _, scope := range []string{"--user", "--system"} {
		output, err := utils.RunCommand("flatpak", "list", scope, "--columns=application,name")
		if err != nil {
			// Installing in another scope can make a listing command fail; keep trying.
			continue
		}
		for _, line := range strings.Split(output, "\n") {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) != 2 {
				continue
			}
			id := strings.TrimSpace(parts[0])
			name := strings.TrimSpace(parts[1])
			if strings.Contains(strings.ToLower(id), appLower) ||
				strings.Contains(strings.ToLower(name), appLower) {
				return id, scope
			}
		}
	}
	return "", ""
}
