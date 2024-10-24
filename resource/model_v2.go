package resource

import (
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Handler url(资源地址),dir(下载目录),filename(文件完整名称),proxy(代理地址)
type Handler func(url, dir, filename string, proxy ...string) error

type Info struct {
	Key     []string //索引
	Name    Name     //不同系统的名称,例windows是 xxx.exe
	Url     []Url    //源
	Handler Handler  //自定义处理,例如压缩文件
}

func (this *Info) GetName() string {
	return this.Name.GetName()
}

func (this *Info) GetUrl() []string {
	ls := []string(nil)
	for _, v := range this.Url {
		ls = append(ls, v.Format(this.Name.GetResourceName()))
	}
	return ls
}

func (this *Info) InsertUrl(u ...Url) {
	this.Url = append(u, this.Url...)
}

func (this *Info) do(dir string, f Handler) (err error) {
	name := this.Name.GetResourceName()
	for _, v := range this.Url {
		if this.Handler != nil {
			err = this.Handler(v.Format(name), dir, name)
			if err == nil {
				return
			}
		}
		err = f(v.Format(name), dir, name)
		if err == nil {
			return
		}
	}
	return
}

/*
Name
下载名称
资源名称
*/
type Name struct {
	Name         string //实际名称,可能下载名字和资源名字不一致
	All          string //支持全部平台,
	LinuxAmd64   string //支持linux amd64
	LinuxArm     string //支持linux arm
	WindowsAmd64 string //支持windows amd64
	//LinuxArm64   string
	//WindowsArm64 string
}

func (this Name) GetName() string {
	if this.Name != "" {
		return this.Name
	}
	return this.GetResourceName()
}

// GetResourceName 获取资源名称,指的下载地址里面的名称
func (this Name) GetResourceName() string {
	switch runtime.GOOS {
	case "windows":
		return this.WindowsAmd64
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return this.LinuxAmd64
		case "arm":
			return this.LinuxArm
		}
	}
	return this.All
}

// Url 例如minio https://oss.xxx.com/store/{name}
type Url string

func (this Url) Format(name string) string {
	return strings.ReplaceAll(string(this), "{name}", name)
}

