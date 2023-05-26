package main

import "testing"

func TestOutdated(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want bool
	}{
		{
			"invalid",
			"invalid",
			true,
		},
		{
			"1.1.1",
			"invalid",
			true,
		},
		{
			"invalid",
			"1.1.1",
			true,
		},
		{
			"1.1.1",
			"1.1.2",
			true,
		},
		{
			"1.1.1",
			"1.2.1",
			true,
		},
		{
			"1.1.1",
			"2.1.1",
			true,
		},
		{
			"1.1.1",
			"1.1.1",
			false,
		},
		{
			"1.1.2",
			"1.1.1",
			false,
		},
		{
			"1.2.1",
			"1.1.1",
			false,
		},
		{
			"2.1.1",
			"1.1.1",
			false,
		},
	}
	for _, test := range tests {
		if got := outdated(test.v1, test.v2); got != test.want {
			t.Errorf("outdated(%q, %q) = %v, want: %v", test.v1, test.v2, got, test.want)
		}
	}
}
