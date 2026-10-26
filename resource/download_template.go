package resource

import (
	"context"
	"embed"
	"io/fs"
	"strings"

	"github.com/injoyai/goutil/oss"
)

//go:embed templates templates/.gitignore
var Templates embed.FS

func getTemplateName(name string) string {
	switch strings.ToLower(name) {
	case "agents", "agents.md":
		return "AGENTS.md"
	case "dockerfile":
		return "Dockerfile"
	case "golang.md":
		return "GOLANG.md"
	case "build", "build.sh":
		return "build.sh"
	case "gitignore", ".gitignore":
		return ".gitignore"
	case "install", "install.sh":
		return "install.sh"
	case "setup_ssh_key", "setup_ssh_key.ps1":
		return "setup_ssh_key.ps1"
	case "install_go", "install_go.sh":
		return "install_go.sh"
	case "ca.crt":
		return "ca.crt"
	case "ca.key":
		return "ca.key"
	default:
		return name
	}
}

type templateDownloader struct{}

func (templateDownloader) Match(resource string) bool {
	info, err := fs.Stat(Templates, "templates/"+getTemplateName(resource))
	return err == nil && !info.IsDir()
}

func (templateDownloader) Download(ctx context.Context, op *Config) (string, bool, error) {
	name := getTemplateName(op.Resource)
	op.setFilename(name, "")
	return downloadFile(ctx, op, func(ctx context.Context, op *Config) error {
		bs, err := Templates.ReadFile("templates/" + name)
		if err != nil {
			return err
		}
		return oss.New(op.Filename(), string(bs))
	})
}
