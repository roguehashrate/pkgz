package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewAptSource returns a source backed by apt/apt-cache/dpkg (Debian/Ubuntu).
func NewAptSource(elevator *utils.Elevator) sources.Source {
	var c *commandSource
	c = &commandSource{
		name:        "Apt",
		description: "native Debian/Ubuntu packages",
		available: availableContains("apt-cache", func(app string) []string {
			return []string{"search", app}
		}),
		installed: func(app string) (bool, error) {
			return aptAppInstalled(app), nil
		},
		install: func(app string) error {
			return c.runOp(elevator, true, "apt", []string{"install", "-y", app})
		},
		remove: func(app string) error {
			return c.runOp(elevator, true, "apt", []string{"remove", "-y", aptRemoveTarget(app)})
		},
		update: func() error {
			return c.runOp(elevator, true, "sh", []string{"-c", "apt update && apt upgrade -y"})
		},
		listUpdates: listUpdatesCmd("apt", []string{"list", "--upgradable"}, parseAptUpgradable),
		search: searchContains("apt-cache", func(app string) []string {
			return []string{"search", app}
		}),
		searchMatches: func(app string) ([]string, error) {
			output, err := utils.RunCommand("apt-cache", "search", app)
			if err != nil {
				return nil, err
			}
			return matchLines(output, func(fields []string) string {
				return fields[0]
			}, func(line string, fields []string) bool {
				return strings.Contains(strings.ToLower(line), strings.ToLower(app))
			}), nil
		},
		installedCount: countOutput("dpkg-query", "-f", ".\n", "-W"),

		installPrivileged: true,
		removePrivileged:  true,
		updatePrivileged:  true,
	}
	return c
}
