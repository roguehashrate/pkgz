package linux

import (
	"strings"
	"sync"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewFlatpakSource returns a source backed by flatpak (cross-distro).
func NewFlatpakSource(elevator *utils.Elevator) sources.Source {
	var c *commandSource
	c = &commandSource{
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
				return c.runOp(elevator, false, "flatpak", []string{"install", "--user", "-y", "flathub", app})
			}
			return c.runOp(elevator, false, "flatpak", []string{"install", "--user", "-y", "flathub", appID})
		},
		remove: func(app string) error {
			appID, scope := flatpakInstalledApp(app)
			if appID == "" {
				appID = app
				scope = "--user"
			}
			return c.runOp(elevator, false, "flatpak", []string{"uninstall", scope, "-y", appID})
		},
		update: func() error {
			return c.runOp(elevator, false, "flatpak", []string{"update", "--user", "-y"})
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
		searchMatches: func(app string) ([]string, error) {
			output, err := utils.RunCommand("flatpak", "search", "--columns=application,name", app)
			if err != nil {
				return nil, nil
			}
			return matchLines(output, func(fields []string) string {
				if len(fields) == 0 || fields[0] == "" {
					return ""
				}
				id := fields[0]
				if len(fields) > 1 {
					return id + " (" + fields[1] + ")"
				}
				return id
			}, func(line string, fields []string) bool {
				return containsAny(fields, strings.ToLower(app))
			}), nil
		},
		installedCount: countOutput("flatpak", "list", "--user", "--app"),
	}
	return c
}

// flatpakMemo guards the per-run lookup caches below. One-shot CLI: results stay
// valid for the process lifetime.
var flatpakMemo struct {
	mu       sync.Mutex
	appID    map[string]string // lowercase app -> resolved search appID ("" = not found)
	insID    map[string]string // lowercase app -> installed appID ("" = not installed)
	insScope map[string]string // lowercase app -> installed scope ("--user"/"--system")
}

func flatpakFindAppID(app string) (string, error) {
	key := strings.ToLower(app)
	flatpakMemo.mu.Lock()
	if flatpakMemo.appID == nil {
		flatpakMemo.appID = make(map[string]string)
	}
	if id, ok := flatpakMemo.appID[key]; ok {
		flatpakMemo.mu.Unlock()
		return id, nil
	}
	flatpakMemo.mu.Unlock()

	output, err := utils.RunCommand("flatpak", "search", "--columns=application,name", app)
	if err != nil {
		flatpakMemo.mu.Lock()
		flatpakMemo.appID[key] = ""
		flatpakMemo.mu.Unlock()
		return "", err
	}

	lines := strings.Split(output, "\n")
	appLower := strings.ToLower(app)

	var id string
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		cand := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		if cand == appLower {
			id = cand
			break
		}
		if id == "" && (strings.Contains(strings.ToLower(name), appLower) ||
			strings.Contains(strings.ToLower(cand), appLower)) {
			id = cand
		}
	}
	flatpakMemo.mu.Lock()
	flatpakMemo.appID[key] = id
	flatpakMemo.mu.Unlock()
	return id, nil
}

// flatpakInstalledApp returns the App ID and scope (--user or --system) of an
// installed flatpak matching app, or empty strings if it is not installed.
func flatpakInstalledApp(app string) (appID, scope string) {
	key := strings.ToLower(app)
	flatpakMemo.mu.Lock()
	if flatpakMemo.insID == nil {
		flatpakMemo.insID = make(map[string]string)
		flatpakMemo.insScope = make(map[string]string)
	}
	if id, ok := flatpakMemo.insID[key]; ok {
		sc := flatpakMemo.insScope[key]
		flatpakMemo.mu.Unlock()
		return id, sc
	}
	flatpakMemo.mu.Unlock()

	appLower := strings.ToLower(app)
	for _, sc := range []string{"--user", "--system"} {
		output, err := utils.RunCommand("flatpak", "list", sc, "--columns=application,name")
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
			if id == appLower || id != "" &&
				(strings.Contains(strings.ToLower(id), appLower) ||
					strings.Contains(strings.ToLower(name), appLower)) {
				flatpakMemo.mu.Lock()
				flatpakMemo.insID[key] = id
				flatpakMemo.insScope[key] = sc
				flatpakMemo.mu.Unlock()
				return id, sc
			}
		}
	}
	flatpakMemo.mu.Lock()
	flatpakMemo.insID[key] = ""
	flatpakMemo.insScope[key] = ""
	flatpakMemo.mu.Unlock()
	return "", ""
}

// containsAny reports whether any field contains the lower-cased needle.
func containsAny(fields []string, needle string) bool {
	for _, f := range fields {
		if strings.Contains(f, needle) {
			return true
		}
	}
	return false
}
