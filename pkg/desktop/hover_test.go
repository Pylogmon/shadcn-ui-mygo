package desktop

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type hoverBackend struct {
	fakeBackend
	focuses             int
	applicationInactive bool
}
type hoverSurface struct {
	*fakeSurface
	backend *hoverBackend
}

func (b *hoverBackend) Open(s Spec, view func(*ui.Context), closed func()) (Surface, error) {
	surface, err := b.fakeBackend.Open(s, view, closed)
	if err != nil {
		return nil, err
	}
	return &hoverSurface{fakeSurface: surface.(*fakeSurface), backend: b}, nil
}
func (s *hoverSurface) Focus()                  { s.backend.focuses++ }
func (b *hoverBackend) ApplicationActive() bool { return !b.applicationInactive }

type focusedHoverBackend struct {
	hoverBackend
	childFocused bool
}
type focusedHoverSurface struct {
	Surface
	backend *focusedHoverBackend
}

func (b *focusedHoverBackend) Open(spec Spec, view func(*ui.Context), closed func()) (Surface, error) {
	surface, err := b.hoverBackend.Open(spec, view, closed)
	if err != nil {
		return nil, err
	}
	return &focusedHoverSurface{Surface: surface, backend: b}, nil
}
func (s *focusedHoverSurface) IsFocused() bool { return s.backend.childFocused }
func (s *focusedHoverSurface) Focus() {
	if focus, ok := s.Surface.(interface{ Focus() }); ok {
		focus.Focus()
	}
}

func moveOver(t *testing.T, view *ui.Tester, label string) {
	t.Helper()
	r, ok := view.Find(label)
	if !ok {
		t.Fatalf("%q missing in %v", label, view.Texts())
	}
	view.Move(r.X+r.W/2, r.Y+r.H/2)
}

func TestTooltipHoverDelayCancellationReentryAndContentRefresh(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	text := "Helpful tip"
	disabled := false
	var popupErr error
	var events []bool
	p := TooltipProps{OpenDelay: 200 * time.Millisecond, CloseDelay: 100 * time.Millisecond, OnOpenChange: func(open bool) { events = append(events, open) }}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(20).Children(func() {
			anchor := control.Button(c, "Trigger", control.ButtonProps{})
			p.Disabled = disabled
			popupErr = Tooltip(c, h, "tip", anchor, text, p)
		})
	}, 420, 220)
	moveOver(t, parent, "Trigger")
	parent.Advance(199 * time.Millisecond)
	if len(b.surfaces) != 0 {
		t.Fatal("tooltip opened before delay")
	}
	parent.Move(400, 200)
	parent.Advance(time.Second)
	if len(b.surfaces) != 0 {
		t.Fatal("cancelled enter timer opened tooltip")
	}
	moveOver(t, parent, "Trigger")
	parent.Advance(200 * time.Millisecond)
	if popupErr != nil || len(b.surfaces) != 1 {
		t.Fatalf("surfaces=%d err=%v", len(b.surfaces), popupErr)
	}
	surface := b.surfaces[0]
	if surface.spec.Focusable || !surface.spec.ShowInactive || !surface.spec.IgnoreMouseEvents || surface.spec.SystemPopover {
		t.Fatalf("tooltip input spec %+v", surface.spec)
	}
	if parent.HasText(text) {
		t.Fatal("tooltip content painted into parent")
	}
	child := ui.NewTester(surface.view, 220, 50)
	if !child.HasText(text) {
		t.Fatal("child omitted tooltip content")
	}
	if _, ok := child.AccessByRole(ui.RoleTooltip, text); !ok {
		t.Fatal("tooltip native role absent")
	}
	parent.Frame()
	parent.Frame()
	surface.invalidates = 0
	text = "Updated tip"
	parent.Frame()
	if surface.invalidates != 1 {
		t.Fatalf("text change redraws=%d", surface.invalidates)
	}
	child.Frame()
	if !child.HasText(text) {
		t.Fatal("open tooltip retained old text")
	}
	parent.Frame()
	if surface.invalidates != 1 {
		t.Fatal("stable tooltip invalidated child again")
	}
	parent.Move(400, 200)
	parent.Advance(60 * time.Millisecond)
	moveOver(t, parent, "Trigger")
	parent.Advance(200 * time.Millisecond)
	if !h.IsOpen("tip") || len(b.surfaces) != 1 {
		t.Fatal("re-entry did not cancel close")
	}
	disabled = true
	parent.Frame()
	if h.IsOpen("tip") || surface.closes != 1 {
		t.Fatal("disabled tooltip was retained")
	}
	if !slices.Equal(events, []bool{true, false}) {
		t.Fatalf("open-change events %v", events)
	}
	parent.Advance(5 * time.Second)
	if parent.FramePending() {
		t.Fatal("closed tooltip scheduled idle frames")
	}
}

