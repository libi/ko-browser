
---

<a name="-中文文档"></a>

# 🇨🇳 中文文档

## 简介

**kbr** (ko-browser) 是简单,快速,节省Token的浏览器自动化工具，使用Go语言开发, 专为 AI Agent 设计。它同时提供 **CLI 命令行工具** 和 **Go Library 库**，自定义的无障碍树快照格式比同类工具**节省 46% 以上 token**。

### ✨ 核心优势

- 🚀 **单一二进制文件** — 无需 Node.js，无需 Playwright 运行时
- 🤖 **AI 优化的快照格式** — `id: role "name" states` 格式节省 46%+ token
- 📦 **双重身份** — 既是 CLI 工具，也是可 `go get` 导入的 Go Library
- ⚡ **启动飞快** — ~50ms（Go 二进制）对比 ~500ms（Node.js 方案）
- 🔢 **简洁的元素引用** — `click 5`
- 🔍 **内置 OCR** — 可选的 Tesseract 集成，处理图片密集页面
- 🌐 **~86 个命令** — 完整对标 agent-browser v0.19.0

> 📖 阅读完整的[快照格式规范](docs/snapshot-format.md)，了解详细的设计决策、BNF 语法和示例。([English Version](docs/snapshot-format-en.md))

---

## 📦 安装

### 预编译二进制

从 [GitHub Releases](https://github.com/libi/ko-browser/releases) 下载：

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/libi/ko-browser/releases/latest/download/ko-browser-darwin-arm64.tar.gz
tar xzf ko-browser-darwin-arm64.tar.gz
mv ko-browser-darwin-arm64 /usr/local/bin/kbr

# macOS (Intel)
curl -LO https://github.com/libi/ko-browser/releases/latest/download/ko-browser-darwin-amd64.tar.gz
tar xzf ko-browser-darwin-amd64.tar.gz
mv ko-browser-darwin-amd64 /usr/local/bin/kbr

# Linux (amd64)
curl -LO https://github.com/libi/ko-browser/releases/latest/download/ko-browser-linux-amd64.tar.gz
tar xzf ko-browser-linux-amd64.tar.gz
mv ko-browser-linux-amd64 /usr/local/bin/kbr
```

### 源码编译

```bash
go install github.com/libi/ko-browser@latest
mv $(go env GOPATH)/bin/ko-browser $(go env GOPATH)/bin/kbr  # 可选重命名
```

### 安装浏览器

```bash
kbr install              # 检查并下载 Chromium
kbr install --with-deps  # 同时安装系统依赖（Linux）
```

---

## 🚀 快速开始

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
├── browser/     ★ 公开包 — 核心浏览器 API（可 go get 导入）
├── axtree/      ★ 公开包 — AX Tree 提取、过滤、格式化
├── selector/    ★ 公开包 — 元素选择器解析（ID/CSS/XPath）
├── ocr/         ★ 公开包 — 可选 Tesseract OCR 引擎
├── cmd/           CLI 层 — cobra 命令定义
└── internal/      内部包 — 守护进程、会话管理（仅 CLI 使用）
```

`browser/`、`axtree/`、`selector/`、`ocr/` 均为公开包，可通过 `go get` 导入。`internal/` 仅供 CLI 守护进程使用。

---

## 🤝 参与贡献

欢迎提交 Pull Request！

```bash
git clone https://github.com/libi/ko-browser.git
cd ko-browser
go build -o kbr .
go test ./tests/ -v -timeout 180s
```

---

## 📄 许可证

MIT
