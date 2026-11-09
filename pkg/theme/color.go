package theme

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// OKLCH converts CSS OKLCH (lightness 0..1, chroma >=0, hue in degrees)
// to display sRGB. Out-of-gamut channels are clipped at the display boundary.
func OKLCH(l, c, h float64, alpha ...float64) ui.Color {
	a := 1.0
	if len(alpha) > 0 {
		a = alpha[0]
	}
	l = clamp(l)
	c = max(0, c)
	h *= math.Pi / 180
	x, y := c*math.Cos(h), c*math.Sin(h)
	lm := l + 0.3963377774*x + 0.2158037573*y
	mm := l - 0.1055613458*x - 0.0638541728*y
	sm := l - 0.0894841775*x - 1.2914855480*y
	lm, mm, sm = lm*lm*lm, mm*mm*mm, sm*sm*sm
	encode := func(v float64) uint8 {
		if v <= 0.0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - 0.055
		}
		return byteOf(v)
	}
	return ui.Color{
		R: encode(4.0767416621*lm - 3.3077115913*mm + 0.2309699292*sm),
		G: encode(-1.2684380046*lm + 2.6097574011*mm - 0.3413193965*sm),
		B: encode(-0.0041960863*lm - 0.7034186147*mm + 1.7076147010*sm), A: byteOf(a),
	}
}

func clamp(v float64) float64 {
	if math.IsNaN(v) {
		return 0
	}
	return max(0, min(1, v))
}
func byteOf(v float64) uint8 { return uint8(math.Round(clamp(v) * 255)) }

// ParseColor reads CSS hex (#rgb, #rgba, #rrggbb, #rrggbbaa) and OKLCH.
// It returns an error instead of silently turning invalid user input black.
func ParseColor(value string) (ui.Color, error) {
	s := strings.TrimSpace(value)
	if strings.HasPrefix(s, "#") {
		h := s[1:]
		if len(h) == 3 || len(h) == 4 {
			var b strings.Builder
			for _, ch := range h {
				b.WriteRune(ch)
				b.WriteRune(ch)
			}
			h = b.String()
		}
		if len(h) == 6 {
			h += "ff"
		}
		if len(h) != 8 {
			return ui.Color{}, fmt.Errorf("invalid hex color %q", value)
		}
		n, err := strconv.ParseUint(h, 16, 32)
		if err != nil {
			return ui.Color{}, fmt.Errorf("invalid hex color %q: %w", value, err)
		}
		return ui.Color{R: uint8(n >> 24), G: uint8(n >> 16), B: uint8(n >> 8), A: uint8(n)}, nil
	}
	if strings.HasPrefix(s, "oklch(") && strings.HasSuffix(s, ")") {
		parts := strings.Fields(strings.ReplaceAll(s[6:len(s)-1], "/", " / "))
		if len(parts) != 3 && (len(parts) != 5 || parts[3] != "/") {
			return ui.Color{}, fmt.Errorf("invalid OKLCH color %q", value)
		}
		values := make([]float64, 0, 4)
		for i, p := range parts {
			if i == 3 {
				continue
			}
			// CSS uses none for a missing channel, notably the hue of
			// neutral colors. It computes to zero for an isolated color.
			if p == "none" {
				p = "0"
			}
			percent := strings.HasSuffix(p, "%")
			v, err := strconv.ParseFloat(strings.TrimSuffix(p, "%"), 64)
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return ui.Color{}, fmt.Errorf("invalid OKLCH channel %q", p)
			}
			if percent {
				if i == 1 {
					v *= .004
				} else if i != 2 {
					v /= 100
				} else {
					return ui.Color{}, fmt.Errorf("hue cannot be a percentage")
				}
			}
			values = append(values, v)
		}
		return OKLCH(values[0], values[1], values[2], values[3:]...), nil
	}
	return ui.Color{}, fmt.Errorf("unsupported color %q", value)
}

// Mix blends colors in premultiplied sRGB, including alpha.
func Mix(a, b ui.Color, amount float32) ui.Color {
	t := clamp(float64(amount))
	aa, ba := float64(a.A)/255, float64(b.A)/255
	out := aa*(1-t) + ba*t
	if out == 0 {
		return ui.Transparent
	}
	ch := func(x, y uint8) uint8 { return uint8(math.Round((float64(x)*aa*(1-t) + float64(y)*ba*t) / out)) }
	return ui.Color{R: ch(a.R, b.R), G: ch(a.G, b.G), B: ch(a.B, b.B), A: byteOf(out)}
}

// Contrast returns the WCAG contrast ratio of two opaque sRGB colors.
func Contrast(a, b ui.Color) float64 {
	luminance := func(c ui.Color) float64 {
		linear := func(v uint8) float64 {
			s := float64(v) / 255
			if s <= .04045 {
				return s / 12.92
			}
			return math.Pow((s+.055)/1.055, 2.4)
		}
		return .2126*linear(c.R) + .7152*linear(c.G) + .0722*linear(c.B)
	}
	x, y := luminance(a), luminance(b)
	return (max(x, y) + .05) / (min(x, y) + .05)
}
