package desktop

import (
	"fmt"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestThemedMenuUsesIndependentFlyoutAndDeliversChoiceOnce(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	state := MenuState{}
	open := false
	callback, returns := 0, 0
	var menuErr error
	items := []overlay.MenuItem{{Key: "open", Label: "Open"}, {Key: "more", Label: "More", Children: []overlay.MenuItem{{Key: "archive", Label: "Archive"}}}}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := control.Button(c, "Actions", control.ButtonProps{})
		key, err := Menu(c, h, "actions", anchor, &open, &state, MenuProps{Label: "Actions menu", Items: items, OnChoose: func(key string) {
			if key != "archive" {
				t.Fatalf("unexpected callback %q", key)
			}
			callback++
		}})
		menuErr = err
		if key != "" {
			returns++
		}
	}, 320, 160)
	if err := parent.Click("Actions"); err != nil {
		t.Fatal(err)
	}
	if menuErr != nil || len(b.surfaces) != 1 || !open {
		t.Fatalf("menu open surfaces=%d open=%v err=%v", len(b.surfaces), open, menuErr)
	}
	if parent.HasText("Archive") || b.surfaces[0].spec.SystemPopover || !b.surfaces[0].spec.FitContent {
		t.Fatalf("menu did not use plain independent Flyout: %+v", b.surfaces[0].spec)
	}
	child := ui.NewTester(b.surfaces[0].view, 500, 230)
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyRight)
	if !child.HasText("Archive") || parent.HasText("Archive") {
		t.Fatal("submenu content did not belong to child UI")
	}
	child.Advance(110 * time.Millisecond)
	defaultPanel, _ := child.AccessByRole(ui.RoleMenu, "More")
	if defaultPanel.Bounds.W != 192 {
		t.Fatalf("unset submenu width changed: %v", defaultPanel.Bounds)
	}
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	parent.Frame()
	if callback != 1 || returns != 1 || open || h.IsOpen("actions") {
		t.Fatalf("callback=%d returns=%d open=%v", callback, returns, open)
	}
	if !parent.Focused("Actions") {
		t.Fatal("selection did not restore parent anchor focus")
	}
	items[1].SubmenuWidth = 176
	if err := parent.Click("Actions"); err != nil {
		t.Fatal(err)
	}
	surface := b.surfaces[len(b.surfaces)-1]
	child = ui.NewTester(surface.view, 600, 230)
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyRight)
	child.Advance(110 * time.Millisecond)
	panel, _ := child.AccessByRole(ui.RoleMenu, "More")
	if panel.Bounds.W != 176 {
		t.Fatalf("explicit owner width=%v", panel.Bounds)
	}
	items[1].SubmenuWidth = 240
	child.Key(0, ui.KeyDown) // independent child event precedes the parent
	panel, _ = child.AccessByRole(ui.RoleMenu, "More")
	if panel.Bounds.W != 240 {
		t.Fatalf("child retained stale owner width: %v", panel.Bounds)
	}
	invalidates, count := surface.invalidates, len(b.surfaces)
	parent.Frame()
	if surface.invalidates <= invalidates || surface.closes != 0 || len(b.surfaces) != count {
		t.Fatal("item width refresh ignored fingerprint or replaced the native surface")
	}
	items[1].SubmenuWidth = -1
	parent.Frame()
	child.Frame()
	panel, _ = child.AccessByRole(ui.RoleMenu, "More")
	if panel.Bounds.W != 192 {
		t.Fatalf("nonpositive owner width lost original shared fallback: %v", panel.Bounds)
	}
	for _, invalid := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		items[1].SubmenuWidth = invalid
		parent.Frame()
		if menuErr == nil || open || surface.closes != 1 || len(b.surfaces) != count {
			t.Fatal("non-finite nested width was accepted or reopened a surface")
		}
	}
}

