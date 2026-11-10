// Package gallery is the native, executable component showcase.
package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func (s *State) heading(c *ui.Context, title, description string) {
	ui.Column(c).Gap(10).Children(func() {
		theme.StyleHeading(c, display.Typography(c, title, display.H1).FontSize(26))
		display.Typography(c, description, display.Muted)
	})
}
func columns(c *ui.Context, builds ...func()) {
	w, _ := c.Size()
	root := ui.Row(c).FillWidth().Gap(20).AlignItems(ui.Start)
	available := w - 48
	if w >= 980 {
		available -= 292
	}
	stacked := available < 720
	if stacked {
		root.Column()
	}
	root.Children(func() {
		for _, build := range builds {
			column := ui.Column(c).MinWidth(0).Gap(20)
			if stacked {
				column.FillWidth().Shrink(0)
			} else {
				column.Basis(0).Grow(1)
			}
			column.Children(build)
		}
	})
}
func (s *State) overview(c *ui.Context) {
	s.heading(c, "Build something beautiful.", "Thoughtful components for your next native Go app.")
	ui.Row(c).Gap(8).Children(func() {
		display.Badge(c, "Native", display.BadgeSecondary)
		display.Badge(c, "Composable", display.BadgeOutline)
		display.Badge(c, "Themeable", display.BadgeOutline)
	})
	columns(c, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Create an account", "Enter your email below to create your account.")
			control.Button(c, "Continue with GitHub", control.ButtonProps{Variant: control.Outline}).FillWidth()
			ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
				display.Separator(c, false).Grow(1)
				display.Typography(c, "OR CONTINUE WITH", display.Muted).FontSize(10)
				display.Separator(c, false).Grow(1)
			})
			control.Field(c, control.FieldProps{Label: "Email"}, func() ui.Element { return control.Input(c, &s.Email, control.InputProps{Placeholder: "m@example.com"}) })
			if control.Button(c, "Create account", control.ButtonProps{}).FillWidth().Clicked() {
				overlay.Toast(c, ui.Toast{Title: "Welcome aboard", Description: "This demo keeps your data on this device."})
			}
		})
		display.Card(c, func() {
			s.cardHeader(c, "Your preferences", "Make yourself at home.")
			control.Switch(c, &s.Notifications, control.SwitchProps{Label: "Push notifications"})
			control.Checkbox(c, &s.Marketing, "Send me product updates", control.CheckboxProps{})
		})
	}, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Activity", "Your projects are moving forward.")
			display.Chart(c, display.ChartProps{Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}, Series: []display.ChartSeries{{Name: "Builds", Values: []float64{32, 48, 36, 72, 56, 88}}}, Height: 170, Label: "Successful builds this week"})
			ui.Row(c).Gap(8).Children(func() {
				display.Badge(c, "+24.8%", display.BadgeSecondary)
				display.Typography(c, "from last week", display.Muted)
			})
		})
		display.Card(c, func() {
			s.cardHeader(c, "Team members", "Invite your team to collaborate.")
			for _, person := range []struct{ name, email, initial string }{{"Sofia Davis", "sofia@example.com", "SD"}, {"Jackson Lee", "jackson@example.com", "JL"}} {
				display.Item(c, display.ItemProps{Title: person.name, Description: person.email, Leading: func() { display.Avatar(c, display.AvatarProps{Fallback: person.initial, Size: 36}) }, Trailing: func() { display.Badge(c, "Member", display.BadgeOutline) }})
			}
		})
	})
}
func (s *State) controls(c *ui.Context) {
	s.heading(c, "Forms & controls", "Keyboard-ready building blocks with controlled state.")
	display.Card(c, func() {
		s.cardHeader(c, "Buttons", "Six variants. Consistent sizing and focus treatment.")
		ui.Row(c).Gap(8).Wrap().Children(func() {
			for i, label := range []string{"Default", "Secondary", "Outline", "Ghost", "Destructive", "Link"} {
				control.Button(c, label, control.ButtonProps{Variant: control.Variant(i)})
			}
			control.Button(c, "Disabled", control.ButtonProps{Disabled: true})
		})
	})
	columns(c, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Profile", "Labels, descriptions and inline validation.")
			control.Field(c, control.FieldProps{Label: "Full name", Required: true}, func() ui.Element { return control.Input(c, &s.Name, control.InputProps{}) })
			control.Field(c, control.FieldProps{Label: "Bio", Description: "Tell us a little about yourself."}, func() ui.Element {
				return control.TextareaWith(c, &s.Bio, control.TextareaProps{Placeholder: "I build things with Go…"}, control.TextareaResizeProps{State: &s.bioResize, Label: "Resize bio"}).Root
			})
			control.InputGroup(c, control.InputGroupProps{Prefix: "https://", Suffix: ".dev", Label: "Website name"}, func() ui.Element {
				return control.Input(c, &s.websiteName, control.InputProps{Placeholder: "your-name"})
			})
			s.otpPreview(c)
		})
	}, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Make a choice", "Select, search, toggle and adjust.")
			selectProps := control.SelectProps{Label: "Framework", Options: []control.Option{{Value: "go", Label: "Go / MyGo"}, {Value: "rust", Label: "Rust"}, {Value: "swift", Label: "Swift"}}}
			if s.Config.NativeMenus {
				desktop.NativeSelect(c, &s.Framework, selectProps)
			} else {
				if _, err := desktop.Select(c, s.Popovers, "preview-framework", &s.Framework, selectProps); err != nil {
					s.popupError(c, err)
				}
			}
			control.RadioGroup(c, &s.Plan, control.RadioGroupProps{Label: "Plan", Orientation: ui.Vertical, Description: "Choose your workspace plan.", Options: []control.Option{{Value: "free", Label: "Free", Description: "For personal projects"}, {Value: "pro", Label: "Pro", Description: "For growing teams"}, {Value: "team", Label: "Team", Description: "For your organization"}}})
			control.Switch(c, &s.Notifications, control.SwitchProps{Label: "Notifications"})
			control.Checkbox(c, &s.Checked, "Accept terms and conditions", control.CheckboxProps{})
			control.Slider(c, &s.Slider, control.SliderProps{Min: 0, Max: 100, Step: 1, Label: "Volume"})
			display.Typography(c, fmt.Sprintf("Volume: %.0f%%", s.Slider), display.Muted)
			control.Toggle(c, &s.Bold, "Bold", control.ToggleProps{Outline: true})
		})
	})
	s.choiceVariantsPreview(c)
	s.recipeChoicesPreview(c)
	s.typedInputs.view(c, s.Window)
	s.inputGroups.view(c, s.Popovers, s.Config.NativeMenus)
	s.sliderRangesPreview(c)
	s.comboboxPreview(c)
	s.comboboxCompoundPreview(c)
}
func (s *State) displays(c *ui.Context) {
	s.heading(c, "Data & display", "Readable information, useful feedback and considered details.")
	columns(c, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Calendar", "Browse months and choose a date.")
			s.previewDatePicker(c)
			s.previewDateRangePicker(c)
			display.Calendar(c, &s.Date)
		})
		display.Card(c, func() {
			s.cardHeader(c, "Loading states", "")
			display.Progress(c, .65)
			display.Skeleton(c, display.SkeletonProps{})
			display.Skeleton(c, display.SkeletonProps{Width: 180})
			display.Spinner(c)
		})
	}, func() {
		s.avatarsPreview(c)
		display.Card(c, func() {
			s.cardHeader(c, "Frequently asked", "Composable disclosure patterns.")
			display.Accordion(c, display.AccordionProps{Single: true, Bordered: true, Items: []display.AccordionItem{{Key: "native", Title: "Is this a native app?", Open: &s.Accordion, Content: func() {
				display.Typography(c, "Yes. Every component is rendered by MyGo using Go, with native input and accessibility.", display.Paragraph)
			}}, {Key: "themes", Title: "How do themes work?", Open: &s.AccordionTheme, Content: func() {
				display.Typography(c, "Semantic colors adapt to your palette, appearance and local theme scope.", display.Paragraph)
			}}, {Key: "future", Title: "Premium options", Open: &s.AccordionDisabled, Disabled: true}}})
			display.CollapsibleWith(c, &s.Collapsible, display.CollapsibleProps{Label: "More details", KeepMounted: true, Content: func() {
				display.Typography(c, "Your notes stay here when this section closes.", display.Muted)
				control.Input(c, &s.Bio, control.InputProps{Label: "Preview notes", Placeholder: "Add a note…"})
			}})
		})
		display.Alert(c, display.AlertProps{Title: "Everything is up to date", Description: "Your workspace is ready for the next idea."})
		display.CodeBlock(c, display.CodeBlockProps{Title: "main.go", Language: "go", Code: "func view(c *ui.Context) {\n    theme.Apply(c)\n    control.Button(c, \"Hello, Go\", control.ButtonProps{})\n}", LineNumbers: true, Height: 130})
	})
	s.recipeProgressPreview(c)
	s.recipeAvatarPreview(c)
	s.recipeSkeletonPreview(c)
	s.tableCompositionPreview(c)
	display.Card(c, func() {
		s.cardHeader(c, "Recent invoices", "Sortable columns and paginated rows.")
		display.DataTable(c, &s.Table, display.DataTableProps[invoice]{Rows: invoices, Label: "Invoices", Key: func(i invoice) any { return i.ID }, Columns: []display.DataColumn[invoice]{{ID: "id", Title: "Invoice", Value: func(i invoice) string { return i.ID }}, {ID: "status", Title: "Status", Value: func(i invoice) string { return i.Status }}, {ID: "amount", Title: "Amount", Value: func(i invoice) string { return fmt.Sprintf("$%.2f", i.Amount) }, Compare: func(a, b invoice) int {
			if a.Amount < b.Amount {
				return -1
			}
			if a.Amount > b.Amount {
				return 1
			}
			return 0
		}}}})
	})
	s.payments.view(c, s.Popovers, s.Config.NativeMenus)
	s.workspacePreview(c)
	s.chartCompositionPreview(c)
	s.chartSourcePreview(c)
	s.carouselPreview(c)
}

