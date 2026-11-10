// Package gallery is a native design-system creator and live component preview.
package gallery

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/chartcatalog"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type State struct {
	lastSolidMenuAccent theme.MenuAccent
	// Theme.Mode remains supported for command-line appearance selection.
	// The creator's Config controls the scoped live preview independently of chrome.
	Theme                                                                          theme.Controller
	Config                                                                         Preset
	Window                                                                         *mygo.Window
	Popovers                                                                       *desktop.Host
	Page                                                                           string
	pendingPreview                                                                 string
	Name, Email, Bio, Framework, Plan, OTP, Query                                  string
	Notifications, Marketing, Checked, Bold, Dialog, Sheet, Popover, Menu, Command bool
	Accordion, Collapsible                                                         bool
	AccordionTheme, AccordionDisabled                                              bool
	Slider                                                                         float64
	SliderRange                                                                    [2]float64
	SliderBands                                                                    []float64
	VerticalLevel                                                                  float64
	WorkspaceLayout, EditorLayout                                                  display.ResizableLayout
	Menubar                                                                        desktop.MenubarState
	Sidebar                                                                        navigation.SidebarState
	Tab, Slide, PageNumber                                                         int
	Date                                                                           time.Time
	DateRange                                                                      display.DateRange
	Table                                                                          display.DataTableState
	tablePreview                                                                   tablePreviewState
	payments                                                                       paymentsPreviewState
	recipes                                                                        recipeCompositionState
	typedInputs                                                                    typedInputPreviewState
	inputGroups                                                                    inputGroupPreviewState
	controlCompounds                                                               controlCompoundsPreviewState
	cardCompositions                                                               cardCompositionPreviewState
	fieldCompositions                                                              fieldCompositionPreviewState
	statusCompositions                                                             statusCompositionPreviewState
	disclosureCompositions                                                         disclosureCompositionPreviewState
	itemCompositions                                                               itemCompositionPreviewState
	popupCompositions                                                              popupCompositionPreviewState
	dropdownCompositions                                                           dropdownCompositionPreviewState
	contextMenuCompositions                                                        contextMenuCompositionPreviewState
	menubarCompositions                                                            menubarCompositionPreviewState
	badgeCompositions                                                              badgeCompositionPreviewState
	buttonCompositions                                                             buttonCompositionPreviewState
	switchCompositions                                                             switchCompositionPreviewState
	checkboxCompositions                                                           checkboxCompositionPreviewState
	radioCompositions                                                              radioCompositionPreviewState
	textareaCompositions                                                           textareaCompositionPreviewState
	inputCompositions                                                              inputCompositionPreviewState
	otpCompositions                                                                otpCompositionPreviewState
	sliderCompositions                                                             sliderCompositionPreviewState
	selectCompositions                                                             selectCompositionPreviewState
	tabsCompositions                                                               tabsCompositionPreviewState
	dialogCompositions                                                             dialogCompositionPreviewState
	chartPreview                                                                   chartPreviewState
	chartCatalogue                                                                 chartcatalog.State
	comboboxCompound                                                               comboboxCompoundPreviewState
	comboboxes                                                                     comboboxPreviewState
	carousels                                                                      carouselPreviewState
	conversationPreview                                                            conversationPreviewState
	ConversationComposition                                                        conversationCompositionState
	Questionnaire                                                                  conversation.QuestionnaireState
	QuestionnairePreview                                                           questionnairePreviewState
	questionnaireRecipes                                                           questionnaireRecipesState
	Trace                                                                          conversation.TraceState
	undo, redo                                                                     []Preset
	lastThemeMode                                                                  theme.Mode
	previewIndex                                                                   int
	lastPreviewPage                                                                string
	lastPreviewWidth                                                               float32
	randomSequence                                                                 uint64
	designRevision                                                                 uint64
	importOpen, tokenOpen                                                          bool
	importText, importError, status                                                string
	fontError                                                                      string
	nativeMenuOpen                                                                 bool
	lineTab, manualTab                                                             int
	navRoute                                                                       string
	navigationMenuAction                                                           string
	sidebarRoute                                                                   string
	contentAction, contentProject                                                  string
	contentPinned, contentDismissed                                                bool
	formStatus, inviteCode, breadcrumbSelection                                    string
	websiteName, commandPreviewQuery                                               string
	formatting                                                                     []string
	readAccess, writeAccess, quietNotifications                                    bool
	bioResize                                                                      control.TextareaResizeState
	loadingStatic, codeWrap, codeCompact                                           bool
	codeDiff, articleDetails, nativeScrollbars                                     bool
	codeLanguage, codeCopied, layoutDraft                                          string
	codeExample                                                                    display.CodeBlockState
	articleScroll                                                                  ui.ScrollState
	toastPreview                                                                   overlay.ToasterState
	toastKind, toastPosition, toastUploadID, toastResult                           string
	toastRich, toastClose, toastExpand, toastKeepOpen                              bool
	toastShow, toastStack, toastStart, toastComplete                               ui.Handle
	toastPageReturn                                                                ui.Handle
	toastReturnFocus                                                               *ui.Handle
	menuStates                                                                     map[string]*creatorMenu
	pendingMenuAction                                                              string
}

