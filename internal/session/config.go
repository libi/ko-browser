package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/libi/ko-browser/browser"
)

type Options struct {
	Name      string
	Headed    bool
	Timeout   time.Duration
	Debug     bool
	Profile   string
	StatePath string

	// Phase 8.3: Global options
	UserAgent         string
	Proxy             string
	ProxyBypass       string
	IgnoreHTTPSErrors bool
	AllowFileAccess   bool
	Extensions        []string
	ExtraArgs         []string
	DownloadPath      string
	ScreenshotDir     string
	ScreenshotFormat  string

	// Phase 8.4: Confirmation
	ConfirmActions []string
}

type Request struct {
	Command         string                  `json:"command"`
	URL             string                  `json:"url,omitempty"`
	ID              int                     `json:"id,omitempty"`
	Text            string                  `json:"text,omitempty"`
	Key             string                  `json:"key,omitempty"`
	Values          []string                `json:"values,omitempty"`
	Direction       string                  `json:"direction,omitempty"`
	Amount          int                     `json:"amount,omitempty"`
	AttrName        string                  `json:"attrName,omitempty"`
	CSSSelector     string                  `json:"cssSelector,omitempty"`
	Expression      string                  `json:"expression,omitempty"`
	WaitDuration    time.Duration           `json:"waitDuration,omitempty"`
	SnapshotOptions browser.SnapshotOptions `json:"snapshotOptions,omitempty"`

	// Phase 3: Screenshot/PDF/Eval/Find
	FilePath       string         `json:"filePath,omitempty"`
	ScreenshotArgs ScreenshotArgs `json:"screenshotArgs,omitempty"`
	PDFArgs        PDFArgs        `json:"pdfArgs,omitempty"`
	Role           string         `json:"role,omitempty"`
	N              int            `json:"n,omitempty"`

	// Phase 4: Drag/Upload/Download/Mouse
	DstID    int      `json:"dstId,omitempty"`
	Files    []string `json:"files,omitempty"`
	SaveDir  string   `json:"saveDir,omitempty"`
	X        float64  `json:"x,omitempty"`
	Y        float64  `json:"y,omitempty"`
	DeltaX   float64  `json:"deltaX,omitempty"`
	DeltaY   float64  `json:"deltaY,omitempty"`
	MouseBtn string   `json:"mouseBtn,omitempty"`

	// Phase 5: Tab/Network/Storage
	TabIndex    int                `json:"tabIndex,omitempty"`
	Pattern     string             `json:"pattern,omitempty"`
	RouteAction string             `json:"routeAction,omitempty"`
	Cookie      browser.CookieInfo `json:"cookie,omitempty"`
	StorageType string             `json:"storageType,omitempty"`
	StorageKey  string             `json:"storageKey,omitempty"`
	StorageVal  string             `json:"storageVal,omitempty"`

	// Phase 6: Settings/Debug/Clipboard
	Width         int                    `json:"width,omitempty"`
	Height        int                    `json:"height,omitempty"`
	DeviceName    string                 `json:"deviceName,omitempty"`
	Lat           float64                `json:"lat,omitempty"`
	Lon           float64                `json:"lon,omitempty"`
	Offline       *bool                  `json:"offline,omitempty"`
	Headers       map[string]string      `json:"headers,omitempty"`
	User          string                 `json:"user,omitempty"`
	Pass          string                 `json:"pass,omitempty"`
	MediaFeatures []browser.MediaFeature `json:"mediaFeatures,omitempty"`
	ColorScheme   string                 `json:"colorScheme,omitempty"`
	ClipboardText string                 `json:"clipboardText,omitempty"`
	ConsoleLevel  string                 `json:"consoleLevel,omitempty"`

	// Phase 7: Advanced features
	Target            string  `json:"target,omitempty"`            // connect target (port or ws URL)
	BaselineFile      string  `json:"baselineFile,omitempty"`      // diff baseline file path
	OutputFile        string  `json:"outputFile,omitempty"`        // output file path for diff/trace/profiler/record
	Threshold         float64 `json:"threshold,omitempty"`         // screenshot diff threshold
	URL2              string  `json:"url2,omitempty"`              // second URL for diff url
	IncludeScreenshot bool    `json:"includeScreenshot,omitempty"` // diff url: also compare screenshots
	Categories        string  `json:"categories,omitempty"`        // trace categories
	StatePath         string  `json:"statePath,omitempty"`         // state import/export path

	// Phase 8: Missing commands/flags
	NewTab    bool    `json:"newTab,omitempty"`    // click with new tab
	Exact     bool    `json:"exact,omitempty"`     // exact matching for find
	Scale     float64 `json:"scale,omitempty"`     // viewport scale factor
	Cursor    bool    `json:"cursor,omitempty"`    // show cursor in snapshot
	ConfirmID string  `json:"confirmId,omitempty"` // confirmation ID for confirm/deny
}

// ScreenshotArgs holds arguments for the screenshot command.
type ScreenshotArgs struct {
	FullPage  bool `json:"fullPage,omitempty"`
	Quality   int  `json:"quality,omitempty"`
	ElementID int  `json:"elementID,omitempty"`
	Annotate  bool `json:"annotate,omitempty"`
}

// PDFArgs holds arguments for the pdf command.
type PDFArgs struct {
	Landscape bool `json:"landscape,omitempty"`
	PrintBG   bool `json:"printBG,omitempty"`
}

type Response struct {
	OK           bool               `json:"ok"`
	Error        string             `json:"error,omitempty"`
	Text         string             `json:"text,omitempty"`
	RawCount     int                `json:"rawCount,omitempty"`
	ElementCount int                `json:"elementCount,omitempty"`
	IntResult    int                `json:"intResult,omitempty"`
	BoolResult   *bool              `json:"boolResult,omitempty"`
	Box          *browser.BoxResult `json:"box,omitempty"`
	FilePath     string             `json:"filePath,omitempty"`

	// Phase 5
	Tabs         []browser.TabInfo        `json:"tabs,omitempty"`
	Cookies      []browser.CookieInfo     `json:"cookies,omitempty"`
	Requests     []browser.NetworkRequest `json:"requests,omitempty"`
	StorageItems map[string]string        `json:"storageItems,omitempty"`

	// Phase 6
	ConsoleMessages []browser.ConsoleMessage `json:"consoleMessages,omitempty"`
	PageErrors      []browser.PageError      `json:"pageErrors,omitempty"`

	// Phase 7
	DiffResult     *browser.DiffSnapshotResult   `json:"diffResult,omitempty"`
	ScreenshotDiff *browser.DiffScreenshotResult `json:"screenshotDiff,omitempty"`
	DiffURLResult  *browser.DiffURLResult        `json:"diffURLResult,omitempty"`

	// Phase 8.4: Confirmation
	ConfirmationRequired bool   `json:"confirmationRequired,omitempty"`
	ConfirmID            string `json:"confirmId,omitempty"`
}

func (o Options) normalized() Options {
	if strings.TrimSpace(o.Name) == "" {
		o.Name = "default"
	}
	if o.Timeout <= 0 {
		o.Timeout = 30 * time.Second
	}
	return o
}

func socketPath(name string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("ko-browser-%s.sock", sanitizeSessionName(name)))
}

func sanitizeSessionName(name string) string {
	var builder strings.Builder
	for _, r := range strings.TrimSpace(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}
	if builder.Len() == 0 {
		return "default"
	}
	return builder.String()
}
