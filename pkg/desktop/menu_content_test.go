package desktop

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func contentTester(state *MenuState, items []overlay.MenuItem, p MenuContentProps, rtl bool) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		container := ui.Column(c).Padding(12)
		if rtl {
			container.Direction(ui.RightToLeft)
		}
		container.Children(func() { MenuContent(c, state, items, p) })
	}, 700, 600)
}
func TestMenuContentKeyboardTypeaheadAndDisabledSkipping(t *testing.T) {
	state := MenuState{}
	items := []overlay.MenuItem{{Key: "open", Label: "Open"}, {Key: "separator", Separator: true}, {Key: "blocked", Label: "Unavailable", Disabled: true}, {Key: "save", Label: "Save"}, {Key: "settings", Label: "Settings"}}
	chosen := []string{}
	closed := 0
	tt := contentTester(&state, items, MenuContentProps{OnChoose: func(key string) { chosen = append(chosen, key) }, OnClose: func() { closed++ }}, false)
	tt.Key(0, ui.KeyDown)
	if state.Highlighted[0] != "save" {
		t.Fatalf("disabled/separator skip=%v", state.Highlighted)
	}
	tt.Key(0, ui.KeyEnd)
	if state.Highlighted[0] != "settings" {
		t.Fatal(state.Highlighted)
	}
	tt.Key(0, ui.KeyDown)
	if state.Highlighted[0] != "open" {
		t.Fatal("Down did not wrap")
	}
	tt.Type("sa")
	if state.Highlighted[0] != "save" {
		t.Fatalf("typeahead=%v", state.Highlighted)
	}
	tt.Advance(time.Second)
	tt.Type("s")
	if state.Highlighted[0] != "settings" {
		t.Fatalf("new prefix should cycle after timeout: %v", state.Highlighted)
	}
	tt.Type("s")
	if state.Highlighted[0] != "save" {
		t.Fatalf("repeated character did not cycle: %v", state.Highlighted)
	}
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(chosen, []string{"save"}) || closed != 1 || state.Chosen != "save" {
		t.Fatalf("selection=%v closed=%d state=%+v", chosen, closed, state)
	}
	item, ok := tt.AccessByRole(ui.RoleMenuItem, "Save")
	if !ok || !item.Focused {
		t.Fatalf("highlight not exposed to native accessibility: %+v", item)
	}
	// A solid decorative child isolates inherited disabled paint from glyph
	// antialiasing. Disabled rows still reject pointer activation and skip keys.
	for _, disabled := range []bool{false, true} {
		state := MenuState{}
		selected := 0
		paint := ui.NewTesterWith(func(c *ui.Context) {
			tokens := theme.Preset("blue", false)
			tokens.Popover, tokens.Border, tokens.Background = ui.RGB(0, 0, 0), ui.RGB(0, 0, 0), ui.RGB(0, 0, 0)
			theme.Scope(c, tokens).Fill().Background(tokens.Background).Padding(20).Children(func() {
				MenuContent(c, &state, []overlay.MenuItem{{Key: "row", Label: "Paint row", Disabled: disabled}}, MenuContentProps{ItemContent: func(c *ui.Context, _ overlay.MenuItem) {
					ui.Box(c).Size(16, 16).Background(ui.RGB(255, 255, 255)).AccessHidden(true)
				}, OnChoose: func(string) { selected++ }})
			})
		}, 240, 120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		row, _ := paint.AccessByRole(ui.RoleMenuItem, "Paint row")
		peak := uint8(0)
		image := paint.Image()
		for y := int(row.Bounds.Y) + 3; y < int(row.Bounds.Y+row.Bounds.H)-3; y++ {
			for x := int(row.Bounds.X) + 3; x < int(row.Bounds.X+row.Bounds.W)-3; x++ {
				peak = max(peak, color.NRGBAModel.Convert(image.At(x, y)).(color.NRGBA).R)
			}
		}
		want := uint8(255)
		if disabled {
			want = 128
		}
		if peak != want || row.Disabled != disabled {
			t.Fatalf("disabled=%v decorative white=%d want%d", disabled, peak, want)
		}
		paint.ClickAt(row.Bounds.X+12, row.Bounds.Y+16)
		paint.Key(0, ui.KeyEnter)
		if disabled && selected != 0 {
			t.Fatal("disabled row accepted native pointer or keyboard input")
		}
	}
}

func TestMenuContentScopedSubmenusChecksAndFocus(t *testing.T) {
	left, right := false, false
	state := MenuState{}
	items := []overlay.MenuItem{{Key: "left", Label: "Left", Children: []overlay.MenuItem{{Key: "pin", Label: "Pin left", Checked: &left}}}, {Key: "right", Label: "Right", Children: []overlay.MenuItem{{Key: "pin", Label: "Pin right", Checked: &right}}}}
	closed := 0
	var measured ui.Rect
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(12).AlignItems(ui.Start).Children(func() {
			measured = MenuContent(c, &state, items, MenuContentProps{OnClose: func() { closed++ }}).LayoutBounds()
		})
	}, 700, 600)
	tt.Key(0, ui.KeyRight)
	if state.ActiveDepth != 1 || !slices.Equal(state.Path, []string{"left"}) {
		t.Fatalf("submenu not opened: %+v", state)
	}
	// Inspect the settled adjacent layout after the source 100ms entrance.
	tt.Advance(110 * time.Millisecond)
	root, _ := tt.AccessByRole(ui.RoleMenu, "Menu")
	child, _ := tt.AccessByRole(ui.RoleMenu, "Left")
	if child.Bounds.X < root.Bounds.X+root.Bounds.W {
		t.Fatalf("submenu is not measured adjacent content: root=%v child=%v", root.Bounds, child.Bounds)
	}
	tt.Key(0, ui.KeyEscape)
	if state.ActiveDepth != 0 || len(state.Path) != 0 || closed != 0 || !tt.Focused("Menu") {
		t.Fatalf("submenu close/focus=%+v", state)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEnter)
	if left || !right || closed != 0 || !slices.Equal(state.ChosenPath, []string{"right", "pin"}) {
		t.Fatalf("duplicate leaf crossed branches: left=%v right=%v state=%+v", left, right, state)
	}
	node, ok := tt.AccessByRole(ui.RoleMenuItemCheckBox, "Pin right")
	if !ok || !node.Checked {
		t.Fatalf("checked menu state missing: %+v", node)
	}
	// Widths belong to each owner, including grandchildren. Unset descendants
	// inherit the shared192 fallback rather than their parent's explicit176.
	items[0].SubmenuWidth = 240
	items[1].SubmenuWidth = 176
	items[1].Children = append(items[1].Children, overlay.MenuItem{Key: "nested", Label: "Nested", Children: []overlay.MenuItem{{Key: "leaf", Label: "Leaf"}}})
	state.Path, state.Highlighted, state.ActiveDepth = []string{"left"}, []string{"left", "pin"}, 1
	tt.Frame()
	tt.Advance(110 * time.Millisecond)
	state.Path, state.Highlighted, state.ActiveDepth = []string{"right", "nested"}, []string{"right", "nested", "leaf"}, 2
	tt.Frame()
	// The old240-DIP sibling is still leaving. Advance into its overlap with
	// the narrower live path, then settle and verify exact native panel geometry.
	tt.Advance(50 * time.Millisecond)
	tt.Frame()
	if len(state.Path) != 2 || !tt.HasText("Leaf") || measured.W != 192+4+240+4+192 {
		t.Fatalf("retiring sibling did not retain its max envelope: bounds=%v path=%v", measured, state.Path)
	}
	items[0].SubmenuWidth = 128
	tt.Frame()
	tt.Frame()
	if measured.W != 192+4+240+4+192 {
		t.Fatalf("retiring owner resize collapsed its in-flight envelope: %v", measured)
	}
	tt.Advance(60 * time.Millisecond)
	root, _ = tt.AccessByRole(ui.RoleMenu, "Menu")
	rightPanel, _ := tt.AccessByRole(ui.RoleMenu, "Right")
	nested, _ := tt.AccessByRole(ui.RoleMenu, "Nested")
	if root.Bounds.W != 192 || rightPanel.Bounds.W != 176 || nested.Bounds.W != 192 || rightPanel.Bounds.X-root.Bounds.X-root.Bounds.W != 4 || nested.Bounds.X-rightPanel.Bounds.X-rightPanel.Bounds.W != 4 {
		t.Fatalf("per-owner width/shared fallback/gap: root=%v right=%v nested=%v", root.Bounds, rightPanel.Bounds, nested.Bounds)
	}
	items[1].Children[1].SubmenuWidth = 144
	tt.Frame()
	nested, _ = tt.AccessByRole(ui.RoleMenu, "Nested")
	if nested.Bounds.W != 144 {
		t.Fatalf("current owner width did not refresh: %v", nested.Bounds)
	}
	for _, invalid := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("direct MenuContent accepted a non-finite nested width")
				}
			}()
			items[1].Children[1].SubmenuWidth = invalid
			_ = contentTester(&MenuState{}, items, MenuContentProps{}, false)
		}()
	}
}

