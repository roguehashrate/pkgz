package linux

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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

func TestMatchLines(t *testing.T) {
	t.Run("apt style first-field names", func(t *testing.T) {
		out := "firefox-esr - Powerful, extensible web browser\n" +
			"fireshot - screenshot tool\n" +
			"firestorm - strategy game\n"
		got := matchLines(out, func(f []string) string {
			return f[0]
		}, func(line string, f []string) bool {
			return strings.Contains(strings.ToLower(line), "firest")
		})
		if !reflect.DeepEqual(got, []string{"firestorm"}) {
			t.Errorf("matchLines = %v, want [firestorm]", got)
		}
	})

	t.Run("flatpak style appid (name)", func(t *testing.T) {
		out := "com.mozilla.Firefox\tMozilla Firefox\n" +
			"com.mozilla.Firefox.Nightly\tMozilla Firefox Nightly\n"
		got := matchLines(out, func(f []string) string {
			return f[0] + " (" + f[1] + ")"
		}, nil)
		want := []string{
			"com.mozilla.Firefox (Mozilla Firefox)",
			"com.mozilla.Firefox.Nightly (Mozilla Firefox Nightly)",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("matchLines = %v, want %v", got, want)
		}
	})

	t.Run("caps matches at 15", func(t *testing.T) {
		var out strings.Builder
		for i := 0; i < 20; i++ {
			out.WriteString("pkg" + strconv.Itoa(i) + " - thing\n")
		}
		got := matchLines(out.String(), func(f []string) string { return f[0] }, nil)
		if len(got) != 15 {
			t.Errorf("matchLines capped at %d, want 15", len(got))
		}
	})
}

func TestCommandSourceMemoizesProbes(t *testing.T) {
	var availCalls, instCalls int
	c := &commandSource{
		available: func(app string) (bool, error) {
			availCalls++
			return true, nil
		},
		installed: func(app string) (bool, error) {
			instCalls++
			return false, nil
		},
	}
	for i := 0; i < 3; i++ {
		if ok, err := c.Available("vim"); err != nil || !ok {
			t.Fatalf("Available(vim) = %v, %v", ok, err)
		}
		if ok, err := c.Installed("vim"); err != nil || ok {
			t.Fatalf("Installed(vim) = %v, %v", ok, err)
		}
	}
	if availCalls != 1 {
		t.Errorf("available probed %d times, want 1", availCalls)
	}
	if instCalls != 1 {
		t.Errorf("installed probed %d times, want 1", instCalls)
	}

	if ok, err := c.Available("VIM"); err != nil || !ok {
		t.Errorf("Available memoized by case-insensitive key: %v, %v", ok, err)
	}
	if availCalls != 1 {
		t.Errorf("available probed %d times after case-different call, want 1", availCalls)
	}
}
