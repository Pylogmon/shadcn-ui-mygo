package theme

import (
	"math"
	"weak"

	"github.com/egoist/mygo/ui"
)

// TextScale returns the desktop's finite, positive text scale, or one.
func TextScale(c *ui.Context) float32 {
	scale := c.Preferences().TextScale
	if !positiveFontSize(scale) {
		return 1
	}
	return scale
}

func positiveFontSize(size float32) bool {
	return size > 0 && !math.IsNaN(float64(size)) && !math.IsInf(float64(size), 0)
}

// ControlFontSize scales a component's reference size against the theme's
// 14-DIP baseline and the desktop text scale. A positive finite override is a
// local size in unscaled DIPs, independent of the theme and Size variant.
// Controller themes and menu snapshots already include the desktop scale;
// their recorded scale prevents applying it a second time.
func ControlFontSize(c *ui.Context, reference, override float32) float32 {
	scale := TextScale(c)
	if positiveFontSize(override) {
		return override * scale
	}
	base := float32(14)
	if native := c.Theme(); native != nil {
		if positiveFontSize(native.FontSize) {
			base = native.FontSize
		}
		if value, ok := metadata.Load(weak.Make(native)); ok {
			applied := value.(themeMetadata).menu.textScale
			if positiveFontSize(applied) {
				base /= applied
			}
		} else {
			// MyGo's platform theme already follows desktop preferences.
			base /= scale
		}
	}
	return reference / 14 * base * scale
}
