// Select source recipes using a shared no-editor listbox runtime.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var emptyChoice = "none"
var fruits = []control.Option{
	{Value: "none", Label: "Select a fruit"},
	{Value: "apple", Label: "Apple"}, {Value: "banana", Label: "Banana"},
	{Value: "blueberry", Label: "Blueberry"}, {Value: "grapes", Label: "Grapes"},
	{Value: "pineapple", Label: "Pineapple"},
}

type plan struct{ name, description string }

var plans = []plan{{"Starter", "Perfect for individuals getting started."}, {"Professional", "Ideal for growing teams and businesses."}, {"Enterprise", "Advanced features for large organizations."}}

type appState struct {
	theme                                                               theme.Controller
	selection, manual, dialogSelection                                  control.SelectState
	recipe, side, native, menuColor                                     string
	dark, rtl, small, disabled, invalid, icons, aligned, inline, dialog bool
	submitted                                                           string
}

func newState() *appState {
	return &appState{theme: theme.Controller{Preset: "blue", Mode: theme.LightMode}, recipe: "basic", side: "bottom", aligned: true, menuColor: "default", manual: control.SelectState{Values: []string{"review"}}}
}
func planVisual(c *ui.Context, p plan) {
	t := theme.TokensOf(c)
	ui.Column(c).MinWidth(0).Gap(2).Children(func() {
		ui.Text(c, p.name).FontSize(14).FontWeight(500)
		ui.Text(c, p.description).FontSize(12).TextColor(t.MutedForeground)
	})
}
func (s *appState) options() []control.Option {
	switch s.recipe {
	case "groups":
		return []control.Option{{Value: "apple", Label: "Apple", Group: "Fruits"}, {Value: "banana", Label: "Banana", Group: "Fruits"}, {Separator: true}, {Value: "carrot", Label: "Carrot", Group: "Vegetables"}, {Value: "broccoli", Label: "Broccoli", Group: "Vegetables"}}
	case "large":
		items := []control.Option{{Value: emptyChoice, Label: "Select an item"}}
		for i := range 100 {
			items = append(items, control.Option{Value: fmt.Sprint(i), Label: fmt.Sprintf("Item %02d", i)})
		}
		return items
	case "plan":
		items := []control.Option{}
		for _, p := range plans {
			items = append(items, control.Option{Value: p.name, Label: p.name})
		}
		return items
	case "multiple":
		return fruits[1:]
	case "aligned":
		items := append([]control.Option(nil), fruits...)
		items[4].Disabled = true
		return items
	default:
		return fruits
	}
}
func (s *appState) view(c *ui.Context) {
	s.theme.Mode = theme.LightMode
	if s.dark {
		s.theme.Mode = theme.DarkMode
	}
	s.theme.MenuColor = theme.MenuColor(s.menuColor)
	s.theme.Apply(c)
	if s.rtl {
		c.Root().Direction(ui.RightToLeft)
	} else {
		c.Root().Direction(ui.LeftToRight)
	}
	t := theme.TokensOf(c)
	ui.Scroll(c).Fill().Background(t.Background).Children(func() {
		ui.Column(c).FillWidth().MaxWidth(1100).MarginX(ui.Auto).Padding(24).Gap(18).Children(func() {
			display.Typography(c, "Select composition", display.H2)
			ui.Text(c, "Choose one or several values, place the list beside its trigger, and use the keyboard through a long collection.").TextColor(t.MutedForeground)
			ui.Row(c).Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
				control.Switch(c, &s.dark, control.SwitchProps{Label: "Dark appearance"})
				control.Switch(c, &s.rtl, control.SwitchProps{Label: "Right to left"})
				control.NativeSelect(c, &s.menuColor, control.SelectProps{Label: "Menu appearance", Options: []control.Option{{Value: "default", Label: "Default"}, {Value: "inverted", Label: "Inverted"}, {Value: "default-translucent", Label: "Translucent"}, {Value: "inverted-translucent", Label: "Inverted translucent"}}}).Width(200)
			})
			display.Card(c, func() {
				ui.Column(c).Gap(14).Children(func() {
					display.Typography(c, "Recipe workbench", display.H3)
					ui.Row(c).Wrap().Gap(12).AlignItems(ui.Center).Children(func() {
						if control.NativeSelect(c, &s.recipe, control.SelectProps{Label: "Recipe", Options: []control.Option{{Value: "basic", Label: "Basic"}, {Value: "groups", Label: "Groups"}, {Value: "large", Label: "Large list"}, {Value: "multiple", Label: "Multiple"}, {Value: "plan", Label: "Subscription plan"}, {Value: "aligned", Label: "Aligned disabled item"}}}).Width(200).Changed() {
							s.selection = control.SelectState{}
							if s.recipe == "plan" {
								s.selection.Value = control.SelectSelection{Value: "Starter", Valid: true}
							}
							if s.recipe == "aligned" {
								s.selection.Value = control.SelectSelection{Value: "banana", Valid: true}
								s.aligned = true
							}
						}
						if control.NativeSelect(c, &s.side, control.SelectProps{Label: "List side", Options: []control.Option{{Value: "top", Label: "Top"}, {Value: "bottom", Label: "Bottom"}, {Value: "left", Label: "Left"}, {Value: "right", Label: "Right"}, {Value: "inline-start", Label: "Inline start"}, {Value: "inline-end", Label: "Inline end"}}}).Width(160).Changed() {
							s.selection.Open = false
						}
						control.Switch(c, &s.aligned, control.SwitchProps{Label: "Align item"})
						control.Switch(c, &s.small, control.SwitchProps{Label: "Small trigger"})
					})
					ui.Row(c).Wrap().Gap(16).Children(func() {
						control.Checkbox(c, &s.inline, "Inline field", control.CheckboxProps{})
						control.Checkbox(c, &s.icons, "Icons", control.CheckboxProps{})
						control.Checkbox(c, &s.disabled, "Disabled", control.CheckboxProps{})
						control.Checkbox(c, &s.invalid, "Invalid", control.CheckboxProps{})
					})
					p := control.SelectWithProps{Root: control.SelectRootProps{Options: s.options(), Multiple: s.recipe == "multiple", Disabled: s.disabled}, Trigger: control.SelectTriggerProps{Label: "Recipe selection", Invalid: s.invalid}, Content: control.SelectContentProps{Side: control.SelectSide(s.side), AlignItemWithTrigger: &s.aligned}, Placeholder: "Select fruits"}
					if s.small {
						p.Trigger.Size = control.Small
					}
					if s.recipe == "basic" || s.recipe == "large" || s.recipe == "aligned" {
						p.Root.UnselectedValue = &emptyChoice
					}
					if s.recipe == "multiple" {
						p.ValueContent = func(c *ui.Context, v control.SelectValueView) {
							text := "Select fruits"
							if len(v.Values) == 1 {
								for _, o := range fruits {
									if o.Value == v.Values[0] {
										text = o.Label
									}
								}
							} else if len(v.Values) > 1 {
								text = fmt.Sprintf("%d fruits selected", len(v.Values))
							}
							ui.Text(c, text)
						}
					}
					if s.recipe == "plan" {
						p.ItemContent = func(c *ui.Context, o control.Option) {
							for _, p := range plans {
								if p.name == o.Value {
									planVisual(c, p)
								}
							}
						}
						p.ValueContent = func(c *ui.Context, v control.SelectValueView) {
							for _, p := range plans {
								if v.Value.Valid && p.name == v.Value.Value {
									planVisual(c, p)
									return
								}
							}
							ui.Text(c, "Choose a plan")
						}
					} else if s.icons {
						p.ItemContent = func(c *ui.Context, o control.Option) {
							ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() { ui.Icon(c, outlined.Favorite()).Size(16, 16).AccessHidden(true); ui.Text(c, o.Label) })
						}
						if s.recipe != "multiple" {
							p.ValueContent = func(c *ui.Context, v control.SelectValueView) {
								ui.Icon(c, outlined.Favorite()).Size(16, 16).AccessHidden(true)
								label := "Select a fruit"
								for _, o := range v.Options {
									if v.Value.Valid && o.Value == v.Value.Value {
										label = o.Label
									}
								}
								ui.Text(c, label)
							}
						}
					}
					orientation := control.FieldVertical
					if s.inline {
						orientation = control.FieldHorizontal
					}
					control.Field(c, control.FieldProps{Label: "Your selection", Orientation: orientation, Description: "Use arrows, Home/End, Enter and Escape. Type a label while closed to choose it.", Error: map[bool]string{true: "Choose an available value."}[s.invalid]}, func() ui.Element {
						var trigger ui.Element
						ui.Row(c).Wrap().Gap(12).AlignItems(ui.Center).Children(func() {
							trigger = control.SelectWith(c, &s.selection, p).Width(320)
							if control.Button(c, "Submit selection", control.ButtonProps{Variant: control.Outline}).Clicked() {
								s.submit()
							}
						})
						return trigger
					})
					ui.Text(c, s.submitted).Role(ui.RoleStatus)

				})
			})
			s.compound(c)
			display.Card(c, func() {
				ui.Column(c).Gap(12).Children(func() {
					display.Typography(c, "Dialog and system menu", display.H3)
					ui.Row(c).Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
						if control.Button(c, "Choose in dialog", control.ButtonProps{Variant: control.Outline}).Clicked() {
							s.dialog = true
						}
						control.NativeSelect(c, &s.native, control.SelectProps{Label: "System fruit", Placeholder: "System menu", Options: fruits[1:]}).Width(220)
					})
				})
			})
		})
	})
	overlay.Dialog(c, &s.dialog, overlay.DialogProps{Title: "Choose a fruit", Description: "The list keeps its keyboard boundary inside this dialog.", Width: 440}, func() {
		control.SelectWith(c, &s.dialogSelection, control.SelectWithProps{Root: control.SelectRootProps{Options: fruits, UnselectedValue: &emptyChoice}, Trigger: control.SelectTriggerProps{Label: "Dialog fruit"}}).FillWidth()
		if control.Button(c, "Done", control.ButtonProps{}).Clicked() {
			s.dialog = false
		}
	})
}
func (s *appState) submit() {
	values := s.selection.Values
	if s.recipe != "multiple" {
		values = nil
		if s.selection.Value.Valid {
			values = []string{s.selection.Value.Value}
		}
	}
	s.submitted = "Submitted: " + strings.Join(values, ", ")
	if len(values) == 0 {
		s.submitted = "No values selected"
	}
}

