package overlay

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestToasterDefaultDoesNotAddCloseButton(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		Toast(c, ui.Toast{ID: "sample", Title: "Saved", Timeout: -1})
		Toaster(c)
	}, 500, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleButton && (node.Name == "Dismiss notification" || node.Name == "Close toast") {
			t.Fatal("source default unexpectedly included a close button")
		}
	}
}

func TestToasterOrdinaryErrorUsesPopoverSurfaceAndNormalText(t *testing.T) {
	tokens := theme.Light()
	tokens.Popover = ui.RGB(245, 235, 215)
	tokens.PopoverForeground = ui.RGB(25, 70, 120)
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Scope(c, tokens).Fill().Children(func() { Toast(c, ui.Toast{ID: "error", Title: "Failed", Type: "error", Timeout: -1}); Toaster(c) })
	}, 500, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	toast, ok := r.AccessByRole(ui.RoleStatus, "Failed")
	if !ok {
		t.Fatal("native status missing")
	}
	pixel := r.Image().RGBAAt(int(toast.Bounds.X+8), int(toast.Bounds.Y+toast.Bounds.H/2))
	want := color.RGBA{R: tokens.Popover.R, G: tokens.Popover.G, B: tokens.Popover.B, A: 255}
	if pixel != want {
		t.Fatalf("ordinary source error did not use Popover: got=%v want=%v", pixel, want)
	}
}

