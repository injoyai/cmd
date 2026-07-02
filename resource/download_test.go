package resource

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestProxyUsesCompiledIgnoreRegex(t *testing.T) {
	op := &Config{
		Resource:     "https://example.com/file.bin",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy.local:8080",
		ProxyIgnore:  []string{"example\\.com"},
	}

	if got := op.Proxy(); got != "" {
		t.Fatalf("Proxy() = %q, want empty when url matches ignore rule", got)
	}
	if len(op.proxyIgnoreRegex) != 1 {
		t.Fatalf("compiled regex count = %d, want 1", len(op.proxyIgnoreRegex))
	}
	first := op.proxyIgnoreRegex[0]
	if first == nil {
		t.Fatal("compiled regex is nil")
	}
	_ = op.Proxy()
	if op.proxyIgnoreRegex[0] != first {
		t.Fatal("compiled regex pointer changed, want cached regex reuse")
	}
}

func TestHTTPClientUsesProxyAddress(t *testing.T) {
	var proxyHits int32
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&proxyHits, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxyServer.Close()

	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer targetServer.Close()

	op := &Config{
		Resource:     targetServer.URL,
		ProxyEnable:  true,
		ProxyAddress: proxyServer.URL,
	}
	client, err := op.HTTPClient()
	if err != nil {
		t.Fatalf("HTTPClient() error = %v", err)
	}
	resp, err := client.Get(targetServer.URL)
	if err != nil {
		t.Fatalf("client.Get() error = %v", err)
	}
	resp.Body.Close()
	if atomic.LoadInt32(&proxyHits) == 0 {
		t.Fatal("expected request to go through proxy server")
	}
}

func TestHTTPClientSkipsProxyWhenIgnored(t *testing.T) {
	var targetHits int32
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&targetHits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer targetServer.Close()

	proxyURL, err := url.Parse(targetServer.URL)
	if err != nil {
		t.Fatalf("parse target url: %v", err)
	}
	op := &Config{
		Resource:     targetServer.URL,
		ProxyEnable:  true,
		ProxyAddress: proxyURL.String(),
		ProxyIgnore:  []string{"127\\.0\\.0\\.1", "localhost"},
	}
	client, err := op.HTTPClient()
	if err != nil {
		t.Fatalf("HTTPClient() error = %v", err)
	}
	resp, err := client.Get(targetServer.URL)
	if err != nil {
		t.Fatalf("client.Get() error = %v", err)
	}
	resp.Body.Close()
	if atomic.LoadInt32(&targetHits) == 0 {
		t.Fatal("expected request to reach target server directly")
	}
}
