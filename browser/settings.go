package browser

import (
	"context"
	"fmt"

	cdpbrowser "github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// DeviceDescriptor describes a device for emulation.
type DeviceDescriptor struct {
	Name      string
	Width     int64
	Height    int64
	Scale     float64
	Mobile    bool
	UserAgent string
}

// predefined devices for SetDevice
var devices = map[string]DeviceDescriptor{
	"iphone 12": {
		Name: "iPhone 12", Width: 390, Height: 844, Scale: 3, Mobile: true,
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
	},
	"iphone 13": {
		Name: "iPhone 13", Width: 390, Height: 844, Scale: 3, Mobile: true,
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
	},
	"iphone 14": {
		Name: "iPhone 14", Width: 390, Height: 844, Scale: 3, Mobile: true,
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	},
	"iphone 14 pro": {
		Name: "iPhone 14 Pro", Width: 393, Height: 852, Scale: 3, Mobile: true,
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	},
	"iphone se": {
		Name: "iPhone SE", Width: 375, Height: 667, Scale: 2, Mobile: true,
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
	},
	"ipad": {
		Name: "iPad", Width: 768, Height: 1024, Scale: 2, Mobile: true,
		UserAgent: "Mozilla/5.0 (iPad; CPU OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
	},
	"ipad pro": {
		Name: "iPad Pro", Width: 1024, Height: 1366, Scale: 2, Mobile: true,
		UserAgent: "Mozilla/5.0 (iPad; CPU OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
	},
	"pixel 5": {
		Name: "Pixel 5", Width: 393, Height: 851, Scale: 2.75, Mobile: true,
		UserAgent: "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.91 Mobile Safari/537.36",
	},
	"samsung galaxy s21": {
		Name: "Samsung Galaxy S21", Width: 360, Height: 800, Scale: 3, Mobile: true,
		UserAgent: "Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.91 Mobile Safari/537.36",
	},
	"desktop 1080p": {
		Name: "Desktop 1080p", Width: 1920, Height: 1080, Scale: 1, Mobile: false,
	},
	"desktop 1440p": {
		Name: "Desktop 1440p", Width: 2560, Height: 1440, Scale: 1, Mobile: false,
	},
}

// SetViewport sets the browser viewport size with optional device scale factor.
// If scale is not provided or is 0, defaults to 1.0.
func (b *Browser) SetViewport(width, height int, scale ...float64) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	s := 1.0
	if len(scale) > 0 && scale[0] > 0 {
		s = scale[0]
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return emulation.SetDeviceMetricsOverride(int64(width), int64(height), s, false).Do(ctx)
	}))
}

// SetDevice emulates a specific device by name (e.g., "iPhone 12", "Pixel 5").
// Returns error if the device is not recognized.
func (b *Browser) SetDevice(name string) error {
	key := toLowerTrimmed(name)
	dev, ok := devices[key]
	if !ok {
		return fmt.Errorf("unknown device %q; available: %s", name, availableDeviceNames())
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		if err := emulation.SetDeviceMetricsOverride(dev.Width, dev.Height, dev.Scale, dev.Mobile).Do(ctx); err != nil {
			return err
		}
		if dev.UserAgent != "" {
			if err := emulation.SetUserAgentOverride(dev.UserAgent).Do(ctx); err != nil {
				return err
			}
		}
		return nil
	}))
}

// SetGeo overrides geolocation to the specified latitude and longitude.
// Use accuracy > 0 for a more precise emulation; defaults to 1 meter.
func (b *Browser) SetGeo(lat, lon float64) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Grant geolocation permission first
		_ = cdpbrowser.SetPermission(
			&cdpbrowser.PermissionDescriptor{Name: "geolocation"},
			cdpbrowser.PermissionSettingGranted,
		).Do(ctx)

		return emulation.SetGeolocationOverride().
			WithLatitude(lat).
			WithLongitude(lon).
			WithAccuracy(1).
			Do(ctx)
	}))
}

// ClearGeo removes the geolocation override.
func (b *Browser) ClearGeo() error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return emulation.ClearGeolocationOverride().Do(ctx)
	}))
}

// SetOffline enables or disables offline mode by emulating network conditions.
func (b *Browser) SetOffline(offline bool) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return network.OverrideNetworkState(offline, 0, -1, -1).Do(ctx)
	}))
}

// SetHeaders sets extra HTTP headers to be sent with every request.
func (b *Browser) SetHeaders(headers map[string]string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	h := make(network.Headers)
	for k, v := range headers {
		h[k] = v
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return network.SetExtraHTTPHeaders(h).Do(ctx)
	}))
}

// SetCredentials sets HTTP Basic Auth credentials for all requests.
// The credentials are injected via extra HTTP headers (Authorization).
func (b *Browser) SetCredentials(user, pass string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Use base64 encoding for Basic auth
		encoded := basicAuthHeader(user, pass)
		h := make(network.Headers)
		h["Authorization"] = "Basic " + encoded
		return network.SetExtraHTTPHeaders(h).Do(ctx)
	}))
}

// SetMedia sets the emulated CSS media type and/or color scheme.
// media can be "screen", "print", or "".
// colorScheme can be "dark", "light", or "".
func (b *Browser) SetMedia(features ...MediaFeature) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var mediaFeatures []*emulation.MediaFeature
		for _, f := range features {
			mediaFeatures = append(mediaFeatures, &emulation.MediaFeature{
				Name:  f.Name,
				Value: f.Value,
			})
		}

		return emulation.SetEmulatedMedia().
			WithFeatures(mediaFeatures).
			Do(ctx)
	}))
}

// MediaFeature represents a CSS media feature to emulate.
type MediaFeature struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SetColorScheme is a convenience method for SetMedia with prefers-color-scheme.
func (b *Browser) SetColorScheme(scheme string) error {
	return b.SetMedia(MediaFeature{Name: "prefers-color-scheme", Value: scheme})
}

// -- helpers --

func toLowerTrimmed(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result = append(result, c)
	}
	return string(result)
}

func availableDeviceNames() string {
	names := make([]string, 0, len(devices))
	for _, d := range devices {
		names = append(names, d.Name)
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}

func basicAuthHeader(user, pass string) string {
	src := []byte(user + ":" + pass)
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result []byte
	for i := 0; i < len(src); i += 3 {
		var b0, b1, b2 byte
		b0 = src[i]
		if i+1 < len(src) {
			b1 = src[i+1]
		}
		if i+2 < len(src) {
			b2 = src[i+2]
		}

		result = append(result, base64Chars[(b0>>2)&0x3F])
		result = append(result, base64Chars[((b0<<4)|(b1>>4))&0x3F])
		if i+1 < len(src) {
			result = append(result, base64Chars[((b1<<2)|(b2>>6))&0x3F])
		} else {
			result = append(result, '=')
		}
		if i+2 < len(src) {
			result = append(result, base64Chars[b2&0x3F])
		} else {
			result = append(result, '=')
		}
	}
	return string(result)
}
