// dmg handles Apple Disk Images (.dmg).
package dmg

import (
	"bytes"
	"encoding/xml"
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
