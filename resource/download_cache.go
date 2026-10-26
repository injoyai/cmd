package resource

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// prepareMediaCache 只复用带有本资源标记的缓存，锁文件避免同时写同一组分片。
func (op *Task) prepareMediaCache(identity string, resume bool) (func(), error) {
	id := fmt.Sprintf("%x", sha256.Sum256([]byte(op.URL()+"\x00"+op.Filename()+"\x00"+identity)))
	parent := filepath.Dir(op.Filename())
	if err := os.MkdirAll(parent, 0755); err != nil {
		return nil, err
	}
	prefix := ".i-download-" + id[:16] + "-"
	claim := func(dir string) (func(), error) {
		lock := filepath.Join(dir, ".lock")
		f, err := os.OpenFile(lock, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return nil, err
		}
		if err := f.Close(); err != nil {
			os.Remove(lock)
			return nil, err
		}
		op.cacheDir = dir
		return func() { os.Remove(lock) }, nil
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
		os.RemoveAll(dir)
		return nil, err
	}
	return release, nil
}
