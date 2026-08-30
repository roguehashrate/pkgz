package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewZypperSource returns a source backed by zypper/rpm (openSUSE).
func NewZypperSource(elevator *utils.Elevator) sources.Source {
	var c *commandSource
	c = &commandSource{
		name: "Zypper",
		available: availableContains("zypper", func(app string) []string {
			return []string{"search", app}
		}),
		installed: installedRedirect("rpm", func(app string) []string {
			return []string{"-q", app}
		}),
		install: func(app string) error {
			return c.runOp(elevator, true, "zypper", []string{"install", "-y", app})
		},
		remove: func(app string) error {
			return c.runOp(elevator, true, "zypper", []string{"remove", "-y", app})
		},
		update: func() error {
			return c.runOp(elevator, true, "sh", []string{"-c", "zypper refresh && zypper update -y"})
		},
		listUpdates: listUpdatesCmd("zypper", []string{"list-updates", "--no-refresh"}, parseZypperUpdates),
		search: searchContains("zypper", func(app string) []string {
			return []string{"search", app}
		}),
		installedCount: countOutput("rpm", "-qa"),

		installPrivileged: true,
		removePrivileged:  true,
		updatePrivileged:  true,
	}
	return c
}

func parseZypperUpdates(output string) []string {
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
	return updates
}
