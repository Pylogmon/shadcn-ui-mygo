package gallery

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Only values and public SelectState survive a build. Scopes, native Elements
// and view callbacks belong to the current parent or selector child context.
type selectCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	drafts                               map[string]*selectCompositionDraft
}
type selectCompositionDraft struct {
	choices          map[string]*control.SelectState
	align, dialog    bool
	query, sort      string
	changes, submits int
	submitted        []string
}

var selectCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Fruit group (docs)"}, {Value: "align", Label: "Align Item switch"},
	{Value: "groups", Label: "Fruit and vegetable groups"}, {Value: "scrollable", Label: "Timezone groups"},
	{Value: "disabled", Label: "Disabled (docs)"}, {Value: "invalid", Label: "Invalid Field (docs)"},
	{Value: "rtl", Label: "Translated groups"},
	{Value: "create-basic", Label: "Create: Basic"}, {Value: "create-sides", Label: "Create: Six sides"},
	{Value: "create-icons", Label: "Create: Chart icons"}, {Value: "create-groups", Label: "Create: Groups and labels"},
	{Value: "create-large", Label: "Create: 100 items"}, {Value: "create-multiple", Label: "Create: Multiple selection"},
	{Value: "create-sizes", Label: "Create: Two sizes"}, {Value: "create-plan", Label: "Create: Subscription Plan"},
	{Value: "create-button", Label: "Create: With Button"}, {Value: "create-aligned", Label: "Create: Item Aligned"},
	{Value: "create-field", Label: "Create: With Field"}, {Value: "create-invalid", Label: "Create: Two invalid layouts"},
	{Value: "create-inline", Label: "Create: Input, filter and NativeSelect"}, {Value: "create-disabled", Label: "Create: Disabled"},
	{Value: "create-dialog", Label: "Create: In Dialog"},
}

const selectSourceNone = "source-null"

func (p *selectCompositionPreviewState) initialize(native bool) {
	if !p.ready {
		p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
		p.drafts = map[string]*selectCompositionDraft{}
	}
}
func (p *selectCompositionPreviewState) draft() *selectCompositionDraft {
	if d := p.drafts[p.recipe]; d != nil {
		return d
	}
	d := &selectCompositionDraft{choices: map[string]*control.SelectState{}, align: true}
	p.drafts[p.recipe] = d
	return d
}
func (d *selectCompositionDraft) choice(key string) *control.SelectState {
	if s := d.choices[key]; s != nil {
		return s
	}
	s := &control.SelectState{}
	if key == "aligned-demo" {
		s.Value = control.SelectSelection{Value: "banana", Valid: true}
	}
	if key == "plan" {
		s.Value = control.SelectSelection{Value: "Starter", Valid: true}
	}
	d.choices[key] = s
	return s
}
func (p *selectCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}
func (p *selectCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-select-config")
			h.Close("source-select-language")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Select compositions"}, display.CardSlots{
		Title: "Select compositions", Description: "Source listboxes, grouped choices, rich plans and independent selections.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-select-config", &p.recipe, control.SelectProps{Label: "Select configuration", Options: selectCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-select-language", &p.language, control.SelectProps{Label: "Select language", Options: statusLanguages()})
			}
			box := ui.Column(child.Key("source-select:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
			if p.recipe == "rtl" {
				box.Direction(childDirection(p.language))
			}
			box.Children(func() { p.draw(child, p.draft()) })
			if p.action != "" {
				ui.Text(child, p.action).FontSize(13).Role(ui.RoleStatus).Label(p.action)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).Role(ui.RoleAlert)
			}
		},
	})
}

