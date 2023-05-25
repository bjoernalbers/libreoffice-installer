package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
)

type App struct {
	Path string
}

func (a *App) IsMissing() bool {
	_, err := os.Stat(a.Path)
	if err == nil {
		return false
	}
	return true
}

// version returns the app's version
func (a *App) version() (string, error) {
	cmd := exec.Command("/usr/bin/defaults",
		"read",
		filepath.Join(a.Path, "Contents/Info.plist"),
		"CFBundleShortVersionString")
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}

// InstalledFromMAS returns a boolean indicating whether the app has been
// installed from the Mac App Store (MAS).
//
// If the app contains an App Store receipts it was probably installed from
// there.
// To be 100 % sure we'd have to validate the receipt, though:
// https://stackoverflow.com/questions/30339568/how-can-i-know-if-an-app-is-installed-from-apple-app-store-or-other-stores-in-os
func (a *App) InstalledFromMAS() bool {
	masReceipt := filepath.Join(a.Path, "Contents", "_MASReceipt", "receipt")
	_, err := os.Stat(masReceipt)
	if err == nil {
		return true
	}
	return false
}

// IsOlderThan returns a boolean indication wether the app is older than the
// given version.
// An error might be returned if any of the versions is invalid or the current
// version could not be obtained.
func (a *App) IsOlderThan(otherVersion string) (bool, error) {
	thisVersion, err := a.version()
	if err != nil {
		return false, err
	}
	this, err := version.NewVersion(thisVersion)
	if err != nil {
		return false, err
	}
	other, err := version.NewVersion(otherVersion)
	if err != nil {
		return false, err
	}
	return this.LessThan(other), nil
}