func TestTooltipKeyboardEscapeSuppressionAndInactiveWindow(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			anchor := control.Button(c, "Trigger", control.ButtonProps{})
			if err := Tooltip(c, h, "tip", anchor, "Keyboard tip", TooltipProps{OpenDelay: time.Second}); err != nil {
				t.Fatal(err)
			}
			control.Button(c, "Next control", control.ButtonProps{})
		})
	}, 420, 220)
	parent.Key(0, ui.KeyTab)
	if !h.IsOpen("tip") || !parent.Focused("Trigger") {
		t.Fatal("focus did not open immediately without stealing focus")
	}
	parent.Key(0, ui.KeyEscape)
	parent.Frame()
	if h.IsOpen("tip") {
		t.Fatal("Escape did not suppress focused tooltip")
	}
	parent.Advance(2 * time.Second)
	if h.IsOpen("tip") {
		t.Fatal("suppressed tooltip reopened on unchanged focus")
	}
	parent.Key(0, ui.KeyTab)
	parent.Key(ui.Shift, ui.KeyTab)
	if !h.IsOpen("tip") {
		t.Fatal("leaving then re-entering focus did not rearm")
	}
	parent.SetWindowActive(false)
	if h.IsOpen("tip") {
		t.Fatal("inactive parent retained tooltip")
	}
	parent.SetWindowActive(true)
	parent.Frame()
	if h.IsOpen("tip") {
		t.Fatal("activation reopened tooltip without re-entry")
	}
	parent.Key(0, ui.KeyTab)
	parent.Key(ui.Shift, ui.KeyTab)
	if !h.IsOpen("tip") {
		t.Fatal("focus did not rearm after activation")
	}
	parent.SetWindowHidden(true)
	if h.IsOpen("tip") {
		t.Fatal("hidden parent retained tooltip")
	}
}

func TestTooltipStylePairUsesIndependentPlatformTheme(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "shadcn", true: "platform"}[native], func(t *testing.T) {
			b := &fakeBackend{}
			h := NewHostWithBackend(b)
			tokens := theme.Preset("violet", true)
			tokens.Foreground = ui.RGB(12, 34, 56)
			parent := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(tokens.Theme())
				h.BeginFrame()
				defer h.EndFrame()
				ui.Column(c).Padding(20).Children(func() {
					anchor := control.Button(c, "Trigger", control.ButtonProps{})
					var err error
					if native {
						err = NativeTooltip(c, h, "tip", anchor, "Description", TooltipProps{})
					} else {
						err = Tooltip(c, h, "tip", anchor, "Description", TooltipProps{})
					}
					if err != nil {
						t.Fatal(err)
					}
				})
			}, 420, 220)
			parent.Key(0, ui.KeyTab)
			if len(b.surfaces) != 1 {
				t.Fatal("focus did not open tooltip")
			}
			var childTheme ui.Theme
			child := ui.NewTester(func(c *ui.Context) { b.surfaces[0].view(c); childTheme = *c.Theme() }, 240, 50)
			if native && childTheme.Text == tokens.Foreground {
				t.Fatal("native tooltip inherited shadcn theme")
			}
			if !native && childTheme.Text != tokens.Foreground {
				t.Fatal("shadcn tooltip lost parent tokens")
			}
			child.Advance(400 * time.Millisecond)
			if child.FramePending() {
				t.Fatal("tooltip animation did not settle")
			}
		})
	}
}

