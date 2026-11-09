package desktop

import (
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type tooltipGeometryBackend struct {
	fakeBackend
	anchor, content ui.Rect
	known           bool
}
type tooltipGeometrySurface struct {
	Surface
	backend *tooltipGeometryBackend
}

func (b *tooltipGeometryBackend) Open(spec Spec, view func(*ui.Context), closed func()) (Surface, error) {
	s, err := b.fakeBackend.Open(spec, view, closed)
	return &tooltipGeometrySurface{s, b}, err
}
func (s *tooltipGeometrySurface) Geometry() (ui.Rect, ui.Rect, bool) {
	return s.backend.anchor, s.backend.content, s.backend.known
}

func TestTooltipRichChildDescriptionRevisionAndRemoval(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	status, revision, mounted := "Active", uint64(0), true
	var popupErr error
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		if !mounted {
			return
		}
		anchor := control.Button(c, "Save", control.ButtonProps{})
		popupErr = TooltipWith(c, h, "rich", anchor, TooltipProps{Description: "Save changes with S", Revision: revision}, func(child *ui.Context) {
			ui.Text(child, status).FontWeight(600)
			display.KbdWith(child, "S", display.KbdProps{InTooltip: true})
		})
	}, 400, 220)
	if popupErr != nil || len(b.surfaces) != 0 {
		t.Fatalf("closed rich tooltip err=%v surfaces=%d", popupErr, len(b.surfaces))
	}
	parent.Key(0, ui.KeyTab)
	trigger, ok := parent.AccessByRole(ui.RoleButton, "Save")
	if !ok || trigger.Description != "Save changes with S" || len(b.surfaces) != 1 {
		t.Fatal("focus tooltip lost semantic description or independent window")
	}
	s := b.surfaces[0]
	if s.spec.Focusable || !s.spec.IgnoreMouseEvents || !s.spec.ShowInactive {
		t.Fatalf("rich tooltip can intercept parent input: %+v", s.spec)
	}
	child := ui.NewTesterWith(s.view, 220, 80, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !child.HasText("Active") || parent.HasText("Active") {
		t.Fatal("rich tooltip content was not built independently")
	}
	if _, ok := child.AccessByRole(ui.RoleTooltip, "Save changes with S"); !ok {
		t.Fatal("rich tooltip lost native tooltip semantics")
	}
	if key, ok := child.Find("S"); !ok || key.H != 20 {
		t.Fatalf("rich tooltip omitted native 20-DIP Kbd content: %+v", key)
	}
	status, revision = "Updated", revision+1
	parent.Frame()
	child.Frame()
	if !child.HasText("Updated") || len(b.surfaces) != 1 {
		t.Fatal("rich content revision remounted or retained stale content")
	}
	parent.Key(0, ui.KeyEscape)
	if h.IsOpen("rich") || !parent.Focused("Save") {
		t.Fatal("tooltip Escape stole parent focus or failed to dismiss")
	}
	parent.Frame()
	if h.IsOpen("rich") {
		t.Fatal("tooltip immediately reopened under the retained focus")
	}
	mounted = false
	parent.Frame()
	if len(h.entries) != 0 {
		t.Fatal("removed rich trigger retained a child callback")
	}
}

