package display

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestAspectRatioRejectsNonfiniteAndUnrepresentableGeometry(t *testing.T) {
	for _, ratio := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1)), 0, -1, math.SmallestNonzeroFloat32} {
		r := ui.NewTester(func(c *ui.Context) {
			ui.Column(c).Width(240).Children(func() { AspectRatio(c, ratio, nil).Label("Ratio") })
		}, 300, 300)
		box, ok := r.Find("Ratio")
		if !ok || math.IsNaN(float64(box.W)) || math.IsNaN(float64(box.H)) || math.IsInf(float64(box.W), 0) || math.IsInf(float64(box.H), 0) || box.H != 240 {
			t.Fatalf("ratio=%g did not fall back to a finite square: %+v", ratio, box)
		}
	}
}

func aspectPicture() *ui.Bitmap {
	im := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			pixel := color.RGBA{R: 220, A: 255}
			if x >= 20 {
				pixel = color.RGBA{B: 220, A: 255}
			}
			im.SetRGBA(x, y, pixel)
		}
	}
	return ui.NewBitmap(im)
}

func TestAspectRatioSourceImageShapesPaintCoverWithinActualBounds(t *testing.T) {
	for _, example := range []struct{ ratio, width, height float32 }{{16.0 / 9, 160, 90}, {21.0 / 9, 210, 90}, {1, 120, 120}, {9.0 / 16, 90, 160}} {
		r := ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			ui.Column(c).Padding(20).Children(func() {
				AspectRatioWith(c, AspectRatioProps{Ratio: example.ratio, Width: example.width, Label: "Photo frame", Clip: true}, func(child *ui.Context) {
					ui.Image(child, aspectPicture()).Fill().Fit(ui.Cover).Label("Photo")
				})
			})
		}, 300, 260)
		frame, _ := r.Find("Photo frame")
		photo, ok := r.AccessByRole(ui.RoleImage, "Photo")
		if !ok || frame.W != example.width || math.Abs(float64(frame.H-example.height)) > .01 || photo.Bounds != frame {
			t.Fatalf("ratio%g did not size the native image: frame=%+v image=%+v", example.ratio, frame, photo)
		}
		im := r.Image()
		left := im.RGBAAt(int(frame.X+frame.W/4), int(frame.Y+frame.H/2))
		right := im.RGBAAt(int(frame.X+3*frame.W/4), int(frame.Y+frame.H/2))
		if left.R != 220 || left.B != 0 || right.B != 220 || right.R != 0 {
			t.Fatalf("ratio%g cover pixels: left=%v right=%v", example.ratio, left, right)
		}
	}
}

func TestAspectRatioWithAllowsSourceOverflowAndLegacyRemainsClipped(t *testing.T) {
	for _, mode := range []string{"visible", "clipped", "legacy"} {
		r := ui.NewTester(func(c *ui.Context) {
			ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() {
				content := func(child *ui.Context) {
					ui.Box(child).Absolute().Left(90).Top(20).Size(40, 20).Background(ui.RGB(230, 20, 40))
				}
				if mode == "legacy" {
					AspectRatio(c, 1, func() { content(c) }).Width(100)
				} else {
					AspectRatioWith(c, AspectRatioProps{Ratio: 1, Width: 100, Clip: mode == "clipped"}, content)
				}
			})
		}, 200, 180)
		pixel := r.Image().RGBAAt(140, 50)
		if mode == "visible" && pixel.R != 230 || mode != "visible" && pixel != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
			t.Fatalf("overflow mode%s pixel=%v", mode, pixel)
		}
	}
}

func TestAspectRatioWidthAndHeightConstraintsAreFiniteAndExplicit(t *testing.T) {
	for _, p := range []AspectRatioProps{
		{Ratio: 2, Width: 200, MaxWidth: 120, Label: "Constrained"},
		{Ratio: 1, Width: 120, MaxHeight: 40, Label: "Constrained"},
		{Ratio: float32(math.Inf(1)), Width: 120, Label: "Constrained"},
		{Ratio: 1, Width: float32(math.NaN()), MaxWidth: 120, Label: "Constrained"},
	} {
		r := ui.NewTester(func(c *ui.Context) {
			ui.Column(c).Width(240).Children(func() { AspectRatioWith(c, p, nil) })
		}, 300, 300)
		frame, ok := r.AccessByRole(ui.RoleGroup, "Constrained")
		if !ok || !aspectFinite(frame.Bounds.W) || !aspectFinite(frame.Bounds.H) {
			t.Fatalf("constraint produced nonfinite bounds: %+v", frame)
		}
		if p.MaxWidth > 0 && frame.Bounds.W != 120 || p.MaxHeight > 0 && frame.Bounds.H != 40 || p.Ratio == 2 && frame.Bounds.H != 60 {
			t.Fatalf("explicit native bounds ignored: props=%+v frame=%+v", p, frame.Bounds)
		}
		if (!aspectFinite(p.Ratio) || !aspectFinite(p.Width)) && !frame.Invalid {
			t.Fatal("invalid native geometry lost its semantic error state")
		}
	}
}

func TestAspectRatioResizeAndEmptyContentKeepNativeChildAndFollowingIdentity(t *testing.T) {
	width, ratio, show := float32(160), float32(16.0/9), false
	inside, following, created := "inside", "following", 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			AspectRatioWith(c, AspectRatioProps{Ratio: ratio, Width: width, Label: "Resizable ratio"}, func(child *ui.Context) {
				owner := ui.Column(child)
				ui.Local(owner, "native editor", func() bool { created++; return true })
				owner.Children(func() { control.Input(child, &inside, control.InputProps{Label: "Inside ratio"}) })
			})
			AspectRatioWith(c, AspectRatioProps{Ratio: 2, Width: width}, func(child *ui.Context) {
				if show {
					ui.Text(child, "Optional caption")
				}
			})
			control.Input(c, &following, control.InputProps{Label: "Following ratio"})
		})
	}, 400, 500)
	displayClick(t, r, "Inside ratio")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("saved native draft")
	width, ratio = 240, 2
	r.Frame()
	if created != 1 || !r.Focused("Inside ratio") || inside != "saved native draft" {
		t.Fatal("ratio/width update recreated the native editor")
	}
	frame, _ := r.Find("Resizable ratio")
	if frame.W != 240 || frame.H != 120 {
		t.Fatalf("controlled ratio failed to resize: %+v", frame)
	}
	displayClick(t, r, "Following ratio")
	show = true
	r.Frame()
	if !r.Focused("Following ratio") || following != "following" {
		t.Fatal("empty ratio callback leaked a pending widget identity")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("a stable ratio retained an idle frame")
	}
}