func TestNativeTooltipPaintsPlatformInverseFallbackAndExplicitColors(t *testing.T) {
	for _, tc := range []struct {
		name       string
		native     *ui.Theme
		fill, text ui.Color
	}{
		{name: "light defaults", native: ui.LightTheme(), fill: ui.LightTheme().Text, text: ui.LightTheme().Background},
		{name: "dark defaults", native: ui.DarkTheme(), fill: ui.DarkTheme().Text, text: ui.DarkTheme().Background},
		{name: "explicit inverse", native: func() *ui.Theme {
			n := ui.LightTheme()
			n.Inverse = ui.RGB(21, 57, 93)
			n.InverseText = ui.RGB(249, 225, 173)
			return n
		}(), fill: ui.RGB(21, 57, 93), text: ui.RGB(249, 225, 173)},
		{name: "explicit fill only", native: func() *ui.Theme { n := ui.LightTheme(); n.Inverse = ui.RGB(21, 57, 93); return n }(), fill: ui.RGB(21, 57, 93), text: ui.LightTheme().Background},
		{name: "explicit text only", native: func() *ui.Theme { n := ui.LightTheme(); n.InverseText = ui.RGB(249, 225, 173); return n }(), fill: ui.LightTheme().Text, text: ui.RGB(249, 225, 173)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := &fakeBackend{}
			h := NewHostWithBackend(b)
			parent := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				h.BeginFrame()
				defer h.EndFrame()
				ui.Column(c).Padding(20).Children(func() {
					anchor := control.Button(c, "Trigger", control.ButtonProps{})
					if err := NativeTooltip(c, h, "tip", anchor, "MMMM WWWW", TooltipProps{}); err != nil {
						t.Fatal(err)
					}
				})
			}, 420, 220)
			parent.Key(0, ui.KeyTab)
			if len(b.surfaces) != 1 {
				t.Fatal("native tooltip was not opened")
			}
			child := ui.NewTester(func(c *ui.Context) { c.SetTheme(tc.native); b.surfaces[0].view(c) }, 240, 60)
			child.Advance(300 * time.Millisecond)
			label, ok := child.Find("MMMM WWWW")
			if !ok {
				t.Fatal("tooltip text missing")
			}
			image := child.Image()
			fill := image.RGBAAt(int(label.X+12), int(label.Y+3))
			if fill.R != tc.fill.R || fill.G != tc.fill.G || fill.B != tc.fill.B || fill.A != 255 {
				t.Fatalf("native tooltip fill=%v want opaque %+v; inverse fields=%+v/%+v", fill, tc.fill, tc.native.Inverse, tc.native.InverseText)
			}
			foregroundPainted := false
			for y := int(label.Y + 6); y < int(label.Y+label.H-6); y++ {
				for x := int(label.X + 12); x < int(label.X+label.W-12); x++ {
					pixel := image.RGBAAt(x, y)
					if pixel.A == 255 && pixel.R == tc.text.R && pixel.G == tc.text.G && pixel.B == tc.text.B {
						foregroundPainted = true
					}
				}
			}
			if !foregroundPainted {
				t.Fatalf("native tooltip painted no opaque foreground pixels of %+v", tc.text)
			}
		})
	}
}

func TestHoverCardPointerGapAndChildKeepAliveWithoutFocusSteal(t *testing.T) {
	b := &hoverBackend{}
	h := NewHostWithBackend(b)
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(20).Children(func() {
			anchor := control.Button(c, "Profile", control.ButtonProps{})
			if err := HoverCard(c, h, "profile", anchor, HoverCardProps{Label: "Profile preview", OpenDelay: 100 * time.Millisecond, CloseDelay: 150 * time.Millisecond}, func(child *ui.Context) {
				ui.Text(child, "Details")
				control.Button(child, "Visit profile", control.ButtonProps{})
			}); err != nil {
				t.Fatal(err)
			}
		})
	}, 420, 220)
	moveOver(t, parent, "Profile")
	parent.Advance(99 * time.Millisecond)
	if len(b.surfaces) != 0 {
		t.Fatal("card opened before hover delay")
	}
	// Native timers coalesce to at least a millisecond after the frame that
	// arms them; a simultaneous hover color transition may arm at 99.999 ms.
	parent.Advance(2 * time.Millisecond)
	if len(b.surfaces) != 1 || b.focuses != 0 {
		t.Fatalf("surfaces=%d focus requests=%d", len(b.surfaces), b.focuses)
	}
	surface := b.surfaces[0]
	if !surface.spec.Focusable || !surface.spec.ShowInactive || surface.spec.IgnoreMouseEvents {
		t.Fatalf("interactive hover spec %+v", surface.spec)
	}
	if surface.spec.MinWidth != 320 || surface.spec.MaxWidth != 320 {
		t.Fatal("FitContent lost requested HoverCard width")
	}
	child := ui.NewTester(surface.view, 320, 180)
	child.SetWindowActive(false)
	parent.Move(400, 200)
	parent.Advance(80 * time.Millisecond)
	if !h.IsOpen("profile") {
		t.Fatal("gap grace closed too soon")
	}
	moveOver(t, child, "Visit profile")
	parent.Frame()
	parent.Advance(300 * time.Millisecond)
	if !h.IsOpen("profile") {
		t.Fatal("cross-window pointer entry did not cancel close")
	}
	if b.focuses != 0 {
		t.Fatal("pointer hover took native keyboard")
	}
	child.Move(600, 400)
	parent.Frame()
	parent.Advance(149 * time.Millisecond)
	if !h.IsOpen("profile") {
		t.Fatal("child leave closed before delay")
	}
	parent.Advance(time.Millisecond)
	if h.IsOpen("profile") || surface.closes != 1 {
		t.Fatal("card did not close after leaving both windows")
	}
	parent.Advance(3 * time.Second)
	if parent.FramePending() {
		t.Fatal("closed card left an idle timer loop")
	}
}

