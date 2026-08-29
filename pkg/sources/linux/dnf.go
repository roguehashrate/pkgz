package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewDnfSource returns a source backed by dnf/rpm (Fedora/RHEL).
func NewDnfSource(elevator *utils.Elevator) sources.Source {
	var c *commandSource
	c = &commandSource{
		name: "DNF",
		available: availableContains("dnf", func(app string) []string {
			return []string{"search", app}
		}),
		installed: installedRedirect("rpm", func(app string) []string {
			return []string{"-q", app}
		}),
		install: func(app string) error {
			return c.runOp(elevator, true, "dnf", []string{"install", "-y", app})
		},
		remove: func(app string) error {
			return c.runOp(elevator, true, "dnf", []string{"remove", "-y", app})
		},
		update: func() error {
			return c.runOp(elevator, true, "dnf", []string{"upgrade", "-y"})
		},
		listUpdates: listUpdatesCmd("dnf", []string{"check-update"}, parseDnfUpdates),
		search: searchContains("dnf", func(app string) []string {
			return []string{"search", app}
		}),
		installedCount: countOutput("rpm", "-qa"),
	}
	return c
}

func parseDnfUpdates(output string) []string {
	var updates []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Last metadata") ||
			strings.HasPrefix(line, "Upgrade") || strings.HasPrefix(line, "Downgrade") ||
			strings.HasPrefix(line, "Installing") || strings.HasPrefix(line, "Removing") ||
			strings.HasPrefix(line, "Package") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		updates = append(updates, fields[0])
	}
	return updates
}
