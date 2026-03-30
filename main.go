package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/roguehashrate/pkgz/pkg/config"
	"github.com/roguehashrate/pkgz/pkg/sources/bsd"
	"github.com/roguehashrate/pkgz/pkg/sources/linux"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

const VERSION = "0.1.9"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pkgz <install|remove|update|search|--version> [app-name]")
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
	if enabledSources["xbps"] {
		sources = append(sources, linux.NewXbpsSource(elevator))
	}
	if enabledSources["alpine"] {
		sources = append(sources, linux.NewApkSource(elevator))
	}
	if enabledSources["pacstall"] {
		sources = append(sources, linux.NewPacstallSource(elevator))
	}
	if enabledSources["zypper"] {
		sources = append(sources, linux.NewZypperSource(elevator))
	}
	if enabledSources["nix"] {
		sources = append(sources, linux.NewNixSource(elevator))
	}
	if enabledSources["xbps_src"] {
		sources = append(sources, linux.NewXbpsSrcSource(elevator))
	}
	// BSD Sources
	if enabledSources["freebsd"] {
		sources = append(sources, bsd.NewFreeBsdSource(elevator))
	}
	if enabledSources["freebsd_ports"] {
		sources = append(sources, bsd.NewFreeBsdPortsSource(elevator))
	}
	if enabledSources["openbsd"] {
		sources = append(sources, bsd.NewOpenBsdSource(elevator))
	}
	if enabledSources["openbsd_ports"] {
		sources = append(sources, bsd.NewOpenBsdPortsSource(elevator))
	}

	// Handle commands
	switch command {
	case "install":
		handleInstall(appName, sources)
	case "remove":
		handleRemove(appName, sources)
	case "update":
		handleUpdate(sources)
	case "search":
		handleSearch(appName, sources)
	case "info":
		handleInfo(appName, sources)
	case "clean":
		handleClean(sources)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: pkgz <install|remove|update|search|clean|info|--version> [app-name]")
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
	Search(app string) (bool, error)
	InstalledCount() (int, error)
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
		source := availableSources[0]
		fmt.Printf("✅ Found '%s' in %s. Installing...\n", appName, source.Name())
		if err := source.Install(appName); err != nil {
			fmt.Printf("❌ Installation failed: %v\n", err)
		}
		return
	}

	fmt.Printf("📦 Found '%s' in multiple sources:\n", appName)
	for i, source := range availableSources {
		fmt.Printf("%d. %s\n", i+1, source.Name())
	}

	fmt.Printf("Which one would you like to use? [1-%d]: ", len(availableSources))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(availableSources) {
		fmt.Println("❌ Invalid choice.")
		return
	}

	selected := availableSources[choice-1]
	fmt.Printf("🚀 Installing with %s...\n", selected.Name())
	if err := selected.Install(appName); err != nil {
		fmt.Printf("❌ Installation failed: %v\n", err)
	}
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
		source := installedSources[0]
		fmt.Printf("🗑️ Removing '%s' from %s...\n", appName, source.Name())
		if err := source.Remove(appName); err != nil {
			fmt.Printf("❌ Removal failed: %v\n", err)
		}
		return
	}

	fmt.Printf("⚠️ '%s' is installed in multiple sources:\n", appName)
	for i, source := range installedSources {
		fmt.Printf("%d. %s\n", i+1, source.Name())
	}

	fmt.Printf("Which one would you like to remove? [1-%d]: ", len(installedSources))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(installedSources) {
		fmt.Println("❌ Invalid choice.")
		return
	}

	selected := installedSources[choice-1]
	fmt.Printf("🗑️ Removing '%s' from %s...\n", appName, selected.Name())
	if err := selected.Remove(appName); err != nil {
		fmt.Printf("❌ Removal failed: %v\n", err)
	}
}

func handleUpdate(sources []Source) {
	for _, source := range sources {
		fmt.Printf("⬆️ Updating %s packages...\n", source.Name())
		if err := source.Update(); err != nil {
			fmt.Printf("❌ Update failed for %s: %v\n", source.Name(), err)
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
	for _, source := range sources {
		switch source.Name() {
		case "Apt":
			fmt.Println("🧹 Cleaning Apt cache...")
			elevator := utils.NewElevator()
			elevator.RunPrivileged("apt", "clean")
		case "Flatpak":
			fmt.Println("🧹 Cleaning Flatpak cache...")
			exec.Command("flatpak", "uninstall", "--user", "--unused", "-y").Run()
		case "Pacman":
			fmt.Println("🧹 Cleaning Pacman cache...")
			elevator := utils.NewElevator()
			elevator.RunPrivileged("pacman", "-Sc", "--noconfirm")
		case "Paru":
			fmt.Println("🧹 Cleaning Paru cache...")
			exec.Command("paru", "-Sc", "--noconfirm").Run()
		case "Yay":
			fmt.Println("🧹 Cleaning Yay cache...")
			exec.Command("yay", "-Sc", "--noconfirm").Run()
		case "DNF":
			fmt.Println("🧹 Cleaning DNF cache...")
			elevator := utils.NewElevator()
			elevator.RunPrivileged("dnf", "clean", "all")
		case "Alpine":
			fmt.Println("🧹 Cleaning Alpine cache...")
			elevator := utils.NewElevator()
			elevator.RunPrivileged("rm", "-rf", "/var/cache/apk/*")
		case "XBPS":
			fmt.Println("🧹 Cleaning XBPS cache...")
			elevator := utils.NewElevator()
			elevator.RunPrivileged("xbps-remove", "-O")
		case "Pacstall":
			fmt.Println("🧹 Cleaning Pacstall cache...")
			elevator := utils.NewElevator()
			elevator.RunPrivileged("pacstall", "-C")
		case "Nix":
			fmt.Println("⚠️  No automatic clean command available for Nix.")
		case "FreeBSD", "FreeBSD Ports", "OpenBSD", "OpenBSD Ports":
			fmt.Printf("⚠️  No automatic clean command available for %s.\n", source.Name())
		default:
			fmt.Printf("⚠️ No automatic clean command available for %s.\n", source.Name())
		}
	}
}
