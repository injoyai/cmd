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
			op.Errorf("执行安装: %s\n", strings.Join(command, " "))
			cmd := exec.CommandContext(op.Ctx, command[0], command[1:]...)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, op.ErrorWriter(), op.ErrorWriter()
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

var Resources = indexResources(map[string]*Info{}, Exclusive)

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

// downloadReleaseBinary 安装指定程序及配套文件，全部解压成功后再替换目标。
func downloadReleaseBinary(name string, companions ...string) Handler {
	return func(op *Task) (result error) {
		dir := filepath.Dir(op.Filename())
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		temp, err := os.MkdirTemp(dir, ".i-install-*")
		if err != nil {
			return err
		}
		cleanup := true
		defer func() {
			if cleanup {
				os.RemoveAll(temp)
			}
		}()
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
		if strings.HasSuffix(ext, ".tar.xz") {
			unpacked := filepath.Join(temp, "package.tar")
			out, err := os.Create(unpacked)
			if err != nil {
				return err
			}
			cmd := exec.CommandContext(op.Ctx, "xz", "-dc", archive)
			cmd.Stdout = out
			runErr := cmd.Run()
			if err := errors.Join(runErr, out.Close()); err != nil {
				return fmt.Errorf("解压 .tar.xz 发布包（需要安装 xz）: %w", err)
			}
			archive, ext = unpacked, ".tar"
		}
		format := ""
		for _, suffix := range []string{".zip", ".tar.gz", ".tar"} {
			if strings.HasSuffix(ext, suffix) {
				format = suffix
				break
			}
		}
		if format != "" {
			if err := extractReleaseBinary(archive, binary, name, format); err != nil {
				return err
			}
		} else {
			binary = archive
		}
		if err := ensureExecutable(binary); err != nil {
			return err
		}
		if len(companions) == 0 {
			return os.Rename(binary, op.Filename())
		}
		if format == "" {
			return fmt.Errorf("配套文件必须从发布归档中提取")
		}
		type installFile struct {
			source, target, backup string
			installed              bool
		}
		files := make([]installFile, 0, len(companions)+1)
		for _, companion := range companions {
			target := filepath.Join(dir, companion)
			if strings.EqualFold(target, op.Filename()) {
				return fmt.Errorf("程序文件名与配套文件冲突: %s", companion)
			}
			filename := filepath.Join(temp, companion)
			if err := extractReleaseBinary(archive, filename, companion, format); err != nil {
				return err
			}
			files = append(files, installFile{source: filename, target: target})
		}
		files = append(files, installFile{source: binary, target: op.Filename()})
		defer func() {
			if result == nil {
				return
			}
			var restoreErr error
			for i := len(files) - 1; i >= 0; i-- {
				file := &files[i]
				if file.installed {
					restoreErr = errors.Join(restoreErr, os.Remove(file.target))
				}
				if file.backup != "" {
					restoreErr = errors.Join(restoreErr, os.Rename(file.backup, file.target))
				}
			}
			if restoreErr != nil {
				cleanup = false
				result = errors.Join(result, fmt.Errorf("恢复旧文件失败，备份保留在 %s: %w", temp, restoreErr))
			}
		}()
		// 配套文件先到位，最后替换主程序；失败时按相反顺序恢复。
		for i := range files {
			file := &files[i]
			if stat, err := os.Lstat(file.target); err == nil {
				if stat.IsDir() {
					return fmt.Errorf("安装目标是目录: %s", file.target)
				}
				backup := filepath.Join(temp, fmt.Sprintf("previous-%d", i))
				if err := os.Rename(file.target, backup); err != nil {
					return err
				}
				file.backup = backup
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err := os.Rename(file.source, file.target); err != nil {
				return err
			}
			file.installed = true
		}
		return nil
	}
}

func extractReleaseBinary(archive, filename, name, format string) error {
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
	if format == ".zip" {
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
		var reader io.Reader = f
		if format == ".tar.gz" {
			gz, err := gzip.NewReader(f)
			if err != nil {
				return err
			}
			defer gz.Close()
			reader = gz
		}
		tr := tar.NewReader(reader)
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
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return err
		}
	}
	if !found {
		return fmt.Errorf("发布包中未找到 %s", name)
	}
	return out.Close()
}
