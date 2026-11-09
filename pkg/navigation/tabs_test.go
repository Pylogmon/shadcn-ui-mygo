package navigation

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func tabView(selected *int, p *TabsProps) func(*ui.Context) {
	return func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(16).Children(func() {
			Tabs(c, selected, *p, func(i int) { ui.Text(c, p.Items[i].Label+" content") })
			ui.Button(c, "After tabs")
		})
	}
}
func TestTabsManualActivationAndPanelRelationships(t *testing.T) {
	selected := 0
	p := TabsProps{Label: "Settings", Manual: true, Items: []TabItem{{Label: "Account"}, {Label: "Unavailable", Disabled: true}, {Label: "Security"}}}
	r := ui.NewTester(tabView(&selected, &p), 520, 250)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyRight)
	if selected != 0 || !r.Focused("Security") || !r.HasText("Account content") {
		t.Fatalf("manual arrow selected=%d focus=%v texts=%v", selected, r.Focused("Security"), r.Texts())
	}
	r.Key(0, ui.KeyEnter)
	if selected != 2 || !r.HasText("Security content") {
		t.Fatal("Enter did not activate manual tab")
	}
	tab, ok := r.AccessByRole(ui.RoleTab, "Security")
	if !ok || !tab.Selected || !slices.Equal(tab.Controls, []string{"Security"}) {
		t.Fatalf("tab relations %+v", tab)
	}
	region, ok := r.AccessByRole(ui.RoleRegion, "Security")
	if !ok || region.RoleDescription != "tab panel" || !region.Focusable {
		t.Fatalf("panel %+v", region)
	}
	r.Key(0, ui.KeyTab)
	if info, _ := r.AccessByRole(ui.RoleRegion, "Security"); !info.Focused {
		t.Fatalf("Tab did not enter active panel %+v", info)
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("After tabs") {
		t.Fatal("Tab did not leave panel")
	}
}
func TestTabsOrientationNoLoopAndRTL(t *testing.T) {
	selected := 0
	p := TabsProps{NoLoop: true, Items: []TabItem{{Label: "First"}, {Label: "Second"}, {Label: "Third"}}}
	r := ui.NewTester(tabView(&selected, &p), 520, 250)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyUp)
	if selected != 0 {
		t.Fatal("horizontal tabs used Up")
	}
	r.Key(0, ui.KeyLeft)
	if selected != 0 || !r.Focused("First") {
		t.Fatal("NoLoop wrapped backward")
	}
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyRight)
	if selected != 2 || !r.Focused("Third") {
		t.Fatal("NoLoop wrapped forward")
	}
	r.SetDirection(ui.RightToLeft)
	r.Key(0, ui.KeyRight)
	if selected != 1 {
		t.Fatalf("RTL right chose %d", selected)
	}
	p.Vertical = true
	r.Frame()
	r.Key(0, ui.KeyLeft)
	if selected != 1 {
		t.Fatal("vertical tabs used Left")
	}
	r.Key(0, ui.KeyDown)
	if selected != 2 {
		t.Fatal("vertical Down did not choose next tab")
	}
	list, _ := r.AccessByRole(ui.RoleTabList, "")
	if !list.Vertical {
		t.Fatalf("vertical list %+v", list)
	}
}
func TestTabsNormalizeDisabledRemovedAndEmpty(t *testing.T) {
	selected := 0
	p := TabsProps{Items: []TabItem{{Key: "a", Label: "First", Disabled: true}, {Key: "b", Label: "Second"}}}
	r := ui.NewTester(tabView(&selected, &p), 520, 250)
	if selected != 1 || !r.HasText("Second content") {
		t.Fatal("disabled selection was not normalized")
	}
	p.Items = p.Items[:1]
	r.Frame()
	if selected != -1 {
		t.Fatalf("all disabled kept selection %d", selected)
	}
	p.Items = nil
	r.Frame()
	if selected != -1 {
		t.Fatal("empty selection not -1")
	}
}
func TestTabsIndicatorAnimationAndReducedMotion(t *testing.T) {
	for _, reduce := range []bool{false, true} {
		t.Run(map[bool]string{false: "animated", true: "reduced"}[reduce], func(t *testing.T) {
			selected := 0
			p := TabsProps{Variant: TabsLine, Items: []TabItem{{Label: "First"}, {Label: "Second"}}}
			r := ui.NewTesterWith(tabView(&selected, &p), 520, 250, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduce}})
			if err := r.Click("Second"); err != nil {
				t.Fatal(err)
			}
			if !reduce && !r.FramePending() {
				t.Fatal("indicator transition did not schedule frames")
			}
			r.Advance(400 * time.Millisecond)
			if r.FramePending() {
				t.Fatal("indicator did not settle")
			}
		})
	}
}
