package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
)

// sourceOrder is the canonical display/enable order for sources.
var sourceOrder = []string{"apt", "flatpak", "pacman", "paru", "yay", "dnf", "zypper"}

// SourceBinaries maps each supported source to the binary that backs it.
// A source is only usable when its binary exists on PATH.
var SourceBinaries = map[string]string{
	"apt":     "apt",
	"flatpak": "flatpak",
	"pacman":  "pacman",
	"paru":    "paru",
	"yay":     "yay",
	"dnf":     "dnf",
	"zypper":  "zypper",
}

// DetectSources returns the names of package managers present on the system,
// in a stable order.
func DetectSources() []string {
	var found []string
	for _, src := range sourceOrder {
		if _, err := exec.LookPath(SourceBinaries[src]); err == nil {
			found = append(found, src)
		}
	}
	return found
}

// DetectElevator returns a sensible privilege-escalation command present on the
// system (preferring doas), or "" when none is available.
func DetectElevator() string {
	for _, cmd := range []string{"doas", "sudo", "pkexec"} {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}

// LoadConfig loads and parses the configuration file. When no config exists it
// creates one automatically from detected sources and returns it, so pkgz works
// out of the box instead of demanding a hand-written file.
func LoadConfig() (*Config, error) {
	configPath := ExpandPath(CONFIG_PATH)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := ensureConfigFile(configPath); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "ℹ️  First run: created default config at %s\n", configPath)
		fmt.Fprintln(os.Stderr, "   It enables the package managers detected on this system. Edit it and re-run pkgz to change the set.")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %v\nPlease fix the file and try again.", configPath, err)
	}

	// Resolve elevator: use the configured command when present, otherwise
	// detect a usable one. Validation against PATH happens at startup.
	if config.Elevator.Command == "" {
		config.Elevator.Command = DetectElevator()
	}

	return &config, nil
}

// ensureConfigFile writes a default config with the detected sources enabled.
func ensureConfigFile(configPath string) error {
	sources := DetectSources()
	elevator := DetectElevator()

	if len(sources) == 0 {
		return fmt.Errorf("no supported package manager found on this system (apt, flatpak, pacman, paru, yay, dnf, zypper)")
	}
	if elevator == "" {
		return fmt.Errorf("no privilege elevation tool found (doas, sudo, pkexec); install sudo or doas first")
	}

	var b strings.Builder
	b.WriteString("# pkgz automatically detected and enabled these package managers.\n")
	b.WriteString("# Set any source to false to disable it.\n")
	b.WriteString("[sources]\n")
	for _, src := range sourceOrder {
		enabled := "false"
		for _, detected := range sources {
			if detected == src {
				enabled = "true"
				break
			}
		}
		b.WriteString(fmt.Sprintf("%s = %s\n", src, enabled))
	}
	b.WriteString("\n[elevator]\n")
	b.WriteString(fmt.Sprintf("command = %q  # or \"doas\"\n", elevator))

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	return os.WriteFile(configPath, []byte(b.String()), 0o644)
}

// EnabledSources returns the list of source names that are both enabled in the
// config and present on the system. Sources enabled in config but missing a
// binary are reported via the returned []string so callers can warn about them.
func (c *Config) EnabledSources() (present []string, missingBinary []string) {
	enabled := c.GetEnabledSources()
	for _, src := range sourceOrder {
		if !enabled[src] {
			continue
		}
		if _, err := exec.LookPath(SourceBinaries[src]); err == nil {
			present = append(present, src)
		} else {
			missingBinary = append(missingBinary, src)
		}
	}
	return present, missingBinary
}

// GetEnabledSources returns a map of enabled source names
func (c *Config) GetEnabledSources() map[string]bool {
	return map[string]bool{
		"apt":     c.Sources.Apt,
		"flatpak": c.Sources.Flatpak,
		"pacman":  c.Sources.Pacman,
		"paru":    c.Sources.Paru,
		"yay":     c.Sources.Yay,
		"dnf":     c.Sources.Dnf,
		"zypper":  c.Sources.Zypper,
	}
}
