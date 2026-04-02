<p align="center">
  <h1 align="center">kbr (ko-browser)</h1>
  <p align="center">A simple, fast, token-efficient browser for AI agents — CLI + Go Library.</p>
</p>

<p align="center">
  <a href="README-CN.md">中文文档</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#agent-notes">Agent Notes</a> •
  <a href="#-commands">Commands</a> •
  <a href="#-library-api">Library</a> •
  <a href="docs/snapshot-format-en.md">Snapshot Format Spec</a>
</p>

<p align="center">
  <a href="https://github.com/libi/ko-browser/releases"><img src="https://img.shields.io/github/v/release/libi/ko-browser?style=flat-square" alt="Release"></a>
  <a href="https://github.com/libi/ko-browser/actions"><img src="https://img.shields.io/github/actions/workflow/status/libi/ko-browser/ci.yml?style=flat-square" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/libi/ko-browser"><img src="https://pkg.go.dev/badge/github.com/libi/ko-browser.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/libi/ko-browser"><img src="https://goreportcard.com/badge/github.com/libi/ko-browser?style=flat-square" alt="Go Report Card"></a>
</p>

---

**ko-browser** is a browser automation tool built in Go for AI agents. It exposes the same core model in two forms:

- a **CLI** for shell-driven agent workflows
- a **Go library** for embedding browser control into agent runtimes and tools

Its custom accessibility-tree snapshot format reduces prompt footprint by **46%+** compared with more verbose alternatives.

### ✨ Key Features

- 🚀 **Single binary** — no Node.js, no Playwright runtime needed
- 🤖 **AI-optimized snapshot format** — `id: role "name" states` saves 46%+ tokens
- 📦 **Dual-use** — works as a CLI tool AND a Go library (`go get`)
- ⚡ **Fast startup** — ~50ms (Go binary) vs ~500ms (Node.js-based tools)
- 🔢 **Simple element references** — `click 5` 
- 🔍 **Optional OCR** — Tesseract integration via `-tags=ocr` build flag for image-heavy pages
- 🌐 **~86 commands** — broad coverage for browser automation workflows

### Snapshot Format Comparison

```
┌─ kbr (46% fewer tokens) ──────────┐  ┌─ verbose snapshot output ──────────┐
│ Page: "Example"                    │  │ - document "Example"               │
│                                    │  │   - navigation "main":             │
│ 1: link "Home"                     │  │     - link "Home" [ref=@e1]        │
│ 2: link "About"                    │  │     - link "About" [ref=@e2]       │
│ 3: textbox "Search" focused        │  │   - search:                        │
│ 4: button "Go"                     │  │     - textbox "Search" [ref=@e3]   │
│                                    │  │     - button "Go" [ref=@e4]        │
└────────────────────────────────────┘  └────────────────────────────────────┘
```
> 📖 Read the full [Snapshot Format Specification](docs/snapshot-format-en.md) for detailed design decisions, BNF grammar, and examples. ([中文版](docs/snapshot-format.md))
---

## 📦 Installation

### Choose an install path

| Use case | Recommended path | Notes |
|---------|------------------|-------|
| macOS local usage | Homebrew | Installs an OCR-enabled build and pulls in `tesseract` |
| Linux/macOS manual deployment | GitHub Releases | Download the prebuilt archive and make sure Tesseract runtime libraries are present |
| Go-based integration or custom builds | From source | Best when embedding `kbr` into your own toolchain |

### Homebrew

```bash
brew tap libi/tap
brew install ko-browser
```

> Homebrew installs `kbr` with OCR enabled.
> It pulls in `tesseract` automatically and builds from source.

### Pre-built binaries