func toastTestView(state *ToasterState, props *ToasterProps) func(*ui.Context) {
	return func(c *ui.Context) { theme.Apply(c); p := *props; p.State = state; ToasterWith(c, p) }
}
func toastTestNode(t *testing.T, r *ui.Tester, title string) ui.AccessInfo {
	t.Helper()
	n, ok := r.AccessByRole(ui.RoleListItem, title)
	if !ok {
		t.Fatalf("missing toast %q: %+v", title, r.AccessAll())
	}
	return n
}
func TestToasterAppQueueVisibleLimitAndExpandedActualHeights(t *testing.T) {
	var state ToasterState
	limit := 5
	gap := float32(19)
	p := ToasterProps{VisibleToasts: &limit, Gap: &gap, Expand: true}
	for i := 0; i < 5; i++ {
		state.Show(ui.Toast{Title: fmt.Sprintf("Notice %d", i), Description: strings.Repeat("Long description ", i+1), Timeout: -1})
	}
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 1000, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(50 * time.Millisecond)
	var last ui.AccessInfo
	for i := 4; i >= 0; i-- {
		n := toastTestNode(t, r, fmt.Sprintf("Notice %d", i))
		if i < 4 && math.Abs(float64(last.Bounds.Y-n.Bounds.Y-n.Bounds.H-gap)) > 1 {
			t.Fatalf("actual heights/gap overlap: newer=%+v older=%+v", last.Bounds, n.Bounds)
		}
		last = n
	}
	if len(state.Toasts()) != 5 {
		t.Fatal("app queue inherited native fixed cap")
	}
}
func TestToasterAppPositionsOffsetsAndMobile(t *testing.T) {
	for _, pos := range []ToastPosition{ToastTopLeft, ToastTopCenter, ToastTopRight, ToastBottomLeft, ToastBottomCenter, ToastBottomRight} {
		t.Run(string(pos), func(t *testing.T) {
			var state ToasterState
			state.Show(ui.Toast{Title: "Position", Timeout: -1})
			offset := ToastInsets{31, 37, 41, 43}
			p := ToasterProps{Position: pos, Offset: &offset}
			r := ui.NewTesterWith(toastTestView(&state, &p), 900, 600, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			r.Advance(30 * time.Millisecond)
			b := toastTestNode(t, r, "Position").Bounds
			if strings.HasPrefix(string(pos), "top") {
				if b.Y != 31 {
					t.Fatal(b)
				}
			} else if b.Y+b.H != 559 {
				t.Fatal(b)
			}
			if strings.HasSuffix(string(pos), "left") {
				if b.X != 43 {
					t.Fatal(b)
				}
			} else if strings.HasSuffix(string(pos), "right") {
				if b.X+b.W != 863 {
					t.Fatal(b)
				}
			} else if math.Abs(float64(b.X+b.W/2-450)) > .1 {
				t.Fatal(b)
			}
			r.SetSize(500, 600)
			r.Advance(30 * time.Millisecond)
			b = toastTestNode(t, r, "Position").Bounds
			if b.X != 16 || b.W != 468 {
				t.Fatalf("mobile width/offset %+v", b)
			}
		})
	}
}
func TestToasterTimersPauseHiddenWithoutInactiveConflation(t *testing.T) {
	var state ToasterState
	p := ToasterProps{}
	auto, dismiss, close := 0, 0, 0
	state.Show(ui.Toast{Title: "Timed", Timeout: time.Second, OnClose: func() { close++ }}, ToastOptions{OnAutoClose: func(ui.Toast) { auto++ }, OnDismiss: func(ui.Toast) { dismiss++ }})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(300 * time.Millisecond)
	r.SetWindowHidden(true)
	r.Advance(10 * time.Second)
	if len(state.Toasts()) != 1 {
		t.Fatal("hidden lifetime aged")
	}
	r.SetWindowHidden(false)
	r.Advance(500 * time.Millisecond)
	if len(state.Toasts()) != 1 {
		t.Fatal("remaining visible lifetime lost")
	}
	r.Advance(500 * time.Millisecond)
	if len(state.Toasts()) != 0 || auto != 1 || dismiss != 0 || close != 1 {
		t.Fatalf("auto=%d dismiss=%d close=%d queue=%d", auto, dismiss, close, len(state.Toasts()))
	}
	state.Show(ui.Toast{Title: "Inactive", Timeout: time.Second})
	r.Advance(30 * time.Millisecond)
	r.SetWindowActive(false)
	r.Advance(1500 * time.Millisecond)
	if len(state.Toasts()) != 0 {
		t.Fatal("inactive window incorrectly treated as hidden")
	}
}
func TestToasterLoadingUpdateKeepsNativeActionIdentity(t *testing.T) {
	var state ToasterState
	p := ToasterProps{CloseButton: true}
	state.Show(ui.Toast{ID: "job", Title: "Loading", Type: "loading", Action: "Inspect", Timeout: time.Second})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(10 * time.Second)
	if len(state.Toasts()) != 1 {
		t.Fatal("loading expired")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Close toast"); ok {
		t.Fatal("loading exposed close")
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyTab)
	focusedBefore := r.Focused("Inspect")
	if !state.Update("job", ui.Toast{Title: "Complete", Type: "success", Action: "Inspect", Timeout: time.Second}) {
		t.Fatal("missing loading")
	}
	r.Advance(30 * time.Millisecond)
	if focusedBefore && !r.Focused("Inspect") {
		t.Fatal("same-ID update lost native action focus")
	}
	r.Move(0, 0)
	r.ClickAt(0, 0)
	r.Advance(1500 * time.Millisecond)
	if len(state.Toasts()) != 0 {
		t.Fatal("completed toast did not acquire finite lifetime")
	}
}
func TestToasterActionPreventionCancelAndDismissCallbacks(t *testing.T) {
	var state ToasterState
	p := ToasterProps{CloseButton: true}
	actions, cancels, dismiss, auto, closed := 0, 0, 0, 0, 0
	state.Show(ui.Toast{Title: "Choose", Action: "Do", Timeout: -1, OnClose: func() { closed++ }}, ToastOptions{Cancel: "Later", OnCancel: func() { cancels++ }, Action: func(e *ToastActionEvent) { actions++; e.PreventDefault() }, OnDismiss: func(ui.Toast) { dismiss++ }, OnAutoClose: func(ui.Toast) { auto++ }})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Do"); err != nil {
		t.Fatal(err)
	}
	if actions != 1 || len(state.Toasts()) != 1 {
		t.Fatal("preventDefault ignored")
	}
	if err := r.Click("Later"); err != nil {
		t.Fatal(err)
	}
	r.Advance(30 * time.Millisecond)
	if cancels != 1 || dismiss != 0 || auto != 0 || closed != 1 || len(state.Toasts()) != 0 {
		t.Fatalf("cancel callbacks %d %d %d %d queue=%d", cancels, dismiss, auto, closed, len(state.Toasts()))
	}
	r.Advance(300 * time.Millisecond)
	state.Show(ui.Toast{Title: "Closable", Timeout: -1}, ToastOptions{OnDismiss: func(ui.Toast) { dismiss++ }})
	r.Advance(30 * time.Millisecond)
	if err := r.Click("Close toast"); err != nil {
		t.Fatal(err)
	}
	r.Advance(30 * time.Millisecond)
	if dismiss != 1 || len(state.Toasts()) != 0 {
		t.Fatal("explicit close callback missing")
	}
}
func TestToasterDismissibleFalseRetainsAction(t *testing.T) {
	var state ToasterState
	no := false
	p := ToasterProps{CloseButton: true}
	n := 0
	state.Show(ui.Toast{Title: "Locked", Action: "Continue", Timeout: -1}, ToastOptions{Dismissible: &no, Cancel: "Cancel", OnCancel: func() { t.Fatal("disabled cancel invoked") }, Action: func(*ToastActionEvent) { n++ }})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	for _, name := range []string{"Cancel", "Close toast"} {
		node, _ := r.AccessByRole(ui.RoleButton, name)
		if !node.Disabled {
			t.Fatalf("%s enabled", name)
		}
	}
	if err := r.Click("Continue"); err != nil {
		t.Fatal(err)
	}
	r.Advance(30 * time.Millisecond)
	if n != 1 || len(state.Toasts()) != 0 {
		t.Fatal("dismissible false incorrectly disabled action")
	}
}
func TestToasterDefaultExpandDoesNotPauseLifetime(t *testing.T) {
	var state ToasterState
	p := ToasterProps{Expand: true}
	state.Show(ui.Toast{Title: "Expanded", Timeout: time.Second})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(1500 * time.Millisecond)
	if len(state.Toasts()) != 0 {
		t.Fatal("expand-by-default permanently paused timer")
	}
}
func TestToasterHotkeyEscapeCollapsesAndRestoresExplicitFocus(t *testing.T) {
	var state ToasterState
	var trigger ui.Handle
	p := ToasterProps{ReturnFocus: &trigger}
	state.Show(ui.Toast{Title: "Keyboard", Action: "Review", Timeout: -1})
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Button(c, "Trigger").Bind(&trigger)
		p.State = &state
		ToasterWith(c, p)
	}, 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Trigger"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Alt, ui.KeyT)
	r.Advance(30 * time.Millisecond)
	if !state.expanded {
		t.Fatal("hotkey did not expand")
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyEscape)
	r.Advance(30 * time.Millisecond)
	if len(state.Toasts()) != 1 {
		t.Fatal("Escape dismissed notification")
	}
	if state.expanded {
		t.Fatal("Escape did not collapse")
	}
	if !r.Focused("Trigger") {
		t.Fatal("explicit return handle not restored")
	}
}

