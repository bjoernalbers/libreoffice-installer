package dmg

import "testing"

func TestMountpointFromPlist(t *testing.T) {
	tests := []struct {
		input   []byte
		want    string
		wantErr bool
	}{
		{
			[]byte(""),
			"",
			true,
		},
		{
			[]byte("<invalid xml"),
			"",
			true,
		},
		{
			[]byte(`
        <key>mount-point</key>
        <string>/Volumes/LibreOffice</string>
      `),
			"/Volumes/LibreOffice",
			false,
		},
		{
			[]byte(`
        <key>mount-point</key>
        <string>/Volumes/LibreOffice 1</string>
        <key>mount-point</key>
        <string>/Volumes/LibreOffice 2</string>
      `),
			"/Volumes/LibreOffice 1",
			false,
		},
	}
	for _, test := range tests {
		got, err := mountpointFromPlist(test.input)
		if (err != nil) != test.wantErr {
			t.Fatalf("mountpointFromPlist() err: %v, wantErr: %v", err, test.wantErr)
		}
		if got != test.want {
			t.Fatalf("mountpointFromPlist() = %q, want: %q", got, test.want)
		}
	}
}
