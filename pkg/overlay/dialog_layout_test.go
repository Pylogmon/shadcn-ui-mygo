package overlay

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func dialogBounds(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.Rect {
	t.Helper()
	info, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing %v %q", role, name)
	}
	return info.Bounds
}

func dialogNear(t *testing.T, name string, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > 1 {
		t.Fatalf("%s = %g, want %g", name, got, want)
	}
}

func TestDialogNoScrollFillsFixedHeightAndPinsFooter(t *testing.T) {
	for _, tc := range []struct {
		name               string
		hideHeader, footer bool
		height, maxHeight  float32
		padding, gap       float32
		fluent             bool
	}{
		{"header-footer", false, true, 360, 0, 24, 24, false},
		{"max-height", false, true, 480, 320, 24, 24, false},
		{"hidden-header", true, true, 360, 0, 24, 24, false},
		{"without-footer", false, false, 360, 0, 24, 24, false},
		{"body-only", true, false, 360, 0, 12, 8, false},
		{"zero-spacing", true, true, 360, 0, 0, 0, false},
		{"fluent-height", false, true, 360, 0, 24, 24, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			open := true
			r := ui.NewTesterWith(view(func(c *ui.Context) {
				p := DialogProps{Title: "Layout", Width: 400, Height: tc.height, MaxHeight: tc.maxHeight, HideHeader: tc.hideHeader,
					HideClose: true, NoScroll: true, ContentPadding: &tc.padding, ContentGap: &tc.gap}
				if tc.fluent {
					p.Height = 0
				}
				if tc.footer {
					p.Footer = func() { ui.Button(c, "Footer").Height(28) }
				}
				panel := Dialog(c, &open, p, func() {
					ui.Scroll(c).Fill().MinHeight(0).Label("Managed list").Children(func() {
						ui.Column(c).FillWidth().Shrink(0).Children(func() {
							for i := 0; i < 30; i++ {
								ui.Text(c, fmt.Sprintf("Managed row %d", i)).Height(24)
							}
							ui.Button(c, "Last managed action")
						})
					})
				})
				if tc.fluent {
					panel.Height(tc.height)
				}
			}), 800, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			panel := dialogBounds(t, r, ui.RoleDialog, "Layout")
			body := dialogBounds(t, r, ui.RoleScroll, "Managed list")
			wantHeight := tc.height
			if tc.maxHeight > 0 {
				wantHeight = min(wantHeight, tc.maxHeight)
			}
			dialogNear(t, "panel height", panel.H, wantHeight)
			dialogNear(t, "body left", body.X, panel.X+1+tc.padding)
			dialogNear(t, "body width", body.W, panel.W-2-2*tc.padding)
			top, bottom := panel.Y+1+tc.padding, panel.Y+panel.H-1-tc.padding
			if !tc.hideHeader {
				heading := dialogBounds(t, r, ui.RoleHeading, "Layout")
				top = heading.Y + heading.H + tc.gap
			}
			if tc.footer {
				footer := dialogBounds(t, r, ui.RoleButton, "Footer")
				dialogNear(t, "footer bottom", footer.Y+footer.H, bottom)
				bottom = footer.Y - tc.gap
			}
			dialogNear(t, "body top", body.Y, top)
			dialogNear(t, "body bottom", body.Y+body.H, bottom)
			if body.H < 40 {
				t.Fatalf("managed list did not receive the remaining height: %+v", body)
			}
			r.Scroll(body.X+body.W/2, body.Y+body.H/2, 0, 2000)
			r.Advance(300 * time.Millisecond)
			last := dialogBounds(t, r, ui.RoleButton, "Last managed action")
			if last.Y < body.Y-1 || last.Y+last.H > body.Y+body.H+1 {
				t.Fatalf("managed list cannot scroll to its last action: body=%+v last=%+v", body, last)
			}
			if tc.footer {
				footer := dialogBounds(t, r, ui.RoleButton, "Footer")
				dialogNear(t, "pinned footer", footer.Y-tc.gap, bottom)
			}
		})
	}
}

