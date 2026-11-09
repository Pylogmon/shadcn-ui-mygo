// The fixed Tabs examples, composed with native slots and public kit controls.
package main

import (
	"flag"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	nav "github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type recipe struct {
	ID, Title                                                          string
	Values                                                             []string
	Line, Content, Vertical, Icons, IconOnly, Dropdown, Input, Compare bool
	Disabled                                                           int
}

var recipes = []recipe{
	{ID: "basic", Title: "Basic", Values: []string{"home", "settings"}, Disabled: -1},
	{ID: "line", Title: "Line", Values: []string{"overview", "analytics", "reports"}, Line: true, Disabled: -1},
	{ID: "comparison", Title: "Variants Alignment", Values: []string{"overview", "analytics"}, Compare: true, Disabled: -1},
	{ID: "disabled", Title: "Disabled", Values: []string{"home", "settings"}, Disabled: 1},
	{ID: "icons", Title: "With Icons", Values: []string{"preview", "code"}, Icons: true, Disabled: -1},
	{ID: "icon-only", Title: "Icon Only", Values: []string{"home", "search", "settings"}, Icons: true, IconOnly: true, Disabled: -1},
	{ID: "multiple", Title: "Multiple", Values: []string{"overview", "analytics", "reports", "settings"}, Disabled: -1},
	{ID: "content", Title: "With Content", Values: []string{"account", "password", "notifications"}, Content: true, Disabled: -1},
	{ID: "line-content", Title: "Line With Content", Values: []string{"account", "password", "notifications"}, Line: true, Content: true, Disabled: -1},
	{ID: "line-disabled", Title: "Line Disabled", Values: []string{"overview", "analytics", "reports"}, Line: true, Disabled: 2},
	{ID: "dropdown", Title: "With Dropdown", Values: []string{"overview", "analytics", "reports"}, Content: true, Dropdown: true, Disabled: -1},
	{ID: "vertical", Title: "Vertical", Values: []string{"account", "password", "notifications"}, Vertical: true, Content: true, Disabled: -1},
	{ID: "input", Title: "With Input and Button", Values: []string{"overview", "analytics"}, Content: true, Input: true, Disabled: -1},
}

type appState struct {
	appearance                theme.Controller
	recipe                    string
	values                    map[string]*string
	notes                     map[string]*string
	second, query, status     string
	manual, noLoop, rtl, menu bool
	lastRecipe                string
}

