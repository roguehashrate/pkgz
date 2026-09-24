package linux

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// commandSource is a configurable implementation of sources.Source. Package
// managers supply small closures for each operation, delegating to the shared
// helpers below for the common patterns (contains-search, redirect-installed,
// first-field parsing, streaming execution).
type commandSource struct {
	name        string
	description string
	task        utils.Task

	available      func(app string) (bool, error)
	installed      func(app string) (bool, error)
	install        func(app string) error
	remove         func(app string) error
	update         func() error
	listUpdates    func() ([]string, error)
	search         func(app string) (bool, error)
	searchMatches  func(app string) ([]string, error)
	installedCount func() (int, error)

	// Privilege flags: whether each op escalates privileges (e.g. via
	// sudo/doas), which means the TUI must release the terminal so password
	// prompts work. Left false for sources that run as the plain user.
	installPrivileged bool
	removePrivileged  bool
	updatePrivileged  bool

	// memo caches per-run lookups so repeated probes within one invocation
	// (install -> available -> install, info, remove) don't re-run slow
	// commands like `apt-cache search` or `flatpak search` (multi-second each).
	memo struct {
		mu       sync.Mutex
		avail    map[string]memoEntry
		inst     map[string]memoEntry
		updates  []string
		updReady bool
	}
}

// memoEntry caches a (ok, err) probe result keyed by lower-cased app name.
type memoEntry struct {
	ok  bool
	err error
}

var _ sources.Source = (*commandSource)(nil)

func (c *commandSource) Name() string        { return c.name }
func (c *commandSource) Description() string { return c.description }
func (c *commandSource) Available(app string) (bool, error) {
	return c.memoizedAppProbe(&c.memo.avail, app, c.available)
}
func (c *commandSource) Installed(app string) (bool, error) {
	return c.memoizedAppProbe(&c.memo.inst, app, c.installed)
}
func (c *commandSource) Install(app string) error        { return c.install(app) }
func (c *commandSource) Remove(app string) error         { return c.remove(app) }
func (c *commandSource) Update() error                   { return c.update() }
func (c *commandSource) ListUpdates() ([]string, error)  { return c.memoizedUpdates() }
func (c *commandSource) Search(app string) (bool, error) { return c.search(app) }
func (c *commandSource) SearchMatches(app string) ([]string, error) {
	if c.searchMatches == nil {
		return nil, nil
	}
	return c.searchMatches(app)
}
func (c *commandSource) InstalledCount() (int, error) { return c.installedCount() }

// memoizedAppProbe runs probe the first time an app is asked about and returns
// the cached result afterwards.
func (c *commandSource) memoizedAppProbe(cache *map[string]memoEntry, app string, probe func(string) (bool, error)) (bool, error) {
	key := strings.ToLower(app)
	c.memo.mu.Lock()
	if *cache == nil {
		*cache = make(map[string]memoEntry)
	}
	if e, ok := (*cache)[key]; ok {
		c.memo.mu.Unlock()
		return e.ok, e.err
	}
	c.memo.mu.Unlock()

	ok, err := probe(app)
	c.memo.mu.Lock()
	(*cache)[key] = memoEntry{ok: ok, err: err}
	c.memo.mu.Unlock()
	return ok, err
}

// memoizedUpdates runs the update-list query once per invocation (it is stable
// within a run and can be slow) and replays the result for refresh/info flows.
func (c *commandSource) memoizedUpdates() ([]string, error) {
	c.memo.mu.Lock()
	if c.memo.updReady {
		c.memo.mu.Unlock()
		return c.memo.updates, nil
	}
	c.memo.mu.Unlock()

	updates, err := c.listUpdates()
	c.memo.mu.Lock()
	c.memo.updates = updates
	c.memo.updReady = true
	c.memo.mu.Unlock()
	return updates, err
}

// Privilege getters let the UI decide whether to run an op with direct terminal
// control (to accept sudo/doas passwords).
func (c *commandSource) InstallPrivileged() bool { return c.installPrivileged }
func (c *commandSource) RemovePrivileged() bool  { return c.removePrivileged }
func (c *commandSource) UpdatePrivileged() bool  { return c.updatePrivileged }

// SetTask attaches a reporting hook for streaming status/output.
func (c *commandSource) SetTask(t utils.Task) { c.task = t }

func (c *commandSource) reportLine(line string) {
	if c.task != nil {
		c.task.AppendOutput(line)
	}
}

