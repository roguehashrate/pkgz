package linux

import (
	"reflect"
	"testing"
)

func TestParseAptUpgradable(t *testing.T) {
	output := "WARNING: apt does not have a stable CLI interface. Use with caution in scripts.\n" +
		"Listing...\n" +
		"curl/stable 8.5.0-2 amd64 [upgradable from: 8.5.0-1]\n" +
		"vim/stable 9.0.123-1 amd64 [upgradable from: 9.0.2-3]\n" +
		"\n"

	want := []string{"curl", "vim"}
	got := parseAptUpgradable(output)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseAptUpgradable() = %v, want %v", got, want)
	}
}

func TestParseAptUpgradableNoUpdates(t *testing.T) {
	output := "WARNING: apt does not have a stable CLI interface. Use with caution in scripts.\n" +
		"Listing...\n" +
		"\n"

	got := parseAptUpgradable(output)
	if len(got) != 0 {
		t.Errorf("parseAptUpgradable() = %v, want empty", got)
	}
}
