package resource

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/injoyai/cmd/global"
)

func TestGetURLsFromList(t *testing.T) {
	previous := global.File
	global.File = global.Open(filepath.Join(t.TempDir(), "config.yaml"))
	t.Cleanup(func() { global.File = previous })
	first := "https://one.example/{name}?a=1,b=2"
	second := "https://two.example/{name}"
	global.File.Set("resource.urls", []string{first, "", second})
	if got, want := GetUrls(), []Url{Url(first), Url(second), DefaultUrl}; !reflect.DeepEqual(got, want) {
		t.Fatalf("GetUrls() = %q, want %q", got, want)
	}
	global.File.Set("resource.urls", []string{})
	if got := GetUrls(); !reflect.DeepEqual(got, []Url{DefaultUrl}) {
		t.Fatalf("empty list must retain default source: %q", got)
	}
}
