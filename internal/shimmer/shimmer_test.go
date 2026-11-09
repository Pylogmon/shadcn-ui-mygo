package shimmer

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

const blocks = "████████████"

func scalar(v float32) *float32 { return &v }
func shimmerTester(p *Props, dark bool) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		native := *c.Theme()
		native.Dark = dark
		c.SetTheme(&native)
		ui.Column(c).Fill().AlignItems(ui.Start).Background(ui.RGB(255, 255, 255)).Children(func() { Text(c, blocks, *p).Role(ui.RoleText) })
	}, 360, 70, ui.TesterOptions{Dark: dark})
}
func textBounds(t *testing.T, r *ui.Tester, text string) ui.Rect {
	t.Helper()
	n, ok := r.AccessByRole(ui.RoleText, text)
	if !ok {
		t.Fatalf("missing native text %q: %+v", text, r.AccessAll())
	}
	return n.Bounds
}
func rgbNear(a color.RGBA, b ui.Color, tolerance int) bool {
	return abs(int(a.R)-int(b.R)) <= tolerance && abs(int(a.G)-int(b.G)) <= tolerance && abs(int(a.B)-int(b.B)) <= tolerance
}
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
func sampleInterior(t *testing.T, reference *image.RGBA, bounds ui.Rect, x float32, base ui.Color) (int, int) {
	t.Helper()
	best := float32(math.MaxFloat32)
	bx, by := -1, -1
	for y := int(bounds.Y); y < int(bounds.Y+bounds.H); y++ {
		for xx := int(bounds.X); xx < int(bounds.X+bounds.W); xx++ {
			p := reference.RGBAAt(xx, y)
			distance := float32(math.Abs(float64(float32(xx) + .5 - x)))
			if rgbNear(p, base, 0) && distance < best {
				bx, by, best = xx, y, distance
			}
		}
	}
	if bx < 0 {
		t.Fatal("native reference has no fully covered glyph pixels")
	}
	return bx, by
}
func TestPlainReducedAndInstantAreNativePixelsAndIdle(t *testing.T) {
	base := ui.RGBA(20, 80, 160, .5)
	for _, mode := range []string{"inactive", "reduced", "instant"} {
		t.Run(mode, func(t *testing.T) {
			render := func(helper bool) *ui.Tester {
				return ui.NewTesterWith(func(c *ui.Context) {
					if mode == "instant" {
						c.DisableAnimations()
					}
					ui.Column(c).Font("monospace").FontSize(19).Children(func() {
						if helper {
							Text(c, "Native 字形 عربي", Props{Active: mode != "inactive", Base: base}).Role(ui.RoleText)
						} else {
							ui.Text(c, "Native 字形 عربي").TextColor(base).Role(ui.RoleText)
						}
					})
				}, 360, 70, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: mode == "reduced"}})
			}
			actual, plain := render(true), render(false)
			if !bytes.Equal(actual.Image().Pix, plain.Image().Pix) {
				t.Fatal("disabled effect replaced native text pixels/font/alpha")
			}
			bounds := textBounds(t, actual, "Native 字形 عربي")
			if bounds.W <= 0 || bounds.H <= 0 {
				t.Fatal("native shaping/layout missing")
			}
			before := append([]byte(nil), actual.Image().Pix...)
			actual.Advance(3 * time.Second)
			if !bytes.Equal(before, actual.Image().Pix) || actual.FramePending() {
				t.Fatal("plain/reduced/instant effect kept moving or requesting frames")
			}
		})
	}
}
func TestLightSweepMovesSpatialBandWithDefaultTwoSecondLifetime(t *testing.T) {
	base := ui.RGB(20, 80, 160)
	p := Props{Active: true, Base: base, Font: ui.Font{Family: "monospace", Size: 20}, Spread: scalar(25), Angle: scalar(0)}
	r := shimmerTester(&p, false)
	bounds := textBounds(t, r, blocks)
	reference := r.Image()
	r.Advance(500 * time.Millisecond)
	quarter := r.Image()
	r.Advance(500 * time.Millisecond)
	middle := r.Image()
	centroid := func(img *image.RGBA) float64 {
		weighted, total := 0.0, 0.0
		for y := 0; y < int(bounds.H); y++ {
			for x := 0; x < int(bounds.W); x++ {
				before, after := reference.RGBAAt(x, y), img.RGBAAt(x, y)
				delta := float64(int(after.R) - int(before.R))
				if delta > 0 {
					weighted += (float64(x) + .5) * delta
					total += delta
				}
			}
		}
		if total == 0 {
			t.Fatal("shimmer never changed any native glyph coverage")
		}
		return weighted / total
	}
	q, m := centroid(quarter), centroid(middle)
	wantQuarter := -25 + .25*(float64(bounds.W)+50)
	if math.Abs(q-wantQuarter) > 4 || math.Abs(m-float64(bounds.W)*.5) > 4 || m <= q {
		t.Fatalf("source 2s linear sweep geometry q=%v m=%v width=%v expected q=%v", q, m, bounds.W, wantQuarter)
	}
	x, y := sampleInterior(t, reference, bounds, bounds.W/2, base)
	got := middle.RGBAAt(x, y)
	want := ui.RGB(208, 220, 236)
	if !rgbNear(got, want, 6) {
		t.Fatalf("light center is not alpha*.2: got=%v want≈%v", got, want)
	}
	leftX, leftY := sampleInterior(t, reference, bounds, 5, base)
	if !rgbNear(middle.RGBAAt(leftX, leftY), base, 1) {
		t.Fatal("effect faded the whole text instead of a moving band")
	}
}
func TestDarkHighlightDerivesOKLCHLightnessAndAddsAlpha(t *testing.T) {
	for _, alpha := range []float32{1, .5, 0} {
		t.Run(string(rune('a'+int(alpha*2))), func(t *testing.T) {
			base := ui.RGBA(51, 51, 51, alpha)
			p := Props{Active: true, Base: base, Font: ui.Font{Family: "monospace", Size: 20}, Spread: scalar(50), Angle: scalar(0), Duration: 100 * time.Millisecond}
			r := shimmerTester(&p, true)
			bounds := textBounds(t, r, blocks)
			referenceProps := p
			referenceProps.Base = ui.RGB(51, 51, 51)
			referenceProps.Active = false
			reference := shimmerTester(&referenceProps, true).Image()
			x, y := sampleInterior(t, reference, bounds, bounds.W/2, ui.RGB(51, 51, 51))
			r.Advance(50 * time.Millisecond)
			r.Frame()
			got := r.Image().RGBAAt(x, y)
			// 51 gray has L≈.321: max(.8,L+.4)=.8 -> sRGB190.
			outAlpha := math.Min(1, float64(base.A)/255+.4)
			channel := uint8(math.Round(190*outAlpha + 255*(1-outAlpha)))
			want := ui.RGB(channel, channel, channel)
			if !rgbNear(got, want, 3) {
				t.Fatalf("dark OKLCH/alpha+.4 mismatch: base=%v got=%v want≈%v", base, got, want)
			}
		})
	}
}
func TestDirectionReverseOnceAndDurationAreVisibleAndSettle(t *testing.T) {
	makeTester := func(direction ui.Direction, reverse, active bool) *ui.Tester {
		p := Props{Active: active, Base: ui.RGB(0, 0, 0), Font: ui.Font{Family: "monospace", Size: 20}, Spread: scalar(25), Angle: scalar(0), Duration: 100 * time.Millisecond, Reverse: reverse}
		return ui.NewTester(func(c *ui.Context) {
			ui.Column(c).Direction(direction).AlignItems(ui.Start).Children(func() { Text(c, blocks, p).Role(ui.RoleText) })
		}, 360, 70)
	}
	ltr, rtl, reversed := makeTester(ui.LeftToRight, false, true), makeTester(ui.RightToLeft, false, true), makeTester(ui.LeftToRight, true, true)
	for _, r := range []*ui.Tester{ltr, rtl, reversed} {
		r.Advance(25 * time.Millisecond)
		r.Frame()
	}
	centroid := func(r, reference *ui.Tester) float64 {
		img, plain := r.Image(), reference.Image()
		bounds := textBounds(t, r, blocks)
		sum, total := 0.0, 0.0
		for y := int(bounds.Y); y < int(bounds.Y+bounds.H); y++ {
			for x := int(bounds.X); x < int(bounds.X+bounds.W); x++ {
				d := float64(int(img.RGBAAt(x, y).R) - int(plain.RGBAAt(x, y).R))
				if d > 0 {
					sum += (float64(x) + .5 - float64(bounds.X)) * d
					total += d
				}
			}
		}
		if total == 0 {
			t.Fatal("reverse test has no moving band")
		}
		return sum / total
	}
	ltrReference, rtlReference := makeTester(ui.LeftToRight, false, false), makeTester(ui.RightToLeft, false, false)
	lc, rc, vc := centroid(ltr, ltrReference), centroid(rtl, rtlReference), centroid(reversed, ltrReference)
	if math.Abs(lc-vc) < 20 {
		t.Fatal("Reverse had no spatial effect")
	}
	if math.Abs(rc-vc) > 1 {
		t.Fatalf("RTL band did not follow reverse: rtl=%v reverse=%v", rc, vc)
	}
	p := Props{Active: true, Base: ui.RGB(20, 80, 160), Font: ui.Font{Size: 20}, Spread: scalar(25), Angle: scalar(0), Duration: 100 * time.Millisecond, Once: true}
	r := shimmerTester(&p, false)
	r.Advance(110 * time.Millisecond)
	r.Frame()
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(time.Second)
	if !bytes.Equal(before, r.Image().Pix) || r.FramePending() {
		t.Fatal("Once did not stop at its source end position")
	}
}

