package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/libi/ko-browser/browser"
)

type Client struct {
	opts Options
}

func NewClient(opts Options) *Client {
	return &Client{opts: opts.normalized()}
}

// SessionEntry represents an active session found on the system.
type SessionEntry struct {
	Name   string `json:"name"`
	Socket string `json:"socket"`
}

// ListSessions scans the temp directory for active ko-browser session sockets.
func ListSessions() ([]SessionEntry, error) {
	pattern := filepath.Join(os.TempDir(), "ko-browser-*.sock")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var sessions []SessionEntry
	for _, m := range matches {
		base := filepath.Base(m)
		name := strings.TrimPrefix(base, "ko-browser-")
		name = strings.TrimSuffix(name, ".sock")

		// Check if socket is alive by attempting a connection
		conn, connErr := net.DialTimeout("unix", m, 200*time.Millisecond)
		if connErr != nil {
			// Socket exists but daemon is dead; clean up
			_ = os.Remove(m)
			continue
		}
		conn.Close()

		sessions = append(sessions, SessionEntry{
			Name:   name,
			Socket: m,
		})
	}
	return sessions, nil
}

func (c *Client) Open(url string) error {
	_, err := c.call(Request{Command: "open", URL: url}, true)
	return err
}

func (c *Client) Snapshot(opts browser.SnapshotOptions) (Response, error) {
	return c.call(Request{Command: "snapshot", SnapshotOptions: opts}, true)
}

func (c *Client) Click(id int) error {
	_, err := c.call(Request{Command: "click", ID: id}, true)
	return err
}

func (c *Client) Type(id int, text string) error {
	_, err := c.call(Request{Command: "type", ID: id, Text: text}, true)
	return err
}

func (c *Client) Fill(id int, text string) error {
	_, err := c.call(Request{Command: "fill", ID: id, Text: text}, true)
	return err
}

func (c *Client) Press(key string) error {
	_, err := c.call(Request{Command: "press", Key: key}, true)
	return err
}

func (c *Client) KeyboardType(text string) error {
	_, err := c.call(Request{Command: "keyboard.type", Text: text}, true)
	return err
}

func (c *Client) Hover(id int) error {
	_, err := c.call(Request{Command: "hover", ID: id}, true)
	return err
}

func (c *Client) Focus(id int) error {
	_, err := c.call(Request{Command: "focus", ID: id}, true)
	return err
}

func (c *Client) Check(id int) error {
	_, err := c.call(Request{Command: "check", ID: id}, true)
	return err
}

func (c *Client) Uncheck(id int) error {
	_, err := c.call(Request{Command: "uncheck", ID: id}, true)
	return err
}

func (c *Client) Select(id int, values ...string) error {
	_, err := c.call(Request{Command: "select", ID: id, Values: values}, true)
	return err
}

func (c *Client) Scroll(direction string, amount int) error {
	_, err := c.call(Request{Command: "scroll", Direction: direction, Amount: amount}, true)
	return err
}

func (c *Client) ScrollIntoView(id int) error {
	_, err := c.call(Request{Command: "scrollintoview", ID: id}, true)
	return err
}

func (c *Client) DblClick(id int) error {
	_, err := c.call(Request{Command: "dblclick", ID: id}, true)
	return err
}

// Phase 2: Navigation

func (c *Client) Back() error {
	_, err := c.call(Request{Command: "back"}, true)
	return err
}

func (c *Client) Forward() error {
	_, err := c.call(Request{Command: "forward"}, true)
	return err
}

func (c *Client) Reload() error {
	_, err := c.call(Request{Command: "reload"}, true)
	return err
}

// Phase 2: Information retrieval

