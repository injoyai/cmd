package main

import (
	"io"
	"strings"
	"testing"
)

func TestDownloadRejectsToolSelection(t *testing.T) {
	isolatedConfig(t)
	for _, name := range []string{"dl", "download"} {
		root := newRootCommand()
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs([]string{name, "agents", "--downloader=aria2"})
		if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "unknown flag") {
			t.Fatalf("%s still accepts a tool selector: %v", name, err)
		}
	}
}
