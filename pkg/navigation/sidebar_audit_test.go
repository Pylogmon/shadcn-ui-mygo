package navigation

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestSidebarExternalTriggerBeforeLayoutControlsPanel(t *testing.T) {
	state, route := &SidebarState{}, "home"
	mobile := false
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			SidebarTrigger(c, state).Label("External sidebar trigger")
			_, err := SidebarWith(c, nil, "external", &route, state, SidebarLayoutProps{Mobile: &mobile, HideTrigger: true, Label: "External navigation", Items: sidebarTestItems()}, nil)
			if err != nil {
				t.Fatal(err)
			}
		})
	}, 1000, 500)
	trigger := sidebarTestAccess(t, r, ui.RoleButton, "External sidebar trigger")
	panel := sidebarTestAccess(t, r, ui.RoleNavigation, "External navigation")
	if len(trigger.Controls) != 1 || trigger.Controls[0] != panel.Name {
		t.Fatalf("external trigger lost panel association: trigger=%+v panel=%+v", trigger, panel)
	}
}
