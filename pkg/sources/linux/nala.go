package linux

import (
	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewNalaSource returns a source backed by nala (a fancier apt frontend).
func NewNalaSource(elevator *utils.Elevator) sources.Source {
	var c *commandSource
	c = &commandSource{
		name: "Nala",
		available: availableContains("nala", func(app string) []string {
			return []string{"search", app}
		}),
		installed: func(app string) (bool, error) {
			return aptAppInstalled(app), nil
		},
		install: func(app string) error {
			return c.runOp(elevator, true, "nala", []string{"install", "-y", app})
		},
		remove: func(app string) error {
			return c.runOp(elevator, true, "nala", []string{"remove", "-y", aptRemoveTarget(app)})
		},
		update: func() error {
			return c.runOp(elevator, true, "sh", []string{"-c", "nala update && nala upgrade -y"})
		},
		listUpdates: listUpdatesCmd("apt", []string{"list", "--upgradable"}, parseAptUpgradable),
		search: searchContains("nala", func(app string) []string {
			return []string{"search", app}
		}),
		installedCount: countOutput("dpkg-query", "-f", ".\n", "-W"),

		installPrivileged: true,
		removePrivileged:  true,
		updatePrivileged:  true,
	}
	return c
}
