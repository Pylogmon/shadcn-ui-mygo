package display

import (
	"fmt"
	"image/color"
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

func TestScrollAreaVegaThumbPixelsHaveSourceInsetsAndSemanticColor(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			t.Run(fmt.Sprintf("dark=%v/rtl=%v", dark, rtl), func(t *testing.T) {
				tokens := theme.Light()
				if dark {
					tokens = theme.Dark()
				}
				tokens.Border = ui.RGB(220, 40, 60)
				state := ui.ScrollState{}
				r := ui.NewTester(func(c *ui.Context) {
					theme.Scope(c, tokens).Fill().Children(func() {
						root := ui.Column(c).Padding(20)
						if rtl {
							root.Direction(ui.RightToLeft)
						}
						root.Children(func() {
							ScrollAreaWith(c, ScrollAreaProps{State: &state, Label: "Articles"}, func(child *ui.Context) {
								ui.Box(child).FillWidth().Height(600).Shrink(0).Background(ui.RGB(20, 80, 180))
							}).Root.Size(200, 120)
						})
					})
				}, 300, 200)
				view := scrollAccess(t, r, ui.RoleScroll, "Articles").Bounds
				x := view.X + view.W - 10
				if rtl {
					x = view.X
				}
				want := color.RGBA{R: 220, G: 40, B: 60, A: 255}
				im := r.Image()
				for offset := 2; offset < 9; offset++ {
					if got := im.RGBAAt(int(x)+offset, int(view.Y)+8); got != want {
						t.Fatalf("7-DIP Border thumb not painted at inset2: offset=%d pixel=%v", offset, got)
					}
				}
				for _, offset := range []int{0, 1, 9} {
					if got := im.RGBAAt(int(x)+offset, int(view.Y)+8); got == want {
						t.Fatalf("track padding/border painted as thumb at%d", offset)
					}
				}
			})
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
	if !r.Focused("Draft") || value != "Edited draft" || state.Y != 0 {
		t.Fatalf("thumb capture/cancel stole editor state/focus: value=%q offset=%g", value, state.Y)
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("After region") {
		t.Fatal("source bars introduced an unexpected keyboard stop")
	}
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleSlider {
			t.Fatal("default source scrollbar invented an AX adjustment target")
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

func scrollThumbCenter(t *testing.T, r *ui.Tester, bar ui.Rect, horizontal bool) (float32, float32) {
	t.Helper()
	im := r.Image()
	first, last := -1, -1
	length := int(bar.H)
	if horizontal {
		length = int(bar.W)
	}
	for i := 0; i < length; i++ {
		x, y := int(bar.X+5), int(bar.Y)+i
		if horizontal {
			x, y = int(bar.X)+i, int(bar.Y+5)
		}
		pixel := im.RGBAAt(x, y)
		if pixel.R > 200 && pixel.G < 80 && pixel.B < 120 {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		t.Fatal("no visible source thumb in the actual track pixels")
	}
	if horizontal {
		return bar.X + float32(first+last)/2, bar.Y + 5
	}
	return bar.X + 5, bar.Y + float32(first+last)/2
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
				ScrollAreaWith(c, ScrollAreaProps{Orientation: orientation, State: state, Label: "Pointer region", FocusableScrollbars: true, OnScroll: onScroll}, func(child *ui.Context) {
					ui.Box(child).Size(600, 600).Shrink(0).Background(ui.RGB(20, 80, 180))
				}).Root.Size(200, 120)
			})
		})
	}, 700, 400)
}

