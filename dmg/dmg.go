// dmg handles Apple Disk Images (.dmg).
package dmg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strings"
)

// InstallApplication install an application from disk image to

// Attach attaches the named disk image and returns its mount point.
func Attach(name string) (string, error) {
	cmd := exec.Command("hdiutil", "attach", name, "-nobrowse", "-plist")
	stdout, err := cmd.Output()
	if err != nil {
		stderr := err.(*exec.ExitError).Stderr
		firstLine, _, _ := strings.Cut(string(stderr), "\n")
		return "", fmt.Errorf("%s: %s", firstLine, name)
	}
	mountpoint, err := mountpointFromPlist(stdout)
	if err != nil {
		return "", err
	}
	return mountpoint, nil
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

// mountpointFromPlist returns first mountpoint from plist.
func mountpointFromPlist(plist []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(plist))
	var tag string
	var isMountpointKey bool
	for {
		token, err := dec.Token()
		if err != nil {
			break
		}
		switch token := token.(type) {
		case xml.StartElement:
			tag = token.Name.Local
		case xml.CharData:
			value := string(token)
			if tag == "key" && value == "mount-point" {
				isMountpointKey = true
			} else if isMountpointKey && tag == "string" {
				return value, nil
			}
		}
	}
	return "", fmt.Errorf("dmg: no mountpoint found")
}
