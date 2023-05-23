// Package download provides utilities for downloading LibreOffice resources.
package download

import (
	"fmt"
	"io"
	"net/http"

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