func New() *State {
	p := DefaultPreset()
	return &State{Config: p, Theme: theme.Controller{Mode: p.Appearance, Preset: p.Accent, Radius: p.Radius}, lastThemeMode: p.Appearance, Page: "overview", Name: "Sofia Davis", Email: "sofia@example.com", Framework: "go", Plan: "pro", Slider: 65, SliderRange: [2]float64{25, 75}, SliderBands: []float64{20, 50, 80}, VerticalLevel: 60, Notifications: true, readAccess: true, formatting: []string{"bold"}, Accordion: true, PageNumber: 1, Date: time.Date(2026, 11, 9, 0, 0, 0, 0, time.Local)}
}

type creatorPage struct {
	Key, Label string
	Keywords   []string
}

var pages = []creatorPage{
	{Key: "overview", Label: "Cards", Keywords: []string{"Card", "CardSpacing", "Card compositions", "Meeting Notes"}},
	{Key: "controls", Label: "Forms & controls", Keywords: []string{"Button", "Combobox", "MultiCombobox", "Autocomplete", "ButtonGroup", "RadioGroup", "Radio", "Field", "FieldComposition", "FieldSet", "FieldGroup", "Input", "InputGroup", "NumberInput", "DateInput", "TimeInput", "FileInput", "DatePickerInput", "Booking", "Jalali", "Persian", "Slider", "Switch", "Checkbox", "Textarea", "InputOTP", "Select"}},
	{Key: "display", Label: "Data & display", Keywords: []string{"Table", "DataTable", "Payments", "Payment", "Calendar", "Avatar", "AvatarDropdown", "Accordion", "Collapsible", "Disclosure", "RadiusSettings", "FileTree", "Chart", "BarChart", "LineChart", "AreaChart", "PieChart", "DonutChart", "RadarChart", "RadialChart", "Carousel", "Slides", "Slideshow", "Progress", "Skeleton"}},
	{Key: "content", Label: "Content & feedback", Keywords: []string{"Alert", "AlertColors", "Empty", "EmptySearch", "Item", "Badge"}},
	{Key: "text-layout", Label: "Text & layout", Keywords: []string{"Typography", "Kbd", "Keyboard", "CodeBlock", "Code", "Spinner", "Skeleton", "ScrollArea", "AspectRatio"}},
	{Key: "overlays", Label: "Overlays", Keywords: []string{"AlertDialog", "Dialog", "Sheet", "Popover", "Tooltip", "HoverCard", "DropdownMenu", "Toast", "Sonner", "Command"}},
	{Key: "conversation", Label: "Messages & questions", Keywords: []string{"Bubble", "Attachment", "Trace", "Message", "Marker", "MessageScroller", "Questionnaire", "Questionnaire workflows", "Migration", "Conditional"}},
	{Key: "navigation", Label: "Navigation", Keywords: []string{"Tabs", "ContextMenu", "Menubar"}},
	{Key: "themes", Label: "Colors & themes"},
}

