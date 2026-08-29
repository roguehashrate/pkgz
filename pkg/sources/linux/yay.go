package linux

import (
	"github.com/roguehashrate/pkgz/pkg/sources"
	"github.com/roguehashrate/pkgz/pkg/utils"
)

// NewYaySource returns a source backed by the yay AUR helper (Arch).
func NewYaySource(elevator *utils.Elevator) sources.Source {
	return newAurSource(elevator, "Yay (AUR)", "yay")
}
