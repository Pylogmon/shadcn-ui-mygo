package desktop

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type presenceBackend struct {
	surfaces    []*presenceSurface
	invalidates int
}
type presenceSurface struct {
	spec                Spec
	view                func(*ui.Context)
	onClosed            func()
	onDismiss           func() bool
	closes, invalidates int
}

func (b *presenceBackend) Open(s Spec, view func(*ui.Context), closed func()) (Surface, error) {
	f := &presenceSurface{spec: s, view: view, onClosed: closed}
	b.surfaces = append(b.surfaces, f)
	return f, nil
}
func (b *presenceBackend) InvalidateParent()                     { b.invalidates++ }
func (f *presenceSurface) SetAnchor(r ui.Rect)                   { f.spec.Anchor = r }
func (f *presenceSurface) Invalidate()                           { f.invalidates++ }
func (f *presenceSurface) SetDismissHandler(handler func() bool) { f.onDismiss = handler }
func (f *presenceSurface) Close()                                { f.closes++; f.onClosed() }
func (f *presenceSurface) dismiss() {
	if f.onDismiss == nil || !f.onDismiss() {
		f.Close()
	}
}

func presenceParent(h *Host, open, visible *bool, dismissed *int, reduced bool) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		if *visible {
			anchor := control.Button(c, "Open presence", control.ButtonProps{})
			if err := Popover(c, h, "presence", anchor, open, PopoverProps{OnDismiss: func() { *dismissed++ }}, func(child *ui.Context) { control.Button(child, "Popup action", control.ButtonProps{}) }); err != nil {
				panic(err)
			}
		}
	}, 400, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
}

func TestPlainFlyoutRetainsSurfaceForExitThenClosesOnce(t *testing.T) {
	b := &presenceBackend{}
	h := NewHostWithBackend(b)
	open, visible := true, true
	dismissed := 0
	parent := presenceParent(h, &open, &visible, &dismissed, false)
	f := b.surfaces[0]
	child := ui.NewTester(f.view, 288, 200)
	child.Advance(300 * time.Millisecond)
	button, ok := child.Find("Popup action")
	if !ok {
		t.Fatal("action missing before close")
	}
	px, py := int(button.X+8), int(button.Y+button.H/2)
	before := child.Image().RGBAAt(px, py)
	open = false
	parent.Frame()
	if f.closes != 0 || h.IsOpen("presence") || !h.IsPresent("presence") {
		t.Fatalf("exit did not retain logical closed surface: closes=%d open=%v present=%v", f.closes, h.IsOpen("presence"), h.IsPresent("presence"))
	}
	child.Frame()
	if !child.FramePending() {
		t.Fatal("exit scheduled no animation frames")
	}
	if _, ok := child.AccessByRole(ui.RoleButton, "Popup action"); ok {
		t.Fatal("leaving flyout still exposed interactive content")
	}
	child.Advance(50 * time.Millisecond)
	if f.closes != 0 {
		t.Fatal("native surface closed before exit completed")
	}
	middle := child.Image().RGBAAt(px, py)
	if middle == before {
		t.Fatalf("exit scheduled frames but painted unchanged pixels: before=%v middle=%v", before, middle)
	}
	child.Advance(100 * time.Millisecond)
	if f.closes != 1 || h.IsPresent("presence") || dismissed != 0 {
		t.Fatalf("close once/policy violated: closes=%d present=%v dismiss=%d", f.closes, h.IsPresent("presence"), dismissed)
	}
	parent.Frame()
	child.Advance(500 * time.Millisecond)
	if f.closes != 1 || child.FramePending() {
		t.Fatal("exit repeated or kept an idle animation loop")
	}
	if !parent.Focused("Open presence") {
		t.Fatal("exit completion did not restore anchor focus")
	}
}

