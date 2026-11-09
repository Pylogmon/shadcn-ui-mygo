package shimmer

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"
	"testing"
)

func TestNativeMaskPreservesShapedTextAndAccessibility(t *testing.T) {
	text := "MMMM MMMM MMMM"
	r := ui.NewTester(func(c *ui.Context) {
		ui.Text(c, text).Font("monospace").FontSize(28).TextColor(ui.RGB(20, 80, 160)).Role(ui.RoleText).MaskLinearGradient(ui.LinearGradient{Angle: 90, Stops: []ui.GradientStop{{Color: ui.RGBA(0, 0, 0, 1), At: -1}, {Color: ui.RGBA(0, 0, 0, 1), At: .3}, {Color: ui.RGBA(0, 0, 0, 0), At: .5}, {Color: ui.RGBA(0, 0, 0, 1), At: .7}, {Color: ui.RGBA(0, 0, 0, 1), At: 2}}})
	}, 400, 100)
	n, ok := r.AccessByRole(ui.RoleText, text)
	if !ok || n.Bounds.W <= 0 {
		t.Fatal("native mask lost real text AX/layout")
	}
	img := r.Image()
	edge, mid := 0, 0
	for y := int(n.Bounds.Y); y < int(n.Bounds.Y+n.Bounds.H); y++ {
		for x := 0; x < int(n.Bounds.W); x++ {
			p := img.RGBAAt(x, y)
			if p.B > p.R+20 {
				if float32(x) < n.Bounds.W*.25 {
					edge++
				}
				if float32(x) > n.Bounds.W*.45 && float32(x) < n.Bounds.W*.55 {
					mid++
				}
			}
		}
	}
	if edge == 0 || mid >= edge {
		t.Fatalf("native arbitrary-stop mask did not paint a selective band: edge=%d middle=%d", edge, mid)
	}
}
func TestNativeColorEffectRecolorsGlyphCoverageWithoutPaintingBackground(t *testing.T) {
	effect := shader.MustCompile("package main\nfunc ColorEffect(pos vec2,color vec4)vec4{return vec4(color.a,0,0,color.a)}\n")
	text := "Native 字形"
	r := ui.NewTester(func(c *ui.Context) {
		ui.Text(c, text).FontSize(28).TextColor(ui.RGB(20, 80, 160)).Role(ui.RoleText).Effect(effect.Effect(shader.Options{}))
	}, 400, 100)
	n, ok := r.AccessByRole(ui.RoleText, text)
	if !ok || n.Bounds.W <= 0 {
		t.Fatal("public color effect replaced native text semantics")
	}
	image := r.Image()
	red, white := 0, 0
	for y := 0; y < int(n.Bounds.H); y++ {
		for x := 0; x < int(n.Bounds.W); x++ {
			p := image.RGBAAt(x, y)
			if p.R > 200 && p.G < 80 && p.B < 80 {
				red++
			}
			if p.R == 255 && p.G == 255 && p.B == 255 {
				white++
			}
		}
	}
	if red == 0 || white == 0 {
		t.Fatalf("effect did not preserve glyph coverage: red=%d empty=%d", red, white)
	}
}
