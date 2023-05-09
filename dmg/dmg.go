// dmg handles Apple Disk Images (.dmg).
package dmg

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Attach attaches the named disk image and returns its mount point.
func Attach(name string) (string, error) {
	dir, err := os.MkdirTemp("/tmp", "")
	if err != nil {
		return "", fmt.Errorf("attach disk image: %v", err)
	}
	cmd := exec.Command("hdiutil", "attach", name, "-mountpoint", dir, "-nobrowse")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		firstLine, _, _ := strings.Cut(stderr.String(), "\n")
		return "", fmt.Errorf("%s: %s", firstLine, name)
	}
	return dir, nil
}

// Detach detaches a disk image by the named device or mountpoint.
func Detach(name string) error {
	cmd := exec.Command("hdiutil", "detach", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		firstLine, _, _ := strings.Cut(stderr.String(), "\n")
		return fmt.Errorf("%s: %s", firstLine, name)
	}
	return nil
}
