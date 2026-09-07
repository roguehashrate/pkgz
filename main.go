package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/roguehashrate/pkgz/pkg/config"
	"github.com/roguehashrate/pkgz/pkg/sources/linux"
	"github.com/roguehashrate/pkgz/pkg/tui"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

const VERSION = "1.3.0"

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "--version", "-v":
		fmt.Printf("pkgz version %s\n", VERSION)
		return
	case "help", "-h", "--help":
		usage(os.Stdout)
		return
	}

	// Load configuration (auto-creates a default on first run).
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Resolve the privilege elevator and validate it up front so privileged
	// operations fail with a clear message instead of mid-run.
	elevator := utils.NewElevator()
	elevator.SetCommand(cfg.Elevator.Command)
	if _, err := elevator.GetElevatorCommand(""); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Install sudo or doas, or set [elevator] command in ~/.config/pkgz/config.toml.")
		os.Exit(1)
	}

	// Build sources from the enabled config, skipping any whose binary is no
	// longer present so they cannot report phantom "up to date" results.
	present, missing := cfg.EnabledSources()
	for _, src := range missing {
		fmt.Fprintf(os.Stderr, "⚠️  \"%s\" is enabled in config but \"%s\" is not installed — skipping it.\n", src, config.SourceBinaries[src])
	}

	sources := buildSources(present, elevator)
	if len(sources) == 0 {
		fmt.Fprintln(os.Stderr, "No usable sources. Enable at least one package manager in ~/.config/pkgz/config.toml and make sure it is installed.")
		os.Exit(1)
	}

	forceName, pkgs := parseArgs(os.Args[2:])

	var runErr error
	switch command {
	case "install":
		runErr = handleInstall(pkgs, forceName, sources)
	case "remove":
		runErr = handleRemove(pkgs, forceName, sources)
	case "update":
		runErr = handleUpdate(sources)
	case "refresh":
		runErr = handleRefresh(sources)
	case "search":
		runErr = handleSearch(pkgs, forceName, sources)
	case "info":
		runErr = handleInfo(pkgs, sources)
	case "clean":
		runErr = handleClean(sources, elevator)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		usage(os.Stderr)
		os.Exit(1)
	}

	if runErr != nil {
		fmt.Fprintln(os.Stderr, runErr)
		os.Exit(1)
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "pkgz — one command to install, remove and update packages")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  pkgz install [--source NAME] PACKAGE...")
	fmt.Fprintln(w, "  pkgz remove  [--source NAME] PACKAGE...")
	fmt.Fprintln(w, "  pkgz update                 apply all available updates")
	fmt.Fprintln(w, "  pkgz refresh                check for updates without applying")
	fmt.Fprintln(w, "  pkgz search [--source NAME] PACKAGE...")
	fmt.Fprintln(w, "  pkgz info [PACKAGE]         package status, or installed counts")
	fmt.Fprintln(w, "  pkgz clean")
	fmt.Fprintln(w, "  pkgz --version")
}

// parseArgs extracts a --source/--from flag and the remaining positional args.
func parseArgs(args []string) (forceName string, pkgs []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--source" || a == "--from":
			if i+1 < len(args) {
				i++
				forceName = args[i]
			}
		case strings.HasPrefix(a, "--source="):
			forceName = strings.TrimPrefix(a, "--source=")
		case strings.HasPrefix(a, "--from="):
			forceName = strings.TrimPrefix(a, "--from=")
		case strings.HasPrefix(a, "-") && a != "-":
			continue
		default:
			pkgs = append(pkgs, a)
		}
	}
	return forceName, pkgs
}

// buildSources constructs Source instances for the given (present) source names.
func buildSources(names []string, elevator *utils.Elevator) []Source {
	sources := make([]Source, 0, len(names))
	for _, name := range names {
		switch name {
		case "apt":
			sources = append(sources, linux.NewAptSource(elevator))
		case "flatpak":
			sources = append(sources, linux.NewFlatpakSource(elevator))
		case "pacman":
			sources = append(sources, linux.NewPacmanSource(elevator))
		case "paru":
			sources = append(sources, linux.NewParuSource(elevator))
		case "yay":
			sources = append(sources, linux.NewYaySource(elevator))
		case "dnf":
			sources = append(sources, linux.NewDnfSource(elevator))
		case "zypper":
			sources = append(sources, linux.NewZypperSource(elevator))
		}
	}
	return sources
}

