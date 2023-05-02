package main

import "fmt"

type URL struct {
	Version string
	Arch    string
}

// DiskImage returns URL for the Disk Image.
func (url URL) DiskImage() (string, error) {
	var arch, filenameArch string
	switch url.Arch {
	case "arm64":
		arch = "aarch64"
		filenameArch = arch
	case "x86_64":
		arch = url.Arch
		filenameArch = "x86-64"
	default:
		return "", fmt.Errorf("Unknown architekture: %q", url.Arch)
	}
	return fmt.Sprintf("https://download.documentfoundation.org/libreoffice/stable/%s/mac/%s/LibreOffice_%s_MacOS_%s.dmg",
		url.Version, arch, url.Version, filenameArch), nil
}

// Checksum returns URL for the Disk Image's SHA-256 Checksum.
func (url URL) Checksum() (string, error) {
	diskImageURL, err := url.DiskImage()
	if err != nil {
		return "", err
	}
	return diskImageURL + ".sha256", nil
}
