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

const VERSION = "1.0.1"

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

// runOps runs operations through the TUI when stdout is a terminal, otherwise
// falls back to plain sequential output (pipes, CI, scripts).
func runOps(title string, ops []tui.Op) error {
	if isTerminal() {
		return tui.Run(title, "", nil, ops)
	}
	return tui.RunPlain(ops)
}

// runInstall presents a source picker inside the TUI when multiple options are
// available, then runs the chosen install. Falls back to a plain numbered
// prompt when stdout is not a terminal.
func runInstall(appName string, availableSources []Source) error {
	return runPick("install", appName, availableSources, func(src Source) tui.Op {
		return tui.Op{
			Label: "Installing " + appName + " via " + src.Name(),
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
			Label: "Removing " + appName + " via " + src.Name(),
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
			Label: "Updating " + src.Name(),
			Run: func(t *tui.Task) error {
				return withTask(src, t).Update()
			},
		})
	}
	runOps("pkgz update", ops)
}

func handleRefresh(sources []Source) {
	for _, source := range sources {
		fmt.Printf("🔄 Checking %s for updates...\n", source.Name())
		updates, err := source.ListUpdates()
		if err != nil {
			fmt.Printf("❌ Update check failed for %s: %v\n", source.Name(), err)
			continue
		}
		if len(updates) == 0 {
			continue
		}
		fmt.Printf("⬆️ %s has %d update(s) available:\n", source.Name(), len(updates))
		for _, pkg := range updates {
			fmt.Printf("  - %s\n", pkg)
		}
	}
}

func handleSearch(appName string, sources []Source) {
	if appName == "" {
		fmt.Println("Usage: pkgz search <app-name>")
		return
	}

	fmt.Printf("🔍 Searching for '%s' across enabled sources...\n", appName)
	anyFound := false

	for _, source := range sources {
		if found, err := source.Search(appName); err == nil && found {
			fmt.Printf("✅ Found in %s\n", source.Name())
			anyFound = true
		} else {
			fmt.Printf("❌ Not found in %s\n", source.Name())
		}
	}

	if !anyFound {
		fmt.Printf("📦 Package '%s' not found in any enabled source.\n", appName)
	}
}

func handleInfo(appName string, sources []Source) {
	if appName == "" {
		fmt.Println("📦 pkgz info")
		fmt.Println()

		for _, source := range sources {
			if count, err := source.InstalledCount(); err == nil {
				fmt.Printf("%s: %d\n", source.Name(), count)
			} else {
				fmt.Printf("%s: unavailable\n", source.Name())
			}
		}
		return
	}

	fmt.Printf("ℹ️ Info for '%s':\n\n", appName)

	foundAny := false
	for _, source := range sources {
		installed, _ := source.Installed(appName)
		available, _ := source.Available(appName)

		var status string
		if installed {
			status = "✔ INSTALLED"
		} else if available {
			status = "○ AVAILABLE"
		} else {
			status = "✖ NOT FOUND"
		}

		fmt.Printf("  %-13s %s\n", status, source.Name())
		foundAny = foundAny || installed || available
	}

	if !foundAny {
		fmt.Println()
		fmt.Printf("❌ '%s' was not found in any enabled source.\n", appName)
	}
}

func handleClean(sources []Source) {
	var ops []tui.Op
	for _, source := range sources {
		var label string
		var run func(*tui.Task) error

		switch source.Name() {
		case "Apt":
			label = "Cleaning Apt cache"
			run = func(t *tui.Task) error {
				e := utils.NewElevator()
				t.SetLabel("Cleaning Apt cache")
				return e.RunPrivilegedStreaming("apt", []string{"clean"}, t.AppendOutput)
			}
		case "Pacman":
			label = "Cleaning Pacman cache"
			run = func(t *tui.Task) error {
				e := utils.NewElevator()
				t.SetLabel("Cleaning Pacman cache")
				return e.RunPrivilegedStreaming("pacman", []string{"-Sc", "--noconfirm"}, t.AppendOutput)
			}
		case "Paru":
			label = "Cleaning Paru cache"
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Paru cache")
				return utils.RunCommandStreaming("paru", []string{"-Sc", "--noconfirm"}, t.AppendOutput)
			}
		case "Yay":
			label = "Cleaning Yay cache"
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Yay cache")
				return utils.RunCommandStreaming("yay", []string{"-Sc", "--noconfirm"}, t.AppendOutput)
			}
		case "DNF":
			label = "Cleaning DNF cache"
			run = func(t *tui.Task) error {
				e := utils.NewElevator()
				t.SetLabel("Cleaning DNF cache")
				return e.RunPrivilegedStreaming("dnf", []string{"clean", "all"}, t.AppendOutput)
			}
		case "Flatpak":
			label = "Cleaning Flatpak cache"
			run = func(t *tui.Task) error {
				t.SetLabel("Cleaning Flatpak cache")
				return utils.RunCommandStreaming("flatpak", []string{"uninstall", "--user", "--unused", "-y"}, t.AppendOutput)
			}
		default:
			continue
		}

		ops = append(ops, tui.Op{Label: label, Run: run})
	}

	if len(ops) == 0 {
		fmt.Println("No cleanable sources enabled.")
		return
	}
	runOps("pkgz clean", ops)
}
