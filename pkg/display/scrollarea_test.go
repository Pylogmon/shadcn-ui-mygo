package display

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func scrollAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	e, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q", role, label)
	}
	return e
}

func TestScrollAreaAxesMoveActualNativeContentAndReportAcceptedOffsets(t *testing.T) {
	for _, axis := range []ui.Orientation{ui.Vertical, ui.Horizontal, ui.Horizontal | ui.Vertical} {
		for _, rtl := range []bool{false, true} {
			t.Run(fmt.Sprintf("axis=%d/rtl=%v", axis, rtl), func(t *testing.T) {
				state, changes := ui.ScrollState{}, 0
				var contentBounds ui.Rect
				var events []ScrollAreaPosition
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					root := ui.Column(c).Padding(20)
					if rtl {
						root.Direction(ui.RightToLeft)
					}
					root.Children(func() {
						parts := ScrollAreaWith(c, ScrollAreaProps{Orientation: axis, State: &state, Label: "Canvas", OnScroll: func(p ScrollAreaPosition) { events = append(events, p) }}, func(child *ui.Context) {
							ui.Box(child).Size(600, 500).Shrink(0).Background(ui.RGB(20, 90, 180)).Draw(func(_ *ui.Painter, bounds ui.Rect) { contentBounds = bounds })
						})
						parts.Root.Size(200, 120)
						if parts.Root.Changed() {
							changes++
						}
					})
				}, 700, 400)
				before := contentBounds
				view := scrollAccess(t, r, ui.RoleScroll, "Canvas")
				dx := float32(120)
				if rtl {
					dx = -dx
				}
				r.Scroll(view.Bounds.X+40, view.Bounds.Y+40, dx, 80)
				after := contentBounds
				x, y := float32(0), float32(0)
				if axis&ui.Horizontal != 0 {
					x = 120
				}
				if axis&ui.Vertical != 0 {
					y = 80
				}
				if state.X != x || state.Y != y || before.Y-after.Y != y || !rtl && before.X-after.X != x || rtl && after.X-before.X != x {
					t.Fatalf("state did not move native content: state=%+v before=%+v after=%+v", state, before, after)
				}
				if len(events) != 1 || changes != 1 || events[0].X != x || events[0].Y != y {
					t.Fatalf("accepted callback/Changed=%+v /%d", events, changes)
				}
				state.X, state.Y = math.MaxFloat32, math.MaxFloat32
				r.Frame()
				if state.X != state.MaxX || state.Y != state.MaxY || len(events) != 2 || events[1].X != state.X || events[1].Y != state.Y {
					t.Fatalf("reported request instead of native clamp: %+v events=%+v", state, events)
				}
				r.Advance(time.Second)
				if r.FramePending() || len(events) != 2 || changes != 2 {
					t.Fatal("settled scrolling retained frames or repeated callbacks")
				}
			})
		}
	}
}

func TestScrollAreaPreservesParentTextAndCompleteScopedTokens(t *testing.T) {
	custom := theme.Light()
	custom.Primary, custom.Border = ui.RGB(90, 20, 210), ui.RGB(230, 40, 60)
	var inside theme.Tokens
	var nativeBefore, nativeAfter *ui.Theme
	r := ui.NewTester(func(c *ui.Context) {
		theme.Scope(c, custom).Fill().Children(func() {
			ui.Column(c).Font("monospace").FontSize(24).TextColor(ui.RGB(10, 40, 190)).Padding(20).Children(func() {
				ui.Box(c).Height(48).Children(func() { ui.Text(c, "Inherited style").Label("Reference text").AlignSelf(ui.Start) })
				nativeBefore = c.Theme()
				ScrollAreaWith(c, ScrollAreaProps{Label: "Styled region"}, func(child *ui.Context) {
					inside = *theme.TokensOf(child)
					ui.Text(child, "Inherited style").Label("Inside text").AlignSelf(ui.Start)
				}).Root.Size(300, 100)
				nativeAfter = c.Theme()
			})
		})
	}, 500, 280)
	out, _ := r.Find("Reference text")
	in, _ := r.Find("Inside text")
	if out.W != in.W || out.H != in.H || in.H < 24 || inside.Primary != custom.Primary || inside.Border != custom.Border || nativeBefore != nativeAfter {
		t.Fatalf("scroll wrapper reset inherited text/theme: outside=%+v inside=%+v tokens=%+v", out, in, inside)
	}
	// Both identical glyph runs must paint the parent's blue rather than the
	// default native Theme.Text. Compare the tight text rectangles row by row.
	im := r.Image()
	for y := 0; y < int(out.H); y++ {
		for x := 0; x < int(out.W); x++ {
			if im.RGBAAt(int(out.X)+x, int(out.Y)+y) != im.RGBAAt(int(in.X)+x, int(in.Y)+y) {
				t.Fatal("scroll wrapper changed the inherited text color or glyphs")
			}
		}
	}
}