func TestMenuArrowOpeningAndMissingHostError(t *testing.T) {
	for _, key := range []ui.Key{ui.KeyDown, ui.KeyUp} {
		t.Run(fmt.Sprint(key), func(t *testing.T) {
			b := &fakeBackend{}
			h := NewHostWithBackend(b)
			state := MenuState{}
			open := false
			parent := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				h.BeginFrame()
				defer h.EndFrame()
				anchor := control.Button(c, "Menu", control.ButtonProps{}).AutoFocus()
				_, err := Menu(c, h, "arrows", anchor, &open, &state, MenuProps{Items: []overlay.MenuItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}, {Key: "blocked", Label: "Unavailable", Disabled: true}}})
				if err != nil {
					t.Fatal(err)
				}
			}, 300, 150)
			parent.Key(0, key)
			if len(b.surfaces) != 1 || !open {
				t.Fatal("arrow did not open Flyout")
			}
			_ = ui.NewTester(b.surfaces[0].view, 250, 180)
			want := "one"
			if key == ui.KeyUp {
				want = "two"
			}
			if len(state.Highlighted) != 1 || state.Highlighted[0] != want {
				t.Fatalf("initial arrow highlight=%v want=%s", state.Highlighted, want)
			}
		})
	}
	state := MenuState{}
	open := false
	sawError := false
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		anchor := control.Button(c, "Missing host", control.ButtonProps{})
		_, err := Menu(c, nil, "missing", anchor, &open, &state, MenuProps{})
		if err != nil {
			sawError = true
		}
	}, 300, 150)
	if sawError {
		t.Fatal("closed menu unnecessarily requires native Host")
	}
	if err := parent.Click("Missing host"); err != nil {
		t.Fatal(err)
	}
	if !sawError || open {
		t.Fatal("missing Host did not return an opening error/reset state")
	}
}

func TestContextMenuSecondaryControlAndKeyboardAnchoring(t *testing.T) {
	for _, mode := range []string{"secondary", "control", "key", "shift-f10"} {
		t.Run(mode, func(t *testing.T) {
			b := &fakeBackend{}
			h := NewHostWithBackend(b)
			state := MenuState{}
			open := false
			dismissed := 0
			parent := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				h.BeginFrame()
				defer h.EndFrame()
				ui.Column(c).Padding(30).Children(func() {
					target := control.Button(c, "Context target", control.ButtonProps{}).AutoFocus()
					if _, err := ContextMenu(c, h, "context", target, &open, &state, MenuProps{Items: []overlay.MenuItem{{Key: "open", Label: "Open"}}, OnDismiss: func() { dismissed++ }}); err != nil {
						t.Fatal(err)
					}
				})
			}, 400, 240)
			target, _ := parent.AccessByRole(ui.RoleButton, "Context target")
			switch mode {
			case "secondary":
				if err := parent.RightClick("Context target"); err != nil {
					t.Fatal(err)
				}
			case "control":
				if err := parent.ClickWith(ui.Ctrl, "Context target"); err != nil {
					t.Fatal(err)
				}
			case "key":
				parent.Key(0, ui.KeyContextMenu)
			case "shift-f10":
				parent.Key(ui.Shift, ui.KeyF10)
			}
			if len(b.surfaces) != 1 || !open {
				t.Fatalf("context did not open: %d open=%v", len(b.surfaces), open)
			}
			r := b.surfaces[0].spec.Anchor
			if mode == "secondary" || mode == "control" {
				if r.X < target.Bounds.X || r.X > target.Bounds.X+target.Bounds.W || r.Y < target.Bounds.Y || r.Y > target.Bounds.Y+target.Bounds.H {
					t.Fatalf("pointer anchor=%v target=%v", r, target.Bounds)
				}
			} else if r.X != target.Bounds.X || r.Y != target.Bounds.Y+target.Bounds.H {
				t.Fatalf("keyboard anchor=%v target=%v", r, target.Bounds)
			}
			child := ui.NewTester(b.surfaces[0].view, 240, 160)
			child.Key(0, ui.KeyEscape)
			parent.Frame()
			if open || dismissed != 1 || !parent.Focused("Context target") {
				t.Fatalf("context dismiss open=%v count=%d focus=%v", open, dismissed, parent.Focused("Context target"))
			}
		})
	}
}

