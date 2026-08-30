package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/roguehashrate/pkgz/pkg/config"
	"github.com/roguehashrate/pkgz/pkg/sources/linux"
	"github.com/roguehashrate/pkgz/pkg/tui"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

const VERSION = "1.1.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pkgz <install|remove|update|refresh|search|--version> [app-name]")
		os.Exit(1)
	}

	command := os.Args[1]
	appName := ""
	if len(os.Args) > 2 {
		appName = os.Args[2]
	}

	if command == "--version" {
		fmt.Printf("pkgz version %s\n", VERSION)
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize elevator
	elevator := utils.NewElevator()

	// Create sources based on enabled configuration
	var sources []Source
	enabledSources := cfg.GetEnabledSources()

	if enabledSources["apt"] {
		sources = append(sources, linux.NewAptSource(elevator))
	}
	if enabledSources["nala"] {
		sources = append(sources, linux.NewNalaSource(elevator))
	}
	if enabledSources["flatpak"] {
		sources = append(sources, linux.NewFlatpakSource(elevator))
	}
	if enabledSources["pacman"] {
		sources = append(sources, linux.NewPacmanSource(elevator))
	}
	if enabledSources["paru"] {
		sources = append(sources, linux.NewParuSource(elevator))
	}
	if enabledSources["yay"] {
		sources = append(sources, linux.NewYaySource(elevator))
	}
	if enabledSources["dnf"] {
		sources = append(sources, linux.NewDnfSource(elevator))
	}
	if enabledSources["zypper"] {
		sources = append(sources, linux.NewZypperSource(elevator))
	}

	// Handle commands
	switch command {
	case "install":
		handleInstall(appName, sources)
	case "remove":
		handleRemove(appName, sources)
	case "update":
		handleUpdate(sources)
	case "refresh":
		handleRefresh(sources)
	case "search":
		handleSearch(appName, sources)
	case "info":
		handleInfo(appName, sources)
	case "clean":
		handleClean(sources)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: pkgz <install|remove|update|refresh|search|clean|info|--version> [app-name]")
		os.Exit(1)
	}
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

// withTask attaches a reporting task to a source it if supports it, so the
// source's streaming output is surfaced in the TUI.
func withTask(src Source, t utils.Task) Source {
	if s, ok := src.(interface{ SetTask(utils.Task) }); ok {
		s.SetTask(t)
	}
	return src
}

// runOps runs operations through the TUI when stdout is a terminal. Because
// bubbletea's Exec releases the terminal into cooked mode while a privileged
// operation runs, sudo/doas password prompts work and the TUI is restored
// afterwards - so privileged and non-privileged ops alike run inside the TUI.
// When stdout is not a terminal (pipes, CI, scripts) it falls back to plain
// sequential output.
func runOps(title string, ops []tui.Op) error {
	if isTerminal() {
		if opErr, progErr := tui.RunAny(title, "", nil, ops); progErr == nil {
			return opErr
		}
	}
	return tui.RunPlain(ops)
}

// runInstall presents a source picker inside the TUI when multiple options are
// available, then runs the chosen install. Falls back to a plain numbered
// prompt when stdout is not a terminal.
func runInstall(appName string, availableSources []Source) error {
	return runPick("install", appName, availableSources, func(src Source) tui.Op {
		return tui.Op{
			Label:      "Installing " + appName + " via " + src.Name(),
			Privileged: installPrivileged(src),
			Run: func(t *tui.Task) error {
				return withTask(src, t).Install(appName)
			},
		}
	})
}

// runRemove presents a source picker inside the TUI when multiple installed
// sources match, then runs the chosen removal.
func runRemove(appName string, installedSources []Source) error {
	return runPick("remove", appName, installedSources, func(src Source) tui.Op {
		return tui.Op{
			Label:      "Removing " + appName + " via " + src.Name(),
			Privileged: removePrivileged(src),
			Run: func(t *tui.Task) error {
				return withTask(src, t).Remove(appName)
			},
		}
	})
}

// runPick runs a single source operation chosen from candidate sources. When
// more than one candidate exists it shows a picker in the TUI (or a numbered
// prompt when not a terminal); a single candidate runs directly.
func runPick(verb, appName string, sources []Source, build func(Source) tui.Op) error {
	if len(sources) == 1 {
		src := sources[0]
		return runOps("pkgz "+verb, []tui.Op{build(src)})
	}

	choices := make([]string, len(sources))
	ops := make([]tui.Op, len(sources))
	for i, src := range sources {
		src := src
		choices[i] = src.Name()
		ops[i] = build(src)
	}

	if isTerminal() {
		return tui.Run("pkgz "+verb, fmt.Sprintf("'%s' is available via multiple sources. Choose one:", appName), choices, ops)
	}

	// Non-TTY: plain numbered prompt.
	fmt.Printf("⚠️ '%s' is available via multiple sources:\n", appName)
	for i, src := range sources {
		fmt.Printf("%d. %s\n", i+1, src.Name())
	}
	fmt.Printf("Which one would you like to use? [1-%d]: ", len(sources))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(sources) {
		fmt.Println("❌ Invalid choice.")
		return nil
	}
	return runOps("pkgz "+verb, []tui.Op{ops[choice-1]})
}

func handleInstall(appName string, sources []Source) {
	if appName == "" {
		fmt.Println("Usage: pkgz install <app-name>")
		return
	}

	fmt.Printf("🔍 Searching for '%s' in sources...\n", appName)

	var availableSources []Source
	for _, source := range sources {
		if available, err := source.Available(appName); err == nil && available {
			availableSources = append(availableSources, source)
		}
	}

	if len(availableSources) == 0 {
		fmt.Printf("❌ App '%s' not found in any source.\n", appName)
		return
	}

	if len(availableSources) == 1 {
		fmt.Printf("✅ Found '%s' in %s.\n", appName, availableSources[0].Name())
	}
	runInstall(appName, availableSources)
}

func handleRemove(appName string, sources []Source) {
	if appName == "" {
		fmt.Println("Usage: pkgz remove <app-name>")
		return
	}

	var installedSources []Source
	for _, source := range sources {
		if installed, err := source.Installed(appName); err == nil && installed {
			installedSources = append(installedSources, source)
		}
	}

	if len(installedSources) == 0 {
		fmt.Printf("❌ '%s' is not installed in any enabled source.\n", appName)
		return
	}

	if len(installedSources) == 1 {
		fmt.Printf("🗑️ Found '%s' in %s.\n", appName, installedSources[0].Name())
	}
	runRemove(appName, installedSources)
}

func handleUpdate(sources []Source) {
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
	runOps("pkgz update", ops)
}

func handleRefresh(sources []Source) {
	if !isTerminal() {
		refreshPlain(sources)
		return
	}

	ops := make([]tui.Op, 0, len(sources))
	for _, src := range sources {
		src := src
		ops = append(ops, tui.Op{
			Label: "Checking " + src.Name() + " for updates",
			Run: func(t *tui.Task) error {
				updates, err := src.ListUpdates()
				if err != nil {
					t.SetStatus("failed")
					t.AppendOutput("✗ update check failed: " + err.Error())
					return nil
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

	if _, progErr := tui.RunAny("pkgz refresh", "", nil, ops); progErr != nil {
		// TUI could not start; fall back to plain output.
		refreshPlain(sources)
	}
}

// refreshPlain prints the per-source update check result without a terminal.
func refreshPlain(sources []Source) {
	if len(sources) == 0 {
		fmt.Println("No sources enabled.")
		return
	}
	for _, src := range sources {
		updates, err := src.ListUpdates()
		if err != nil {
			fmt.Printf("❌ %s: update check failed: %v\n", src.Name(), err)
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
}

func handleSearch(appName string, sources []Source) {
	if appName == "" {
		fmt.Println("Usage: pkgz search <app-name>")
		return
	}

	if !isTerminal() {
		searchPlain(appName, sources)
		return
	}

	ops := make([]tui.Op, 0, len(sources))
	for _, src := range sources {
		src := src
		ops = append(ops, tui.Op{
			Label: "Searching " + src.Name(),
			Run: func(t *tui.Task) error {
				found, err := src.Search(appName)
				if err != nil {
					t.SetStatus("failed")
					t.AppendOutput("✗ search failed: " + err.Error())
					return nil
				}
				if found {
					t.SetStatus("done")
					t.SetLabel("Found in " + src.Name())
					t.AppendOutput(fmt.Sprintf("'%s' is available via %s.", appName, src.Name()))
					return nil
				}
				t.SetStatus("failed")
				t.SetLabel("Not found in " + src.Name())
				t.AppendOutput(fmt.Sprintf("'%s' was not found in %s.", appName, src.Name()))
				return nil
			},
		})
	}

	if _, progErr := tui.RunAny("pkgz search "+appName, "", nil, ops); progErr != nil {
		searchPlain(appName, sources)
	}
}

// searchPlain prints the per-source search result without a terminal.
func searchPlain(app string, sources []Source) {
	foundAny := false
	for _, src := range sources {
		found, err := src.Search(app)
		if err != nil {
			fmt.Printf("❌ %s: search failed: %v\n", src.Name(), err)
			continue
		}
		if found {
			fmt.Printf("✅ Found in %s\n", src.Name())
			foundAny = true
		} else {
			fmt.Printf("❌ Not found in %s\n", src.Name())
		}
	}
	if !foundAny {
		fmt.Printf("📦 Package '%s' not found in any enabled source.\n", app)
	}
}

func handleInfo(appName string, sources []Source) {
	if appName == "" {
		handleInfoCounts(sources)
		return
	}

	ops := make([]tui.Op, 0, len(sources))
	for _, src := range sources {
		src := src
		ops = append(ops, tui.Op{
			Label: src.Name(),
			Run: func(t *tui.Task) error {
				installed, _ := src.Installed(appName)
				available, _ := src.Available(appName)
				switch {
				case installed:
					t.SetStatus("done")
					t.SetLabel(src.Name() + " — INSTALLED")
					t.AppendOutput(fmt.Sprintf("'%s' is installed via %s.", appName, src.Name()))
				case available:
					t.SetStatus("updates")
					t.SetLabel(src.Name() + " — AVAILABLE")
					t.AppendOutput(fmt.Sprintf("'%s' is available (not installed) via %s.", appName, src.Name()))
				default:
					t.SetStatus("failed")
					t.SetLabel(src.Name() + " — NOT FOUND")
					t.AppendOutput(fmt.Sprintf("'%s' was not found in %s.", appName, src.Name()))
				}
				return nil
			},
		})
	}

	if isTerminal() {
		if _, progErr := tui.RunAny("pkgz info "+appName, "", nil, ops); progErr == nil {
			return
		}
	}
	// Non-TTY fallback.
	tui.RunPlain(ops)
}

// handleInfoCounts shows the installed package count per source.
func handleInfoCounts(sources []Source) {
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
					return nil
				}
				t.SetStatus("done")
				t.SetLabel(fmt.Sprintf("%s — %d installed", src.Name(), count))
				return nil
			},
		})
	}

	if isTerminal() {
		if _, progErr := tui.RunAny("pkgz info", "", nil, ops); progErr == nil {
			return
		}
	}
	// Non-TTY fallback.
	tui.RunPlain(ops)
}

func handleClean(sources []Source) {
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
				e := utils.NewElevator()
				t.SetLabel("Cleaning Apt cache")
				return e.RunPrivilegedStreaming("apt", []string{"clean"}, t.AppendOutput)
			}
		case "Nala":
			label = "Cleaning Nala cache"
			privileged = true
			run = func(t *tui.Task) error {
				e := utils.NewElevator()
				t.SetLabel("Cleaning Nala cache")
				return e.RunPrivilegedStreaming("nala", []string{"clean"}, t.AppendOutput)
			}
		case "Pacman":
			label = "Cleaning Pacman cache"
			privileged = true
			run = func(t *tui.Task) error {
				e := utils.NewElevator()
				t.SetLabel("Cleaning Pacman cache")
				return e.RunPrivilegedStreaming("pacman", []string{"-Sc", "--noconfirm"}, t.AppendOutput)
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
				e := utils.NewElevator()
				t.SetLabel("Cleaning DNF cache")
				return e.RunPrivilegedStreaming("dnf", []string{"clean", "all"}, t.AppendOutput)
			}
		case "Zypper":
			label = "Cleaning Zypper cache"
			privileged = true
			run = func(t *tui.Task) error {
				e := utils.NewElevator()
				t.SetLabel("Cleaning Zypper cache")
				return e.RunPrivilegedStreaming("zypper", []string{"clean"}, t.AppendOutput)
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
		return
	}
	runOps("pkgz clean", ops)
}