func TestScrollAreaNativeKeyboardAndBarsDoNotStealEditorFocus(t *testing.T) {
	state, value := ui.ScrollState{}, "draft"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			ScrollAreaWith(c, ScrollAreaProps{State: &state, Label: "Keyboard region"}, func(child *ui.Context) {
				control.Input(child, &value, control.InputProps{Label: "Draft"})
				ui.Box(child).Height(500).Shrink(0)
			}).Root.Size(240, 140)
			ui.Button(c, "After region")
		})
	}, 400, 350)
	r.Key(0, ui.KeyTab)
	if !r.Focused("Keyboard region") {
		t.Fatal("overflowing native viewport was not a keyboard target")
	}
	r.Key(0, ui.KeyPageDown)
	if state.Y <= 0 {
		t.Fatal("native viewport PageDown did not scroll")
	}
	r.Key(0, ui.KeyHome)
	displayClick(t, r, "Draft")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Edited draft")
	view := scrollAccess(t, r, ui.RoleScroll, "Keyboard region").Bounds
	r.Press(view.X+view.W-5, view.Y+8)
	r.Move(view.X+view.W+30, view.Y+40)
	r.CancelPress()
	if !r.Focused("Draft") || value != "Edited draft" {
		t.Fatalf("thumb capture/cancel stole editor state/focus: value=%q offset=%g", value, state.Y)
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("After region") {
		t.Fatal("native scrollbars introduced an unexpected keyboard stop")
	}
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleSlider {
			t.Fatal("scroll wrapper introduced an independent AX adjustment target")
		}
	}
}

func TestScrollAreaDefaultNestedBoundaryChainingUsesNativeOffsets(t *testing.T) {
	inner, outer := ui.ScrollState{}, ui.ScrollState{}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			ScrollAreaWith(c, ScrollAreaProps{State: &outer, Label: "Outer region"}, func(child *ui.Context) {
				ScrollAreaWith(child, ScrollAreaProps{State: &inner, Label: "Inner region"}, func(nested *ui.Context) {
					ui.Box(nested).FillWidth().Height(400).Shrink(0)
				}).Root.Size(200, 100).Shrink(0)
				ui.Box(child).Height(600).Shrink(0)
			}).Root.Size(260, 220)
		})
	}, 400, 350)
	view := scrollAccess(t, r, ui.RoleScroll, "Inner region").Bounds
	r.Scroll(view.X+30, view.Y+30, 0, 80)
	if inner.Y != 80 || outer.Y != 0 {
		t.Fatal("inner native region did not consume a scroll within its range")
	}
	inner.Y = math.MaxFloat32
	r.Frame()
	r.Scroll(view.X+30, view.Y+30, 0, 80)
	if inner.Y != inner.MaxY || outer.Y != 80 {
		t.Fatalf("edge did not chain outward: inner=%+v outer=%+v", inner, outer)
	}
}

func TestScrollAreaEmptyContentAndAxisChangesKeepFollowingEditorIdentity(t *testing.T) {
	show, axis, initialized := false, ui.Vertical, 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			ScrollAreaWith(c, ScrollAreaProps{Orientation: axis}, func(child *ui.Context) {
				if show {
					ui.Box(child).Size(500, 400)
				}
			}).Root.Size(240, 100)
			owner := ui.Column(c)
			value := ui.Local(owner, "draft", func() string { initialized++; return "draft" })
			owner.Children(func() { control.Input(c, value, control.InputProps{Label: "Following editor"}) })
		})
	}, 400, 300)
	displayClick(t, r, "Following editor")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("edited")
	show, axis = true, ui.Horizontal|ui.Vertical
	r.Frame()
	if initialized != 1 || !r.Focused("Following editor") || scrollAccess(t, r, ui.RoleTextField, "Following editor").Value != "edited" {
		t.Fatal("empty callback or scrollbar insertion leaked identity")
	}
}

