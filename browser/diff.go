package browser

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
)

// DiffSnapshotResult holds the result of comparing two accessibility snapshots.
type DiffSnapshotResult struct {
	// Text is the unified diff output.
	Text string `json:"text"`
	// Added is the number of added lines.
	Added int `json:"added"`
	// Removed is the number of removed lines.
	Removed int `json:"removed"`
	// Changed is true if there are any differences.
	Changed bool `json:"changed"`
}

// DiffSnapshotOptions configures snapshot diff behavior.
type DiffSnapshotOptions struct {
	// BaselineFile is the path to a saved baseline snapshot file.
	// If empty, the previous snapshot (lastSnap) is used as baseline.
	BaselineFile string
	// SnapshotOptions for taking the current snapshot.
	SnapshotOptions SnapshotOptions
}

// DiffSnapshot compares the current page snapshot with a baseline.
// If opts.BaselineFile is empty, compares with the previous Snapshot() result.
// Returns a unified diff of the two snapshots.
func (b *Browser) DiffSnapshot(opts ...DiffSnapshotOptions) (*DiffSnapshotResult, error) {
	config := DiffSnapshotOptions{}
	if len(opts) > 0 {
		config = opts[0]
	}

	// Get baseline text
	var baselineText string
	if config.BaselineFile != "" {
		data, err := os.ReadFile(config.BaselineFile)
		if err != nil {
			return nil, fmt.Errorf("read baseline file: %w", err)
		}
		baselineText = string(data)
	} else {
		if b.lastSnap == nil {
			return nil, fmt.Errorf("no previous snapshot for comparison; call Snapshot() first or provide a baseline file")
		}
		baselineText = b.lastSnap.Text
	}

	// Take current snapshot
	current, err := b.Snapshot(config.SnapshotOptions)
	if err != nil {
		return nil, fmt.Errorf("take current snapshot: %w", err)
	}

	// Compute unified diff
	diff := unifiedDiff(baselineText, current.Text, "before", "after")

	added, removed := countDiffStats(diff)

	return &DiffSnapshotResult{
		Text:    diff,
		Added:   added,
		Removed: removed,
		Changed: added > 0 || removed > 0,
	}, nil
}

// DiffScreenshotResult holds the result of a pixel-level screenshot comparison.
type DiffScreenshotResult struct {
	// DiffCount is the number of pixels that differ.
	DiffCount int `json:"diffCount"`
	// TotalPixels is the total number of pixels compared.
	TotalPixels int `json:"totalPixels"`
	// DiffPercent is the percentage of differing pixels (0-100).
	DiffPercent float64 `json:"diffPercent"`
	// Changed is true if there are any pixel differences above threshold.
	Changed bool `json:"changed"`
	// DiffImage holds the diff image bytes (PNG), if OutputPath was set.
	DiffImage []byte `json:"-"`
}

// DiffScreenshotOptions configures screenshot diff behavior.
type DiffScreenshotOptions struct {
	// BaselineFile is the path to a baseline PNG image (required).
	BaselineFile string
	// OutputPath is the path to save the diff image. If empty, no diff image is saved.
	OutputPath string
	// Threshold is the color distance threshold (0-1). Default: 0.1.
	// Pixels with a normalized distance below this are considered identical.
	Threshold float64
	// FullPage captures the full scrollable page.
	FullPage bool
	// ElementID if > 0, capture only this element for comparison.
	ElementID int
}

// DiffScreenshot performs a pixel-level comparison between the current page
// screenshot and a baseline PNG image.
func (b *Browser) DiffScreenshot(opts DiffScreenshotOptions) (*DiffScreenshotResult, error) {
	if opts.BaselineFile == "" {
		return nil, fmt.Errorf("baseline file is required for screenshot diff")
	}
	if opts.Threshold <= 0 {
		opts.Threshold = 0.1
	}

	// Read baseline image
	baselineData, err := os.ReadFile(opts.BaselineFile)
	if err != nil {
		return nil, fmt.Errorf("read baseline: %w", err)
	}

	baselineImg, err := png.Decode(bytes.NewReader(baselineData))
	if err != nil {
		return nil, fmt.Errorf("decode baseline PNG: %w", err)
	}

	// Take current screenshot
	screenshotData, err := b.ScreenshotToBytes(ScreenshotOptions{
		FullPage:  opts.FullPage,
		ElementID: opts.ElementID,
	})
	if err != nil {
		return nil, fmt.Errorf("take screenshot: %w", err)
	}

	currentImg, err := png.Decode(bytes.NewReader(screenshotData))
	if err != nil {
		return nil, fmt.Errorf("decode current screenshot: %w", err)
	}

	// Compare images
	result := compareImages(baselineImg, currentImg, opts.Threshold)

	// If output path is set, create and save diff image
	if opts.OutputPath != "" && result.Changed {
		diffImg := createDiffImage(baselineImg, currentImg, opts.Threshold)
		var buf bytes.Buffer
		if err := png.Encode(&buf, diffImg); err != nil {
			return nil, fmt.Errorf("encode diff image: %w", err)
		}
		result.DiffImage = buf.Bytes()
		if err := os.WriteFile(opts.OutputPath, result.DiffImage, 0644); err != nil {
			return nil, fmt.Errorf("write diff image: %w", err)
		}
	}

	return result, nil
}