func (s *State) View(c *ui.Context) {
	if s.Popovers != nil {
		s.Popovers.BeginFrame()
		defer s.Popovers.EndFrame()
	}
	// A command-line mode is an initial application preference, not an undo step.
	if s.Theme.Mode != s.lastThemeMode {
		mode := s.Theme.Mode
		if mode == "" {
			mode = theme.System
		}
		s.Config.Appearance = mode
		s.lastThemeMode = mode
		s.Theme.Mode = mode
		s.designRevision++
	}
	systemDark := c.Theme().Dark
	if s.Config.ReduceMotion {
		c.DisableAnimations()
	}
	for i, page := range pages {
		if page.Key == s.Page {
			s.previewIndex = i
			break
		}
	}
	preview := s.Config.Tokens(systemDark)
	chrome := chromeTokens()
	theme.Scope(c, chrome).Fill().Background(chrome.Background).TextColor(chrome.Foreground).FontSize(13).Children(func() {
		s.header(c, preview)
		width, _ := c.Size()
		compact := width < 980
		body := ui.Row(c).Grow(1).MinHeight(0).FillWidth().Padding(20).Gap(20).AlignItems(ui.Stretch)
		if compact {
			body.Column().Padding(12).Gap(12)
		}
		body.Children(func() {
			s.customizer(c, compact, preview)
			theme.FontScope(c, s.previewFonts()).Grow(1).MinHeight(0).MinWidth(0).Children(func() {
				s.preview(c, preview)
			})
		})
		s.creatorDialogs(c)
	})
	if c.Shortcut(ui.Cmd|ui.Shift, ui.KeyZ) {
		s.Redo()
	} else if c.Shortcut(ui.Cmd, ui.KeyZ) {
		s.Undo()
	}
	if c.Shortcut(ui.Cmd, ui.KeyK) {
		s.Command = true
	}
}

func (s *State) header(c *ui.Context, preview *theme.Tokens) {
	t := theme.TokensOf(c)
	ui.Row(c).Height(58).FillWidth().PaddingX(24).Gap(10).AlignItems(ui.Center).InsetShadow(0, -1, 0, 0, t.Border).Children(func() {
		ui.Text(c, "◩").FontSize(25).FontWeight(700)
		ui.Text(c, "shadcn/ui").FontSize(16).FontWeight(600)
		ui.Text(c, "/").TextColor(t.MutedForeground).Margin(0, 6)
		ui.Text(c, "Create").FontSize(14).FontWeight(500)
		display.Badge(c, "Native Go", display.BadgeOutline).AlignSelf(ui.Center)
		ui.Box(c).Grow(1)
		if control.Button(c, "Search", control.ButtonProps{Variant: control.Ghost, Size: control.Small, Label: "Search components"}).Clicked() {
			s.Command = true
		}
		if control.Button(c, "◐", control.ButtonProps{Variant: control.Ghost, Size: control.IconSmall, Label: "Toggle theme"}).Clicked() {
			p := s.Config
			if preview.Dark {
				p.Appearance = theme.LightMode
			} else {
				p.Appearance = theme.DarkMode
			}
			_ = s.ApplyPreset(p)
		}
	})
}

