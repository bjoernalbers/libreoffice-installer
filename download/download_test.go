package download

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLatestVersion(t *testing.T) {
	if _, err := LatestVersion(""); err == nil {
		t.Fatal("LatestVersion(): no error with invalid URL.")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	if _, err := LatestVersion(server.URL); err == nil {
		t.Fatal("LatestVersion(): no error on HTTP 404.")
	}
	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", `<!DOCTYPE html></html>`)
	}))
	if _, err := LatestVersion(server.URL); err == nil {
		t.Fatal("LatestVersion(): no error when HTML contains no versions.")
	}
	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", `<!DOCTYPE html>
<head>
  <title>Download LibreOffice</title>
</head>
<body class="Download" id="download-libreoffice">
  <span class="dl_version_number">7.5.3</span><br />
  <span class="dl_description_text">If you're a technology enthusiast, early adopter or power user, this version is for you!</span>
</body>
</html>
`)
	}))
	if _, err := LatestVersion(server.URL); err == nil {
		t.Fatal("LatestVersion(): no error when HTML contains only one version.")
	}
	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", `<!DOCTYPE html>
<head>
  <title>Download LibreOffice</title>
</head>
<body class="Download" id="download-libreoffice">
  <span class="dl_version_number">7.5.3</span><br />
  <span class="dl_description_text">If you're a technology enthusiast, early adopter or power user, this version is for you!</span>
  <span class="dl_version_number">7.4.7</span><br />
  <span class="dl_description_text">This version is slightly older and does not have the latest features, but it has been tested for longer. For business deployments, we <a href="https://www.libreoffice.org/download/libreoffice-in-business/">strongly recommend support from certified partners</a> which also offer long-term support versions of LibreOffice.</span>
</body>
</html>
`)
	}))
	want := "7.4.7"
	got, err := LatestVersion(server.URL)
	server.Close()
	if err != nil {
		t.Fatalf("LatestVersion() err: %v", err)
	}
	if got != want {
		t.Fatalf("LatestVersion() = %q, want: %q", got, want)
	}
}

func TestDownloadableVersions(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			"",
			[]string{},
		},
		{
			"/invalid",
			[]string{},
		},
		{
			`<div>1.2.3</span>`,
			[]string{},
		},
		{
			`<div class="dl_version_number">1.2.3</span>`,
			[]string{},
		},
		{
			`<span>1.2.3</span>`,
			[]string{},
		},
		{
			`<span class="wrongclass">1.2.3</span>`,
			[]string{},
		},
		{
			`<span class="dl_version_number">1.2.3</span>`,
			[]string{"1.2.3"},
		},
		{
			`<span class="dl_version_number">1.2.3</span>
			<span class="dl_version_number">4.5.6</span>
			<span class="dl_version_number">7.8.9</span>`,
			[]string{"1.2.3", "4.5.6", "7.8.9"},
		},
	}
	for _, test := range tests {
		if got := downloadableVersions(strings.NewReader(test.input)); !equal(got, test.want) {
			t.Errorf("downloadableVersions() = %q, want: %q", got, test.want)
		}
	}
}

func equal(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func TestDiskImageURL(t *testing.T) {
	tests := []struct {
		version string
		arch    string
		want    string
		wantErr bool
	}{
		{
			"7.4.6",
			"amd64",
			"https://download.documentfoundation.org/libreoffice/stable/7.4.6/mac/x86_64/LibreOffice_7.4.6_MacOS_x86-64.dmg",
			false,
		},
		{
			"7.4.6",
			"arm64",
			"https://download.documentfoundation.org/libreoffice/stable/7.4.6/mac/aarch64/LibreOffice_7.4.6_MacOS_aarch64.dmg",
			false,
		},
		{
			"7.5.2",
			"amd64",
			"https://download.documentfoundation.org/libreoffice/stable/7.5.2/mac/x86_64/LibreOffice_7.5.2_MacOS_x86-64.dmg",
			false,
		},
		{
			"7.5.2",
			"arm64",
			"https://download.documentfoundation.org/libreoffice/stable/7.5.2/mac/aarch64/LibreOffice_7.5.2_MacOS_aarch64.dmg",
			false,
		},
		{
			"7.4.6",
			"",
			"",
			true,
		},
	}
	for _, test := range tests {
		got, err := diskImageURL(test.version, test.arch)
		if (err != nil) != test.wantErr {
			t.Fatalf("diskImageURL(%q, %v) err = %v, wantErr: %v", test.version, test.arch, err, test.wantErr)
		}
		if got != test.want {
			t.Errorf("diskImageURL(%q, %q):\ngot:\t%q\nwant:\t%q", test.version, test.arch, got, test.want)
		}
	}
}

func TestChecksum(t *testing.T) {
	want := "9a13c79fd185a5737cf5a28741143fa67d2b9980ed33a2ce61a24c67fe03dae8"
	got, err := checksum("../LICENSE")
	if err != nil {
		t.Fatalf("checksum() return an error: %v", err)
	}
	if got != want {
		t.Fatalf("checksum() = %v, want: %v", got, want)
	}
}
