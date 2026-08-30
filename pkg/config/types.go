package config

import "os"

const (
	VERSION     = "1.1.0"
	CONFIG_PATH = "~/.config/pkgz/config.toml"
)

type Config struct {
	Sources  SourcesConfig  `toml:"sources"`
	Elevator ElevatorConfig `toml:"elevator"`
}

type SourcesConfig struct {
	Apt     bool `toml:"apt"`
	Nala    bool `toml:"nala"`
	Flatpak bool `toml:"flatpak"`
	Pacman  bool `toml:"pacman"`
	Paru    bool `toml:"paru"`
	Yay     bool `toml:"yay"`
	Dnf     bool `toml:"dnf"`
	Zypper  bool `toml:"zypper"`
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
