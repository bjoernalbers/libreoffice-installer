// libreoffice-installer installs LibreOffice on macOS
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
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
	// Algorithm:
	//
	// Get architecture of host (Intel vs. Apple Silicon)
	// Set current LibreOffice version
	// Read installed LibreOffice version and installation method (Mac App Store vs. download)
	//   Exit successfully if expected LibreOffice version (or newer) is already installed
	// Download SHA-256 checksum file for current LibreOffice Disk Image
	//   Abort with explanation when download failed
	// Read checksum from downloaded checksumm file
	//   Abort with explanation when read failed
	// Download Disk Image of current LibreOffice version by host architecture
	//   Abort with explanation when download failed
	// Compare expected checksum with actual checksum of downloaded Disk Image
	//   Abort with explanation when verification failed
	// Quit LibreOffice when running
	//   Abort when quit failed
	// Remove directory /Applications/LibreOffice.app
	//   Abort when quit failed
	// Mount Disk Image to temp. folder
	//   Abort with explanation when mount failed
	// Defer unmount of Disk Image
	// Copy LibreOffice.app from mounted Disk Image to /Applications
	//   Abort with explanation when copy failed
	// Exit successfully with status message

	app := App{"/Applications/LibreOffice.app"}

	// Exit if the expected LibreOffice version is already installed
	if !needsInstallation(app, LibreOfficeVersion) {
		return
	}
	log.Println("Installation required.")

	/*
		// Download LibreOffice Disk Image
		url := URL{LibreOfficeVersion, runtime.GOARCH}
		diskImageURL, err := url.DiskImage()
		if err != nil {
			log.Fatal(err)
		}
		dmgPath, err := download(diskImageURL)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("LibreOffice %s downloaded to %q", LibreOfficeVersion, dmgPath)
	*/
}

// needsInstallation returns true if installation of LibreOffice is required.
func needsInstallation(app App, version string) bool {
	// true if LibreOffice is not installed at all
	if app.IsMissing() {
		log.Println("LibreOffice is missing.")
		return true
	}
	// true if LibreOffice has been installed from Mac App Store since that
	// version is currently not fit for production:
	// https://bugs.documentfoundation.org/show_bug.cgi?id=153927
	if app.InstalledFromMAS() {
		log.Println("LibreOffice has been installed from Mac App Store.")
		return true
	}
	// true if current LibreOffice version is outdated or the version could not
	// be determined.
	older, err := app.IsOlderThan(version)
	if err != nil || older {
		log.Println("LibreOffice is probably outdated.")
		return true
	}
	return false
}

// Download downloads the given URL to the named file.
func Download(name, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %q: %s", url, resp.Status)
	}
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// checksum returns the SHA-256 checksum from input filename
func checksum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(content)), nil
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