func TestNativeMenuScopesSystemPolicyWithoutChangingOtherMenus(t *testing.T) {
	choice := ""
	var policy ui.MenuPresentation
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		c.SetMenuPresentation(ui.PopoverMenus)
		value, err := NativeMenu(c, overlay.MenuProps{Items: []overlay.MenuItem{{Key: "open", Label: "Open native", SubmenuWidth: float32(math.NaN())}}}, func(anchor ui.Element) { control.StyleButton(c, anchor, "Native menu", control.ButtonProps{}) })
		if err != nil {
			t.Fatal(err)
		}
		if value != "" {
			choice = value
		}
		policy = c.MenuPresentation()
	}, 320, 160)
	if err := parent.Click("Native menu"); err != nil {
		t.Fatal(err)
	}
	if policy != ui.PopoverMenus || len(parent.Menu()) != 1 || parent.HasText("Open native") {
		t.Fatalf("native route/policy broken: menu=%v policy=%v", parent.Menu(), policy)
	}
	if err := parent.ChooseMenuItem("Open native"); err != nil {
		t.Fatal(err)
	}
	if choice != "open" {
		t.Fatalf("native menu choice=%q", choice)
	}
	// The legacy in-window adapter retains its original default sizing, but a
	// positive per-owner override is exact. Its disabled native rows dim once.
	icon := ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path fill="currentColor" d="M0 0h16v16H0z"/></svg>`))
	for _, disabled := range []bool{false, true} {
		legacyOpen := true
		r := ui.NewTesterWith(func(c *ui.Context) {
			tokens := theme.Preset("blue", false)
			tokens.Popover, tokens.Border, tokens.Background, tokens.PopoverForeground = ui.RGB(0, 0, 0), ui.RGB(0, 0, 0), ui.RGB(0, 0, 0), ui.RGB(255, 255, 255)
			theme.Scope(c, tokens).Fill().Background(tokens.Background).Padding(20).Children(func() {
				overlay.InWindowDropdownMenu(c, &legacyOpen, overlay.MenuProps{Items: []overlay.MenuItem{{Key: "row", Label: "Legacy row", Icon: icon, Disabled: disabled}}}, nil)
			})
		}, 320, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		row, ok := r.AccessByRole(ui.RoleMenuItem, "Legacy row")
		if !ok {
			t.Fatal("legacy menu did not expose its actual native row")
		}
		image, peak := r.Image(), uint8(0)
		for y := int(row.Bounds.Y) + 3; y < int(row.Bounds.Y+row.Bounds.H)-3; y++ {
			for x := int(row.Bounds.X) + 3; x < int(row.Bounds.X+row.Bounds.W)-3; x++ {
				peak = max(peak, color.NRGBAModel.Convert(image.At(x, y)).(color.NRGBA).R)
			}
		}
		want := uint8(255)
		if disabled {
			want = 128
		}
		if peak != want {
			t.Fatalf("legacy disabled=%v white=%d want%d", disabled, peak, want)
		}
	}
	legacyOpen := true
	legacyItems := []overlay.MenuItem{{Key: "more", Label: "Legacy more", SubmenuWidth: 176, Children: []overlay.MenuItem{{Key: "nested", Label: "Legacy nested", Children: []overlay.MenuItem{{Key: "leaf", Label: "Legacy leaf"}}}}}}
	legacy := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		overlay.InWindowDropdownMenu(c, &legacyOpen, overlay.MenuProps{Width: 208, Items: legacyItems}, nil)
	}, 800, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := legacy.Click("Legacy more"); err != nil {
		t.Fatal(err)
	}
	legacyPanel, _ := legacy.AccessByRole(ui.RoleMenu, "Legacy more")
	if legacyPanel.Bounds.W != 176 {
		t.Fatalf("legacy explicit submenu width=%v", legacyPanel.Bounds)
	}
	if err := legacy.Click("Legacy nested"); err != nil {
		t.Fatal(err)
	}
	legacyNested, _ := legacy.AccessByRole(ui.RoleMenu, "Legacy nested")
	if legacyNested.Bounds.W != 208 {
		t.Fatalf("legacy grandchild lost shared208 fallback: %v", legacyNested.Bounds)
	}
	for _, invalid := range []float32{float32(math.NaN()), float32(math.Inf(1))} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("legacy menu accepted a non-finite nested width")
				}
			}()
			legacyItems[0].SubmenuWidth = invalid
			legacy.Frame()
		}()
	}
}