// DiffURLResult holds the result of comparing two URLs.
type DiffURLResult struct {
	// SnapshotDiff is the snapshot comparison result.
	SnapshotDiff *DiffSnapshotResult `json:"snapshotDiff"`
	// ScreenshotDiff is the screenshot comparison result (if requested).
	ScreenshotDiff *DiffScreenshotResult `json:"screenshotDiff,omitempty"`
}

// DiffURLOptions configures URL comparison behavior.
type DiffURLOptions struct {
	// IncludeScreenshot also compares screenshots, not just snapshots.
	IncludeScreenshot bool
	// FullPage captures full page screenshots.
	FullPage bool
	// SnapshotOptions for taking snapshots.
	SnapshotOptions SnapshotOptions
	// Threshold for screenshot comparison (0-1). Default: 0.1.
	Threshold float64
}

// DiffURL navigates to two URLs and compares their snapshots (and optionally screenshots).
// This method navigates the browser to url1, takes a snapshot, then navigates to url2
// and takes another snapshot, then computes the diff.
func (b *Browser) DiffURL(url1, url2 string, opts ...DiffURLOptions) (*DiffURLResult, error) {
	config := DiffURLOptions{}
	if len(opts) > 0 {
		config = opts[0]
	}
	if config.Threshold <= 0 {
		config.Threshold = 0.1
	}

	// Navigate to url1 and take snapshot
	if err := b.Open(url1); err != nil {
		return nil, fmt.Errorf("open url1: %w", err)
	}
	snap1, err := b.Snapshot(config.SnapshotOptions)
	if err != nil {
		return nil, fmt.Errorf("snapshot url1: %w", err)
	}

	var screenshot1 []byte
	if config.IncludeScreenshot {
		screenshot1, err = b.ScreenshotToBytes(ScreenshotOptions{FullPage: config.FullPage})
		if err != nil {
			return nil, fmt.Errorf("screenshot url1: %w", err)
		}
	}

	// Navigate to url2 and take snapshot
	if err := b.Open(url2); err != nil {
		return nil, fmt.Errorf("open url2: %w", err)
	}
	snap2, err := b.Snapshot(config.SnapshotOptions)
	if err != nil {
		return nil, fmt.Errorf("snapshot url2: %w", err)
	}

	// Compute snapshot diff
	diff := unifiedDiff(snap1.Text, snap2.Text, url1, url2)
	added, removed := countDiffStats(diff)
	snapshotDiff := &DiffSnapshotResult{
		Text:    diff,
		Added:   added,
		Removed: removed,
		Changed: added > 0 || removed > 0,
	}

	result := &DiffURLResult{SnapshotDiff: snapshotDiff}

	// Optionally compare screenshots
	if config.IncludeScreenshot {
		screenshot2, err := b.ScreenshotToBytes(ScreenshotOptions{FullPage: config.FullPage})
		if err != nil {
			return nil, fmt.Errorf("screenshot url2: %w", err)
		}

		img1, err := png.Decode(bytes.NewReader(screenshot1))
		if err != nil {
			return nil, fmt.Errorf("decode screenshot1: %w", err)
		}
		img2, err := png.Decode(bytes.NewReader(screenshot2))
		if err != nil {
			return nil, fmt.Errorf("decode screenshot2: %w", err)
		}

		result.ScreenshotDiff = compareImages(img1, img2, config.Threshold)
	}

	return result, nil
}

// --- internal helpers ---

// unifiedDiff produces a simple unified diff between two strings.
func unifiedDiff(a, b, labelA, labelB string) string {
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	// Use a simple LCS-based diff algorithm
	lcs := computeLCS(linesA, linesB)

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("--- %s\n", labelA))
	buf.WriteString(fmt.Sprintf("+++ %s\n", labelB))

	idxA, idxB, idxLCS := 0, 0, 0

	for idxLCS < len(lcs) {
		// Find lines only in A (removed)
		for idxA < len(linesA) && linesA[idxA] != lcs[idxLCS] {
			buf.WriteString(fmt.Sprintf("-%s\n", linesA[idxA]))
			idxA++
		}
		// Find lines only in B (added)
		for idxB < len(linesB) && linesB[idxB] != lcs[idxLCS] {
			buf.WriteString(fmt.Sprintf("+%s\n", linesB[idxB]))
			idxB++
		}
		// Common line
		buf.WriteString(fmt.Sprintf(" %s\n", lcs[idxLCS]))
		idxA++
		idxB++
		idxLCS++
	}

	// Remaining lines in A
	for idxA < len(linesA) {
		buf.WriteString(fmt.Sprintf("-%s\n", linesA[idxA]))
		idxA++
	}
	// Remaining lines in B
	for idxB < len(linesB) {
		buf.WriteString(fmt.Sprintf("+%s\n", linesB[idxB]))
		idxB++
	}

	return buf.String()
}

