package browser

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// recordState tracks an in-progress screen recording.
type recordState struct {
	mu       sync.Mutex
	active   bool
	path     string
	frames   [][]byte
	cancelFn context.CancelFunc
}

// RecordStart begins recording screenshots (screencast frames) from the browser.
// The frames are collected as PNG images. When stopped, they are saved to the output path.
// For simplicity, the recording captures individual frames rather than encoding a video format.
// The output is saved as a series of PNG files or as a concatenated binary.
func (b *Browser) RecordStart(outputPath string) error {
	if b.recording != nil && b.recording.active {
		return fmt.Errorf("recording already in progress; call RecordStop first")
	}

	ctx := b.activeContext()
	listenCtx, listenCancel := context.WithCancel(ctx)

	rs := &recordState{
		active:   true,
		path:     outputPath,
		cancelFn: listenCancel,
	}

	// Listen for screencast frames
	chromedp.ListenTarget(listenCtx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventScreencastFrame:
			// Decode the base64 frame data
			data, err := base64.StdEncoding.DecodeString(e.Data)
			if err == nil {
				rs.mu.Lock()
				rs.frames = append(rs.frames, data)
				rs.mu.Unlock()
			}

			// Acknowledge the frame
			go func() {
				opCtx, opCancel := b.operationContext()
				defer opCancel()
				_ = chromedp.Run(opCtx, chromedp.ActionFunc(func(ctx context.Context) error {
					return page.ScreencastFrameAck(e.SessionID).Do(ctx)
				}))
			}()
		}
	})

	// Start screencast
	opCtx, opCancel := b.operationContext()
	defer opCancel()

	err := chromedp.Run(opCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		return page.StartScreencast().
			WithFormat(page.ScreencastFormatPng).
			WithEveryNthFrame(2). // capture every 2nd frame for performance
			Do(ctx)
	}))
	if err != nil {
		listenCancel()
		return fmt.Errorf("start screencast: %w", err)
	}

	b.recording = rs
	return nil
}

// RecordStop stops the recording and saves the captured frames.
// Frames are saved as individual PNG files: {outputPath}_001.png, {outputPath}_002.png, etc.
// Returns the number of frames captured.
func (b *Browser) RecordStop() (int, error) {
	if b.recording == nil || !b.recording.active {
		return 0, fmt.Errorf("no recording in progress; call RecordStart first")
	}

	// Stop screencast
	opCtx, opCancel := b.operationContext()
	defer opCancel()

	err := chromedp.Run(opCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		return page.StopScreencast().Do(ctx)
	}))
	if err != nil {
		// Continue anyway to save what we have
		_ = err
	}

	// Cancel the listener
	if b.recording.cancelFn != nil {
		b.recording.cancelFn()
	}

	b.recording.mu.Lock()
	frames := b.recording.frames
	path := b.recording.path
	b.recording.mu.Unlock()

	b.recording.active = false
	b.recording = nil

	// Save frames
	if path != "" && len(frames) > 0 {
		if err := saveRecordingFrames(path, frames); err != nil {
			return len(frames), fmt.Errorf("save recording: %w", err)
		}
	}

	return len(frames), nil
}

// saveRecordingFrames saves captured frames as individual PNG files.
// If there's only one frame, saves directly as the path.
// Otherwise, saves as path_001.png, path_002.png, etc.
func saveRecordingFrames(basePath string, frames [][]byte) error {
	if len(frames) == 1 {
		return os.WriteFile(basePath, frames[0], 0644)
	}

	// Remove extension from basePath for numbered files
	ext := ".png"
	base := basePath
	for _, suffix := range []string{".png", ".webm", ".mp4"} {
		if len(base) > len(suffix) && base[len(base)-len(suffix):] == suffix {
			base = base[:len(base)-len(suffix)]
			break
		}
	}

	for i, frame := range frames {
		path := fmt.Sprintf("%s_%03d%s", base, i+1, ext)
		if err := os.WriteFile(path, frame, 0644); err != nil {
			return fmt.Errorf("write frame %d: %w", i+1, err)
		}
	}

	return nil
}