func (c *Client) GetTitle() (string, error) {
	resp, err := c.call(Request{Command: "get.title"}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) GetURL() (string, error) {
	resp, err := c.call(Request{Command: "get.url"}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) GetText(id int) (string, error) {
	resp, err := c.call(Request{Command: "get.text", ID: id}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) GetHTML(id int) (string, error) {
	resp, err := c.call(Request{Command: "get.html", ID: id}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) GetValue(id int) (string, error) {
	resp, err := c.call(Request{Command: "get.value", ID: id}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) GetAttr(id int, name string) (string, error) {
	resp, err := c.call(Request{Command: "get.attr", ID: id, AttrName: name}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) GetCount(cssSelector string) (int, error) {
	resp, err := c.call(Request{Command: "get.count", CSSSelector: cssSelector}, true)
	if err != nil {
		return 0, err
	}
	return resp.IntResult, nil
}

func (c *Client) GetBox(id int) (*browser.BoxResult, error) {
	resp, err := c.call(Request{Command: "get.box", ID: id}, true)
	if err != nil {
		return nil, err
	}
	return resp.Box, nil
}

func (c *Client) GetStyles(id int) (string, error) {
	resp, err := c.call(Request{Command: "get.styles", ID: id}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

// Phase 2: State queries

func (c *Client) IsVisible(id int) (bool, error) {
	resp, err := c.call(Request{Command: "is.visible", ID: id}, true)
	if err != nil {
		return false, err
	}
	if resp.BoolResult != nil {
		return *resp.BoolResult, nil
	}
	return false, nil
}

func (c *Client) IsEnabled(id int) (bool, error) {
	resp, err := c.call(Request{Command: "is.enabled", ID: id}, true)
	if err != nil {
		return false, err
	}
	if resp.BoolResult != nil {
		return *resp.BoolResult, nil
	}
	return false, nil
}

func (c *Client) IsChecked(id int) (bool, error) {
	resp, err := c.call(Request{Command: "is.checked", ID: id}, true)
	if err != nil {
		return false, err
	}
	if resp.BoolResult != nil {
		return *resp.BoolResult, nil
	}
	return false, nil
}

// Phase 2: Wait commands

func (c *Client) Wait(d time.Duration) error {
	_, err := c.call(Request{Command: "wait", WaitDuration: d}, true)
	return err
}

func (c *Client) WaitSelector(cssSelector string) error {
	_, err := c.call(Request{Command: "wait.selector", CSSSelector: cssSelector}, true)
	return err
}

func (c *Client) WaitURL(pattern string) error {
	_, err := c.call(Request{Command: "wait.url", Text: pattern}, true)
	return err
}

func (c *Client) WaitLoad() error {
	_, err := c.call(Request{Command: "wait.load"}, true)
	return err
}

func (c *Client) WaitText(text string) error {
	_, err := c.call(Request{Command: "wait.text", Text: text}, true)
	return err
}

func (c *Client) WaitFunc(expression string) error {
	_, err := c.call(Request{Command: "wait.func", Expression: expression}, true)
	return err
}

// Phase 3: Screenshot/PDF/Eval/Find

func (c *Client) Screenshot(path string, args ScreenshotArgs) error {
	_, err := c.call(Request{Command: "screenshot", FilePath: path, ScreenshotArgs: args}, true)
	return err
}

func (c *Client) PDF(path string, args PDFArgs) error {
	_, err := c.call(Request{Command: "pdf", FilePath: path, PDFArgs: args}, true)
	return err
}

func (c *Client) Eval(expression string) (string, error) {
	resp, err := c.call(Request{Command: "eval", Expression: expression}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindRole(role, name string) (string, error) {
	resp, err := c.call(Request{Command: "find.role", Role: role, Text: name}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindText(text string) (string, error) {
	resp, err := c.call(Request{Command: "find.text", Text: text}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindLabel(label string) (string, error) {
	resp, err := c.call(Request{Command: "find.label", Text: label}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindNth(cssSelector string, n int) (string, error) {
	resp, err := c.call(Request{Command: "find.nth", CSSSelector: cssSelector, N: n}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindLast(cssSelector string) (string, error) {
	resp, err := c.call(Request{Command: "find.last", CSSSelector: cssSelector}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

// Phase 4: Drag

func (c *Client) Drag(srcID, dstID int) error {
	_, err := c.call(Request{Command: "drag", ID: srcID, DstID: dstID}, true)
	return err
}

// Phase 4: Upload

func (c *Client) Upload(id int, files ...string) error {
	_, err := c.call(Request{Command: "upload", ID: id, Files: files}, true)
	return err
}

// Phase 4: Download

func (c *Client) Download(id int, saveDir string) (string, error) {
	resp, err := c.call(Request{Command: "download", ID: id, SaveDir: saveDir}, true)
	if err != nil {
		return "", err
	}
	return resp.FilePath, nil
}

// Phase 4: Mouse operations

func (c *Client) MouseMove(x, y float64) error {
	_, err := c.call(Request{Command: "mouse.move", X: x, Y: y}, true)
	return err
}

func (c *Client) MouseDown(x, y float64, button string) error {
	_, err := c.call(Request{Command: "mouse.down", X: x, Y: y, MouseBtn: button}, true)
	return err
}

func (c *Client) MouseUp(x, y float64, button string) error {
	_, err := c.call(Request{Command: "mouse.up", X: x, Y: y, MouseBtn: button}, true)
	return err
}

func (c *Client) MouseWheel(x, y, deltaX, deltaY float64) error {
	_, err := c.call(Request{Command: "mouse.wheel", X: x, Y: y, DeltaX: deltaX, DeltaY: deltaY}, true)
	return err
}

func (c *Client) MouseClick(x, y float64, button string) error {
	_, err := c.call(Request{Command: "mouse.click", X: x, Y: y, MouseBtn: button}, true)
	return err
}

// Phase 5: Tab management

func (c *Client) TabList() ([]browser.TabInfo, error) {
	resp, err := c.call(Request{Command: "tab.list"}, true)
	if err != nil {
		return nil, err
	}
	return resp.Tabs, nil
}

func (c *Client) TabNew(url string) error {
	_, err := c.call(Request{Command: "tab.new", URL: url}, true)
	return err
}

func (c *Client) TabClose(index int) error {
	_, err := c.call(Request{Command: "tab.close", TabIndex: index}, true)
	return err
}

func (c *Client) TabSwitch(index int) error {
	_, err := c.call(Request{Command: "tab.switch", TabIndex: index}, true)
	return err
}

// Phase 5: Network

func (c *Client) NetworkRoute(pattern, action string) error {
	_, err := c.call(Request{Command: "network.route", Pattern: pattern, RouteAction: action}, true)
	return err
}

func (c *Client) NetworkUnroute(pattern string) error {
	_, err := c.call(Request{Command: "network.unroute", Pattern: pattern}, true)
	return err
}

func (c *Client) NetworkRequests() ([]browser.NetworkRequest, error) {
	resp, err := c.call(Request{Command: "network.requests"}, true)
	if err != nil {
		return nil, err
	}
	return resp.Requests, nil
}

func (c *Client) NetworkStartLogging() error {
	_, err := c.call(Request{Command: "network.start-logging"}, true)
	return err
}

func (c *Client) NetworkClearRequests() error {
	_, err := c.call(Request{Command: "network.clear-requests"}, true)
	return err
}

// Phase 5: Cookies

func (c *Client) CookiesGet() ([]browser.CookieInfo, error) {
	resp, err := c.call(Request{Command: "cookies.get"}, true)
	if err != nil {
		return nil, err
	}
	return resp.Cookies, nil
}

func (c *Client) CookieSet(cookie browser.CookieInfo) error {
	_, err := c.call(Request{Command: "cookies.set", Cookie: cookie}, true)
	return err
}

func (c *Client) CookieDelete(name string) error {
	_, err := c.call(Request{Command: "cookies.delete", Text: name}, true)
	return err
}

func (c *Client) CookiesClear() error {
	_, err := c.call(Request{Command: "cookies.clear"}, true)
	return err
}

// Phase 5: Storage

func (c *Client) StorageGet(storageType, key string) (string, error) {
	resp, err := c.call(Request{Command: "storage.get", StorageType: storageType, StorageKey: key}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) StorageSet(storageType, key, value string) error {
	_, err := c.call(Request{Command: "storage.set", StorageType: storageType, StorageKey: key, StorageVal: value}, true)
	return err
}

func (c *Client) StorageDelete(storageType, key string) error {
	_, err := c.call(Request{Command: "storage.delete", StorageType: storageType, StorageKey: key}, true)
	return err
}

func (c *Client) StorageClear(storageType string) error {
	_, err := c.call(Request{Command: "storage.clear", StorageType: storageType}, true)
	return err
}

func (c *Client) StorageGetAll(storageType string) (map[string]string, error) {
	resp, err := c.call(Request{Command: "storage.getall", StorageType: storageType}, true)
	if err != nil {
		return nil, err
	}
	return resp.StorageItems, nil
}

// Phase 6: Settings

func (c *Client) SetViewport(width, height int) error {
	_, err := c.call(Request{Command: "set.viewport", Width: width, Height: height}, true)
	return err
}

func (c *Client) SetDevice(name string) error {
	_, err := c.call(Request{Command: "set.device", DeviceName: name}, true)
	return err
}

func (c *Client) SetGeo(lat, lon float64) error {
	_, err := c.call(Request{Command: "set.geo", Lat: lat, Lon: lon}, true)
	return err
}

func (c *Client) ClearGeo() error {
	_, err := c.call(Request{Command: "clear.geo"}, true)
	return err
}

func (c *Client) SetOffline(offline bool) error {
	_, err := c.call(Request{Command: "set.offline", Offline: &offline}, true)
	return err
}

func (c *Client) SetHeaders(headers map[string]string) error {
	_, err := c.call(Request{Command: "set.headers", Headers: headers}, true)
	return err
}

func (c *Client) SetCredentials(user, pass string) error {
	_, err := c.call(Request{Command: "set.credentials", User: user, Pass: pass}, true)
	return err
}

func (c *Client) SetMedia(features []browser.MediaFeature) error {
	_, err := c.call(Request{Command: "set.media", MediaFeatures: features}, true)
	return err
}

func (c *Client) SetColorScheme(scheme string) error {
	_, err := c.call(Request{Command: "set.colorscheme", ColorScheme: scheme}, true)
	return err
}

// Phase 6: Console/Debug

func (c *Client) ConsoleStart() error {
	_, err := c.call(Request{Command: "console.start"}, true)
	return err
}

func (c *Client) ConsoleMessages(level string) ([]browser.ConsoleMessage, error) {
	resp, err := c.call(Request{Command: "console.messages", ConsoleLevel: level}, true)
	if err != nil {
		return nil, err
	}
	return resp.ConsoleMessages, nil
}

func (c *Client) ConsoleClear() error {
	_, err := c.call(Request{Command: "console.clear"}, true)
	return err
}

func (c *Client) PageErrors() ([]browser.PageError, error) {
	resp, err := c.call(Request{Command: "page.errors"}, true)
	if err != nil {
		return nil, err
	}
	return resp.PageErrors, nil
}

func (c *Client) PageErrorsClear() error {
	_, err := c.call(Request{Command: "page.errors.clear"}, true)
	return err
}

func (c *Client) Highlight(id int) error {
	_, err := c.call(Request{Command: "highlight", ID: id}, true)
	return err
}

func (c *Client) OpenDevTools() error {
	_, err := c.call(Request{Command: "devtools"}, true)
	return err
}

// Phase 6: Clipboard

func (c *Client) ClipboardRead() (string, error) {
	resp, err := c.call(Request{Command: "clipboard.read"}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) ClipboardWrite(text string) error {
	_, err := c.call(Request{Command: "clipboard.write", ClipboardText: text}, true)
	return err
}

// Phase 7: Diff

func (c *Client) DiffSnapshot(baselineFile string, snapOpts browser.SnapshotOptions) (Response, error) {
	return c.call(Request{Command: "diff.snapshot", BaselineFile: baselineFile, SnapshotOptions: snapOpts}, true)
}

func (c *Client) DiffScreenshot(baselineFile, outputFile string, threshold float64, fullPage bool) (Response, error) {
	return c.call(Request{
		Command:        "diff.screenshot",
		BaselineFile:   baselineFile,
		OutputFile:     outputFile,
		Threshold:      threshold,
		ScreenshotArgs: ScreenshotArgs{FullPage: fullPage},
	}, true)
}

func (c *Client) DiffURL(url1, url2 string, includeScreenshot, fullPage bool, threshold float64, snapOpts browser.SnapshotOptions) (Response, error) {
	return c.call(Request{
		Command:           "diff.url",
		URL:               url1,
		URL2:              url2,
		IncludeScreenshot: includeScreenshot,
		ScreenshotArgs:    ScreenshotArgs{FullPage: fullPage},
		Threshold:         threshold,
		SnapshotOptions:   snapOpts,
	}, true)
}

// Phase 7: Trace

func (c *Client) TraceStart(categories string) error {
	_, err := c.call(Request{Command: "trace.start", Categories: categories}, true)
	return err
}

func (c *Client) TraceStop(outputFile string) error {
	_, err := c.call(Request{Command: "trace.stop", OutputFile: outputFile}, true)
	return err
}

// Phase 7: Profiler

func (c *Client) ProfilerStart() error {
	_, err := c.call(Request{Command: "profiler.start"}, true)
	return err
}

func (c *Client) ProfilerStop(outputFile string) error {
	_, err := c.call(Request{Command: "profiler.stop", OutputFile: outputFile}, true)
	return err
}

// Phase 7: Record

func (c *Client) RecordStart(outputPath string) error {
	_, err := c.call(Request{Command: "record.start", OutputFile: outputPath}, true)
	return err
}

func (c *Client) RecordStop() (int, error) {
	resp, err := c.call(Request{Command: "record.stop"}, true)
	if err != nil {
		return 0, err
	}
	return resp.IntResult, nil
}

// Phase 7: State

func (c *Client) ExportState(path string) error {
	_, err := c.call(Request{Command: "state.export", StatePath: path}, true)
	return err
}

func (c *Client) ImportState(path string) error {
	_, err := c.call(Request{Command: "state.import", StatePath: path}, true)
	return err
}

// Phase 8: Missing commands

func (c *Client) KeyboardInsertText(text string) error {
	_, err := c.call(Request{Command: "keyboard.inserttext", Text: text}, true)
	return err
}

func (c *Client) ClickNewTab(id int) error {
	_, err := c.call(Request{Command: "click.newtab", ID: id}, true)
	return err
}

func (c *Client) ClipboardCopy() error {
	_, err := c.call(Request{Command: "clipboard.copy"}, true)
	return err
}

func (c *Client) ClipboardPaste() error {
	_, err := c.call(Request{Command: "clipboard.paste"}, true)
	return err
}

func (c *Client) FindPlaceholder(text string, exact bool) (string, error) {
	resp, err := c.call(Request{Command: "find.placeholder", Text: text, Exact: exact}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindAlt(text string, exact bool) (string, error) {
	resp, err := c.call(Request{Command: "find.alt", Text: text, Exact: exact}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindTitle(text string, exact bool) (string, error) {
	resp, err := c.call(Request{Command: "find.title", Text: text, Exact: exact}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindTestID(testID string) (string, error) {
	resp, err := c.call(Request{Command: "find.testid", Text: testID}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindRoleExact(role, name string, exact bool) (string, error) {
	resp, err := c.call(Request{Command: "find.role", Role: role, Text: name, Exact: exact}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindTextExact(text string, exact bool) (string, error) {
	resp, err := c.call(Request{Command: "find.text", Text: text, Exact: exact}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) FindLabelExact(label string, exact bool) (string, error) {
	resp, err := c.call(Request{Command: "find.label", Text: label, Exact: exact}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) WaitHidden(cssSelector string) error {
	_, err := c.call(Request{Command: "wait.hidden", CSSSelector: cssSelector}, true)
	return err
}

func (c *Client) WaitDownload(saveDir string) (string, error) {
	resp, err := c.call(Request{Command: "wait.download", SaveDir: saveDir}, true)
	if err != nil {
		return "", err
	}
	return resp.FilePath, nil
}

func (c *Client) GetCDPURL() (string, error) {
	resp, err := c.call(Request{Command: "get.cdp-url"}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) SetViewportWithScale(width, height int, scale float64) error {
	_, err := c.call(Request{Command: "set.viewport", Width: width, Height: height, Scale: scale}, true)
	return err
}

func (c *Client) ScreenshotAnnotated(path string, args ScreenshotArgs) error {
	args.Annotate = true
	_, err := c.call(Request{Command: "screenshot", FilePath: path, ScreenshotArgs: args}, true)
	return err
}

func (c *Client) SessionInfo() (string, error) {
	resp, err := c.call(Request{Command: "session.info"}, true)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func (c *Client) Confirm(id string) error {
	_, err := c.call(Request{Command: "confirm", ConfirmID: id}, true)
	return err
}

func (c *Client) Deny(id string) error {
	_, err := c.call(Request{Command: "deny", ConfirmID: id}, true)
	return err
}

func (c *Client) Close() error {
	_, err := c.call(Request{Command: "close"}, false)
	if err != nil && isConnectionError(err) {
		return nil
	}
	return err
}

func (c *Client) call(req Request, autoStart bool) (Response, error) {
	resp, err := c.sendOnce(req)
	if err == nil {
		return resp, nil
	}
	if !autoStart || !isConnectionError(err) {
		return Response{}, err
	}
	if err := c.startDaemon(); err != nil {
		return Response{}, err
	}
	return c.sendOnce(req)
}

func (c *Client) sendOnce(req Request) (Response, error) {
	conn, err := net.DialTimeout("unix", socketPath(c.opts.Name), 500*time.Millisecond)
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(c.opts.Timeout)); err != nil {
		return Response{}, err
	}

	if err := json.NewEncoder(conn).Encode(&req); err != nil {
		return Response{}, err
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return Response{}, err
	}
	if !resp.OK {
		return Response{}, errors.New(resp.Error)
	}
	return resp, nil
}

func (c *Client) startDaemon() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{"_daemon", "--session", c.opts.Name, "--timeout", c.opts.Timeout.String()}
	if c.opts.Headed {
		args = append(args, "--headed")
	}
	if c.opts.Debug {
		args = append(args, "--debug")
	}
	if c.opts.Profile != "" {
		args = append(args, "--profile", c.opts.Profile)
	}
	if c.opts.StatePath != "" {
		args = append(args, "--state", c.opts.StatePath)
	}
	// Phase 8.3: Global options
	if c.opts.UserAgent != "" {
		args = append(args, "--user-agent", c.opts.UserAgent)
	}
	if c.opts.Proxy != "" {
		args = append(args, "--proxy", c.opts.Proxy)
	}
	if c.opts.ProxyBypass != "" {
		args = append(args, "--proxy-bypass", c.opts.ProxyBypass)
	}
	if c.opts.IgnoreHTTPSErrors {
		args = append(args, "--ignore-https-errors")
	}
	if c.opts.AllowFileAccess {
		args = append(args, "--allow-file-access")
	}
	for _, ext := range c.opts.Extensions {
		args = append(args, "--extension", ext)
	}
	for _, a := range c.opts.ExtraArgs {
		args = append(args, "--args", a)
	}
	if c.opts.DownloadPath != "" {
		args = append(args, "--download-path", c.opts.DownloadPath)
	}
	if c.opts.ScreenshotDir != "" {
		args = append(args, "--screenshot-dir", c.opts.ScreenshotDir)
	}
	if c.opts.ScreenshotFormat != "" {
		args = append(args, "--screenshot-format", c.opts.ScreenshotFormat)
	}

	cmd := exec.Command(executable, args...)
	if c.opts.Debug {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	return c.waitForDaemon()
}

func (c *Client) waitForDaemon() error {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", socketPath(c.opts.Name), 300*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("daemon did not become ready in time")
}

func isConnectionError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) || errors.Is(err, os.ErrNotExist)
}