func TestScrollAreaSourceTrackAndCapturedThumbManipulateNativeViewport(t *testing.T) {
	for _, horizontal := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			t.Run(fmt.Sprintf("horizontal=%v/rtl=%v", horizontal, rtl), func(t *testing.T) {
				state := ui.ScrollState{}
				axis, name := ui.Vertical, "Vertical scroll bar"
				if horizontal {
					axis, name = ui.Horizontal, "Horizontal scroll bar"
				}
				r := scrollPointerTester(&state, axis, rtl, nil)
				bar := scrollAccess(t, r, ui.RoleSlider, name)
				view := scrollAccess(t, r, ui.RoleScroll, "Pointer region")
				if len(bar.Controls) != 1 || bar.Controls[0] != view.Name || bar.Min != 0 || bar.Max <= 0 || bar.RoleDescription != "scroll bar" || bar.Horizontal != horizontal {
					t.Fatalf("opt-in supported native adjustment semantics=%+v", bar)
				}
				x, y := scrollThumbCenter(t, r, bar.Bounds, horizontal)
				r.Press(x, y)
				if horizontal {
					dx := float32(1000)
					if rtl {
						dx = -dx
					}
					r.Move(x+dx, y+40)
					if state.X != state.MaxX {
						t.Fatalf("captured horizontal thumb did not reach native end: %+v", state)
					}
				} else {
					r.Move(x+200, y+1000)
					if state.Y != state.MaxY {
						t.Fatalf("captured vertical thumb did not reach native end: %+v", state)
					}
				}
				r.CancelPress()
				if state.X != 0 || state.Y != 0 {
					t.Fatalf("native capture cancellation did not restore start: %+v", state)
				}
				bar = scrollAccess(t, r, ui.RoleSlider, name)
				x, y = bar.Bounds.X+bar.Bounds.W/2, bar.Bounds.Y+bar.Bounds.H*.75
				if horizontal {
					x, y = bar.Bounds.X+bar.Bounds.W*.75, bar.Bounds.Y+5
					if rtl {
						x = bar.Bounds.X + bar.Bounds.W*.25
					}
				}
				r.Press(x, y)
				r.Release(x, y)
				value, maximum := state.Y, state.MaxY
				if horizontal {
					value, maximum = state.X, state.MaxX
				}
				if value <= maximum/2 || value >= maximum {
					t.Fatalf("track press did not center the real thumb around the clicked point: value=%g max=%g", value, maximum)
				}
				// The range now reports the exact offset, and native scrollbar keys
				// can go home/end while retaining their actual adjustment focus.
				r.Key(0, ui.KeyHome)
				if state.X != 0 || state.Y != 0 {
					t.Fatal("source bar Home did not return native content to start")
				}
				r.Key(0, ui.KeyEnd)
				bar = scrollAccess(t, r, ui.RoleSlider, name)
				if bar.Now != bar.Max || !r.Focused(name) {
					t.Fatal("source bar End lost native range/focus")
				}
			})
		}
	}
}

func TestScrollAreaThumbPressStopsNativeWheelAndProgrammaticMotion(t *testing.T) {
	for _, programmatic := range []bool{false, true} {
		t.Run(fmt.Sprint(programmatic), func(t *testing.T) {
			state := ui.ScrollState{}
			r := scrollPointerTester(&state, ui.Vertical, false, nil)
			view := scrollAccess(t, r, ui.RoleScroll, "Pointer region")
			if programmatic {
				state.ScrollTo(0, 400, ui.Animation{Duration: time.Second, Ease: ui.Linear})
				r.Frame()
			} else {
				r.Wheel(view.Bounds.X+20, view.Bounds.Y+20, 0, 240)
			}
			r.Advance(30 * time.Millisecond)
			before := state.Y
			if before <= 0 || !r.FramePending() {
				t.Fatalf("native motion did not begin: offset=%g", before)
			}
			bar := scrollAccess(t, r, ui.RoleSlider, "Vertical scroll bar")
			x, y := scrollThumbCenter(t, r, bar.Bounds, false)
			r.Press(x, y)
			r.Advance(time.Second)
			if state.Y != before || state.Scrolling() || r.FramePending() {
				t.Fatalf("holding thumb retained old native motion: before=%g after=%g pending=%v", before, state.Y, r.FramePending())
			}
			r.Move(x, y+30)
			if state.Y <= before {
				t.Fatal("canceling native motion also canceled the real capture")
			}
			r.CancelPress()
			if state.Y != before {
				t.Fatal("canceled hold did not restore its accepted native offset")
			}
		})
	}
}

