package app

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

// Version returns the app's version
func (a *App) Version() (string, error) {
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

// FromMacAppStore returns a boolean indicating whether the app has been
// installed from the Mac App Store (MAS).
//
// If the app contains an App Store receipts it was probably installed from
// there.
// To be 100 % sure we'd have to validate the receipt, though:
// https://stackoverflow.com/questions/30339568/how-can-i-know-if-an-app-is-installed-from-apple-app-store-or-other-stores-in-os
func (a *App) FromMacAppStore() bool {
	masReceipt := filepath.Join(a.Path, "Contents", "_MASReceipt", "receipt")
	_, err := os.Stat(masReceipt)
	if err == nil {
		return true
	}
	return false
}

func QuitLibreOffice() error {
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
			return fmt.Errorf("unable to quit LibreOffice")
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

// quitApp quits macOS application by username.
func quitApp(app, username string) error {
	appleScript := fmt.Sprintf("quit app %q", app)
	var cmd *exec.Cmd
	cmd = exec.Command("sudo", "--non-interactive", "--user", username, "osascript", "-e", appleScript)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// pidsToOpt returns PIDs as string joined by commas.
func pidsToOpt(pids []int) string {
	var s []string
	for _, pid := range pids {
		s = append(s, strconv.Itoa(pid))
	}
	return strings.Join(s, ",")
}
