package resource

import (
	"runtime"
	"strings"

	"github.com/injoyai/cmd/global"
)

const DefaultUrl = "https://oss.002246.xyz/in-store/{name}"

// GetUrls 从配置中读取配置的基础地址,和默认地址,例 https://oss.xxx.com/store/{name}
func GetUrls() []Url {
	urls := []Url(nil)
	for _, u := range global.File.GetStrings("resource.urls") {
		if len(u) != 0 {
			urls = append(urls, Url(u))
		}
	}
	urls = append(urls, DefaultUrl)
	return urls
}

// Url 例如minio https://oss.xxx.com/store/{name}_{os}_{arch}
type Url string

func (this Url) Format(name string) string {
	s := strings.ReplaceAll(string(this), "{name}", name)
	s = strings.ReplaceAll(s, "{os}", runtime.GOOS)
	s = strings.ReplaceAll(s, "{arch}", runtime.GOARCH)
	return s
}