func TestScrollAreaBothAxesLeaveSourceCornerAndPreserveNativeLogicalX(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprint(rtl), func(t *testing.T) {
			state := ui.ScrollState{}
			r := scrollPointerTester(&state, ui.Horizontal|ui.Vertical, rtl, nil)
			view := scrollAccess(t, r, ui.RoleScroll, "Pointer region").Bounds
			vertical := scrollAccess(t, r, ui.RoleSlider, "Vertical scroll bar").Bounds
			horizontal := scrollAccess(t, r, ui.RoleSlider, "Horizontal scroll bar").Bounds
			if vertical.W != 10 || vertical.H != view.H-10 || horizontal.H != 10 || horizontal.W != view.W-10 {
				t.Fatalf("source corner dimensions: view=%+v v=%+v h=%+v", view, vertical, horizontal)
			}
			if !rtl && (vertical.X != view.X+view.W-10 || horizontal.X != view.X) || rtl && (vertical.X != view.X || horizontal.X != view.X+10) {
				t.Fatal("RTL source bars/corner were not mirrored")
			}
			state.X, state.Y = math.MaxFloat32, math.MaxFloat32
			r.Frame()
			if state.X != 400 || state.Y != 480 {
				t.Fatalf("corner changed the actual native viewport range: %+v", state)
			}
			x, _ := scrollThumbCenter(t, r, horizontal, true)
			if !rtl && x < horizontal.X+horizontal.W/2 || rtl && x > horizontal.X+horizontal.W/2 {
				t.Fatal("logical end did not put the actual thumb at the correct side")
			}
			r.Scroll(horizontal.X+horizontal.W/2, horizontal.Y+5, 50, 0)
			want := float32(400)
			if rtl {
				want = 350
			}
			if state.X != want || state.Y != 480 {
				t.Fatalf("wheel over horizontal source track used the wrong native axis/direction: %+v", state)
			}
		})
	}
}

func TestScrollAreaStyleVisibilityChangesPreserveEditorAndRealScrollState(t *testing.T) {
	state := ui.ScrollState{}
	native, visibility, initialized := false, ScrollbarsAlways, 0
	value := "draft"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			ScrollAreaWith(c, ScrollAreaProps{State: &state, Label: "Stateful region", NativeScrollbars: native, Scrollbars: visibility}, func(child *ui.Context) {
				owner := ui.Column(child)
				ui.Local(owner, "lifecycle", func() bool { initialized++; return true })
				owner.Children(func() {
					control.Input(child, &value, control.InputProps{Label: "Persistent editor"})
					ui.Box(child).Height(600).Shrink(0)
				})
			}).Root.Size(240, 140)
		})
	}, 400, 300)
	displayClick(t, r, "Persistent editor")
	r.Type("!")
	for _, nextNative := range []bool{true, false} {
		for _, nextVisibility := range []ScrollbarVisibility{ScrollbarsAuto, ScrollbarsHidden, ScrollbarsAlways} {
			native, visibility = nextNative, nextVisibility
			r.Frame()
			if initialized != 1 || !r.Focused("Persistent editor") || value != "!draft" && value != "draft!" {
				t.Fatal("presentation change replaced the native editor or focus")
			}
			view := scrollAccess(t, r, ui.RoleScroll, "Stateful region").Bounds
			r.Scroll(view.X+30, view.Y+60, 0, 40)
			if state.Y <= 0 {
				t.Fatal("hidden/native indicator mode disabled actual scrolling")
			}
			state.Y = 0
			r.Frame()
		}
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("indicator switches retained idle animation frames")
	}
}

func TestScrollAreaAutoBarsKeepCaptureThenHideWithoutNewKeyboardStops(t *testing.T) {
	state := ui.ScrollState{}
	r := ui.NewTester(func(c *ui.Context) {
		tokens := theme.Light()
		tokens.Border = ui.RGB(230, 30, 70)
		theme.Scope(c, tokens).Fill().Children(func() {
			ui.Column(c).Padding(20).Children(func() {
				ScrollAreaWith(c, ScrollAreaProps{State: &state, Label: "Auto region", Scrollbars: ScrollbarsAuto}, func(child *ui.Context) {
					ui.Box(child).FillWidth().Height(600).Shrink(0).Background(ui.RGB(20, 80, 180))
				}).Root.Size(200, 120)
			})
		})
	}, 500, 400)
	view := scrollAccess(t, r, ui.RoleScroll, "Auto region").Bounds
	if got := r.Image().RGBAAt(int(view.X+195), int(view.Y+8)); got.R > 200 && got.G < 80 {
		t.Fatal("auto thumb showed before hover")
	}
	r.Move(view.X+100, view.Y+60)
	x, y := scrollThumbCenter(t, r, ui.Rect{X: view.X + view.W - 10, Y: view.Y, W: 10, H: view.H}, false)
	r.Press(x, y)
	r.Move(x+150, y+30)
	if state.Y <= 0 {
		t.Fatal("auto hiding interrupted captured pointer movement")
	}
	r.Release(x+150, y+30)
	if got := r.Image().RGBAAt(int(view.X+195), int(view.Y+8)); got.R > 200 && got.G < 80 {
		t.Fatal("auto bar retained its hover after capture outside the region")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("auto bars scheduled idle frames")
	}
}

