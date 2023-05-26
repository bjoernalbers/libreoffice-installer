// libreoffice-installer installs LibreOffice on macOS
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/bjoernalbers/libreoffice-installer/app"
	"github.com/bjoernalbers/libreoffice-installer/dmg"
	"github.com/bjoernalbers/libreoffice-installer/download"
	"github.com/hashicorp/go-version"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("libreoffice-installer: ")
}

func main() {
	latestVersion, err := download.LatestVersion(download.VersionURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("latest version: %s", latestVersion)
	volume := os.Args[3]
	appPath := filepath.Join(volume, "/Applications/LibreOffice.app")
	a := app.App{appPath}
	if !needsInstallation(a, latestVersion) {
		log.Println("LibreOffice", latestVersion, "or newer is already installed.")
		return
	}
	diskimage, err := download.DiskImage(latestVersion, runtime.GOARCH)
	if err != nil {
		log.Fatal(err)
	}
	err = app.QuitLibreOffice()
	if err != nil {
		log.Fatal(err)
	}
	err = installApplication(appPath, diskimage)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Installation completed successfully")
}

// needsInstallation returns true if installation of LibreOffice is required.
func needsInstallation(a app.App, version string) bool {
	if a.IsMissing() {
		return true
	}
	// If LibreOffice has been installed from the Mac App Store it must be
	// replaced since that version is currently not fit for production:
	// https://bugs.documentfoundation.org/show_bug.cgi?id=153927
	if a.FromMacAppStore() {
		return true
	}
	installedVersion, err := a.Version()
	if err != nil || outdated(installedVersion, version) {
		return true
	}
	return false
}

// outdated returns true if version 1 is less than version 2 or if the versions
// are not comparable.
func outdated(version1, version2 string) bool {
	v1, err := version.NewVersion(version1)
	if err != nil {
		return true
	}
	v2, err := version.NewVersion(version2)
	if err != nil {
		return true
	}
	return v1.LessThan(v2)
}

// installApplication installs application from disk image to destination,
// which is the application's target bundle path.
func installApplication(destination, diskimage string) error {
	mountpoint, err := dmg.Attach(diskimage)
	if err != nil {
		return err
	}
	defer dmg.Detach(mountpoint)
	err = os.RemoveAll(destination)
	if err != nil {
		return err
	}
	source := filepath.Join(mountpoint, filepath.Base(destination))
	cmd := exec.Command("cp", "-R", source, destination)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %v", cmd, err)
	}
	return nil
}
