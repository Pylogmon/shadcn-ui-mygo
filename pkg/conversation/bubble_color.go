package conversation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"math"
)

// Native tokens are display sRGB. Recover their perceptual coordinates for the
// source's relative OKLCH tint and shorter-hue color-mix instead of an RGB wash.
func bubbleLCH(color ui.Color) (l, c, h float64) {
	linear := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= .04045 {
			return x / 12.92
		}
		return math.Pow((x+.055)/1.055, 2.4)
	}
	r, g, b := linear(color.R), linear(color.G), linear(color.B)
	lm := math.Cbrt(.4122214708*r + .5363325363*g + .0514459929*b)
	mm := math.Cbrt(.2119034982*r + .6806995451*g + .1073969566*b)
	sm := math.Cbrt(.0883024619*r + .2817188376*g + .6299787005*b)
	l = .2104542553*lm + .793617785*mm - .0040720468*sm
	a := 1.9779984951*lm - 2.428592205*mm + .4505937099*sm
	v := .0259040371*lm + .7827717662*mm - .808675766*sm
	c = math.Hypot(a, v)
	h = math.Atan2(v, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return
}
func bubbleMix(a, b ui.Color, amount float64) ui.Color {
	al, ac, ah := bubbleLCH(a)
	bl, bc, bh := bubbleLCH(b)
	aa, ba := float64(a.A)/255, float64(b.A)/255
	alpha := aa*(1-amount) + ba*amount
	if alpha == 0 {
		return ui.Transparent
	}
	weight := ba * amount / alpha
	// Achromatic colors have a missing hue and borrow the other's hue.
	if ac < .00001 {
		ah = bh
	}
	if bc < .00001 {
		bh = ah
	}
	delta := math.Mod(bh-ah+540, 360) - 180
	return theme.OKLCH(al+(bl-al)*weight, ac+(bc-ac)*weight, ah+delta*weight, alpha)
}
func bubbleColors(t *theme.Tokens, variant BubbleVariant, hover bool) (bg, fg, border ui.Color) {
	bg, fg, border = t.Primary, t.PrimaryForeground, ui.Transparent
	switch variant {
	case Secondary:
		bg, fg = t.Secondary, t.SecondaryForeground
		if hover {
			bg = bubbleMix(t.Secondary, t.Foreground, .05)
		}
	case Muted:
		bg, fg = t.Muted, t.Foreground
		if hover {
			bg = bubbleMix(t.Muted, t.Foreground, .05)
		}
	case Tinted:
		_, chroma, hue := bubbleLCH(t.Primary)
		lightness, factor := .93, .4
		if t.Dark {
			lightness = .3
		}
		if hover {
			lightness, factor = .88, .5
			if t.Dark {
				lightness = .35
			}
		}
		bg, fg = theme.OKLCH(lightness, chroma*factor, hue, float64(t.Primary.A)/255), t.Foreground
	case Outline:
		bg, fg, border = t.Background, t.Foreground, t.Border
		if hover {
			bg = t.Muted
			if t.Dark {
				bg = t.Input.Alpha(.3)
			}
		}
	case Ghost:
		bg, fg = ui.Transparent, t.Foreground
		if hover {
			bg = t.Muted
			if t.Dark {
				bg = t.Muted.Alpha(.5)
			}
		}
	case Destructive:
		bg, fg = t.Destructive.Alpha(.1), t.Destructive
		if t.Dark {
			bg = t.Destructive.Alpha(.2)
		}
		if hover {
			bg = t.Destructive.Alpha(.2)
			if t.Dark {
				bg = t.Destructive.Alpha(.3)
			}
		}
	default:
		if hover {
			bg = t.Primary.Alpha(.8)
		}
	}
	return
}
