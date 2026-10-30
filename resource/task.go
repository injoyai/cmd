package resource

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/injoyai/goutil/oss"
)

// Task 是一次下载的配置快照及运行状态，通过NewTask创建。
type Task struct {
	Config
	Ctx        context.Context
	Client     *http.Client
	CacheDir   string
	url        string
	filename   string
	proxyRules []*regexp.Regexp
	tools      map[string]string
}

// NewTask 复制输入配置，校验代理并解析初始输出路径，不修改调用方配置。
func NewTask(config *Config) (*Task, error) {
	if config == nil {
		return nil, errors.New("下载配置不能为空")
	}
	task := &Task{Config: *config, url: config.SourceURL, Ctx: context.Background()}
	task.ProxyIgnore = append([]string(nil), config.ProxyIgnore...)
	if task.ProxyAddress != "" {
		proxy, err := url.Parse(task.ProxyAddress)
		if err != nil || proxy.Hostname() == "" {
			// 底层URL解析错误会回显完整代理地址，避免把凭据带入错误日志。
			return nil, errors.New("代理地址格式无效，请检查主机、端口和转义字符")
		}
		switch proxy.Scheme {
		case "", "http", "https", "socks5", "socks5h":
		default:
			return nil, errors.New("不支持的代理协议，请使用HTTP、HTTPS或SOCKS5代理")
		}
	}
	rules, err := task.compileProxyRules()
	if err != nil {
		return nil, err
	}
	task.proxyRules = rules
	task.SetFilename("", "")
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		address := task.proxyFor(req.URL.String())
		if address == "" {
			return nil, nil
		}
		return url.Parse(address)
	}
	task.Client = &http.Client{Transport: transport}
	return task, nil
}

// Close 释放任务持有的空闲连接；包级Download会自动调用。
func (this *Task) Close() {
	this.Client.CloseIdleConnections()
}

func (this *Task) SetFilename(name, extension string) {
	if name == "" {
		name = sourceFilename(this.Resource)
	}
	this.filename = resolveFilename(name, this.Name, extension)
}

func sourceFilename(resource string) string {
	if u, err := url.Parse(resource); err == nil {
		name := path.Base(u.Path)
		if name != "" && name != "." && name != "/" {
			return name
		}
	}
	return time.Now().Format("20060102150405")
}

// resolveFilename 保留完整默认名称；只在需要转换格式或自定义名称缺后缀时处理扩展名。
func resolveFilename(defaultName, customName, extension string) string {
	defaultExt := filenameExtension(defaultName)
	if extension == "" {
		extension = defaultExt
	}
	if customName != "" {
		if filenameExtension(customName) == "" {
			return customName + extension
		}
		return customName
	}
	return strings.TrimSuffix(defaultName, defaultExt) + extension
}

func filenameExtension(name string) string {
	ext := filepath.Ext(name)
	// .gitignore等隐藏文件的整个名称不作为扩展名，避免得到空缓存目录名。
	if ext == filepath.Base(name) {
		return ""
	}
	return ext
}

func (this *Task) String() string {
	return fmt.Sprintf("开始下载: %s  代理: %s", RedactURL(this.URL()), RedactURL(this.Proxy()))
}

// RedactURL 只用于展示，实际请求继续使用原始地址。
func RedactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<无效地址>"
	}
	if u.User != nil {
		u.User = url.User("***")
	}
	return u.String()
}

func (this *Task) SetURL(raw string) {
	this.url = raw
}

func (this *Task) URL() string {
	if this.url != "" {
		return this.url
	}
	return this.Resource
}

func (this *Task) compileProxyRules() ([]*regexp.Regexp, error) {
	rules := make([]*regexp.Regexp, 0, len(this.ProxyIgnore))
	if this.ProxyAddress == "" {
		return rules, nil
	}
	for i, pattern := range this.ProxyIgnore {
		if pattern == "" {
			continue
		}
		rule, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("代理忽略规则 %d 无效: %w", i+1, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// Proxy 根据当前实际地址和本任务已校验的规则选择代理。
func (this *Task) Proxy() string {
	return this.proxyFor(this.URL())
}

func (this *Task) proxyFor(raw string) string {
	if this.ProxyAddress == "" {
		return ""
	}
	for _, rule := range this.proxyRules {
		if rule.MatchString(raw) {
			return ""
		}
	}
	return this.ProxyAddress
}

// GetName 返回输出名称去掉最后一层扩展名后的部分。
func (this *Task) GetName() string {
	return strings.TrimSuffix(this.filename, filenameExtension(this.filename))
}

func (this *Task) Filename() string {
	return filepath.Join(this.Dir, this.filename)
}

func (this *Task) exists() bool {
	return !this.Cover && oss.Exists(this.Filename())
}

func (this *Task) TempFilename() string {
	return this.Filename() + ".temp"
}

func (this *Task) TempDir() string {
	if this.CacheDir != "" {
		return this.CacheDir
	}
	id := sha256.Sum256([]byte(this.URL() + "\x00" + this.Filename()))
	return filepath.Join(filepath.Dir(this.Filename()), fmt.Sprintf(".i-download-%x", id[:16]))
}
