package material_test

import (
	"bytes"
	"image"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/rounded"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/sharp"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func redPixels(im *image.RGBA) image.Rectangle {
	b := image.Rectangle{Min: im.Bounds().Max, Max: im.Bounds().Min}
	for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
			p := im.RGBAAt(x, y)
			if int(p.R)-int(max(p.G, p.B)) > 100 {
				b.Min.X, b.Min.Y = min(b.Min.X, x), min(b.Min.Y, y)
				b.Max.X, b.Max.Y = max(b.Max.X, x+1), max(b.Max.Y, y+1)
			}
		}
	}
	return b
}

func TestSymbolUsesNativeVectorColorAndFixedLayout(t *testing.T) {
	for _, size := range []float32{16, 24, 32} {
		t.Run(fmtSize(size), func(t *testing.T) {
			var images []*image.RGBA
			for _, font := range []float32{8, 38} {
				r := ui.NewTester(func(c *ui.Context) {
					ui.Column(c).Fill().Padding(10).Background(ui.RGB(255, 255, 255)).TextColor(ui.RGB(255, 0, 0)).FontSize(font).LineHeight(2).Children(func() {
						material.Icon(c, outlined.Home(), material.Props{Size: size, Label: "Home symbol"})
					})
				}, 100, 100)
				n, ok := r.AccessByRole(ui.RoleImage, "Home symbol")
				if !ok || n.Bounds.W != size || n.Bounds.H != size {
					t.Fatalf("symbol box = %+v, found=%v", n, ok)
				}
				ink := redPixels(r.Image())
				if ink.Empty() || ink.Dx() > int(size) || ink.Dy() > int(size) {
					t.Fatalf("native SVG did not inherit red within its fixed box: %v", ink)
				}
				images = append(images, r.Image())
			}
			if !bytes.Equal(images[0].Pix, images[1].Pix) {
				t.Fatal("surrounding font size or line height changed the symbol")
			}
		})
	}
}

func fmtSize(v float32) string {
	return map[float32]string{16: "16DIP", 24: "24DIP", 32: "32DIP"}[v]
}

func TestSymbolStylesAndFillUseDifferentNativeShapes(t *testing.T) {
	sources := []*ui.SVG{outlined.Home(), outlined.Home(material.Filled), rounded.Home(), sharp.Home()}
	images := []*image.RGBA{}
	for _, svg := range sources {
		w, h := svg.Size()
		if w != 24 || h != 24 {
			t.Fatalf("official source dimensions = %v,%v", w, h)
		}
		images = append(images, svg.Image(48, 48, nil))
	}
	// Google's home outline and sharp sources are intentionally identical;
	// fill and the rounded family change this particular symbol's geometry.
	for i := 1; i < 3; i++ {
		if bytes.Equal(images[0].Pix, images[i].Pix) {
			t.Fatalf("variant%d did not change the native vector shape", i)
		}
	}
	if outlined.Home() != outlined.Home() {
		t.Fatal("requested immutable SVG was parsed repeatedly")
	}
}

func TestSymbolComposesInRealNativeButton(t *testing.T) {
	clicks := 0
	r := ui.NewTester(func(c *ui.Context) {
		c.DisableAnimations()
		theme.Apply(c)
		ui.Column(c).Padding(10).Children(func() {
			if control.Button(c, "Save", control.ButtonProps{Icon: outlined.Save()}).Clicked() {
				clicks++
			}
			material.Icon(c, rounded.Favorite(material.Filled))
		})
	}, 240, 140)
	b, ok := r.AccessByRole(ui.RoleButton, "Save")
	if !ok {
		t.Fatal("symbol replaced the native button semantics")
	}
	r.ClickAt(b.Bounds.X+14, b.Bounds.Y+b.Bounds.H/2)
	if clicks != 1 {
		t.Fatalf("icon area did not activate its real button: %d", clicks)
	}
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleImage {
			t.Fatal("decorative symbol added an unnamed accessibility image")
		}
	}
}

func TestSymbolRejectsNonfiniteLayout(t *testing.T) {
	for _, size := range []float32{-1, float32(math.NaN()), float32(math.Inf(1))} {
		r := ui.NewTester(func(c *ui.Context) {
			material.Icon(c, outlined.Check(), material.Props{Size: size, Label: "Check"})
		}, 80, 80)
		n, ok := r.AccessByRole(ui.RoleImage, "Check")
		if !ok || n.Bounds.W != 24 || n.Bounds.H != 24 {
			t.Fatalf("invalid size did not use a finite default: %+v", n)
		}
	}
}