Download from [GitHub Releases](https://github.com/libi/ko-browser/releases):

> Release binaries are also built with OCR enabled.
> Install Tesseract first so the runtime OCR libraries are available: `brew install tesseract` on macOS, `apt install libtesseract-dev` on Linux.

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

### From source

```bash
# Install kbr with OCR support (requires Tesseract to be installed)
CGO_ENABLED=1 go install -tags=ocr github.com/libi/ko-browser/cmd/kbr@latest
```

> OCR is required for the published packages.
> This requires Tesseract: `brew install tesseract` (macOS) / `apt install libtesseract-dev` (Linux).

### Install Chrome (if not already installed)

```bash
kbr install              # check & download Chromium
kbr install --with-deps  # also install system dependencies (Linux)
```

### Runtime requirements

- Chrome or Chromium is required. Run `kbr install` if you do not already have a compatible browser.
- OCR-enabled builds expect Tesseract runtime libraries on the host.
- On Linux CI or sandboxed environments, Chrome may require `--no-sandbox`-style fallbacks depending on the runner configuration.

---

## 🚀 Quick Start

### Agent loop

This is the most common shell-driven agent workflow:

```bash
kbr open https://example.com
kbr snapshot
kbr click 5
kbr type 8 "hello world"
kbr press Enter
kbr wait load
kbr get text 12
```

Use `snapshot` to get the current page structure, act on numeric refs, then re-snapshot after the page changes.

### CLI

```bash
# Open a page and take a snapshot
kbr open https://www.google.com
kbr snapshot

# Output:
# Page: "Google"
#
# 1: link "Gmail"
# 2: link "Images"
# 3: textbox "Search" focused
# 4: button "Google Search"
# 5: button "I'm Feeling Lucky"

# Interact with elements by ID
kbr click 3
kbr type 3 "ko-browser github"
kbr press Enter

# Take a screenshot
kbr screenshot result.png

# Close the browser
kbr close
```

### Go Library

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

    b.Open("https://www.google.com")

    snap, _ := b.Snapshot()
    fmt.Println(snap.Text)
    // 1: link "Gmail"
    // 2: link "Images"
    // 3: textbox "Search" focused
    // ...

    b.Click(3)
    b.Type(3, "ko-browser github")
    b.Press("Enter")
}
```

---

## 📖 Commands

### Common flows

| Goal | Commands |
|------|----------|
| Navigate and inspect | `open`, `snapshot`, `get title`, `get url` |
| Interact with forms | `click`, `type`, `fill`, `select`, `check`, `press` |
| Wait for state changes | `wait load`, `wait selector`, `wait text`, `wait url` |
| Debug and verify | `screenshot`, `console messages`, `errors list`, `highlight` |
| Work with sessions | `tab *`, `cookies *`, `storage *`, `state *` |

### Core Interaction

| Command | CLI Usage | Library API |
|---------|----------|-------------|
| **open** | `kbr open <url>` | `b.Open(url)` |
| **click** | `kbr click <id>` | `b.Click(id)` |
| **dblclick** | `kbr dblclick <id>` | `b.DblClick(id)` |
| **type** | `kbr type <id> "text"` | `b.Type(id, text)` |
| **fill** | `kbr fill <id> "text"` | `b.Fill(id, text)` |
| **press** | `kbr press <key>` | `b.Press(key)` |
| **hover** | `kbr hover <id>` | `b.Hover(id)` |
| **focus** | `kbr focus <id>` | `b.Focus(id)` |
| **check** | `kbr check <id>` | `b.Check(id)` |
| **uncheck** | `kbr uncheck <id>` | `b.Uncheck(id)` |
| **select** | `kbr select <id> "val"` | `b.Select(id, vals...)` |
| **scroll** | `kbr scroll down 500` | `b.Scroll("down", 500)` |
| **drag** | `kbr drag <src> <dst>` | `b.Drag(srcID, dstID)` |
| **close** | `kbr close` | `b.Close()` |

### Keyboard

```bash
kbr press Enter
kbr press Control+a
kbr keyboard type "Hello, World!"
kbr keyboard inserttext "pasted content"
```

### Snapshot & Screenshot

```bash
kbr snapshot                     # full accessibility tree
kbr snapshot -i                  # interactive elements only
kbr snapshot -c                  # compact mode
kbr snapshot -d 5                # max depth 5
kbr snapshot -C                  # include cursor elements
kbr snapshot -s "#main"          # scope to CSS selector
kbr snapshot --ocr               # with OCR for images

