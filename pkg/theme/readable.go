package theme

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// Composite returns foreground over background in display sRGB, including
// source alpha. Contrast must be measured after composition, not on tint RGB.
func Composite(foreground, background ui.Color) ui.Color {
	f, b := foreground.SRGB(), background.SRGB()
	fa, ba := float64(f.A)/255, float64(b.A)/255
	a := fa + ba*(1-fa)
	if a == 0 {
		return ui.Transparent
	}
	ch := func(x, y uint8) uint8 {
		return uint8(math.Round((float64(x)*fa + float64(y)*ba*(1-fa)) / a))
	}
	return ui.Color{R: ch(f.R, b.R), G: ch(f.G, b.G), B: ch(f.B, b.B), A: byteOf(a)}
}

// EnsureContrast is a native extension, not a shadcn token. It preserves a
// readable color exactly, otherwise chooses the shortest sRGB blend towards
// black or white that meets minimum on the opaque background. This also raises
// translucent text's alpha as needed. It never changes the source accent.
// Zero/invalid minimum means 4.5. If the requested ratio is unattainable (e.g.
// 7 on middle gray), the higher-contrast black/white endpoint is returned.
func EnsureContrast(color, background ui.Color, minimum float64) ui.Color {
	return ensureContrast(color, []ui.Color{background}, minimum, false)
}

func ensureContrast(color ui.Color, backgrounds []ui.Color, minimum float64, roundoff bool) ui.Color {
	color = color.SRGB()
	if minimum <= 0 || math.IsNaN(minimum) || math.IsInf(minimum, 0) {
		minimum = 4.5
	}
	quality := func(candidate ui.Color) float64 {
		ratio := float64(21)
		for _, background := range backgrounds {
			background = background.SRGB()
			background.A = 255
			if !roundoff {
				ratio = min(ratio, Contrast(Composite(candidate, background), background))
				continue
			}
			// Raster source-over can differ by one display step from repeated
			// CPU composition. Check both luminance extremes rather than adding
			// an arbitrary contrast cushion to every color.
			for _, delta := range []int{-1, 1} {
				bg := shiftRGB(background, delta)
				fg := Composite(candidate, bg)
				if candidate.A == 255 {
					ratio = min(ratio, Contrast(fg, bg))
				} else {
					ratio = min(ratio, Contrast(shiftRGB(fg, -1), bg), Contrast(shiftRGB(fg, 1), bg))
				}
			}
		}
		return ratio
	}
	if quality(color) >= minimum {
		return color
	}
	best, distance, bestRatio := color, float32(2), float64(0)
	for _, target := range []ui.Color{ui.RGB(0, 0, 0), ui.RGB(255, 255, 255)} {
		ratio := quality(target)
		if ratio > bestRatio && distance == 2 {
			best, bestRatio = target, ratio
		}
		if ratio < minimum {
			continue
		}
		lo, hi := float32(0), float32(1)
		// Search the actual quantized display colors; hi always passes.
		for range 24 {
			mid := (lo + hi) / 2
			if quality(Mix(color, target, mid)) >= minimum {
				hi = mid
			} else {
				lo = mid
			}
		}
		if hi < distance {
			best, distance = Mix(color, target, hi), hi
		}
	}
	if distance == 2 && len(backgrounds) > 1 {
		// No single color can cover backgrounds of opposite polarity. Use
		// the actual state; the component suppresses its color interpolation.
		return ensureContrast(color, backgrounds[:1], minimum, roundoff)
	}
	return best
}

func shiftRGB(color ui.Color, delta int) ui.Color {
	channel := func(value uint8) uint8 { return uint8(max(0, min(255, int(value)+delta))) }
	return ui.Color{R: channel(color.R), G: channel(color.G), B: channel(color.B), A: color.A}
}

type surfaceKey struct{}
type surfaceScope struct {
	color  ui.Color
	active bool
}

// SurfaceOf returns the current composited carrying surface. Outside an
// explicit/library surface scope it defaults to the theme page background.
func SurfaceOf(c *ui.Context) ui.Color {
	return SurfaceOr(c, TokensOf(c).Background)
}

// SurfaceOr is SurfaceOf with a component-specific fallback when no enclosing
// carrying surface has been declared (e.g. a standalone sidebar action).
func SurfaceOr(c *ui.Context, fallback ui.Color) ui.Color {
	s := ui.Local(c.Root(), surfaceKey{}, func() surfaceScope { return surfaceScope{} })
	if s.active {
		return s.color
	}
	return Composite(fallback, ui.RGB(255, 255, 255))
}

// ResolveSurface resolves a per-control surface override over its enclosing
// surface. Nil inherits. This describes the carrying surface, not the control's
// own fill, which must subsequently be composited on top.
func ResolveSurface(c *ui.Context, override *ui.Color) ui.Color {
	background := SurfaceOf(c)
	if override != nil {
		background = Composite(*override, background)
	}
	return background
}

// WithSurface scopes foreground calculations synchronously without inserting
// a layout node or changing tokens. Paint the same background on the container.
// Nested translucent surfaces compose in order; siblings and panics restore
// the previous scope. No Context/Element is retained across build passes.
func WithSurface(c *ui.Context, background ui.Color, content func()) {
	resolved := Composite(background, SurfaceOf(c))
	s := ui.Local(c.Root(), surfaceKey{}, func() surfaceScope { return surfaceScope{} })
	previous := *s
	*s = surfaceScope{color: resolved, active: true}
	defer func() { *s = previous }()
	if content != nil {
		content()
	}
}

// ReadableForeground resolves small colored text on its actual backgrounds.
// It targets 4.5:1, or 7:1 where attainable with the desktop HighContrast
// preference, accounting for one display step of compositing roundoff.
// Supply all interactive fills to keep foreground stable during background
// animation; the first is the actual state if no common foreground is possible.
// Backgrounds may be translucent over the current surface scope; none inherits.
// Primary/PrimaryForeground and official presets remain untouched.
func ReadableForeground(c *ui.Context, color ui.Color, backgrounds ...ui.Color) ui.Color {
	minimum := 4.5
	if c.Preferences().HighContrast {
		minimum = 7
	}
	resolved := make([]ui.Color, 0, max(1, len(backgrounds)))
	for _, background := range backgrounds {
		resolved = append(resolved, Composite(background, SurfaceOf(c)))
	}
	if len(resolved) == 0 {
		resolved = append(resolved, SurfaceOf(c))
	}
	return ensureContrast(color, resolved, minimum, true)
}