func TestScrollAreaQueuedScrollbarKeysAccumulateBeforeNextNativeLayout(t *testing.T) {
	state := ui.ScrollState{}
	var handler *scrollBarRuntime
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			parts := ScrollAreaWith(c, ScrollAreaProps{State: &state, FocusableScrollbars: true}, func(child *ui.Context) {
				ui.Box(child).FillWidth().Height(600).Shrink(0)
			})
			parts.Root.Size(200, 120)
			handler = ui.Local(parts.VerticalBar, scrollBarRuntimeKey{}, func() scrollBarRuntime { return scrollBarRuntime{} })
		})
	}, 400, 300)
	for _, key := range []ui.Key{ui.KeyDown, ui.KeyDown, ui.KeyUp} {
		if !handler.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: key}) {
			t.Fatal("queued native adjustment key was not handled")
		}
	}
	if state.Y != 40 {
		t.Fatalf("queued keys lost cumulative native request: %g", state.Y)
	}
	r.Frame()
	if state.Y != 40 || state.Scrolling() || scrollAccess(t, r, ui.RoleSlider, "Vertical scroll bar").Now != 40 {
		t.Fatal("a native motion cancel overwrote the final coalesced keyboard target")
	}
}

func TestScrollAreaInvalidBarParametersAndTinyTracksRemainFinite(t *testing.T) {
	for _, thickness := range []float32{float32(math.NaN()), float32(math.Inf(1)), -1, 3, 0} {
		state := ui.ScrollState{}
		value := ""
		r := ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			ui.Column(c).Padding(20).Children(func() {
				ScrollAreaWith(c, ScrollAreaProps{State: &state, Thickness: thickness, FocusableScrollbars: true}, func(child *ui.Context) {
					control.Input(child, &value, control.InputProps{Label: "Usable native editor"})
					ui.Box(child).Height(500).Shrink(0)
				}).Root.Size(200, 60)
			})
		}, 400, 250)
		bar := scrollAccess(t, r, ui.RoleSlider, "Vertical scroll bar")
		if bar.Bounds.W != 10 || math.IsNaN(bar.Max) || math.IsInf(bar.Max, 0) || thickness != 0 && (!bar.Disabled || !bar.Invalid) {
			t.Fatalf("invalid parameter leaked nonfinite geometry/range: thickness=%g bar=%+v", thickness, bar)
		}
		displayClick(t, r, "Usable native editor")
		r.Type("Native")
		if value != "Native" {
			t.Fatal("invalid scrollbar parameters disabled the actual editor")
		}
	}
	state := ui.ScrollState{}
	r := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Padding(20).Children(func() {
			ScrollAreaWith(c, ScrollAreaProps{State: &state, FocusableScrollbars: true}, func(child *ui.Context) {
				ui.Box(child).Height(600).Shrink(0)
			}).Root.Size(80, 8)
		})
	}, 200, 100)
	bar := scrollAccess(t, r, ui.RoleSlider, "Vertical scroll bar")
	r.Press(bar.Bounds.X+5, bar.Bounds.Y+4)
	r.Move(bar.Bounds.X+5, bar.Bounds.Y+100)
	r.Release(bar.Bounds.X+5, bar.Bounds.Y+100)
	if state.Y != 0 || math.IsNaN(float64(state.Y)) || math.IsInf(float64(state.Y), 0) {
		t.Fatal("a track shorter than the source thumb floor divided by zero")
	}
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

func TestScrollAreaReconfigurationNeverReplaysOldCapturedOffset(t *testing.T) {
	for _, change := range []string{"native", "hidden", "axis", "state", "direction"} {
		t.Run(change, func(t *testing.T) {
			first, replacement := ui.ScrollState{Y: 100}, ui.ScrollState{Y: 60}
			state := &first
			native, rtl, visibility, axis := false, false, ScrollbarsAlways, ui.Vertical
			r := ui.NewTester(func(c *ui.Context) {
				tokens := theme.Light()
				tokens.Border = ui.RGB(230, 30, 70)
				theme.Scope(c, tokens).Fill().Children(func() {
					root := ui.Column(c).Padding(20)
					if rtl {
						root.Direction(ui.RightToLeft)
					}
					root.Children(func() {
						ScrollAreaWith(c, ScrollAreaProps{State: state, Label: "Reconfigured region", Orientation: axis, NativeScrollbars: native, Scrollbars: visibility, FocusableScrollbars: true}, func(child *ui.Context) {
							ui.Box(child).Size(600, 600).Shrink(0).Background(ui.RGB(20, 80, 180))
						}).Root.Size(200, 120)
					})
				})
			}, 700, 400)
			bar := scrollAccess(t, r, ui.RoleSlider, "Vertical scroll bar")
			x, y := scrollThumbCenter(t, r, bar.Bounds, false)
			r.Press(x, y)
			r.Move(x, y+20)
			accepted := state.Y
			switch change {
			case "native":
				native = true
			case "hidden":
				visibility = ScrollbarsHidden
			case "axis":
				axis, accepted = ui.Horizontal, 0
			case "state":
				state, accepted = &replacement, 60
			case "direction":
				rtl = true
			}
			r.Frame()
			r.Move(x+20, y+80)
			r.Release(x+20, y+80)
			if state.Y != accepted {
				t.Fatalf("%s replayed old capture after reconfiguration: offset=%g accepted=%g", change, state.Y, accepted)
			}
		})
	}
}

