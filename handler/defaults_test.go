package handler

import "testing"

func TestNormalizePushAddress(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "localhost:10087"},
		{"self", "localhost:10087"},
		{"all", "255.255.255.255:10087"},
		{"ll", "255.255.255.255:10087"},
		{"192.168.1.1", "192.168.1.1:10087"},
		{"192.168.1.1:20000", "192.168.1.1:20000"},
		{"[::1]", "[::1]:10087"},
		{"[::1]:20000", "[::1]:20000"},
	}
	for _, c := range cases {
		if got := normalizePushAddress(c.in); got != c.want {
			t.Errorf("normalizePushAddress(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
