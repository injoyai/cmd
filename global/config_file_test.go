package global

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLegacyConfigMigration(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, legacyConfigName)
	filename := filepath.Join(dir, "injoy", "i", "config.yaml")
	original := []byte(`{"proxy":"old","custom":{"value":"keep"}}`)
	if err := os.WriteFile(legacy, original, 0644); err != nil {
		t.Fatal(err)
	}
	f := loadFile(filename, legacy)
	if f.err != nil || f.GetString("proxy") != "old" || f.GetString("custom.value") != "keep" {
		t.Fatalf("legacy load: error=%v", f.err)
	}
	if _, err := os.Stat(filepath.Dir(filename)); !os.IsNotExist(err) {
		t.Fatalf("load created config directory: %v", err)
	}
	f.Set("proxy", "new")
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(legacy)
	if err != nil || !bytes.Equal(old, original) {
		t.Fatalf("legacy file changed: %v", err)
	}
	f = loadFile(filename, legacy)
	if f.err != nil || f.GetString("proxy") != "new" || f.GetString("custom.value") != "keep" {
		t.Fatalf("new config was not preferred: error=%v", f.err)
	}
	if runtime.GOOS != "windows" {
		for path, mode := range map[string]os.FileMode{filename: 0600, filepath.Dir(filename): 0700} {
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != mode {
				t.Errorf("%s mode=%o want=%o", path, info.Mode().Perm(), mode)
			}
		}
	}
}

func TestInvalidConfigIsNotOverwritten(t *testing.T) {
	for _, original := range []string{
		`{"broken":`, `[]`, `null`, `"text"`, ``,
		"proxy: [", "proxy: first\nproxy: second\n", "proxy: first\n---\nproxy: second\n",
	} {
		t.Run(original, func(t *testing.T) {
			dir := t.TempDir()
			filename := filepath.Join(dir, "config.yaml")
			legacy := filepath.Join(dir, "legacy")
			if err := os.WriteFile(filename, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(legacy, []byte(`{"proxy":"legacy"}`), 0600); err != nil {
				t.Fatal(err)
			}
			f := loadFile(filename, legacy)
			if f.err == nil {
				t.Fatal("invalid config accepted")
			}
			f.Set("proxy", "replacement")
			if err := f.Save(); err == nil {
				t.Fatal("save succeeded after failed read")
			}
			data, err := os.ReadFile(filename)
			if err != nil || string(data) != original {
				t.Fatalf("invalid config was replaced: %v", err)
			}
		})
	}
}

func TestYAMLConfigRoundTrip(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.yaml")
	original := `# 手动配置
proxy: http://localhost:1081
enabled: false
literalBool: "true"
numericText: "00123"
empty: ""
id: 9007199254740993
uploadMinio:
  endpoint: localhost:9000
  rename: false
customOpen:
  .md: TextEdit
items:
  - one
  - a,b=c
message: |
  第一行
  第二行
`
	if err := os.WriteFile(filename, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	f := Open(filename)
	if f.err != nil || f.GetString("uploadMinio.endpoint") != "localhost:9000" || f.GetBool("enabled") {
		t.Fatalf("YAML load: error=%v", f.err)
	}
	f.Set("proxy", "http://localhost:1082")
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "uploadMinio:\n  endpoint:") {
		t.Fatalf("expected readable indented YAML, got %s", data)
	}
	var before, after map[string]any
	if err := yaml.Unmarshal([]byte(original), &before); err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &after); err != nil {
		t.Fatal(err)
	}
	before["proxy"] = "http://localhost:1082"
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("values or types changed: before=%v after=%v", before, after)
	}
}

func TestConfigReadErrorDoesNotFallBack(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "legacy")
	if err := os.WriteFile(legacy, []byte(`{"proxy":"legacy"}`), 0600); err != nil {
		t.Fatal(err)
	}
	f := loadFile(dir, legacy)
	if f.err == nil || f.GetString("proxy") != "" {
		t.Fatal("read error silently fell back to legacy config")
	}
}

func TestSaveReplacesExistingConfigPrivately(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.yaml")
	original := []byte(`{"proxy":"old"}`)
	if err := os.WriteFile(filename, original, 0644); err != nil {
		t.Fatal(err)
	}
	f := Open(filename)
	f.Set("proxy", "new")
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	if got := Open(filename).GetString("proxy"); got != "new" {
		t.Fatalf("saved proxy=%q", got)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("existing config retained broad mode %o", info.Mode().Perm())
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(filename), ".config-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary files remain: %v error=%v", matches, err)
	}
}