func TestHoverCardPointerLeaveClosesWhileChildRemainsDrawnActive(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			b := &hoverBackend{}
			h := NewHostWithBackend(b)
			parent := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				h.BeginFrame()
				defer h.EndFrame()
				anchor := control.Button(c, "Profile", control.ButtonProps{})
				card := HoverCard
				if native {
					card = NativeHoverCard
				}
				if err := card(c, h, "profile", anchor, HoverCardProps{OpenDelay: -1, CloseDelay: 50 * time.Millisecond}, func(child *ui.Context) {
					control.Button(child, "Preview action", control.ButtonProps{})
				}); err != nil {
					t.Fatal(err)
				}
			}, 420, 220)
			moveOver(t, parent, "Profile")
			if len(b.surfaces) != 1 || b.focuses != 0 {
				t.Fatal("pointer opening did not create an inactive child")
			}
			drawnActive := false
			child := ui.NewTester(func(c *ui.Context) {
				drawnActive = c.WindowActive()
				b.surfaces[0].view(c)
			}, 320, 180)
			if !drawnActive {
				t.Fatal("regression requires the drawn-active state inherited from an active parent")
			}
			moveOver(t, child, "Preview action")
			parent.Move(400, 200)
			parent.Frame()
			parent.Advance(100 * time.Millisecond)
			if !h.IsOpen("profile") {
				t.Fatal("pointer inside the child failed to keep the card open")
			}
			child.Move(600, 400)
			parent.Frame()
			parent.Advance(51 * time.Millisecond)
			if h.IsOpen("profile") || b.surfaces[0].closes != 1 || b.focuses != 0 {
				t.Fatal("drawn-active child kept a pointer-only card after both pointers left")
			}
			parent.Advance(time.Second)
			if parent.FramePending() {
				t.Fatal("pointer leave left an idle timer loop")
			}
		})
	}
}

func TestHoverCardUsesRealChildFocusThenClosesAfterFocusAndPointerLeave(t *testing.T) {
	b := &focusedHoverBackend{}
	h := NewHostWithBackend(b)
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := control.Button(c, "Profile", control.ButtonProps{})
		if err := HoverCard(c, h, "profile", anchor, HoverCardProps{OpenDelay: -1, CloseDelay: 50 * time.Millisecond}, func(child *ui.Context) {
			ui.Text(child, "Preview")
		}); err != nil {
			t.Fatal(err)
		}
	}, 420, 220)
	moveOver(t, parent, "Profile")
	child := ui.NewTester(b.surfaces[0].view, 320, 180)
	b.childFocused = true
	child.Frame()
	parent.Move(400, 200)
	child.Move(600, 400)
	parent.Frame()
	parent.Advance(100 * time.Millisecond)
	if !h.IsOpen("profile") {
		t.Fatal("real focused child was dismissed on pointer leave")
	}
	b.childFocused = false
	child.Frame()
	parent.Frame()
	parent.Advance(51 * time.Millisecond)
	if h.IsOpen("profile") || b.surfaces[0].closes != 1 {
		t.Fatal("child retained the card after true window focus and both pointers left")
	}
}