func scrollPointerTester(state *ui.ScrollState, orientation ui.Orientation, rtl bool, onScroll func(ScrollAreaPosition)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		tokens := theme.Light()
		tokens.Border = ui.RGB(230, 30, 70)
		theme.Scope(c, tokens).Fill().Children(func() {
			root := ui.Column(c).Padding(20)
			if rtl {
				root.Direction(ui.RightToLeft)
			}
			root.Children(func() {
				ScrollAreaWith(c, ScrollAreaProps{Orientation: orientation, State: state, Label: "Pointer region", Scrollbars: ScrollbarsAlways, OnScroll: onScroll}, func(child *ui.Context) {
					ui.Box(child).Size(600, 600).Shrink(0).Background(ui.RGB(20, 80, 180))
				}).Root.Size(200, 120)
			})
		})
	}, 700, 400)
}

func TestScrollAreaCallbackOverrideUsesNextAcceptedNativeOffset(t *testing.T) {
	state := ui.ScrollState{}
	var reports []float32
	r := scrollPointerTester(&state, ui.Vertical, false, func(p ScrollAreaPosition) {
		reports = append(reports, p.Y)
		if len(reports) == 1 {
			state.Y = 0
		}
	})
	view := scrollAccess(t, r, ui.RoleScroll, "Pointer region").Bounds
	r.Scroll(view.X+20, view.Y+20, 0, 100)
	if state.Y != 0 || len(reports) != 2 || reports[0] != 100 || reports[1] != 0 {
		t.Fatalf("reentrant callback swallowed the replacement offset: state=%+v reports=%v", state, reports)
	}
	r.Advance(time.Second)
	if r.FramePending() || len(reports) != 2 {
		t.Fatal("accepted callback replacement created an idle loop")
	}
}

func TestScrollAreaInvalidOrientationMarksRealViewportWithoutDisablingContent(t *testing.T) {
	state, value := ui.ScrollState{}, ""
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			ScrollAreaWith(c, ScrollAreaProps{Orientation: ui.Orientation(255), State: &state, Label: "Invalid region"}, func(child *ui.Context) {
				control.Input(child, &value, control.InputProps{Label: "Editable child"})
				ui.Box(child).Size(500, 500).Shrink(0)
			}).Root.Size(240, 140)
		})
	}, 400, 300)
	view := scrollAccess(t, r, ui.RoleScroll, "Invalid region")
	if !view.Invalid || view.Disabled || state.MaxX != 0 || state.MaxY <= 0 {
		t.Fatalf("invalid axis was not a finite, usable native fallback: %+v state=%+v", view, state)
	}
	displayClick(t, r, "Editable child")
	r.Type("still editable")
	if value != "still editable" {
		t.Fatal("invalid scroll configuration disabled independent content")
	}
}