func (s *State) customizer(c *ui.Context, compact bool, preview *theme.Tokens) {
	t := theme.TokensOf(c)
	card := ui.Column(c).Width(232).MinHeight(0).Radius(14).Border(1, t.Border).Background(t.Card).Clip()
	if compact {
		card.FillWidth().Height(252)
	}
	card.Children(func() {
		ui.Row(c).Height(48).FillWidth().PaddingX(14).Gap(6).AlignItems(ui.Center).InsetShadow(0, -1, 0, 0, t.Border).Children(func() {
			ui.Text(c, "Design system").FontSize(13).FontWeight(600).Grow(1)
			if control.Button(c, "↶", control.ButtonProps{Variant: control.Ghost, Size: control.IconSmall, Disabled: !s.CanUndo(), Label: "Undo design change"}).Clicked() {
				s.Undo()
			}
			if control.Button(c, "↷", control.ButtonProps{Variant: control.Ghost, Size: control.IconSmall, Disabled: !s.CanRedo(), Label: "Redo design change"}).Clicked() {
				s.Redo()
			}
		})
		ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Children(func() {
			ui.Column(c).FillWidth().Padding(14).Gap(14).Children(func() {
				s.choice(c, "Base color", s.Config.Base, baseColors, func(v string) { p := s.Config; p.Base = v; _ = s.ApplyPreset(p) })
				s.choice(c, "Theme", s.Config.Accent, theme.Presets(), func(v string) { p := s.Config; p.Accent = v; _ = s.ApplyPreset(p) })
				s.choice(c, "Chart color", s.Config.Chart, theme.Presets(), func(v string) { p := s.Config; p.Chart = v; _ = s.ApplyPreset(p) })
				display.Separator(c, false)
				s.choice(c, "Heading", s.Config.HeadingFont, headingFonts, func(v string) { s.setFont(true, v) })
				s.choice(c, "Font", s.Config.BodyFont, bodyFonts, func(v string) { s.setFont(false, v) })
				display.Separator(c, false)
				radii := make([]string, len(radiusValues))
				for i, r := range radiusValues {
					radii[i] = strconv.FormatFloat(float64(r), 'f', -1, 32)
				}
				radiusValue := strconv.FormatFloat(float64(s.Config.Radius), 'f', -1, 32)
				foundRadius := false
				for _, r := range radii {
					if r == radiusValue {
						foundRadius = true
						break
					}
				}
				if !foundRadius {
					radii = append(radii, radiusValue)
				}
				s.choice(c, "Radius", radiusValue, radii, func(v string) {
					r, _ := strconv.ParseFloat(v, 32)
					p := s.Config
					p.Radius = float32(r)
					_ = s.ApplyPreset(p)
				})
				s.choice(c, "Appearance", string(s.Config.Appearance), []string{"system", "light", "dark"}, func(v string) { p := s.Config; p.Appearance = theme.Mode(v); _ = s.ApplyPreset(p) })
				display.Separator(c, false)
				presentation := "flyout"
				if s.Config.NativeMenus {
					presentation = "native"
				}
				s.choice(c, "Menu style", presentation, []string{"flyout", "native"}, func(value string) { p := s.Config; p.NativeMenus = value == "native"; _ = s.ApplyPreset(p) })
				if !(theme.MenuOptions{Color: s.Config.MenuColor}).ColorIsTranslucent() {
					s.lastSolidMenuAccent = s.Config.MenuAccent
				}
				s.choice(c, "Menu color", string(s.Config.MenuColor), []string{"default", "inverted", "default-translucent", "inverted-translucent"}, func(value string) { s.setMenuColor(theme.MenuColor(value)) })
				accents := []string{"subtle", "bold"}
				if (theme.MenuOptions{Color: s.Config.MenuColor}).ColorIsTranslucent() {
					accents = []string{"subtle"}
				}
				s.choice(c, "Menu accent", string(s.Config.MenuAccent), accents, func(value string) { p := s.Config; p.MenuAccent = theme.MenuAccent(value); _ = s.ApplyPreset(p) })
				s.choiceSwitch(c, "Reduce motion", s.Config.ReduceMotion, func(value bool) { p := s.Config; p.ReduceMotion = value; _ = s.ApplyPreset(p) })
			})
		})
		ui.Column(c).FillWidth().Padding(12).Gap(8).InsetShadow(0, 1, 0, 0, t.Border).Children(func() {
			ui.Row(c).FillWidth().Gap(6).Children(func() {
				if control.Button(c, "Copy preset", control.ButtonProps{Variant: control.Outline, Size: control.Small}).Grow(1).Clicked() {
					c.WriteClipboard(s.ExportPreset())
					s.status = "Preset copied"
				}
				if control.Button(c, "Open", control.ButtonProps{Variant: control.Outline, Size: control.Small, Label: "Open preset"}).Clicked() {
					s.importOpen = true
					s.importError = ""
				}
			})
			ui.Row(c).FillWidth().Gap(6).Children(func() {
				if control.Button(c, "Shuffle", control.ButtonProps{Variant: control.Outline, Size: control.Small, Label: "Randomize design"}).Grow(1).Clicked() {
					s.randomSequence++
					s.Randomize(uint64(c.Now().UnixNano()) ^ s.randomSequence)
				}
				if control.Button(c, "Reset", control.ButtonProps{Variant: control.Ghost, Size: control.Small}).Clicked() {
					s.Reset()
				}
			})
			anchor := control.Button(c, "Inspect theme", control.ButtonProps{Variant: control.Ghost, Size: control.Small}).FillWidth()
			s.inspector(c, anchor, preview)
			if s.status != "" {
				ui.Text(c, s.status).FontSize(11).TextColor(t.MutedForeground).Role(ui.RoleStatus)
			}
		})
	})
}

