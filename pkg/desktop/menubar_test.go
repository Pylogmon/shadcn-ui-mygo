package desktop

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func testMenubarItems() []overlay.MenubarItem {
	return []overlay.MenubarItem{
		{Key: "file", Label: "File", Items: []overlay.MenuItem{{Key: "new", Label: "New"}, {Key: "recent", Label: "Recent", Children: []overlay.MenuItem{{Key: "one", Label: "First project"}}}}},
		{Key: "disabled", Label: "Unavailable", Disabled: true, Items: []overlay.MenuItem{{Key: "bad", Label: "Disabled action"}}},
		{Key: "edit", Label: "Edit", Items: []overlay.MenuItem{{Key: "copy", Label: "Copy"}, {Key: "paste", Label: "Paste"}}},
		{Key: "view", Label: "View", Items: []overlay.MenuItem{{Key: "zoom", Label: "Zoom"}}},
	}
}

func menubarTester(t *testing.T, b *fakeBackend, s *MenubarState, p *MenubarProps, rtl bool, before, after *ui.Handle, result func(string)) (*ui.Tester, *Host) {
	t.Helper()
	h := NewHostWithBackend(b)
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		column := ui.Column(c).Padding(16).Gap(8)
		if rtl {
			column.Direction(ui.RightToLeft)
		}
		column.Children(func() {
			ui.Button(c, "Before").Bind(before)
			value, err := Menubar(c, h, "main-menubar", s, *p)
			if err != nil {
				t.Fatal(err)
			}
			if value != "" && result != nil {
				result(value)
			}
			ui.Button(c, "After").Bind(after)
		})
	}, 520, 230)
	return r, h
}

func menubarChild(t *testing.T, b *fakeBackend) *ui.Tester {
	t.Helper()
	if len(b.surfaces) != 1 {
		t.Fatalf("menubar surfaces=%d", len(b.surfaces))
	}
	r := ui.NewTester(b.surfaces[0].view, 500, 260)
	r.Advance(250 * time.Millisecond)
	return r
}

func TestMenubarKeyboardGroupAndArrowOpening(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprintf("rtl=%v", rtl), func(t *testing.T) {
			b, s := &fakeBackend{}, &MenubarState{}
			p := MenubarProps{Items: testMenubarItems()}
			parent, _ := menubarTester(t, b, s, &p, rtl, nil, nil, nil)
			parent.Key(0, ui.KeyTab)
			parent.Key(0, ui.KeyTab)
			if !parent.Focused("File") {
				t.Fatal("Tab did not enter the menubar at its first enabled trigger")
			}
			next := ui.KeyRight
			if rtl {
				next = ui.KeyLeft
			}
			parent.Key(0, next)
			if !parent.Focused("Edit") || s.Open {
				t.Fatal("closed-bar arrow did not skip disabled item without opening")
			}
			parent.Key(0, ui.KeyUp)
			child := menubarChild(t, b)
			if !s.Open || s.Active != "edit" || len(s.menu.Highlighted) != 1 || s.menu.Highlighted[0] != "paste" {
				t.Fatalf("Up did not open at last enabled item: %+v", s)
			}
			if parent.HasText("Copy") || b.surfaces[0].spec.SystemPopover || !child.HasText("Copy") {
				t.Fatal("themed menu did not belong to an independent plain Flyout")
			}
			child.Key(0, ui.KeyEscape)
			parent.Frame()
			if s.Open || !parent.Focused("Edit") || b.surfaces[0].closes != 1 {
				t.Fatal("Escape did not close once and restore active trigger")
			}
			parent.Key(0, ui.KeyTab)
			if !parent.Focused("After") {
				t.Fatal("closed menubar was more than one Tab stop")
			}
		})
	}
}

