# ko-browser Library Usage Examples

This directory contains three examples showing how to use `ko-browser` as a Go library.

## Core Concept

ko-browser uses a **snapshot-driven** interaction model:

1. Call `b.Snapshot()` to get the current page's accessibility tree
2. Each interactive element in the snapshot gets a numeric ID like `[1]`, `[2]`, `[3]`
3. All interactions (Click / Fill / Hover etc.) reference elements by this ID

```
snap.Text output example:
[1] heading "Example Domain"
[2] paragraph "This domain is for use in illustrative examples..."
[3] link "More information..."

-> Call b.Click(3) to click the "More information..." link
```

## Examples

### basic/ - Basic Usage

Core workflow: create browser -> open page -> snapshot -> query/click -> screenshot

```bash
go run ./examples/basic/
```

### form/ - Form Interaction

Form operations: Fill / Type / Check / Select, element finding (FindRole / FindLabel), state queries (IsVisible / IsEnabled)

```bash
go run ./examples/form/
```

### advanced/ - Advanced Features

Tabs, viewport/device emulation, screenshot options, Cookie/Storage, Console capture, network logging, snapshot options

```bash
go run ./examples/advanced/
```

## Quick Reference

### Create a Browser

```go
import "github.com/libi/ko-browser/browser"

// Launch a new Chrome instance
b, err := browser.New(browser.Options{
    Headless: true,
    Timeout:  30 * time.Second,
})
defer b.Close()

// Or connect to an already-running Chrome (CDP)
b, err := browser.Connect("9222", browser.Options{})
```

### Common API

| Category | Methods | Description |
|----------|---------|-------------|
| Navigate | `Open(url)`, `Back()`, `Forward()`, `Reload()` | Page navigation |
| Snapshot | `Snapshot(opts...)` | Get accessibility tree with element IDs |
| Click | `Click(id)`, `DblClick(id)`, `Hover(id)` | Mouse interaction |
| Input | `Fill(id, text)`, `Type(id, text)`, `KeyboardPress(key)` | Keyboard input |
| Form | `Check(id)`, `Uncheck(id)`, `Select(id, values...)` | Form controls |
| Query | `GetTitle()`, `GetURL()`, `GetText(id)`, `GetValue(id)` | Read page/element info |
| Find | `FindRole(role, name)`, `FindText(text)`, `FindLabel(label)` | Search elements in snapshot |
| Wait | `WaitSelector(css)`, `WaitText(text)`, `WaitURL(pattern)` | Wait for conditions |
| State | `IsVisible(id)`, `IsEnabled(id)`, `IsChecked(id)` | Element state |
| Screenshot | `Screenshot(path)`, `ScreenshotToBytes()`, `PDF(path)` | Capture screenshots/PDF |
| JS | `Eval(expression)` | Execute JavaScript |
| Tab | `TabList()`, `TabNew(url)`, `TabSwitch(i)`, `TabClose(i)` | Tab management |
| Storage | `CookiesGet()`, `CookieSet(...)`, `StorageGet/Set(...)` | Cookie & Storage |
| Settings | `SetViewport(w,h)`, `SetDevice(name)`, `SetOffline(bool)` | Browser settings |
| Network | `NetworkStartLogging()`, `NetworkRequests()` | Network request capture |
| Console | `ConsoleStart()`, `ConsoleMessages()`, `PageErrors()` | Console log capture |
