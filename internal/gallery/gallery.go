// Package gallery is the native, executable component showcase.
package gallery

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type State struct {
	Theme                                                                          theme.Controller
	Page                                                                           string
	Name, Email, Bio, Framework, Plan, OTP, Query                                  string
	Notifications, Marketing, Checked, Bold, Dialog, Sheet, Popover, Menu, Command bool
	Accordion, Collapsible                                                         bool
	Slider                                                                         float64
	Tab, Slide, PageNumber                                                         int
	Date                                                                           time.Time
	Table                                                                          display.DataTableState
	Questionnaire                                                                  conversation.QuestionnaireState
	Trace                                                                          conversation.TraceState
}

func New() *State {
	return &State{Page: "overview", Name: "Sofia Davis", Email: "sofia@example.com", Framework: "go", Plan: "pro", Slider: 65, Notifications: true, Accordion: true, PageNumber: 1, Date: time.Date(2026, 11, 9, 0, 0, 0, 0, time.Local)}
}

var pages = []navigation.NavItem{
	{Key: "overview", Label: "Overview", Group: "Get started"},
	{Key: "controls", Label: "Forms & controls", Group: "Components"},
	{Key: "display", Label: "Data & display"},
	{Key: "overlays", Label: "Overlays"},
	{Key: "conversation", Label: "Messages & questions"},
	{Key: "themes", Label: "Colors & themes", Group: "Design system"},
}

