package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/injoyai/cmd/global"
	"gopkg.in/yaml.v3"
)

func isolatedGlobal(t *testing.T) *global.Config {
	t.Helper()
	previous := global.File
	global.File = global.Open(filepath.Join(t.TempDir(), "config"))
	t.Cleanup(func() { global.File = previous })
	return global.File
}

func runGlobal(t *testing.T, args ...string) (map[string]any, error) {
	t.Helper()
	cmd := newRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(append([]string{"global"}, args...))
	if err := cmd.Execute(); err != nil {
		return nil, err
	}
	var values map[string]any
	if err := yaml.Unmarshal(out.Bytes(), &values); err != nil {
		t.Fatalf("invalid config output %q: %v", out.String(), err)
	}
	return values, nil
}

func TestGlobalQueryDoesNotWrite(t *testing.T) {
	f := isolatedGlobal(t)
	if _, err := runGlobal(t); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(f.Filename()); !os.IsNotExist(err) {
		t.Fatalf("query created a file: %v", err)
	}
	f.Set("custom", "visible")
	if err := global.Save(); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Unix(1_000_000, 0)
	if err := os.Chtimes(f.Filename(), oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	values, err := runGlobal(t)
	if err != nil || values["custom"] != "visible" {
		t.Fatalf("query: values=%v error=%v", values, err)
	}
	info, err := os.Stat(f.Filename())
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(oldTime) {
		t.Fatal("query rewrote existing config")
	}
}

func TestGlobalRepeatedFlagsPreserveValues(t *testing.T) {
	f := isolatedGlobal(t)
	resource := "https://one.example/{name},https://two.example/{name}"
	values, err := runGlobal(t,
		"-s", "resource="+resource, "--set", "testUploadUrl=a=b=c", "-s", "empty=",
		"-s", "uploadMinio.endpoint=localhost:9000",
		"-a", "items=first,second", "--append", "items=a=b")
	if err != nil {
		t.Fatal(err)
	}
	if values["resource"] != resource || values["testUploadUrl"] != "a=b=c" || values["empty"] != "" {
		t.Fatalf("values were lost: %v", values)
	}
	if !reflect.DeepEqual(values["items"], []any{"first,second", "a=b"}) {
		t.Fatalf("items=%v", values["items"])
	}
	reloaded := global.Open(f.Filename())
	if reloaded.GetString("uploadMinio.endpoint") != "localhost:9000" || reloaded.GetString("resource") != resource {
		t.Fatal("saved values differ after reloading")
	}
}

func TestGlobalInvalidArgsKeepConfig(t *testing.T) {
	for _, args := range [][]string{
		{"-s", "=clobber"}, {"-s", "..=clobber"}, {"-s", " =clobber"},
		{"-s", "broken"}, {"-s", ""}, {"-d", ""}, {"-d", "proxy,"},
		{"-s", "proxy=changed", "-a", "broken"},
		{"-s", "proxy=changed", "-d", ".."}, {"unknown"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			f := isolatedGlobal(t)
			f.Set("proxy", "original")
			if err := global.Save(); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(f.Filename())
			if err != nil {
				t.Fatal(err)
			}
			if _, err := runGlobal(t, args...); err == nil {
				t.Fatalf("accepted invalid args %q", args)
			}
			after, err := os.ReadFile(f.Filename())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) || f.GetString("proxy") != "original" {
				t.Fatal("invalid input changed config")
			}
		})
	}
}

func TestGlobalOperationOrder(t *testing.T) {
	isolatedGlobal(t)
	for i := 0; i < 20; i++ {
		values, err := runGlobal(t, "-d", "proxy,items", "-a", "items=extra", "-s", "proxy=x", "-s", "items=initial")
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := values["proxy"]; ok {
			t.Fatal("set ran after delete")
		}
		if _, ok := values["items"]; ok {
			t.Fatal("append ran after delete")
		}
	}
	values, err := runGlobal(t, "-s", "proxy=x", "-s", "temporary=y", "-d", "proxy", "-d", "temporary")
	if err != nil || len(values) != 0 {
		t.Fatalf("repeated delete: values=%v error=%v", values, err)
	}
}

func TestGlobalSaveFailure(t *testing.T) {
	f := isolatedGlobal(t)
	if err := os.WriteFile(f.Filename(), []byte("block directory creation"), 0600); err != nil {
		t.Fatal(err)
	}
	global.File = global.Open(filepath.Join(f.Filename(), "child"))
	if _, err := runGlobal(t, "-s", "proxy=x"); err == nil {
		t.Fatal("save failure was ignored")
	}
}

func TestGlobalUserConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("APPDATA", filepath.Join(dir, "appdata"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	base, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(base, "injoy", "i", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		t.Fatal(err)
	}
	jsonFilename := filepath.Join(filepath.Dir(filename), "config.json")
	jsonData := []byte(`{"custom":"ignored"}`)
	if err := os.WriteFile(jsonFilename, jsonData, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := cliProcess("global").CombinedOutput(); err != nil {
		t.Fatalf("query: %v %s", err, out)
	} else if strings.Contains(string(out), "ignored") {
		t.Fatal("query loaded config.json")
	}
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Fatalf("query created default config: %v", err)
	}
	if out, err := cliProcess("global", "-s", "custom=preserved").CombinedOutput(); err != nil {
		t.Fatalf("save: %v %s", err, out)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]any
	if err := yaml.Unmarshal(data, &values); err != nil || values["custom"] != "preserved" {
		t.Fatalf("default config: error=%v values=%v", err, values)
	}
	if data, err := os.ReadFile(jsonFilename); err != nil || !bytes.Equal(data, jsonData) {
		t.Fatalf("config.json was modified: %v", err)
	}
}
