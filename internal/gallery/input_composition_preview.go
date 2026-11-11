package gallery

import (
	"regexp"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Only application values and public handles survive a view. The picker gets
// native dialog values, never a frame Context. Native editors own their buffers.
type inputCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	drafts                               map[string]*inputCompositionDraft
	picker                               control.FilePicker
}
type inputCompositionDraft struct {
	texts              map[string]*string
	targets            map[string]*ui.Handle
	choice             string
	number             float64
	numberEmpty        bool
	date, clock        time.Time
	files              []string
	saved              map[string]string
	submitted, changes int
	error              string
}

var inputCompositionRecipes = []control.Option{
	{Value: "demo", Label: "API Key"}, {Value: "basic", Label: "Basic"},
	{Value: "field", Label: "Username Field"}, {Value: "fieldgroup", Label: "Name and Email Field Group"},
	{Value: "disabled", Label: "Disabled Field"}, {Value: "invalid", Label: "Invalid Field"},
	{Value: "file", Label: "Picture File"}, {Value: "inline", Label: "Inline Search"},
	{Value: "grid", Label: "First and Last Name Grid"}, {Value: "required", Label: "Required Field"},
	{Value: "badge", Label: "Webhook Beta Badge"}, {Value: "input-group", Label: "Website Input Group"},
	{Value: "button-group", Label: "Search Button Group"}, {Value: "form", Label: "Contact Form (docs)"},
	{Value: "rtl", Label: "Translated API Key"},
	{Value: "create-basic", Label: "Create: Basic Email"}, {Value: "create-invalid", Label: "Create: Invalid"},
	{Value: "create-label", Label: "Create: With Label"}, {Value: "create-description", Label: "Create: With Description"},
	{Value: "create-disabled", Label: "Create: Disabled"}, {Value: "create-types", Label: "Create: Input Types"},
	{Value: "create-select", Label: "Create: With Select"}, {Value: "create-button", Label: "Create: With Button"},
	{Value: "create-native-select", Label: "Create: With Native Select"}, {Value: "create-form", Label: "Create: Contact Form"},
}

func (p *inputCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.drafts = map[string]*inputCompositionDraft{}
}
func (p *inputCompositionPreviewState) draft() *inputCompositionDraft {
	if d := p.drafts[p.recipe]; d != nil {
		return d
	}
	now := time.Now()
	d := &inputCompositionDraft{texts: map[string]*string{}, targets: map[string]*ui.Handle{}, choice: "us", numberEmpty: true, date: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), clock: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)}
	if p.recipe == "create-select" {
		d.choice = "usd"
	}
	if p.recipe == "create-native-select" {
		d.choice = "+1"
	}
	p.drafts[p.recipe] = d
	return d
}
func (d *inputCompositionDraft) text(key string) *string {
	if v := d.texts[key]; v != nil {
		return v
	}
	v := ""
	d.texts[key] = &v
	return &v
}
func (d *inputCompositionDraft) target(key string) *ui.Handle {
	if h := d.targets[key]; h != nil {
		return h
	}
	h := &ui.Handle{}
	d.targets[key] = h
	return h
}
func (p *inputCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) ui.Element {
	if native {
		return control.NativeSelect(c, value, props).MinWidth(0)
	}
	e, err := desktop.Select(c, h, key, value, props)
	if err != nil {
		p.popupError = err.Error()
	}
	return e.MinWidth(0)
}
func (p *inputCompositionPreviewState) view(c *ui.Context, window *mygo.Window, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-input-config")
			h.Close("source-input-language")
			h.Close("source-input-country")
			h.Close("source-input-currency")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: tr(c, "Input compositions")}, display.CardSlots{Title: tr(c, "Input compositions"), Description: tr(c, "Source fields, native editors, typed values and local form drafts."), Content: func(child *ui.Context) {
		p.choose(child, h, native, "source-input-config", &p.recipe, control.SelectProps{Label: tr(child, "Input configuration"), Options: localizedOptions(child, inputCompositionRecipes)})
		if p.recipe == "rtl" {
			p.choose(child, h, native, "source-input-language", &p.language, control.SelectProps{Label: tr(child, "Input language"), Options: localizedOptions(child, statusLanguages())})
		}
		box := ui.Column(child.Key("source-input:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
		if p.recipe == "rtl" {
			box.Direction(childDirection(p.language))
		}
		box.Children(func() { p.draw(child, window, h, native) })
		if p.action != "" {
			ui.Text(child, tr(child, p.action)).FontSize(13).Role(ui.RoleStatus).Label(tr(child, p.action))
		}
		if p.popupError != "" {
			ui.Text(child, p.popupError).Role(ui.RoleAlert).TextColor(theme.TokensOf(child).Destructive)
		}
	}})
}