var Resources = map[string]*Info{
	"hfs":        {Name: Name{WindowsAmd64: "hfs.exe"}},
	"swag":       {Name: Name{WindowsAmd64: "swag.exe"}},
	"win_active": {Name: Name{WindowsAmd64: "win_active.exe"}},
	"rsrc":       {Name: Name{WindowsAmd64: "rsrc.exe"}},
	"nac":        {Name: Name{WindowsAmd64: "nac.syso"}},
	"upx":        {Name: Name{WindowsAmd64: "upx.exe"}},
	"npc":        {Name: Name{WindowsAmd64: "npc.exe"}},
	"ffmpeg":     {Name: Name{WindowsAmd64: "ffmpeg.exe"}},
	"ffplay":     {Name: Name{WindowsAmd64: "ffplay.exe"}},
	"ffprobe":    {Name: Name{WindowsAmd64: "ffprobe.exe"}},
	"livego":     {Name: Name{WindowsAmd64: "livego.exe"}},
	"motrix":     {Name: Name{WindowsAmd64: "motrix.exe"}},
	"frpc":       {Name: Name{WindowsAmd64: "frpc.exe"}},
	"frps":       {Name: Name{WindowsAmd64: "frps.exe"}},
	"ModbusPoll": {Name: Name{WindowsAmd64: "ModbusPoll.exe"}, Key: []string{"modbuspoll"}},

	"proxy":         {Name: Name{WindowsAmd64: "proxy.exe"}},
	"listen":        {Name: Name{WindowsAmd64: "listen.exe"}},
	"timer":         {Name: Name{WindowsAmd64: "timer.exe"}},
	"edge":          {Name: Name{WindowsAmd64: "edge.exe"}},
	"edge_mini":     {Name: Name{WindowsAmd64: "edge_mini.exe"}},
	"notice_client": {Name: Name{WindowsAmd64: "notice_client.exe"}, Key: []string{"notice_cli", "notice-cli"}},
	"upgrade":       {Name: Name{WindowsAmd64: "in_upgrade.exe"}, Key: []string{"in_upgrade"}},
	"server":        {Name: Name{WindowsAmd64: "in_server.exe"}, Key: []string{"in_server"}},
	"in":            {Name: Name{WindowsAmd64: "in.exe", LinuxAmd64: "in", LinuxArm: "in7"}},

	"build.sh":           {Name: Name{All: "build.sh"}, Key: []string{"build"}},
	"service.service":    {Name: Name{All: "service.service"}, Key: []string{"service"}},
	"dockerfile":         {Name: Name{All: "dockerfile"}, Key: []string{"Dockerfile"}},
	"install_minio.sh":   {Name: Name{All: "install_minio.sh"}, Key: []string{"install_minio"}},
	"install_nodered.sh": {Name: Name{All: "install_nodered.sh"}, Key: []string{"install_nodered"}},
	"install_v2raya.sh":  {Name: Name{All: "install_v2raya.sh"}, Key: []string{"install_v2raya"}},

	"downloader": {
		Key:  []string{"download"},
		Name: Name{WindowsAmd64: "downloader.exe"},
		Url:  []Url{"https://github.com/injoyai/downloader/releases/latest/download/downloader.exe"},
	},

	"chrome104": {
		Name: Name{WindowsAmd64: "chrome.zip"},
		Url:  []Url{"https://github.com/injoyai/resource/releases/download/v0.0.0/chrome.zip"},
		Handler: func(url, dir, filename string, proxy ...string) error {
			zipFilename := filepath.Join(dir, "chrome.zip")
			if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			return zip.Decode(zipFilename, dir)
		},
	},

	"influxdb": {
		Key:  []string{"influx", "influxd"},
		Name: Name{Name: "influxd.exe"},
		Handler: func(url, dir, filename string, proxy ...string) error {
			url = "https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip"
			zipFilename := filepath.Join(dir, "influxdb.zip")
			if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
				return err
			}
			if err := zip.Decode(zipFilename, dir); err != nil {
				return err
			}
			logs.PrintErr(os.Remove(zipFilename))

			folder := "/influxdb-1.8.10-1"
			logs.PrintErr(os.Rename(filepath.Join(dir, folder, "/influxd.exe"), filename))
			logs.PrintErr(os.RemoveAll(filepath.Join(dir, folder)))
			return nil
		},
	},

	"ps5": {
		Name: Name{WindowsAmd64: "PhotoShop CS5.zip"},
		Handler: func(url, dir, filename string, proxy ...string) error {
			zipFilename := filepath.Join(dir, "ps5.zip")
			if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
				return err
			}
			if err := zip.Decode(zipFilename, filepath.Join(dir, "PhotoShop CS5/")); err != nil {
				return err
			}
			logs.PrintErr(os.Remove(zipFilename))
			return nil
		},
	},

	"ipinfo": {
		Name: Name{WindowsAmd64: "ipinfo.exe"},
		Handler: func(url, dir, filename string, proxy ...string) error {
			url = "https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_windows_amd64.zip"
			zipFilename := filepath.Join(dir, "ipinfo.zip")
			if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
				return err
			}
			if err := zip.Decode(zipFilename, dir); err != nil {
				return err
			}
			logs.PrintErr(os.Remove(zipFilename))
			logs.PrintErr(os.Rename(filepath.Join(dir, "/ipinfo_3.3.1_windows_amd64.exe"), filename))
			return nil
		},
	},
}

type Infos map[string]*Info

//func (this Infos) FindAndDownload(key string, cfg *Config) (inResource bool, filename string, inLocal bool) {
//	_, inResource = Resources[key]
//	if inResource {
//		filename, inLocal = MustDownload(g.Ctx(), cfg)
//	}
//	return
//}

func init() {

	lsConfig := []Url(nil)
	for _, v := range strings.Split(global.GetString("resource"), ",") {
		if len(v) != 0 {
			lsConfig = append(lsConfig, Url(v))
		}
	}

	lsBase := []Url{
		"https://oss.injoy.ink/in-store/{name}",
		"https://github.com/injoyai/cmd/raw/main/resource/{name}",
	}

	for _, v := range Resources {
		v.InsertUrl(lsBase...)
		v.InsertUrl(lsConfig...)
		for _, k := range v.Key {
			Resources[k] = v
		}
	}
}