func TestMenuContentRTLSubmenuArrows(t *testing.T) {
	state := MenuState{}
	items := []overlay.MenuItem{{Key: "more", Label: "More", SubmenuWidth: 176, Children: []overlay.MenuItem{{Key: "one", Label: "One", SubmenuWidth: 144, Children: []overlay.MenuItem{{Key: "leaf", Label: "Leaf"}}}}}, {Key: "wide", Label: "Wide", SubmenuWidth: 240, Children: []overlay.MenuItem{{Key: "old", Label: "Old"}}}}
	markers := map[string]ui.Rect{}
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(12).Direction(ui.RightToLeft).Children(func() {
			MenuContent(c, &state, items, MenuContentProps{ItemContent: func(c *ui.Context, item overlay.MenuItem) {
				markers[item.Label] = ui.Box(c).FillWidth().Height(16).LayoutBounds()
			}})
		})
	}, 700, 600)
	tt.Key(0, ui.KeyLeft)
	if state.ActiveDepth != 1 || len(state.Path) != 1 {
		t.Fatalf("RTL Left did not open: %+v", state)
	}
	// The initial 8-DIP slide can cross the final 4-DIP gap temporarily.
	tt.Advance(110 * time.Millisecond)
	root, _ := tt.AccessByRole(ui.RoleMenu, "Menu")
	child, _ := tt.AccessByRole(ui.RoleMenu, "More")
	if child.Bounds.W != 176 || root.Bounds.X-child.Bounds.X-child.Bounds.W != 4 {
		t.Fatalf("RTL submenu not on leading side: root=%v child=%v", root.Bounds, child.Bounds)
	}
	tt.Key(0, ui.KeyLeft)
	tt.Advance(110 * time.Millisecond)
	nested, _ := tt.AccessByRole(ui.RoleMenu, "One")
	if nested.Bounds.W != 144 || child.Bounds.X-nested.Bounds.X-nested.Bounds.W != 4 {
		t.Fatalf("RTL grandchild lost actual owner gap: parent=%v nested=%v", child.Bounds, nested.Bounds)
	}
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyRight)
	if state.ActiveDepth != 0 || len(state.Path) != 0 {
		t.Fatal("RTL Right did not close submenu")
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyLeft)
	tt.Advance(110 * time.Millisecond)
	if err := tt.Click("More"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyLeft)
	tt.Advance(50 * time.Millisecond)
	tt.Frame()
	// Scalar LayoutBounds omit entry transforms. A retiring240-DIP column must
	// not leave64 extra DIPs between the current176 parent and144 grandchild.
	one, leaf := markers["One"], markers["Leaf"]
	if math.Abs(float64((one.X+one.W)-(leaf.X+leaf.W)-(176+4))) > .1 {
		t.Fatalf("RTL retained wide branch left a gap before the live grandchild: one=%v leaf=%v", one, leaf)
	}
}

