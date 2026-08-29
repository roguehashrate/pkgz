package linux

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// commandSource is a configurable implementation of sources.Source. Package
// managers supply small closures for each operation, delegating to the shared
// helpers below for the common patterns (contains-search, redirect-installed,
// first-field parsing, streaming execution).
type commandSource struct {
	name string
	task utils.Task

	available      func(app string) (bool, error)
	installed      func(app string) (bool, error)
	install        func(app string) error
	remove         func(app string) error
	update         func() error
	listUpdates    func() ([]string, error)
	search         func(app string) (bool, error)
	installedCount func() (int, error)
}

var _ sources.Source = (*commandSource)(nil)

func (c *commandSource) Name() string                       { return c.name }
func (c *commandSource) Available(app string) (bool, error) { return c.available(app) }
func (c *commandSource) Installed(app string) (bool, error) { return c.installed(app) }
func (c *commandSource) Install(app string) error           { return c.install(app) }
func (c *commandSource) Remove(app string) error            { return c.remove(app) }
func (c *commandSource) Update() error                      { return c.update() }
func (c *commandSource) ListUpdates() ([]string, error)     { return c.listUpdates() }
func (c *commandSource) Search(app string) (bool, error)    { return c.search(app) }
func (c *commandSource) InstalledCount() (int, error)       { return c.installedCount() }

// SetTask attaches a reporting hook for streaming status/output.
func (c *commandSource) SetTask(t utils.Task) { c.task = t }

func (c *commandSource) reportLine(line string) {
	if c.task != nil {
		c.task.AppendOutput(line)
	}
}

// runOp executes a (possibly privileged) command, streaming its output into the
// attached task. When privileged, stdin is forwarded for sudo/doas prompts.
func (c *commandSource) runOp(e *utils.Elevator, privileged bool, bin string, args []string) error {
	if privileged {
		return e.RunPrivilegedStreaming(bin, args, c.reportLine)
	}
	return utils.RunCommandStreaming(bin, args, c.reportLine)
}

// availableContains builds an Available closure that reports whether `bin`
// output contains the app string (case-sensitive, as the original did).
func availableContains(bin string, args func(app string) []string) func(string) (bool, error) {
	return func(app string) (bool, error) {
		output, err := utils.RunCommand(bin, args(app)...)
		if err != nil {
			return false, nil
		}
		return strings.Contains(output, app), nil
	}
}

// searchContains builds a Search closure (case-insensitive contains).
func searchContains(bin string, args func(app string) []string) func(string) (bool, error) {
	return func(app string) (bool, error) {
		output, err := utils.RunCommand(bin, args(app)...)
		if err != nil {
			return false, nil
		}
		return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
	}
}

// installedRedirect builds an Installed closure that checks exit code with
// stdout/stderr discarded.
func installedRedirect(bin string, args func(app string) []string) func(string) (bool, error) {
	return func(app string) (bool, error) {
		return utils.RunCommandWithRedirect(bin, args(app)...), nil
	}
}

// countOutput builds an InstalledCount closure returning the number of lines.
func countOutput(bin string, args ...string) func() (int, error) {
	return func() (int, error) {
		lines, err := utils.GetCommandOutput(bin, args...)
		if err != nil {
			return 0, nil
		}
		return len(lines), nil
	}
}

// listFirstField parses an update list as the first whitespace-delimited field
// of each line, skipping lines starting with any of the given prefixes.
func listFirstField(output string, skip ...string) []string {
	var updates []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		skipLine := false
		for _, p := range skip {
			if strings.HasPrefix(line, p) {
				skipLine = true
				break
			}
		}
		if skipLine {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		updates = append(updates, fields[0])
	}
	return updates
}

// listUpdatesCmd builds a ListUpdates closure running `bin args` and parsing
// the output with parse.
func listUpdatesCmd(bin string, args []string, parse func(string) []string) func() ([]string, error) {
	return func() ([]string, error) {
		output, err := utils.RunCommand(bin, args...)
		if err != nil && strings.TrimSpace(output) == "" {
			return nil, nil
		}
		return parse(output), nil
	}
}

