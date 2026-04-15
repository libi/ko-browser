# Agents Guide — ko-browser

本文件面向 AI Agent / Copilot，在修改本项目时请遵守以下规范。

---

## ⚠️ 跨平台兼容（强制）

**所有改动必须同时兼容 Windows、macOS 和 Linux。** 提交前须通过以下交叉编译验证：

```bash
GOOS=darwin  GOARCH=arm64 go build ./...
GOOS=linux   GOARCH=amd64 go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

### 常见陷阱

| 场景 | ❌ 错误做法 | ✅ 正确做法 |
|------|-----------|-----------|
| 进程属性 | 直接写 `syscall.SysProcAttr{Setsid: true}` | 用构建标签拆分为 `_unix.go` / `_windows.go` |
| 路径拼接 | 手写 `/` 或 `\` | 使用 `filepath.Join()` |
| 路径分隔符 | 硬编码 `"/"` | 使用 `filepath.Separator` 或 `filepath.ToSlash()` |
| 换行符 | 假设 `\n` | 输出用 `fmt.Println`；解析用 `bufio.Scanner` |
| 临时目录 | 硬编码 `/tmp` | 使用 `os.TempDir()` |
| 空设备 | 硬编码 `/dev/null` | 使用 `os.DevNull` |
| 可执行文件名 | 假设无后缀 | Windows 上需要 `.exe`；使用 `os.Executable()` |
| 文件权限 | `os.Chmod(path, 0600)` | Windows 上 chmod 无效果但不会报错，可接受 |
| 信号处理 | 使用 Unix-only 信号（`SIGUSR1`） | 仅使用 `os.Interrupt`；平台特定信号用构建标签隔离 |
| IPC 通信 | Unix domain socket | Win10 1803+ 支持；如需更广兼容可考虑 named pipe |
| 环境变量 | 区分大小写假设 | Windows 环境变量不区分大小写 |
| exec 命令 | `exec.Command("sh", "-c", ...)` | 需要 shell 时按平台区分，或直接 exec 目标二进制 |

### 构建标签模式

当有平台差异时，使用以下文件命名和构建标签：

```
foo_unix.go      //go:build !windows
foo_windows.go   //go:build windows
```

两个文件导出相同的函数签名，调用方无感知。参考本项目中的 `internal/session/daemon_unix.go` 和 `daemon_windows.go`。

---

## Go 编码规范

### 项目结构

```
cmd/kbr/        CLI 入口（main.go）
cmd/            cobra 命令定义
browser/        公开包 — 核心浏览器 API
axtree/         公开包 — AX Tree 提取 / 过滤 / 格式化
selector/       公开包 — 元素选择器解析
ocr/            公开包 — OCR 引擎（需 -tags=ocr）
internal/       内部包 — 仅 CLI 使用（daemon、session 管理）
tests/          集成测试
testdata/       测试用 HTML 等静态资源
```

- `browser/`、`axtree/`、`selector/`、`ocr/` 是公开 API，修改时注意向后兼容。
- `internal/` 下的类型不会被外部引用，可自由重构。

### 命名

- 文件名：小写 + 下划线，如 `dom_helpers.go`、`screenshot_cmd.go`。
- 包名：短小、小写、单个单词（`session`，不要 `sessionManager`）。
- 导出函数/类型：`CamelCase`；未导出：`camelCase`。
- 接口：单方法接口用 `-er` 后缀（`Reader`、`Closer`）。
- Error 变量：`Err` 前缀（`ErrNotFound`）；error 类型：`Error` 后缀。
- 常量：`CamelCase`（Go 风格），不要 `ALL_CAPS`。

### 错误处理

```go
// ✅ 包装上下文
if err := doSomething(); err != nil {
    return fmt.Errorf("do something: %w", err)
}

// ✅ 哨兵错误用 errors.Is 判断
if errors.Is(err, os.ErrNotExist) { ... }

// ❌ 不要用字符串比较错误
if err.Error() == "not found" { ... }
```

- 优先用 `fmt.Errorf("context: %w", err)` 包装。
- 不要忽略 error，除非有明确注释：`_ = conn.Close() // best-effort`。
- 不要 `panic`，除非是程序初始化阶段的不可恢复错误。

### 并发

- 共享可变状态必须加锁（`sync.Mutex`）或通过 channel 通信。
- goroutine 必须有明确的退出条件，不允许泄漏。
- 使用 `context.Context` 控制超时和取消。
- `sync.WaitGroup` 配合 goroutine 确保优雅退出。

### 依赖与导入

- 标准库优先。引入新的第三方依赖需要有充分理由。
- 导入分三组，空行分隔：标准库 → 第三方 → 项目内部。

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/libi/ko-browser/browser"
)
```

### 测试

- 测试文件与被测文件同目录或放在 `tests/`（集成测试）。
- 表驱动测试优先：

```go
tests := []struct {
    name string
    input string
    want  string
}{
    {"empty", "", "default"},
    {"normal", "hello", "hello"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got := process(tt.input)
        if got != tt.want {
            t.Errorf("process(%q) = %q, want %q", tt.input, got, tt.want)
        }
    })
}
```

- 使用 `testdata/` 存放测试用的静态文件。
- 集成测试通过 `go test ./tests/ -v -timeout 180s` 运行。

### 文档

- 所有导出的函数、类型、常量必须有 godoc 注释。
- 注释以被描述的名字开头：`// Open navigates to the given URL.`
- CLI 命令的 `Short` 和 `Long` 字段要填写清楚。

### 提交

- 提交信息格式：`type: short description`
- type 可选：`feat`、`fix`、`refactor`、`docs`、`test`、`chore`
- 每个提交做一件事；大改动拆分多次提交。

---

## 检查清单

每次修改完成后，对照以下清单：

- [ ] `go build ./...` 通过
- [ ] `go vet ./...` 无警告
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` 通过
- [ ] `GOOS=linux GOARCH=amd64 go build ./...` 通过
- [ ] 新增导出 API 有 godoc 注释
- [ ] 错误用 `%w` 包装了上下文
- [ ] 无平台特定的硬编码路径或系统调用
- [ ] 新增 goroutine 有退出机制