func TestMenuContentHoverAndRadioSemantics(t *testing.T) {
	state := MenuState{}
	items := []overlay.MenuItem{{Key: "themes", Label: "Themes", Children: []overlay.MenuItem{{Key: "light", Label: "Light", Radio: true, Selected: true}, {Key: "dark", Label: "Dark", Radio: true}}}, {Key: "blocked", Label: "Unavailable", Disabled: true}}
	tt := contentTester(&state, items, MenuContentProps{KeepOpenOnSelect: true}, false)
	row, _ := tt.AccessByRole(ui.RoleMenuItem, "Themes")
	tt.Move(row.Bounds.X+8, row.Bounds.Y+16)
	if !slices.Equal(state.Path, []string{"themes"}) {
		t.Fatalf("hover did not open submenu: %+v", state)
	}
	if err := tt.Click("Dark"); err != nil {
		t.Fatal(err)
	}
	dark, _ := tt.AccessByRole(ui.RoleMenuItemRadio, "Dark")
	light, _ := tt.AccessByRole(ui.RoleMenuItemRadio, "Light")
	if !dark.Checked || light.Checked || items[0].Children[0].Selected || !items[0].Children[1].Selected {
		t.Fatalf("radio state light=%+v dark=%+v", light, dark)
	}
}

func TestMenuContentKeyboardRevealsLongListAndHonorsReducedMotion(t *testing.T) {
	state := MenuState{}
	items := make([]overlay.MenuItem, 40)
	for i := range items {
		items[i] = overlay.MenuItem{Key: fmt.Sprint(i), Label: fmt.Sprintf("Option %02d", i)}
	}
	var reduced bool
	tt := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		reduced = motion.Reduced(c)
		MenuContent(c, &state, items, MenuContentProps{MaxHeight: 120})
	}, 260, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	tt.Key(0, ui.KeyEnd)
	item, _ := tt.AccessByRole(ui.RoleMenuItem, "Option 39")
	panel, _ := tt.AccessByRole(ui.RoleMenu, "Menu")
	if state.Highlighted[0] != "39" || item.Offscreen || item.Bounds.H <= 0 || item.Bounds.Y+item.Bounds.H > panel.Bounds.Y+panel.Bounds.H+1 {
		t.Fatalf("End did not reveal last row: item=%+v panel=%v state=%+v", item, panel.Bounds, state)
	}
	if !reduced || tt.FramePending() {
		t.Fatal("reduced motion menu scheduled idle animation")
	}
}
