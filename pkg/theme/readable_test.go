package theme

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestReadableForegroundAllPresetsAndSurfaces(t *testing.T) {
	for _, name := range Presets() {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dark=%v", name, dark), func(t *testing.T) {
				tokens := Preset(name, dark)
				original := *tokens
				for _, surface := range []ui.Color{tokens.Background, tokens.Sidebar, tokens.Card, tokens.Popover,
					ui.RGB(230, 209, 181), ui.RGB(45, 69, 90), ui.RGB(120, 120, 120)} {
					for _, accent := range []ui.Color{tokens.Primary, tokens.SidebarPrimary, tokens.Destructive,
						ui.RGB(22, 163, 74), ui.RGB(255, 200, 30), ui.RGB(100, 10, 180),
						ui.RGB(0, 0, 0), ui.RGB(255, 255, 255), ui.Transparent, ui.RGBA(120, 40, 220, .25)} {
						for _, alpha := range []float32{0, .1, .2, .3} {
							bg := Composite(accent.Alpha(alpha), surface)
							got := EnsureContrast(accent, bg, 4.5)
							if ratio := Contrast(Composite(got, bg), bg); ratio < 4.5 {
								t.Fatalf("accent=%v surface=%v alpha=%v foreground=%v contrast=%.6f", accent, surface, alpha, got, ratio)
							}
							if Contrast(Composite(accent, bg), bg) >= 4.5 && got != accent {
								t.Fatalf("already readable accent changed: %v -> %v", accent, got)
							}
						}
					}
				}
				if *tokens != original || *Preset(name, dark) != original {
					t.Fatal("readable foreground changed official tokens")
				}
			})
		}
	}
}

func TestEnsureContrastMakesOnlyNecessaryFamilyAdjustment(t *testing.T) {
	for _, bg := range []ui.Color{ui.RGB(255, 255, 255), ui.RGB(20, 20, 20)} {
		for _, source := range []ui.Color{ui.RGB(170, 105, 0), ui.RGB(100, 160, 180), ui.RGBA(120, 40, 220, .4)} {
			got := EnsureContrast(source, bg, 4.5)
			// Independent coarse reference: no smaller passing mix on either
			// endpoint path should exist. Keep a one-step quantization allowance.
			bestStep := 1001
			for _, target := range []ui.Color{ui.RGB(0, 0, 0), ui.RGB(255, 255, 255)} {
				for i := 0; i <= 1000; i++ {
					candidate := Mix(source, target, float32(i)/1000)
					if Contrast(Composite(candidate, bg), bg) >= 4.5 {
						bestStep = min(bestStep, i)
						break
					}
				}
			}
			found := false
			for _, target := range []ui.Color{ui.RGB(0, 0, 0), ui.RGB(255, 255, 255)} {
				candidate := Mix(source, target, float32(bestStep)/1000)
				if colorDistance(got, candidate) <= 3 {
					found = true
				}
			}
			if !found {
				t.Fatalf("excessive color adjustment: source=%v bg=%v got=%v first step=%d", source, bg, got, bestStep)
			}
		}
	}
}

func colorDistance(a, b ui.Color) int {
	abs := func(n int) int { return max(n, -n) }
	return abs(int(a.R)-int(b.R)) + abs(int(a.G)-int(b.G)) + abs(int(a.B)-int(b.B)) + abs(int(a.A)-int(b.A))
}

func TestReadableSurfaceCompositionAndIsolation(t *testing.T) {
	tokens := Light()
	tokens.Background = ui.RGB(255, 255, 255)
	outer, inner := ui.RGBA(0, 0, 0, .8), ui.RGBA(200, 60, 0, .2)
	ui.NewTester(func(c *ui.Context) {
		c.SetTheme(tokens.Theme())
		original := SurfaceOf(c)
		WithSurface(c, outer, func() {
			want := Composite(outer, original)
			if got := SurfaceOf(c); got != want {
				t.Fatalf("outer=%v want %v", got, want)
			}
			WithSurface(c, inner, func() {
				bg := Composite(inner, want)
				if SurfaceOf(c) != bg || ResolveSurface(c, &inner) != Composite(inner, bg) {
					t.Fatal("nested/explicit alpha composition lost enclosing surface")
				}
				fg := ReadableForeground(c, tokens.Primary, ui.Transparent)
				if Contrast(Composite(fg, bg), bg) < 4.5 {
					t.Fatal("used the page background instead of the carrying surface")
				}
			})
			func() {
				defer func() { _ = recover() }()
				WithSurface(c, inner, func() { panic("scope restoration") })
			}()
			if SurfaceOf(c) != want {
				t.Fatal("nested/panicking scope leaked")
			}
		})
		if SurfaceOf(c) != original {
			t.Fatal("surface leaked to sibling")
		}
	}, 100, 100)
}

func TestReadableHighContrastAndUnattainableTarget(t *testing.T) {
	for _, high := range []bool{false, true} {
		ui.NewTesterWith(func(c *ui.Context) {
			c.SetTheme(Dark().Theme())
			bg := ui.RGB(30, 30, 30)
			fg := ReadableForeground(c, ui.RGB(170, 105, 0), bg)
			minimum := 4.5
			if high {
				minimum = 7
			}
			if Contrast(Composite(fg, bg), bg) < minimum {
				t.Fatal("ignored high contrast preference")
			}
		}, 100, 100, ui.TesterOptions{Preferences: ui.Preferences{HighContrast: high}})
	}
	gray := ui.RGB(120, 120, 120)
	got := EnsureContrast(ui.RGB(120, 120, 120), gray, 7)
	if got != ui.RGB(0, 0, 0) || Contrast(got, gray) < 4.5 {
		t.Fatalf("unattainable target should choose best endpoint: %v", got)
	}
	for _, minimum := range []float64{0, math.NaN(), math.Inf(1)} {
		if EnsureContrast(gray, ui.RGB(255, 255, 255), minimum) != EnsureContrast(gray, ui.RGB(255, 255, 255), 4.5) {
			t.Fatal("invalid minimum did not use 4.5")
		}
	}
}
