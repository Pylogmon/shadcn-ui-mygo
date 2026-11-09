package desktop

import (
	"errors"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type fakeBackend struct {
	surfaces           []*fakeSurface
	invalidates        int
	onParentInvalidate func()
	err                error
	closeDuringOpen    bool
}
type fakeSurface struct {
	spec                       Spec
	view                       func(*ui.Context)
	dismiss                    func()
	closes, invalidates, moves int
}

func (b *fakeBackend) Open(s Spec, view func(*ui.Context), closed func()) (Surface, error) {
	if b.err != nil {
		return nil, b.err
	}
	f := &fakeSurface{spec: s, view: view, dismiss: closed}
	b.surfaces = append(b.surfaces, f)
	if b.closeDuringOpen {
		closed()
	}
	return f, nil
}
func (b *fakeBackend) InvalidateParent() {
	b.invalidates++
	if b.onParentInvalidate != nil {
		b.onParentInvalidate()
	}
}
func (f *fakeSurface) SetAnchor(r ui.Rect) { f.moves++; f.spec.Anchor = r }
func (f *fakeSurface) Invalidate()         { f.invalidates++ }
func (f *fakeSurface) Close()              { f.closes++ }

func TestPopoverIndependentContextThemeAndAnchorRefresh(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	open := true
	left := float32(20)
	shift := float32(0)
	tokens := *theme.Preset("violet", true)
	tokens.Chart4 = ui.RGB(12, 34, 56)
	tokens.Popover = ui.RGB(31, 28, 40)
	var parentContext *ui.Context
	var popupErr error
	var childTokens theme.Tokens
	var childReduced bool
	parent := ui.NewTester(func(c *ui.Context) {
		parentContext = c
		c.SetTheme(tokens.Theme())
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(left).Children(func() {
			anchor := control.Button(c, "Open settings", control.ButtonProps{})
			anchor.Translate(shift, 0)
			popupErr = h.Popover(c, "settings", anchor, &open, PopoverProps{Label: "Settings", Plain: true, Placement: mygo.PlacementRightStart, Width: 320}, func(child *ui.Context) {
				if child == parentContext {
					t.Fatal("parent context was reused in child")
				}
				childTokens = *theme.TokensOf(child)
				childReduced = motion.Reduced(child)
				control.Input(child, new(string), control.InputProps{Label: "Setting"})
			})
		})
	}, 500, 300)
	if popupErr != nil {
		t.Fatal(popupErr)
	}
	if len(b.surfaces) != 1 || !h.IsOpen("settings") {
		t.Fatalf("surfaces=%d open=%v", len(b.surfaces), h.IsOpen("settings"))
	}
	surface := b.surfaces[0]
	if surface.spec.SystemPopover || !surface.spec.Focusable || surface.spec.Anchor.W <= 0 {
		t.Fatalf("spec=%+v", surface.spec)
	}
	child := ui.NewTester(surface.view, 320, 200)
	if childTokens != tokens {
		t.Fatalf("semantic tokens were lost: got %+v want %+v", childTokens, tokens)
	}
	if parentContext == nil {
		t.Fatal("absent parent")
	}
	if _, ok := child.AccessByRole(ui.RoleDialog, "Settings"); !ok {
		t.Fatal("native content lacks dialog metadata")
	}
	// Settle first-frame instant preferences, then prove stable parent builds
	// do not request more child frames or native moves.
	parent.Frame()
	parent.Frame()
	surface.invalidates, surface.moves = 0, 0
	parent.Frame()
	if surface.invalidates != 0 || surface.moves != 0 {
		t.Fatalf("idle refresh: %d redraws, %d moves", surface.invalidates, surface.moves)
	}
	left = 40
	parent.Frame()
	if surface.moves != 1 {
		t.Fatalf("anchor moves=%d", surface.moves)
	}
	shift = 30
	parent.Frame()
	if surface.moves != 2 {
		t.Fatalf("transformed anchor moves=%d", surface.moves)
	}
	tokens.Popover = ui.RGB(6, 9, 12)
	parent.Frame()
	if surface.invalidates != 1 || len(b.surfaces) != 1 {
		t.Fatal("theme update recreated or failed to invalidate native content")
	}
	child.Frame()
	if childTokens.Popover != tokens.Popover {
		t.Fatal("open child retained stale theme")
	}
	parent.SetPreferences(ui.Preferences{ReduceMotion: true})
	child.Frame()
	if !childReduced {
		t.Fatal("child did not inherit reduced motion")
	}
	if parent.FramePending() {
		t.Fatal("steady host scheduled an idle render loop")
	}
}

func TestPopoverNativeDismissalAndDisappearingAnchor(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	open, visible := true, true
	dismissed := 0
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		if visible {
			anchor := control.Button(c, "Open popup", control.ButtonProps{})
			if err := h.Popover(c, "popup", anchor, &open, PopoverProps{OnDismiss: func() { dismissed++ }}, func(c *ui.Context) { ui.Text(c, "Content") }); err != nil {
				t.Fatal(err)
			}
		}
	}, 400, 200)
	first := b.surfaces[0]
	first.dismiss()
	first.dismiss()
	if open || h.IsOpen("popup") || dismissed != 1 {
		t.Fatalf("native dismissal open=%v count=%d", open, dismissed)
	}
	parent.Frame()
	if !parent.Focused("Open popup") {
		t.Fatal("native dismissal did not restore anchor focus")
	}
	open = true
	parent.Frame()
	if len(b.surfaces) != 2 {
		t.Fatal("popup did not reopen")
	}
	second := b.surfaces[1]
	visible = false
	parent.Frame()
	if open || second.closes != 1 || h.IsOpen("popup") {
		t.Fatalf("disappearing anchor open=%v closes=%d", open, second.closes)
	}
	second.dismiss()
	if dismissed != 1 {
		t.Fatal("programmatic cleanup called native OnDismiss")
	}
	h.EndFrame()
	h.CloseAll()
	if second.closes != 1 {
		t.Fatal("surface closed more than once")
	}
}

