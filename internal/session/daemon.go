package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/libi/ko-browser/browser"
)

func RunDaemon(opts Options) error {
	opts = opts.normalized()
	path := socketPath(opts.Name)
	_ = os.Remove(path)

	listener, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	defer listener.Close()
	defer os.Remove(path)
	_ = os.Chmod(path, 0600)

	confirmStore := NewConfirmStore(opts.ConfirmActions, 0)

	b, err := browser.New(browser.Options{
		Headless:          !opts.Headed,
		Timeout:           opts.Timeout,
		Profile:           opts.Profile,
		StatePath:         opts.StatePath,
		UserAgent:         opts.UserAgent,
		Proxy:             opts.Proxy,
		ProxyBypass:       opts.ProxyBypass,
		IgnoreHTTPSErrors: opts.IgnoreHTTPSErrors,
		AllowFileAccess:   opts.AllowFileAccess,
		Extensions:        opts.Extensions,
		ExtraArgs:         opts.ExtraArgs,
		DownloadPath:      opts.DownloadPath,
		ScreenshotDir:     opts.ScreenshotDir,
		ScreenshotFormat:  opts.ScreenshotFormat,
	})
	if err != nil {
		return err
	}
	defer b.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			continue
		}

		exit, _ := handleConn(conn, b, opts.Name, confirmStore)
		conn.Close()
		if exit {
			return nil
		}
	}
}

