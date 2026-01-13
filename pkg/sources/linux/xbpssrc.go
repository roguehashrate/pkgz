package linux

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type XbpsSrcSource struct {
	elevator *utils.Elevator
	voidPkgs string
}

func NewXbpsSrcSource(elevator *utils.Elevator) sources.Source {
	return &XbpsSrcSource{
		elevator: elevator,
		voidPkgs: "/usr/src/void-packages",
	}
}

func (x *XbpsSrcSource) Name() string {
	return "XBPS-SRC"
}

func (x *XbpsSrcSource) Available(app string) (bool, error) {
	srcpkgDir := filepath.Join(x.voidPkgs, "srcpkgs", app)
	info, err := os.Stat(srcpkgDir)
	if err != nil {
		return false, nil
	}
	return info.IsDir(), nil
}

func (x *XbpsSrcSource) Installed(app string) (bool, error) {
	return utils.RunCommandWithRedirect("xbps-query", "-p", "pkgver", app), nil
}

func (x *XbpsSrcSource) Install(app string) error {
	// Check if xbps-src directory exists
	if _, err := os.Stat(x.voidPkgs); os.IsNotExist(err) {
		fmt.Printf("❌ xbps-src not found at %s\n", x.voidPkgs)
		fmt.Println("Install it with:")
		fmt.Println("  sudo xbps-install -S void-packages")
		return nil
	}

	// Build and install package
	cmd := fmt.Sprintf("cd %s && ./xbps-src pkg %s && xi %s", x.voidPkgs, app, app)
	return x.elevator.RunPrivileged("sh", "-c", cmd)
}

func (x *XbpsSrcSource) Remove(app string) error {
	return x.elevator.RunPrivileged("xbps-remove", "-Ry", app)
}

func (x *XbpsSrcSource) Update() error {
	fmt.Println("⚠️  xbps-src does not support mass updates.")
	fmt.Println("Use: pkgz install <pkg> to rebuild when needed.")
	return nil
}

func (x *XbpsSrcSource) Search(app string) (bool, error) {
	srcpkgDir := filepath.Join(x.voidPkgs, "srcpkgs")

	// Get all directories in srcpkgs
	entries, err := os.ReadDir(srcpkgDir)
	if err != nil {
		return false, nil
	}

	appLower := strings.ToLower(app)
	for _, entry := range entries {
		if entry.IsDir() {
			name := strings.ToLower(entry.Name())
			if strings.Contains(name, appLower) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (x *XbpsSrcSource) InstalledCount() (int, error) {
	// XBPS-SRC doesn't have a simple way to count source-installed packages
	// Return nil to indicate unavailable
	return 0, nil
}
