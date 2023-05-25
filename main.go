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
	// true if LibreOffice is not installed at all
	if a.IsMissing() {
		return true
	}
	// true if LibreOffice has been installed from Mac App Store since that
	// version is currently not fit for production:
	// https://bugs.documentfoundation.org/show_bug.cgi?id=153927
	if a.FromMacAppStore() {
		return true
	}
	// true if current LibreOffice version is outdated or the version could not
	// be determined.
	older, err := a.Outdated(version)
	if err != nil || older {
		return true
	}
	return false
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