func selectSourceFruits(count int, includeEmpty, disabledGrapes bool) []control.Option {
	all := []control.Option{{Value: "apple", Label: "Apple"}, {Value: "banana", Label: "Banana"}, {Value: "blueberry", Label: "Blueberry"}, {Value: "grapes", Label: "Grapes", Disabled: disabledGrapes}, {Value: "pineapple", Label: "Pineapple"}, {Value: "strawberry", Label: "Strawberry"}, {Value: "watermelon", Label: "Watermelon"}}
	items := slices.Clone(all[:count])
	if includeEmpty {
		items = append([]control.Option{{Value: selectSourceNone, Label: "Select a fruit"}}, items...)
	}
	return items
}
func selectSourceGroups(language string, full bool) (string, []control.Option) {
	words := []string{"Select a fruit", "Fruits", "Apple", "Banana", "Blueberry", "Grapes", "Pineapple", "Vegetables", "Carrot", "Broccoli", "Spinach"}
	if language == "ar" {
		words = []string{"اختر فاكهة", "الفواكه", "تفاح", "موز", "توت أزرق", "عنب", "أناناس", "الخضروات", "جزر", "بروكلي", "سبانخ"}
	} else if language == "he" {
		words = []string{"בחר פרי", "פירות", "תפוח", "בננה", "אוכמניה", "ענבים", "אננס", "ירקות", "גזר", "ברוקולי", "תרד"}
	}
	items := []control.Option{}
	keys := []string{"apple", "banana", "blueberry", "grapes", "pineapple"}
	count := 3
	if full {
		count = 5
	}
	for i := range count {
		items = append(items, control.Option{Value: keys[i], Label: words[i+2], Group: words[1]})
	}
	items = append(items, control.Option{Separator: true})
	for i, key := range []string{"carrot", "broccoli", "spinach"} {
		items = append(items, control.Option{Value: key, Label: words[i+8], Group: words[7]})
	}
	return words[0], items
}

var selectSourceTimezones = []control.Option{
	{Value: "est", Label: "Eastern Standard Time", Group: "North America"}, {Value: "cst", Label: "Central Standard Time", Group: "North America"}, {Value: "mst", Label: "Mountain Standard Time", Group: "North America"}, {Value: "pst", Label: "Pacific Standard Time", Group: "North America"}, {Value: "akst", Label: "Alaska Standard Time", Group: "North America"}, {Value: "hst", Label: "Hawaii Standard Time", Group: "North America"},
	{Value: "gmt", Label: "Greenwich Mean Time", Group: "Europe & Africa"}, {Value: "cet", Label: "Central European Time", Group: "Europe & Africa"}, {Value: "eet", Label: "Eastern European Time", Group: "Europe & Africa"}, {Value: "west", Label: "Western European Summer Time", Group: "Europe & Africa"}, {Value: "cat", Label: "Central Africa Time", Group: "Europe & Africa"}, {Value: "eat", Label: "East Africa Time", Group: "Europe & Africa"},
	{Value: "msk", Label: "Moscow Time", Group: "Asia"}, {Value: "ist", Label: "India Standard Time", Group: "Asia"}, {Value: "cst_china", Label: "China Standard Time", Group: "Asia"}, {Value: "jst", Label: "Japan Standard Time", Group: "Asia"}, {Value: "kst", Label: "Korea Standard Time", Group: "Asia"}, {Value: "ist_indonesia", Label: "Indonesia Central Standard Time", Group: "Asia"},
	{Value: "awst", Label: "Australian Western Standard Time", Group: "Australia & Pacific"}, {Value: "acst", Label: "Australian Central Standard Time", Group: "Australia & Pacific"}, {Value: "aest", Label: "Australian Eastern Standard Time", Group: "Australia & Pacific"}, {Value: "nzst", Label: "New Zealand Standard Time", Group: "Australia & Pacific"}, {Value: "fjt", Label: "Fiji Time", Group: "Australia & Pacific"},
	{Value: "art", Label: "Argentina Time", Group: "South America"}, {Value: "bot", Label: "Bolivia Time", Group: "South America"}, {Value: "brt", Label: "Brasilia Time", Group: "South America"}, {Value: "clt", Label: "Chile Standard Time", Group: "South America"},
}

type selectSourcePlan struct{ name, description string }

var selectSourcePlans = []selectSourcePlan{{"Starter", "Perfect for individuals getting started."}, {"Professional", "Ideal for growing teams and businesses."}, {"Enterprise", "Advanced features for large organizations."}}

