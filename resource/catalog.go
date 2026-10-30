package resource

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// Handler 处理已解析的资源下载，例如下载发布包后解压安装。
type Handler func(op *Task) error

type Info struct {
	Key     []string          //别名
	Local   string            //本地资源名称
	Remote  map[string]string //架构对应的远程文件名，空键表示通用文件；省略时使用 Local
	FullUrl []Url             //下载地址模板，省略时使用配置的仓库地址
	Handler Handler           //自定义处理，例如解压安装
	Install Handler           //无对应发布包时通过系统包管理器安装
}

func (this *Info) GetLocalName() string {
	return this.Local
}

func (this *Info) GetFullUrls() ([]string, error) {
	name, err := this.GetRemote()
	if err != nil {
		return nil, err
	}
	urls := this.FullUrl
	if len(urls) == 0 {
		urls = GetUrls()
	}
	ls := make([]string, 0, len(urls))
	for _, v := range urls {
		ls = append(ls, v.Format(name))
	}
	return ls, nil
}

func (this *Info) GetRemote() (string, error) {
	return this.remoteForArch(runtime.GOARCH)
}

func (this *Info) remoteForArch(arch string) (string, error) {
	if name := this.Remote[arch]; name != "" {
		return name, nil
	}
	if name := this.Remote[""]; name != "" {
		return name, nil
	}
	if len(this.Remote) == 0 && this.Local != "" && this.Install == nil {
		return this.Local, nil
	}
	if this.Install != nil {
		return "", fmt.Errorf("%s 使用系统包管理器安装，没有 %s/%s 独立发布包", this.Local, runtime.GOOS, arch)
	}
	return "", fmt.Errorf("%s 暂无 %s/%s 资源", this.Local, runtime.GOOS, arch)
}

// installPackage 保留包管理器维护的完整依赖，仅在目标目录创建命令链接。
func installPackage(binary string, command ...string) Handler {
	return func(op *Task) error {
		if err := os.MkdirAll(filepath.Dir(op.Filename()), 0755); err != nil {
			return err
		}
		temp, err := os.MkdirTemp(filepath.Dir(op.Filename()), ".i-install-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(temp)
		filename, err := exec.LookPath(binary)
		if err != nil || op.Cover {
			fmt.Printf("执行安装: %s\n", strings.Join(command, " "))
			cmd := exec.CommandContext(op.Ctx, command[0], command[1:]...)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			// 显式代理（包括空值）覆盖父进程的代理环境。
			for _, value := range os.Environ() {
				key, _, _ := strings.Cut(value, "=")
				if !strings.HasSuffix(strings.ToLower(key), "_proxy") {
					cmd.Env = append(cmd.Env, value)
				}
			}
			for _, key := range []string{"http_proxy", "https_proxy", "all_proxy", "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY"} {
				cmd.Env = append(cmd.Env, key+"="+op.ProxyAddress)
			}
			cmd.Env = append(cmd.Env, "no_proxy=", "NO_PROXY=")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("%s 安装失败: %w", binary, errors.Join(op.Ctx.Err(), err))
			}
			filename, err = exec.LookPath(binary)
			if err != nil {
				return fmt.Errorf("安装完成但找不到 %s: %w", binary, err)
			}
		}
		filename, err = filepath.Abs(filename)
		if err != nil {
			return err
		}
		target, err := filepath.Abs(op.Filename())
		if err != nil || target == filename {
			return err
		}
		link := filepath.Join(temp, "binary")
		if err := os.Symlink(filename, link); err != nil {
			return err
		}
		return os.Rename(link, target)
	}
}

var Resources = indexResources(map[string]*Info{
	"build_win.sh":       {Local: "build_win.sh", Key: []string{"build_win"}},
	"service.service":    {Local: "service.service", Key: []string{"service"}},
	"install_minio.sh":   {Local: "install_minio.sh", Key: []string{"install_minio"}},
	"install_nodered.sh": {Local: "install_nodered.sh", Key: []string{"install_nodered"}},
	"install_v2raya.sh":  {Local: "install_v2raya.sh", Key: []string{"install_v2raya"}},
	"golang.md":          {Local: "GOLANG.md", Key: []string{"go.md", "GOLANG.md"}},
}, Exclusive)

// indexResources 先合并定义、补齐名称，再建立独立索引；后面的清单覆盖同名定义。
func indexResources(groups ...map[string]*Info) map[string]*Info {
	definitions := make(map[string]*Info)
	for _, group := range groups {
		for key, info := range group {
			definitions[key] = info
		}
	}
	index := make(map[string]*Info)
	for key, info := range definitions {
		if info.Local == "" {
			info.Local = key
			if name, err := info.GetRemote(); err == nil {
				info.Local = name
			}
		}
		index[info.Local] = info
		for _, alias := range info.Key {
			if alias != "" {
				index[alias] = info
			}
		}
	}
	for key, info := range definitions {
		index[key] = info
	}
	return index
}

// downloadReleaseBinary 只安装发布包中的指定程序，成功后再替换目标。
func downloadReleaseBinary(name string) Handler {
	return func(op *Task) error {
		dir := filepath.Dir(op.Filename())
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		temp, err := os.MkdirTemp(dir, ".i-install-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(temp)
		u, err := url.Parse(op.URL())
		if err != nil {
			return err
		}
		archive := filepath.Join(temp, "package")
		binary := filepath.Join(temp, "binary")
		if err := op.DownloadTo(archive); err != nil {
			return err
		}
		ext := strings.ToLower(u.Path)
		if strings.HasSuffix(ext, ".zip") || strings.HasSuffix(ext, ".tar.gz") {
			if err := extractReleaseBinary(archive, binary, name, strings.HasSuffix(ext, ".zip")); err != nil {
				return err
			}
		} else {
			binary = archive
		}
		if err := ensureExecutable(binary); err != nil {
			return err
		}
		return os.Rename(binary, op.Filename())
	}
}

func extractReleaseBinary(archive, filename, name string, isZip bool) error {
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()
	found := false
	copyBinary := func(r io.Reader) error {
		if found {
			return fmt.Errorf("发布包中存在多个 %s", name)
		}
		found = true
		_, err := io.Copy(out, r)
		return err
	}
	if isZip {
		z, err := zip.OpenReader(archive)
		if err != nil {
			return err
		}
		defer z.Close()
		for _, f := range z.File {
			if path.Base(f.Name) != name || !f.Mode().IsRegular() {
				continue
			}
			r, err := f.Open()
			if err != nil {
				return err
			}
			err = copyBinary(r)
			closeErr := r.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		}
	} else {
		f, err := os.Open(archive)
		if err != nil {
			return err
		}
		defer f.Close()
		gz, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gz.Close()
		tr := tar.NewReader(gz)
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if path.Base(h.Name) == name && (h.Typeflag == tar.TypeReg || h.Typeflag == tar.TypeRegA) {
				if err := copyBinary(tr); err != nil {
					return err
				}
			}
		}
		// tar 的结束标记可能早于 gzip 尾部，读完以校验压缩流。
		if _, err := io.Copy(io.Discard, gz); err != nil {
			return err
		}
	}
	if !found {
		return fmt.Errorf("发布包中未找到 %s", name)
	}
	return out.Close()
}