func TestScrollAreaCornerThumbUsesUsableTrackForNativeDragRatio(t *testing.T) {
	for _, horizontal := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			t.Run(fmt.Sprintf("horizontal=%v/rtl=%v", horizontal, rtl), func(t *testing.T) {
				state := ui.ScrollState{}
				r := scrollPointerTester(&state, ui.Horizontal|ui.Vertical, rtl, nil)
				name, viewportSize, maximum := "Vertical scroll bar", float32(120), state.MaxY
				if horizontal {
					name, viewportSize, maximum = "Horizontal scroll bar", 200, state.MaxX
				}
				bar := scrollAccess(t, r, ui.RoleSlider, name).Bounds
				track := bar.H
				if horizontal {
					track = bar.W
				}
				// The native physical travel excludes both one-DIP track paddings.
				// Moving half that travel must show half of the scrollable content.
				usable := track - 2
				thumb := usable * viewportSize / (viewportSize + maximum)
				distance := (usable - thumb) / 2
				x, y := scrollThumbCenter(t, r, bar, horizontal)
				r.Press(x, y)
				if horizontal {
					if rtl {
						distance = -distance
					}
					r.Move(x+distance, y)
				} else {
					r.Move(x, y+distance)
				}
				value := state.Y
				if horizontal {
					value = state.X
				}
				if math.Abs(float64(value-maximum/2)) > .001 {
					t.Fatalf("corner track mapped physical half-travel to %g instead of %g", value, maximum/2)
				}
				r.CancelPress()
			})
		}
	}
}

func TestScrollAreaAutoBarRemainsVisibleForKeyboardFocusInsideRegion(t *testing.T) {
	state := ui.ScrollState{}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			ScrollAreaWith(c, ScrollAreaProps{State: &state, Label: "Auto keyboard region", Scrollbars: ScrollbarsAuto, FocusableScrollbars: true}, func(child *ui.Context) {
				ui.Box(child).FillWidth().Height(600).Shrink(0)
			}).Root.Size(200, 120)
		})
	}, 500, 400)
	r.Key(0, ui.KeyTab)
	if !r.Focused("Auto keyboard region") {
		t.Fatal("source viewport was not keyboard reachable")
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("Vertical scroll bar") {
		t.Fatal("opt-in scrollbar lost focus when the pointer was outside")
	}
	if _, ok := r.AccessByRole(ui.RoleSlider, "Vertical scroll bar"); !ok {
		t.Fatal("auto visibility removed the focused native adjustment target")
	}
	r.Key(0, ui.KeyDown)
	if state.Y != 40 {
		t.Fatal("focused auto scrollbar stopped accepting native adjustment keys")
	}
}

func TestScrollAreaSwitchFromFocusedCustomBarRestoresNativeViewport(t *testing.T) {
	state := ui.ScrollState{}
	native := false
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			ScrollAreaWith(c, ScrollAreaProps{State: &state, Label: "Restored region", NativeScrollbars: native, FocusableScrollbars: true}, func(child *ui.Context) {
				ui.Box(child).Height(600).Shrink(0)
			}).Root.Size(200, 120)
		})
	}, 400, 300)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyTab)
	if !r.Focused("Vertical scroll bar") {
		t.Fatal("custom native adjustment target was not focused")
	}
	native = true
	r.Frame()
	if !r.Focused("Restored region") {
		t.Fatal("presentation switch discarded keyboard navigation focus")
	}
	r.Key(0, ui.KeyDown)
	if state.Y != 40 {
		t.Fatal("restored native viewport did not receive the next navigation key")
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