func choiceLabel(value string) string {
	switch value {
	case "system":
		return "System"
	case "flyout":
		return "Shadcn Flyout"
	case "native":
		return "System native"
	case "default-translucent":
		return "Default translucent"
	case "inverted-translucent":
		return "Inverted translucent"
	case "same":
		return "Same as body"
	case theme.FontInherit:
		return "Inherit"
	case "serif":
		return "Serif"
	case "mono":
		return "Monospace"
	}
	for _, font := range galleryFonts {
		if value == font.ID {
			return font.Name
		}
	}
	if len(value) > 0 {
		return strings.ToUpper(value[:1]) + value[1:]
	}
	return value
}
func (s *State) choice(c *ui.Context, label, value string, values []string, changed func(string)) {
	t := theme.TokensOf(c)
	root := ui.Column(c.Key(label)).FillWidth().Gap(6)
	state := ui.Local(root, "value", func() creatorChoice { return creatorChoice{Value: value, Last: value} })
	if state.Last != value {
		state.Value, state.Last = value, value
	}
	root.Children(func() {
		ui.Text(c, label).FontSize(11).FontWeight(500).TextColor(t.MutedForeground)
		options := make([]control.Option, len(values))
		for i, v := range values {
			options[i] = control.Option{Value: v, Label: choiceLabel(v)}
		}
		props := control.SelectProps{Label: label, Options: options, Size: control.Small}
		var field ui.Element
		if s.Config.NativeMenus {
			field = desktop.NativeSelect(c, &state.Value, props)
		} else {
			var err error
			field, err = desktop.Select(c, s.Popovers, "creator-choice:"+label, &state.Value, props)
			if err != nil {
				s.popupError(c, err)
			}
		}
		field.FillWidth().MinHeight(32).Height(32).FontSize(12)
		if field.Changed() {
			state.Last = state.Value
			changed(state.Value)
			c.Invalidate()
		}
	})
}

type creatorChoice struct{ Value, Last string }
type creatorSwitch struct{ Value, Last bool }

func (s *State) choiceSwitch(c *ui.Context, label string, value bool, changed func(bool)) {
	root := ui.Column(c.Key(label))
	state := ui.Local(root, "value", func() creatorSwitch { return creatorSwitch{Value: value, Last: value} })
	if state.Last != value {
		state.Value, state.Last = value, value
	}
	root.Children(func() {
		if control.Switch(c, &state.Value, control.SwitchProps{Label: label}).Changed() {
			state.Last = state.Value
			changed(state.Value)
			c.Invalidate()
		}
	})
}

