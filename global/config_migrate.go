package global

import (
	"fmt"
	"strings"
)

// migrateConfig 只整理内存；首次修改后才把完整配置写回原 YAML 路径。
func migrateConfig(values map[string]any) error {
	for _, pair := range [][2]string{{"proxy", "addr"}, {"resource", "urls"}} {
		value, exists := values[pair[0]]
		if _, grouped := value.(map[string]any); !exists || grouped {
			continue
		}
		if text, ok := value.(string); ok && pair[0] == "resource" {
			urls := []string{}
			for _, url := range strings.Split(text, ",") {
				if url != "" {
					urls = append(urls, url)
				}
			}
			value = urls
		}
		values[pair[0]] = map[string]any{pair[1]: value}
	}
	for _, pair := range [][2]string{
		{"proxyIgnore", "proxy.ignore"},
		{"downloadDir", "download.dir"},
		{"downloadNoticeEnable", "download.noticeEnable"},
		{"downloadNoticeText", "download.noticeText"},
		{"uploadMinio", "upload.minio"},
		{"customOpen", "open.custom"},
	} {
		if value, exists := values[pair[0]]; exists {
			if err := mergeConfigValue(values, strings.Split(pair[1], "."), value); err != nil {
				return fmt.Errorf("迁移 %s 到 %s: %w", pair[0], pair[1], err)
			}
			delete(values, pair[0])
		}
	}
	return nil
}

// 合并时已有新值优先；逐项搬移对象，保留未登记的自定义字段和含点的字面键。
func mergeConfigValue(values map[string]any, path []string, value any) error {
	key := path[0]
	current, exists := values[key]
	if len(path) > 1 {
		child, ok := current.(map[string]any)
		if !exists {
			child = make(map[string]any)
			values[key] = child
		} else if !ok {
			return fmt.Errorf("%s 已被非对象配置占用", key)
		}
		return mergeConfigValue(child, path[1:], value)
	}
	if !exists {
		values[key] = value
		return nil
	}
	if oldObject, ok := value.(map[string]any); ok {
		newObject, ok := current.(map[string]any)
		if !ok {
			return fmt.Errorf("%s 已被非对象配置占用", key)
		}
		for name, item := range oldObject {
			if err := mergeConfigValue(newObject, []string{name}, item); err != nil {
				return err
			}
		}
	}
	return nil
}
