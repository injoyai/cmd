package global

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGroupedConfigMigration(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(map[bool]string{false: "yaml", true: "legacy"}[legacy], func(t *testing.T) {
			dir := t.TempDir()
			filename := filepath.Join(dir, "config.yaml")
			source := filename
			if legacy {
				source = filepath.Join(dir, legacyConfigName)
			}
			original := []byte(`proxy: http://old.example:1081
proxyIgnore: localhost,127.0.0.1
downloadDir: ./old
downloadNoticeEnable: false
downloadNoticeText: ""
resource: https://one.example/{name},https://two.example/{name}
uploadMinio:
  endpoint: localhost:9000
  access: example-access
  secret: example-secret
  bucket: old
  rename: false
  custom: keep
customOpen:
  .md: TextEdit
testDownloadUrl: https://test.example/download
testUploadUrl: https://test.example/upload
download:
  dir: ./new
  custom: 9007199254740993
upload:
  minio:
    bucket: new
open:
  custom:
    .md: OtherEditor
    key.with.dots: keep
custom:
  values: [true, false, null, "00123"]
`)
			if err := os.WriteFile(source, original, 0600); err != nil {
				t.Fatal(err)
			}
			f := loadFile(filename, source)
			var out bytes.Buffer
			if err := f.writeTo(&out); err != nil {
				t.Fatal(err)
			}
			if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, original) {
				t.Fatalf("query rewrote source: %v", err)
			}
			if legacy {
				if _, err := os.Stat(filename); !os.IsNotExist(err) {
					t.Fatalf("query created YAML: %v", err)
				}
			}
			// 修改一项后，其他旧值、新值和自定义字段必须完整保留。
			f.Set("download.dir", "./changed")
			if err := f.Save(); err != nil {
				t.Fatal(err)
			}
			want := `proxy:
  addr: http://old.example:1081
  ignore: localhost,127.0.0.1
download:
  dir: ./changed
  noticeEnable: false
  noticeText: ""
  custom: 9007199254740993
resource:
  urls:
    - https://one.example/{name}
    - https://two.example/{name}
upload:
  minio:
    endpoint: localhost:9000
    access: example-access
    secret: example-secret
    bucket: new
    rename: false
    custom: keep
open:
  custom:
    .md: OtherEditor
    key.with.dots: keep
test:
  downloadUrl: https://test.example/download
  uploadUrl: https://test.example/upload
custom:
  values: [true, false, null, "00123"]
`
			data, err := os.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			var expected, actual map[string]any
			if err := yaml.Unmarshal([]byte(want), &expected); err != nil {
				t.Fatal(err)
			}
			if err := yaml.Unmarshal(data, &actual); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("migration changed values or types:\n%s", data)
			}
			reloaded := Open(filename)
			if reloaded.GetString("download.dir") != "./changed" || reloaded.GetBool("download.noticeEnable", true) || !reflect.DeepEqual(reloaded.GetStrings("resource.urls"), []string{"https://one.example/{name}", "https://two.example/{name}"}) {
				t.Fatal("grouped config differs after reloading")
			}
			if legacy {
				if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, original) {
					t.Fatalf("save changed legacy file: %v", err)
				}
			}
		})
	}
}

func TestMigrationConflictDoesNotOverwrite(t *testing.T) {
	for _, original := range []string{
		"download: custom-value\ndownloadDir: ./old\n",
		"upload: custom-value\nuploadMinio: {bucket: old}\n",
		"upload: {minio: custom-value}\nuploadMinio: {bucket: old}\n",
		"open: {custom: custom-value}\ncustomOpen: {.md: TextEdit}\n",
		"test: custom-value\ntestDownloadUrl: https://example.com\n",
	} {
		t.Run(original, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(filename, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			f := Open(filename)
			if f.err == nil {
				t.Fatal("migration accepted a scalar in a destination group")
			}
			f.Set("unrelated", "change")
			if err := f.Save(); err == nil {
				t.Fatal("saved after migration failed")
			}
			if data, err := os.ReadFile(filename); err != nil || string(data) != original {
				t.Fatalf("conflicting configuration was overwritten: %v", err)
			}
		})
	}
}

func TestNewGroupedValuesWin(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(filename, []byte(`proxy: {addr: "", ignore: ""}
proxyIgnore: old-rule
downloadDir: ./old
downloadNoticeEnable: true
downloadNoticeText: old-text
download: {dir: "", noticeEnable: false, noticeText: ""}
resource:
  urls: ["https://example.com/{name}?a=1,b=2"]
testDownloadUrl: old-url
test: {downloadUrl: ""}
`), 0600); err != nil {
		t.Fatal(err)
	}
	f := Open(filename)
	if f.err != nil {
		t.Fatal(f.err)
	}
	for _, key := range []string{"proxy.addr", "proxy.ignore", "download.dir", "download.noticeText", "test.downloadUrl"} {
		if got := f.GetString(key, "missing"); got != "" {
			t.Fatalf("new empty value at %s was replaced: %q", key, got)
		}
	}
	if f.GetBool("download.noticeEnable", true) {
		t.Fatal("new false value was replaced")
	}
	if got := f.GetStrings("resource.urls"); !reflect.DeepEqual(got, []string{"https://example.com/{name}?a=1,b=2"}) {
		t.Fatalf("new URL list was split: %q", got)
	}
}