type selectSourceProps struct {
	key, label, placeholder string
	options                 []control.Option
	width                   float32
	size                    control.Size
	disabled, invalid       bool
	icons, plan, multiple   bool
	align                   *bool
	side                    control.SelectSide
}

func selectSourcePlanVisual(c *ui.Context, key string) {
	for _, plan := range selectSourcePlans {
		if plan.name == key {
			ui.Column(c).FillWidth().MinWidth(0).Gap(0).Children(func() {
				ui.Text(c, plan.name).FontSize(14).FontWeight(500).LineHeight(20.0 / 14)
				ui.Text(c, plan.description).FontSize(12).FontWeight(400).LineHeight(4.0 / 3).TextColor(theme.TokensOf(c).MutedForeground)
			})
		}
	}
}
func selectSourceIcon(c *ui.Context, key string) {
	icon := outlined.ShowChart()
	switch key {
	case "bar":
		icon = outlined.BarChart()
	case "pie":
		icon = outlined.PieChart()
	}
	ui.Icon(c, icon).Size(16, 16).AccessHidden(true)
}
func (p *selectCompositionPreviewState) selectView(c *ui.Context, d *selectCompositionDraft, props selectSourceProps) ui.Element {
	state := d.choice(props.key)
	var trigger ui.Element
	control.SelectRoot(c.Key(props.key), state, control.SelectRootProps{Options: props.options, Disabled: props.disabled, Multiple: props.multiple, UnselectedValue: stringPointer(selectSourceNone),
		OnValueChange: func(next control.SelectSelection, _ *control.SelectChange) {
			if next != state.Value {
				d.changes++
				p.action = "Local select choice: " + next.Value
			}
		}, OnValuesChange: func(next []string, _ *control.SelectChange) {
			if !slices.Equal(next, state.Values) {
				d.changes++
				p.action = "Local select choices: " + strings.Join(next, ", ")
			}
		}}, func(b *control.SelectScope) {
		trigger = b.Trigger(c, control.SelectTriggerProps{Label: props.label, Size: props.size, Invalid: props.invalid}, func() {
			var content func(control.SelectValueView)
			if props.plan || props.icons || props.multiple {
				content = func(v control.SelectValueView) {
					if props.plan {
						selectSourcePlanVisual(c, v.Value.Value)
						return
					}
					text, key := props.placeholder, selectSourceNone
					if props.multiple {
						if len(v.Values) > 1 {
							text = fmt.Sprintf("%d fruits selected", len(v.Values))
						} else if len(v.Values) == 1 {
							key = v.Values[0]
						}
					} else if v.Value.Valid {
						key = v.Value.Value
					}
					for _, o := range props.options {
						if o.Value == key && !o.Separator && (key != selectSourceNone || !props.multiple) {
							text = o.Label
						}
					}
					if props.icons {
						selectSourceIcon(c, key)
					}
					ui.Text(c, text).SingleLine()
				}
			}
			b.Value(c, control.SelectValueProps{Placeholder: props.placeholder}, content)
		}).MinWidth(0)
		if props.width > 0 {
			trigger.Width(props.width).MaxWidthPercent(100)
		} else {
			trigger.FillWidth()
		}
		if props.plan {
			trigger.MinHeight(0).PaddingY(8)
		}
		b.Content(c, control.SelectContentProps{Side: props.side, AlignItemWithTrigger: props.align}, func() {
			b.List(c, control.SelectListProps{Label: props.label + " options"}, func() {
				for i := 0; i < len(props.options); {
					if props.options[i].Separator {
						b.Separator(c)
						i++
						continue
					}
					start, group := i, props.options[i].Group
					for i < len(props.options) && !props.options[i].Separator && props.options[i].Group == group {
						i++
					}
					b.Group(c.Key(start), group, func() {
						for _, o := range props.options[start:i] {
							var content func()
							if props.plan {
								content = func() { selectSourcePlanVisual(c, o.Value) }
							} else if props.icons {
								content = func() { selectSourceIcon(c, o.Value); ui.Text(c, o.Label).SingleLine() }
							}
							b.Item(c, o.Value, control.SelectItemProps{}, content)
						}
					})
				}
			})
			b.ScrollUpButton(c)
			b.ScrollDownButton(c)
		})
	})
	return trigger
}
func stringPointer(v string) *string { return &v }