func newState(dark bool, id string) *appState {
	s := &appState{appearance: theme.Controller{Preset: "neutral", Mode: theme.LightMode}, recipe: id, values: map[string]*string{}, notes: map[string]*string{}, second: "overview"}
	if dark {
		s.appearance.Mode = theme.DarkMode
	}
	for _, r := range recipes {
		v := r.Values[0]
		s.values[r.ID] = &v
	}
	return s
}
func label(value string) string {
	if value == "code" {
		return "Code"
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
func tabIcon(value string) *ui.SVG {
	switch value {
	case "home":
		return outlined.Home()
	case "search":
		return outlined.Search()
	case "settings":
		return outlined.Settings()
	case "code":
		return outlined.Code()
	default:
		return outlined.Window()
	}
}
func (s *appState) view(c *ui.Context) {
	s.appearance.Apply(c)
	t := theme.TokensOf(c)
	ui.Column(c).Fill().Padding(24).Gap(16).Background(t.Background).TextColor(t.Foreground).Children(func() {
		display.Typography(c, "Tabs composition", display.H2)
		ui.Text(c, "Choose an example, edit a note, or use the controls beside the tab list.").TextColor(t.MutedForeground)
		ui.Row(c).Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
			options := make([]control.Option, 0, len(recipes))
			for _, r := range recipes {
				options = append(options, control.Option{Value: r.ID, Label: r.Title})
			}
			control.NativeSelect(c, &s.recipe, control.SelectProps{Label: "Tabs example", Options: options}).Width(230)
			control.Switch(c, &s.manual, control.SwitchProps{Label: "Manual activation", Size: control.Small})
			control.Switch(c, &s.noLoop, control.SwitchProps{Label: "Stop at ends", Size: control.Small})
			control.Switch(c, &s.rtl, control.SwitchProps{Label: "Right to left", Size: control.Small})
		})
		selected := recipes[0]
		for _, r := range recipes {
			if r.ID == s.recipe {
				selected = r
				break
			}
		}
		if s.lastRecipe != s.recipe {
			s.menu = false
			s.lastRecipe = s.recipe
		}
		display.Card(c, func() {
			display.Typography(c, selected.Title, display.H3)
			wrapper := ui.Box(c).FillWidth().MinWidth(0)
			if s.rtl {
				wrapper.Direction(ui.RightToLeft)
			}
			wrapper.Children(func() {
				if selected.Compare {
					ui.Row(c).Wrap().Gap(24).Children(func() {
						s.group(c, selected, s.values[selected.ID], nav.TabsDefault, "default")
						s.group(c, selected, &s.second, nav.TabsLine, "line")
					})
				} else {
					variant := nav.TabsDefault
					if selected.Line {
						variant = nav.TabsLine
					}
					s.group(c, selected, s.values[selected.ID], variant, selected.ID)
				}
			})
		}).FillWidth()
		if s.status != "" {
			ui.Text(c, s.status).Role(ui.RoleStatus)
		}
	})
}
func (s *appState) group(c *ui.Context, r recipe, value *string, variant nav.TabsVariant, key string) {
	root := nav.TabsRoot(c, value, nav.TabsRootProps{Key: key, Label: r.Title + " " + key, Vertical: r.Vertical, Manual: s.manual, NoLoop: s.noLoop, OnValueChange: func(v string) { s.status = "Selected " + label(v) }}, func(c *ui.Context) {
		header := ui.Row(c).Gap(16).AlignItems(ui.Center)
		if r.Vertical {
			header.Shrink(0).AlignSelf(ui.Start)
		} else if !r.Compare {
			header.FillWidth().Wrap()
		}
		header.Children(func() {
			nav.TabsList(c, nav.TabsListProps{Variant: variant}, func(c *ui.Context) {
				for i, v := range r.Values {
					p := nav.TabsTriggerProps{Label: label(v), Disabled: i == r.Disabled, HideLabel: r.IconOnly}
					if r.ID == "disabled" && i == r.Disabled {
						p.Label = "Disabled"
					}
					if r.Icons {
						p.Icon = tabIcon(v)
					}
					nav.TabsTrigger(c, v, p)
				}
			})
			if r.Input || r.Dropdown {
				ui.Box(c).Grow(1).MinWidth(0)
			}
			if r.Input {
				control.Input(c, &s.query, control.InputProps{Label: "Search tabs", Placeholder: "Search..."}).Width(176)
				if control.Button(c, "Action", control.ButtonProps{}).Clicked() {
					s.status = "Action for " + *value + ": " + s.query
				}
			}
			if r.Dropdown {
				chosen := overlay.InWindowDropdownMenu(c, &s.menu, overlay.MenuProps{Items: []overlay.MenuItem{{Key: "settings", Label: "Settings"}, {Key: "export", Label: "Export"}, {Separator: true}, {Key: "archive", Label: "Archive"}}}, func(e ui.Element) {
					control.StyleButton(c, e, "", control.ButtonProps{Label: "More options", Icon: outlined.MoreHoriz(), Variant: control.Ghost, Size: control.IconSmall})
				})
				if chosen != "" {
					s.status = "Menu " + chosen + " for " + *value
				}
			}
		})
		if r.Content {
			panel := ui.Column(c).MinWidth(0).Padding(16).Radius(theme.TokensOf(c).RadiusMD()).Border(1, theme.TokensOf(c).Border)
			if r.Vertical {
				panel.Basis(0).Grow(1)
			} else {
				panel.FillWidth()
			}
			panel.Children(func() {
				values := r.Values
				if r.Input {
					values = append(append([]string(nil), values...), "reports")
				}
				for _, v := range values {
					nav.TabsContent(c, v, nav.TabsContentProps{}, func(c *ui.Context) {
						description := map[string]string{"account": "Manage your account preferences and profile information.", "password": "Update your password to keep your account secure.", "notifications": "Configure how you receive notifications and alerts.", "overview": "View your dashboard metrics and key performance indicators.", "analytics": "Detailed analytics and insights about your data.", "reports": "Generate and view custom reports."}[v]
						ui.Text(c, description)
						noteKey := r.ID + "/" + v
						if s.notes[noteKey] == nil {
							s.notes[noteKey] = new(string)
						}
						control.Input(c, s.notes[noteKey], control.InputProps{Label: "Note for " + label(v), Placeholder: "Keep a note for this panel"}).FillWidth()
					})
				}
			})
		}
	})
	if !r.Compare {
		root.FillWidth()
	}
}
func main() {
	snapshot := flag.String("snapshot", "", "render the actual example as PNG")
	dark := flag.Bool("dark", false, "use dark appearance")
	id := flag.String("recipe", "input", "fixed example ID")
	flag.Parse()
	s := newState(*dark, *id)
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		if err = png.Encode(f, ui.Render(s.view, 1000, 620, 1)); err != nil {
			f.Close()
			log.Fatal(err)
		}
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "Tabs composition", Width: 1000, Height: 620, Content: ui.View(s.view)})
		w.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