func TestDefaultSpreadTracksThreeNativeZeroAdvancesPlusForty(t *testing.T) {
	font := ui.Font{Family: "monospace", Size: 20, Weight: 500}
	zero := ui.ShapeText("0", font, 0)
	w, _ := zero.Size()
	spread := 3*w + 40
	a := Props{Active: true, Base: ui.RGB(40, 80, 120), Font: font, Duration: 100 * time.Millisecond}
	b := a
	b.Spread = &spread
	ar, br := shimmerTester(&a, false), shimmerTester(&b, false)
	for _, r := range []*ui.Tester{ar, br} {
		r.Advance(50 * time.Millisecond)
		r.Frame()
	}
	if !bytes.Equal(ar.Image().Pix, br.Image().Pix) {
		t.Fatal("default band did not use actual 3ch + 40px and source background size")
	}
}
func TestDefaultAngleIsTwentyDegreesAcrossNativeGlyphRows(t *testing.T) {
	base := ui.RGB(0, 0, 0)
	p := Props{Active: true, Base: base, Font: ui.Font{Family: "monospace", Size: 28}, Spread: scalar(30), Duration: 100 * time.Millisecond}
	r := shimmerTester(&p, false)
	bounds := textBounds(t, r, blocks)
	plain := p
	plain.Active = false
	reference := shimmerTester(&plain, false).Image()
	r.Advance(50 * time.Millisecond)
	r.Frame()
	img := r.Image()
	rows := []int{}
	for y := 0; y < int(bounds.H); y++ {
		count := 0
		for x := 0; x < int(bounds.W); x++ {
			if rgbNear(reference.RGBAAt(x, y), base, 0) {
				count++
			}
		}
		if count > int(bounds.W*.7) {
			rows = append(rows, y)
		}
	}
	if len(rows) < 10 {
		t.Fatal("native block glyphs did not cover enough rows")
	}
	top, bottom := rows[2], rows[len(rows)-3]
	centroid := func(y int) float64 {
		sum, total := 0.0, 0.0
		for x := 0; x < int(bounds.W); x++ {
			delta := float64(int(img.RGBAAt(x, y).R) - int(reference.RGBAAt(x, y).R))
			if delta > 0 {
				sum += (float64(x) + .5) * delta
				total += delta
			}
		}
		if total == 0 {
			t.Fatal("angled band missing")
		}
		return sum / total
	}
	difference := centroid(top) - centroid(bottom)
	want := float64(bottom-top) * math.Tan(20*math.Pi/180)
	if math.Abs(difference-want) > 1.5 {
		t.Fatalf("source20deg band tilt=%v expected≈%v", difference, want)
	}
}
func TestCustomHighlightUsesPerceptualMidpointAndIndependentAlpha(t *testing.T) {
	// Both input colors are in-gamut, so their public RGB coordinates preserve
	// the source's independently calculable OKLCH hue/lightness reference.
	base := ui.RGB(128, 128, 128)
	highlight := ui.RGBA(0, 80, 180, .6)
	p := Props{Active: true, Base: base, Highlight: &highlight, Font: ui.Font{Family: "monospace", Size: 20}, Angle: scalar(0), Spread: scalar(45), Duration: 100 * time.Millisecond}
	r := shimmerTester(&p, false)
	bounds := textBounds(t, r, blocks)
	referenceProps := p
	referenceProps.Active = false
	reference := shimmerTester(&referenceProps, false).Image()
	r.Advance(50 * time.Millisecond)
	r.Frame()
	img := r.Image()
	x, y := sampleInterior(t, reference, bounds, bounds.W/2, base)
	got := img.RGBAAt(x, y)
	want := highlight.Over(ui.RGB(255, 255, 255))
	if !rgbNear(got, want, 4) {
		t.Fatalf("custom highlight color/alpha missing: got=%v want≈%v", got, want)
	}
	// The half-band combines colors in OKLCH; it must retain blue chroma and
	// differs visibly from simply fading a same-color text overlay.
	mx, my := sampleInterior(t, reference, bounds, bounds.W/2-22.5, base)
	mid := img.RGBAAt(mx, my)
	if mid.B <= mid.R+15 || mid.B <= mid.G+10 {
		t.Fatalf("custom perceptual midpoint lost blue chroma: %v", mid)
	}
}
func TestSpreadPercentAndAngleOverridesChangeActualBand(t *testing.T) {
	p := Props{Active: true, Base: ui.RGB(0, 0, 0), Font: ui.Font{Family: "monospace", Size: 20}, Angle: scalar(0), Spread: scalar(8), Duration: 100 * time.Millisecond}
	q := p
	q.SpreadPercent = scalar(15)
	a, b := shimmerTester(&p, false), shimmerTester(&q, false)
	for _, r := range []*ui.Tester{a, b} {
		r.Advance(50 * time.Millisecond)
		r.Frame()
	}
	if bytes.Equal(a.Image().Pix, b.Image().Pix) {
		t.Fatal("percentage band was only stored as configuration")
	}
	bounds := textBounds(t, a, blocks)
	count := func(img *image.RGBA) int {
		n := 0
		for y := 0; y < int(bounds.H); y++ {
			for x := 0; x < int(bounds.W); x++ {
				pixel := img.RGBAAt(x, y)
				if pixel.R > 30 && pixel.R < 240 {
					n++
				}
			}
		}
		return n
	}
	if count(b.Image()) <= count(a.Image())*2 {
		t.Fatal("percentage spread did not expand the source gradient band")
	}
}
func TestNativeEllipsisFontInheritanceAndAXSurviveEffectsAndPreferenceChanges(t *testing.T) {
	active := true
	measure := ui.Font{Family: "monospace", Size: 14, Weight: 500}
	p := Props{Active: true, Base: ui.RGB(40, 80, 120), Font: ui.Font{Weight: 500}, MeasureFont: &measure, Duration: 100 * time.Millisecond}
	text := "Native shaping 字形 mixed text that must truncate"
	mounts := 0
	r := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Font("monospace").FontSize(24).Children(func() {
			p.Active = active
			e := Text(c, text, p).Width(160).SingleLine().Ellipsis("…").Role(ui.RoleText)
			ui.Local(e, "native identity", func() int { mounts++; return 0 })
		})
	}, 300, 90)
	bounds := textBounds(t, r, text)
	if bounds.W != 160 || bounds.H < 25 {
		t.Fatalf("effect bypassed fluent native ellipsis/font inheritance: %+v", bounds)
	}
	r.Advance(30 * time.Millisecond)
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(time.Second)
	if !bytes.Equal(before, r.Image().Pix) || r.FramePending() {
		t.Fatal("dynamic reduced preference did not restore native plain/idle")
	}
	active = false
	r.Frame()
	if mounts != 1 || textBounds(t, r, text) != bounds {
		t.Fatal("activity/preferences remounted native text or changed its font/layout")
	}
	// A reference native node confirms this is true host ellipsis, rather than
	// the helper slicing text and inventing its own visual string.
	reference := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Font("monospace").FontSize(24).Children(func() {
			ui.Text(c, text).TextColor(p.Base).FontWeight(500).Width(160).SingleLine().Ellipsis("…").Role(ui.RoleText)
		})
	}, 300, 90)
	if !bytes.Equal(r.Image().Pix, reference.Image().Pix) {
		t.Fatal("inactive helper did not return exact native ellipsis pixels")
	}
}
func TestInvalidOptionsFallBackToFiniteNativeText(t *testing.T) {
	p := Props{Active: true, Base: ui.RGB(30, 60, 90), Angle: scalar(float32(math.NaN())), Spread: scalar(float32(math.Inf(1))), SpreadPercent: scalar(-1), Duration: -time.Second}
	r := shimmerTester(&p, false)
	n, ok := r.AccessByRole(ui.RoleText, blocks)
	if !ok || !n.Invalid || n.Bounds.W <= 0 || n.Bounds.H <= 0 {
		t.Fatalf("invalid options lost finite fallback/AX: %+v", n)
	}
	r.Advance(50 * time.Millisecond)
	p.Active = false
	r.Frame()
	r.Advance(50 * time.Millisecond)
	if r.FramePending() {
		t.Fatal("disabling invalid-option fallback left an animation alive")
	}
}