func TestMenubarChildSiblingSwitchSubmenuPriorityAndReanchor(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprintf("rtl=%v", rtl), func(t *testing.T) {
			b, s := &fakeBackend{}, &MenubarState{}
			p := MenubarProps{Items: testMenubarItems()}
			parent, _ := menubarTester(t, b, s, &p, rtl, nil, nil, nil)
			if err := parent.Click("File"); err != nil {
				t.Fatal(err)
			}
			child := menubarChild(t, b)
			anchor := b.surfaces[0].spec.Anchor
			forward, backward := ui.KeyRight, ui.KeyLeft
			if rtl {
				forward, backward = backward, forward
			}
			child.Key(0, ui.KeyEnd)
			child.Key(0, forward)
			if s.Active != "file" || !child.HasText("First project") || s.menu.ActiveDepth != 1 {
				t.Fatal("submenu arrow switched top-level menus prematurely")
			}
			rootPanel, _ := child.AccessByRole(ui.RoleMenu, "File")
			subPanel, _ := child.AccessByRole(ui.RoleMenu, "Recent")
			if rtl && subPanel.Bounds.X >= rootPanel.Bounds.X || !rtl && subPanel.Bounds.X <= rootPanel.Bounds.X {
				t.Fatal("independent child did not inherit the parent's submenu direction")
			}
			child.Key(0, backward)
			if s.Active != "file" || s.menu.ActiveDepth != 0 {
				t.Fatal("submenu close arrow skipped closing its own submenu")
			}
			child.Key(0, ui.KeyHome)
			child.Key(0, forward)
			parent.Frame()
			child.Frame()
			if s.Active != "edit" || !s.Open || len(b.surfaces) != 1 || b.surfaces[0].closes != 0 || !child.HasText("Copy") || child.HasText("New") {
				t.Fatalf("sibling switch replaced/lost its Flyout: active=%q open=%v surfaces=%d texts=%v", s.Active, s.Open, len(b.surfaces), child.Texts())
			}
			if anchor == b.surfaces[0].spec.Anchor {
				t.Fatal("same native Flyout was not reanchored to Edit")
			}
			child.Key(0, backward)
			parent.Frame()
			child.Frame()
			if s.Active != "file" || !child.HasText("New") {
				t.Fatal("root close direction did not switch to previous enabled menu")
			}
		})
	}
}

func TestMenubarHoverSwitchAndDynamicContentDoNotScheduleIdleFrames(t *testing.T) {
	b, s := &fakeBackend{}, &MenubarState{}
	p := MenubarProps{Items: testMenubarItems()}
	parent, _ := menubarTester(t, b, s, &p, false, nil, nil, nil)
	view, _ := parent.Find("View")
	parent.Move(view.X+view.W/2, view.Y+view.H/2)
	if s.Open || len(b.surfaces) != 0 {
		t.Fatal("closed menubar opened merely on hover")
	}
	if err := parent.Click("File"); err != nil {
		t.Fatal(err)
	}
	child := menubarChild(t, b)
	edit, _ := parent.Find("Edit")
	parent.Move(edit.X+edit.W/2, edit.Y+edit.H/2)
	child.Frame()
	if s.Active != "edit" || !child.HasText("Copy") || len(b.surfaces) != 1 {
		t.Fatalf("hover did not switch the open sibling: active=%q texts=%v", s.Active, child.Texts())
	}
	p.Items[2].Items[0].Label = "Copy updated"
	parent.Frame()
	if b.surfaces[0].invalidates == 0 {
		t.Fatal("structural menu update did not invalidate independent child")
	}
	child.Frame()
	if !child.HasText("Copy updated") {
		t.Fatal("independent menu retained stale content")
	}
	parent.Advance(time.Second)
	child.Advance(time.Second)
	b.surfaces[0].invalidates, b.surfaces[0].moves = 0, 0
	parent.Frame()
	if b.surfaces[0].invalidates != 0 || b.surfaces[0].moves != 0 || parent.FramePending() || child.FramePending() {
		t.Fatal("idle menubar scheduled native moves or redraws")
	}
}

func TestMenubarSelectionCallbackAndReturnEachOccurOnce(t *testing.T) {
	b, s := &fakeBackend{}, &MenubarState{}
	callbacks, returns, dismissals := 0, 0, 0
	checked, closeOnSelect := false, true
	p := MenubarProps{Items: []overlay.MenubarItem{{Key: "view", Label: "View", Items: []overlay.MenuItem{{Key: "grid", Label: "Show grid", Checked: &checked, CloseOnSelect: &closeOnSelect}}}},
		OnChoose: func(value string) {
			callbacks++
			if value != "grid" {
				t.Fatal(value)
			}
		}, OnDismiss: func() { dismissals++ }}
	parent, _ := menubarTester(t, b, s, &p, false, nil, nil, func(value string) {
		returns++
		if value != "grid" {
			t.Fatal(value)
		}
	})
	if err := parent.Click("View"); err != nil {
		t.Fatal(err)
	}
	child := menubarChild(t, b)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	parent.Frame()
	if !checked || callbacks != 1 || returns != 1 || dismissals != 0 || s.Open || b.surfaces[0].closes != 1 || !parent.Focused("View") {
		t.Fatalf("selection checked=%v callbacks=%d returns=%d dismiss=%d open=%v", checked, callbacks, returns, dismissals, s.Open)
	}
}