// Source interface to match the package sources
type Source interface {
	Name() string
	Available(app string) (bool, error)
	Installed(app string) (bool, error)
	Install(app string) error
	Remove(app string) error
	Update() error
	ListUpdates() ([]string, error)
	Search(app string) (bool, error)
	InstalledCount() (int, error)
}

// privSource is implemented by sources that can report whether each of their
// operations escalates privileges (requiring sudo/doas), so the TUI knows it
// must release the terminal to accept a password prompt.
type privSource interface {
	InstallPrivileged() bool
	RemovePrivileged() bool
	UpdatePrivileged() bool
}

func installPrivileged(s Source) bool {
	if p, ok := s.(privSource); ok {
		return p.InstallPrivileged()
	}
	return false
}

func removePrivileged(s Source) bool {
	if p, ok := s.(privSource); ok {
		return p.RemovePrivileged()
	}
	return false
}

func updatePrivileged(s Source) bool {
	if p, ok := s.(privSource); ok {
		return p.UpdatePrivileged()
	}
	return false
}

// isTerminal reports whether stdout is a TTY (used to pick TUI vs plain output).
func isTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

// stdinTerminal reports whether stdin is a TTY (required for interactive prompts).
func stdinTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
}

// withTask attaches a reporting task to a source if it supports it, so the
// source's streaming output is surfaced in the TUI.
func withTask(src Source, t utils.Task) Source {
	if s, ok := src.(interface{ SetTask(utils.Task) }); ok {
		s.SetTask(t)
	}
	return src
}

// runOps runs operations through the TUI when stdout is a terminal and falls
// back to plain sequential output otherwise. Returns the first operation error.
func runOps(title string, ops []tui.Op) error {
	if isTerminal() {
		if opErr, progErr := tui.RunAny(title, "", nil, ops); progErr == nil {
			return opErr
		}
	}
	return tui.RunPlain(ops)
}

// --- install / remove -------------------------------------------------------