func TestHoverCardCompletedKeyboardHandoffDoesNotKeepCardAfterChildBlur(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			b := &focusedHoverBackend{}
			h := NewHostWithBackend(b)
			value := ""
			parent := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				h.BeginFrame()
				defer h.EndFrame()
				ui.Column(c).Padding(16).Gap(12).Children(func() {
					anchor := control.Button(c, "Profile", control.ButtonProps{})
					card := HoverCard
					if native {
						card = NativeHoverCard
					}
					if err := card(c, h, "profile", anchor, HoverCardProps{Label: "User profile", CloseDelay: 50 * time.Millisecond}, func(child *ui.Context) {
						control.Button(child, "Visit profile", control.ButtonProps{})
					}); err != nil {
						t.Fatal(err)
					}
					control.Input(c, &value, control.InputProps{Label: "Other input"})
				})
			}, 420, 220)
			parent.Key(0, ui.KeyTab)
			parent.Key(0, ui.KeyDown)
			if b.focuses != 1 {
				t.Fatal("Down did not request the child keyboard")
			}
			b.childFocused = true
			child := ui.NewTester(b.surfaces[0].view, 320, 180)
			child.Advance(250 * time.Millisecond)
			parent.Frame()
			if !h.IsOpen("profile") || !child.Focused("User profile") {
				t.Fatal("keyboard handoff did not establish real child focus")
			}
			b.childFocused = false
			child.Frame()
			child.Move(600, 400)
			if err := parent.Click("Other input"); err != nil {
				t.Fatal(err)
			}
			parent.Frame()
			parent.Advance(51 * time.Millisecond)
			if h.IsOpen("profile") || b.surfaces[0].closes != 1 {
				t.Fatal("application activity kept a completed keyboard handoff alive after child blur")
			}
			if !parent.Focused("Other input") {
				t.Fatal("closing an unfocused child stole the parent's new input focus")
			}
			parent.Advance(time.Second)
			if parent.FramePending() {
				t.Fatal("completed handoff left idle timer frames")
			}
		})
	}
}

func TestHoverCardNativeDismissalBeforeChildBlurFramePreservesNewParentFocus(t *testing.T) {
	b := &focusedHoverBackend{}
	h := NewHostWithBackend(b)
	value := ""
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(16).Gap(12).Children(func() {
			anchor := control.Button(c, "Profile", control.ButtonProps{})
			if err := NativeHoverCard(c, h, "profile", anchor, HoverCardProps{Label: "User profile"}, func(child *ui.Context) {
				ui.Text(child, "Preview")
			}); err != nil {
				t.Fatal(err)
			}
			control.Input(c, &value, control.InputProps{Label: "Other input"})
		})
	}, 420, 220)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyDown)
	b.childFocused = true
	child := ui.NewTester(b.surfaces[0].view, 320, 180)
	child.Frame()
	b.childFocused = false
	if err := parent.Click("Other input"); err != nil {
		t.Fatal(err)
	}
	// Native destruction can deliver OnClosed before a child blur frame had
	// time to update its last-known focus. That stale bit must not steal focus.
	b.surfaces[0].dismiss()
	parent.Frame()
	if h.IsOpen("profile") || !parent.Focused("Other input") {
		t.Fatal("late native focus report stole the parent's newly focused editor")
	}
}

func TestHoverCardKeyboardEntryEscapeAndNativeDismissal(t *testing.T) {
	b := &hoverBackend{}
	h := NewHostWithBackend(b)
	visits := 0
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			anchor := control.Button(c, "Profile", control.ButtonProps{})
			if err := HoverCard(c, h, "profile", anchor, HoverCardProps{Label: "User profile"}, func(child *ui.Context) {
				if control.Button(child, "Visit profile", control.ButtonProps{}).Clicked() {
					visits++
				}
			}); err != nil {
				t.Fatal(err)
			}
			control.Button(c, "Next", control.ButtonProps{})
		})
	}, 420, 220)
	parent.Key(0, ui.KeyTab)
	if !h.IsOpen("profile") || !parent.Focused("Profile") {
		t.Fatal("anchor focus did not open card immediately")
	}
	parent.Key(0, ui.KeyDown)
	if b.focuses != 1 {
		t.Fatalf("keyboard entry native focus requests=%d", b.focuses)
	}
	surface := b.surfaces[0]
	child := ui.NewTester(surface.view, 320, 180)
	if info, ok := child.AccessByRole(ui.RoleGroup, "User profile"); !ok || !info.Focusable {
		t.Fatalf("keyboard group %+v", info)
	}
	child.Key(0, ui.KeyTab)
	child.Key(0, ui.KeyEnter)
	if visits != 1 {
		t.Fatal("child native action could not be reached by keyboard")
	}
	parent.SetWindowActive(false)
	parent.Frame()
	if !h.IsOpen("profile") {
		t.Fatal("focused child was closed with inactive parent")
	}
	child.Key(0, ui.KeyEscape)
	parent.Frame()
	if h.IsOpen("profile") {
		t.Fatal("child Escape did not close card")
	}
	parent.SetWindowActive(true)
	if !parent.Focused("Profile") {
		t.Fatal("close did not restore trigger focus")
	}
	parent.Frame()
	if h.IsOpen("profile") {
		t.Fatal("focused trigger reopened dismissed card")
	}
	parent.Key(0, ui.KeyTab)
	parent.Key(ui.Shift, ui.KeyTab)
	if !h.IsOpen("profile") {
		t.Fatal("focus leave/re-entry did not rearm card")
	}
	latest := b.surfaces[len(b.surfaces)-1]
	latest.dismiss()
	parent.Frame()
	if h.IsOpen("profile") {
		t.Fatal("native dismissal reopened from restored focus")
	}
}

