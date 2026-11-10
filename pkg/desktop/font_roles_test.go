package desktop_test

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"math"
	"testing"
)

type fontProbeSurface struct {
	view        func(*ui.Context)
	invalidates int
}

func (f *fontProbeSurface) SetAnchor(ui.Rect) {}
func (f *fontProbeSurface) Invalidate()       { f.invalidates++ }
func (f *fontProbeSurface) Close()            {}

type fontProbeBackend struct{ child *fontProbeSurface }

func (b *fontProbeBackend) Open(_ desktop.Spec, view func(*ui.Context), _ func()) (desktop.Surface, error) {
	b.child = &fontProbeSurface{view: view}
	return b.child, nil
}
func (b *fontProbeBackend) InvalidateParent() {}
func TestPublicStyledFlyoutInheritsActualNativeMono(t *testing.T) {
	b := &fontProbeBackend{}
	host := desktop.NewHostWithBackend(b)
	open := true
	family := "monospace"
	native := theme.Light().Theme()
	sample := "iiiiiiiiWWWWWWWW01234567"
	var parentWidth, childWidth float32
	var childFamily string
	parent := ui.NewTester(func(c *ui.Context) {
		native.Font = family
		c.SetTheme(native)
		host.BeginFrame()
		defer host.EndFrame()
		parentWidth, _ = c.MeasureText(1000, ui.Span{Text: sample, Size: 20})
		anchor := ui.Button(c, "Open font probe")
		if err := host.Popover(c, "font-probe", anchor, &open, desktop.PopoverProps{Plain: true, Width: 500, ManualTrigger: true}, func(child *ui.Context) {
			childFamily = child.Theme().Font
			childWidth, _ = child.MeasureText(1000, ui.Span{Text: sample, Size: 20})
			ui.Row(child).Children(func() { ui.Text(child, sample).FontSize(20).TextColor(ui.RGB(255, 0, 0)).Role(ui.RoleHeading) })
		}); err != nil {
			t.Fatal(err)
		}
	}, 500, 150)
	if b.child == nil {
		t.Fatal("no styled child")
	}
	child := ui.NewTester(b.child.view, 500, 150)
	_ = child.Image()
	t.Logf("native public measurement: parent family=%q width=%g child family=%q width=%g", native.Font, parentWidth, childFamily, childWidth)
	if childFamily != "monospace" || math.Abs(float64(parentWidth-childWidth)) > .1 {
		t.Errorf("styled child lost the native body family")
	}
	parent.Frame()
	parent.Frame()
	b.child.invalidates = 0
	family = "Georgia"
	parent.Frame()
	child.Frame()
	t.Logf("font-only update: invalidates=%d parent width=%g child family=%q width=%g", b.child.invalidates, parentWidth, childFamily, childWidth)
	if b.child.invalidates != 1 || childFamily != "Georgia" || math.Abs(float64(parentWidth-childWidth)) > .1 {
		t.Errorf("font-only update did not refresh the existing child")
	}
}

func TestPublicNativeFlyoutKeepsPlatformFont(t *testing.T) {
	b := &fontProbeBackend{}
	host := desktop.NewHostWithBackend(b)
	open := true
	fonts := theme.FontRoles{Body: "monospace", Heading: "Georgia"}
	controller := theme.Controller{Fonts: &fonts}
	var actual theme.FontRoles
	parent := ui.NewTester(func(c *ui.Context) {
		controller.Apply(c)
		host.BeginFrame()
		defer host.EndFrame()
		anchor := ui.Button(c, "Open platform font probe")
		if err := desktop.NativePopover(c, host, "native-font", anchor, &open, desktop.PopoverProps{ManualTrigger: true}, func(child *ui.Context) {
			actual = theme.FontsOf(child)
			ui.Text(child, "Native family")
		}); err != nil {
			t.Fatal(err)
		}
	}, 500, 150)
	platform := *ui.LightTheme()
	platform.Font = "Georgia"
	child := ui.NewTester(func(c *ui.Context) { c.SetTheme(&platform); b.child.view(c) }, 400, 150)
	if actual != (theme.FontRoles{Body: "Georgia", Heading: theme.FontInherit}) {
		t.Fatalf("NativePopover inherited styled parent roles: %+v", actual)
	}
	fonts.Body, fonts.Heading = "Georgia", "monospace"
	parent.Frame()
	child.Frame()
	if actual != (theme.FontRoles{Body: "Georgia", Heading: theme.FontInherit}) {
		t.Fatal("live styled font changes replaced native theme")
	}
}
