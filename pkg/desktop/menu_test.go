package desktop

import (
	"fmt"
	"testing"

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
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := control.Button(c, "Actions", control.ButtonProps{})
		key, err := Menu(c, h, "actions", anchor, &open, &state, MenuProps{Label: "Actions menu", Items: []overlay.MenuItem{{Key: "open", Label: "Open"}, {Key: "more", Label: "More", Children: []overlay.MenuItem{{Key: "archive", Label: "Archive"}}}}, OnChoose: func(key string) {
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
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	parent.Frame()
	if callback != 1 || returns != 1 || open || h.IsOpen("actions") {
		t.Fatalf("callback=%d returns=%d open=%v", callback, returns, open)
	}
	if !parent.Focused("Actions") {
		t.Fatal("selection did not restore parent anchor focus")
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
		value, err := NativeMenu(c, overlay.MenuProps{Items: []overlay.MenuItem{{Key: "open", Label: "Open native"}}}, func(anchor ui.Element) { control.StyleButton(c, anchor, "Native menu", control.ButtonProps{}) })
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
}