func (s *State) View(c *ui.Context) {
	s.Theme.Apply(c)
	t := theme.TokensOf(c)
	ui.Column(c).Fill().Background(t.Background).TextColor(t.Foreground).FontSize(14).Children(func() {
		ui.Row(c).Height(64).FillWidth().PaddingX(28).Gap(16).AlignItems(ui.Center).InsetShadow(0, -1, 0, 0, t.Border).Children(func() {
			ui.Text(c, "◩").FontSize(28).FontWeight(700)
			ui.Text(c, "shadcn/ui").FontSize(18).FontWeight(600)
			display.Badge(c, "GO", display.BadgeOutline)
			ui.Box(c).Grow(1)
			b := control.Button(c, "Search components…", control.ButtonProps{Variant: control.Outline}).Width(200).Justify(ui.Start)
			if b.Clicked() {
				s.Command = true
			}
			b = control.Button(c, "◐", control.ButtonProps{Variant: control.Ghost, Size: control.Icon}).Label("Toggle theme")
			if b.Clicked() {
				s.Theme.Toggle(c.Theme().Dark)
			}
		})
		ui.Row(c).Grow(1).MinHeight(0).FillWidth().AlignItems(ui.Stretch).Children(func() {
			navigation.Sidebar(c, &s.Page, navigation.SidebarProps{Items: pages, Width: 210, Label: "Gallery sections"}, func() {
				display.Typography(c, "Native Go components", display.Small)
				display.Typography(c, "Powered by MyGo", display.Muted)
			}).InsetShadow(-1, 0, 0, 0, t.Border)
			ui.Scroll(c.Key(s.Page)).Grow(1).MinWidth(0).Children(func() {
				ui.Column(c).FillWidth().MaxWidth(1120).Padding(32).Gap(28).Children(func() {
					switch s.Page {
					case "controls":
						s.controls(c)
					case "display":
						s.displays(c)
					case "overlays":
						s.overlays(c)
					case "conversation":
						s.conversation(c)
					case "themes":
						s.themes(c)
					default:
						s.overview(c)
					}
				})
			})
		})
	})
	if c.Shortcut(ui.Cmd, ui.KeyK) {
		s.Command = true
	}
	if choice := overlay.CommandDialog(c, &s.Command, &s.Query, overlay.CommandProps{Items: []overlay.CommandItem{{Value: "overview", Label: "Overview"}, {Value: "controls", Label: "Forms & controls"}, {Value: "display", Label: "Data & display"}, {Value: "overlays", Label: "Overlays"}, {Value: "themes", Label: "Colors & themes"}}, Placeholder: "Find a section…"}); choice != "" {
		s.Page = choice
		s.Command = false
	}
	overlay.Dialog(c, &s.Dialog, overlay.DialogProps{Title: "Edit profile", Description: "Make changes to your profile. Save when you are done."}, func() {
		control.Field(c, control.FieldProps{Label: "Name"}, func() ui.Element { return control.Input(c, &s.Name, control.InputProps{}) })
		control.Field(c, control.FieldProps{Label: "Email"}, func() ui.Element { return control.Input(c, &s.Email, control.InputProps{}) })
		if control.Button(c, "Save changes", control.ButtonProps{}).AlignSelf(ui.End).Clicked() {
			s.Dialog = false
			overlay.Toast(c, ui.Toast{Title: "Profile updated", Description: "Your changes have been saved."})
		}
	})
	overlay.Sheet(c, &s.Sheet, overlay.SheetProps{Title: "Settings", Description: "Your workspace preferences", Side: overlay.Right}, func() { control.Switch(c, &s.Notifications, control.SwitchProps{Label: "Notifications"}) })
	overlay.Toaster(c)
}
func heading(c *ui.Context, title, description string) {
	ui.Column(c).Gap(10).Children(func() { display.Typography(c, title, display.H1); display.Typography(c, description, display.Lead) })
}
func columns(c *ui.Context, builds ...func()) {
	w, _ := c.Size()
	root := ui.Row(c).FillWidth().Gap(20).AlignItems(ui.Start)
	if w < 950 {
		root.Column()
	}
	root.Children(func() {
		for _, build := range builds {
			ui.Column(c).Basis(0).Grow(1).MinWidth(0).Gap(20).Children(build)
		}
	})
}
func (s *State) overview(c *ui.Context) {
	heading(c, "Build something beautiful.", "Thoughtful components for your next native Go app.")
	ui.Row(c).Gap(8).Children(func() {
		display.Badge(c, "Native", display.BadgeSecondary)
		display.Badge(c, "Composable", display.BadgeOutline)
		display.Badge(c, "Themeable", display.BadgeOutline)
	})
	columns(c, func() {
		display.Card(c, func() {
			display.CardHeader(c, "Create an account", "Enter your email below to create your account.")
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
			display.CardHeader(c, "Your preferences", "Make yourself at home.")
			control.Switch(c, &s.Notifications, control.SwitchProps{Label: "Push notifications"})
			control.Checkbox(c, &s.Marketing, "Send me product updates", control.CheckboxProps{})
		})
	}, func() {
		display.Card(c, func() {
			display.CardHeader(c, "Activity", "Your projects are moving forward.")
			display.Chart(c, display.ChartProps{Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}, Series: []display.ChartSeries{{Name: "Builds", Values: []float64{32, 48, 36, 72, 56, 88}}}, Height: 170, Label: "Successful builds this week"})
			ui.Row(c).Gap(8).Children(func() {
				display.Badge(c, "+24.8%", display.BadgeSecondary)
				display.Typography(c, "from last week", display.Muted)
			})
		})
		display.Card(c, func() {
			display.CardHeader(c, "Team members", "Invite your team to collaborate.")
			for _, person := range []struct{ name, email, initial string }{{"Sofia Davis", "sofia@example.com", "SD"}, {"Jackson Lee", "jackson@example.com", "JL"}} {
				display.Item(c, display.ItemProps{Title: person.name, Description: person.email, Leading: func() { display.Avatar(c, display.AvatarProps{Fallback: person.initial, Size: 36}) }, Trailing: func() { display.Badge(c, "Member", display.BadgeOutline) }})
			}
		})
	})
}
func (s *State) controls(c *ui.Context) {
	heading(c, "Forms & controls", "Keyboard-ready building blocks with controlled state.")
	display.Card(c, func() {
		display.CardHeader(c, "Buttons", "Six variants. Consistent sizing and focus treatment.")
		ui.Row(c).Gap(8).Wrap().Children(func() {
			for i, label := range []string{"Default", "Secondary", "Outline", "Ghost", "Destructive", "Link"} {
				control.Button(c, label, control.ButtonProps{Variant: control.Variant(i)})
			}
			control.Button(c, "Disabled", control.ButtonProps{Disabled: true})
		})
	})
	columns(c, func() {
		display.Card(c, func() {
			display.CardHeader(c, "Profile", "Labels, descriptions and inline validation.")
			control.Field(c, control.FieldProps{Label: "Full name", Required: true}, func() ui.Element { return control.Input(c, &s.Name, control.InputProps{}) })
			control.Field(c, control.FieldProps{Label: "Bio", Description: "Tell us a little about yourself."}, func() ui.Element {
				return control.Textarea(c, &s.Bio, control.TextareaProps{Placeholder: "I build things with Go…"})
			})
			control.InputGroup(c, control.InputGroupProps{Prefix: "https://", Suffix: ".dev"}, func() ui.Element { return control.Input(c, &s.Query, control.InputProps{Placeholder: "your-name"}) })
			control.InputOTP(c, &s.OTP, control.InputOTPProps{Length: 6, Label: "Verification code"})
		})
	}, func() {
		display.Card(c, func() {
			display.CardHeader(c, "Make a choice", "Select, search, toggle and adjust.")
			control.Select(c, &s.Framework, control.SelectProps{Label: "Framework", Options: []control.Option{{Value: "go", Label: "Go / MyGo"}, {Value: "rust", Label: "Rust"}, {Value: "swift", Label: "Swift"}}})
			control.RadioGroup(c, &s.Plan, control.RadioGroupProps{Label: "Plan", Options: []control.Option{{Value: "free", Label: "Free"}, {Value: "pro", Label: "Pro"}, {Value: "team", Label: "Team"}}})
			control.Switch(c, &s.Notifications, control.SwitchProps{Label: "Notifications"})
			control.Checkbox(c, &s.Checked, "Accept terms and conditions", control.CheckboxProps{})
			control.Slider(c, &s.Slider, control.SliderProps{Min: 0, Max: 100, Step: 1, Label: "Volume"})
			display.Typography(c, fmt.Sprintf("Volume: %.0f%%", s.Slider), display.Muted)
			control.Toggle(c, &s.Bold, "Bold", control.ToggleProps{Outline: true})
		})
	})
}
func (s *State) displays(c *ui.Context) {
	heading(c, "Data & display", "Readable information, useful feedback and considered details.")
	columns(c, func() {
		display.Card(c, func() {
			display.CardHeader(c, "Calendar", "Browse months and choose a date.")
			display.DatePicker(c, &s.Date, display.DatePickerProps{Label: "Project date"})
			display.Calendar(c, &s.Date)
		})
		display.Card(c, func() {
			display.CardHeader(c, "Loading states", "")
			display.Progress(c, .65)
			display.Skeleton(c, display.SkeletonProps{})
			display.Skeleton(c, display.SkeletonProps{Width: 180})
			display.Spinner(c)
		})
	}, func() {
		display.Card(c, func() {
			display.CardHeader(c, "Frequently asked", "Composable disclosure patterns.")
			display.Accordion(c, display.AccordionProps{Single: true, Items: []display.AccordionItem{{Key: "native", Title: "Is this a native app?", Open: &s.Accordion, Content: func() {
				display.Typography(c, "Yes. Every component is rendered by MyGo using Go, with native input and accessibility.", display.Paragraph)
			}}}})
			display.Collapsible(c, "More details", &s.Collapsible, func() { display.Typography(c, "Components use your application's own state.", display.Muted) })
		})
		display.Alert(c, display.AlertProps{Title: "Everything is up to date", Description: "Your workspace is ready for the next idea."})
		display.CodeBlock(c, display.CodeBlockProps{Title: "main.go", Language: "go", Code: "func view(c *ui.Context) {\n    theme.Apply(c)\n    control.Button(c, \"Hello, Go\", control.ButtonProps{})\n}", LineNumbers: true, Height: 130})
	})
	display.Card(c, func() {
		display.CardHeader(c, "Recent invoices", "Sortable columns and paginated rows.")
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
}

type invoice struct {
	ID, Status string
	Amount     float64
}

var invoices = []invoice{{"INV001", "Paid", 250}, {"INV002", "Pending", 150}, {"INV003", "Paid", 350}, {"INV004", "Processing", 450}}

func (s *State) overlays(c *ui.Context) {
	heading(c, "Overlays", "Focused interactions, with native dismissal and focus restoration.")
	display.Card(c, func() {
		display.CardHeader(c, "Open something", "Try Escape, Tab, the arrow keys and an outside click.")
		ui.Row(c).Gap(12).Wrap().Children(func() {
			if control.Button(c, "Edit profile", control.ButtonProps{Variant: control.Outline}).Clicked() {
				s.Dialog = true
			}
			if control.Button(c, "Open sheet", control.ButtonProps{Variant: control.Outline}).Clicked() {
				s.Sheet = true
			}
			b := control.Button(c, "Popover", control.ButtonProps{Variant: control.Outline})
			overlay.Popover(c, b, &s.Popover, overlay.PopoverProps{}, func() {
				display.CardHeader(c, "Dimensions", "Set the dimensions for the layer.")
				control.Input(c, &s.Name, control.InputProps{Label: "Layer name"})
			})
			b = control.Button(c, "Hover for a tip", control.ButtonProps{Variant: control.Outline})
			overlay.Tooltip(c, b, "Small details make a difference.", overlay.TooltipProps{})
			overlay.DropdownMenu(c, &s.Menu, overlay.MenuProps{Items: []overlay.MenuItem{{Key: "profile", Label: "Profile"}, {Key: "settings", Label: "Settings"}, {Separator: true}, {Key: "logout", Label: "Log out", Destructive: true}}}, func(b ui.Element) {
				control.StyleButton(c, b, "Open menu", control.ButtonProps{Variant: control.Outline})
			})
			if control.Button(c, "Show toast", control.ButtonProps{}).Clicked() {
				overlay.Toast(c, ui.Toast{Title: "Event created", Description: "Monday, November 9 at 9:00 AM"})
			}
		})
	})
	display.Card(c, func() {
		display.CardHeader(c, "Command palette", "Search by name. Select with the keyboard.")
		overlay.Command(c, &s.Query, overlay.CommandProps{Items: []overlay.CommandItem{{Value: "calendar", Label: "Calendar", Group: "Suggestions"}, {Value: "search", Label: "Search", Group: "Suggestions"}, {Value: "settings", Label: "Settings", Group: "Workspace"}}})
	})
}
func (s *State) themes(c *ui.Context) {
	heading(c, "Make it yours.", "Semantic colors, light and dark appearance, shared across every component.")
	display.Card(c, func() {
		display.CardHeader(c, "Appearance", "Switch at runtime or follow your desktop.")
		ui.Row(c).Gap(8).Children(func() {
			for _, mode := range []theme.Mode{theme.System, theme.LightMode, theme.DarkMode} {
				if control.Button(c, string(mode), control.ButtonProps{Variant: control.Outline}).Clicked() {
					s.Theme.Mode = mode
				}
			}
		})
		control.Switch(c, &s.Theme.ReduceMotion, control.SwitchProps{Label: "Reduce motion"})
		ui.Row(c).Gap(8).Wrap().Children(func() {
			for _, preset := range theme.Presets() {
				tk := theme.Preset(preset, c.Theme().Dark)
				b := control.Button(c, preset, control.ButtonProps{Variant: control.Outline}).Border(2, tk.Primary)
				if b.Clicked() {
					s.Theme.Preset = preset
				}
			}
		})
	})
	for _, dark := range []bool{false, true} {
		t := theme.Preset(s.Theme.Preset, dark)
		theme.Scope(c, t).FillWidth().Padding(24).Gap(16).Radius(t.Radius).Background(t.Background).TextColor(t.Foreground).Border(1, t.Border).Children(func() {
			label := "Light"
			if dark {
				label = "Dark"
			}
			display.Typography(c, label, display.H3)
			ui.Row(c).Gap(8).Children(func() {
				control.Button(c, "Primary", control.ButtonProps{})
				control.Button(c, "Secondary", control.ButtonProps{Variant: control.Secondary})
				control.Button(c, "Outline", control.ButtonProps{Variant: control.Outline})
				control.Button(c, "Destructive", control.ButtonProps{Variant: control.Destructive})
			})
			ui.Row(c).Gap(12).Wrap().Children(func() {
				for _, color := range []struct {
					name  string
					color ui.Color
				}{{"background", t.Background}, {"foreground", t.Foreground}, {"primary", t.Primary}, {"secondary", t.Secondary}, {"muted", t.Muted}, {"accent", t.Accent}, {"destructive", t.Destructive}, {"border", t.Border}, {"ring", t.Ring}} {
					ui.Column(c).Gap(6).Children(func() {
						ui.Box(c).Size(68, 48).Radius(6).Background(color.color).Border(1, t.Border)
						ui.Text(c, color.name).FontSize(10).TextColor(t.MutedForeground)
					})
				}
			})
		})
	}
}
