// libreoffice-installer installs LibreOffice on macOS
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
)

const (
	LibreOfficeVersion = "7.4.6"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("libreoffice-installer: ")
}

func main() {
	app := App{"/Applications/LibreOffice.app"}

	// Exit if the expected LibreOffice version is already installed
	if !needsInstallation(app, LibreOfficeVersion) {
		return
	}
	fmt.Println("Installation required.")
}

// needsInstallation returns true if installation of LibreOffice is required.
func needsInstallation(app App, version string) bool {
	// true if LibreOffice is not installed at all
	if app.IsMissing() {
		fmt.Println("LibreOffice is missing.")
		return true
	}
	// true if LibreOffice has been installed from Mac App Store since that
	// version is currently not fit for production:
	// https://bugs.documentfoundation.org/show_bug.cgi?id=153927
	if app.InstalledFromMAS() {
		fmt.Println("LibreOffice has been installed from Mac App Store.")
		return true
	}
	// true if current LibreOffice version is outdated or the version could not
	// be determined.
	older, err := app.IsOlderThan(version)
	if err != nil || older {
		fmt.Println("LibreOffice is probably outdated.")
		return true
	}
	return false
}

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
// version could not be optained.
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
