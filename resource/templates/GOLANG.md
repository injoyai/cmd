# Golang 编码规范与 AI 约束

> 本文件为 Go 项目的权威编码规范。
>
> 当 `AGENTS.md` 加载本文件后，所有 Go 代码生成、修改、重构、Review 建议必须遵守以下规则。
>
> 本规范优先级高于通用代码约定。

---

# 1. 基础原则

## 1.1 编码目标

代码应满足：

- 可读性优先
- 简单优先
- 明确优先
- 可维护优先
- 避免过度设计

禁止：

- 为未来不存在的需求提前设计复杂架构
- 引入没有实际收益的抽象
- 为了所谓性能优化牺牲代码可读性

---

# 2. 格式化与代码风格

## 2.1 gofmt

所有 Go 代码必须通过：

```bash
gofmt -s
```

要求：

- AI 输出代码必须符合 gofmt 格式
- 不允许存在明显格式问题

## 2.2 import 管理

推荐使用：

```bash
goimports
```

或 IDE 自动整理。

import 推荐分组：

```go
import (
    "context"
    "fmt"

    "github.com/pkg/errors"

    "project/internal/user"
)
```

分为：

1. 标准库
2. 第三方库
3. 项目内部包

> 优先使用标准库 `errors.Is` / `errors.As` + `fmt.Errorf("%w", err)` 完成错误包装与判别；
> `github.com/pkg/errors` 已进入 maintenance-only 模式，新项目不建议引入。

## 2.3 行长度

推荐：

- 单行不超过 120 字符

但：

- 可读性优先
- 不强制为了长度拆分代码

---

# 3. 命名规范

## 3.1 包名

规则：

- 全小写
- 不使用下划线
- 不使用复数
- 简短、有意义

推荐：

```go
user
device
config
server
```

避免：

```go
utils
common
helpers
misc
users
```

---

## 3.2 文件命名

规则：

- 小写
- 单词之间允许 `_`

推荐：

```
user.go
user_service.go
http_client.go
```

避免：

```
UserService.go
userService.go
```

测试文件：

```
xxx_test.go
```

---

## 3.3 标识符

导出：

```go
UserService
HTTPServer
DeviceID
```

非导出：

```go
userService
httpServer
deviceID
```

缩写规则：

推荐：

```go
HTTP
URL
ID
IP
TCP
JSON
```

例如：

正确：

```go
userID
HTTPServer
URLParser
```

错误：

```go
userId
HttpServer
UrlParser
```

---

# 4. 变量与常量

## 4.1 变量命名

短作用域：

允许：

```go
for i := 0; i < 10; i++ {
}
```

允许：

```go
ctx
err
req
resp
```

禁止：

```go
data
info
obj
tmp
```

作为长期变量名。

---

## 4.2 常量

禁止：

```go
MAX_SIZE
DEFAULT_TIMEOUT
```

推荐：

```go
MaxSize
DefaultTimeout
```

枚举除外。

---

# 5. 错误处理规范

## 5.1 错误必须处理

必须处理影响程序正确性的 error。

推荐：

```go
if err != nil {
    return fmt.Errorf("create user: %w", err)
}
```

错误包装：

优先：

```go
fmt.Errorf("%w", err)
```

支持：

```go
errors.Is
errors.As
```

---

## 5.2 允许忽略的错误

以下情况允许：

```go
defer file.Close()
```

或者：

```go
_ = conn.Close()
```

但需要说明原因：

```go
// Ignore close error because resource cleanup failure is not recoverable.
_ = file.Close()
```

---

## 5.3 错误信息

错误字符串：

正确：

```go
user not found
```

错误：

```go
User not found.
```

规则：

- 小写开头
- 不加句号

---

## 5.4 panic 使用

禁止业务代码使用 panic。

允许：

- 程序初始化失败
- 不可恢复状态

例如：

```go
func initConfig() {
    cfg, err := loadConfig()
    if err != nil {
        panic(fmt.Errorf("init config: %w", err))
    }
    _ = cfg
}
```

---

## 5.5 recover

recover 只能用于边界层：

例如：

- HTTP入口
- RPC入口
- goroutine入口

禁止：

- 普通业务逻辑 recover

---

# 6. Context 规范

## 6.1 使用规则

涉及：

- IO
- 网络
- 数据库
- RPC
- 长时间任务

必须：

```go
func Do(ctx context.Context) error
```

Context 应作为第一个参数显式传递。

---

## 6.2 禁止保存 Context

禁止：

```go
type Service struct {
    ctx context.Context
}
```

Context 应显式传递。

---

