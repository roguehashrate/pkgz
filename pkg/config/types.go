package config

import "os"

const (
	VERSION     = "0.1.9"
	CONFIG_PATH = "~/.config/pkgz/config.toml"
)

type Config struct {
	Sources  SourcesConfig  `toml:"sources"`
	Elevator ElevatorConfig `toml:"elevator"`
}

type SourcesConfig struct {
	Apt          bool `toml:"apt"`
	Nala         bool `toml:"nala"`
	Flatpak      bool `toml:"flatpak"`
	Pacman       bool `toml:"pacman"`
	Paru         bool `toml:"paru"`
	Yay          bool `toml:"yay"`
	Dnf          bool `toml:"dnf"`
	Pacstall     bool `toml:"pacstall"`
	Zypper       bool `toml:"zypper"`
	Xbps         bool `toml:"xbps"`
	XbpsSrc      bool `toml:"xbps_src"`
	Alpine       bool `toml:"alpine"`
	Nix          bool `toml:"nix"`
	FreeBsd      bool `toml:"freebsd"`
	FreeBsdPorts bool `toml:"freebsd_ports"`
	OpenBsd      bool `toml:"openbsd"`
	OpenBsdPorts bool `toml:"openbsd_ports"`
}

type ElevatorConfig struct {
	Command string `toml:"command"`
}

// ExpandPath expands ~ to the user's home directory
func ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		if home != "" {
			return home + path[1:]
		}
	}
	return path
}