func (s *State) preview(c *ui.Context, t *theme.Tokens) {
	theme.MenuScope(c, t, s.Config.MenuOptions()).Grow(1).MinHeight(0).MinWidth(0).Radius(14).Border(1, t.Border).Clip().Background(t.Background).TextColor(t.Foreground).Children(func() {
		if s.fontError != "" {
			ui.Text(c, s.fontError).Role(ui.RoleAlert).TextColor(t.Destructive).Padding(12)
		}
		var parts ui.TabsParts
		triggers := make([]ui.Element, len(pages))
		ui.Row(c).Height(48).FillWidth().InsetShadow(0, -1, 0, 0, t.Border).Background(t.Card).Children(func() {
			ui.ScrollHorizontal(c).Grow(1).MinWidth(0).Children(func() {
				parts = ui.TabsBase(c, &s.previewIndex, len(pages))
				parts.List.Label("Component previews").Gap(2).PaddingX(12).Height(48).Children(func() {
					for i, page := range pages {
						b := parts.Tab(i).Height(48).PaddingX(11).FontSize(12).TextColor(t.MutedForeground)
						if i == s.previewIndex {
							b.TextColor(t.Foreground).InsetShadow(0, -2, 0, 0, t.Primary)
						}
						triggers[i] = b
						b.Children(func() { ui.Text(c, page.Label).SingleLine() })
					}
				})
			})
		})
		s.previewIndex = max(0, min(s.previewIndex, len(pages)-1))
		s.Page = pages[s.previewIndex].Key
		triggers[s.previewIndex].Bind(&s.toastPageReturn)
		width, _ := c.Size()
		if s.Page != s.lastPreviewPage || width != s.lastPreviewWidth {
			triggers[s.previewIndex].ScrollIntoView()
			s.lastPreviewPage = s.Page
			s.lastPreviewWidth = width
		}
		panel := ui.Scroll(c.Key(s.Page)).Grow(1).MinHeight(0).FillWidth().Role(ui.RoleRegion).Label(pages[s.previewIndex].Label)
		panel.LabelledBy(triggers[s.previewIndex])
		triggers[s.previewIndex].Controls(panel)
		panel.Children(func() {
			ui.Column(c).FillWidth().Padding(24).Gap(20).Children(func() {
				switch s.Page {
				case "controls":
					s.controls(c)
				case "display":
					s.displays(c)
				case "content":
					s.contents(c)
				case "text-layout":
					s.textLayout(c)
				case "overlays":
					s.overlays(c)
				case "conversation":
					s.conversation(c)
				case "navigation":
					s.navigationPreview(c)
				case "themes":
					s.themes(c)
				default:
					s.overview(c)
				}
			})
		})
		s.previewDialogs(c)
	})
}

// Each component name is an alternative way to find its page. Scoring a long
// concatenation lets unrelated names jointly outrank an actual page title.
func creatorPageScore(item overlay.CommandItem, query string) float64 {
	score := max(overlay.CommandScore(item.Label, query), overlay.CommandScore(item.Value, query))
	for _, keyword := range item.Keywords {
		score = max(score, overlay.CommandScore(keyword, query))
	}
	return score
}

