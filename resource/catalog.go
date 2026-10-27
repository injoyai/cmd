package resource

import (
	"runtime"
)

// Handler 处理已解析的资源下载，例如下载发布包后解压安装。
type Handler func(op *Task) error

type Info struct {
	Key         []string //索引
	Local       string   //本地资源名称
	Remote      string   //默认远程资源名称
	RemoteAMD64 string   //远程资源名称,amd64架构
	Remote386   string   //远程资源名称,386架构
	RemoteArm   string   //远程资源名称,arm架构
	RemoteArm64 string   //远程资源名称,arm64架构
	FullUrl     []Url    //完整的资源地址,todo 缓存和最新目前在一起
	Handler     Handler  //自定义处理,例如压缩文件
}

func (this *Info) GetLocalName() string {
	return this.Local
}

func (this *Info) GetFullUrls() []string {
	ls := []string(nil)
	for _, v := range this.FullUrl {
		ls = append(ls, v.Format(this.GetRemote()))
	}
	return ls
}

func (this *Info) GetRemote() string {
	switch runtime.GOARCH {
	case "amd64":
		if len(this.RemoteAMD64) > 0 {
			return this.RemoteAMD64
		}
	case "386":
		if len(this.Remote386) > 0 {
			return this.Remote386
		}
	case "arm":
		//arm专属资源为空则回退默认资源
		if len(this.RemoteArm) > 0 {
			return this.RemoteArm
		}
	case "arm64":
		//arm64专属资源为空则回退默认资源
		if len(this.RemoteArm64) > 0 {
			return this.RemoteArm64
		}
	}
	return this.Remote
}

// init 初始化,补充各个架构的值
func (this *Info) init() {
	if len(this.Local) == 0 && len(this.GetRemote()) > 0 {
		this.Local = this.GetRemote()
	} else if len(this.GetRemote()) == 0 && len(this.Local) > 0 {
		this.Remote = this.Local
		this.RemoteArm = this.Local
	}
}

var Resources = map[string]*Info{
	"build_win.sh":       {Local: "build_win.sh", Key: []string{"build_win"}},
	"service.service":    {Local: "service.service", Key: []string{"service"}},
	"install_minio.sh":   {Local: "install_minio.sh", Key: []string{"install_minio"}},
	"install_nodered.sh": {Local: "install_nodered.sh", Key: []string{"install_nodered"}},
	"install_v2raya.sh":  {Local: "install_v2raya.sh", Key: []string{"install_v2raya"}},
	"golang.md":          {Local: "GOLANG.md", Key: []string{"go.md", "GOLANG.md"}},
}

func init() {

	//从配置中读取配置的基础地址,例 https://oss.xxx.com/store/{name}
	urls := GetUrls()

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
		if len(v.FullUrl) == 0 {
			v.FullUrl = append(v.FullUrl, urls...)
		}
	}

}