func TestToasterCustomHotkeyAdvertisesAndFocusesActualNativeRegion(t *testing.T) {
	var state ToasterState
	state.Show(ui.Toast{Title: "Custom shortcut", Timeout: -1})
	p := ToasterProps{Hotkey: &ToastHotkey{Modifiers: ui.Cmd | ui.Shift, Key: ui.KeyK}}
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	region, ok := r.AccessByRole(ui.RoleRegion, "Notifications")
	if !ok || !strings.Contains(region.KeyShortcuts, "K") || strings.Contains(region.KeyShortcuts, "T") {
		t.Fatalf("native region advertised the wrong shortcut: %+v", region)
	}
	r.Key(ui.Alt, ui.KeyT)
	if state.expanded {
		t.Fatal("replaced default shortcut still activated the region")
	}
	r.Key(ui.Cmd|ui.Shift, ui.KeyK)
	r.Key(0, ui.KeyTab)
	if !r.Focused("Custom shortcut") {
		t.Fatal("advertised custom shortcut did not enter the actual native notification")
	}
}

func TestToasterOrdinaryAndRichStatusPixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, rich := range []bool{false, true} {
			t.Run(fmt.Sprintf("dark=%v/rich=%v", dark, rich), func(t *testing.T) {
				tokens := theme.Light()
				if dark {
					tokens = theme.Dark()
				}
				tokens.Popover = ui.RGB(245, 235, 215)
				tokens.PopoverForeground = ui.RGB(25, 70, 120)
				var state ToasterState
				state.Show(ui.Toast{Title: "Error", Type: "error", Timeout: -1})
				p := ToasterProps{State: &state, RichColors: rich}
				r := ui.NewTesterWith(func(c *ui.Context) { theme.Scope(c, tokens).Fill().Children(func() { ToasterWith(c, p) }) }, 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				r.Advance(30 * time.Millisecond)
				b := toastTestNode(t, r, "Error").Bounds
				got := r.Image().RGBAAt(int(b.X+8), int(b.Y+b.H/2))
				want := color.RGBA{245, 235, 215, 255}
				if rich {
					want = color.RGBA{255, 240, 240, 255}
					if dark {
						want = color.RGBA{45, 6, 7, 255}
					}
				}
				if got != want {
					t.Fatalf("surface %v want %v", got, want)
				}
			})
		}
	}
}
func TestToasterNormalRemovalHasVisibleExitAndNoGhost(t *testing.T) {
	var state ToasterState
	state.Show(ui.Toast{Title: "Leaving", Timeout: -1})
	p := ToasterProps{}
	r := ui.NewTester(toastTestView(&state, &p), 800, 400)
	r.Advance(time.Second)
	b := toastTestNode(t, r, "Leaving").Bounds
	state.Dismiss("")
	r.Frame()
	atStart := r.Image()
	r.Advance(100 * time.Millisecond)
	middle := r.Image()
	// The translated, partially transparent surface still paints below its old
	// bottom at 100ms; source removal at 200ms leaves no native transition copy.
	x, y := int(b.X+60), int(b.Y+b.H+5)
	start := atStart.RGBAAt(x, y)
	during := middle.RGBAAt(x, y)
	if during == start {
		t.Fatalf("removal jumped invisible instead of painting an exit: start=%v mid=%v", start, during)
	}
	r.Advance(100 * time.Millisecond)
	r.Frame() // Native timer wakes have millisecond granularity at this boundary.
	after := r.Image().RGBAAt(x, y)
	if after != (color.RGBA{255, 255, 255, 255}) {
		t.Fatalf("native ghost remained beyond source 200ms lifecycle: start=%v after=%v", start, after)
	}
	if len(state.entries) != 0 {
		t.Fatal("entry survived removal lifecycle")
	}
}
func TestToasterHiddenRowsAreNotKeyboardStopsAndKeepAging(t *testing.T) {
	var state ToasterState
	limit := 1
	p := ToasterProps{VisibleToasts: &limit}
	state.Show(ui.Toast{Title: "Old", Action: "Old action", Timeout: time.Second})
	state.Show(ui.Toast{Title: "Front", Action: "Front action", Timeout: -1})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if _, ok := r.AccessByRole(ui.RoleButton, "Old action"); ok {
		t.Fatal("invisible old row remained accessible")
	}
	for i := 0; i < 6; i++ {
		r.Key(0, ui.KeyTab)
		if r.Focused("Old action") {
			t.Fatal("invisible row caught native keyboard focus")
		}
	}
	// Release focused/hover pause before advancing the hidden old model.
	r.ClickAt(0, 0)
	r.Move(0, 0)
	r.Advance(1500 * time.Millisecond)
	if models := state.Toasts(); len(models) != 1 || models[0].Title != "Front" {
		t.Fatalf("invisible row waited rather than aged: %v", models)
	}
}
func TestToasterCallbackUsesSnapshotAndPreventDefaultSurvivesSameIDUpdate(t *testing.T) {
	var state ToasterState
	p := ToasterProps{}
	old, newCalls := 0, 0
	state.Show(ui.Toast{ID: "same", Title: "Before", Action: "Apply", Timeout: -1, OnAction: func() {
		state.Update("same", ui.Toast{Title: "After", Action: "Apply", Timeout: -1}, ToastOptions{Action: func(*ToastActionEvent) { newCalls++ }})
	}}, ToastOptions{Action: func(e *ToastActionEvent) { old++; e.PreventDefault() }})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Apply"); err != nil {
		t.Fatal(err)
	}
	r.Advance(30 * time.Millisecond)
	if old != 1 || newCalls != 0 || len(state.Toasts()) != 1 || !r.HasText("After") {
		t.Fatalf("callback snapshot/default prevention lost: old=%d new=%d models=%v", old, newCalls, state.Toasts())
	}
}

