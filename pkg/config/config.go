package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

// LoadConfig loads and parses the configuration file
func LoadConfig() (*Config, error) {
	configPath := ExpandPath(CONFIG_PATH)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("❌ Config file not found at %s\nPlease create it manually with the sources you want enabled.\n\n[sources]\napt = true\nnala = false\nflatpak = true\nparu = false\nyay = false\npacman = false\ndnf = false\npacstall = true\n\n[elevator]\ncommand = \"sudo\"", configPath)
	}

	// Read and parse TOML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Set default elevator if not configured
	if config.Elevator.Command == "" {
		config.Elevator.Command = "sudo"
	}

	return &config, nil
}

// GetEnabledSources returns a map of enabled source names
func (c *Config) GetEnabledSources() map[string]bool {
	return map[string]bool{
		"apt":           c.Sources.Apt,
		"nala":          c.Sources.Nala,
		"flatpak":       c.Sources.Flatpak,
		"pacman":        c.Sources.Pacman,
		"paru":          c.Sources.Paru,
		"yay":           c.Sources.Yay,
		"dnf":           c.Sources.Dnf,
		"pacstall":      c.Sources.Pacstall,
		"zypper":        c.Sources.Zypper,
		"xbps":          c.Sources.Xbps,
		"xbps_src":      c.Sources.XbpsSrc,
		"alpine":        c.Sources.Alpine,
		"nix":           c.Sources.Nix,
		"freebsd":       c.Sources.FreeBsd,
		"freebsd_ports": c.Sources.FreeBsdPorts,
		"openbsd":       c.Sources.OpenBsd,
		"openbsd_ports": c.Sources.OpenBsdPorts,
	}
}
