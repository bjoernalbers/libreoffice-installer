// libreoffice-installer installs LibreOffice on macOS
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/bjoernalbers/libreoffice-installer/dmg"
	"github.com/hashicorp/go-version"
	"golang.org/x/net/html"
)

const LibreOfficeDownloadURL = "https://update.libreoffice.org/description"

func init() {
	log.SetFlags(0)
	log.SetPrefix("libreoffice-installer: ")
}

func main() {
	LibreOfficeVersion, err := latestVersion(LibreOfficeDownloadURL)
	if err != nil {
		log.Fatal(err)
	}
	volume := os.Args[3]
	app := App{filepath.Join(volume, "/Applications/LibreOffice.app")}
	if !needsInstallation(app, LibreOfficeVersion) {
		log.Println("LibreOffice", LibreOfficeVersion, "or newer is already installed.")
		return
	}
	url := URL{LibreOfficeVersion, runtime.GOARCH}
	checksumURL, err := url.Checksum()
	if err != nil {
		log.Fatal(err)
	}
	diskImageURL, err := url.DiskImage()
	if err != nil {
		log.Fatal(err)
	}
	checksumFilename := filepath.Join(os.TempDir(), path.Base(checksumURL))
	diskImageFilename := filepath.Join(os.TempDir(), path.Base(diskImageURL))
	err = Download(checksumFilename, checksumURL)
	if err != nil {
		log.Fatal(err)
	}
	err = Download(diskImageFilename, diskImageURL)
	if err != nil {
		log.Fatal(err)
	}
	content, err := os.ReadFile(checksumFilename)
	if err != nil {
		log.Fatal(err)
	}
	expectedChecksum := strings.Split(string(content), " ")[0]
	actualChecksum, err := Checksum(diskImageFilename)
	if err != nil {
		log.Fatal(err)
	}
	if actualChecksum != expectedChecksum {
		log.Fatal("Checksum validation failed: ", diskImageFilename)
	}

	err = quitLibreOffice()
	if err != nil {
		log.Fatal(err)
	}

	// Remove directory /Applications/LibreOffice.app

	mountpoint, err := dmg.Attach(diskImageFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer dmg.Detach(mountpoint)
	cmd := exec.Command("cp", "-R", filepath.Join(mountpoint, "LibreOffice.app"), filepath.Join(volume, "Applications"))
	log.Print(cmd) // debug
	err = cmd.Run()
	if err != nil {
		log.Fatal("Copy of LibreOffice failed: ", err)
	}
	log.Print("Installation completed successfully")
}

// latestVersion returns the lastest available version of LibreOffice.
func latestVersion(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", url, resp.Status)
	}
	version, err := parseVersion(resp.Body)
	if err != nil {
		return "", err
	}
	return version, nil
}

// parseVersion parsed the version from reader
func parseVersion(r io.Reader) (string, error) {
	z := html.NewTokenizer(r)
	var foundVersionNumber bool
	var versions []string
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if err := z.Err(); err != io.EOF {
				return "", err
			}
			if len(versions) != 2 {
				return "", fmt.Errorf("expected two version numbers: %v", versions)
			}
			return versions[1], nil
		case html.StartTagToken:
			t := z.Token()
			if t.Data != "span" {
				continue
			}
			for _, a := range t.Attr {
				if a.Key == "class" && a.Val == "dl_version_number" {
					foundVersionNumber = true
				}
			}
		case html.TextToken:
			if !foundVersionNumber {
				continue
			}
			t := z.Token()
			versions = append(versions, t.Data)
		case html.EndTagToken:
			foundVersionNumber = false
		}
	}
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

// Checksum returns the SHA-256 checksum from named file
func Checksum(name string) (string, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %v", err)
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

// quitApp quits macOS applicattion by username.
func quitApp(app, username string) error {
	appleScript := fmt.Sprintf("quit app %q", app)
	var cmd *exec.Cmd
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	if currentUser.Username != username {
		cmd = exec.Command("sudo", "--non-interactive", "--user", username, "osascript", "-e", appleScript)
	} else {
		cmd = exec.Command("osascript", "-e", appleScript)
	}
	log.Print(cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