type inputSourceField struct {
	key, label, placeholder, description, kind                        string
	disabled, fieldDisabled, invalid, required, requiredMarker, badge bool
}

func (p *inputCompositionPreviewState) editor(c *ui.Context, window *mygo.Window, d *inputCompositionDraft, s inputSourceField) ui.Element {
	changed := func() { d.changes++; p.action = "Local input edit: " + s.label }
	switch s.kind {
	case "number":
		return control.NumberInput(c, &d.number, control.NumberInputProps{Label: tr(c, s.label), Placeholder: tr(c, s.placeholder), Empty: &d.numberEmpty, OnEmptyChange: func(empty bool) {
			if empty {
				changed()
			}
		}, OnChange: func(float64) { changed() }})
	case "date":
		return control.DateInput(c, &d.date, control.DateInputProps{Label: tr(c, s.label), OnChange: func(time.Time) { changed() }})
	case "time":
		return control.TimeInput(c, &d.clock, control.TimeInputProps{Label: tr(c, s.label), OnChange: func(time.Time) { changed() }})
	case "file":
		return control.FileInput(c, &d.files, control.FileInputProps{Label: tr(c, s.label), Parent: window, Picker: p.picker, OnChange: func([]string) { changed() }, OnError: func(err error) { p.action = "File selection error: " + err.Error() }})
	}
	return control.Input(c, d.text(s.key), control.InputProps{Label: tr(c, s.label), Placeholder: tr(c, s.placeholder), Disabled: s.disabled, Invalid: s.invalid, Password: s.kind == "password", Search: s.kind == "search", Handle: d.target(s.key), OnChange: func(string) { changed() }})
}
func (p *inputCompositionPreviewState) field(c *ui.Context, window *mygo.Window, d *inputCompositionDraft, s inputSourceField) {
	control.FieldWith(c.Key(s.key), control.FieldProps{Required: s.required, Disabled: s.fieldDisabled, Invalid: s.invalid}, func(f *control.FieldScope) {
		if s.badge || s.required {
			f.Label(c, control.FieldLabelProps{Label: tr(c, s.label)}, func() {
				ui.Text(c, tr(c, s.label)).FontSize(14).FontWeight(500)
				if s.requiredMarker {
					ui.Text(c, "*").FontSize(14).TextColor(theme.TokensOf(c).Destructive)
				}
				if s.badge {
					ui.Box(c).Grow(1)
					display.Badge(c, tr(c, "Beta"), display.BadgeSecondary)
				}
			})
		} else {
			f.Label(c, control.FieldLabelProps{Text: tr(c, s.label)}, nil)
		}
		f.Control(c, func() ui.Element { return p.editor(c, window, d, s) })
		if s.description != "" {
			f.Description(c, tr(c, s.description))
		}
	})
}
func (p *inputCompositionPreviewState) apiWords() (string, string) {
	switch p.language {
	case "ar":
		return "مفتاح API", "مفتاح API الخاص بك مشفر ومخزن بأمان."
	case "he":
		return "מפתח API", "מפתח ה-API שלך מוצפן ונשמר בצורה מאובטחת."
	}
	return "API Key", "Your API key is encrypted and stored securely."
}
func (p *inputCompositionPreviewState) search(d *inputCompositionDraft, key string) {
	d.submitted++
	d.saved = map[string]string{"query": *d.text(key)}
	p.action = "Local input action: Search"
}
func (p *inputCompositionPreviewState) draw(c *ui.Context, window *mygo.Window, h *desktop.Host, native bool) {
	d := p.draft()
	width := float32(320)
	if strings.HasPrefix(p.recipe, "create-") || p.recipe == "form" || p.recipe == "grid" {
		width = 384
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(width).Gap(28).Children(func() {
		switch p.recipe {
		case "demo", "rtl":
			label, description := "API Key", "Your API key is encrypted and stored securely."
			if p.recipe == "rtl" {
				label, description = p.apiWords()
			}
			p.field(c, window, d, inputSourceField{key: "api", label: label, placeholder: "sk-...", description: description, kind: "password"})
		case "basic":
			p.editor(c, window, d, inputSourceField{key: "text", label: "Enter text", placeholder: "Enter text"})
		case "create-basic":
			p.editor(c, window, d, inputSourceField{key: "email", label: "Email", placeholder: "Email", kind: "email"})
		case "field", "create-description":
			p.field(c, window, d, inputSourceField{key: "username", label: "Username", placeholder: "Enter your username", description: "Choose a unique username for your account."})
		case "create-label":
			p.field(c, window, d, inputSourceField{key: "email", label: "Email", placeholder: "name@example.com", kind: "email"})
		case "disabled":
			p.field(c, window, d, inputSourceField{key: "email", label: "Email", placeholder: "Email", description: "This field is currently disabled.", kind: "email", disabled: true, fieldDisabled: true})
		case "create-disabled":
			p.field(c, window, d, inputSourceField{key: "email", label: "Email", placeholder: "Email", kind: "email", disabled: true})
		case "invalid":
			p.field(c, window, d, inputSourceField{key: "invalid", label: "Invalid Input", placeholder: "Error", description: "This field contains validation errors.", invalid: true})
		case "create-invalid":
			p.editor(c, window, d, inputSourceField{key: "invalid", label: "Error", placeholder: "Error", invalid: true})
		case "required":
			p.field(c, window, d, inputSourceField{key: "required", label: "Required Field", placeholder: "This field is required", description: "This field must be filled out.", required: true, requiredMarker: true})
		case "badge":
			p.field(c, window, d, inputSourceField{key: "webhook", label: "Webhook URL", placeholder: "https://api.example.com/webhook", kind: "url", badge: true})
		case "file":
			p.field(c, window, d, inputSourceField{key: "picture", label: "Picture", description: "Select a picture to upload.", kind: "file"})
		case "grid":
			ui.Column(c).FillWidth().MinWidth(0).Grid().Columns(2).Gap(28).Children(func() {
				p.field(c, window, d, inputSourceField{key: "first", label: "First Name", placeholder: "Jordan"})
				p.field(c, window, d, inputSourceField{key: "last", label: "Last Name", placeholder: "Lee"})
			})
		case "input-group":
			control.FieldWith(c, control.FieldProps{Label: tr(c, "Website URL")}, func(f *control.FieldScope) {
				f.Label(c, control.FieldLabelProps{Text: tr(c, "Website URL")}, nil)
				f.Control(c, func() ui.Element {
					return control.InputGroup(c, control.InputGroupProps{Prefix: "https://", InlineEnd: func() { ui.Icon(c, outlined.Info()).Size(16, 16).AccessHidden(true) }}, func() ui.Element {
						return p.editor(c, window, d, inputSourceField{key: "website", label: "Website URL", placeholder: "example.com"})
					})
				})
			})
		case "button-group":
			control.FieldWith(c, control.FieldProps{Label: tr(c, "Search")}, func(f *control.FieldScope) {
				f.Label(c, control.FieldLabelProps{Text: tr(c, "Search")}, nil)
				control.ButtonGroup(c, func() {
					f.Control(c, func() ui.Element {
						return p.editor(c, window, d, inputSourceField{key: "search", label: "Search", placeholder: "Type to search..."}).Grow(1).MinWidth(0)
					})
					if control.Button(c, tr(c, "Search"), control.ButtonProps{Variant: control.Outline}).Clicked() {
						p.search(d, "search")
					}
				}).FillWidth()
			})
		case "inline", "create-button":
			gap := float32(8)
			if p.recipe == "inline" {
				gap = 12
			}
			ui.Row(c).FillWidth().MinWidth(0).Gap(gap).Children(func() {
				p.editor(c, window, d, inputSourceField{key: "search", label: "Search query", placeholder: "Search...", kind: "search"}).Grow(1).MinWidth(0)
				if control.Button(c, tr(c, "Search"), control.ButtonProps{}).Clicked() {
					p.search(d, "search")
				}
			})
		case "create-select":
			ui.Row(c).FillWidth().MinWidth(0).Gap(8).Children(func() {
				p.editor(c, window, d, inputSourceField{key: "amount", label: "Amount", placeholder: "Enter amount"}).Grow(1).MinWidth(0)
				p.choose(c, h, native, "source-input-currency", &d.choice, control.SelectProps{Label: tr(c, "Currency"), Options: localizedOptions(c, []control.Option{{Value: "usd", Label: "USD"}, {Value: "eur", Label: "EUR"}, {Value: "gbp", Label: "GBP"}})}).Width(128).Shrink(0)
			})
		case "create-native-select":
			ui.Row(c).FillWidth().MinWidth(0).Gap(8).Children(func() {
				p.editor(c, window, d, inputSourceField{key: "phone", label: "Phone", placeholder: "(555) 123-4567", kind: "tel"}).Grow(1).MinWidth(0)
				control.NativeSelect(c, &d.choice, control.SelectProps{Label: tr(c, "Phone country code"), Options: localizedOptions(c, []control.Option{{Value: "+1", Label: "+1"}, {Value: "+44", Label: "+44"}, {Value: "+46", Label: "+46"}})}).Shrink(0)
			})
		case "fieldgroup":
			p.simpleForm(c, window, d)
		case "form", "create-form":
			p.form(c, window, h, native, d)
		case "create-types":
			ui.Column(c).FillWidth().MinWidth(0).Gap(24).Children(func() {
				for _, s := range []inputSourceField{{key: "password", label: "Password", placeholder: "Password", kind: "password"}, {key: "phone", label: "Phone", placeholder: "+1 (555) 123-4567", kind: "tel"}, {key: "url", label: "URL", placeholder: "https://example.com", kind: "url"}, {key: "search", label: "Search", placeholder: "Search", kind: "search"}, {key: "number", label: "Number", placeholder: "123", kind: "number"}, {key: "date", label: "Date", kind: "date"}, {key: "time", label: "Time", kind: "time"}, {key: "file", label: "File", kind: "file"}} {
					p.field(c, window, d, s)
				}
			})
		}
		if d.error != "" {
			ui.Text(c, tr(c, d.error)).FontSize(13).TextColor(theme.TokensOf(c).Destructive).Role(ui.RoleAlert)
		}
		if d.saved != nil {
			ui.Text(c, tr(c, "Saved input values: ")+inputSavedSummary(d.saved)).FontSize(12).Role(ui.RoleStatus)
		}
	})
}
func inputSavedSummary(values map[string]string) string {
	var parts []string
	for _, key := range []string{"name", "email", "phone", "country", "address", "query"} {
		if v, ok := values[key]; ok {
			parts = append(parts, key+"="+v)
		}
	}
	return strings.Join(parts, " · ")
}

// ASCII email shape from the WHATWG email-state definition. Browser IDNA and
// constraint bubbles are not supplied by this native example.
var inputEmailPattern = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func inputEmailValid(s string) bool { return s == "" || inputEmailPattern.MatchString(s) }
func (p *inputCompositionPreviewState) save(d *inputCompositionDraft, required bool) {
	d.error = ""
	if required && *d.text("name") == "" {
		d.error = "Name is required."
		d.target("name").Focus()
	} else if !inputEmailValid(strings.Trim(*d.text("email"), " \t\n\r\f")) {
		d.error = "Enter a valid email address."
		d.target("email").Focus()
	}
	if d.error != "" {
		p.action = "Local input validation: " + d.error
		return
	}
	d.saved = map[string]string{}
	for _, key := range []string{"name", "email", "phone", "address"} {
		d.saved[key] = *d.text(key)
	}
	d.saved["country"] = d.choice
	d.saved["email"] = strings.Trim(d.saved["email"], " \t\n\r\f")
	d.submitted++
	p.action = "Local input action: Submit"
}
func (p *inputCompositionPreviewState) simpleForm(c *ui.Context, window *mygo.Window, d *inputCompositionDraft) {
	p.field(c, window, d, inputSourceField{key: "name", label: "Name", placeholder: "Jordan Lee"})
	p.field(c, window, d, inputSourceField{key: "email", label: "Email", placeholder: "name@example.com", description: "We'll send updates to this address.", kind: "email"})
	ui.Row(c).Gap(12).Children(func() {
		if control.Button(c, tr(c, "Reset"), control.ButtonProps{Variant: control.Outline}).Clicked() {
			for _, v := range d.texts {
				*v = ""
			}
			d.saved = nil
			d.error = ""
			p.action = "Local input action: Reset"
		}
		if control.Button(c, tr(c, "Submit"), control.ButtonProps{}).Clicked() {
			p.save(d, false)
		}
	})
}
func (p *inputCompositionPreviewState) form(c *ui.Context, window *mygo.Window, h *desktop.Host, native bool, d *inputCompositionDraft) {
	namePlaceholder := "Evil Rabbit"
	if p.recipe == "create-form" {
		namePlaceholder = "John Doe"
	}
	p.field(c, window, d, inputSourceField{key: "name", label: "Name", placeholder: namePlaceholder, required: p.recipe == "form"})
	p.field(c, window, d, inputSourceField{key: "email", label: "Email", placeholder: "john@example.com", description: "We'll never share your email with anyone.", kind: "email"})
	ui.Column(c).FillWidth().MinWidth(0).Grid().Columns(2).Gap(16).Children(func() {
		p.field(c, window, d, inputSourceField{key: "phone", label: "Phone", placeholder: "+1 (555) 123-4567", kind: "tel"})
		control.FieldWith(c, control.FieldProps{Label: tr(c, "Country")}, func(f *control.FieldScope) {
			f.Label(c, control.FieldLabelProps{Text: tr(c, "Country")}, nil)
			f.Control(c, func() ui.Element {
				return p.choose(c, h, native, "source-input-country", &d.choice, control.SelectProps{Label: tr(c, "Country"), Options: localizedOptions(c, []control.Option{{Value: "us", Label: "United States"}, {Value: "uk", Label: "United Kingdom"}, {Value: "ca", Label: "Canada"}})})
			})
		})
	})
	p.field(c, window, d, inputSourceField{key: "address", label: "Address", placeholder: "123 Main St"})
	ui.Row(c).Gap(12).Children(func() {
		if control.Button(c, tr(c, "Cancel"), control.ButtonProps{Variant: control.Outline}).Clicked() {
			p.action = "Local input action: Cancel"
		}
		if control.Button(c, tr(c, "Submit"), control.ButtonProps{}).Clicked() {
			p.save(d, p.recipe == "form")
		}
	})
}
