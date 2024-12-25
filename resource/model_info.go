package resource

import (
	"github.com/injoyai/cmd/global"
	"strings"
)

type Info struct {
	Key     []string //索引
	Local   string   //本地资源名称
	Remote  string   //远程资源名称
	FullUrl []Url    //完整的资源地址,todo 缓存和最新目前在一起
	Handler Handler  //自定义处理,例如压缩文件
}

func (this *Info) GetLocalName() string {
	return this.Local
}

func (this *Info) GetFullUrls() []string {
	ls := []string(nil)
	for _, v := range this.FullUrl {
		ls = append(ls, v.Format(this.Remote))
	}
	return ls
}

func (this *Info) GetHandler() Handler {
	return this.Handler
}

func (this *Info) init() {
	if len(this.Local) == 0 && len(this.Remote) >= 0 {
		this.Local = this.Remote
	} else if len(this.Remote) == 0 && len(this.Local) >= 0 {
		this.Remote = this.Local
	}
}

type MResource map[string]*Info

func (this MResource) Get(key string) (Resource, bool) {
	r, ok := this[key]
	return r, ok
}

var Resources = MResource{
	"build.sh":           {Local: "build.sh"},
	"build_win.sh":       {Local: "build_win.sh", Key: []string{"build_win"}},
	"service.service":    {Local: "service.service", Key: []string{"service"}},
	"dockerfile":         {Local: "Dockerfile", Key: []string{"Dockerfile"}},
	"install_minio.sh":   {Local: "install_minio.sh", Key: []string{"install_minio"}},
	"install_nodered.sh": {Local: "install_nodered.sh", Key: []string{"install_nodered"}},
	"install_v2raya.sh":  {Local: "install_v2raya.sh", Key: []string{"install_v2raya"}},
}

func init() {

	//从配置中读取配置的基础地址,例 https://oss.xxx.com/store/{name}
	lsConfig := []Url(nil)
	for _, v := range strings.Split(global.GetString("resource"), ",") {
		if len(v) != 0 {
			lsConfig = append(lsConfig, Url(v))
		}
	}

	//建立索引
	for k, v := range Resources {
		Resources[k] = v
		Resources[v.Local] = v
		for _, k2 := range v.Key {
			Resources[k2] = v
		}
	}

	//合并资源
	for k, v := range Exclusive {
		Resources[k] = v
		Resources[v.Local] = v
		for _, k2 := range v.Key {
			Resources[k2] = v
		}
	}

	//补充配置的地址
	for _, v := range Resources {
		v.init()
		v.FullUrl = append(v.FullUrl, lsConfig...)
	}

}