func TestPlainFlyoutExplicitCloseCanReverseAndLaterRemoveEntry(t *testing.T) {
	b := &presenceBackend{}
	h := NewHostWithBackend(b)
	open, visible := true, true
	dismissed := 0
	parent := presenceParent(h, &open, &visible, &dismissed, false)
	f := b.surfaces[0]
	child := ui.NewTester(f.view, 288, 200)
	child.Advance(300 * time.Millisecond)
	h.Close("presence")
	child.Frame()
	child.Advance(30 * time.Millisecond)
	open = true
	parent.Frame()
	child.Frame()
	child.Advance(200 * time.Millisecond)
	if f.closes != 0 || len(b.surfaces) != 1 || !h.IsOpen("presence") {
		t.Fatal("explicit close could not reverse in retained native window")
	}
	h.Close("presence")
	child.Frame()
	child.Advance(150 * time.Millisecond)
	if f.closes != 1 || h.entries["presence"] != nil || dismissed != 0 {
		t.Fatal("completed explicit close retained entry or delivered native dismissal")
	}
	open = true
	parent.Frame()
	if len(b.surfaces) != 2 {
		t.Fatal("menu key could not reopen after completed close")
	}
	f.onClosed()
	if !open || !h.IsOpen("presence") {
		t.Fatal("old closed callback affected new window")
	}
}

func TestPlainFlyoutExitReversesWithoutReplacingNativeWindow(t *testing.T) {
	b := &presenceBackend{}
	h := NewHostWithBackend(b)
	open, visible := true, true
	dismissed := 0
	parent := presenceParent(h, &open, &visible, &dismissed, false)
	f := b.surfaces[0]
	child := ui.NewTester(f.view, 288, 200)
	child.Advance(300 * time.Millisecond)
	f.dismiss()
	parent.Frame()
	child.Frame()
	child.Advance(35 * time.Millisecond)
	if open || f.closes != 0 || dismissed != 0 {
		t.Fatal("native dismissal was not retained")
	}
	open = true
	parent.Frame()
	child.Frame()
	child.Advance(250 * time.Millisecond)
	if len(b.surfaces) != 1 || f.closes != 0 || !h.IsOpen("presence") || dismissed != 0 {
		t.Fatal("reverse replaced or closed surface / delivered cancelled dismissal")
	}
	if !child.HasText("Popup action") {
		t.Fatal("reopened content was not accessible")
	}
	if child.FramePending() {
		t.Fatal("reverse animation did not settle")
	}
}

func TestPlainFlyoutNativeDismissalIsPreventedAndNotifiedAfterExit(t *testing.T) {
	b := &presenceBackend{}
	h := NewHostWithBackend(b)
	open, visible := true, true
	dismissed := 0
	parent := presenceParent(h, &open, &visible, &dismissed, false)
	f := b.surfaces[0]
	child := ui.NewTester(f.view, 288, 200)
	child.Advance(300 * time.Millisecond)
	f.dismiss()
	f.dismiss()
	if open || f.closes != 0 || dismissed != 0 {
		t.Fatal("native close was not prevented until presence finishes")
	}
	parent.Frame()
	child.Frame()
	child.Advance(160 * time.Millisecond)
	parent.Frame()
	if f.closes != 1 || dismissed != 1 || h.IsPresent("presence") {
		t.Fatalf("closes=%d dismissed=%d present=%v", f.closes, dismissed, h.IsPresent("presence"))
	}
	f.onClosed()
	parent.Frame()
	if dismissed != 1 {
		t.Fatal("late native callback delivered OnDismiss twice")
	}
	if !parent.Focused("Open presence") {
		t.Fatal("native dismiss completion lost focus restoration")
	}
}

func TestLeavingPlainFlyoutDoesNotActivateItsOldControls(t *testing.T) {
	b := &presenceBackend{}
	h := NewHostWithBackend(b)
	open := true
	actions := 0
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := control.Button(c, "Owner", control.ButtonProps{})
		_ = Popover(c, h, "inert", anchor, &open, PopoverProps{}, func(child *ui.Context) {
			if control.Button(child, "Delete", control.ButtonProps{}).Clicked() {
				actions++
			}
		})
	}, 400, 240)
	f := b.surfaces[0]
	child := ui.NewTester(f.view, 288, 200)
	child.Advance(300 * time.Millisecond)
	button, ok := child.Find("Delete")
	if !ok {
		t.Fatal("button absent")
	}
	open = false
	parent.Frame()
	child.Frame()
	child.ClickAt(button.X+button.W/2, button.Y+button.H/2)
	child.Key(0, ui.KeyEnter)
	if actions != 0 {
		t.Fatal("an exiting native flyout still activated old content")
	}
	child.Advance(150 * time.Millisecond)
	if f.closes != 1 {
		t.Fatal("inert exit did not finish")
	}
}