kbr screenshot out.png           # viewport screenshot
kbr screenshot --full out.png    # full page
kbr screenshot --annotate out.png # annotated with element IDs
```

### Navigation

```bash
kbr back
kbr forward
kbr reload
```

### Get Information

```bash
kbr get title                    # page title
kbr get url                      # current URL
kbr get text <id>                # element inner text
kbr get html <id>                # element innerHTML
kbr get value <id>               # input value
kbr get attr <id> href           # element attribute
kbr get count ".item"            # count matching elements
kbr get box <id>                 # bounding box
kbr get styles <id>              # computed styles
kbr get cdp-url                  # CDP WebSocket URL
```

### Check State

```bash
kbr is visible <id>              # → true/false
kbr is enabled <id>
kbr is checked <id>
```

### Find Elements

```bash
kbr find role button --name Submit
kbr find text "Sign In"
kbr find label "Email"
kbr find placeholder "Search..."
kbr find alt "Logo"
kbr find title "tooltip"
kbr find testid "login-form"
kbr find first ".card"
kbr find last ".card"
kbr find nth 2 ".card"
# Add --exact for exact text matching
kbr find text "Sign In" --exact
```

### Wait

```bash
kbr wait time 2s                 # wait 2 seconds
kbr wait selector "#loading"     # wait for element
kbr wait url "**/dashboard"      # wait for URL match
kbr wait load                    # wait for networkidle
kbr wait text "Welcome"          # wait for text
kbr wait fn "window.ready"       # wait for JS expression
kbr wait hidden "#spinner"       # wait for element to hide
kbr wait download ./file.pdf     # wait for download
```

### Mouse (Low-level)

```bash
kbr mouse move 100 200
kbr mouse click 100 200
kbr mouse down 100 200
kbr mouse up 100 200
kbr mouse wheel 100 200 0 300    # x y deltaX deltaY
```

### Tabs

```bash
kbr tab list
kbr tab new https://example.com
kbr tab switch 2
kbr tab close 1
```

### Network

```bash
kbr network route "**/api/*" --action block
kbr network unroute "**/api/*"
kbr network requests
kbr network start-logging
kbr network clear
```

### Storage & Cookies

```bash
kbr cookies get
kbr cookies set session_id "abc123" --domain example.com
kbr cookies delete session_id
kbr cookies clear

kbr storage get theme --type local
kbr storage set theme "dark" --type local
kbr storage delete theme --type local
kbr storage clear --type local
kbr storage list --type session
```

### Browser Settings

```bash
kbr set viewport 1920 1080
kbr set viewport 1920 1080 2     # with 2x scale
kbr set device "iPhone 12"
kbr set geo 37.7749 -122.4194
kbr set offline true
kbr set headers '{"X-Custom":"value"}'
kbr set credentials admin secret
kbr set media dark
kbr set colorscheme dark
```

### File Operations

```bash
kbr upload <id> ./document.pdf
kbr download <id> ./output/
```

### JavaScript Evaluation

```bash
kbr eval "document.title"
kbr eval -b "ZG9jdW1lbnQudGl0bGU="   # base64 encoded
cat script.js | kbr eval --stdin
```

### Diff

```bash
kbr diff snapshot                              # compare with last snapshot
kbr diff snapshot --baseline before.txt
kbr diff screenshot --baseline before.png
kbr diff url https://v1.example.com https://v2.example.com
```

### Debug & Clipboard

```bash
kbr console messages
kbr console clear
kbr errors list
kbr highlight <id>
kbr inspect                      # open DevTools

kbr clipboard read
kbr clipboard write "text"
kbr clipboard copy               # Ctrl+C
kbr clipboard paste              # Ctrl+V
```

### Trace & Record

```bash
kbr trace start
kbr trace stop ./trace.zip

kbr profiler start
kbr profiler stop ./profile.json

