<p align="center">
  <h1 align="center">kbr (ko-browser)</h1>
  <p align="center">一个简单、快速、节省 Token 的 AI Agent 浏览器工具，支持 CLI 和 Go Library。</p>
</p>

<p align="center">
  <a href="README.md">English</a> •
  <a href="#-快速开始">快速开始</a> •
  <a href="#-命令速查">命令速查</a> •
  <a href="#agent-notes-cn">Agent 阅读说明</a> •
  <a href="docs/snapshot-format.md">快照格式规范</a>
</p>

<p align="center">
  <a href="https://github.com/libi/ko-browser/releases"><img src="https://img.shields.io/github/v/release/libi/ko-browser?style=flat-square" alt="Release"></a>
  <a href="https://github.com/libi/ko-browser/actions"><img src="https://img.shields.io/github/actions/workflow/status/libi/ko-browser/ci.yml?style=flat-square" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/libi/ko-browser"><img src="https://pkg.go.dev/badge/github.com/libi/ko-browser.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/libi/ko-browser"><img src="https://goreportcard.com/badge/github.com/libi/ko-browser?style=flat-square" alt="Go Report Card"></a>
</p>

---

## 简介

**kbr** (ko-browser) 是一个使用 Go 开发的浏览器自动化工具，专为 AI Agent 设计。它同时提供两种使用方式：

- **CLI 命令行工具**，适合 shell 驱动的 agent 工作流
- **Go Library 库**，适合嵌入到 Agent Runtime、工具层或服务端程序中

它的自定义无障碍树快照格式，相比更冗长的同类输出可**节省 46% 以上 token**。

### ✨ 核心优势

- 🚀 **单一二进制文件** — 无需 Node.js，无需 Playwright 运行时
- 🤖 **AI 优化的快照格式** — `id: role "name" states` 格式节省 46%+ token
- 📦 **双重身份** — 既是 CLI 工具，也是可 `go get` 导入的 Go Library
- ⚡ **启动飞快** — ~50ms（Go 二进制）对比 ~500ms（Node.js 方案）
- 🔢 **简洁的元素引用** — `click 5`
- 🔍 **可选 OCR** — 通过 `-tags=ocr` 编译启用 Tesseract，处理图片密集页面
- 🌐 **~86 个命令** — 覆盖常见浏览器自动化工作流

> 📖 阅读完整的[快照格式规范](docs/snapshot-format.md)，了解详细的设计决策、BNF 语法和示例。([English Version](docs/snapshot-format-en.md))

---

## 📦 安装

### 安装方式选择

| 使用场景 | 推荐方式 | 说明 |
|---------|----------|------|
| macOS 本地使用 | Homebrew | 默认安装带 OCR 的版本，并自动拉取 `tesseract` |
| Linux / macOS 手动部署 | GitHub Releases | 下载预编译压缩包，并确保系统里有 Tesseract 运行时 |
| Go 项目集成或自定义构建 | 源码编译 | 适合把 `kbr` 集成进自己的工具链或服务中 |

### Homebrew

```bash
brew tap libi/tap
brew install ko-browser
# 或者不先 tap，直接安装：
brew install libi/tap/ko-browser
```