# 7. 并发规范

## 7.1 goroutine 生命周期

启动 goroutine 必须明确：

- 如何退出
- 谁负责关闭
- 如何释放资源

禁止：

```go
go func() {
    for {
        work()
    }
}()
```

除非存在明确退出机制。

推荐使用 `errgroup.Group` 管理一组 goroutine 并传递 context 取消：

```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error {
    return worker(ctx)
})
if err := g.Wait(); err != nil {
    return err
}
```

---

## 7.2 channel

规则：

- 创建者负责关闭
- 接收者不要关闭
- 明确所有权

推荐：

```go
func Worker(out chan<- Result)
```

---

## 7.3 同步

优先：

- sync.WaitGroup
- errgroup
- Mutex
- RWMutex

避免：

- 全局变量
- 复杂锁嵌套

---

## 7.4 竞态检测

重要代码必须运行：

```bash
go test -race ./...
```

---

# 8. 接口设计

## 8.1 原则

接口应该：

- 小
- 专注
- 面向行为

推荐：

```go
type Reader interface {
    Read([]byte) error
}
```

避免：

```go
type UserManager interface {
    Create()
    Update()
    Delete()
    Query()
}
```

---

## 8.2 接受接口，返回结构体

推荐：

```go
func NewService(repo Repository) *Service
```

---

## 8.3 接口实现检查

允许：

```go
var _ Reader = (*FileReader)(nil)
```

推荐用于：

- 公共库
- 复杂接口
- 框架代码

---

# 9. 测试规范

## 9.1 测试文件

必须：

```
xxx_test.go
```

---

## 9.2 表驱动测试