func TestReverseOnceReturnsToUnderlyingZeroBackgroundPosition(t *testing.T) {
	p := Props{Active: true, Base: ui.RGB(10, 40, 90), Font: ui.Font{Family: "monospace", Size: 20}, Duration: 100 * time.Millisecond, Angle: scalar(70), Once: true}
	q := p
	q.Reverse = true
	a, b := shimmerTester(&p, false), shimmerTester(&q, false)
	for _, r := range []*ui.Tester{a, b} {
		r.Advance(110 * time.Millisecond)
		r.Frame()
	}
	// shimmer-once has no animation-fill-mode. Both directions therefore fall
	// back to background-position:0 0 after their single iteration.
	if !bytes.Equal(a.Image().Pix, b.Image().Pix) {
		t.Fatal("reverse-once retained animation's start position rather than source underlying background position")
	}
}

func TestDarkChromaHighlightUsesPerceptualGamutMapping(t *testing.T) {
	// Public Oklch.SRGB promises CSS Color 4 gamut mapping. Compare the actual
	// shader pixels with that independent host implementation at a highlight
	// which lies far outside sRGB; clipping RGB channels is visibly different.
	base := ui.RGB(0, 70, 180)
	p := Props{Active: true, Base: base, Font: ui.Font{Family: "monospace", Size: 20}, Duration: 100 * time.Millisecond, Angle: scalar(0), Spread: scalar(120)}
	r := shimmerTester(&p, true)
	bounds := textBounds(t, r, blocks)
	plain := p
	plain.Active = false
	reference := shimmerTester(&plain, true).Image()
	x, y := sampleInterior(t, reference, bounds, bounds.W/2, base)
	r.Advance(50 * time.Millisecond)
	r.Frame()
	_, _, high := colors(base, nil, true)
	chroma := float32(math.Hypot(float64(high[1]), float64(high[2])))
	hue := float32(math.Atan2(float64(high[2]), float64(high[1])) * 180 / math.Pi)
	want := ui.Oklch(high[0], chroma, hue).SRGB()
	got := r.Image().RGBAAt(x, y)
	if !rgbNear(got, want, 5) {
		t.Fatalf("dark wide-chroma highlight was channel-clipped or remixed in RGB: got=%v mapped=%v", got, want)
	}
}