> Homebrew 安装的是带 OCR 的 `kbr`。
> 它会自动安装 `tesseract`，并通过源码编译。
> Tap 仓库是 [libi/homebrew-tap](https://github.com/libi/homebrew-tap)。

### 预编译二进制

从 [GitHub Releases](https://github.com/libi/ko-browser/releases) 下载：

> GitHub Release 提供的二进制同样是带 OCR 的版本。
> 使用前请先安装 Tesseract 运行时库：macOS 用 `brew install tesseract`，Linux 用 `apt install libtesseract-dev`。

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/libi/ko-browser/releases/latest/download/ko-browser-darwin-arm64.tar.gz
tar xzf ko-browser-darwin-arm64.tar.gz
mv kbr /usr/local/bin/kbr

# macOS (Intel)
curl -LO https://github.com/libi/ko-browser/releases/latest/download/ko-browser-darwin-amd64.tar.gz
tar xzf ko-browser-darwin-amd64.tar.gz
mv kbr /usr/local/bin/kbr

# Linux (amd64)
curl -LO https://github.com/libi/ko-browser/releases/latest/download/ko-browser-linux-amd64.tar.gz
tar xzf ko-browser-linux-amd64.tar.gz
mv kbr /usr/local/bin/kbr
```

### 源码编译

```bash
# 安装带 OCR 支持的 kbr（需要先安装 Tesseract）
CGO_ENABLED=1 go install -tags=ocr github.com/libi/ko-browser/cmd/kbr@latest
```

> 发布出去的安装方式统一使用 OCR 版本。
> OCR 依赖 Tesseract：`brew install tesseract`（macOS）/ `apt install libtesseract-dev`（Linux）。

### 安装浏览器

```bash
kbr install              # 检查并下载 Chromium
kbr install --with-deps  # 同时安装系统依赖（Linux）
```

### 运行依赖

- 必须有可用的 Chrome 或 Chromium；如果本机没有，先执行 `kbr install`
- 带 OCR 的发布版本依赖宿主机上的 Tesseract 运行时库
- 在 Linux CI 或受限沙箱环境中，Chrome 可能需要无沙箱模式等兼容参数才能启动

---

## 🚀 快速开始

### Agent 常用循环

这是最常见的 shell 驱动 Agent 工作流：

```bash
kbr open https://example.com
kbr snapshot
kbr click 5
kbr type 8 "hello world"
kbr press Enter
kbr wait load
kbr get text 12
```

思路就是：先 `snapshot` 读页面结构，再根据数字 ID 操作元素；页面变化后重新获取快照。

### CLI 使用

```bash
# 打开页面并获取快照
kbr open https://www.baidu.com
kbr snapshot

# 输出：
# Page: "百度一下，你就知道"
#
# 1: link "新闻"
# 2: link "hao123"
# 3: textbox "搜索" focused
# 4: button "百度一下"

# 通过 ID 与元素交互
kbr click 3
kbr type 3 "ko-browser"
kbr press Enter

# 截图
kbr screenshot result.png

# 关闭浏览器
kbr close
```

### Go Library 使用

```go
package main

import (
    "fmt"
    "log"

    "github.com/libi/ko-browser/browser"
)

func main() {
    b, err := browser.New(browser.Options{Headless: true})
    if err != nil {
        log.Fatal(err)
    }
    defer b.Close()

    b.Open("https://www.baidu.com")

    snap, _ := b.Snapshot()
    fmt.Println(snap.Text)
    // 1: link "新闻"
    // 2: link "hao123"
    // 3: textbox "搜索" focused
    // 4: button "百度一下"

    b.Click(3)
    b.Type(3, "hello world")
    b.Press("Enter")
}
```

---

## 📖 命令速查

### 常见任务入口

| 目标 | 常用命令 |
|------|----------|
| 打开页面并读取信息 | `open`、`snapshot`、`get title`、`get url` |
| 表单和交互操作 | `click`、`type`、`fill`、`select`、`check`、`press` |
| 等待页面变化 | `wait load`、`wait selector`、`wait text`、`wait url` |
| 调试和核对结果 | `screenshot`、`console messages`、`errors list`、`highlight` |
| 会话与状态管理 | `tab *`、`cookies *`、`storage *`、`state *` |

### 核心交互

| 命令 | CLI 用法 | Library API |
|------|---------|-------------|
| **打开** | `kbr open <url>` | `b.Open(url)` |
| **点击** | `kbr click <id>` | `b.Click(id)` |
| **双击** | `kbr dblclick <id>` | `b.DblClick(id)` |
| **输入** | `kbr type <id> "文本"` | `b.Type(id, text)` |
| **填充** | `kbr fill <id> "文本"` | `b.Fill(id, text)` |
| **按键** | `kbr press <key>` | `b.Press(key)` |
| **悬停** | `kbr hover <id>` | `b.Hover(id)` |
| **聚焦** | `kbr focus <id>` | `b.Focus(id)` |
| **勾选** | `kbr check <id>` | `b.Check(id)` |
| **取消勾选** | `kbr uncheck <id>` | `b.Uncheck(id)` |
| **下拉选择** | `kbr select <id> "值"` | `b.Select(id, vals...)` |
| **滚动** | `kbr scroll down 500` | `b.Scroll("down", 500)` |
| **拖拽** | `kbr drag <源> <目标>` | `b.Drag(srcID, dstID)` |
| **关闭** | `kbr close` | `b.Close()` |

### 快照与截图

```bash
kbr snapshot                     # 完整无障碍树
kbr snapshot -i                  # 仅可交互元素
kbr snapshot -c                  # 紧凑模式
kbr snapshot -d 5                # 最大深度 5
kbr snapshot --ocr               # 启用 OCR

kbr screenshot out.png           # 视口截图
kbr screenshot --full out.png    # 全页截图
kbr screenshot --annotate out.png # 带 ID 标注的截图
```

### 导航

```bash
kbr back                         # 后退
kbr forward                      # 前进
kbr reload                       # 刷新
```

### 获取信息

```bash
kbr get title                    # 页面标题
kbr get url                      # 当前 URL
kbr get text <id>                # 元素文本
kbr get html <id>                # 元素 HTML
kbr get value <id>               # 输入框的值
kbr get attr <id> href           # 元素属性
kbr get count ".item"            # 匹配元素数量
kbr get box <id>                 # 边界框
kbr get styles <id>              # 计算样式
```

### 状态检查

```bash
kbr is visible <id>              # 是否可见 → true/false
kbr is enabled <id>              # 是否启用
kbr is checked <id>              # 是否选中
```

### 元素查找

```bash
kbr find role button --name Submit   # 按角色查找
kbr find text "登录"                  # 按文本查找
kbr find label "邮箱"                 # 按标签查找
kbr find placeholder "搜索..."        # 按占位符查找
kbr find testid "login-form"         # 按 data-testid 查找
kbr find first ".card"               # 第一个匹配
kbr find nth 2 ".card"               # 第 N 个匹配
```

### 等待

```bash
kbr wait time 2s                 # 等待 2 秒
kbr wait selector "#loading"     # 等待元素出现
kbr wait url "**/dashboard"      # 等待 URL 匹配
kbr wait load                    # 等待页面加载完成
kbr wait text "欢迎"              # 等待文本出现
kbr wait fn "window.ready"       # 等待 JS 表达式为真
kbr wait hidden "#spinner"       # 等待元素隐藏
kbr wait download ./file.pdf     # 等待下载完成
```

### 多标签页

```bash
kbr tab list                     # 列出标签页
kbr tab new https://example.com  # 新建标签页
kbr tab switch 2                 # 切换到第 2 个
kbr tab close 1                  # 关闭第 1 个
```

### 网络控制

```bash
kbr network route "**/api/*" --action block   # 拦截请求
kbr network unroute "**/api/*"                # 移除规则
kbr network requests                          # 查看请求日志
```

### Cookie 与存储

```bash
kbr cookies get                  # 获取所有 Cookie
kbr cookies set token "abc" --domain example.com
kbr cookies clear                # 清除所有 Cookie

kbr storage get theme --type local    # 读取 localStorage
kbr storage set theme "dark" --type local
kbr storage clear --type local
```

### 浏览器设置

```bash
kbr set viewport 1920 1080       # 设置视口大小
kbr set device "iPhone 12"       # 模拟设备
kbr set geo 39.9 116.4           # 设置地理位置
kbr set offline true             # 启用离线模式
kbr set media dark               # 深色模式
```

### 调试

```bash
kbr console messages             # 查看控制台日志
kbr errors list                  # 查看页面错误
kbr highlight <id>               # 高亮元素
kbr inspect                      # 打开 DevTools
kbr clipboard read               # 读取剪贴板
```

### Selector 语法

kbr 支持三种选择器格式，自动识别：

| 输入 | 类型 | 说明 |
|------|------|------|
| `5` | 快照 ID | 对应快照中的 `5:` |
| `"#submit"` | CSS 选择器 | 以 `.` `#` `[` 等开头 |
| `"//button"` | XPath | 以 `//` 或 `/` 开头 |

```bash
kbr click 5                           # 点击 5 号元素
kbr click "#submit-button"            # CSS 选择器
kbr click "//button[@type='submit']"  # XPath
```

---

## ⚙️ 配置文件

kbr 按以下优先级加载配置（低 → 高）：

1. `~/.ko-browser/config.json` — 用户级默认
2. `./ko-browser.json` — 项目级覆盖
3. 环境变量（`KO_BROWSER_*`）
4. CLI 标志 — 覆盖一切

```json
{
  "headed": true,
  "proxy": "http://localhost:8080",
  "profile": "./browser-data",
  "screenshotFormat": "jpeg",
  "downloadPath": "./downloads"
}
```

---

## 🔧 全局选项

| 选项 | 说明 |
|------|------|
| `--session <name>` | 隔离会话名称（默认: "default"） |
| `--headed` | 显示浏览器窗口 |
| `--json` | JSON 格式输出 |
| `--timeout <duration>` | 操作超时时间（默认: 30s） |
| `--profile <path>` | Chrome 用户数据目录（持久化会话） |
| `--state <path>` | 加载保存的浏览器状态 |
| `--config <path>` | 配置文件路径 |
| `--user-agent <ua>` | 自定义 User-Agent |
| `--proxy <url>` | 代理服务器 URL |
| `--ignore-https-errors` | 忽略 HTTPS 证书错误 |
| `--extension <path>` | 加载 Chrome 扩展（可重复） |
| `--download-path <path>` | 默认下载目录 |
| `--screenshot-dir <path>` | 默认截图输出目录 |
| `--screenshot-format <fmt>` | 截图格式: png, jpeg |
| `--content-boundaries` | 用边界标记包裹输出 |
| `--debug` | 调试输出 |

---

## 🏗️ 项目架构

```
kbr
├── cmd/kbr/      ★ CLI 入口 — `go install .../cmd/kbr@latest`
├── browser/      ★ 公开包 — 核心浏览器 API（可 go get 导入）
├── axtree/       ★ 公开包 — AX Tree 提取、过滤、格式化
├── selector/     ★ 公开包 — 元素选择器解析（ID/CSS/XPath）
├── ocr/          ★ 公开包 — Tesseract OCR 引擎（需 build tag: ocr）
├── cmd/            CLI 层 — cobra 命令定义
└── internal/       内部包 — 守护进程、会话管理（仅 CLI 使用）
```

`browser/`、`axtree/`、`selector/`、`ocr/` 均为公开包，可通过 `go get` 导入。`internal/` 仅供 CLI 守护进程使用。OCR 需编译时添加 `-tags=ocr`。

---

<a name="agent-notes-cn"></a>

## Agent 阅读说明

如果你是正在阅读本 README 的 Agent，可以把 `kbr` 理解成一个带紧凑文本接口的可交互浏览器。

- 优先使用 `kbr snapshot` 获取页面结构。它最省 token，也会返回稳定的数字引用。
- 拿到快照后，优先用数字 ID 操作元素，例如 `kbr click 5`、`kbr type 8 "hello"`、`kbr get text 12`。
- 页面跳转或 DOM 变化后要重新获取快照。元素 ID 只在当前快照内有效，不是永久选择器。
- 优先使用 `kbr get`、`kbr find`、`kbr wait` 做定向读取，不要反复依赖整页截图。
- 只有在页面内容主要存在于图片中，或 DOM / 无障碍树拿不到信息时，再启用 OCR。
- 当你已经明确知道目标元素，或当前上下文里没有快照 ID 时，再退回使用 CSS / XPath 选择器。
- 通过 Homebrew 或 GitHub Release 安装的发布版本都带 OCR，因此宿主机需要可用的 Tesseract 运行时库。

### 作为 Agent 技能安装

如果你的 Agent 框架支持本地技能，先安装 `kbr`，再把仓库里的 `skills/ko-browser` 目录复制到 Agent 的技能目录中，并命名为 `ko-browser/`。

```yaml
<agent-skills-dir>/
  ko-browser/
    SKILL.md
```

仓库里的技能文件在这里：

- [`skills/ko-browser/SKILL.md`](https://github.com/libi/ko-browser/blob/main/skills/ko-browser/SKILL.md)

示例：

```bash
mkdir -p <agent-skills-dir>/ko-browser
cp skills/ko-browser/SKILL.md <agent-skills-dir>/ko-browser/SKILL.md
```

复制完成后，再确保 Agent 运行环境的 `PATH` 中可以找到 `kbr` 二进制。如果 Agent 宿主还支持直接注册可执行工具，也可以同时把 `kbr` 暴露为可调用工具。

---

## 🤝 参与贡献

欢迎提交 Pull Request！

```bash
git clone https://github.com/libi/ko-browser.git
cd ko-browser
go build -o kbr ./cmd/kbr/              # 不含 OCR
go build -tags=ocr -o kbr ./cmd/kbr/     # 含 OCR
go test ./tests/ -v -timeout 180s
```

---

## 📄 许可证

MIT