func (p *selectCompositionPreviewState) field(c *ui.Context, d *selectCompositionDraft, label, description, errorText string, props selectSourceProps) {
	control.FieldWith(c.Key(props.key+":field"), control.FieldProps{Invalid: props.invalid, Error: errorText}, func(f *control.FieldScope) {
		f.Label(c, control.FieldLabelProps{Text: label}, nil)
		f.Control(c, func() ui.Element { return p.selectView(c, d, props) })
		if description != "" {
			f.Description(c, description)
		}
		if errorText != "" {
			f.Error(c)
		}
	})
}
func (p *selectCompositionPreviewState) draw(c *ui.Context, d *selectCompositionDraft) {
	width := float32(384)
	switch p.recipe {
	case "demo", "groups", "disabled", "invalid":
		width = 192
	case "scrollable":
		width = 256
	case "align":
		width = 320
	case "rtl":
		width = 128
	case "create-plan", "create-multiple":
		width = 288
	case "create-sides", "create-inline":
		width = 560
	}
	gap := float32(16)
	if p.recipe == "align" {
		gap = 28
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(width).Gap(gap).Children(func() {
		base := selectSourceProps{key: "main", label: "Source fruit", placeholder: "Select a fruit", options: selectSourceFruits(5, true, false)}
		switch p.recipe {
		case "demo":
			for i := range base.options {
				base.options[i].Group = "Fruits"
			}
			p.selectView(c, d, base)
		case "align":
			control.FieldWith(c, control.FieldProps{Orientation: control.FieldHorizontal}, func(f *control.FieldScope) {
				f.Content(c, func() {
					f.Label(c, control.FieldLabelProps{Text: "Align Item"}, nil)
					f.Description(c, "Toggle to align the item with the trigger.")
				})
				f.Control(c, func() ui.Element { return control.Switch(c, &d.align, control.SwitchProps{Label: "Align Item"}) })
			})
			base.key, base.align = "aligned-demo", &d.align
			p.selectView(c, d, base)
		case "groups", "create-groups", "rtl":
			language := "en"
			if p.recipe == "rtl" {
				language = p.language
			}
			base.placeholder, base.options = selectSourceGroups(language, p.recipe == "rtl")
			p.selectView(c, d, base)
		case "scrollable":
			base.label, base.placeholder, base.options = "Source timezone", "Select a timezone", selectSourceTimezones
			p.selectView(c, d, base)
		case "disabled", "create-disabled":
			base.disabled, base.options = true, selectSourceFruits(5, true, true)
			p.selectView(c, d, base)
		case "invalid":
			base.label, base.invalid, base.options = "Fruit", true, selectSourceFruits(3, true, false)
			p.field(c, d, "Fruit", "", "Please select a fruit.", base)
		case "create-basic":
			p.selectView(c, d, base)
		case "create-aligned":
			base.options, base.align = selectSourceFruits(5, true, true), &d.align
			p.selectView(c, d, base)
		case "create-sizes", "create-icons", "create-button":
			for i, size := range []control.Size{control.Small, control.DefaultSize} {
				base.key, base.label, base.size = fmt.Sprint(i), []string{"Small fruit", "Default fruit"}[i], size
				base.options = selectSourceFruits(3, true, false)
				if p.recipe == "create-icons" {
					base.label, base.placeholder, base.icons = []string{"Small chart type", "Default chart type"}[i], "Chart Type", true
					base.options = []control.Option{{Value: selectSourceNone, Label: "Chart Type"}, {Value: "line", Label: "Line"}, {Value: "bar", Label: "Bar"}, {Value: "pie", Label: "Pie"}}
				}
				if p.recipe == "create-button" {
					ui.Row(c.Key(i)).FillWidth().MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
						p.selectView(c, d, base).Grow(1)
						if control.Button(c, "Submit", control.ButtonProps{Variant: control.Outline, Size: size, Label: []string{"Submit small fruit", "Submit default fruit"}[i]}).Clicked() {
							d.submits++
							d.submitted = nil
							if v := d.choice(base.key).Value; v.Valid {
								d.submitted = []string{v.Value}
							}
							p.action = "Local select action: Submit"
						}
					})
				} else {
					p.selectView(c, d, base)
				}
			}
		case "create-sides":
			ui.Row(c).Wrap().FillWidth().Gap(8).Justify(ui.Center).Children(func() {
				aligned := false
				for _, side := range []control.SelectSide{control.SelectSideInlineStart, control.SelectSideLeft, control.SelectSideTop, control.SelectSideBottom, control.SelectSideRight, control.SelectSideInlineEnd} {
					base.key, base.side, base.align, base.width = string(side), side, &aligned, 112
					base.label, base.placeholder, base.options = "Side "+string(side), strings.Title(strings.ReplaceAll(string(side), "-", " ")), selectSourceFruits(3, true, false)
					base.options[0].Label = "Select"
					p.selectView(c, d, base)
				}
			})
		case "create-large":
			base.placeholder, base.options = "Select an item", []control.Option{{Value: selectSourceNone, Label: "Select an item"}}
			for i := range 100 {
				base.options = append(base.options, control.Option{Value: fmt.Sprintf("item-%d", i), Label: fmt.Sprintf("Item %d", i)})
			}
			p.selectView(c, d, base)
		case "create-multiple":
			base.label, base.multiple, base.placeholder, base.options = "Source fruits", true, "Select fruits", selectSourceFruits(7, false, false)
			p.selectView(c, d, base)
		case "create-plan":
			base.key, base.label, base.plan, base.options = "plan", "Subscription Plan", true, nil
			for _, plan := range selectSourcePlans {
				base.options = append(base.options, control.Option{Value: plan.name, Label: plan.name})
			}
			p.selectView(c, d, base)
		case "create-field":
			base.label = "Favorite Fruit"
			p.field(c, d, "Favorite Fruit", "Choose your favorite fruit from the list.", "", base)
		case "create-invalid":
			base.invalid, base.label = true, "Invalid fruit"
			p.selectView(c, d, base)
			base.key, base.label = "field", "Favorite Fruit"
			p.field(c, d, "Favorite Fruit", "", "Please select a valid fruit.", base)
		case "create-inline":
			ui.ScrollHorizontal(c).FillWidth().MinWidth(0).Children(func() {
				ui.Row(c).MinWidth(328).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
					control.Input(c, &d.query, control.InputProps{Label: "Source inline search", Placeholder: "Search..."}).Grow(1).MinWidth(32)
					base.label, base.placeholder, base.width = "Source filter", "Filter", 140
					base.options = []control.Option{{Value: selectSourceNone, Label: "Filter"}, {Value: "all", Label: "All"}, {Value: "active", Label: "Active"}, {Value: "inactive", Label: "Inactive"}}
					p.selectView(c, d, base).Shrink(0)
					control.NativeSelect(c, &d.sort, control.SelectProps{Label: "Source sort order", Options: []control.Option{{Value: "", Label: "Sort by"}, {Value: "name", Label: "Name"}, {Value: "date", Label: "Date"}, {Value: "status", Label: "Status"}}}).MinWidth(0).Width(140).Shrink(0)
				})
			})
		case "create-dialog":
			if control.Button(c, "Open Dialog", control.ButtonProps{Variant: control.Outline}).Clicked() {
				d.dialog = true
			}
			overlay.Dialog(c, &d.dialog, overlay.DialogProps{Title: "Select Example", Description: "Use the select below to choose a fruit."}, func() {
				base.label = "Dialog fruit"
				p.selectView(c, d, base)
			})
		}
	})
}
