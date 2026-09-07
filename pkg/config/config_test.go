package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeHomePointers the test at a fresh HOME with PATH prepended to a dir that
// contains the given fake binaries, so source/elevator detection is hermetic.
func fakeHomeWithBins(t *testing.T, bins ...string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	bin := filepath.Join(home, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	for _, name := range bins {
		path := filepath.Join(bin, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return home
}

func TestLoadConfigCreatesDefaultOnFirstRun(t *testing.T) {
	home := fakeHomeWithBins(t, "apt", "flatpak", "sudo")

	loadConfig(t)

	configPath := filepath.Join(home, ".config", "pkgz", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config to be written: %v", err)
	}
	for _, want := range []string{"apt = true", "flatpak = true", `command = "sudo"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("config missing %q:\n%s", want, data)
		}
	}
	if strings.Contains(string(data), "dnf = true") {
		t.Errorf("config should not enable undetected dnf:\n%s", data)
	}
}

func TestEnabledSourcesSkipsMissingBinary(t *testing.T) {
	home := fakeHomeWithBins(t, "apt", "sudo")
	// Restrict PATH to the fake binaries so the real system's package managers
	// cannot leak in and muddy the detection.
	t.Setenv("PATH", filepath.Join(home, "bin"))

	// Hand-written config enabling a source whose binary is absent.
	dir := filepath.Join(home, ".config", "pkgz")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfg := "[sources]\napt = true\nflatpak = true\ndnf = true\n\n[elevator]\ncommand = \"sudo\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	c := loadConfig(t)
	present, missing := c.EnabledSources()
	if len(present) != 1 || present[0] != "apt" {
		t.Errorf("present = %v, want [apt]", present)
	}
	if len(missing) != 2 {
		t.Errorf("missing = %v, want flatpak and dnf", missing)
	}
}

func loadConfig(t *testing.T) *Config {
	t.Helper()
	c, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	return c
}
