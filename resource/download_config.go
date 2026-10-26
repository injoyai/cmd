package resource

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/notice"
)

type Dir string

func (this Dir) Join(s ...string) string {
	return filepath.Join(append([]string{string(this)}, s...)...)
}

type Config struct {
	Resource     string   //资源 upx , https://xxx.com/xxx
	Dir          string   //下载目录
	Name         string   //保存名称,不含扩展名时补充默认后缀
	Retry        int      //m3u8分片重试次数
	Coroutine    int      //m3u8分片并发数量
	ProxyEnable  bool     //是否使用代理
	ProxyAddress string   //代理地址
	ProxyIgnore  []string //忽略代理地址
	NoticeEnable bool     //是否使用消息通知
	NoticeText   string   //通知消息内容
	VoiceEnable  bool     //是否使用通知语音
	VoiceText    string   //通知语音内容
	Cover        bool     //是否覆盖下载

	url   string // 通过SetUrl显式设置的地址，不保存下载过程中的镜像选择。
	state *downloadState
}

// downloadState 只属于一次下载，不写回调用方的配置。
type downloadState struct {
	url        string
	filename   string
	proxyRules []*regexp.Regexp
}

func (this *Config) newTask() (*Config, error) {
	task := *this
	task.ProxyIgnore = append([]string(nil), this.ProxyIgnore...)
	if task.ProxyEnable && task.ProxyAddress != "" {
		if _, err := url.Parse(task.ProxyAddress); err != nil {
			// 底层URL解析错误会回显完整代理地址，避免把凭据带入错误日志。
			return nil, errors.New("代理地址格式无效，请检查端口和转义字符")
		}
	}
	rules, err := task.compileProxyRules()
	if err != nil {
		return nil, err
	}
	task.state = &downloadState{
		url:        this.url,
		filename:   resolveFilename(sourceFilename(task.Resource), task.Name, ""),
		proxyRules: rules,
	}
	return &task, nil
}

func (this *Config) setFilename(name, extension string) {
	if name == "" {
		name = sourceFilename(this.Resource)
	}
	if this.state == nil {
		this.state = &downloadState{url: this.url}
	}
	this.state.filename = resolveFilename(name, this.Name, extension)
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

func (this *Config) outputName() string {
	if this.state != nil && this.state.filename != "" {
		return this.state.filename
	}
	return resolveFilename(sourceFilename(this.Resource), this.Name, "")
}

func (this *Config) String() string {
	return fmt.Sprintf("开始下载: %s  代理: %s", redactURL(this.Url()), redactURL(this.Proxy()))
}

// redactURL 只用于展示，实际请求继续使用原始地址。
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<无效地址>"
	}
	if u.User != nil {
		u.User = url.User("***")
	}
	return u.String()
}

func (this *Config) SetUrl(raw string) {
	if this.state != nil {
		this.state.url = raw
		return
	}
	this.url = raw
}

func (this *Config) Url() string {
	current := this.url
	if this.state != nil {
		current = this.state.url
	}
	if current != "" {
		return current
	}
	return this.Resource
}

func (this *Config) compileProxyRules() ([]*regexp.Regexp, error) {
	rules := make([]*regexp.Regexp, 0, len(this.ProxyIgnore))
	if !this.ProxyEnable {
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

// Proxy 保留字符串返回值兼容旧调用；下载入口会先校验所有启用的代理规则。
func (this *Config) Proxy() string {
	if !this.ProxyEnable {
		return ""
	}
	var rules []*regexp.Regexp
	if this.state != nil {
		rules = this.state.proxyRules
	}
	if rules == nil {
		var err error
		rules, err = this.compileProxyRules()
		if err != nil {
			// 无法表达错误的旧接口保留代理；下载入口会返回校验错误。
			return this.ProxyAddress
		}
	}
	for _, rule := range rules {
		if rule.MatchString(this.Url()) {
			return ""
		}
	}
	return this.ProxyAddress
}

// GetName 返回不含最后一层扩展名的名称，不修改配置。
func (this *Config) GetName() string {
	name := this.outputName()
	return strings.TrimSuffix(name, filenameExtension(name))
}

// Filename 返回完整输出路径，不修改配置。
func (this *Config) Filename() string {
	return filepath.Join(this.Dir, this.outputName())
}

func (this *Config) TempFilename() string {
	return this.Filename() + ".temp"
}

func (this *Config) TempDir() string {
	return filepath.Join(this.Dir, this.GetName())
}

func (this *Config) PlayNotice() {
	if this.NoticeEnable {
		tool.PublishNotice(&notice.Message{
			Title:   "下载完成",
			Content: this.NoticeText,
		})
	}
}

func (this *Config) PlayVoice() {
	if this.VoiceEnable {
		notice.DefaultVoice.Speak(this.VoiceText)
	}
}