func (s *appState) compound(c *ui.Context) {
	display.Card(c, func() {
		ui.Column(c).Gap(12).Children(func() {
			display.Typography(c, "Independent slots", display.H3)
			ui.Text(c, "A manually registered collection shares the same selection, focus, scrolling and labels.").TextColor(theme.TokensOf(c).MutedForeground)
			control.SelectRoot(c, &s.manual, control.SelectRootProps{Multiple: true}, func(b *control.SelectScope) {
				b.Trigger(c, control.SelectTriggerProps{Label: "Workflow stages"}, func() {
					b.Value(c, control.SelectValueProps{}, func(v control.SelectValueView) { ui.Text(c, fmt.Sprintf("%d stages selected", len(v.Values))) })
				}).Width(260)
				b.Content(c, control.SelectContentProps{}, func() {
					b.List(c, control.SelectListProps{Label: "Workflow stages"}, func() {
						b.Group(c, "Preparation", func() {
							b.Item(c, "draft", control.SelectItemProps{Label: "Draft"}, nil)
							b.Item(c, "review", control.SelectItemProps{Label: "Review"}, nil)
						})
						b.Separator(c)
						b.Group(c, "Delivery", func() {
							b.Item(c, "approved", control.SelectItemProps{Label: "Approved"}, nil)
							b.Item(c, "published", control.SelectItemProps{Label: "Published"}, nil)
						})
					})
					b.ScrollUpButton(c)
					b.ScrollDownButton(c)
				})
			})
		})
	})
}
func main() {
	snapshot := flag.String("snapshot", "", "write a native PNG instead of opening a window")
	dark := flag.Bool("dark", false, "use dark appearance")
	open := flag.Bool("open", false, "initially show the selection list")
	flag.Parse()
	s := newState()
	s.dark = *dark
	s.selection.Open = *open
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		err = png.Encode(f, ui.Render(func(c *ui.Context) { c.DisableAnimations(); s.view(c) }, 1050, 1000, 1))
		closeErr := f.Close()
		if err != nil {
			log.Fatal(err)
		}
		if closeErr != nil {
			log.Fatal(closeErr)
		}
		return
	}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{Title: "Select composition", Width: 1050, Height: 1000, Content: ui.View(s.view)}).Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
