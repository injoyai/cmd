package handler

import "testing"

func TestKillCmd(t *testing.T) {
	cases := []struct{ in, want string }{
		{"12345", "taskkill /f /t /pid 12345"},
		{`"12345"`, "taskkill /f /t /pid 12345"},
		{"chrome.exe", `taskkill /f /t /im "chrome.exe"`},
		{`"chrome.exe"`, `taskkill /f /t /im "chrome.exe"`},
	}
	for _, c := range cases {
		if got := killCmd(c.in); got != c.want {
			t.Errorf("killCmd(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
