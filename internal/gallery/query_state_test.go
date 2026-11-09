package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func queryStateTester(s *State) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		c.DisableAnimations()
		s.View(c)
	}, 1280, 2400)
}

func websiteQueryField(t *testing.T, r *ui.Tester) ui.AccessInfo {
	t.Helper()
	if field, ok := r.AccessByRole(ui.RoleTextField, "Website name"); ok {
		return field
	}
	// The pre-fix website input had no name. Keep the baseline runnable so
	// the regression fails on real overwritten text, not a missing new label.
	for _, field := range r.AccessAll() {
		if field.Role == ui.RoleTextField && field.Name == "" {
			return field
		}
	}
	t.Fatal("website editor missing")
	return ui.AccessInfo{}
}

func commandPreviewField(t *testing.T, r *ui.Tester) ui.AccessInfo {
	t.Helper()
	if field, ok := r.AccessByRole(ui.RoleComboBox, "Example command search"); ok {
		return field
	}
	if field, ok := r.AccessByRole(ui.RoleComboBox, "Search commands"); ok {
		return field
	}
	t.Fatal("inline command editor missing")
	return ui.AccessInfo{}
}

func queryFieldClick(r *ui.Tester, field ui.AccessInfo) {
	r.ClickAt(field.Bounds.X+8, field.Bounds.Y+field.Bounds.H/2)
}

func TestGalleryWebsiteDraftIsIndependentOfComponentSearch(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := queryStateTester(s)
	queryFieldClick(r, websiteQueryField(t, r))
	r.Type("native-project")
	if got := websiteQueryField(t, r).Value; got != "native-project" {
		t.Fatalf("website did not accept native typing: %q", got)
	}
	if err := r.Click("Search components"); err != nil {
		t.Fatal(err)
	}
	if got := advancedAccess(t, r, ui.RoleComboBox, "Search commands").Value; got != "" {
		t.Fatalf("website typing changed the global search draft: %q", got)
	}
	r.Type("Combobox")
	r.Key(0, ui.KeyEnter)
	if s.Page != "controls" || s.Command {
		t.Fatal("component search did not return to the actual Forms page")
	}
	if got := websiteQueryField(t, r).Value; got != "native-project" {
		t.Fatalf("search/navigation overwrote the website draft: %q", got)
	}
	queryFieldClick(r, websiteQueryField(t, r))
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("renamed-project")
	if s.Query != "Combobox" {
		t.Fatalf("website edit rewrote the global search query: %q", s.Query)
	}
	s.Page = "overview"
	r.Frame()
	s.Page = "controls"
	r.Frame()
	if got := websiteQueryField(t, r).Value; got != "renamed-project" {
		t.Fatalf("leaving Forms lost the independent website draft: %q", got)
	}
}

func TestGalleryInlineCommandHasAnIndependentSearchDraft(t *testing.T) {
	s := New()
	s.Page = "overlays"
	r := queryStateTester(s)
	queryFieldClick(r, commandPreviewField(t, r))
	r.Type("settings")
	if got := commandPreviewField(t, r).Value; got != "settings" {
		t.Fatalf("inline command did not accept native input: %q", got)
	}
	if err := r.Click("Search components"); err != nil {
		t.Fatal(err)
	}
	if got := advancedAccess(t, r, ui.RoleComboBox, "Search commands").Value; got != "" {
		t.Fatalf("inline example changed the global search draft: %q", got)
	}
	r.Type("Combobox")
	r.Key(0, ui.KeyEnter)
	if s.Page != "controls" || s.Command {
		t.Fatal("global search did not navigate independently of the inline example")
	}
	s.Page = "overlays"
	r.Frame()
	if got := commandPreviewField(t, r).Value; got != "settings" {
		t.Fatalf("global search replaced the inline example query: %q", got)
	}
	queryFieldClick(r, commandPreviewField(t, r))
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("calendar")
	if s.Query != "Combobox" {
		t.Fatalf("inline command rewrote global search: %q", s.Query)
	}
}
