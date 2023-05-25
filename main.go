// libreoffice-installer installs LibreOffice on macOS
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

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
	app := App{appPath}
	if !needsInstallation(app, latestVersion) {
		log.Println("LibreOffice", latestVersion, "or newer is already installed.")
		return
	}
	diskimage, err := download.DiskImage(latestVersion, runtime.GOARCH)
	if err != nil {
		log.Fatal(err)
	}
	err = quitLibreOffice()
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

func quitLibreOffice() error {
	procname := "soffice"
	pids, err := pgrep(procname)
	if err != nil {
		return err
	}
	if len(pids) == 0 {
		return nil
	}
	users, err := processUsers(pids)
	if err != nil {
		return err
	}
	for _, user := range users {
		err := quitApp("LibreOffice", user)
		if err != nil {
			return err
		}
	}
	pids, err = pgrep(procname)
	if len(pids) != 0 {
		return fmt.Errorf("unable to quit LibreOffice")
	}
	return nil
}

// pgrep returns list of process IDs (PIDs) found by process name.
func pgrep(name string) ([]int, error) {
	var pids []int
	cmd := exec.Command("pgrep", "-x", name)
	output, err := cmd.Output()
	if err != nil {
		exiterr, ok := err.(*exec.ExitError)
		if !ok {
			return nil, err
		}
		if exiterr.ProcessState.ExitCode() != 1 {
			return nil, exiterr
		}
		return nil, nil
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		pid, err := strconv.Atoi(scanner.Text())
		if err != nil {
			return nil, err
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// processUsers returns list of user names by corresponding PIDs
func processUsers(pids []int) ([]string, error) {
	cmd := exec.Command("ps", "-p", pidsToOpt(pids), "-o", "user=")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	users := make(map[string]int)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		users[scanner.Text()]++
	}
	var uniqueUsers []string
	for u := range users {
		uniqueUsers = append(uniqueUsers, u)
	}
	return uniqueUsers, nil
}

// pidsToOpt returns PIDs as string joined by commas.
func pidsToOpt(pids []int) string {
	var s []string
	for _, pid := range pids {
		s = append(s, strconv.Itoa(pid))
	}
	return strings.Join(s, ",")
}

// quitApp quits macOS application by username.
func quitApp(app, username string) error {
	appleScript := fmt.Sprintf("quit app %q", app)
	var cmd *exec.Cmd
	cmd = exec.Command("sudo", "--non-interactive", "--user", username, "osascript", "-e", appleScript)
	log.Print(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
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