func TestPopoverReplacedSurfaceRejectsLateNativeCloseAndTransfersOpenState(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	firstOpen, secondOpen := true, true
	open := &firstOpen
	width := 280
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := control.Button(c, "Open", control.ButtonProps{})
		if err := h.Popover(c, "same", anchor, open, PopoverProps{Width: width}, func(c *ui.Context) { ui.Text(c, "Popup") }); err != nil {
			t.Fatal(err)
		}
	}, 400, 200)
	first := b.surfaces[0]
	width = 360
	parent.Frame()
	if len(b.surfaces) != 2 || first.closes != 1 {
		t.Fatal("window configuration was not replaced")
	}
	first.dismiss()
	if !firstOpen || !h.IsOpen("same") {
		t.Fatal("stale native close dismissed replacement")
	}
	open = &secondOpen
	parent.Frame()
	if firstOpen || !secondOpen || len(b.surfaces) != 2 {
		t.Fatal("controlled open pointer transfer orphaned state or reopened surface")
	}
	h.Close("same")
	if secondOpen || b.surfaces[1].closes != 1 {
		t.Fatal("explicit close did not close latest model")
	}
}

func TestPopoverChildMutationInvalidatesParentAfterBuildingContent(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	open := true
	mutated := false
	notifiedAfterMutation := false
	b.onParentInvalidate = func() {
		if mutated && !open {
			notifiedAfterMutation = true
		}
	}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := control.Button(c, "Open editor", control.ButtonProps{})
		if err := h.Popover(c, "edit", anchor, &open, PopoverProps{}, func(child *ui.Context) {
			if control.Button(child, "Done", control.ButtonProps{}).Clicked() {
				mutated = true
				open = false
			}
		}); err != nil {
			t.Fatal(err)
		}
	}, 400, 200)
	surface := b.surfaces[0]
	child := ui.NewTester(surface.view, 288, 200)
	if err := child.Click("Done"); err != nil {
		t.Fatal(err)
	}
	if !mutated || open || !notifiedAfterMutation {
		t.Fatalf("mutation/close ordering mutated=%v open=%v notified=%v", mutated, open, notifiedAfterMutation)
	}
	if surface.closes != 0 {
		t.Fatal("surface destroyed during its own view build")
	}
	parent.Frame()
	if surface.closes != 1 || h.IsOpen("edit") {
		t.Fatal("parent did not reconcile child-requested close")
	}
}

