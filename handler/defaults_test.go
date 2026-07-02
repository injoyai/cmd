package handler

import (
	"testing"
)

func TestDefaultPortConstants(t *testing.T) {
	if DefaultPushPort != 10087 {
		t.Fatalf("DefaultPushPort = %d, want 10087", DefaultPushPort)
	}
	if DefaultTCPPort != 10086 {
		t.Fatalf("DefaultTCPPort = %d, want 10086", DefaultTCPPort)
	}
	if DefaultUDPPort != 10088 {
		t.Fatalf("DefaultUDPPort = %d, want 10088", DefaultUDPPort)
	}
	if DefaultMQTTPort != 1883 {
		t.Fatalf("DefaultMQTTPort = %d, want 1883", DefaultMQTTPort)
	}
	if DefaultSeleniumPort != 20165 {
		t.Fatalf("DefaultSeleniumPort = %d, want 20165", DefaultSeleniumPort)
	}
	if DefaultHTTPPort != 8080 {
		t.Fatalf("DefaultHTTPPort = %d, want 8080", DefaultHTTPPort)
	}
	if DefaultWebsitePort != 8000 {
		t.Fatalf("DefaultWebsitePort = %d, want 8000", DefaultWebsitePort)
	}
}

func TestNormalizePushAddress(t *testing.T) {
	cases := map[string]string{
		"":            "localhost:10087",
		"self":        "localhost:10087",
		"all":         "255.255.255.255:10087",
		"ll":          "255.255.255.255:10087",
		"1":           "255.255.255.255:10087",
		"192.168.1.2": "192.168.1.2:10087",
	}
	for in, want := range cases {
		if got := normalizePushAddress(in); got != want {
			t.Fatalf("normalizePushAddress(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveBroadcastUDPAddrUsesBroadcastIP(t *testing.T) {
	addr, err := resolveBroadcastUDPAddr("all")
	if err != nil {
		t.Fatalf("resolveBroadcastUDPAddr returned error: %v", err)
	}
	if got := addr.IP.String(); got != "255.255.255.255" {
		t.Fatalf("broadcast ip = %s, want 255.255.255.255", got)
	}
	if addr.Port != DefaultPushPort {
		t.Fatalf("broadcast port = %d, want %d", addr.Port, DefaultPushPort)
	}
}
