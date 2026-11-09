package conversation

import (
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func attachmentStrip(c *ui.Context, state *ui.ScrollState, count int) ui.Element {
	return AttachmentGroupWith(c, AttachmentGroupProps{State: state, Label: "Attachment strip"}, func(c *ui.Context) {
		for i := 0; i < count; i++ {
			AttachmentWith(c.Key(i), AttachmentProps{}, nil).Root.Size(180, 50).Radius(0).Border(0, ui.Transparent).Background(ui.RGB(200, 40, 60))
		}
	}).Size(260, 70)
}

func TestAttachmentGroupFadeUsesNativeOffsetAndLogicalEdges(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(map[ui.Direction]string{ui.LeftToRight: "LTR", ui.RightToLeft: "RTL"}[direction], func(t *testing.T) {
			var state ui.ScrollState
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				ui.Box(c).Fill().Background(ui.RGB(255, 255, 255)).Children(func() { attachmentStrip(c, &state, 4) })
			}, 400, 120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			r.SetDirection(direction)
			bounds, _ := r.Find("Attachment strip")
			left, right := int(bounds.X+1), int(bounds.X+bounds.W-2)
			start, end := left, right
			if direction == ui.RightToLeft {
				start, end = right, left
			}
			full := color.RGBA{200, 40, 60, 255}
			sample := func(x int) color.RGBA { return r.Image().RGBAAt(x, int(bounds.Y+40)) }
			if colorDistance(sample(start), full) > 2 || sample(end).G < 200 {
				t.Fatalf("initial fade did not reveal native logical start: start=%v end=%v max=%v", sample(start), sample(end), state.MaxX)
			}
			state.X = 100
			r.Frame()
			if sample(start).G < 180 || sample(end).G < 180 {
				t.Fatalf("middle scroll should fade both edges: start=%v end=%v", sample(start), sample(end))
			}
			state.X = state.MaxX
			r.Frame()
			if sample(start).G < 200 || colorDistance(sample(end), full) > 2 {
				t.Fatalf("final fade did not reveal native logical end: start=%v end=%v", sample(start), sample(end))
			}
			if r.FramePending() {
				t.Fatal("static scroll mask scheduled an animation loop")
			}
		})
	}
}

func TestAttachmentGroupNativeWheelSnapsAndKeyboardScrolls(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(map[ui.Direction]string{ui.LeftToRight: "LTR", ui.RightToLeft: "RTL"}[direction], func(t *testing.T) {
			var state ui.ScrollState
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				attachmentStrip(c, &state, 4)
			}, 400, 120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			r.SetDirection(direction)
			bounds, _ := r.Find("Attachment strip")
			delta := float32(100)
			if direction == ui.RightToLeft {
				delta = -delta
			}
			r.Scroll(bounds.X+100, bounds.Y+30, delta, 0)
			r.Advance(151 * time.Millisecond)
			if math.Abs(float64(state.X-192)) > .1 {
				t.Fatalf("real native wheel did not snap to next180+12 card: X=%v", state.X)
			}
			state.Handle.Focus()
			r.Frame()
			r.Key(0, ui.KeyEnd)
			if state.X != state.MaxX || state.MaxX <= 0 {
				t.Fatalf("native End did not reach strip end: X=%v max=%v", state.X, state.MaxX)
			}
			r.Key(0, ui.KeyHome)
			if state.X != 0 {
				t.Fatalf("native Home did not return to strip start: X=%v", state.X)
			}
		})
	}
}

func TestAttachmentGroupContainsHorizontalBoundaryAndLeavesVerticalWheel(t *testing.T) {
	var parent, group ui.ScrollState
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.ScrollBoth(c).Size(360, 100).TrackScroll(&parent).Children(func() {
			ui.Column(c).Width(700).Children(func() {
				attachmentStrip(c, &group, 4)
				ui.Box(c).Size(700, 400)
			})
		})
	}, 400, 150, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	group.X = group.MaxX
	r.Frame()
	r.Scroll(100, 30, 100, 0)
	if parent.X != 0 || group.X != group.MaxX {
		t.Fatalf("horizontal boundary chained to parent: parent=%v group=%v/%v", parent.X, group.X, group.MaxX)
	}
	r.Scroll(100, 30, 0, 30)
	if parent.Y != 30 {
		t.Fatalf("vertical wheel was swallowed by horizontal strip: parentY=%v", parent.Y)
	}
}

func TestAttachmentGroupKeysPreserveEditorFocusAndCompositionWhenReordered(t *testing.T) {
	ids := []string{"a", "b", "c"}
	values := map[string]*string{}
	for _, id := range []string{"a", "b", "c", "new"} {
		value := ""
		values[id] = &value
	}
	var state ui.ScrollState
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		AttachmentGroupWith(c, AttachmentGroupProps{State: &state, Label: "Editable attachments"}, func(c *ui.Context) {
			for _, id := range ids {
				AttachmentWith(c, AttachmentProps{ID: id, Name: id}, func(a *AttachmentParts) {
					a.Content(func(c *ui.Context) { ui.TextInput(c, values[id]).Label("Rename " + id).Width(140) })
				}).Root.Width(180)
			}
		}).Size(260, 80)
	}, 400, 130, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Rename a"); err != nil {
		t.Fatal(err)
	}
	r.Type("first")
	r.Compose("文", 1)
	ids = []string{"new", "a", "b", "c"}
	r.Frame()
	if !r.Focused("Rename a") {
		t.Fatal("stable attachment ID lost focused native editor on prepend")
	}
	r.Compose("文档", 2)
	r.Type("文档")
	if *values["a"] != "first文档" || *values["new"] != "" {
		t.Fatalf("IME composition moved to another keyed attachment: a=%q new=%q", *values["a"], *values["new"])
	}
}

func TestAttachmentGroupHomeEndLeavesDescendantEditorCaretKeys(t *testing.T) {
	var state ui.ScrollState
	value := "initial"
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		AttachmentGroupWith(c, AttachmentGroupProps{State: &state}, func(c *ui.Context) {
			AttachmentWith(c, AttachmentProps{ID: "editor"}, func(a *AttachmentParts) {
				a.Content(func(c *ui.Context) { ui.TextInput(c, &value).Label("Editable filename").Width(140) })
			}).Root.Width(180)
			Attachment(c, AttachmentProps{ID: "next", Name: "Another file"}).Width(180)
		}).Size(260, 90)
	}, 400, 140, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Editable filename"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyEnd)
	r.Type(" end")
	r.Key(0, ui.KeyHome)
	r.Type("start ")
	if value != "start initial end" || state.X != 0 {
		t.Fatalf("strip consumed descendant editor caret keys: value=%q X=%v", value, state.X)
	}
}
