package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsMissing(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{
			"/Applications/DoesNotExist.app",
			true,
		},
		{
			"/System/Library/CoreServices/Finder.app",
			false,
		},
	}
	for _, test := range tests {
		app := App{test.path}
		if got := app.IsMissing(); got != test.want {
			t.Errorf("IsMissing(%v) = %v, want: %v", test.path, got, test.want)
		}
	}
}

func TestDownload(t *testing.T) {
	return // Disabled test because it queries external resources
	url := "https://download.documentfoundation.org/libreoffice/stable/7.4.6/mac/x86_64/LibreOffice_7.4.6_MacOS_x86-64.dmg.md5"
	want := filepath.Join(os.TempDir(), "LibreOffice_7.4.6_MacOS_x86-64.dmg.md5")
	got, err := download(url)
	if err != nil {
		t.Fatalf("download() return an error: %v", err)
	}
	if got != want {
		t.Fatalf("download() = %v, want: %v", got, want)
	}
	if _, err = os.Stat(want); err != nil {
		t.Fatalf("download(): Problem with downloaded file: %v", err)
	}
}

func TestChecksum(t *testing.T) {
	want := "9a13c79fd185a5737cf5a28741143fa67d2b9980ed33a2ce61a24c67fe03dae8"
	got, err := checksum("LICENSE")
	if err != nil {
		t.Fatalf("checksum() return an error: %v", err)
	}
	if got != want {
		t.Fatalf("checksum() = %v, want: %v", got, want)
	}
}

func TestDmgURL(t *testing.T) {
	tests := []struct {
		version string
		arch    string
		want    string
		wantErr bool
	}{
		{
			"7.4.6",
			"x86_64",
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
			"x86_64",
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
		got, err := dmgURL(test.version, test.arch)
		if (err != nil) != test.wantErr {
			t.Errorf("dmgTest(%q, %q) err = %v, wantErr: %v", test.version, test.arch, err, test.wantErr)
		}
		if got != test.want {
			t.Errorf("dmgTest(%q, %q):\ngot:\t%q\nwant:\t%q", test.version, test.arch, got, test.want)
		}
	}
}
