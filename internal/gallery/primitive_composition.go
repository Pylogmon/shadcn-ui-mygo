package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
)

func (s *State) composedProjectTabs(c *ui.Context) {
	items := []struct{ value, label, content string }{
		{"overview", "Overview", "Your project at a glance."},
		{"activity", "Activity", "Your latest project updates."},
		{"settings", "Settings", "Manage your project preferences."},
	}
	if s.lineTab < 0 || s.lineTab >= len(items) {
		s.lineTab = 0
	}
	// The existing application model remains authoritative. TabsRoot consumes
	// the current value only during this synchronous build and retains no pointer.
	value := items[s.lineTab].value
	navigation.TabsRoot(c, &value, navigation.TabsRootProps{Key: "project-tabs", Label: tr(c, "Project view")}, func(c *ui.Context) {
		ui.Row(c).FillWidth().Wrap().Gap(12).AlignItems(ui.Center).Children(func() {
			navigation.TabsList(c, navigation.TabsListProps{Variant: navigation.TabsLine}, func(c *ui.Context) {
				for _, item := range items {
					navigation.TabsTrigger(c, item.value, navigation.TabsTriggerProps{Label: tr(c, item.label)})
				}
			})
			ui.Box(c).Grow(1).MinWidth(0)
			if control.Button(c, tr(c, "Show current view"), control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
				for _, item := range items {
					if item.value == value {
						s.status = "Current view: " + item.label
					}
				}
			}
		})
		for _, item := range items {
			navigation.TabsContent(c, item.value, navigation.TabsContentProps{}, func(c *ui.Context) {
				ui.Text(c, tr(c, item.content)).Padding(12)
			})
		}
	})
	for index, item := range items {
		if item.value == value {
			s.lineTab = index
			break
		}
	}
}
