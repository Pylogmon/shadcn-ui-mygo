package desktop

import (
	"fmt"
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
}

func TestMenuContentScopedSubmenusChecksAndFocus(t *testing.T) {
	left, right := false, false
	state := MenuState{}
	items := []overlay.MenuItem{{Key: "left", Label: "Left", Children: []overlay.MenuItem{{Key: "pin", Label: "Pin left", Checked: &left}}}, {Key: "right", Label: "Right", Children: []overlay.MenuItem{{Key: "pin", Label: "Pin right", Checked: &right}}}}
	closed := 0
	tt := contentTester(&state, items, MenuContentProps{OnClose: func() { closed++ }}, false)
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
}

func TestMenuContentRTLSubmenuArrows(t *testing.T) {
	state := MenuState{}
	tt := contentTester(&state, []overlay.MenuItem{{Key: "more", Label: "More", Children: []overlay.MenuItem{{Key: "one", Label: "One"}}}}, MenuContentProps{}, true)
	tt.Key(0, ui.KeyLeft)
	if state.ActiveDepth != 1 || len(state.Path) != 1 {
		t.Fatalf("RTL Left did not open: %+v", state)
	}
	// The initial 8-DIP slide can cross the final 4-DIP gap temporarily.
	tt.Advance(110 * time.Millisecond)
	root, _ := tt.AccessByRole(ui.RoleMenu, "Menu")
	child, _ := tt.AccessByRole(ui.RoleMenu, "More")
	if child.Bounds.X+child.Bounds.W > root.Bounds.X {
		t.Fatalf("RTL submenu not on leading side: root=%v child=%v", root.Bounds, child.Bounds)
	}
	tt.Key(0, ui.KeyRight)
	if state.ActiveDepth != 0 || len(state.Path) != 0 {
		t.Fatal("RTL Right did not close submenu")
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