func TestPopoverImmediateDismissDuringOpenAndBackendErrors(t *testing.T) {
	for _, immediate := range []bool{false, true} {
		t.Run(map[bool]string{false: "backend-error", true: "immediate-dismiss"}[immediate], func(t *testing.T) {
			b := &fakeBackend{closeDuringOpen: immediate}
			if !immediate {
				b.err = errors.New("test backend unavailable")
			}
			h := NewHostWithBackend(b)
			open := true
			var got error
			dismissed := 0
			ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				anchor := control.Button(c, "Open", control.ButtonProps{})
				if err := h.Popover(c, "error", anchor, &open, PopoverProps{OnDismiss: func() { dismissed++ }}, nil); err != nil {
					got = err
				}
			}, 400, 200)
			if open || h.IsOpen("error") {
				t.Fatal("failed/immediately dismissed open retained live model")
			}
			if immediate {
				if len(b.surfaces) != 1 || b.surfaces[0].closes != 1 || dismissed != 1 {
					t.Fatal("immediate native dismissal leaked a surface")
				}
			} else if got == nil {
				t.Fatal("backend error was hidden")
			}
		})
	}
}

func TestPopoverValidationAndNativeRectRounding(t *testing.T) {
	if _, err := normalize(PopoverProps{Placement: "invalid"}, ui.Rect{W: 10, H: 10}, 10); err == nil {
		t.Fatal("invalid native placement accepted")
	}
	if _, err := normalize(PopoverProps{MinWidth: 400, MaxWidth: 100}, ui.Rect{W: 10, H: 10}, 10); err == nil {
		t.Fatal("invalid native bounds accepted")
	}
	if err := (*Host)(nil).Popover(nil, "", ui.Element{}, nil, PopoverProps{}, nil); err == nil {
		t.Fatal("nil host accepted")
	}
	r := nativeRect(ui.Rect{X: 10.5, Y: 20.25, W: 40.25, H: 10.5})
	if r != (mygo.Rectangle{X: 10, Y: 20, Width: 41, Height: 11}) {
		t.Fatalf("outward anchor rounding=%+v", r)
	}
}