func handleConn(conn net.Conn, b *browser.Browser, sessionName string, confirmStore *ConfirmStore) (bool, error) {
	var req Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		_ = json.NewEncoder(conn).Encode(Response{OK: false, Error: err.Error()})
		return false, err
	}

	resp := Response{OK: true}
	var err error

	switch req.Command {
	case "open":
		if req.URL == "" {
			err = fmt.Errorf("url is required")
			break
		}
		err = b.Open(req.URL)
	case "snapshot":
		snapOpts := req.SnapshotOptions
		if req.Cursor {
			snapOpts.Cursor = true
		}
		if req.CSSSelector != "" {
			snapOpts.Selector = req.CSSSelector
		}
		var snap *browser.SnapshotResult
		snap, err = b.Snapshot(snapOpts)
		if err == nil {
			resp.Text = snap.Text
			resp.RawCount = snap.RawCount
			resp.ElementCount = len(snap.IDMap)
		}
	case "click":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		if req.NewTab {
			err = b.ClickNewTab(req.ID)
		} else {
			err = b.Click(req.ID)
		}
	case "type":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.Type(req.ID, req.Text)
	case "fill":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.Fill(req.ID, req.Text)
	case "press":
		err = b.Press(req.Key)
	case "keyboard.type":
		err = b.KeyboardType(req.Text)
	case "hover":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.Hover(req.ID)
	case "focus":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.Focus(req.ID)
	case "check":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.Check(req.ID)
	case "uncheck":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.Uncheck(req.ID)
	case "select":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.Select(req.ID, req.Values...)
	case "scroll":
		err = b.Scroll(req.Direction, req.Amount)
	case "scrollintoview":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.ScrollIntoView(req.ID)
	case "dblclick":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.DblClick(req.ID)
	case "close":
		if encodeErr := json.NewEncoder(conn).Encode(resp); encodeErr != nil {
			return true, encodeErr
		}
		return true, nil

	case "back":
		err = b.Back()
	case "forward":
		err = b.Forward()
	case "reload":
		err = b.Reload()

	case "get.title":
		var text string
		text, err = b.GetTitle()
		if err == nil {
			resp.Text = text
		}
	case "get.url":
		var text string
		text, err = b.GetURL()
		if err == nil {
			resp.Text = text
		}
	case "get.text":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var text string
		text, err = b.GetText(req.ID)
		if err == nil {
			resp.Text = text
		}
	case "get.html":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var text string
		text, err = b.GetHTML(req.ID)
		if err == nil {
			resp.Text = text
		}
	case "get.value":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var text string
		text, err = b.GetValue(req.ID)
		if err == nil {
			resp.Text = text
		}
	case "get.attr":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var text string
		text, err = b.GetAttr(req.ID, req.AttrName)
		if err == nil {
			resp.Text = text
		}
	case "get.count":
		if req.CSSSelector == "" {
			err = fmt.Errorf("css selector is required")
			break
		}
		var n int
		n, err = b.GetCount(req.CSSSelector)
		if err == nil {
			resp.IntResult = n
		}
	case "get.box":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var box *browser.BoxResult
		box, err = b.GetBox(req.ID)
		if err == nil {
			resp.Box = box
		}
	case "get.styles":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var text string
		text, err = b.GetStyles(req.ID)
		if err == nil {
			resp.Text = text
		}

	case "is.visible":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var v bool
		v, err = b.IsVisible(req.ID)
		if err == nil {
			resp.BoolResult = &v
		}
	case "is.enabled":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var v bool
		v, err = b.IsEnabled(req.ID)
		if err == nil {
			resp.BoolResult = &v
		}
	case "is.checked":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		var v bool
		v, err = b.IsChecked(req.ID)
		if err == nil {
			resp.BoolResult = &v
		}

	case "wait":
		err = b.Wait(req.WaitDuration)
	case "wait.selector":
		if req.CSSSelector == "" {
			err = fmt.Errorf("css selector is required")
			break
		}
		err = b.WaitSelector(req.CSSSelector)
	case "wait.url":
		if req.Text == "" {
			err = fmt.Errorf("url pattern is required")
			break
		}
		err = b.WaitURL(req.Text)
	case "wait.load":
		err = b.WaitLoad()
	case "wait.text":
		if req.Text == "" {
			err = fmt.Errorf("text is required")
			break
		}
		err = b.WaitText(req.Text)
	case "wait.func":
		if req.Expression == "" {
			err = fmt.Errorf("expression is required")
			break
		}
		err = b.WaitFunc(req.Expression)

	case "screenshot":
		if req.FilePath == "" {
			err = fmt.Errorf("file path is required")
			break
		}
		if req.ScreenshotArgs.Annotate {
			err = b.ScreenshotAnnotated(req.FilePath, browser.ScreenshotOptions{
				FullPage:  req.ScreenshotArgs.FullPage,
				Quality:   req.ScreenshotArgs.Quality,
				ElementID: req.ScreenshotArgs.ElementID,
			})
		} else {
			err = b.Screenshot(req.FilePath, browser.ScreenshotOptions{
				FullPage:  req.ScreenshotArgs.FullPage,
				Quality:   req.ScreenshotArgs.Quality,
				ElementID: req.ScreenshotArgs.ElementID,
			})
		}

	case "pdf":
		if req.FilePath == "" {
			err = fmt.Errorf("file path is required")
			break
		}
		err = b.PDF(req.FilePath, browser.PDFOptions{
			Landscape: req.PDFArgs.Landscape,
			PrintBG:   req.PDFArgs.PrintBG,
		})

	case "eval":
		if req.Expression == "" {
			err = fmt.Errorf("expression is required")
			break
		}
		var result string
		result, err = b.Eval(req.Expression)
		if err == nil {
			resp.Text = result
		}

	case "find.role":
		var result *browser.FindResults
		var findOpts []browser.FindOption
		if req.Exact {
			findOpts = append(findOpts, browser.WithExact())
		}
		result, err = b.FindRole(req.Role, req.Text, findOpts...)
		if err == nil {
			resp.Text = result.Text
		}
	case "find.text":
		if req.Text == "" {
			err = fmt.Errorf("text is required")
			break
		}
		var result *browser.FindResults
		var findOpts []browser.FindOption
		if req.Exact {
			findOpts = append(findOpts, browser.WithExact())
		}
		result, err = b.FindText(req.Text, findOpts...)
		if err == nil {
			resp.Text = result.Text
		}
	case "find.label":
		if req.Text == "" {
			err = fmt.Errorf("label is required")
			break
		}
		var result *browser.FindResults
		var findOpts []browser.FindOption
		if req.Exact {
			findOpts = append(findOpts, browser.WithExact())
		}
		result, err = b.FindLabel(req.Text, findOpts...)
		if err == nil {
			resp.Text = result.Text
		}
	case "find.nth":
		if req.CSSSelector == "" {
			err = fmt.Errorf("css selector is required")
			break
		}
		var result *browser.FindResults
		result, err = b.FindNth(req.CSSSelector, req.N)
		if err == nil {
			resp.Text = result.Text
		}
	case "find.last":
		if req.CSSSelector == "" {
			err = fmt.Errorf("css selector is required")
			break
		}
		var result *browser.FindResults
		result, err = b.FindLast(req.CSSSelector)
		if err == nil {
			resp.Text = result.Text
		}

	case "drag":
		if req.ID <= 0 {
			err = fmt.Errorf("source id must be positive")
			break
		}
		if req.DstID <= 0 {
			err = fmt.Errorf("destination id must be positive")
			break
		}
		err = b.Drag(req.ID, req.DstID)

	case "upload":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		if len(req.Files) == 0 {
			err = fmt.Errorf("at least one file is required")
			break
		}
		err = b.Upload(req.ID, req.Files...)

	case "download":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		if req.SaveDir == "" {
			err = fmt.Errorf("save directory is required")
			break
		}
		var path string
		path, err = b.Download(req.ID, req.SaveDir)
		if err == nil {
			resp.FilePath = path
		}

	case "mouse.move":
		err = b.MouseMove(req.X, req.Y)
	case "mouse.down":
		opts := browser.MouseOptions{}
		if req.MouseBtn != "" {
			opts.Button = browser.MouseButton(req.MouseBtn)
		}
		err = b.MouseDown(req.X, req.Y, opts)
	case "mouse.up":
		opts := browser.MouseOptions{}
		if req.MouseBtn != "" {
			opts.Button = browser.MouseButton(req.MouseBtn)
		}
		err = b.MouseUp(req.X, req.Y, opts)
	case "mouse.wheel":
		err = b.MouseWheel(req.X, req.Y, req.DeltaX, req.DeltaY)
	case "mouse.click":
		opts := browser.MouseOptions{}
		if req.MouseBtn != "" {
			opts.Button = browser.MouseButton(req.MouseBtn)
		}
		err = b.MouseClick(req.X, req.Y, opts)

	case "tab.list":
		var tabs []browser.TabInfo
		tabs, err = b.TabList()
		if err == nil {
			resp.Tabs = tabs
		}
	case "tab.new":
		err = b.TabNew(req.URL)
	case "tab.close":
		err = b.TabClose(req.TabIndex)
	case "tab.switch":
		err = b.TabSwitch(req.TabIndex)

	case "network.route":
		action := browser.RouteBlock
		if req.RouteAction == "continue" {
			action = browser.RouteContinue
		}
		err = b.NetworkRoute(req.Pattern, action)
	case "network.unroute":
		err = b.NetworkUnroute(req.Pattern)
	case "network.requests":
		var reqs []browser.NetworkRequest
		reqs, err = b.NetworkRequests()
		if err == nil {
			resp.Requests = reqs
		}
	case "network.start-logging":
		err = b.NetworkStartLogging()
	case "network.clear-requests":
		b.NetworkClearRequests()

	case "cookies.get":
		var cookies []browser.CookieInfo
		cookies, err = b.CookiesGet()
		if err == nil {
			resp.Cookies = cookies
		}
	case "cookies.set":
		err = b.CookieSet(req.Cookie)
	case "cookies.delete":
		err = b.CookieDelete(req.Text)
	case "cookies.clear":
		err = b.CookiesClear()

	case "storage.get":
		var val string
		val, err = b.StorageGet(req.StorageType, req.StorageKey)
		if err == nil {
			resp.Text = val
		}
	case "storage.set":
		err = b.StorageSet(req.StorageType, req.StorageKey, req.StorageVal)
	case "storage.delete":
		err = b.StorageDelete(req.StorageType, req.StorageKey)
	case "storage.clear":
		err = b.StorageClear(req.StorageType)
	case "storage.getall":
		var items map[string]string
		items, err = b.StorageGetAll(req.StorageType)
		if err == nil {
			resp.StorageItems = items
		}

	case "set.viewport":
		if req.Width <= 0 || req.Height <= 0 {
			err = fmt.Errorf("width and height must be positive")
			break
		}
		if req.Scale > 0 {
			err = b.SetViewport(req.Width, req.Height, req.Scale)
		} else {
			err = b.SetViewport(req.Width, req.Height)
		}
	case "set.device":
		if req.DeviceName == "" {
			err = fmt.Errorf("device name is required")
			break
		}
		err = b.SetDevice(req.DeviceName)
	case "set.geo":
		err = b.SetGeo(req.Lat, req.Lon)
	case "clear.geo":
		err = b.ClearGeo()
	case "set.offline":
		if req.Offline == nil {
			err = fmt.Errorf("offline flag is required")
			break
		}
		err = b.SetOffline(*req.Offline)
	case "set.headers":
		if req.Headers == nil {
			err = fmt.Errorf("headers map is required")
			break
		}
		err = b.SetHeaders(req.Headers)
	case "set.credentials":
		err = b.SetCredentials(req.User, req.Pass)
	case "set.media":
		err = b.SetMedia(req.MediaFeatures...)
	case "set.colorscheme":
		if req.ColorScheme == "" {
			err = fmt.Errorf("color scheme is required")
			break
		}
		err = b.SetColorScheme(req.ColorScheme)

	case "console.start":
		err = b.ConsoleStart()
	case "console.messages":
		var msgs []browser.ConsoleMessage
		if req.ConsoleLevel != "" {
			msgs, err = b.ConsoleMessagesByLevel(req.ConsoleLevel)
		} else {
			msgs, err = b.ConsoleMessages()
		}
		if err == nil {
			resp.ConsoleMessages = msgs
		}
	case "console.clear":
		b.ConsoleClear()
	case "page.errors":
		var errs []browser.PageError
		errs, err = b.PageErrors()
		if err == nil {
			resp.PageErrors = errs
		}
	case "page.errors.clear":
		b.PageErrorsClear()
	case "highlight":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.Highlight(req.ID)
	case "devtools":
		err = b.OpenDevTools()

	case "clipboard.read":
		var text string
		text, err = b.ClipboardRead()
		if err == nil {
			resp.Text = text
		}
	case "clipboard.write":
		err = b.ClipboardWrite(req.ClipboardText)

	case "diff.snapshot":
		var result *browser.DiffSnapshotResult
		result, err = b.DiffSnapshot(browser.DiffSnapshotOptions{
			BaselineFile:    req.BaselineFile,
			SnapshotOptions: req.SnapshotOptions,
		})
		if err == nil {
			resp.Text = result.Text
			resp.DiffResult = result
		}
	case "diff.screenshot":
		if req.BaselineFile == "" {
			err = fmt.Errorf("baseline file is required")
			break
		}
		var result *browser.DiffScreenshotResult
		result, err = b.DiffScreenshot(browser.DiffScreenshotOptions{
			BaselineFile: req.BaselineFile,
			OutputPath:   req.OutputFile,
			Threshold:    req.Threshold,
			FullPage:     req.ScreenshotArgs.FullPage,
			ElementID:    req.ScreenshotArgs.ElementID,
		})
		if err == nil {
			resp.ScreenshotDiff = result
		}
	case "diff.url":
		if req.URL == "" || req.URL2 == "" {
			err = fmt.Errorf("two URLs are required")
			break
		}
		var result *browser.DiffURLResult
		result, err = b.DiffURL(req.URL, req.URL2, browser.DiffURLOptions{
			IncludeScreenshot: req.IncludeScreenshot,
			FullPage:          req.ScreenshotArgs.FullPage,
			SnapshotOptions:   req.SnapshotOptions,
			Threshold:         req.Threshold,
		})
		if err == nil {
			resp.DiffURLResult = result
			if result.SnapshotDiff != nil {
				resp.Text = result.SnapshotDiff.Text
			}
		}

	case "trace.start":
		if req.Categories != "" {
			err = b.TraceStart(req.Categories)
		} else {
			err = b.TraceStart()
		}
	case "trace.stop":
		err = b.TraceStop(req.OutputFile)

	case "profiler.start":
		err = b.ProfilerStart()
	case "profiler.stop":
		err = b.ProfilerStop(req.OutputFile)

	case "record.start":
		err = b.RecordStart(req.OutputFile)
	case "record.stop":
		var frameCount int
		frameCount, err = b.RecordStop()
		if err == nil {
			resp.IntResult = frameCount
		}

	case "state.export":
		if req.StatePath == "" {
			err = fmt.Errorf("state file path is required")
			break
		}
		err = b.ExportState(req.StatePath)
	case "state.import":
		if req.StatePath == "" {
			err = fmt.Errorf("state file path is required")
			break
		}
		err = b.ImportState(req.StatePath)

	case "keyboard.inserttext":
		if req.Text == "" {
			err = fmt.Errorf("text is required")
			break
		}
		err = b.KeyboardInsertText(req.Text)
	case "click.newtab":
		if req.ID <= 0 {
			err = fmt.Errorf("id must be positive")
			break
		}
		err = b.ClickNewTab(req.ID)
	case "clipboard.copy":
		err = b.ClipboardCopy()
	case "clipboard.paste":
		err = b.ClipboardPaste()
	case "find.placeholder":
		if req.Text == "" {
			err = fmt.Errorf("text is required")
			break
		}
		var result *browser.FindResults
		var findOpts []browser.FindOption
		if req.Exact {
			findOpts = append(findOpts, browser.WithExact())
		}
		result, err = b.FindPlaceholder(req.Text, findOpts...)
		if err == nil {
			resp.Text = result.Text
		}
	case "find.alt":
		if req.Text == "" {
			err = fmt.Errorf("text is required")
			break
		}
		var result *browser.FindResults
		var findOpts []browser.FindOption
		if req.Exact {
			findOpts = append(findOpts, browser.WithExact())
		}
		result, err = b.FindAlt(req.Text, findOpts...)
		if err == nil {
			resp.Text = result.Text
		}
	case "find.title":
		if req.Text == "" {
			err = fmt.Errorf("text is required")
			break
		}
		var result *browser.FindResults
		var findOpts []browser.FindOption
		if req.Exact {
			findOpts = append(findOpts, browser.WithExact())
		}
		result, err = b.FindTitle(req.Text, findOpts...)
		if err == nil {
			resp.Text = result.Text
		}
	case "find.testid":
		if req.Text == "" {
			err = fmt.Errorf("testid is required")
			break
		}
		var result *browser.FindResults
		result, err = b.FindTestID(req.Text)
		if err == nil {
			resp.Text = result.Text
		}
	case "wait.hidden":
		if req.CSSSelector == "" {
			err = fmt.Errorf("css selector is required")
			break
		}
		err = b.WaitHidden(req.CSSSelector)
	case "wait.download":
		if req.SaveDir == "" {
			err = fmt.Errorf("save directory is required")
			break
		}
		var path string
		path, err = b.WaitDownload(req.SaveDir)
		if err == nil {
			resp.FilePath = path
		}
	case "get.cdp-url":
		var text string
		text, err = b.GetCDPURL()
		if err == nil {
			resp.Text = text
		}
	case "session.info":
		resp.Text = sessionName

	case "confirm":
		if req.ConfirmID == "" {
			err = fmt.Errorf("confirmation ID is required")
			break
		}
		err = confirmStore.Confirm(req.ConfirmID)

	case "deny":
		if req.ConfirmID == "" {
			err = fmt.Errorf("confirmation ID is required")
			break
		}
		err = confirmStore.Deny(req.ConfirmID)

	default:
		err = fmt.Errorf("unknown command %q", req.Command)
	}

	if err != nil {
		resp.OK = false
		resp.Error = err.Error()
	}

	if encodeErr := json.NewEncoder(conn).Encode(resp); encodeErr != nil {
		return false, encodeErr
	}
	return false, err
}