// runOp executes a (possibly privileged) command. Both paths stream their
// combined output into the attached task so it is surfaced inside the TUI's log
// pane rather than dumped to the terminal. Privileged commands keep stdin
// attached (via the used streaming helper) so sudo/doas can still prompt for a
// password; in the TUI they run inside bubbletea's Exec, which releases the
// terminal for the duration of the prompt.
func (c *commandSource) runOp(e *utils.Elevator, privileged bool, bin string, args []string) error {
	if privileged {
		return e.RunPrivilegedStreaming(bin, args, c.reportLine)
	}
	return utils.RunCommandStreaming(bin, args, c.reportLine)
}

// availableContains builds an Available closure that reports whether `bin`
// output contains the app string (case-insensitive, like the PMs themselves).
func availableContains(bin string, args func(app string) []string) func(string) (bool, error) {
	return func(app string) (bool, error) {
		output, err := utils.RunCommand(bin, args(app)...)
		if err != nil {
			return false, nil
		}
		return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
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

// splitFields splits a search result line into its fields. Tab-delimited output
// (flatpak: "appid\tname") keeps the name intact; space-delimited output
// (apt-cache: "pkg - description") falls back to whitespace splitting.
func splitFields(line string) []string {
	if fields := strings.SplitN(line, "\t", 2); len(fields) == 2 {
		return []string{strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])}
	}
	return strings.Fields(line)
}

// matchLines parses tab/first-field separated search output, keeping at most
// maxResults lines where keep matches. format maps a line's fields to the
// human-friendly result string ("" to skip the line).
func matchLines(output string, format func([]string) string, keep func(string, []string) bool) []string {
	matches := make([]string, 0, 15)
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := splitFields(line)
		if len(fields) == 0 {
			continue
		}
		if keep != nil && !keep(line, fields) {
			continue
		}
		if m := format(fields); m != "" {
			matches = append(matches, m)
			if len(matches) >= 15 {
				break
			}
		}
	}
	return matches
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

// noUpdateNonzeroExit lists binaries whose update-list query exits non-zero when
// there is nothing to update (e.g. `pacman -Qu` returns 1 when up to date). For
// these, a non-empty exit code with no parseable output is a clean "no updates",
// not a failure.
var noUpdateNonzeroExit = map[string]bool{
	"pacman": true,
	"paru":   true,
	"yay":    true,
}

// listUpdatesCmd builds a ListUpdates closure running `bin args` and parsing
// the output with parse. Parseable packages win even when the command exits
// non-zero (dnf returns 100 when updates exist). A genuine command failure with
// no parseable output is surfaced as an error instead of being silently
// reported as "up to date".
func listUpdatesCmd(bin string, args []string, parse func(string) []string) func() ([]string, error) {
	return func() ([]string, error) {
		output, err := utils.RunCommand(bin, args...)
		updates := parse(output)
		if len(updates) > 0 {
			return updates, nil
		}
		if err != nil {
			if noUpdateNonzeroExit[bin] {
				return nil, nil
			}
			return nil, fmt.Errorf("%s failed: %w", bin, err)
		}
		return nil, nil
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

// aptStatusLines splits the `dpkg-query -W` output into status/package lines.
// The full installed list is queried once per invocation (it is stable) and
// reused by every installe/remove probe.
var aptStatusOnce struct {
	sync.Once
	lines []string
}

func aptStatusLines() []string {
	aptStatusOnce.Do(func() {
		out, err := utils.RunCommand("dpkg-query", "-W", "-f=${db:Status-Abbrev} ${Package}\n")
		if err != nil {
			return
		}
		aptStatusOnce.lines = strings.Split(strings.TrimSpace(out), "\n")
	})
	return aptStatusOnce.lines
}

// aptMatchPackages returns fully-installed package names that match app:
// exactly equal to app, or starting with "<app>-" / "<app>." (e.g. emacs is
// provided by emacs-gtk, emacs-common). This lets apt detect apps installed
// via their Debian subpackages rather than an exact metapackage. Matching is
// case-insensitive because apt package names are lowercase.
func aptMatchPackages(app string) []string {
	var matches []string
	lower := strings.ToLower(app)
	prefix := lower + "-"
	for _, line := range aptStatusLines() {
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "ii") {
			continue
		}
		pkg := strings.ToLower(fields[1])
		if pkg == lower || strings.HasPrefix(pkg, prefix) || strings.HasPrefix(pkg, lower+".") {
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
	if strings.ToLower(app) != app && utils.RunCommandWithRedirect("dpkg", "-s", strings.ToLower(app)) {
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
	if strings.ToLower(app) != app && utils.RunCommandWithRedirect("dpkg", "-s", strings.ToLower(app)) {
		return strings.ToLower(app)
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
