package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

// Application values and public scroll state persist. Each native control and
// Field scope is constructed by the current view and never retained here.
type checkboxCompositionPreviewState struct {
	ready, native            bool
	recipe, language, action string
	popupError               string
	drafts                   map[string]*checkboxCompositionDraft
	changes                  map[string]int
}

type checkboxCompositionDraft struct {
	values [4]bool
	rows   [4]bool
	scroll ui.ScrollState
}

var checkboxCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Demo"},
	{Value: "basic", Label: "Basic"},
	{Value: "description", Label: "With Description"},
	{Value: "disabled", Label: "Disabled"},
	{Value: "group", Label: "Desktop Preferences (docs)"},
	{Value: "invalid", Label: "Invalid"},
	{Value: "table", Label: "In Table (docs)"},
	{Value: "rtl", Label: "Translated Checkboxes"},
	{Value: "create-basic", Label: "Create: Basic"},
	{Value: "create-description", Label: "Create: With Description"},
	{Value: "create-invalid", Label: "Create: Invalid"},
	{Value: "create-disabled", Label: "Create: Disabled"},
	{Value: "create-title", Label: "Create: With Title"},
	{Value: "create-table", Label: "Create: In Table"},
	{Value: "create-group", Label: "Create: Desktop Preferences"},
}

func (p *checkboxCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.drafts = map[string]*checkboxCompositionDraft{}
	p.changes = map[string]int{}
}

func (p *checkboxCompositionPreviewState) draft() *checkboxCompositionDraft {
	if m := p.drafts[p.recipe]; m != nil {
		return m
	}
	m := &checkboxCompositionDraft{}
	switch p.recipe {
	case "demo", "rtl":
		m.values[1] = true
	case "description", "create-description", "create-title":
		m.values[0] = true
	case "group":
		m.values[0], m.values[1] = true, true
	}
	if p.recipe == "table" || p.recipe == "create-table" {
		m.rows[0] = true
	}
	p.drafts[p.recipe] = m
	return m
}

func (p *checkboxCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *checkboxCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-checkbox-config")
			h.Close("source-checkbox-language")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Checkbox compositions"}, display.CardSlots{
		Title: "Checkbox compositions", Description: "Source fields, cards, desktop preferences and controlled table selection.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-checkbox-config", &p.recipe, control.SelectProps{Label: "Checkbox configuration", Options: checkboxCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-checkbox-language", &p.language, control.SelectProps{Label: "Checkbox language", Options: statusLanguages()})
			}
			ui.Column(child.Key(p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(func() {
				if p.recipe == "rtl" {
					ui.Column(child).FillWidth().MinWidth(0).Direction(childDirection(p.language)).Children(func() { p.draw(child) })
				} else {
					p.draw(child)
				}
			})
			if p.action != "" {
				ui.Text(child, p.action).FontSize(13).Role(ui.RoleStatus).Label(p.action)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).FontSize(13).Role(ui.RoleAlert)
			}
		},
	})
}

func (p *checkboxCompositionPreviewState) changed(label string) {
	p.changes[p.recipe]++
	p.action = "Local checkbox change: " + label
}

type checkboxSourceField struct {
	label, description                     string
	index                                  int
	disabled, fieldDisabled, invalid, card bool
	plain, normalWeight                    bool
}

func checkboxSourceWords(language string) [4]string {
	switch language {
	case "ar":
		return [4]string{"قبول الشروط والأحكام", "بالنقر على هذا المربع، فإنك توافق على الشروط.", "تفعيل الإشعارات", "يمكنك تفعيل أو إلغاء تفعيل الإشعارات في أي وقت."}
	case "he":
		return [4]string{"קבל תנאים והגבלות", "על ידי לחיצה על תיבת הסימון הזו, אתה מסכים לתנאים.", "הפעל התראות", "אתה יכול להפעיל או להשבית התראות בכל עת."}
	default:
		return [4]string{"Accept terms and conditions", "By clicking this checkbox, you agree to the terms.", "Enable notifications", "You can enable or disable notifications at any time."}
	}
}

func (p *checkboxCompositionPreviewState) field(c *ui.Context, m *checkboxCompositionDraft, s checkboxSourceField) {
	control.FieldWith(c.Key(s.index), control.FieldProps{Disabled: s.fieldDisabled, Invalid: s.invalid, Orientation: control.FieldHorizontal}, func(f *control.FieldScope) {
		checkbox := func() {
			f.Control(c, func() ui.Element {
				e := control.Checkbox(c, &m.values[s.index], "", control.CheckboxProps{Disabled: s.disabled, Invalid: s.invalid, OnCheckedChange: func(bool) { p.changed(s.label) }})
				if s.description != "" {
					e.Margin(1, 0, 0, 0)
				}
				return e
			})
		}
		if s.card {
			card := f.Label(c, control.FieldLabelProps{Label: s.label, Choice: true, Selected: m.values[s.index], Disabled: s.fieldDisabled}, func() {
				ui.Row(c).FillWidth().MinWidth(0).Gap(12).AlignItems(ui.Start).Children(func() {
					checkbox()
					f.Content(c, func() {
						title := control.FieldTitle(c, s.label)
						if s.fieldDisabled {
							title.Opacity(.5)
						}
						f.Description(c, s.description)
					})
				})
			})
			// The source's data-disabled styles dim the title and native check,
			// while its description and card boundary remain at normal opacity.
			if s.fieldDisabled {
				card.Opacity(1)
			}
			return
		}
		checkbox()
		caption := func() {
			label := f.Label(c, control.FieldLabelProps{Text: s.label}, nil)
			if s.plain {
				label.LineHeight(1)
			}
			if s.normalWeight {
				label.FontWeight(400)
			}
			if s.description != "" {
				f.Description(c, s.description)
			}
		}
		if s.description != "" {
			f.Content(c, caption)
		} else {
			caption()
		}
	})
}

