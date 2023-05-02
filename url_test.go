package main

import "testing"

func TestDiskImageURL(t *testing.T) {
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
		url := URL{test.version, test.arch}
		got, err := url.DiskImage()
		if (err != nil) != test.wantErr {
			t.Errorf("%#v.DiskImage() err = %v, wantErr: %v", url, err, test.wantErr)
		}
		if got != test.want {
			t.Errorf("%#v.DiskImage():\ngot:\t%q\nwant:\t%q", url, got, test.want)
		}
	}
}
