package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorTabsCompositionSearchChoiceAndPageReturn(t *testing.T) {
	s := New()
	s.Config.NativeMenus, s.Config.ReduceMotion = true, true
	r := ui.NewTester(s.View, 500, 900)
	r.SetDirection(ui.RightToLeft)
	navigate := func(query, page string) {
		t.Helper()
		r.Key(ui.Cmd, ui.KeyK)
		r.Frame()
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(query)
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if s.Page != page || s.Command {
			t.Fatalf("search %q reached %q", query, s.Page)
		}
	}
	tab := func(name string) ui.AccessInfo {
		t.Helper()
		card, hasCard := r.AccessByRole(ui.RoleGroup, "Tabs compositions")
		v, _ := r.AccessByRole(ui.RoleRegion, "Navigation")
		for _, a := range r.AccessAll() {
			if a.Role != ui.RoleTab || a.Name != name || !hasCard || a.Bounds.Y < card.Bounds.Y || a.Bounds.Y+a.Bounds.H > card.Bounds.Y+card.Bounds.H {
				continue
			}
			if a.Bounds.Y < v.Bounds.Y || a.Bounds.Y+a.Bounds.H > v.Bounds.Y+v.Bounds.H {
				t.Fatalf("Tabs search did not reveal the actual source %s tab", name)
			}
			return a
		}
		t.Fatalf("missing source %s tab inside its actual card", name)
		return ui.AccessInfo{}
	}
	oldTab, oldLineTab := s.Tab, s.lineTab
	navigate("Tabs", "navigation")
	if !tab("Overview").Selected || s.tabsCompositions.drafts["demo"].value != "overview" {
		t.Fatal("the source dashboard tab default was lost")
	}
	a := tab("Reports")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	if !tab("Reports").Selected || s.tabsCompositions.drafts["demo"].value != "reports" || s.tabsCompositions.drafts["demo"].changes != 1 {
		t.Fatal("the real Reports tab failed one source selection")
	}
	r.Key(0, ui.KeyRight)
	r.Frame()
	if !tab("Analytics").Selected || s.tabsCompositions.drafts["demo"].value != "analytics" || s.tabsCompositions.drafts["demo"].changes != 2 {
		t.Fatal("the source tabs failed native RTL keyboard navigation")
	}
	navigate("Badge", "content")
	navigate("Tabs", "navigation")
	if !tab("Analytics").Selected || s.tabsCompositions.recipe != "demo" || s.tabsCompositions.drafts["demo"].changes != 2 || s.Tab != oldTab || s.lineTab != oldLineTab {
		t.Fatal("Tabs page return lost independent source state or changed the original tabs")
	}
}