func TestToasterChangingDurationResetsDefaultLifetime(t *testing.T) {
	var state ToasterState
	p := ToasterProps{Duration: time.Second}
	state.Show(ui.Toast{Title: "Default timer"})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(500 * time.Millisecond)
	p.Duration = 2 * time.Second
	r.Frame()
	r.Advance(time.Second)
	if len(state.Toasts()) != 1 {
		t.Fatal("changed toaster duration did not reset inherited lifetime")
	}
	r.Advance(1100 * time.Millisecond)
	if len(state.Toasts()) != 0 {
		t.Fatal("changed lifetime did not expire")
	}
}

func TestToasterFontSansOverridesMonospaceAncestorPixels(t *testing.T) {
	render := func(ancestor string) *ui.Tester {
		var state ToasterState
		state.Show(ui.Toast{Title: "Minimum width: iii WWW", Description: "Readable notifications", Timeout: -1})
		return ui.NewTesterWith(func(c *ui.Context) {
			theme.Apply(c)
			native := *c.Theme()
			native.Font = ""
			ui.Themed(c, &native).Children(func() {
				ui.Column(c).Font(ancestor).Fill().Children(func() { ToasterWith(c, ToasterProps{State: &state}) })
			})
		}, 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	}
	a, b := render("monospace"), render("system-ui")
	a.Advance(30 * time.Millisecond)
	b.Advance(30 * time.Millisecond)
	if !bytes.Equal(a.Image().Pix, b.Image().Pix) {
		t.Fatal("toast font-sans inherited monospace instead of native body fallback")
	}
}

func TestToasterUnstyledContentOwnsEntireBodyAndPreservesNativeEditor(t *testing.T) {
	var state ToasterState
	value := ""
	p := ToasterProps{CloseButton: true}
	closed := 0
	state.Show(ui.Toast{ID: "custom", Title: "Custom notice", Action: "Default action", Type: "error", Timeout: -1, OnClose: func() { closed++ }}, ToastOptions{Unstyled: true, Content: func(c *ui.Context) {
		ui.Column(c).FillWidth().Padding(8).Gap(4).Background(ui.RGB(20, 80, 140)).Children(func() {
			control.Input(c, &value, control.InputProps{Label: "Custom editor"})
			if ui.Button(c, "Custom dismiss").Clicked() {
				state.Dismiss("custom")
			}
		})
	}})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(30 * time.Millisecond)
	if _, ok := r.AccessByRole(ui.RoleButton, "Close toast"); ok {
		t.Fatal("unstyled content acquired default close chrome")
	}
	if r.HasText("Default action") {
		t.Fatal("unstyled content acquired default action")
	}
	b := toastTestNode(t, r, "Custom notice").Bounds
	pixel := r.Image().RGBAAt(int(b.X+4), int(b.Y+4))
	if pixel != (color.RGBA{20, 80, 140, 255}) {
		t.Fatalf("unstyled body retained default padding/surface: %v", pixel)
	}
	if err := r.Click("Custom editor"); err != nil {
		t.Fatal(err)
	}
	r.Type("native")
	state.Update("custom", ui.Toast{Title: "Updated custom", Timeout: -1, OnClose: func() { closed++ }})
	r.Frame()
	if !r.Focused("Custom editor") {
		t.Fatal("custom editor lost native identity during same-ID update")
	}
	r.Type(" editor")
	if value != "native editor" {
		t.Fatalf("custom native input lost value/selection: %q", value)
	}
	if err := r.Click("Custom dismiss"); err != nil {
		t.Fatal(err)
	}
	r.Advance(250 * time.Millisecond)
	if len(state.Toasts()) != 0 || closed != 1 {
		t.Fatalf("custom dismissal failed: models=%v closed=%d", state.Toasts(), closed)
	}
}

func TestToasterPointerLeaveResumesTimerDespiteFocusedAction(t *testing.T) {
	var state ToasterState
	p := ToasterProps{}
	state.Show(ui.Toast{Title: "Focused action", Action: "Keep", Timeout: time.Second}, ToastOptions{Action: func(e *ToastActionEvent) { e.PreventDefault() }})
	r := ui.NewTesterWith(toastTestView(&state, &p), 800, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Keep"); err != nil {
		t.Fatal(err)
	}
	if !r.Focused("Keep") {
		t.Fatal("native button did not retain focus")
	}
	r.Advance(2 * time.Second)
	if len(state.Toasts()) != 1 {
		t.Fatal("hover failed to pause")
	}
	r.Move(0, 0)
	r.Advance(1500 * time.Millisecond)
	if len(state.Toasts()) != 0 {
		t.Fatal("pointer leave incorrectly retained focus-based timer pause")
	}
}