// newAurSource builds a source for an AUR helper (paru/yay). The two helpers
// share identical command syntax, differing only by binary, and run without
// privilege elevation.
func newAurSource(elevator *utils.Elevator, name, binary string) sources.Source {
	var c *commandSource
	c = &commandSource{
		name: name,
		available: availableContains(binary, func(app string) []string {
			return []string{"-Ss", app}
		}),
		installed: installedRedirect(binary, func(app string) []string {
			return []string{"-Qm", app}
		}),
		install: func(app string) error {
			return c.runOp(elevator, false, binary, []string{"-S", "--noconfirm", app})
		},
		remove: func(app string) error {
			return c.runOp(elevator, false, binary, []string{"-R", "--noconfirm", app})
		},
		update: func() error {
			return c.runOp(elevator, false, binary, []string{"-Syu", "--noconfirm"})
		},
		listUpdates: listUpdatesCmd(binary, []string{"-Qua"}, func(output string) []string {
			return listFirstField(output, "::", "warning", "error")
		}),
		search: searchContains(binary, func(app string) []string {
			return []string{"-Ss", app}
		}),
		installedCount: countOutput(binary, "-Qm"),
	}
	return c
}

// aptStatusPair splits the `dpkg-query -W` output line into status and package.
func aptStatusLines() []string {
	out, err := utils.RunCommand("dpkg-query", "-W", "-f=${db:Status-Abbrev} ${Package}\n")
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSpace(out), "\n")
}

// aptMatchPackages returns fully-installed package names that match app:
// exactly equal to app, or starting with "<app>-" / "<app>." (e.g. emacs is
// provided by emacs-gtk, emacs-common). This lets apt detect apps installed
// via their Debian subpackages rather than an exact metapackage.
func aptMatchPackages(app string) []string {
	var matches []string
	prefix := app + "-"
	for _, line := range aptStatusLines() {
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "ii") {
			continue
		}
		pkg := fields[1]
		if pkg == app || strings.HasPrefix(pkg, prefix) || strings.HasPrefix(pkg, app+".") {
			matches = append(matches, pkg)
		}
	}
	return matches
}

// aptAppInstalled reports whether an apt package matching app is fully
// installed, either exactly or via an installed subpackage.
func aptAppInstalled(app string) bool {
	if utils.RunCommandWithRedirect("dpkg", "-s", app) {
		return true
	}
	return len(aptMatchPackages(app)) > 0
}

// aptRemoveTarget resolves the concrete installed package to remove for app:
// the exact metapackage when installed; otherwise the package that owns the
// app's binary (e.g. emacs -> emacs-gtk via dpkg -S /usr/bin/emacs); otherwise
// the shortest matching subpackage name; falling back to app.
func aptRemoveTarget(app string) string {
	if utils.RunCommandWithRedirect("dpkg", "-s", app) {
		return app
	}
	matches := aptMatchPackages(app)
	if len(matches) == 0 {
		return app
	}

	// Prefer the package that owns the app's real binary, since that is the
	// concrete application package the user expects to remove (emacs-gtk, not
	// emacs-bin-common or emacs-common).
	if bin, err := exec.LookPath(app); err == nil {
		if abs, aerr := filepath.EvalSymlinks(bin); aerr == nil {
			bin = abs
		}
		if out, serr := utils.RunCommand("dpkg", "-S", bin); serr == nil {
			if owner := aptBinaryOwner(out); owner != "" {
				for _, m := range matches {
					if m == owner {
						return owner
					}
				}
			}
		}
	}

	// Degradation: remove the name closest to the app (shortest subpackage).
	sort.Slice(matches, func(i, j int) bool { return len(matches[i]) < len(matches[j]) })
	return matches[0]
}

// aptBinaryOwner extracts the owning package from `dpkg -S <path>` output,
// which is a comma-separated package list then ": <path>".
func aptBinaryOwner(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, ":") {
			continue
		}
		first := strings.SplitN(line, ":", 2)[0]
		first = strings.TrimSpace(first)
		if pkg := strings.Split(first, ",")[0]; pkg != "" {
			return pkg
		}
	}
	return ""
}

// parseAptUpgradable parses `apt list --upgradable` output into package names.
func parseAptUpgradable(output string) []string {
	var updates []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "WARNING") ||
			strings.HasPrefix(line, "Notice") || strings.HasPrefix(line, "Listing") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || !strings.Contains(fields[0], "/") {
			continue
		}
		name := fields[0][:strings.Index(fields[0], "/")]
		updates = append(updates, name)
	}
	return updates
}
