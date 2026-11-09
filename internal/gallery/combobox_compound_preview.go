package gallery

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type compoundDestination struct{ Code, Name, Region string }
type comboboxCompoundPreviewState struct {
	initialized                                                                                          bool
	side, align, highlight                                                                               string
	rtl, extraGap, iconAddon, disabled, invalid, readonly, noRemove, multiple, popup, dialog, separators bool
	layout, grid, virtual, modal                                                                         control.ComboboxState[string]
	destination                                                                                          control.ComboboxState[compoundDestination]
	scroll                                                                                               ui.ScrollState
	destinations                                                                                         []compoundDestination
	large                                                                                                []string
	email, submitted                                                                                     string
}

var compoundFrameworks = []string{"Next.js", "SvelteKit", "Nuxt.js", "Remix", "Astro"}

func (p *comboboxCompoundPreviewState) initialize() {
	if p.initialized {
		return
	}
	p.initialized = true
	p.side, p.align, p.highlight = "bottom", "start", "input"
	p.multiple, p.separators = true, true
	p.destinations = []compoundDestination{{"jp", "Japan", "Asia"}, {"sg", "Singapore", "Asia"}, {"de", "Germany", "Europe"}, {"fr", "France", "Europe"}, {"no", "Norway", "Europe"}}
	p.destination.Values = []compoundDestination{p.destinations[0]}
	for i := range 100 {
		p.large = append(p.large, fmt.Sprintf("Item %03d", i+1))
	}
}