func TestPairedPopoverStylesAndIndependentSystemAppearance(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	themedOpen, nativeOpen := true, true
	parentTokens := theme.Preset("violet", true)
	var themedTokens, nativeTokens theme.Tokens
	var themedMenus, nativeMenus ui.MenuPresentation
	ui.NewTester(func(c *ui.Context) {
		c.SetTheme(parentTokens.Theme())
		c.SetMenuPresentation(ui.PopoverMenus)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Children(func() {
			anchor := control.Button(c, "Themed", control.ButtonProps{})
			if err := Popover(c, h, "themed", anchor, &themedOpen, PopoverProps{Plain: false, SystemStyle: true}, func(child *ui.Context) {
				themedTokens = *theme.TokensOf(child)
				themedMenus = child.MenuPresentation()
				ui.Button(child, "Themed content")
			}); err != nil {
				t.Fatal(err)
			}
			anchor = control.Button(c, "Native", control.ButtonProps{})
			if err := NativePopover(c, h, "native", anchor, &nativeOpen, PopoverProps{Plain: true}, func(child *ui.Context) {
				nativeTokens = *theme.TokensOf(child)
				nativeMenus = child.MenuPresentation()
				ui.Button(child, "System content")
			}); err != nil {
				t.Fatal(err)
			}
		})
	}, 500, 300)
	if len(b.surfaces) != 2 {
		t.Fatalf("surfaces=%d", len(b.surfaces))
	}
	if b.surfaces[0].spec.SystemPopover || !b.surfaces[1].spec.SystemPopover {
		t.Fatal("paired API did not force intended window presentation")
	}
	themed := ui.NewTester(b.surfaces[0].view, 288, 200)
	native := ui.NewTester(b.surfaces[1].view, 288, 200)
	if themedTokens != *parentTokens {
		t.Fatal("themed popover did not use parent semantic tokens")
	}
	if nativeTokens.Dark || nativeTokens.Primary == parentTokens.Primary {
		t.Fatal("native popover inherited parent shadcn appearance")
	}
	if themedMenus != ui.PopoverMenus || nativeMenus != ui.SystemMenus {
		t.Fatalf("child menu policies themed=%v native=%v", themedMenus, nativeMenus)
	}
	native.SetDark(true)
	if !nativeTokens.Dark {
		t.Fatal("native popover did not follow its own system appearance")
	}
	themed.SetDark(false)
	if !themedTokens.Dark || themedTokens.Primary != parentTokens.Primary {
		t.Fatal("themed popover lost parent appearance when platform source changed")
	}
}

func TestStyledPopoverEntranceAndReducedMotion(t *testing.T) {
	for _, reduced := range []bool{false, true} {
		t.Run(map[bool]string{false: "animated", true: "reduced"}[reduced], func(t *testing.T) {
			b := &fakeBackend{}
			h := NewHostWithBackend(b)
			open := true
			ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				anchor := control.Button(c, "Open", control.ButtonProps{})
				if err := Popover(c, h, "motion", anchor, &open, PopoverProps{}, func(c *ui.Context) { ui.Text(c, "Content") }); err != nil {
					t.Fatal(err)
				}
			}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
			child := ui.NewTester(b.surfaces[0].view, 288, 200)
			if child.FramePending() == reduced {
				t.Fatalf("entrance scheduling reduced=%v snapshot=%v pending=%v", reduced, h.entries["motion"].reduced, child.FramePending())
			}
			child.Advance(300 * time.Millisecond)
			if child.FramePending() {
				t.Fatal("native content entrance never settled")
			}
		})
	}
}

func TestCompactRoleOptionsComparePaddingValues(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	open := true
	padding := float32(0)
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		anchor := control.Button(c, "Open", control.ButtonProps{})
		localPadding := padding
		if err := Popover(c, h, "compact", anchor, &open, PopoverProps{ContentRole: ui.RoleGroup, ContentPadding: &localPadding, Label: "Menu host"}, func(c *ui.Context) { ui.Box(c).Size(80, 20).Role(ui.RoleMenu).Label("Menu") }); err != nil {
			t.Fatal(err)
		}
	}, 400, 200)
	surface := b.surfaces[0]
	child := ui.NewTester(surface.view, 200, 100)
	child.Advance(300 * time.Millisecond)
	group, ok := child.AccessByRole(ui.RoleGroup, "Menu host")
	if !ok {
		t.Fatal("compact wrapper retained dialog semantics")
	}
	menu, ok := child.AccessByRole(ui.RoleMenu, "Menu")
	if !ok || menu.Bounds.X != group.Bounds.X || menu.Bounds.Y != group.Bounds.Y {
		t.Fatalf("zero padding did not give content ownership: %+v %+v", group.Bounds, menu.Bounds)
	}
	parent.Frame()
	surface.invalidates = 0
	parent.Frame()
	if surface.invalidates != 0 {
		t.Fatal("equal padding in new pointer invalidated native content")
	}
	padding = 8
	parent.Frame()
	if surface.invalidates != 1 {
		t.Fatal("changed padding did not invalidate native content")
	}
}
