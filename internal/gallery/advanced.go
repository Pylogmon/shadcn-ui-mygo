package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func (s *State) sliderRangesPreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "Ranges & levels", "Drag a thumb or focus it with Tab and use the arrow keys.")
		ui.Row(c).Gap(24).FillWidth().Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Gap(16).Children(func() {
				display.Typography(c, fmt.Sprintf("Price range: $%.0f–$%.0f", s.SliderRange[0], s.SliderRange[1]), display.Small)
				control.RangeSlider(c, &s.SliderRange[0], &s.SliderRange[1], control.SliderProps{Label: "Price", Step: 5, MinGap: 10,
					FormatValue: func(_ int, value float64) string { return fmt.Sprintf("$%.0f", value) }})
				display.Typography(c, "Equalizer bands", display.Small)
				control.SliderValues(c, &s.SliderBands, control.SliderProps{Label: "Equalizer", Step: 5, MinGap: 5,
					ThumbIDs: []string{"bass", "mid", "treble"}, ThumbLabels: []string{"Bass", "Midrange", "Treble"}})
			})
			control.Slider(c, &s.VerticalLevel, control.SliderProps{Label: "Output level", Vertical: true, Step: 5}).Height(120)
		})
	})
}

func (s *State) workspacePreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "Resizable workspace", "Drag a divider, use its arrow keys, or press Enter to collapse the file pane.")
		t := theme.TokensOf(c)
		display.ResizablePanelGroup(c, &s.WorkspaceLayout, display.ResizablePanelGroupProps{Label: "Workspace layout", WithHandle: true, Panels: []display.ResizablePanelProps{
			{ID: "files", Label: "Files", DefaultSize: 24, MinSize: 15, MaxSize: 40, Collapsible: true, Content: func(child *ui.Context) {
				ui.Column(child).Fill().Padding(16).Gap(8).Background(t.Muted).Children(func() {
					display.Typography(child, "Explorer", display.Small)
					ui.Text(child, "main.go")
					ui.Text(child, "theme.go")
				})
			}},
			{ID: "editor", Label: "Editor", DefaultSize: 52, MinSize: 30, Content: func(child *ui.Context) {
				display.ResizablePanelGroup(child, &s.EditorLayout, display.ResizablePanelGroupProps{Label: "Editor panes", Vertical: true, WithHandle: true, Panels: []display.ResizablePanelProps{
					{ID: "source", Label: "Source", DefaultSize: 65, MinSize: 25, Content: func(editor *ui.Context) {
						ui.Column(editor).Fill().Padding(16).Gap(8).Children(func() {
							display.Typography(editor, "Source", display.Small)
							ui.Text(editor, "func main() {\n    app.Run()\n}").Font("monospace").FontSize(12)
						})
					}},
					{ID: "terminal", Label: "Terminal", DefaultSize: 35, MinSize: 20, Content: func(editor *ui.Context) {
						ui.Column(editor).Fill().Padding(16).Background(t.Muted).Children(func() { ui.Text(editor, "✓ Build complete").FontSize(12) })
					}},
				}})
			}},
			{ID: "inspector", Label: "Inspector", DefaultSize: 24, MinSize: 15, MaxSize: 40, Content: func(child *ui.Context) {
				ui.Column(child).Fill().Padding(16).Gap(8).Children(func() {
					display.Typography(child, "Inspector", display.Small)
					ui.Text(child, "Native Go\nTheme aware").FontSize(12)
				})
			}},
		}}).Height(280).Border(1, t.Border).Radius(t.RadiusMD()).Clip()
	})
}

var previewRoutes = []navigation.NavItem{
	{Key: "docs", Label: "Documentation", Children: []navigation.NavItem{
		{Key: "start", Label: "Getting started", Description: "Create your first native application."},
		{Key: "themes", Label: "Themes", Description: "Colors and appearance that work together."},
		{Key: "motion", Label: "Motion", Description: "Responsive transitions and reduced motion."},
	}},
	{Key: "components", Label: "Components", Children: []navigation.NavItem{
		{Key: "buttons", Label: "Buttons"}, {Key: "inputs", Label: "Inputs"}, {Key: "panels", Label: "Panels"},
	}},
	{Key: "updates", Label: "Updates"},
}

func (s *State) routePanelsPreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "Explore the kit", "Hover for a preview, or open a section with the keyboard.")
		var err error
		if s.Config.NativeMenus {
			navigation.NativeNavigationMenu(c, &s.navRoute, previewRoutes)
		} else {
			_, err = navigation.NavigationMenuWith(c, s.Popovers, "preview-route-panels", &s.navRoute, navigation.NavigationMenuProps{
				Label: "Explore navigation", Items: []navigation.NavigationMenuItem{
					{Key: "docs", Label: "Documentation", Width: 420, Links: previewRoutes[0].Children, Columns: 1},
					{Key: "components", Label: "Components", Width: 520, Content: func(child *ui.Context, choose func(string)) {
						ui.Row(child).Gap(16).FillWidth().Children(func() {
							display.Avatar(child, display.AvatarProps{Fallback: "Go", Size: 48})
							ui.Column(child).Grow(1).Gap(10).Children(func() {
								display.CardHeader(child, "Composable by design", "Build native experiences from small, focused components.")
								if control.Button(child, "Browse components", control.ButtonProps{Variant: control.Outline}).Clicked() {
									choose("buttons")
								}
							})
						})
					}},
					{Key: "updates", Label: "Updates"},
				},
			})
		}
		if err != nil {
			s.popupError(c, err)
		}
		if s.navRoute != "" {
			display.Typography(c, "Selected: "+s.navRoute, display.Muted)
		}
	})
}

func (s *State) menubarPreview(c *ui.Context) {
	menus := []overlay.MenubarItem{
		{Key: "file", Label: "File", Items: []overlay.MenuItem{{Key: "new", Label: "New project"}, {Key: "open", Label: "Open project"}, {Separator: true}, {Key: "close", Label: "Close project"}}},
		{Key: "edit", Label: "Edit", Items: []overlay.MenuItem{{Key: "undo", Label: "Undo"}, {Key: "redo", Label: "Redo"}}},
	}
	var chosen string
	var err error
	if s.Config.NativeMenus {
		chosen, err = desktop.NativeMenubar(c, menus)
	} else {
		chosen, err = desktop.Menubar(c, s.Popovers, "preview-menubar", &s.Menubar, desktop.MenubarProps{Items: menus})
	}
	if err != nil {
		s.popupError(c, err)
	}
	if chosen != "" {
		s.pendingMenuAction = chosen
	}
}
