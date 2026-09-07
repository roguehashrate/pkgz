package linux

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fakeBin writes an executable script into a temp dir on PATH and returns the
// directory plus a cleanup function.
func fakeBin(t *testing.T, name, script string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake bin: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestListUpdatesCmd(t *testing.T) {
	t.Run("clean no updates", func(t *testing.T) {
		fakeBin(t, "fakecheck", "#!/bin/sh\nexit 0\n")
		updates, err := listUpdatesCmd("fakecheck", nil, func(string) []string { return nil })()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(updates) != 0 {
			t.Fatalf("expected no updates, got %v", updates)
		}
	})

	t.Run("nonzero exit with package lines is updates", func(t *testing.T) {
		fakeBin(t, "fakecheck", "#!/bin/sh\necho \"pkgA-1.0\npkgB-2.0\"\nexit 100\n")
		parse := func(o string) []string { return strings.Fields(o) }
		updates, err := listUpdatesCmd("fakecheck", nil, parse)()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		want := []string{"pkgA-1.0", "pkgB-2.0"}
		if !reflect.DeepEqual(updates, want) {
			t.Fatalf("updates = %v, want %v", updates, want)
		}
	})

	t.Run("genuine failure without packages is an error", func(t *testing.T) {
		fakeBin(t, "fakecheck", "#!/bin/sh\necho \"Error: failed to sync\" >&2\nexit 1\n")
		if _, err := listUpdatesCmd("fakecheck", nil, func(string) []string { return nil })(); err == nil {
			t.Fatal("expected error for failed update check")
		}
	})

	t.Run("nonzero exit means no updates for pacman-style binaries", func(t *testing.T) {
		fakeBin(t, "fakecheck", "#!/bin/sh\nexit 1\n")
		old := noUpdateNonzeroExit["fakecheck"]
		noUpdateNonzeroExit["fakecheck"] = true
		defer func() { noUpdateNonzeroExit["fakecheck"] = old }()
		updates, err := listUpdatesCmd("fakecheck", nil, func(string) []string { return nil })()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(updates) != 0 {
			t.Fatalf("expected no updates, got %v", updates)
		}
	})

	t.Run("missing binary is an error", func(t *testing.T) {
		if _, err := listUpdatesCmd("definitely-not-a-real-binary-xyz", nil, func(string) []string { return nil })(); err == nil {
			t.Fatal("expected error for missing binary")
		}
	})
}

func TestParseDnfUpdatesSkipsErrorLines(t *testing.T) {
	output := "Last metadata expiration check: 1:00:00 ago on Sat Sep  6 20:00:00 2026.\n" +
		"Error: Failed to synchronize cache for repo 'updates'\n" +
		"\n" +
		"NotPackageLine\n" +
		"mypkg.x86_64  2.0.1  fedora\n"
	want := []string{"mypkg.x86_64"}
	got := parseDnfUpdates(output)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDnfUpdates() = %v, want %v", got, want)
	}
}