func TestPlainFlyoutReducedMotionAndOwningUITeardownAreImmediate(t *testing.T) {
	for _, mode := range []string{"reduced", "missing-anchor", "owner-closed"} {
		t.Run(mode, func(t *testing.T) {
			b := &presenceBackend{}
			h := NewHostWithBackend(b)
			open, visible := true, true
			dismissed := 0
			parent := presenceParent(h, &open, &visible, &dismissed, mode == "reduced")
			f := b.surfaces[0]
			child := ui.NewTester(f.view, 288, 200)
			child.Advance(300 * time.Millisecond)
			switch mode {
			case "reduced":
				open = false
				parent.Frame()
			case "missing-anchor":
				visible = false
				parent.Frame()
			case "owner-closed":
				h.CloseAll()
			}
			if f.closes != 1 || open || h.IsPresent("presence") {
				t.Fatalf("immediate teardown failed: closes=%d open=%v present=%v", f.closes, open, h.IsPresent("presence"))
			}
			child.Advance(300 * time.Millisecond)
			if f.closes != 1 || dismissed != 0 {
				t.Fatal("teardown repeated close or emitted user-dismiss callback")
			}
		})
	}
}

func TestSystemPopoverDoesNotInterceptOSOwnedDismissal(t *testing.T) {
	b := &presenceBackend{}
	h := NewHostWithBackend(b)
	open := true
	dismissed := 0
	ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		anchor := control.Button(c, "Native", control.ButtonProps{})
		_ = NativePopover(c, h, "native", anchor, &open, PopoverProps{OnDismiss: func() { dismissed++ }}, func(child *ui.Context) { ui.Text(child, "System") })
	}, 400, 240)
	f := b.surfaces[0]
	child := ui.NewTester(f.view, 288, 200)
	if f.onDismiss != nil {
		t.Fatal("system popover installed custom close interception")
	}
	f.dismiss()
	if open || f.closes != 1 || dismissed != 1 || h.IsPresent("native") {
		t.Fatal("system-native dismissal was not immediate / OS owned")
	}
	if child.FramePending() {
		t.Fatal("system popover scheduled custom presence")
	}
}

func TestNonFocusableStyleSwitchUpdatesDismissalCapabilityInPlace(t *testing.T) {
	b := &presenceBackend{}
	h := NewHostWithBackend(b)
	open, native := true, true
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := control.Button(c, "Hover", control.ButtonProps{})
		p := PopoverProps{NonFocusable: true, ShowInactive: true, IgnoreMouseEvents: true}
		content := func(child *ui.Context) { ui.Text(child, "Hover content") }
		if native {
			_ = NativePopover(c, h, "hover", anchor, &open, p, content)
		} else {
			_ = Popover(c, h, "hover", anchor, &open, p, content)
		}
	}, 400, 240)
	f := b.surfaces[0]
	child := ui.NewTester(f.view, 288, 200)
	if f.onDismiss != nil || !f.spec.ShowInactive || !f.spec.IgnoreMouseEvents {
		t.Fatal("native-style passive surface configuration wrong")
	}
	native = false
	parent.Frame()
	child.Frame()
	child.Advance(250 * time.Millisecond)
	if len(b.surfaces) != 1 || f.onDismiss == nil {
		t.Fatal("style switch failed to install interception without recreating equal native configuration")
	}
	f.dismiss()
	child.Frame()
	child.Advance(150 * time.Millisecond)
	if f.closes != 1 {
		t.Fatal("styled passive exit failed")
	}
}
