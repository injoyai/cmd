package global

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/injoyai/conv"
	"gopkg.in/yaml.v3"
)

// 旧缓存组件使用 global@MD5("cmd") 命名。
const legacyConfigName = "global@dfff0a7fa1a55c8c1a4966c19f6da452"

type Config struct {
	*conv.Map
	filename string
	err      error
}

// Open 读取配置；文件不存在时使用空配置，不创建目录或文件。
func Open(filename string) *Config {
	return loadFile(filename, "")
}

func loadFile(filename, legacy string) *Config {
	f := &Config{Map: conv.NewMap(nil), filename: filename}
	source := filename
	data, err := os.ReadFile(source)
	if os.IsNotExist(err) && legacy != "" {
		source = legacy
		data, err = os.ReadFile(source)
	}
	if os.IsNotExist(err) {
		data, err = []byte(`{}`), nil
	}
	if err == nil {
		var object map[string]any
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		err = decoder.Decode(&object)
		if err == nil && object == nil {
			err = fmt.Errorf("配置必须是 YAML 映射")
		}
		if err == nil {
			var extra any
			if next := decoder.Decode(&extra); next != io.EOF {
				err = fmt.Errorf("配置只允许一个 YAML 文档")
			}
		}
		if err == nil {
			err = migrateConfig(object)
		}
		if err == nil {
			f.Map = conv.NewMap(object)
		}
	}
	if err != nil {
		f.err = fmt.Errorf("读取配置 %s: %w", source, err)
		return f
	}
	return f
}

func (f *Config) Filename() string {
	return f.filename
}

func (f *Config) Save() error {
	var data bytes.Buffer
	if err := f.writeTo(&data); err != nil {
		return err
	}
	dir := filepath.Dir(f.filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".config-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if err := tmp.Chmod(0600); err != nil {
		return err
	}
	if _, err := tmp.Write(data.Bytes()); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), f.filename)
}

func (f *Config) writeTo(w io.Writer) error {
	// 读取失败时禁止覆盖，避免把损坏或不可读的配置替换为空配置。
	if f.err != nil {
		return f.err
	}
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	if err := encoder.Encode(f.Refresh().Val()); err != nil {
		return err
	}
	return encoder.Close()
}