kbr record start ./recording
kbr record stop
```

### Auth

```bash
kbr auth save github --url https://github.com/login --username user --password pass
kbr auth login github
kbr auth list
kbr auth show github
kbr auth delete github
```

### Session Management

```bash
kbr session                      # show current session
kbr session list                 # list all active sessions
kbr --session test open example.com  # use named session
```

### Selector Syntax

kbr supports three selector formats, auto-detected:

| Input | Type | Example |
|-------|------|---------|
| Number | Snapshot ID | `kbr click 5` |
| CSS | CSS Selector | `kbr click "#submit"` |
| XPath | XPath | `kbr click "//button[@type='submit']"` |

---

## 🔧 Global Options

| Flag | Description |
|------|------------|
| `--session <name>` | Isolated session name (default: "default") |
| `--headed` | Show browser window |
| `--json` | JSON output |
| `--timeout <duration>` | Operation timeout (default: 30s) |
| `--profile <path>` | Persistent Chrome user data directory |
| `--state <path>` | Load saved browser state (cookies + localStorage) |
| `--config <path>` | Config file path |
| `--user-agent <ua>` | Custom User-Agent |
| `--proxy <url>` | Proxy server URL |
| `--proxy-bypass <hosts>` | Hosts to bypass proxy |
| `--ignore-https-errors` | Ignore HTTPS certificate errors |
| `--allow-file-access` | Allow file:// URLs |
| `--extension <path>` | Load Chrome extension (repeatable) |
| `--args <args>` | Extra Chrome arguments |
| `--download-path <path>` | Default download directory |
| `--screenshot-dir <path>` | Default screenshot output directory |
| `--screenshot-format <fmt>` | Screenshot format: png, jpeg |
| `--content-boundaries` | Wrap output with boundary markers |
| `--debug` | Debug output |

---

## 📚 Library API

### Installation

```bash
go get github.com/libi/ko-browser@latest
```

### Quick Example

```go
package main

import (
    "fmt"
    "time"

    "github.com/libi/ko-browser/browser"
)

func main() {
    b, _ := browser.New(browser.Options{
        Headless: true,
        Timeout:  30 * time.Second,
    })
    defer b.Close()

    // Navigate
    b.Open("https://www.baidu.com")

    // Snapshot → interact
    snap, _ := b.Snapshot()
    fmt.Println(snap.Text)
    // Page: "百度一下，你就知道"
    //
    // 1: link "新闻"
    // 2: link "hao123"
    // 3: textbox "搜索" focused
    // 4: button "百度一下"

    b.Click(3)
    b.Type(3, "hello world")
    b.Press("Enter")

    // Wait & screenshot
    b.WaitLoad()
    b.Screenshot("result.png")
}
```

### Advanced: Direct AX Tree Access

```go
import (
    "github.com/libi/ko-browser/axtree"
    "github.com/libi/ko-browser/browser"
    "github.com/libi/ko-browser/ocr"
)

b, _ := browser.New(browser.Options{Headless: true})
defer b.Close()
b.Open("https://example.com")

// Low-level AX Tree API
rawNodes, _ := axtree.Extract(b.Context())
tree := axtree.BuildAndFilter(rawNodes)
text := axtree.Format(tree)
idMap := axtree.BuildIDMap(tree)

// Snapshot with OCR
snap, _ := b.Snapshot(browser.SnapshotOptions{
    EnableOCR:    true,
    OCRLanguages: []string{"eng", "chi_sim"},
})
```

### Connect to Existing Browser

```go
// Connect via CDP WebSocket
b, _ := browser.Connect("ws://localhost:9222/devtools/browser/...", browser.Options{})
defer b.Close()

