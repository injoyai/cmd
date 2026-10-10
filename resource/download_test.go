package resource

import "testing"

func TestProxySkipsWhenUrlMatchesIgnoreRule(t *testing.T) {
	op := &Config{
		Resource:     "https://example.com/file.bin",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy.local:8080",
		ProxyIgnore:  []string{"example\\.com"},
	}

	if got := op.Proxy(); got != "" {
		t.Fatalf("Proxy() = %q, want empty when url matches ignore rule", got)
	}
}

func TestProxyReturnsAddressWhenNotIgnored(t *testing.T) {
	op := &Config{
		Resource:     "https://example.com/file.bin",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy.local:8080",
		ProxyIgnore:  []string{"localhost", "127\\.0\\.0\\.1"},
	}

	if got := op.Proxy(); got != "http://proxy.local:8080" {
		t.Fatalf("Proxy() = %q, want %q", got, "http://proxy.local:8080")
	}
}

func TestProxyEmptyWhenDisabled(t *testing.T) {
	op := &Config{
		Resource:     "https://example.com/file.bin",
		ProxyAddress: "http://proxy.local:8080",
	}

	if got := op.Proxy(); got != "" {
		t.Fatalf("Proxy() = %q, want empty when proxy disabled", got)
	}
}
