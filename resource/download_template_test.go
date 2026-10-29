package resource

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDownloadTemplatePriorityAndCover(t *testing.T) {
	previous, existed := Resources["agents"]
	Resources["agents"] = &Info{
		Local:   "wrong.txt",
		FullUrl: []Url{"http://unused/template"},
		Handler: func(op *Task) error {
			t.Error("embedded template must take priority over built-in resources")
			return os.WriteFile(op.Filename(), []byte("wrong source"), 0644)
		},
	}
	t.Cleanup(func() {
		if existed {
			Resources["agents"] = previous
		} else {
			delete(Resources, "agents")
		}
	})
	want, err := Templates.ReadFile("templates/AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	op := &Config{Resource: "agents", Dir: t.TempDir()}
	filename, exist, err := Download(context.Background(), op)
	if err != nil || exist || filename != filepath.Join(op.Dir, "AGENTS.md") {
		t.Fatalf("Download() = %q, %v, %v", filename, exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != string(want) {
		t.Fatalf("template content mismatch: %v", err)
	}
	if err := os.WriteFile(filename, []byte("user content"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, exist, err := Download(context.Background(), op); err != nil || !exist {
		t.Fatalf("existing template: exist = %v, err = %v", exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "user content" {
		t.Fatalf("existing template was overwritten: %v", err)
	}
	op.Cover = true
	if _, exist, err := Download(context.Background(), op); err != nil || exist {
		t.Fatalf("cover template: exist = %v, err = %v", exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != string(want) {
		t.Fatalf("template was not restored with Cover: %v", err)
	}
}

func TestDownloadCustomTemplateExtension(t *testing.T) {
	op := &Config{Resource: "agents", Name: "custom", Dir: t.TempDir()}
	before := *op
	filename, _, err := Download(context.Background(), op)
	if err != nil || filename != filepath.Join(op.Dir, "custom.md") {
		t.Fatalf("custom template: filename = %q, err = %v", filename, err)
	}
	if !reflect.DeepEqual(*op, before) {
		t.Fatal("template download changed caller configuration")
	}
}