func TestDialogScrollViewportReachesEdgesAndPreservesContentPadding(t *testing.T) {
	for _, header := range []bool{false, true} {
		for _, padding := range []float32{0, 12, 24} {
			t.Run(fmt.Sprintf("header-footer=%v/padding=%g", header, padding), func(t *testing.T) {
				open := true
				r := ui.NewTesterWith(view(func(c *ui.Context) {
					p := DialogProps{Title: "Edge viewport", Width: 400, Height: 320, HideHeader: !header, HideClose: true, ContentPadding: &padding}
					if header {
						p.Footer = func() { ui.Button(c, "Footer").Height(28) }
					}
					Dialog(c, &open, p, func() {
						ui.Box(c).FillWidth().Height(24).Role(ui.RoleRegion).Label("First row")
						for i := 0; i < 30; i++ {
							ui.Text(c, fmt.Sprintf("Row %d", i)).Height(24)
						}
						ui.Button(c, "Last action").Height(28)
					})
				}), 800, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				panel := dialogBounds(t, r, ui.RoleDialog, "Edge viewport")
				viewport := dialogBounds(t, r, ui.RoleScroll, "Dialog content")
				first := dialogBounds(t, r, ui.RoleRegion, "First row")
				dialogNear(t, "viewport left edge", viewport.X, panel.X+1)
				dialogNear(t, "viewport right edge", viewport.X+viewport.W, panel.X+panel.W-1)
				dialogNear(t, "content inset", first.X, panel.X+1+padding)
				dialogNear(t, "content width", first.W, panel.W-2-2*padding)
				if !header {
					dialogNear(t, "viewport top edge", viewport.Y, panel.Y+1)
					dialogNear(t, "viewport bottom edge", viewport.Y+viewport.H, panel.Y+panel.H-1)
					dialogNear(t, "content top padding", first.Y, viewport.Y+padding)
				}
				// The old outer padding kept this edge outside the scroll hit area.
				r.Scroll(panel.X+panel.W-4, viewport.Y+viewport.H/2, 0, 2000)
				r.Advance(300 * time.Millisecond)
				last := dialogBounds(t, r, ui.RoleButton, "Last action")
				if last.Y < viewport.Y-1 || last.Y+last.H > viewport.Y+viewport.H+1 {
					t.Fatalf("wheel at dialog edge missed the viewport: viewport=%+v last=%+v", viewport, last)
				}
				if header {
					footer := dialogBounds(t, r, ui.RoleButton, "Footer")
					dialogNear(t, "footer bottom padding", footer.Y+footer.H, panel.Y+panel.H-1-padding)
				} else {
					dialogNear(t, "content bottom padding", last.Y+last.H, viewport.Y+viewport.H-padding)
				}
			})
		}
	}
}

func TestDialogNaturalHeightPreservesSectionSpacing(t *testing.T) {
	for _, noScroll := range []bool{false, true} {
		for _, sections := range []struct{ header, body, footer bool }{
			{true, true, true}, {false, true, false}, {true, false, false}, {false, false, true}, {false, false, false},
		} {
			t.Run(fmt.Sprintf("noScroll=%v/%v", noScroll, sections), func(t *testing.T) {
				open := true
				r := ui.NewTesterWith(view(func(c *ui.Context) {
					var content func()
					p := DialogProps{Title: "Natural height", HideHeader: !sections.header, HideClose: true, NoScroll: noScroll}
					if sections.footer {
						p.Footer = func() { ui.Button(c, "Footer").Height(28) }
					}
					if sections.body {
						content = func() { ui.Box(c).FillWidth().Height(40).Role(ui.RoleRegion).Label("Body") }
					}
					Dialog(c, &open, p, content)
				}), 800, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				panel := dialogBounds(t, r, ui.RoleDialog, "Natural height")
				want, count := float32(50), 0 // two 24-DIP insets and the border
				for _, part := range []struct {
					shown bool
					role  ui.Role
					name  string
				}{{sections.header, ui.RoleHeading, "Natural height"}, {sections.body, ui.RoleRegion, "Body"}, {sections.footer, ui.RoleButton, "Footer"}} {
					if part.shown {
						want += dialogBounds(t, r, part.role, part.name).H
						count++
					}
				}
				want += float32(max(0, count-1)) * 24
				dialogNear(t, "natural panel height", panel.H, want)
			})
		}
	}
}