// computeLCS computes the Longest Common Subsequence of two string slices.
func computeLCS(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	// Backtrack to find LCS
	lcs := make([]string, 0, dp[m][n])
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append(lcs, a[i-1])
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	// Reverse
	for left, right := 0, len(lcs)-1; left < right; left, right = left+1, right-1 {
		lcs[left], lcs[right] = lcs[right], lcs[left]
	}
	return lcs
}

// countDiffStats counts added and removed lines in a unified diff.
func countDiffStats(diff string) (added, removed int) {
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removed++
		}
	}
	return
}

// compareImages compares two images pixel-by-pixel.
func compareImages(img1, img2 image.Image, threshold float64) *DiffScreenshotResult {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// Use the intersection of the two images for comparison
	maxX := bounds1.Max.X
	if bounds2.Max.X < maxX {
		maxX = bounds2.Max.X
	}
	maxY := bounds1.Max.Y
	if bounds2.Max.Y < maxY {
		maxY = bounds2.Max.Y
	}

	totalPixels := maxX * maxY
	if totalPixels == 0 {
		return &DiffScreenshotResult{TotalPixels: 0, Changed: false}
	}

	diffCount := 0

	// Count pixels outside the common area as different
	area1 := bounds1.Dx() * bounds1.Dy()
	area2 := bounds2.Dx() * bounds2.Dy()
	if area1 > totalPixels {
		diffCount += area1 - totalPixels
	}
	if area2 > totalPixels {
		diffCount += area2 - totalPixels
	}

	for y := 0; y < maxY; y++ {
		for x := 0; x < maxX; x++ {
			if colorDistance(img1.At(x, y), img2.At(x, y)) > threshold {
				diffCount++
			}
		}
	}

	maxTotal := totalPixels
	if area1 > maxTotal {
		maxTotal = area1
	}
	if area2 > maxTotal {
		maxTotal = area2
	}

	percent := 0.0
	if maxTotal > 0 {
		percent = float64(diffCount) / float64(maxTotal) * 100
	}

	return &DiffScreenshotResult{
		DiffCount:   diffCount,
		TotalPixels: maxTotal,
		DiffPercent: math.Round(percent*100) / 100,
		Changed:     diffCount > 0,
	}
}

// createDiffImage creates a visual diff image highlighting differences in red.
func createDiffImage(img1, img2 image.Image, threshold float64) image.Image {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	width := bounds1.Dx()
	if bounds2.Dx() > width {
		width = bounds2.Dx()
	}
	height := bounds1.Dy()
	if bounds2.Dy() > height {
		height = bounds2.Dy()
	}

	diff := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			inBounds1 := x < bounds1.Dx() && y < bounds1.Dy()
			inBounds2 := x < bounds2.Dx() && y < bounds2.Dy()

			if !inBounds1 || !inBounds2 {
				// Out-of-bounds pixels are marked in magenta
				diff.Set(x, y, color.RGBA{R: 255, G: 0, B: 255, A: 255})
			} else {
				c1 := img1.At(x+bounds1.Min.X, y+bounds1.Min.Y)
				c2 := img2.At(x+bounds2.Min.X, y+bounds2.Min.Y)
				if colorDistance(c1, c2) > threshold {
					// Differing pixel → red overlay
					diff.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 200})
				} else {
					// Same pixel → dimmed version of img2
					r, g, b2, a := c2.RGBA()
					diff.Set(x, y, color.RGBA{
						R: uint8(r >> 8 / 3),
						G: uint8(g >> 8 / 3),
						B: uint8(b2 >> 8 / 3),
						A: uint8(a >> 8),
					})
				}
			}
		}
	}

	return diff
}

// colorDistance computes the normalized Euclidean distance between two colors (0-1).
func colorDistance(c1, c2 color.Color) float64 {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	dr := float64(r1) - float64(r2)
	dg := float64(g1) - float64(g2)
	db := float64(b1) - float64(b2)
	da := float64(a1) - float64(a2)

	// Max possible distance: sqrt(4 * 65535^2)
	maxDist := 65535.0 * 2.0

	return math.Sqrt(dr*dr+dg*dg+db*db+da*da) / maxDist
}