func TestMenubarChildSwitchCommitsCurrentSiblingBeforeParentFrame(t *testing.T) {
	b, s := &fakeBackend{}, &MenubarState{}
	chosen := ""
	p := MenubarProps{Items: testMenubarItems(), OnChoose: func(value string) { chosen = value }}
	parent, _ := menubarTester(t, b, s, &p, false, nil, nil, nil)
	if err := parent.Click("File"); err != nil {
		t.Fatal(err)
	}
	child := menubarChild(t, b)
	child.Key(ui.Ctrl, ui.KeyRight)
	if s.Active != "file" {
		t.Fatal("modified arrow was treated as an unmodified menu switch")
	}
	child.Key(0, ui.KeyRight)
	// Native parent and child windows have independent frame queues. Another
	// child key can arrive before the reanchoring parent has built its frame.
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyEnter)
	if chosen != "paste" || s.Open {
		t.Fatalf("fast sibling key sequence acted on stale menu: chosen=%q active=%q", chosen, s.Active)
	}
	parent.Frame()
	if !parent.Focused("Edit") {
		t.Fatal("fast selection restored a stale sibling trigger")
	}
}

func TestMenubarEmptyMenuCanSwitchAndDismiss(t *testing.T) {
	b, s := &fakeBackend{}, &MenubarState{}
	p := MenubarProps{Items: []overlay.MenubarItem{{Key: "empty", Label: "Empty"}, {Key: "edit", Label: "Edit", Items: []overlay.MenuItem{{Key: "copy", Label: "Copy"}}}}}
	parent, _ := menubarTester(t, b, s, &p, false, nil, nil, nil)
	if err := parent.Click("Empty"); err != nil {
		t.Fatal(err)
	}
	child := menubarChild(t, b)
	child.Key(0, ui.KeyRight)
	parent.Frame()
	child.Frame()
	if s.Active != "edit" || !child.HasText("Copy") {
		t.Fatal("empty menu trapped horizontal keyboard navigation")
	}
	child.Key(0, ui.KeyLeft)
	parent.Frame()
	child.Frame()
	child.Key(0, ui.KeyEscape)
	parent.Frame()
	if s.Open || !parent.Focused("Empty") {
		t.Fatal("empty menu could not dismiss with Escape")
	}
}

func TestMenubarTabDismissalCanFocusAdjacentParentControls(t *testing.T) {
	for _, backward := range []bool{false, true} {
		t.Run(fmt.Sprintf("backward=%v", backward), func(t *testing.T) {
			b, s := &fakeBackend{}, &MenubarState{}
			var before, after ui.Handle
			p := MenubarProps{Items: testMenubarItems(), OnTab: func(back bool) {
				if back {
					before.Focus()
				} else {
					after.Focus()
				}
			}}
			parent, _ := menubarTester(t, b, s, &p, false, &before, &after, nil)
			if err := parent.Click("File"); err != nil {
				t.Fatal(err)
			}
			child := menubarChild(t, b)
			mods := ui.Modifiers(0)
			if backward {
				mods = ui.Shift
			}
			child.Key(mods, ui.KeyTab)
			parent.Frame()
			label := "After"
			if backward {
				label = "Before"
			}
			if s.Open || !parent.Focused(label) || b.surfaces[0].closes != 1 {
				t.Fatalf("Tab did not cross back into parent %s", label)
			}
		})
	}
}

func TestMenubarDisabledRemovedTriggersAndHostCleanup(t *testing.T) {
	b, s := &fakeBackend{}, &MenubarState{}
	p := MenubarProps{Items: testMenubarItems()}
	parent, h := menubarTester(t, b, s, &p, false, nil, nil, nil)
	if err := parent.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	_ = menubarChild(t, b)
	p.Items[2].Disabled = true
	parent.Frame()
	if s.Open || h.IsOpen("main-menubar") || !parent.Focused("File") {
		t.Fatal("disabled active trigger retained menu or stranded focus")
	}
	p.Items = p.Items[1:]
	parent.Frame()
	if s.Active != "view" {
		t.Fatalf("removed key did not normalize to enabled survivor: %q", s.Active)
	}
	s.Open = true
	parent.Frame()
	if !h.IsOpen("main-menubar") {
		t.Fatal("controlled Open did not reopen its retained active key")
	}
	h.CloseAll()
	if s.Open {
		t.Fatal("Host cleanup did not clear caller-owned Open")
	}
}