func handleInstall(apps []string, forceName string, sources []Source) error {
	if len(apps) == 0 {
		return fmt.Errorf("usage: pkgz install [--source NAME] PACKAGE...")
	}

	if forceName != "" {
		forced := matchSource(sources, forceName)
		if forced == nil {
			return fmt.Errorf("unknown or unavailable source %q (enabled: %s)", forceName, sourceNames(sources))
		}
		return installAll(apps, forced)
	}

	if len(apps) == 1 {
		return handleInstallOne(apps[0], sources)
	}

	if shared, ok := commonSingleSource(apps, sources); ok {
		return installAll(apps, shared)
	}

	var firstErr error
	for _, app := range apps {
		if err := handleInstallOne(app, sources); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func handleInstallOne(app string, sources []Source) error {
	fmt.Printf("🔍 Searching for '%s' in sources...\n", app)
	srcs := availableSourcesFor(app, sources)
	if len(srcs) == 0 {
		return fmt.Errorf("❌ App '%s' not found in any enabled source.", app)
	}
	if len(srcs) == 1 {
		fmt.Printf("✅ Found '%s' in %s.\n", app, srcs[0].Name())
	}
	return runPick("install", app, srcs, func(src Source) tui.Op {
		return tui.Op{
			Label:      "Installing " + app + " via " + src.Name(),
			Privileged: installPrivileged(src),
			Run: func(t *tui.Task) error {
				return withTask(src, t).Install(app)
			},
		}
	})
}

func installAll(apps []string, src Source) error {
	ops := make([]tui.Op, 0, len(apps))
	for _, app := range apps {
		app := app
		ops = append(ops, tui.Op{
			Label:      "Installing " + app + " via " + src.Name(),
			Privileged: installPrivileged(src),
			Run: func(t *tui.Task) error {
				return withTask(src, t).Install(app)
			},
		})
	}
	return runOps("pkgz install", ops)
}

func handleRemove(apps []string, forceName string, sources []Source) error {
	if len(apps) == 0 {
		return fmt.Errorf("usage: pkgz remove [--source NAME] PACKAGE...")
	}

	if forceName != "" {
		forced := matchSource(sources, forceName)
		if forced == nil {
			return fmt.Errorf("unknown or unavailable source %q (enabled: %s)", forceName, sourceNames(sources))
		}
		return removeAll(apps, forced)
	}

	if len(apps) == 1 {
		return handleRemoveOne(apps[0], sources)
	}

	if shared, ok := commonInstalledSource(apps, sources); ok {
		return removeAll(apps, shared)
	}

	var firstErr error
	for _, app := range apps {
		if err := handleRemoveOne(app, sources); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func handleRemoveOne(app string, sources []Source) error {
	srcs := installedSourcesFor(app, sources)
	if len(srcs) == 0 {
		return fmt.Errorf("❌ '%s' is not installed in any enabled source.", app)
	}
	if len(srcs) == 1 {
		fmt.Printf("🗑️ Found '%s' in %s.\n", app, srcs[0].Name())
	}
	return runPick("remove", app, srcs, func(src Source) tui.Op {
		return tui.Op{
			Label:      "Removing " + app + " via " + src.Name(),
			Privileged: removePrivileged(src),
			Run: func(t *tui.Task) error {
				return withTask(src, t).Remove(app)
			},
		}
	})
}

func removeAll(apps []string, src Source) error {
	ops := make([]tui.Op, 0, len(apps))
	for _, app := range apps {
		app := app
		ops = append(ops, tui.Op{
			Label:      "Removing " + app + " via " + src.Name(),
			Privileged: removePrivileged(src),
			Run: func(t *tui.Task) error {
				return withTask(src, t).Remove(app)
			},
		})
	}
	return runOps("pkgz remove", ops)
}

// runPick picks one source when more than one candidate exists, then runs its
// operation. A single candidate runs directly. In a terminal the picker is the
// TUI; on an interactive non-TTY stdin it is a numbered prompt; otherwise it
// fails with guidance to use --source instead of hanging on a pipe.
func runPick(verb, app string, sources []Source, build func(Source) tui.Op) error {
	if len(sources) == 1 {
		return runOps("pkgz "+verb, []tui.Op{build(sources[0])})
	}

	choices := make([]string, len(sources))
	ops := make([]tui.Op, len(sources))
	for i, src := range sources {
		choices[i] = src.Name()
		ops[i] = build(src)
	}

	if isTerminal() {
		prompt := fmt.Sprintf("'%s' is available via multiple sources. Choose one:", app)
		if opErr, progErr := tui.RunAny("pkgz "+verb, prompt, choices, ops); progErr == nil {
			return opErr
		}
	}

	if stdinTerminal() {
		fmt.Printf("⚠️ '%s' is available via multiple sources:\n", app)
		for i, src := range sources {
			fmt.Printf("%d. %s\n", i+1, src.Name())
		}
		fmt.Printf("Which one would you like to use? [1-%d]: ", len(sources))
		input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(sources) {
			fmt.Println("❌ Invalid choice.")
			return fmt.Errorf("invalid source selection for '%s'", app)
		}
		return runOps("pkgz "+verb, []tui.Op{ops[choice-1]})
	}

	fmt.Fprintf(os.Stderr, "⚠️ '%s' is available via multiple sources: %s\n", app, strings.Join(choices, ", "))
	return fmt.Errorf("select a source explicitly with --source NAME for '%s'", app)
}

// availableSourcesFor returns the sources where the app is available.
func availableSourcesFor(app string, sources []Source) []Source {
	var out []Source
	for _, s := range sources {
		if ok, _ := s.Available(app); ok {
			out = append(out, s)
		}
	}
	return out
}

// installedSourcesFor returns the sources where the app is installed.
func installedSourcesFor(app string, sources []Source) []Source {
	var out []Source
	for _, s := range sources {
		if ok, _ := s.Installed(app); ok {
			out = append(out, s)
		}
	}
	return out
}

// commonSingleSource returns the single source in which every app is available,
// if exactly one such common source exists.
func commonSingleSource(apps []string, sources []Source) (Source, bool) {
	var shared []Source
	for i, app := range apps {
		srcs := availableSourcesFor(app, sources)
		if i == 0 {
			shared = srcs
		} else {
			shared = intersect(shared, srcs)
		}
	}
	if len(shared) == 1 {
		return shared[0], true
	}
	return nil, false
}

// commonInstalledSource is commonSingleSource for installed apps.
func commonInstalledSource(apps []string, sources []Source) (Source, bool) {
	var shared []Source
	for i, app := range apps {
		srcs := installedSourcesFor(app, sources)
		if i == 0 {
			shared = srcs
		} else {
			shared = intersect(shared, srcs)
		}
	}
	if len(shared) == 1 {
		return shared[0], true
	}
	return nil, false
}

func intersect(a, b []Source) []Source {
	var out []Source
	for _, x := range a {
		for _, y := range b {
			if x == y {
				out = append(out, x)
				break
			}
		}
	}
	return out
}

// matchSource resolves a user-supplied source name against the enabled sources,
// accepting names like "apt", "Apt", "paru", "Paru (AUR)".
func matchSource(sources []Source, name string) Source {
	key := strings.ToLower(strings.TrimSpace(name))
	for _, s := range sources {
		lower := strings.ToLower(s.Name())
		if lower == key {
			return s
		}
		if strings.Contains(lower, key) {
			return s
		}
	}
	return nil
}

func sourceNames(sources []Source) string {
	names := make([]string, len(sources))
	for i, s := range sources {
		names[i] = s.Name()
	}
	return strings.Join(names, ", ")
}

// --- update / refresh -------------------------------------------------------

func handleUpdate(sources []Source) error {
	ops := make([]tui.Op, 0, len(sources))
	for _, src := range sources {
		src := src
		ops = append(ops, tui.Op{
			Label:      "Updating " + src.Name(),
			Privileged: updatePrivileged(src),
			Run: func(t *tui.Task) error {
				return withTask(src, t).Update()
			},
		})
	}
	return runOps("pkgz update", ops)
}

func handleRefresh(sources []Source) error {
	ops := refreshOps(sources)
	if isTerminal() {
		if opErr, progErr := tui.RunAny("pkgz refresh", "", nil, ops); progErr == nil {
			return opErr
		}
	}
	return refreshPlain(sources)
}

func refreshOps(sources []Source) []tui.Op {
	ops := make([]tui.Op, 0, len(sources))
	for _, src := range sources {
		src := src
		ops = append(ops, tui.Op{
			Label: "Checking " + src.Name() + " for updates",
			Run: func(t *tui.Task) error {
				updates, err := src.ListUpdates()
				if err != nil {
					t.SetStatus("failed")
					t.SetLabel("Checking " + src.Name() + " — check failed")
					t.AppendOutput("✗ update check failed: " + err.Error())
					return fmt.Errorf("%s: %w", src.Name(), err)
				}
				if len(updates) == 0 {
					t.SetStatus("done")
					t.SetLabel("Checking " + src.Name() + " — up to date")
					t.AppendOutput("No updates available.")
					return nil
				}
				t.SetStatus("updates")
				t.SetLabel(fmt.Sprintf("Checking %s — %d update(s)", src.Name(), len(updates)))
				t.AppendOutput(fmt.Sprintf("%d update(s) available:", len(updates)))
				for _, pkg := range updates {
					t.AppendOutput("  - " + pkg)
				}
				return nil
			},
		})
	}
	return ops
}

// refreshPlain prints the per-source update check result without a terminal.
// Returns an error if any source's update check genuinely failed.
func refreshPlain(sources []Source) error {
	if len(sources) == 0 {
		fmt.Println("No sources enabled.")
		return nil
	}
	var failed bool
	for _, src := range sources {
		updates, err := src.ListUpdates()
		if err != nil {
			fmt.Printf("❌ %s: update check failed: %v\n", src.Name(), err)
			failed = true
			continue
		}
		if len(updates) == 0 {
			fmt.Printf("✓ %s: up to date\n", src.Name())
			continue
		}
		fmt.Printf("▲ %s: %d update(s) available\n", src.Name(), len(updates))
		for _, pkg := range updates {
			fmt.Printf("    - %s\n", pkg)
		}
	}
	if failed {
		return fmt.Errorf("one or more update checks failed")
	}
	return nil
}

// --- search -----------------------------------------------------------------

func handleSearch(apps []string, forceName string, sources []Source) error {
	if len(apps) == 0 {
		return fmt.Errorf("usage: pkgz search [--source NAME] PACKAGE...")
	}
	if len(apps) > 1 {
		var firstErr error
		for _, app := range apps {
			if err := handleSearchOne(app, forceName, sources); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
	return handleSearchOne(apps[0], forceName, sources)
}

func handleSearchOne(app string, forceName string, sources []Source) error {
	search := sources
	if forceName != "" {
		forced := matchSource(sources, forceName)
		if forced == nil {
			return fmt.Errorf("unknown or unavailable source %q (enabled: %s)", forceName, sourceNames(sources))
		}
		search = []Source{forced}
	}

	ops := make([]tui.Op, 0, len(search))
	for _, src := range search {
		src := src
		ops = append(ops, tui.Op{
			Label: "Searching " + src.Name(),
			Run: func(t *tui.Task) error {
				found, err := src.Search(app)
				if err != nil {
					t.SetStatus("failed")
					t.SetLabel(src.Name() + " — search failed")
					t.AppendOutput("✗ search failed: " + err.Error())
					return fmt.Errorf("%s: %w", src.Name(), err)
				}
				if found {
					t.SetStatus("done")
					t.SetLabel("Found in " + src.Name())
					t.AppendOutput(fmt.Sprintf("'%s' is available via %s.", app, src.Name()))
					return nil
				}
				t.SetStatus("done")
				t.SetLabel("Not found in " + src.Name())
				t.AppendOutput(fmt.Sprintf("'%s' was not found in %s.", app, src.Name()))
				return nil
			},
		})
	}

	if isTerminal() {
		if opErr, progErr := tui.RunAny("pkgz search "+app, "", nil, ops); progErr == nil {
			return opErr
		}
	}
	return searchPlain(app, search)
}

// searchPlain prints the per-source search result without a terminal.
func searchPlain(app string, sources []Source) error {
	var failed bool
	foundAny := false
	for _, src := range sources {
		found, err := src.Search(app)
		if err != nil {
			fmt.Printf("❌ %s: search failed: %v\n", src.Name(), err)
			failed = true
			continue
		}
		if found {
			fmt.Printf("✅ Found in %s\n", src.Name())
			foundAny = true
		} else {
			fmt.Printf("— Not found in %s\n", src.Name())
		}
	}
	if !foundAny && !failed {
		fmt.Printf("📦 Package '%s' not found in any enabled source.\n", app)
	}
	if failed {
		return fmt.Errorf("one or more searches failed")
	}
	return nil
}

// --- info -------------------------------------------------------------------

func handleInfo(apps []string, sources []Source) error {
	if len(apps) == 0 {
		return handleInfoCounts(sources)
	}
	if len(apps) > 1 {
		var firstErr error
		for _, app := range apps {
			if err := handleInfoOne(app, sources); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
	return handleInfoOne(apps[0], sources)
}

func handleInfoOne(app string, sources []Source) error {
	ops := make([]tui.Op, 0, len(sources))
	for _, src := range sources {
		src := src
		ops = append(ops, tui.Op{
			Label: src.Name(),
			Run: func(t *tui.Task) error {
				installed, _ := src.Installed(app)
				if installed {
					t.SetStatus("done")
					t.SetLabel(src.Name() + " — installed")
					t.AppendOutput(fmt.Sprintf("'%s' is installed via %s.", app, src.Name()))
					return nil
				}
				available, err := src.Available(app)
				if err != nil {
					t.SetStatus("failed")
					t.SetLabel(src.Name() + " — check failed")
					return fmt.Errorf("%s: %w", src.Name(), err)
				}
				if available {
					t.SetStatus("updates")
					t.SetLabel(src.Name() + " — available")
					t.AppendOutput(fmt.Sprintf("'%s' is available (not installed) via %s.", app, src.Name()))
					return nil
				}
				t.SetStatus("done")
				t.SetLabel(src.Name() + " — not found")
				t.AppendOutput(fmt.Sprintf("'%s' was not found in %s.", app, src.Name()))
				return nil
			},
		})
	}

	if isTerminal() {
		if opErr, progErr := tui.RunAny("pkgz info "+app, "", nil, ops); progErr == nil {
			return opErr
		}
	}
	return tui.RunPlain(ops)
}

// handleInfoCounts shows the installed package count per source.
func handleInfoCounts(sources []Source) error {
	ops := make([]tui.Op, 0, len(sources))
	for _, src := range sources {
		src := src
		ops = append(ops, tui.Op{
			Label: src.Name(),
			Run: func(t *tui.Task) error {
				count, err := src.InstalledCount()
				if err != nil {
					t.SetStatus("failed")
					t.SetLabel(src.Name() + " — unavailable")
					return fmt.Errorf("%s: %w", src.Name(), err)
				}
				t.SetStatus("done")
				t.SetLabel(fmt.Sprintf("%s — %d installed", src.Name(), count))
				t.AppendOutput(fmt.Sprintf("%s has %d package(s) installed.", src.Name(), count))
				return nil
			},
		})
	}

	if isTerminal() {
		if opErr, progErr := tui.RunAny("pkgz info", "", nil, ops); progErr == nil {
			return opErr
		}
	}
	return tui.RunPlain(ops)
}

// --- clean ------------------------------------------------------------------

func handleClean(sources []Source, elevator *utils.Elevator) error {
	var ops []tui.Op
	for _, source := range sources {
		var label string
		var privileged bool
		var run func(*tui.Task) error

		switch source.Name() {
		case "Apt":
			label = "Cleaning Apt cache"
			privileged = true
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Apt cache")
				return elevator.RunPrivilegedStreaming("apt", []string{"clean"}, t.AppendOutput)
			}
		case "Pacman":
			label = "Cleaning Pacman cache"
			privileged = true
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Pacman cache")
				return elevator.RunPrivilegedStreaming("pacman", []string{"-Sc", "--noconfirm"}, t.AppendOutput)
			}
		case "Paru (AUR)":
			label = "Cleaning Paru cache"
			privileged = false
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Paru cache")
				return utils.RunCommandStreaming("paru", []string{"-Sc", "--noconfirm"}, t.AppendOutput)
			}
		case "Yay (AUR)":
			label = "Cleaning Yay cache"
			privileged = false
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Yay cache")
				return utils.RunCommandStreaming("yay", []string{"-Sc", "--noconfirm"}, t.AppendOutput)
			}
		case "DNF":
			label = "Cleaning DNF cache"
			privileged = true
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning DNF cache")
				return elevator.RunPrivilegedStreaming("dnf", []string{"clean", "all"}, t.AppendOutput)
			}
		case "Zypper":
			label = "Cleaning Zypper cache"
			privileged = true
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Zypper cache")
				return elevator.RunPrivilegedStreaming("zypper", []string{"clean"}, t.AppendOutput)
			}
		case "Flatpak":
			label = "Cleaning Flatpak cache"
			privileged = false
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Flatpak cache")
				return utils.RunCommandStreaming("flatpak", []string{"uninstall", "--user", "--unused", "-y"}, t.AppendOutput)
			}
		default:
			continue
		}

		ops = append(ops, tui.Op{Label: label, Privileged: privileged, Run: run})
	}

	if len(ops) == 0 {
		fmt.Println("No cleanable sources enabled.")
		return nil
	}
	return runOps("pkgz clean", ops)
}
