package main

import (
	"reflect"
	"testing"
)

type stubSource struct{ name string }

func (s stubSource) Name() string                   { return s.name }
func (s stubSource) Available(string) (bool, error) { return false, nil }
func (s stubSource) Installed(string) (bool, error) { return false, nil }
func (s stubSource) Install(string) error           { return nil }
func (s stubSource) Remove(string) error            { return nil }
func (s stubSource) Update() error                  { return nil }
func (s stubSource) ListUpdates() ([]string, error) { return nil, nil }
func (s stubSource) Search(string) (bool, error)    { return false, nil }
func (s stubSource) InstalledCount() (int, error)   { return 0, nil }

func TestParseArgs(t *testing.T) {
	force, pkgs := parseArgs([]string{"--source", "apt", "vim", "curl"})
	if force != "apt" || !reflect.DeepEqual(pkgs, []string{"vim", "curl"}) {
		t.Errorf("parseArgs = (%q, %v), want (apt, [vim curl])", force, pkgs)
	}

	force, pkgs = parseArgs([]string{"emacs", "--from=flatpak"})
	if force != "flatpak" || !reflect.DeepEqual(pkgs, []string{"emacs"}) {
		t.Errorf("parseArgs = (%q, %v), want (flatpak, [emacs])", force, pkgs)
	}

	force, pkgs = parseArgs([]string{"vim", "--source", "apt"})
	if force != "apt" || !reflect.DeepEqual(pkgs, []string{"vim"}) {
		t.Errorf("parseArgs = (%q, %v), want (apt, [vim])", force, pkgs)
	}
}

func TestMatchSource(t *testing.T) {
	sources := []Source{
		stubSource{name: "Apt"},
		stubSource{name: "Paru (AUR)"},
		stubSource{name: "Flatpak"},
	}

	for _, key := range []string{"apt", "Apt", "flatpak", "paru", "Paru (AUR)"} {
		if matchSource(sources, key) == nil {
			t.Errorf("matchSource(%q) = nil, want a match", key)
		}
	}
	if got := matchSource(sources, "dnf"); got != nil {
		t.Errorf("matchSource(dnf) = %v, want nil", got)
	}
}

func neverStub(name string) stubSource { return stubSource{name: name} }

func TestIntersect(t *testing.T) {
	a := []Source{neverStub("Apt"), neverStub("Flatpak")}
	b := []Source{neverStub("Flatpak"), neverStub("DNF")}
	out := intersect(a, b)
	if len(out) != 1 || out[0].Name() != "Flatpak" {
		t.Errorf("intersect = %v, want [Flatpak]", out)
	}
}