func TestScrollAreaMatchesMyGoScrollbarPixelsAndInput(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			for _, axis := range []ui.Orientation{ui.Vertical, ui.Horizontal, ui.Horizontal | ui.Vertical} {
				for _, visibility := range []ScrollbarVisibility{ScrollbarsAuto, ScrollbarsAlways, ScrollbarsHidden} {
					t.Run(fmt.Sprintf("dark=%v/rtl=%v/axis=%d/visibility=%d", dark, rtl, axis, visibility), func(t *testing.T) {
						var states [2]ui.ScrollState
						render := func(reference bool, state *ui.ScrollState) *ui.Tester {
							return ui.NewTester(func(c *ui.Context) {
								tokens := theme.Light()
								if dark {
									tokens = theme.Dark()
								}
								native := tokens.Theme()
								native.Scrollbar, native.ScrollbarWidth = ui.RGB(210, 20, 120), 8
								ui.Themed(c, native).Fill().Background(native.Background).Children(func() {
									root := ui.Column(c).Padding(20)
									if rtl {
										root.Direction(ui.RightToLeft)
									}
									root.Children(func() {
										content := func(c *ui.Context) { ui.Box(c).Size(600, 500).Shrink(0).Background(ui.RGB(20, 80, 180)) }
										if !reference {
											ScrollAreaWith(c, ScrollAreaProps{Orientation: axis, State: state, Scrollbars: visibility, Label: "Parity viewport"}, content).Root.Size(200, 120)
											return
										}
										ui.Box(c).Size(200, 120).Clip().Children(func() {
											var viewport ui.Element
											switch axis {
											case ui.Horizontal:
												viewport = ui.ScrollHorizontal(c)
											case ui.Horizontal | ui.Vertical:
												viewport = ui.ScrollBoth(c)
											default:
												viewport = ui.Scroll(c)
											}
											viewport.Fill().MinWidth(0).MinHeight(0).TrackScroll(state).FocusRing(false).Label("Parity viewport").
												Scrollbars(visibility != ScrollbarsHidden).KeepScrollbars(visibility == ScrollbarsAlways).
												Children(func() { content(c) })
										})
									})
								})
							}, 300, 200)
						}
						actual, reference := render(false, &states[0]), render(true, &states[1])
						view := scrollAccess(t, actual, ui.RoleScroll, "Parity viewport").Bounds
						hoverX, hoverY := view.X+view.W/2, view.Y+view.H/2
						check := func(stage string) {
							t.Helper()
							if states[0].X != states[1].X || states[0].Y != states[1].Y || states[0].MaxX != states[1].MaxX || states[0].MaxY != states[1].MaxY || !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
								t.Fatalf("%s differs from ui.Scroll: actual=%+v reference=%+v", stage, states[0], states[1])
							}
						}
						check("initial")
						actual.Move(hoverX, hoverY)
						reference.Move(hoverX, hoverY)
						check("hover")
						dx := float32(70)
						if rtl {
							dx = -dx
						}
						actual.Scroll(hoverX, hoverY, dx, 50)
						reference.Scroll(hoverX, hoverY, dx, 50)
						check("wheel")
						if visibility != ScrollbarsHidden {
							for i := range states {
								states[i].X, states[i].Y = 0, 0
							}
							actual.Frame()
							reference.Frame()
							x, y := view.X+view.W-6, view.Y+8
							if rtl {
								x = view.X + 6
							}
							if axis == ui.Horizontal {
								x, y = view.X+10, view.Y+view.H-6
								if rtl {
									x = view.X + view.W - 10
								}
							}
							actual.Press(x, y)
							reference.Press(x, y)
							check("press")
							actual.Move(x+dx, y+30)
							reference.Move(x+dx, y+30)
							check("drag")
							if states[0].X == 0 && states[0].Y == 0 {
								t.Fatal("native thumb did not move the content")
							}
							actual.Release(x+dx, y+30)
							reference.Release(x+dx, y+30)
							check("release")
						}
					})
				}
			}
		}
	}
}

func TestScrollAreaVisibilityChangesPreserveEditorAndScrollState(t *testing.T) {
	var state ui.ScrollState
	visibility, initialized := ScrollbarsAuto, 0
	value := "draft"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ScrollAreaWith(c, ScrollAreaProps{State: &state, Label: "Stateful region", Scrollbars: visibility}, func(child *ui.Context) {
			owner := ui.Column(child)
			ui.Local(owner, "lifecycle", func() bool { initialized++; return true })
			owner.Children(func() {
				control.Input(child, &value, control.InputProps{Label: "Persistent editor"})
				ui.Box(child).Height(600).Shrink(0)
			})
		}).Root.Size(240, 140)
	}, 400, 300)
	displayClick(t, r, "Persistent editor")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("edited")
	for _, next := range []ScrollbarVisibility{ScrollbarsHidden, ScrollbarsAlways, ScrollbarsAuto} {
		visibility = next
		r.Frame()
		if initialized != 1 || !r.Focused("Persistent editor") || value != "edited" {
			t.Fatal("visibility change replaced the editor or lost focus")
		}
		r.Scroll(30, 60, 0, 40)
		if state.Y <= 0 {
			t.Fatal("visibility policy disabled native scrolling")
		}
		state.Y = 0
		r.Frame()
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("settled native viewport retained animation frames")
	}
}