// The source's 12 standalone examples and 20 full-collection examples share
// behavior. These three explorers add the independent slots and configurations
// missing from the existing basic/grouped/popup/multiple/state cards.
func (s *State) comboboxCompoundPreview(c *ui.Context) {
	p := &s.comboboxCompound
	p.initialize()
	display.Card(c.Key("combobox-content-explorer"), func() {
		s.cardHeader(c, "Content placement", "Choose a side, alignment and reading direction, then search the framework field.")
		ui.Row(c).FillWidth().Wrap().Gap(12).Children(func() {
			control.Select(c, &p.side, control.SelectProps{Label: "Content side", Options: []control.Option{{Value: "bottom", Label: "Bottom"}, {Value: "top", Label: "Top"}, {Value: "left", Label: "Left"}, {Value: "right", Label: "Right"}, {Value: "inline-start", Label: "Inline start"}, {Value: "inline-end", Label: "Inline end"}}}).Width(150)
			control.Select(c, &p.align, control.SelectProps{Label: "Content alignment", Options: []control.Option{{Value: "start", Label: "Start"}, {Value: "center", Label: "Center"}, {Value: "end", Label: "End"}}}).Width(130)
			control.Select(c, &p.highlight, control.SelectProps{Label: "Highlight mode", Options: []control.Option{{Value: "off", Label: "Manual"}, {Value: "input", Label: "On input"}, {Value: "always", Label: "Always"}}}).Width(150)
		})
		ui.Row(c).Wrap().Gap(16).Children(func() {
			control.Switch(c, &p.rtl, control.SwitchProps{Label: "Right to left", Size: control.Small})
			control.Switch(c, &p.extraGap, control.SwitchProps{Label: "12 DIP gap", Size: control.Small})
			control.Switch(c, &p.iconAddon, control.SwitchProps{Label: "Icon addon", Size: control.Small})
		})
		direction := ui.LeftToRight
		if p.rtl {
			direction = ui.RightToLeft
		}
		ui.Box(c).FillWidth().Direction(direction).Children(func() {
			auto := control.ComboboxHighlightOnInput
			switch p.highlight {
			case "off":
				auto = control.ComboboxHighlightOff
			case "always":
				auto = control.ComboboxHighlightAlways
			}
			control.ComboboxRoot(c, &p.layout, control.ComboboxRootProps[string]{Items: compoundFrameworks, Adapter: control.ComboboxStringAdapter(), AutoHighlight: auto}, func(b *control.ComboboxScope[string]) {
				input := control.ComboboxInputProps{Label: "Placement framework", Placeholder: "Select a framework", ShowClear: true}
				content := control.ComboboxContentProps{Side: control.ComboboxSide(p.side), Align: control.ComboboxAlign(p.align)}
				if p.extraGap {
					gap := float32(12)
					content.SideOffset = &gap
				}
				if p.iconAddon {
					input.Addons.InlineStart = func() { material.Icon(c, outlined.Globe()).Size(16, 16) }
					content.AlignOffset = -28
					content.Width = 240
				}
				b.Input(c, input)
				b.Content(c, content, func() { compoundStringList(c, b, "No frameworks found.") })
			})
		})
	}).Role(ui.RoleGroup).Label("Combobox content explorer")
	display.Card(c.Key("combobox-object-explorer"), func() {
		s.cardHeader(c, "Destination form", "Combine object values, custom rows and chips with an ordinary form.")
		ui.Row(c).Wrap().Gap(16).Children(func() {
			control.Switch(c, &p.multiple, control.SwitchProps{Label: "Multiple destinations", Size: control.Small})
			control.Switch(c, &p.popup, control.SwitchProps{Label: "Search in dialog popup", Size: control.Small})
			control.Switch(c, &p.separators, control.SwitchProps{Label: "Group separators", Size: control.Small})
			control.Switch(c, &p.noRemove, control.SwitchProps{Label: "Hide chip remove buttons", Size: control.Small})
			control.Switch(c, &p.disabled, control.SwitchProps{Label: "Disable destination field", Size: control.Small})
			control.Switch(c, &p.readonly, control.SwitchProps{Label: "Read only destination", Size: control.Small})
			control.Switch(c, &p.invalid, control.SwitchProps{Label: "Show destination error", Size: control.Small})
		})
		errorText := ""
		if p.invalid {
			errorText = "Choose a destination before continuing."
		}
		control.Field(c, control.FieldProps{Label: "Trip destinations", Error: errorText}, func() ui.Element {
			return p.destinationField(c)
		})
		control.Input(c, &p.email, control.InputProps{Label: "Trip contact email", Placeholder: "you@example.com"})
		if control.Button(c, "Submit trip", control.ButtonProps{}).Clicked() {
			p.submitTrip()
		}
		if p.submitted != "" {
			ui.Text(c, p.submitted).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
		}
		if control.Button(c, "Choose framework in dialog", control.ButtonProps{Variant: control.Outline}).Clicked() {
			p.dialog = true
		}
		overlay.Dialog(c, &p.dialog, overlay.DialogProps{Title: "Choose a framework", Description: "The suggestions stay within this dialog's focus scope.", Width: 425, Footer: func() {
			if control.Button(c, "Confirm framework", control.ButtonProps{}).Clicked() {
				p.dialog = false
			}
		}}, func() {
			control.ComboboxRoot(c, &p.modal, control.ComboboxRootProps[string]{Items: compoundFrameworks, Adapter: control.ComboboxStringAdapter()}, func(b *control.ComboboxScope[string]) {
				b.Input(c, control.ComboboxInputProps{Label: "Dialog framework", Placeholder: "Select a framework"})
				b.Content(c, control.ComboboxContentProps{}, func() { compoundStringList(c, b, "No frameworks found.") })
			})
		})
	}).Role(ui.RoleGroup).Label("Combobox destination form")
	columns(c, func() {
		display.Card(c.Key("combobox-grid-explorer"), func() {
			s.cardHeader(c, "Framework grid", "Use all four arrow keys to navigate the two-column suggestions.")
			control.ComboboxRoot(c, &p.grid, control.ComboboxRootProps[string]{Items: compoundFrameworks, Adapter: control.ComboboxStringAdapter(), Grid: true}, func(b *control.ComboboxScope[string]) {
				b.Input(c, control.ComboboxInputProps{Label: "Grid framework", Placeholder: "Search framework grid"})
				b.Content(c, control.ComboboxContentProps{}, func() {
					b.Empty(c, "No frameworks found.")
					b.List(c, control.ComboboxListProps{}, func() {
						items := b.FilteredItems()
						for first := 0; first < len(items); first += 2 {
							b.Row(c, first/2, func() {
								for _, value := range items[first:min(first+2, len(items))] {
									b.Item(c, value, control.ComboboxItemProps{}, nil)
								}
							})
						}
					})
				})
			})
		}).Role(ui.RoleGroup).Label("Combobox grid explorer")
	}, func() {
		display.Card(c.Key("combobox-virtual-explorer"), func() {
			s.cardHeader(c, "Large collection", "Search 100 items or use the arrow keys to reveal a choice in the scroll window.")
			control.ComboboxRoot(c, &p.virtual, control.ComboboxRootProps[string]{Items: p.large, Adapter: control.ComboboxStringAdapter(), Virtualized: true, Reveal: func(index int) { p.scroll.Y = float32(index * 32) }}, func(b *control.ComboboxScope[string]) {
				b.Input(c, control.ComboboxInputProps{Label: "Large collection search", Placeholder: "Search 100 items"})
				b.Content(c, control.ComboboxContentProps{}, func() {
					b.Empty(c, "No matching items.")
					b.List(c, control.ComboboxListProps{Scroll: &p.scroll}, func() {
						filtered := b.FilteredItems()
						first := min(max(0, int(p.scroll.Y/32)-1), max(0, len(filtered)-1))
						end := min(len(filtered), first+12)
						ui.Box(c).FillWidth().Height(float32(first * 32))
						for _, value := range filtered[first:end] {
							b.Item(c, value, control.ComboboxItemProps{}, nil)
						}
						ui.Box(c).FillWidth().Height(float32((len(filtered) - end) * 32))
					})
				})
			})
		}).Role(ui.RoleGroup).Label("Combobox large collection explorer")
	})
}
func compoundStringList(c *ui.Context, b *control.ComboboxScope[string], empty string) {
	b.Empty(c, empty)
	b.List(c, control.ComboboxListProps{}, func() {
		b.Collection(c, func(c *ui.Context, value string, _ int) { b.Item(c, value, control.ComboboxItemProps{}, nil) })
	})
}
func (p *comboboxCompoundPreviewState) destinationField(c *ui.Context) ui.Element {
	adapter := control.ComboboxAdapter[compoundDestination]{Key: func(v compoundDestination) string { return v.Code }, Label: func(v compoundDestination) string { return v.Name }, FormValue: func(v compoundDestination) string { return v.Code }, Equal: func(a, b compoundDestination) bool { return a.Code == b.Code }, Disabled: func(v compoundDestination) bool { return v.Code == "no" }}
	groups := []control.ComboboxGroupData[compoundDestination]{{Key: "asia", Label: "Asia", Items: p.destinations[:2]}, {Key: "europe", Label: "Europe", Items: p.destinations[2:]}}
	var root ui.Element
	control.ComboboxRoot(c, &p.destination, control.ComboboxRootProps[compoundDestination]{Groups: groups, Adapter: adapter, Multiple: p.multiple, Disabled: p.disabled, ReadOnly: p.readonly, AutoHighlight: control.ComboboxHighlightOnInput, Filter: func(v compoundDestination, q string) bool {
		return strings.Contains(strings.ToLower(v.Name+" "+v.Region), strings.ToLower(strings.TrimSpace(q)))
	}}, func(b *control.ComboboxScope[compoundDestination]) {
		content := control.ComboboxContentProps{}
		var anchor ui.Handle
		if p.popup {
			root = b.Trigger(c, control.ComboboxTriggerProps{Label: "Trip destinations", Button: control.ButtonProps{Variant: control.Outline, Invalid: p.invalid}}, func() { b.Value(c, control.ComboboxValueProps{Placeholder: "Select destinations"}) }).FillWidth()
		} else if p.multiple {
			root = b.Chips(c, control.ComboboxChipsProps{Label: "Trip destinations", Invalid: p.invalid}, func() {
				for _, value := range p.destination.Values {
					b.Chip(c, value, control.ComboboxChipProps{HideRemove: p.noRemove}, nil)
				}
				editor := b.ChipsInput(c, control.ComboboxInputProps{Label: "Trip destinations", Placeholder: "Add a destination"})
				editor.OnSubmit(p.submitTrip)
			})
			root.Bind(&anchor)
			content.Anchor = &anchor
		} else {
			parts := b.Input(c, control.ComboboxInputProps{Label: "Trip destinations", Placeholder: "Select a destination", Invalid: p.invalid, ShowClear: true})
			root = parts.Root
			parts.Editor.OnSubmit(p.submitTrip)
		}
		b.Content(c, content, func() {
			if p.popup {
				parts := b.Input(c, control.ComboboxInputProps{Label: "Trip destinations", Placeholder: "Search countries or regions", HideTrigger: true})
				parts.Editor.OnSubmit(p.submitTrip)
			}
			b.Empty(c, "No destinations found.")
			b.List(c, control.ComboboxListProps{}, func() {
				b.Groups(c, func(c *ui.Context, group control.ComboboxGroupData[compoundDestination], index int) {
					if index > 0 && p.separators {
						b.Separator(c)
					}
					b.Group(c, group, func() {
						b.Label(c, group.Label)
						b.Collection(c, func(c *ui.Context, value compoundDestination, _ int) {
							b.Item(c, value, control.ComboboxItemProps{}, func() {
								ui.Column(c).Gap(2).MinWidth(0).Children(func() {
									ui.Text(c, value.Name).FontSize(14).SingleLine()
									ui.Text(c, value.Region+" · "+strings.ToUpper(value.Code)).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
								})
							})
						})
					})
				})
			})
		})
	})
	return root
}
func (p *comboboxCompoundPreviewState) submitTrip() {
	names := []string{}
	if p.multiple {
		for _, value := range p.destination.Values {
			names = append(names, value.Name)
		}
	} else if p.destination.Value.Valid {
		names = append(names, p.destination.Value.Value.Name)
	}
	if len(names) == 0 {
		p.submitted = "Choose at least one destination."
		return
	}
	p.submitted = "Trip saved: " + strings.Join(names, ", ")
}