核心逻辑推荐：

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name    string
        a, b    int
        want    int
        wantErr bool
    }{
        {"positive", 1, 2, 3, false},
        {"zero", 0, 0, 0, false},
        {"negative", -1, 1, 0, false},
        {"overflow", math.MaxInt, 1, 0, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Add(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
            }
            if !tt.wantErr && got != tt.want {
                t.Errorf("Add(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

覆盖：

- 正常情况
- 边界情况
- 错误情况

---

## 9.3 测试辅助函数

辅助函数必须：

```go
t.Helper()
```

---

## 9.4 并行测试

纯逻辑测试：

允许：

```go
t.Parallel()
```

注意：

- 数据隔离
- 文件隔离
- 数据库隔离

---

# 10. 项目结构

根据项目规模选择。

## 小型项目

推荐：

```
.
├── main.go
├── internal
├── config
```

## 中大型项目

推荐：

```
.
├── cmd
├── internal
├── pkg
├── api
├── config
├── scripts
```

说明：

- cmd：程序入口
- internal：私有代码
- pkg：公共库
- api：接口定义

不要强制套用固定目录。

---

# 11. 依赖管理

必须使用：

```
go.mod
go.sum
```

定期：

```bash
go mod tidy
```

新增第三方依赖必须考虑：

- 是否必要
- 是否长期维护
- 是否有标准库替代

禁止：

为了简单功能引入大型依赖。

---

# 12. 文档与注释

## 12.1 导出注释

所有导出：

必须：

```go
// Client represents a client connection.
type Client struct{}
```

---

## 12.2 注释原则

注释解释：

- 为什么这样做

不要解释：

- 代码已经表达的内容

---

# 13. 日志规范

日志按使用场景区分，禁止一刀切。

## 13.1 服务端 / 库代码

禁止：

```go
fmt.Println()
log.Println()
```

推荐结构化日志：

- slog（标准库，首选）
- zap
- zerolog

日志必须包含：

- 时间
- 级别
- 关键上下文（request id / user id / module 等）

## 13.2 CLI 工具的用户可见输出

CLI 工具给用户看的标准输出允许使用 `fmt.Println` / `fmt.Fprintf(os.Stdout, ...)`。

注意：

- 用户可见输出 ≠ 调试日志，调试信息仍应走结构化日志或 stderr；
- 避免在库代码中使用 `fmt.Println`，库不应擅自向 stdout 写内容。

## 13.3 禁止打印

- 密码
- token
- 密钥

---

# 14. 配置规范

禁止：

硬编码：

```go
password := "123456"
```

推荐：

- 环境变量
- yaml
- json
- 配置中心

例如：

```yaml
database:
  host: localhost
  port: 3306
```

---

# 15. 数据库规范

数据库访问必须：

- 使用 context
- 参数化查询
- 明确事务范围

禁止：

```go
sql := "select * from user where id=" + id
```

推荐：

```go
db.QueryContext(ctx,
    "select * from user where id=?",
    id,
)
```

资源释放：

```go
defer rows.Close()
```

---

# 16. HTTP 客户端规范

## 16.1 必须设置超时

禁止：

```go
http.Get(url) // 使用 DefaultClient，无超时
```

推荐：

```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("http get: %w", err)
}
defer resp.Body.Close()
```

## 16.2 资源释放

必须 `defer resp.Body.Close()`，并读取到 EOF 以便连接可复用：

```go
defer io.Copy(io.Discard, resp.Body)
defer resp.Body.Close()
```

## 16.3 限制响应体

读取外部响应时使用 `io.LimitReader` 防止超大响应导致内存爆炸：

```go
body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB
```

---

# 17. 性能规范

原则：

不要过早优化。

注意：

- 避免复制大型结构体
- 注意内存分配
- 注意锁竞争

slice/map/channel：

正常传递即可。

不要：

为了性能滥用指针。

---

## sync.Pool

仅在：

- 高频创建对象
- benchmark 证明有效

情况下使用。

---

## 性能分析

推荐：

```bash
go test -bench .
```

使用：

- pprof
- trace

---

# 18. 安全规范

禁止：

代码中保存：

- 密码
- token
- API Key

随机数：

安全场景：

使用：

```go
crypto/rand
```

普通随机：

允许：

```go
math/rand
```

> Go 1.20+ 已默认自动 seed，不要再手动调用 `rand.Seed`（该函数在 1.20 已废弃）。

外部输入必须验证：

包括：

- HTTP 参数
- 文件
- 网络数据
- 用户输入

防止：

- SQL 注入
- 命令注入
- XSS
- 路径穿越

HTML：

必须：

```go
html/template
```

---

# 19. 常见陷阱

## 19.1 defer 参数立即求值

```go
// 错误：i 在 defer 语句时即被捕获
defer fmt.Println(i) // 会立即求值 i
```

如需延迟求值，使用闭包：

```go
defer func() { fmt.Println(i) }()
```

## 19.2 循环变量捕获

Go 1.22+ 已修复循环变量复用问题，但跨 goroutine 使用时仍建议显式传参：

```go
for _, item := range items {
    item := item // Go 1.22 之前必需；之后可省略
    go process(item)
}
```

## 19.3 接口 nil 比较

一个包裹了 nil 指针的 interface 不等于 nil：

```go
var p *T
var i any = p
fmt.Println(i == nil) // false
```

返回 error 时尤其注意，应直接返回 `nil` 而非包裹 nil 的具体类型。

---

# 20. 泛型使用

## 20.1 适用场景

合理场景：

- 通用容器 / 数据结构（如 `Set[T]`、`Pool[T]`）
- 通用算法（如 `Filter`、`Map`、`Reduce`）
- 减少重复的样板代码

## 20.2 避免滥用

禁止：

- 为了泛型而泛型（仅一处使用的泛型函数）
- 把简单函数强行抽象成难以理解的形式

原则：可读性优先；只有当泛型版**确有收益**且**保持清晰**时才使用。

---

# 21. Linter 规范

项目推荐配置：

golangci-lint

至少开启：

```yaml
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - ineffassign
    - gosimple
    - unused
    - bodyclose
    - noctx
    - sqlclosecheck
    - errorlint   # 强制 %w + errors.Is/As
    - gosec       # 安全审计
    - revive      # 替代 golint 的规则集
    - misspell    # 拼写检查
    - goconst     # 重复字面量提取为常量
```

CI 要求：

```bash
go test -race -cover ./...
golangci-lint run
```

---

# 22. AI 代码生成约束

AI 生成代码必须：

## 22.1 编译要求

必须：

- import 正确
- 无未使用变量
- 无未使用包
- 类型正确

---

## 22.2 禁止假设不存在代码

AI 不得：

- 创建不存在的函数
- 假设不存在的接口
- 假设未知依赖

如果上下文不足：

必须说明假设。

---

## 22.3 多文件修改

必须说明：

```
文件路径
package 名称
修改内容
```

---

## 22.4 API 兼容

修改已有代码：

禁止：

- 删除公开接口
- 修改函数参数
- 改变结构体字段语义

除非用户明确要求。

---

## 22.5 新增依赖

必须说明：

- 为什么需要
- 替代方案
- 维护情况

---

## 22.6 重构原则

优先：

- 小步修改
- 保持兼容
- 使用接口隔离

避免：

一次性大规模重构。

---

# 版本

版本：1.2

更新时间：2026-07-20

适用范围：

所有 Go 项目。

本规范根据项目实践持续更新。