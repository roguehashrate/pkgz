package linux

import (
	"strings"

	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

type NixSource struct {
	elevator *utils.Elevator
}

func NewNixSource(elevator *utils.Elevator) sources.Source {
	return &NixSource{elevator: elevator}
}

func (n *NixSource) Name() string {
	return "Nix"
}

func (n *NixSource) Available(app string) (bool, error) {
	output, err := utils.RunCommand("nix", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(output, app), nil
}

func (n *NixSource) Installed(app string) (bool, error) {
	packages, err := n.installedPackages()
	if err != nil {
		return false, nil
	}

	for _, pkg := range packages {
		if strings.Contains(pkg, app) {
			return true, nil
		}
	}
	return false, nil
}

func (n *NixSource) installedPackages() ([]string, error) {
	output, err := utils.RunCommand("nix-env", "-q")
	if err != nil {
		return []string{}, nil
	}

	packages := strings.Split(strings.TrimSpace(output), "\n")
	result := make([]string, 0, len(packages))

	for _, pkg := range packages {
		if strings.TrimSpace(pkg) != "" {
			// Remove version part (everything after first dash)
			if idx := strings.Index(pkg, "-"); idx > 0 {
				pkg = pkg[:idx]
			}
			result = append(result, strings.TrimSpace(pkg))
		}
	}
	return result, nil
}

func (n *NixSource) Install(app string) error {
	_, err := utils.RunCommand("nix-env", "-iA", "nixpkgs."+app)
	return err
}

func (n *NixSource) Remove(app string) error {
	_, err := utils.RunCommand("nix-env", "-e", app)
	return err
}

func (n *NixSource) Update() error {
	_, err := utils.RunCommand("nix-env", "-u", "*")
	return err
}

func (n *NixSource) Search(app string) (bool, error) {
	output, err := utils.RunCommand("nix", "search", app)
	if err != nil {
		return false, nil
	}
	return strings.Contains(strings.ToLower(output), strings.ToLower(app)), nil
}

func (n *NixSource) InstalledCount() (int, error) {
	packages, err := n.installedPackages()
	if err != nil {
		return 0, nil
	}
	return len(packages), nil
}