func (p *checkboxCompositionPreviewState) draw(c *ui.Context) {
	m := p.draft()
	if p.recipe == "table" || p.recipe == "create-table" {
		p.table(c, m)
		return
	}
	width := float32(384)
	switch p.recipe {
	case "basic", "disabled", "invalid":
		width = 224
	case "description":
		width = 288
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(width).AlignSelf(ui.Center).Gap(28).Children(func() {
		words := checkboxSourceWords("en")
		if p.recipe == "rtl" {
			words = checkboxSourceWords(p.language)
		}
		switch p.recipe {
		case "demo", "rtl":
			p.field(c, m, checkboxSourceField{label: words[0], index: 0, plain: true})
			p.field(c, m, checkboxSourceField{label: words[0], description: words[1], index: 1})
			p.field(c, m, checkboxSourceField{label: words[2], index: 2, disabled: true, fieldDisabled: true})
			p.field(c, m, checkboxSourceField{label: words[2], description: words[3], index: 3, card: true})
		case "basic", "create-basic":
			p.field(c, m, checkboxSourceField{label: words[0]})
		case "description", "create-description":
			p.field(c, m, checkboxSourceField{label: words[0], description: "By clicking this checkbox, you agree to the terms and conditions."})
		case "disabled", "create-disabled":
			p.field(c, m, checkboxSourceField{label: words[2], disabled: true, fieldDisabled: p.recipe == "disabled"})
		case "invalid", "create-invalid":
			p.field(c, m, checkboxSourceField{label: words[0], invalid: true})
		case "create-title":
			p.field(c, m, checkboxSourceField{label: words[2], description: words[3], card: true})
			p.field(c, m, checkboxSourceField{label: words[2], description: words[3], index: 1, card: true, disabled: true, fieldDisabled: true})
		case "group", "create-group":
			fields := func(child *ui.Context) {
				ui.Column(child).FillWidth().MinWidth(0).Gap(12).Children(func() {
					for i, label := range []string{"Hard disks", "External disks", "CDs, DVDs, and iPods", "Connected servers"} {
						p.field(child, m, checkboxSourceField{label: label, index: i, normalWeight: true})
					}
				})
			}
			if p.recipe == "group" {
				control.FieldSetWith(c, control.FieldSetProps{Legend: "Show these items on the desktop:", Description: "Select the items you want to show on the desktop."}, control.FieldSetSlots{
					Legend: func(child *ui.Context) ui.Element {
						return control.FieldLegend(child, "Show these items on the desktop:", control.FieldLegendLabel)
					}, Content: fields,
				})
			} else {
				ui.Column(c).FillWidth().MinWidth(0).Gap(12).Role(ui.RoleGroup).Label("Show these items on the desktop:").Children(func() {
					control.FieldLabel(c, control.FieldLabelProps{Text: "Show these items on the desktop:"}, nil)
					fields(c)
				})
			}
		}
	})
}

var checkboxSourcePeople = [4]struct{ name, email, role string }{
	{"Sarah Chen", "sarah.chen@example.com", "Admin"},
	{"Marcus Rodriguez", "marcus.rodriguez@example.com", "User"},
	{"Priya Patel", "priya.patel@example.com", "User"},
	{"David Kim", "david.kim@example.com", "Editor"},
}

func (p *checkboxCompositionPreviewState) table(c *ui.Context, m *checkboxCompositionDraft) {
	all := true
	for _, selected := range m.rows {
		all = all && selected
	}
	header := display.TableRow(display.TableRowProps{Key: "header"},
		display.TableHead("", display.TableCellProps{Checkbox: true, Content: func(child *ui.Context) {
			// The pinned source uses boolean all-selected, not indeterminate.
			// Keep a partial selection visually unchecked, while deriving this
			// controlled header from the actual independent row values.
			control.Checkbox(child, &all, "", control.CheckboxProps{OnCheckedChange: func(checked bool) {
				for i := range m.rows {
					m.rows[i] = checked
				}
				p.changed("Select all rows")
			}}).Label("Select all rows")
		}}),
		display.TableHead("Name"), display.TableHead("Email"), display.TableHead("Role"),
	)
	rows := make([]display.TableRowProps, 0, len(checkboxSourcePeople))
	for i, person := range checkboxSourcePeople {
		rows = append(rows, display.TableRow(display.TableRowProps{Key: person.email, Selected: m.rows[i]},
			display.TableCell("", display.TableCellProps{Checkbox: true, Content: func(child *ui.Context) {
				control.Checkbox(child, &m.rows[i], "", control.CheckboxProps{OnCheckedChange: func(bool) { p.changed(person.name) }}).Label("Select " + person.name)
			}}),
			display.TableCell(person.name, display.TableCellProps{Medium: true}), display.TableCell(person.email), display.TableCell(person.role),
		))
	}
	display.TableWith(c, display.TableOptions{Label: "Source checkbox table", MinWidth: 600, Scroll: &m.scroll, Columns: []display.TableColumn{{Key: "selected", Width: 32}, {Key: "name"}, {Key: "email"}, {Key: "role"}}}, display.TableHeader(header), display.TableBody(rows...))
}

func checkboxRecipeLabel(key string) string {
	for _, recipe := range checkboxCompositionRecipes {
		if recipe.Value == key {
			return recipe.Label
		}
	}
	return strings.TrimPrefix(key, "create-")
}
