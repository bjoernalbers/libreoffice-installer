package app

import (
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