func TestHoverCardKeyboardTransferSurvivesInactiveChildFrameButNotAppDeactivation(t *testing.T) {
	for _, native := range []bool{false, true} {
		for _, completed := range []bool{false, true} {
			t.Run(fmt.Sprintf("native=%v/completed=%v", native, completed), func(t *testing.T) {
				b := &hoverBackend{}
				h := NewHostWithBackend(b)
				parent := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					h.BeginFrame()
					defer h.EndFrame()
					anchor := control.Button(c, "Profile", control.ButtonProps{})
					card := HoverCard
					if native {
						card = NativeHoverCard
					}
					if err := card(c, h, "profile", anchor, HoverCardProps{Label: "User profile"}, func(child *ui.Context) { control.Button(child, "Visit profile", control.ButtonProps{}) }); err != nil {
						t.Fatal(err)
					}
				}, 420, 220)
				parent.Key(0, ui.KeyTab)
				parent.Key(0, ui.KeyDown)
				if b.focuses != 1 {
					t.Fatalf("native focus requests=%d", b.focuses)
				}
				parent.SetWindowActive(false)
				if !h.IsOpen("profile") {
					t.Fatal("parent blur cancelled explicit child keyboard entry")
				}
				renderChild := false
				child := ui.NewTester(func(c *ui.Context) {
					if renderChild {
						b.surfaces[0].view(c)
					}
				}, 320, 180)
				child.SetWindowActive(false)
				renderChild = true
				child.Frame()
				parent.Frame()
				if !h.IsOpen("profile") {
					t.Fatal("inactive first child frame cancelled keyboard transfer")
				}
				child.Advance(300 * time.Millisecond)
				parent.Advance(300 * time.Millisecond)
				if child.FramePending() || parent.FramePending() {
					t.Fatal("pending keyboard transfer scheduled idle frames")
				}
				if completed {
					child.SetWindowActive(true)
					parent.Frame()
					if !h.IsOpen("profile") {
						t.Fatal("completed keyboard transfer lost its child")
					}
				}
				b.applicationInactive = true
				child.SetWindowActive(false)
				parent.Frame()
				if h.IsOpen("profile") || b.surfaces[0].closes != 1 {
					t.Fatal("application deactivation retained the keyboard child")
				}
				parent.Advance(time.Second)
				if parent.FramePending() {
					t.Fatal("dismissed keyboard transfer left idle frames")
				}
			})
		}
	}
}

func TestHoverCardKeyboardTransferDoesNotAssumeUnknownApplicationIsActive(t *testing.T) {
	b := &hoverBackend{}
	// Expose only the required Backend methods, not ApplicationActive.
	h := NewHostWithBackend(struct{ Backend }{b})
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := control.Button(c, "Profile", control.ButtonProps{})
		if err := NativeHoverCard(c, h, "profile", anchor, HoverCardProps{}, func(child *ui.Context) { ui.Text(child, "Details") }); err != nil {
			t.Fatal(err)
		}
	}, 420, 220)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyDown)
	if b.focuses != 1 {
		t.Fatal("keyboard did not request native child focus")
	}
	parent.SetWindowActive(false)
	if h.IsOpen("profile") || b.surfaces[0].closes != 1 {
		t.Fatal("unknown application activity kept the inactive child alive")
	}
}