func (s *State) creatorDialogs(c *ui.Context) {
	commands := make([]overlay.CommandItem, len(pages))
	for i, page := range pages {
		commands[i] = overlay.CommandItem{Key: page.Key, Value: page.Key, Label: page.Label, Keywords: page.Keywords}
	}
	if choice := overlay.CommandDialog(c, &s.Command, &s.Query, overlay.CommandProps{Items: commands, Filter: creatorPageScore, Placeholder: "Find a preview…"}); choice != "" {
		s.Page = choice
		s.pendingPreview = componentSearchTarget(choice, s.Query)
		s.Command = false
		c.Invalidate()
	}
	overlay.Dialog(c, &s.importOpen, overlay.DialogProps{Title: "Open preset", Description: "Paste a Native Go design preset. Changes apply only after validation."}, func() {
		control.Textarea(c, &s.importText, control.TextareaProps{Label: "Preset JSON", Placeholder: DefaultPreset().JSON(), MinRows: 8, MaxRows: 12, Invalid: s.importError != ""}).FillWidth().Font("monospace")
		if s.importError != "" {
			ui.Text(c, s.importError).TextColor(theme.TokensOf(c).Destructive).FontSize(12).Role(ui.RoleAlert)
		}
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			if control.Button(c, "Paste clipboard", control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
				s.importText = c.ReadClipboard()
			}
			if control.Button(c, "Apply preset", control.ButtonProps{Size: control.Small}).Clicked() {
				if err := s.ImportPreset(s.importText); err != nil {
					s.importError = err.Error()
				} else {
					s.importOpen = false
					s.importError = ""
					s.status = "Preset applied"
				}
			}
		})
	})
}
func (s *State) previewDialogs(c *ui.Context) {
	if s.pendingMenuAction != "" {
		overlay.Toast(c, ui.Toast{Title: "Menu action", Description: menuActionLabel(s.pendingMenuAction)})
		s.pendingMenuAction = ""
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
	s.notificationViewport(c)
}

func (s *State) inspector(c *ui.Context, anchor ui.Element, t *theme.Tokens) {
	if s.Popovers == nil {
		if anchor.Clicked() {
			s.popupError(c, fmt.Errorf("this preview has no desktop window"))
		}
		return
	}
	revision := s.designRevision
	if t.Dark {
		revision ^= 1 << 63
	}
	p := desktop.PopoverProps{Label: "Theme tokens", Placement: mygo.PlacementRightStart, Width: 360, Height: 420, FitContent: true, Revision: revision}
	var err error
	if s.Config.NativeMenus {
		err = desktop.NativePopover(c, s.Popovers, "theme-inspector", anchor, &s.tokenOpen, p, func(child *ui.Context) { s.nativeInspectorContent(child, t) })
	} else {
		err = desktop.Popover(c, s.Popovers, "theme-inspector", anchor, &s.tokenOpen, p, func(child *ui.Context) { s.inspectorContent(child, t) })
	}
	if err != nil {
		s.tokenOpen = false
		s.popupError(c, err)
	}
}
func (s *State) popupError(c *ui.Context, err error) {
	overlay.Toast(c, ui.Toast{Title: "Could not open popup", Description: err.Error(), Type: "error"})
}
func (s *State) nativeInspectorContent(c *ui.Context, t *theme.Tokens) {
	ui.Column(c).FillWidth().Gap(12).Children(func() {
		ui.Text(c, "Theme tokens").FontSize(16).FontWeight(600)
		for _, v := range []struct {
			name  string
			color ui.Color
		}{{"Background", t.Background}, {"Foreground", t.Foreground}, {"Primary", t.Primary}, {"Muted", t.Muted}, {"Border", t.Border}, {"Chart 1", t.Chart1}} {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(24, 24).Radius(5).Background(v.color)
				ui.Text(c, v.name).Grow(1)
				ui.Text(c, fmt.Sprintf("#%02x%02x%02x", v.color.R, v.color.G, v.color.B)).Font("monospace").FontSize(11)
			})
		}
		if ui.Button(c, "Copy preset").Clicked() {
			c.WriteClipboard(s.ExportPreset())
		}
	})
}

func (s *State) inspectorContent(c *ui.Context, t *theme.Tokens) {
	theme.FontScope(c, s.previewFonts()).FillWidth().Children(func() {
		theme.MenuScope(c, t, s.Config.MenuOptions()).FillWidth().Padding(16).Gap(12).Background(t.Background).TextColor(t.Foreground).Children(func() {
			s.cardHeader(c, "Theme tokens", "Your current preview colors.")
			for _, v := range []struct {
				name  string
				color ui.Color
			}{{"Background", t.Background}, {"Foreground", t.Foreground}, {"Primary", t.Primary}, {"Muted", t.Muted}, {"Border", t.Border}, {"Chart 1", t.Chart1}} {
				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(24, 24).Radius(5).Background(v.color).Border(1, t.Border)
					ui.Text(c, v.name).Grow(1).FontSize(12)
					ui.Text(c, fmt.Sprintf("#%02x%02x%02x", v.color.R, v.color.G, v.color.B)).Font("monospace").FontSize(11).TextColor(t.MutedForeground)
				})
			}
			if control.Button(c, "Copy preset", control.ButtonProps{Variant: control.Outline, Size: control.Small}).FillWidth().Clicked() {
				c.WriteClipboard(s.ExportPreset())
			}
		})
	})
}
