package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionHistory(t *testing.T) {
	originalVersions, originalBuildDate := versions, BuildDate
	t.Cleanup(func() {
		versions, BuildDate = originalVersions, originalBuildDate
	})
	versions = make([]VersionInfo, 12)
	for i := range versions {
		versions[i] = VersionInfo{Version: "v1.0.0", Details: "历史记录"}
	}
	versions[0] = VersionInfo{Version: "v1.1.0", Date: "2026-10-29", Details: "更新详情"}
	BuildDate = "2026-10-29/15:00"

	for _, tt := range []struct {
		name string
		args []string
		want int
	}{
		{name: "recent", want: 10},
		{name: "all after recent", args: []string{"all"}, want: 12},
		{name: "recent after all", want: 10},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			root := newRootCommand()
			root.SetOut(&out)
			root.SetArgs(append([]string{"version"}, tt.args...))
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			got := out.String()
			if count := strings.Count(got, ": "); count != tt.want+1 {
				t.Fatalf("record count = %d, want %d; output: %s", count-1, tt.want, got)
			}
			for _, want := range []string{
				"v1.1.0 (2026-10-29): 更新详情\n",
				"v1.0.0: 历史记录\n",
				"编译日期: " + BuildDate,
			} {
				if !strings.Contains(got, want) {
					t.Errorf("missing %q in output: %s", want, got)
				}
			}
			if len(versions) != 12 {
				t.Fatalf("version history was truncated to %d records", len(versions))
			}
		})
	}
}
