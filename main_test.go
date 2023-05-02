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
	name := filepath.Join(os.TempDir(), "LibreOffice_7.4.6_MacOS_x86-64.dmg.md5")
	url := "https://download.documentfoundation.org/libreoffice/stable/7.4.6/mac/x86_64/LibreOffice_7.4.6_MacOS_x86-64.dmg.md5"
	err := Download(name, url)
	if err != nil {
		t.Fatalf("Download() return an error: %v", err)
	}
	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("Download(): failed to read downloaded file: %v", err)
	}
	got := string(content)
	want := "e5df6adeac2b87298df1b3fe691b486d  LibreOffice_7.4.6_MacOS_x86-64.dmg\n"
	if got != want {
		t.Fatalf("Download(): expected download differs:\ngot:\t%q\nwant:\t%q", got, want)
	}
	url = "https://download.documentfoundation.org/missing"
	err = Download(name, url)
	if err == nil {
		t.Fatal("Download(): got no error on missing file")
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