func TestHoverCardAppDeactivationAndNativeStyle(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	tokens := theme.Preset("violet", true)
	tokens.Foreground = ui.RGB(12, 34, 56)
	parent := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(tokens.Theme())
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(20).Children(func() {
			anchor := control.Button(c, "Profile", control.ButtonProps{})
			if err := NativeHoverCard(c, h, "card", anchor, HoverCardProps{Label: "Platform preview", OpenDelay: -1}, func(child *ui.Context) { ui.Text(child, "Platform content") }); err != nil {
				t.Fatal(err)
			}
		})
	}, 420, 220)
	moveOver(t, parent, "Profile")
	if len(b.surfaces) != 1 {
		t.Fatal("native card did not open")
	}
	var childTheme ui.Theme
	child := ui.NewTester(func(c *ui.Context) { b.surfaces[0].view(c); childTheme = *c.Theme() }, 320, 180)
	child.SetWindowActive(false)
	if childTheme.Text == tokens.Foreground {
		t.Fatal("native card inherited shadcn theme")
	}
	if b.surfaces[0].spec.SystemPopover {
		t.Fatal("hover preview used a focus-stealing NSPopover")
	}
	moveOver(t, child, "Platform content")
	parent.Frame()
	parent.SetWindowActive(false)
	if h.IsOpen("card") {
		t.Fatal("stale pointer hover kept card after app deactivation")
	}
}

func TestHoverCardTabEntryAndBackwardReturn(t *testing.T) {
	b := &hoverBackend{}
	h := NewHostWithBackend(b)
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			anchor := control.Button(c, "Trigger", control.ButtonProps{})
			if err := HoverCard(c, h, "card", anchor, HoverCardProps{Label: "Details"}, func(child *ui.Context) { control.Button(child, "Child action", control.ButtonProps{}) }); err != nil {
				t.Fatal(err)
			}
			control.Button(c, "After trigger", control.ButtonProps{})
		})
	}, 420, 220)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyTab)
	if b.focuses != 1 || parent.Focused("After trigger") {
		t.Fatal("Tab bypassed opened hover content")
	}
	child := ui.NewTester(b.surfaces[0].view, 320, 150)
	child.Key(ui.Shift, ui.KeyTab)
	parent.Frame()
	if h.IsOpen("card") || !parent.Focused("Trigger") {
		t.Fatal("Shift+Tab from entry group did not return to trigger")
	}
}

func TestHoverMissingHostErrorsOnlyWhenOpening(t *testing.T) {
	var seen error
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			anchor := control.Button(c, "Trigger", control.ButtonProps{})
			if err := Tooltip(c, nil, "tip", anchor, "Hint", TooltipProps{}); err != nil {
				seen = err
			}
		})
	}, 420, 220)
	if seen != nil {
		t.Fatal("closed tooltip required provider")
	}
	moveOver(t, parent, "Trigger")
	if seen == nil {
		t.Fatal("opening did not report missing Host")
	}
	if parent.HasText("Hint") {
		t.Fatal("missing provider silently drew parent-window fallback")
	}
}

func TestNonInteractiveHoverCardCannotReceiveChildActions(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	actions := 0
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(20).Children(func() {
			anchor := control.Button(c, "Trigger", control.ButtonProps{})
			if err := HoverCard(c, h, "card", anchor, HoverCardProps{NonInteractive: true, OpenDelay: -1, CloseDelay: -1}, func(child *ui.Context) {
				if control.Button(child, "Preview action", control.ButtonProps{}).Clicked() {
					actions++
				}
			}); err != nil {
				t.Fatal(err)
			}
		})
	}, 420, 220)
	moveOver(t, parent, "Trigger")
	if len(b.surfaces) != 1 || b.surfaces[0].spec.Focusable || !b.surfaces[0].spec.IgnoreMouseEvents {
		t.Fatalf("noninteractive card spec %+v", b.surfaces)
	}
	child := ui.NewTester(b.surfaces[0].view, 320, 150)
	button, ok := child.AccessByRole(ui.RoleButton, "Preview action")
	if !ok || !button.Disabled {
		t.Fatal("noninteractive preview exposed enabled action")
	}
	_ = child.Click("Preview action")
	if actions != 0 {
		t.Fatal("noninteractive preview action ran")
	}
	parent.Move(400, 200)
	if h.IsOpen("card") {
		t.Fatal("noninteractive card remained open after parent leave")
	}
}
