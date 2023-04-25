// libreoffice-installer installs LibreOffice on macOS
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("libreoffice-installer: ")
}

func main() {
	app := App{"/Applications/LibreOffice.app"}
	if app.IsMissing() {
		fmt.Println("LibreOffice is missing. Installation required.")
	}

	version, err := app.version()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Installed version:", version)

	if app.InstalledFromMAS() {
		fmt.Println("LibreOffice has been installed from Mac App Store.")
	}
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

// InstalledFromMAS returns true if App was installed from Mac App Store (MAS).
func (a *App) InstalledFromMAS() bool {
  masReceipt := filepath.Join(a.Path, "Contents", "_MASReceipt", "receipt")
  _, err := os.Stat(masReceipt)
  if err == nil {
    return true
  }
  return false
}