type invoice struct {
	ID, Status string
	Amount     float64
}

var invoices = []invoice{{"INV001", "Paid", 250}, {"INV002", "Pending", 150}, {"INV003", "Paid", 350}, {"INV004", "Processing", 450}}

func (s *State) overlays(c *ui.Context) {
	s.heading(c, "Overlays", "Focused interactions, with native dismissal and focus restoration.")
	display.Card(c, func() {
		s.cardHeader(c, "Open something", "Try Escape, Tab, the arrow keys and an outside click.")
		ui.Row(c).Gap(12).Wrap().Children(func() {
			if control.Button(c, "Edit profile", control.ButtonProps{Variant: control.Outline}).Clicked() {
				s.Dialog = true
			}
			if control.Button(c, "Open sheet", control.ButtonProps{Variant: control.Outline}).Clicked() {
				s.Sheet = true
			}
			b := control.Button(c, "Popover", control.ButtonProps{Variant: control.Outline})
			s.previewPopover(c, b)
			b = control.Button(c, "Hover for a tip", control.ButtonProps{Variant: control.Outline})
			s.previewTooltip(c, b)
			b = control.Button(c, "Preview author", control.ButtonProps{Variant: control.Outline})
			s.previewHoverCard(c, b)
			s.previewMenu(c, "preview-dropdown", "Open menu", overlay.MenuProps{Items: []overlay.MenuItem{{Key: "profile", Label: "Profile"}, {Key: "settings", Label: "Settings"}, {Separator: true}, {Key: "logout", Label: "Log out", Destructive: true}}})
			if control.Button(c, "Show toast", control.ButtonProps{}).Clicked() {
				overlay.Toast(c, ui.Toast{Title: "Event created", Description: "Monday, November 9 at 9:00 AM"})
			}
		})
	})
	s.notificationsPreview(c)
	display.Card(c, func() {
		s.cardHeader(c, "Command palette", "Search by name. Select with the keyboard.")
		overlay.Command(c, &s.commandPreviewQuery, overlay.CommandProps{InputLabel: "Example command search", Items: []overlay.CommandItem{{Value: "calendar", Label: "Calendar", Group: "Suggestions"}, {Value: "search", Label: "Search", Group: "Suggestions"}, {Value: "settings", Label: "Settings", Group: "Workspace"}}})
	})
}
func (s *State) themes(c *ui.Context) {
	s.heading(c, "Colors that work together.", "Your base, accent and chart selections apply to every semantic token.")
	for _, dark := range []bool{false, true} {
		p := s.Config
		if dark {
			p.Appearance = theme.DarkMode
		} else {
			p.Appearance = theme.LightMode
		}
		t := p.Tokens(dark)
		theme.Scope(c, t).FillWidth().Padding(24).Gap(16).Radius(t.Radius).Border(1, t.Border).Background(t.Background).TextColor(t.Foreground).Children(func() {
			label := "Light"
			if dark {
				label = "Dark"
			}
			theme.StyleHeading(c, ui.Text(c, label).FontSize(20).FontWeight(600))
			ui.Row(c).Gap(8).Wrap().Children(func() {
				control.Button(c, "Primary", control.ButtonProps{})
				control.Button(c, "Secondary", control.ButtonProps{Variant: control.Secondary})
				control.Button(c, "Outline", control.ButtonProps{Variant: control.Outline})
				control.Button(c, "Destructive", control.ButtonProps{Variant: control.Destructive})
			})
			ui.Row(c).Gap(10).Wrap().Children(func() {
				for _, color := range []struct {
					name  string
					color ui.Color
				}{{"background", t.Background}, {"foreground", t.Foreground}, {"primary", t.Primary}, {"muted", t.Muted}, {"accent", t.Accent}, {"border", t.Border}, {"ring", t.Ring}, {"chart-1", t.Chart1}, {"chart-2", t.Chart2}, {"chart-3", t.Chart3}, {"chart-4", t.Chart4}, {"chart-5", t.Chart5}} {
					ui.Column(c).Gap(6).Children(func() {
						ui.Box(c).Size(56, 40).Radius(5).Background(color.color).Border(1, t.Border)
						ui.Text(c, color.name).FontSize(10).TextColor(t.MutedForeground)
					})
				}
			})
		})
	}
}
