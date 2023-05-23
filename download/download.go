// Package download provides utilities for downloading LibreOffice resources.
package download

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

const VersionDownloadURL = "https://update.libreoffice.org/description"

// LatestVersion returns the latest "fresh" version number of LibreOffice.
func LatestVersion(url string) (string, error) {
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