snap, _ := b.Snapshot()
fmt.Println(snap.Text)
```

### Full API Reference

<details>
<summary>Click to expand all Browser methods</summary>

**Navigation**
- `Open(url string) error`
- `Back() error`
- `Forward() error`
- `Reload() error`

**Snapshot**
- `Snapshot(opts ...SnapshotOptions) (*SnapshotResult, error)`

**Interaction**
- `Click(id int) error`
- `ClickNewTab(id int) error`
- `DblClick(id int) error`
- `Type(id int, text string) error`
- `Fill(id int, text string) error`
- `Press(key string) error`
- `KeyboardType(text string) error`
- `KeyboardInsertText(text string) error`
- `Hover(id int) error`
- `Focus(id int) error`
- `Check(id int) error`
- `Uncheck(id int) error`
- `Select(id int, values ...string) error`
- `Scroll(direction string, pixels int) error`
- `ScrollIntoView(id int) error`

**Mouse**
- `MouseMove(x, y float64) error`
- `MouseClick(x, y float64, opts ...MouseOptions) error`
- `MouseDown(x, y float64, opts ...MouseOptions) error`
- `MouseUp(x, y float64, opts ...MouseOptions) error`
- `MouseWheel(x, y, deltaX, deltaY float64) error`
- `Drag(srcID, dstID int) error`
- `DragCoords(srcX, srcY, dstX, dstY float64) error`

**Query**
- `GetTitle() (string, error)`
- `GetURL() (string, error)`
- `GetText(id int) (string, error)`
- `GetHTML(id int) (string, error)`
- `GetValue(id int) (string, error)`
- `GetAttr(id int, name string) (string, error)`
- `GetCount(cssSelector string) (int, error)`
- `GetBox(id int) (*BoxResult, error)`
- `GetStyles(id int) (string, error)`
- `GetCDPURL() (string, error)`

**State**
- `IsVisible(id int) (bool, error)`
- `IsEnabled(id int) (bool, error)`
- `IsChecked(id int) (bool, error)`

**Find**
- `FindRole(role, name string, opts ...FindOption) (*FindResults, error)`
- `FindText(text string, opts ...FindOption) (*FindResults, error)`
- `FindLabel(label string, opts ...FindOption) (*FindResults, error)`
- `FindPlaceholder(text string, opts ...FindOption) (*FindResults, error)`
- `FindAlt(text string, opts ...FindOption) (*FindResults, error)`
- `FindTitle(text string, opts ...FindOption) (*FindResults, error)`
- `FindTestID(testID string) (*FindResults, error)`
- `FindFirst(css string) (*FindResults, error)`
- `FindLast(css string) (*FindResults, error)`
- `FindNth(css string, n int) (*FindResults, error)`

**Wait**
- `Wait(d time.Duration) error`
- `WaitSelector(css string, timeout ...time.Duration) error`
- `WaitURL(pattern string, timeout ...time.Duration) error`
- `WaitLoad(timeout ...time.Duration) error`
- `WaitText(text string, timeout ...time.Duration) error`
- `WaitFunc(expression string, timeout ...time.Duration) error`
- `WaitHidden(css string, timeout ...time.Duration) error`
- `WaitDownload(savePath string, timeout ...time.Duration) (string, error)`

**Screenshot & PDF**
- `Screenshot(path string, opts ...ScreenshotOptions) error`
- `ScreenshotToBytes(opts ...ScreenshotOptions) ([]byte, error)`
- `ScreenshotAnnotated(path string, opts ...ScreenshotOptions) error`
- `PDF(path string, opts ...PDFOptions) error`

**Eval**
- `Eval(expression string) (string, error)`

**File**
- `Upload(id int, files ...string) error`
- `UploadCSS(css string, files ...string) error`
- `Download(id int, saveDir string, opts ...DownloadOptions) (string, error)`

**Tabs**
- `TabList() ([]TabInfo, error)`
- `TabNew(url string) error`
- `TabClose(index int) error`
- `TabSwitch(index int) error`

**Network**
- `NetworkRoute(pattern string, action RouteAction) error`
- `NetworkUnroute(pattern string) error`
- `NetworkRequests() ([]NetworkRequest, error)`
- `NetworkStartLogging() error`
- `NetworkClearRequests()`

**Storage & Cookies**
- `CookiesGet() ([]CookieInfo, error)`
- `CookieSet(cookie CookieInfo) error`
- `CookieDelete(name string) error`
- `CookiesClear() error`
- `StorageGet(storageType, key string) (string, error)`
- `StorageSet(storageType, key, value string) error`
- `StorageDelete(storageType, key string) error`
- `StorageClear(storageType string) error`
- `StorageGetAll(storageType string) (map[string]string, error)`

**Settings**
- `SetViewport(width, height int, scale ...float64) error`
- `SetDevice(name string) error`
- `SetGeo(lat, lon float64) error`
- `ClearGeo() error`
- `SetOffline(offline bool) error`
- `SetHeaders(headers map[string]string) error`
- `SetCredentials(user, pass string) error`
- `SetMedia(features ...MediaFeature) error`
- `SetColorScheme(scheme string) error`

**Debug**
- `ConsoleStart() error`
- `ConsoleMessages() ([]ConsoleMessage, error)`
- `ConsoleMessagesByLevel(level string) ([]ConsoleMessage, error)`
- `ConsoleClear()`
- `PageErrors() ([]PageError, error)`
- `PageErrorsClear()`
- `Highlight(id int) error`
- `OpenDevTools() error`

**Clipboard**
- `ClipboardRead() (string, error)`
- `ClipboardWrite(text string) error`
- `ClipboardCopy() error`
- `ClipboardPaste() error`

**Diff**
- `DiffSnapshot(opts ...DiffSnapshotOptions) (*DiffSnapshotResult, error)`
- `DiffScreenshot(opts DiffScreenshotOptions) (*DiffScreenshotResult, error)`
- `DiffURL(url1, url2 string, opts ...DiffURLOptions) (*DiffURLResult, error)`

**Trace & Record**
- `TraceStart(categories ...string) error`
- `TraceStop(outputPath string) error`
- `ProfilerStart() error`
- `ProfilerStop(outputPath string) error`
- `RecordStart(outputPath string) error`
- `RecordStop() (int, error)`

**State**
- `ExportState(outputPath string) error`
- `ImportState(inputPath string) error`
- `ApplyState(state *BrowserState) error`

</details>

---

## ⚙️ Configuration

kbr loads configuration in this priority (low → high):

1. `~/.ko-browser/config.json` — user-level defaults
2. `./ko-browser.json` — project-level overrides
3. Environment variables (`KO_BROWSER_*`)
4. CLI flags — override everything

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

## 🏗️ Architecture

```
kbr
├── cmd/kbr/      ★ CLI entry point — `go install .../cmd/kbr@latest`
├── browser/      ★ Public Go library — core browser API
├── axtree/       ★ Public — AX Tree extraction, filtering, formatting
├── selector/     ★ Public — element selector parsing (ID/CSS/XPath)
├── ocr/          ★ Public — Tesseract OCR engine (build tag: ocr)
├── cmd/            CLI — cobra command definitions
└── internal/       CLI-only — daemon, session management
```

The `browser/`, `axtree/`, `selector/`, and `ocr/` packages are all public and importable via `go get`. The `internal/` package is only used by the CLI daemon. OCR requires `-tags=ocr` at build time.

---

<a name="agent-notes"></a>

## Agent Notes

If you are an agent reading this README, use `kbr` as an interactive browser with a compact text interface.

- Prefer `kbr snapshot` as the primary page representation. It is the most token-efficient output and gives stable numeric refs.
- Use numeric refs like `kbr click 5`, `kbr type 8 "hello"`, and `kbr get text 12` after taking a fresh snapshot.
- Re-snapshot after navigation or DOM mutations. Element IDs are snapshot-local, not permanent selectors.
- Use `kbr get`, `kbr find`, and `kbr wait` for targeted reads instead of repeatedly taking full screenshots.
- Use OCR only when the page content is image-heavy or inaccessible through the DOM/accessibility tree.
- Prefer CSS/XPath selectors only when you already know the target precisely or when numeric refs are unavailable.
- The published Homebrew package and release binaries are built with OCR enabled, so Tesseract runtime libraries are expected on the host.

### Install as an agent skill

If your agent framework supports local skills, install `kbr` first, then copy the repository skill directory into your agent's skills directory as `ko-browser/`.

```yaml
<agent-skills-dir>/
  ko-browser/
    SKILL.md
```

The source skill lives here:

- [`skills/ko-browser/SKILL.md`](https://github.com/libi/ko-browser/blob/main/skills/ko-browser/SKILL.md)

Example:

```bash
mkdir -p <agent-skills-dir>/ko-browser
cp skills/ko-browser/SKILL.md <agent-skills-dir>/ko-browser/SKILL.md
```

After that, make sure the `kbr` binary is available in the agent runtime `PATH`. For agent hosts that also support executable-tool registration, you can expose `kbr` directly in addition to the skill file.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

```bash
git clone https://github.com/libi/ko-browser.git
cd ko-browser
go build -o kbr ./cmd/kbr/              # without OCR
go build -tags=ocr -o kbr ./cmd/kbr/     # with OCR
go test ./tests/ -v -timeout 180s
```

---

## 📄 License

MIT

---
