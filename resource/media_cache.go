package resource

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// prepareCache 只复用带有本资源标记的缓存，锁文件避免同时写同一组分片。
func prepareCache(op *Task, identity string, resume bool) (func(), error) {
	id := fmt.Sprintf("%x", sha256.Sum256([]byte(op.URL()+"\x00"+op.Filename()+"\x00"+identity)))
	parent := filepath.Dir(op.Filename())
	if err := os.MkdirAll(parent, 0755); err != nil {
		return nil, err
	}
	prefix := ".i-download-" + id[:16] + "-"
	claim := func(dir string) (func(), error) {
		release, err := lockCache(filepath.Join(dir, ".lock"))
		if err != nil {
			return nil, err
		}
		op.CacheDir = dir
		return release, nil
	}
	if resume {
		entries, err := os.ReadDir(parent)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
				continue
			}
			dir := filepath.Join(parent, entry.Name())
			marker, err := os.ReadFile(filepath.Join(dir, ".source"))
			if err == nil && string(marker) == id {
				if release, err := claim(dir); err == nil {
					return release, nil
				}
			}
		}
	}
	dir, err := os.MkdirTemp(parent, prefix)
	if err != nil {
		return nil, err
	}
	release, err := claim(dir)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, ".source"), []byte(id), 0600); err != nil {
		release()
		os.RemoveAll(dir)
		return nil, err
	}
	return release, nil
}
