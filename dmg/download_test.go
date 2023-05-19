package dmg

import "testing"

func TestDownload(t *testing.T) {
	tests := []struct {
		version string
		arch    string
		wantErr bool
	}{
		{
			"7.4.7",
			"invalid",
			true,
		},
		{
			"invalid",
			"arm64",
			true,
		},
		{
			"7.4.7",
			"arm64",
			false,
		},
	}
	for _, test := range tests {
		_, err := Download(test.version, test.arch)
		if (err != nil) != test.wantErr {
			t.Fatalf("Download(%q, %q) err = %v, wantErr: %v",
				test.version, test.arch, err, test.wantErr)
		}
	}
}

func TestUrl(t *testing.T) {
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
		got, err := url(test.version, test.arch)
		if (err != nil) != test.wantErr {
			t.Fatalf("url(%q, %v) err = %v, wantErr: %v", test.version, test.arch, err, test.wantErr)
		}
		if got != test.want {
			t.Errorf("url(%q, %q):\ngot:\t%q\nwant:\t%q", test.version, test.arch, got, test.want)
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
