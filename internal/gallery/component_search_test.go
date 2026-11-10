package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// One actual narrow RTL search/edit/scroll workflow exercises the post-layout
// reveal, including same-page searches and independently built sibling cards.
func TestCreatorComponentSearchRevealsSourceOnce(t *testing.T) {
	s := New()
	s.Config.NativeMenus, s.Config.ReduceMotion = true, true
	r := ui.NewTester(s.View, 500, 900)
	r.SetDirection(ui.RightToLeft)
	navigate := func(query, page, card, configuration string) ui.AccessInfo {
		t.Helper()
		r.Key(ui.Cmd, ui.KeyK)
		r.Frame()
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(query)
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if s.Page != page || s.Command || s.pendingPreview != "" {
			t.Fatalf("search %q page=%q open=%v pending=%q", query, s.Page, s.Command, s.pendingPreview)
		}
		panel, ok := r.AccessByRole(ui.RoleRegion, pages[s.previewIndex].Label)
		if !ok {
			t.Fatal("missing actual page scroll viewport")
		}
		group, ok := r.AccessByRole(ui.RoleGroup, card)
		if !ok {
			t.Fatalf("missing actual source card %q", card)
		}
		control, ok := r.AccessByRole(ui.RoleMenuButton, configuration)
		if !ok || control.Bounds.Y < panel.Bounds.Y || control.Bounds.Y+control.Bounds.H > panel.Bounds.Y+panel.Bounds.H || control.Bounds.Y < group.Bounds.Y || control.Bounds.X < panel.Bounds.X-1 || control.Bounds.X+control.Bounds.W > panel.Bounds.X+panel.Bounds.W+1 {
			t.Fatalf("search %q did not reveal source selector %q: %v in %v", query, configuration, control.Bounds, panel.Bounds)
		}
		return control
	}
	snapshot := func(name string) {
		t.Helper()
		if dir := os.Getenv("GALLERY_COMPONENT_SEARCH_SNAPSHOT_DIR"); dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			file, err := os.Create(filepath.Join(dir, name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, r.Image())
			closeErr := file.Close()
			if err != nil || closeErr != nil {
				t.Fatalf("snapshot %v / %v", err, closeErr)
			}
		}
	}
	navigate("Input", "controls", "Input compositions", "Input configuration")
	editor, ok := r.AccessByRole(ui.RoleTextField, "API Key")
	if !ok || !editor.Password {
		t.Fatal("search did not retain the real protected source editor")
	}
	r.ClickAt(editor.Bounds.X+editor.Bounds.W/2, editor.Bounds.Y+editor.Bounds.H/2)
	r.Type("sk-search-draft")
	r.Frame()
	if *s.inputCompositions.drafts["demo"].texts["api"] != "sk-search-draft" {
		t.Fatal("revealed source editor did not accept native input")
	}
	snapshot("component-search-input-rtl")
	panel, _ := r.AccessByRole(ui.RoleRegion, "Forms & controls")
	before, _ := r.AccessByRole(ui.RoleGroup, "Input compositions")
	r.Scroll(panel.Bounds.X+panel.Bounds.W-12, panel.Bounds.Y+panel.Bounds.H/2, 0, 300)
	after, _ := r.AccessByRole(ui.RoleGroup, "Input compositions")
	if after.Bounds.Y >= before.Bounds.Y-100 {
		t.Fatal("search reveal prevented manual scrolling")
	}
	for range 3 {
		r.Frame()
	}
	settled, _ := r.AccessByRole(ui.RoleGroup, "Input compositions")
	if settled.Bounds.Y != after.Bounds.Y {
		t.Fatal("the consumed search request snapped the page back on a later frame")
	}
	for _, target := range []struct{ query, page, card, configuration string }{
		{"Card", "overview", "Card compositions", "Card configuration"},
		{"Button", "controls", "Button compositions", "Button configuration"},
		{"Textarea", "controls", "Textarea compositions", "Textarea configuration"},
		{"Switch", "controls", "Switch compositions", "Switch configuration"},
		{"Checkbox", "controls", "Checkbox compositions", "Checkbox configuration"},
		{"RadioGroup", "controls", "Radio Group compositions", "Radio Group configuration"},
		{"Field", "controls", "Field compositions", "Field configuration"},
		{"ButtonGroup", "controls", "Button group examples", "Button group configuration"},
		{"Disclosure", "display", "Disclosure compositions", "Disclosure configuration"},
		{"Badge", "content", "Badge compositions", "Badge configuration"},
		{"Item", "content", "Item compositions", "Item configuration"},
		{"Empty", "content", "Empty compositions", "Empty configuration"},
		{"Alert", "content", "Alert compositions", "Alert configuration"},
		{"HoverCard", "overlays", "Hover card compositions", "Hover card configuration"},
		{"Popover", "overlays", "Popover compositions", "Popover configuration"},
		{"DropdownMenu", "overlays", "Dropdown Menu compositions", "Dropdown Menu configuration"},
		{"ContextMenu", "navigation", "Context menu compositions", "Context menu configuration"},
		{"Menubar", "navigation", "Menubar compositions", "Menubar configuration"},
	} {
		navigate(target.query, target.page, target.card, target.configuration)
	}
	s.Config.Appearance = "dark"
	r.Frame()
	snapshot("component-search-menubar-rtl-dark")
	navigate("Input", "controls", "Input compositions", "Input configuration")
	if *s.inputCompositions.drafts["demo"].texts["api"] != "sk-search-draft" || s.Name != "Sofia Davis" || s.Bio != "" {
		t.Fatal("search changed the independent draft or original Profile data")
	}
	// An unregistered fuzzy query still selects its existing page and consumes
	// no component target; a page choice cannot accidentally reveal another page.
	if componentSearchTarget("content", "Input") != "" || componentSearchTarget("controls", "inp") != "" {
		t.Fatal("page/fuzzy search was turned into an unrelated exact component request")
	}
}
