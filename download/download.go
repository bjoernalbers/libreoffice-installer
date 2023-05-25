// Package download provides utilities for downloading LibreOffice resources.
package download

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

const VersionURL = "https://update.libreoffice.org/description"

// LatestVersion returns the latest "fresh" version number of LibreOffice.
func LatestVersion(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download page: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download page: %s: %s", url, resp.Status)
	}
	versions := downloadableVersions(resp.Body)
	expectedVersions := 2
	if foundVersions := len(versions); foundVersions != expectedVersions {
		return "", fmt.Errorf("download page: expected %d versions, found %d: %q", expectedVersions, foundVersions, versions)
	}
	// versions[0] is experimental, versions[1] is stable a.k.a. "fresh"
	return versions[1], nil
}

// downloadableVersions extract the version numbers from the download page.
func downloadableVersions(r io.Reader) []string {
	var versions []string
	var versionStartTag bool
	tokenizer := html.NewTokenizer(r)
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			return versions
		case html.StartTagToken:
			token := tokenizer.Token()
			if token.Data != "span" {
				continue
			}
			for _, attribute := range token.Attr {
				if attribute.Key == "class" && attribute.Val == "dl_version_number" {
					versionStartTag = true
				}
			}
		case html.TextToken:
			if !versionStartTag {
				continue
			}
			versions = append(versions, tokenizer.Token().Data)
		case html.EndTagToken:
			versionStartTag = false
		}
	}
}

// DiskImage downloads a LibreOffice disk image by version and architecture and
// returns its path.
//
// If the download or its checksum verification fails an error is returned.
func DiskImage(version, arch string) (string, error) {
	dmgURL, err := diskImageURL(version, arch)
	if err != nil {
		return "", err
	}
	diskImageFilename := filepath.Join(os.TempDir(), path.Base(dmgURL))
	err = download(diskImageFilename, dmgURL)
	if err != nil {
		return "", err
	}
	checksumURL := dmgURL + ".sha256"
	checksumFilename := filepath.Join(os.TempDir(), path.Base(checksumURL))
	err = download(checksumFilename, checksumURL)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(checksumFilename)
	if err != nil {
		return "", err
	}
	expectedChecksum := strings.Split(string(content), " ")[0]
	actualChecksum, err := checksum(diskImageFilename)
	if err != nil {
		return "", err
	}
	if actualChecksum != expectedChecksum {
		return "", fmt.Errorf("downloaded disk image has wrong checkum: %s",
			actualChecksum)
	}
	return diskImageFilename, nil
}

// diskImageURL returns a LibreOffice disk image URL by version and architecture.
//
// If the architecture is unsupported an error is returned.
func diskImageURL(version, arch string) (string, error) {
	var dirArch, baseArch string
	switch arch {
	case "arm64":
		dirArch = "aarch64"
		baseArch = dirArch
	case "amd64":
		dirArch = "x86_64"
		baseArch = "x86-64"
	default:
		return "", fmt.Errorf("unsupported architecture: %q", arch)
	}
	return fmt.Sprintf("https://download.documentfoundation.org/libreoffice/stable/%s/mac/%s/LibreOffice_%s_MacOS_%s.dmg",
		version, dirArch, version, baseArch), nil
}

// download downloads the given URL to the named file.
//
// If the download fails an error is returned.
func download(name, url string) error {
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

// checksum returns the SHA-256 checksum from named file
func checksum(name string) (string, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %v", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(content)), nil
}
