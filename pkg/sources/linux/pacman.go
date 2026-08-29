package linux

import (
	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewPacmanSource returns a source backed by pacman (Arch).
func NewPacmanSource(elevator *utils.Elevator) sources.Source {
	var c *commandSource
	c = &commandSource{
		name: "Pacman",
		available: availableContains("pacman", func(app string) []string {
			return []string{"-Ss", app}
		}),
		installed: installedRedirect("pacman", func(app string) []string {
			return []string{"-Qn", app}
		}),
		install: func(app string) error {
			return c.runOp(elevator, true, "pacman", []string{"-S", "--noconfirm", app})
		},
		remove: func(app string) error {
			return c.runOp(elevator, true, "pacman", []string{"-R", "--noconfirm", app})
		},
		update: func() error {
			return c.runOp(elevator, true, "pacman", []string{"-Syu", "--noconfirm"})
		},
		listUpdates: listUpdatesCmd("pacman", []string{"-Qu"}, func(output string) []string {
			return listFirstField(output)
		}),
		search: searchContains("pacman", func(app string) []string {
			return []string{"-Ss", app}
		}),
		installedCount: countOutput("pacman", "-Qn"),
	}
	return c
}