func TestTooltipRichNativeThemeAndLongContent(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "styled", true: "platform"}[native], func(t *testing.T) {
			b := &fakeBackend{}
			h := NewHostWithBackend(b)
			tokens := *theme.Preset("rose", false)
			tokens.Foreground = ui.RGB(8, 77, 111)
			var color ui.Color
			parent := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(tokens.Theme())
				h.BeginFrame()
				defer h.EndFrame()
				anchor := ui.Button(c, "Status")
				build := TooltipWith
				if native {
					build = NativeTooltipWith
				}
				err := build(c, h, "status", anchor, TooltipProps{Description: "Active. Last updated two hours ago.", MaxWidth: 160}, func(child *ui.Context) {
					color = child.Theme().Text
					ui.Column(child).MaxWidth(120).Gap(4).Children(func() {
						ui.Text(child, "Active").FontWeight(600)
						ui.Text(child, "Last updated two hours ago. More information wraps onto several lines.").Opacity(.8)
					})
				})
				if err != nil {
					t.Fatal(err)
				}
			}, 300, 150)
			parent.Key(0, ui.KeyTab)
			child := ui.NewTesterWith(b.surfaces[0].view, 174, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if native == (color == tokens.Foreground) {
				t.Fatal("native/styled content theme routes crossed")
			}
			first, ok := child.Find("Active")
			if !ok {
				t.Fatal("formatted heading missing")
			}
			second, ok := child.Find("Last updated two hours ago. More information wraps onto several lines.")
			if !ok || second.Y <= first.Y || second.H <= first.H || second.W > 120.1 {
				t.Fatalf("formatted tooltip did not wrap bounded native text: first%+v second%+v", first, second)
			}
		})
	}
}

func TestTooltipLogicalSideAndRealNativeArrow(t *testing.T) {
	for _, tc := range []struct {
		side      TooltipSide
		direction ui.Direction
		want      mygo.Placement
	}{{TooltipTop, ui.LeftToRight, mygo.PlacementTop}, {TooltipBottom, ui.LeftToRight, mygo.PlacementBottom}, {TooltipLeft, ui.LeftToRight, mygo.PlacementLeft}, {TooltipRight, ui.LeftToRight, mygo.PlacementRight}, {TooltipInlineStart, ui.LeftToRight, mygo.PlacementLeft}, {TooltipInlineEnd, ui.LeftToRight, mygo.PlacementRight}, {TooltipInlineStart, ui.RightToLeft, mygo.PlacementRight}, {TooltipInlineEnd, ui.RightToLeft, mygo.PlacementLeft}} {
		placement, err := tooltipPlacement(TooltipProps{Side: tc.side}, tc.direction)
		if err != nil || placement != tc.want {
			t.Fatalf("side%s direction%v resolved%s/%v", tc.side, tc.direction, placement, err)
		}
	}
	b := &tooltipGeometryBackend{anchor: ui.Rect{X: 60, Y: 200, W: 80, H: 30}, content: ui.Rect{X: 0, Y: 250, W: 220, H: 50}, known: true}
	h := NewHostWithBackend(b)
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := ui.Button(c, "Arrow")
		if err := Tooltip(c, h, "arrow", anchor, "Hint", TooltipProps{Placement: mygo.PlacementTop}); err != nil {
			t.Fatal(err)
		}
	}, 400, 250)
	parent.Key(0, ui.KeyTab)
	child := ui.NewTesterWith(b.surfaces[0].view, 220, 60, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	// Requested top, but the actual native window is below. The arrow must
	// appear on its top facing the original anchor, never the requested side.
	img := child.Image()
	if img.RGBAAt(100, 3).A == 0 || img.RGBAAt(20, 3).A != 0 {
		t.Fatalf("actual flipped arrow/corner pixels = %v / %v", img.RGBAAt(100, 3), img.RGBAAt(20, 3))
	}
	b.known = false
	child.Frame()
	if child.Image().RGBAAt(100, 3).A != 0 {
		t.Fatal("unknown native geometry invented an arrow direction")
	}
}

func TestTooltipRichRequiresDescription(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		anchor := ui.Button(c, "Unlabelled")
		if err := TooltipWith(c, nil, "missing", anchor, TooltipProps{}, func(child *ui.Context) { ui.Text(child, "Visual only") }); err == nil {
			t.Fatal("rich tooltip accepted missing semantic description")
		}
		if _, err := tooltipPlacement(TooltipProps{Side: "invalid"}, ui.LeftToRight); err == nil {
			t.Fatal("invalid tooltip side accepted")
		}
	}, 200, 100)
}
