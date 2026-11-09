package navigation

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestTabsKeyboardSkipsDisabled(t *testing.T) {
	selected := 0
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(24).Children(func() {
			Tabs(c, &selected, TabsProps{Items: []TabItem{{Label: "Account"}, {Label: "Unavailable", Disabled: true}, {Label: "Password"}}}, func(i int) { ui.Text(c, []string{"Account panel", "Disabled panel", "Password panel"}[i]) })
		})
	}, 480, 320)
	if err := tt.Click("Account"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyRight)
	if selected != 2 || !tt.HasText("Password panel") {
		t.Fatalf("selected=%d, texts=%q", selected, tt.Texts())
	}
	info, ok := tt.AccessByRole(ui.RoleTab, "Password")
	if !ok || !info.Selected {
		t.Fatalf("selected tab accessibility %+v, %v", info, ok)
	}
	tt.Key(0, ui.KeyHome)
	if selected != 0 {
		t.Fatal("Home did not select the first tab")
	}
}

func TestPaginationBoundsAndCurrentPage(t *testing.T) {
	page := 1
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(24).Children(func() { Pagination(c, &page, PaginationProps{Pages: 8}) })
	}, 800, 200)
	previous, ok := tt.AccessByRole(ui.RoleButton, "‹ Previous")
	if !ok || !previous.Disabled {
		t.Fatal("previous was not disabled on first page")
	}
	if err := tt.Click("Next ›"); err != nil {
		t.Fatal(err)
	}
	if page != 2 {
		t.Fatalf("page=%d", page)
	}
	current, ok := tt.AccessByRole(ui.RoleButton, "Page 2")
	if !ok || current.Current != ui.CurrentPage {
		t.Fatalf("current page=%+v found=%v", current, ok)
	}
	if err := tt.Click("8"); err != nil {
		t.Fatal(err)
	}
	if page != 8 {
		t.Fatalf("page=%d", page)
	}
	next, _ := tt.AccessByRole(ui.RoleButton, "Next ›")
	if !next.Disabled {
		t.Fatal("next was not disabled on last page")
	}
}

func TestBreadcrumbAndSidebarExposeCurrentRoute(t *testing.T) {
	selected := "home"
	crumb := ""
	collapsed := false
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Row(c).Fill().Children(func() {
			Sidebar(c, &selected, SidebarProps{Label: "Workspace", Collapsed: &collapsed, Items: []NavItem{{Key: "home", Label: "Home"}, {Key: "settings", Label: "Settings"}}}, nil)
			ui.Column(c).Padding(24).Children(func() {
				if value := Breadcrumb(c, []BreadcrumbItem{{Key: "home", Label: "All workspaces"}, {Key: "settings", Label: "Preferences"}}); value != "" {
					crumb = value
				}
			})
		})
	}, 800, 400)
	if err := tt.Click("Settings"); err != nil {
		t.Fatal(err)
	}
	if selected != "settings" {
		t.Fatal("sidebar did not change route")
	}
	current, _ := tt.AccessByRole(ui.RoleButton, "Settings")
	if current.Current != ui.CurrentPage {
		t.Fatal("sidebar did not expose current route")
	}
	if err := tt.Click("All workspaces"); err != nil {
		t.Fatal(err)
	}
	if crumb != "home" {
		t.Fatalf("breadcrumb chose %q", crumb)
	}
	if err := tt.Click("Collapse sidebar"); err != nil {
		t.Fatal(err)
	}
	if !collapsed {
		t.Fatal("collapse toggle failed")
	}
}
