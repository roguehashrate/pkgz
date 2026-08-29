package linux

import (
	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewParuSource returns a source backed by the paru AUR helper (Arch).
func NewParuSource(elevator *utils.Elevator) sources.Source {
	return newAurSource(elevator, "Paru (AUR)", "paru")
}