func TestMenubarMissingHostAndDuplicateKeysAreExplicitErrors(t *testing.T) {
	s := MenubarState{}
	var currentErr error
	sawMissing := false
	p := MenubarProps{Items: testMenubarItems()}
	parent := ui.NewTester(func(c *ui.Context) {
		_, currentErr = Menubar(c, nil, "missing", &s, p)
		if currentErr != nil {
			sawMissing = true
		}
	}, 420, 180)
	if currentErr != nil {
		t.Fatal("closed menubar unnecessarily required a Host")
	}
	if err := parent.Click("File"); err != nil {
		t.Fatal(err)
	}
	if !sawMissing || s.Open || parent.HasText("New") {
		t.Fatal("opening without Host did not report error without an in-window fallback")
	}
	p.Items[1].Key = "file"
	parent.Frame()
	if currentErr == nil {
		t.Fatal("ambiguous top-level keys were accepted")
	}
}

func TestMenubarOutsideDismissLateCallbackAndDisappearingOwner(t *testing.T) {
	b, s := &fakeBackend{}, &MenubarState{}
	h := NewHostWithBackend(b)
	visible, dismissals := true, 0
	p := MenubarProps{Items: testMenubarItems(), OnDismiss: func() { dismissals++ }}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		if visible {
			if _, err := Menubar(c, h, "owned-bar", s, p); err != nil {
				t.Fatal(err)
			}
		}
	}, 450, 170)
	if err := parent.Click("File"); err != nil {
		t.Fatal(err)
	}
	first := b.surfaces[0]
	first.dismiss()
	parent.Frame()
	if s.Open || dismissals != 1 || !parent.Focused("File") {
		t.Fatal("native outside dismissal did not close and restore once")
	}
	if err := parent.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	if len(b.surfaces) != 2 || !s.Open || s.Active != "edit" {
		t.Fatal("menubar did not reopen another sibling")
	}
	first.dismiss()
	parent.Frame()
	if !s.Open || dismissals != 1 || !h.IsOpen("owned-bar") {
		t.Fatal("late native close callback dismissed the replacement menu")
	}
	visible = false
	parent.Frame()
	if s.Open || h.IsPresent("owned-bar") || b.surfaces[1].closes != 1 || dismissals != 1 {
		t.Fatal("disappearing owner left an open menubar surface or duplicate callback")
	}
}

func TestNativeMenubarUsesSystemMenusAndRestoresCallerPolicy(t *testing.T) {
	items := testMenubarItems()
	checked := false
	items[2].Items = []overlay.MenuItem{{Key: "grid", Label: "Grid", Checked: &checked}, {Key: "blocked", Label: "Blocked", Disabled: true}}
	chosen, returns := "", 0
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		c.SetMenuPresentation(ui.PopoverMenus)
		value, err := NativeMenubar(c, items)
		if err != nil {
			t.Fatal(err)
		}
		if c.MenuPresentation() != ui.PopoverMenus {
			t.Fatal("NativeMenubar leaked the system-menu policy")
		}
		if value != "" {
			chosen, returns = value, returns+1
		}
	}, 450, 160)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyRight)
	if !parent.Focused("Edit") {
		t.Fatal("native trigger group did not skip disabled sibling")
	}
	parent.Key(0, ui.KeyDown)
	if !slices.Equal(parent.Menu(), []string{"Grid", "Blocked"}) || parent.HasText("Grid") {
		t.Fatalf("NativeMenubar did not use system menus: %v", parent.Menu())
	}
	if err := parent.ChooseMenuItem("Blocked"); err == nil {
		t.Fatal("disabled native item was accepted")
	}
	if err := parent.ChooseMenuItem("Grid"); err != nil {
		t.Fatal(err)
	}
	parent.Frame()
	if !checked || chosen != "grid" || returns != 1 {
		t.Fatalf("native check/return checked=%v chosen=%q returns=%d", checked, chosen, returns)
	}
}
