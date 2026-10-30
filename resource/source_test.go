package resource

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/injoyai/cmd/config"
)

func TestGetURLsFromList(t *testing.T) {
	previous := config.File
	config.File = config.Open(filepath.Join(t.TempDir(), "config.yaml"))
	t.Cleanup(func() { config.File = previous })
	first := "https://one.example/{name}?a=1,b=2"
	second := "https://two.example/{name}"
	config.File.Set("resource.urls", []string{first, "", second})
	if got, want := GetUrls(), []Url{Url(first), Url(second), DefaultUrl}; !reflect.DeepEqual(got, want) {
		t.Fatalf("GetUrls() = %q, want %q", got, want)
	}
	config.File.Set("resource.urls", []string{})
	if got := GetUrls(); !reflect.DeepEqual(got, []Url{DefaultUrl}) {
		t.Fatalf("empty list must retain default source: %q", got)
	}
}
