package dmg

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Download downloads a LibreOffice disk image by version and architecture and
// returns its path.
//
// If the download or the checksum verification fails an error is returned.
func Download(version, arch string) (string, error) {
	diskImageUrl, err := url(version, arch)
	if err != nil {
		return "", err
	}
	diskImageFilename := filepath.Join(os.TempDir(), path.Base(diskImageUrl))
	err = download(diskImageFilename, diskImageUrl)
	if err != nil {
		return "", err
	}
	checksumUrl := diskImageUrl + ".sha256"
	checksumFilename := filepath.Join(os.TempDir(), path.Base(checksumUrl))
	err = download(checksumFilename, checksumUrl)
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
		return "", fmt.Errorf("Checksum validation failed: %s", diskImageFilename)
	}
	return diskImageFilename, nil
}

// url returns a LibreOffice disk image URL by version and architecture.
//
// If the architecture is unsupported an error is returned.
func url(version, arch string) (string, error) {
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
