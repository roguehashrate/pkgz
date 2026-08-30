package linux

import (
	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewAptSource returns a source backed by apt/apt-cache/dpkg (Debian/Ubuntu).
func NewAptSource(elevator *utils.Elevator) sources.Source {
	var c *commandSource
	c = &commandSource{
		name: "Apt",
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
		installedCount: countOutput("dpkg-query", "-f", ".\n", "-W"),

		installPrivileged: true,
		removePrivileged:  true,
		updatePrivileged:  true,
	}
	return c
}
