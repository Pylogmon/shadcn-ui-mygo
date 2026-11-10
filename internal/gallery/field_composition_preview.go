package gallery

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// These are application drafts and saved local examples. Builders and native
// elements belong only to the current frame.
type fieldCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, status, popupError string
	variants                             map[string]string
	texts                                map[string]*string
	checks                               map[string]*bool
	numbers                              map[string]*[]float64
	selectKeys                           map[string]bool
	profileError, profileSaved           string
	profileName                          ui.Handle
	paymentSaved                         map[string]fieldPaymentSnapshot
}
type fieldPaymentSnapshot struct {
	Name, Number, Month, Year, CVV, Comments string
	Shipping                                 bool
}

var fieldCompositionRecipes = []control.Option{
	{Value: "responsive", Label: "Responsive profile"}, {Value: "payment", Label: "Payment fields"},
	{Value: "input", Label: "Username and password"}, {Value: "textarea", Label: "Feedback field"},
	{Value: "select", Label: "Department field"}, {Value: "slider", Label: "Budget range"},
	{Value: "fieldset", Label: "Address fieldset"}, {Value: "checkbox", Label: "Desktop checkboxes"},
	{Value: "radio", Label: "Subscription plan"}, {Value: "switch", Label: "Authentication switch"},
	{Value: "choice", Label: "Compute environment"}, {Value: "group", Label: "Notification groups"},
	{Value: "rtl", Label: "Translated payment"},
	{Value: "create-input", Label: "Create: Input fields"}, {Value: "create-textarea", Label: "Create: Textarea fields"},
	{Value: "create-select", Label: "Create: Select fields"}, {Value: "create-checkbox", Label: "Create: Checkbox fields"},
	{Value: "create-radio", Label: "Create: Radio fields"}, {Value: "create-switch", Label: "Create: Switch fields"},
	{Value: "create-slider", Label: "Create: Slider fields"}, {Value: "create-native", Label: "Create: Native select fields"},
	{Value: "create-otp", Label: "Create: OTP fields"}, {Value: "create-horizontal", Label: "Create: Horizontal fields"},
}

type fieldRecipe struct {
	rows                                                      int
	minHeight                                                 float32
	id, label, kind, description, placeholder                 string
	before, disabled, invalid, required, badge, right, choice bool
	options                                                   []control.Option
	values                                                    []float64
	min, max, step                                            float64
	length                                                    int
	groups                                                    []int
}

func (p *fieldCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "responsive", "ar"
	p.variants = map[string]string{}
	p.texts = map[string]*string{}
	p.checks = map[string]*bool{}
	p.numbers = map[string]*[]float64{}
	p.selectKeys = map[string]bool{}
	p.paymentSaved = map[string]fieldPaymentSnapshot{}
}
func (p *fieldCompositionPreviewState) text(key, initial string) *string {
	if v := p.texts[key]; v != nil {
		return v
	}
	v := initial
	p.texts[key] = &v
	return &v
}
func (p *fieldCompositionPreviewState) checked(key string, initial bool) *bool {
	if v := p.checks[key]; v != nil {
		return v
	}
	v := initial
	p.checks[key] = &v
	return &v
}
func (p *fieldCompositionPreviewState) values(key string, initial ...float64) *[]float64 {
	if v := p.numbers[key]; v != nil {
		return v
	}
	v := append([]float64(nil), initial...)
	p.numbers[key] = &v
	return &v
}
func (p *fieldCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) ui.Element {
	p.selectKeys[key] = true
	if native {
		return control.NativeSelect(c, value, props)
	}
	e, err := desktop.Select(c, h, key, value, props)
	if err != nil {
		p.popupError = err.Error()
	}
	return e
}
func (p *fieldCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			for key := range p.selectKeys {
				h.Close(key)
			}
		}
		p.native = native
		p.popupError = ""
	}
	display.CardWith(c, display.CardProps{Label: "Field compositions"}, display.CardSlots{Title: "Field compositions", Description: "Independent labels, descriptions, controls and grouped form layouts.", Content: func(child *ui.Context) {
		p.choose(child, h, native, "field-recipe", &p.recipe, control.SelectProps{Label: "Field configuration", Options: fieldCompositionRecipes})
		if strings.HasPrefix(p.recipe, "create-") {
			recipes := fieldCreateRecipes(p.recipe)
			options := make([]control.Option, 0, len(recipes))
			for _, d := range recipes {
				options = append(options, control.Option{Value: d.id, Label: d.label})
			}
			selected := p.variants[p.recipe]
			if selected == "" && len(recipes) > 0 {
				selected = recipes[0].id
			}
			p.choose(child, h, native, "field-variant", &selected, control.SelectProps{Label: "Field variant", Options: options})
			p.variants[p.recipe] = selected
		}
		if p.recipe == "rtl" {
			p.choose(child, h, native, "field-language", &p.language, control.SelectProps{Label: "Field language", Options: []control.Option{{Value: "en", Label: "English"}, {Value: "ar", Label: "العربية"}, {Value: "he", Label: "עברית"}}})
		}
		dir := child.Direction()
		if p.recipe == "rtl" {
			dir = ui.LeftToRight
			if p.language != "en" {
				dir = ui.RightToLeft
			}
		}
		display.Direction(child, dir, func() {
			ui.Column(child.Key("field-preview")).FillWidth().MaxWidth(p.previewWidth()).MinWidth(0).Gap(28).Children(func() { p.draw(child, h, native) })
		})
		if p.status != "" {
			ui.Text(child, p.status).Role(ui.RoleStatus).FontSize(13).TextColor(theme.TokensOf(child).MutedForeground)
		}
		if p.popupError != "" {
			ui.Text(child, p.popupError).Role(ui.RoleAlert).TextColor(theme.TokensOf(child).Destructive)
		}
	}})
}

func (p *fieldCompositionPreviewState) previewWidth() float32 {
	if strings.HasPrefix(p.recipe, "create-") || p.recipe == "responsive" {
		return 512
	}
	if p.recipe == "payment" || p.recipe == "rtl" {
		return 448
	}
	if p.recipe == "fieldset" {
		return 384
	}
	return 320
}

func fieldCompositionGroup(c *ui.Context, build func()) ui.Element {
	return control.FieldGroup(c, build).Gap(28)
}

func fieldSet(c *ui.Context, title, description string, label bool, content func(*ui.Context)) ui.Element {
	variant := control.FieldLegendDefault
	if label {
		variant = control.FieldLegendLabel
	}
	return control.FieldSetWith(c, control.FieldSetProps{Legend: title, Description: description}, control.FieldSetSlots{
		Legend: func(child *ui.Context) ui.Element { return control.FieldLegend(child, title, variant) },
		Description: func(child *ui.Context) ui.Element {
			if description == "" {
				return ui.Element{}
			}
			return control.FieldDescription(child, description)
		}, Content: content,
	})
}
func (p *fieldCompositionPreviewState) basic(c *ui.Context, h *desktop.Host, native bool, d fieldRecipe) ui.Element {
	key := p.recipe + ":" + d.id
	fieldDisabled := d.disabled && d.id != "disabled-input"
	props := control.FieldProps{Required: d.required, Disabled: fieldDisabled, Invalid: d.invalid}
	if p.recipe == "create-horizontal" {
		props.Orientation = control.FieldHorizontal
	}
	result := control.FieldWith(c.Key(d.id), props, func(f *control.FieldScope) {
		label := func() {
			f.Label(c, control.FieldLabelProps{Text: d.label, Label: d.label, Disabled: fieldDisabled}, func() {
				control.FieldTitle(c, d.label)
				if d.required {
					ui.Text(c, "*").TextColor(theme.TokensOf(c).Destructive)
				}
				if d.badge {
					ui.Box(c).Grow(1)
					display.Badge(c, "Recommended", display.BadgeSecondary)
				}
			})
		}
		desc := func() {
			if d.description != "" {
				description := f.Description(c, p.fieldDescription(key, d))
				if d.before && props.Orientation != control.FieldHorizontal {
					description.Margin(-4, 0, 0, 0)
				}
			}
		}
		if props.Orientation == control.FieldHorizontal {
			f.Content(c, func() { label(); desc() }).MinWidth(192)
		} else {
			label()
			if d.before {
				desc()
			}
		}
		f.Control(c, func() ui.Element {
			e := p.editor(c, h, native, key, d)
			if props.Orientation == control.FieldHorizontal && d.kind != "switch" {
				e.Basis(0).Grow(1).MinWidth(0)
			}
			return e
		})
		if props.Orientation != control.FieldHorizontal && !d.before {
			desc()
		}
	})
	if d.invalid {
		result.TextColor(theme.TokensOf(c).Destructive)
	}
	return result
}
func (p *fieldCompositionPreviewState) fieldDescription(key string, d fieldRecipe) string {
	if d.kind != "slider" {
		return d.description
	}
	v := *p.values(key, d.values...)
	switch d.id {
	case "brightness":
		return fmt.Sprintf("Current brightness: %.0f%%", v[0])
	case "temperature":
		return fmt.Sprintf("Range: %.1f - %.1f", v[0], v[1])
	case "price":
		return fmt.Sprintf("$%.0f - $%.0f", v[0], v[1])
	case "balance":
		return fmt.Sprintf("Red: %.0f%%, Green: %.0f%%, Blue: %.0f%%", v[0], v[1], v[2])
	}
	return d.description
}
func (p *fieldCompositionPreviewState) editor(c *ui.Context, h *desktop.Host, native bool, key string, d fieldRecipe) ui.Element {
	switch d.kind {
	case "textarea":
		e := control.Textarea(c, p.text(key, ""), control.TextareaProps{Label: d.label, Placeholder: d.placeholder, Disabled: d.disabled, Invalid: d.invalid, Rows: d.rows})
		if d.minHeight > 0 {
			e.MinHeight(d.minHeight)
		}
		return e
	case "select", "native":
		return p.choose(c, h, native || d.kind == "native", "field-option:"+key, p.text(key, ""), control.SelectProps{Label: d.label, Placeholder: d.placeholder, Options: d.options, Disabled: d.disabled, Invalid: d.invalid})
	case "slider":
		values := p.values(key, d.values...)
		return control.SliderValues(c, values, control.SliderProps{Label: d.label, Min: d.min, Max: d.max, Step: d.step, Disabled: d.disabled, Invalid: d.invalid})
	case "otp":
		return control.InputOTP(c, p.text(key, ""), control.InputOTPProps{Label: d.label, Length: d.length, Groups: d.groups, DigitsOnly: d.length == 4, Disabled: d.disabled, Invalid: d.invalid})
	case "switch":
		return control.Switch(c, p.checked(key, d.id == "profile"), control.SwitchProps{Label: d.label, Disabled: d.disabled, Invalid: d.invalid})
	default:
		return control.Input(c, p.text(key, ""), control.InputProps{Label: d.label, Placeholder: d.placeholder, Password: d.kind == "password", Disabled: d.disabled, Invalid: d.invalid})
	}
}
func (p *fieldCompositionPreviewState) choice(c *ui.Context, key, label, description, kind, value string, checked, disabled, invalid, right, card bool) {
	stateKey := p.recipe + ":" + key
	control.FieldWith(c.Key(key), control.FieldProps{Disabled: disabled, Invalid: invalid, Orientation: control.FieldHorizontal}, func(f *control.FieldScope) {
		controlBuild := func() {
			f.Control(c, func() ui.Element {
				switch kind {
				case "radio":
					return control.RadioWith(c, p.text(stateKey, ""), value, "", control.RadioProps{Disabled: disabled, Invalid: invalid})
				case "switch":
					return control.Switch(c, p.checked(stateKey, checked), control.SwitchProps{Disabled: disabled, Invalid: invalid})
				default:
					return control.Checkbox(c, p.checked(stateKey, checked), "", control.CheckboxProps{Disabled: disabled, Invalid: invalid})
				}
			})
		}
		caption := func() {
			f.Content(c, func() {
				control.FieldTitle(c, label)
				if description != "" {
					f.Description(c, description)
				}
			})
		}
		if card {
			selected := *p.checked(stateKey, checked)
			if kind == "radio" {
				selected = *p.text(stateKey, "") == value
			}
			f.Label(c, control.FieldLabelProps{Label: label, Choice: true, Selected: selected, Disabled: disabled}, func() {
				ui.Row(c).FillWidth().MinWidth(0).Gap(12).AlignItems(ui.Start).Children(func() {
					if right {
						caption()
						controlBuild()
					} else {
						controlBuild()
						caption()
					}
				})
			})
		} else {
			caption = func() {
				if description != "" {
					f.Content(c, func() { f.Label(c, control.FieldLabelProps{Text: label}, nil); f.Description(c, description) })
				} else {
					f.Label(c, control.FieldLabelProps{Text: label}, nil).Grow(1)
				}
			}
			if right {
				caption()
				controlBuild()
			} else {
				controlBuild()
				caption()
			}
		}
	})
}

func (p *fieldCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	switch p.recipe {
	case "responsive":
		p.profile(c)
	case "payment", "rtl":
		p.payment(c, h, native)
	case "input":
		fieldCompositionGroup(c, func() {
			p.basic(c, h, native, fieldRecipe{id: "username", label: "Username", placeholder: "Max Leiter", description: "Choose a unique username for your account."})
			p.basic(c, h, native, fieldRecipe{id: "password", label: "Password", kind: "password", placeholder: "••••••••", description: "Must be at least 8 characters long.", before: true})
		})
	case "textarea":
		p.basic(c, h, native, fieldRecipe{id: "feedback", label: "Feedback", kind: "textarea", rows: 4, placeholder: "Your feedback helps us improve...", description: "Share your thoughts about our service."})
	case "select":
		p.basic(c, h, native, fieldRecipe{id: "department", label: "Department", kind: "select", placeholder: "Choose department", description: "Select your department or area of work.", options: fieldDepartments})
	case "slider":
		key := "slider:budget"
		values := p.values(key, 200, 800)
		control.FieldWith(c, control.FieldProps{Label: "Price Range"}, func(f *control.FieldScope) {
			control.FieldTitle(c, "Price Range")
			f.Description(c, fmt.Sprintf("Set your budget range ($%.0f - $%.0f).", (*values)[0], (*values)[1]))
			f.Control(c, func() ui.Element {
				return control.SliderValues(c, values, control.SliderProps{Label: "Price Range", Max: 1000, Step: 10})
			})
		})
	case "fieldset":
		fieldSet(c, "Address Information", "We need your address to deliver your order.", false, func(child *ui.Context) {
			fieldCompositionGroup(child, func() {
				p.basic(child, h, native, fieldRecipe{id: "street", label: "Street Address", placeholder: "123 Main St"})
				ui.Row(child).FillWidth().Gap(16).Children(func() {
					ui.Column(child).Basis(0).Grow(1).MinWidth(0).Children(func() { p.basic(child, h, native, fieldRecipe{id: "city", label: "City", placeholder: "New York"}) })
					ui.Column(child).Basis(0).Grow(1).MinWidth(0).Children(func() { p.basic(child, h, native, fieldRecipe{id: "zip", label: "Postal Code", placeholder: "90502"}) })
				})
			})
		})
	case "checkbox":
		p.desktopChoices(c)
	case "radio":
		p.radioGroup(c, "plan", "Subscription Plan", "Yearly and lifetime plans offer significant savings.", []control.Option{{Value: "monthly", Label: "Monthly ($9.99/month)"}, {Value: "yearly", Label: "Yearly ($99.99/year)"}, {Value: "lifetime", Label: "Lifetime ($299.99)"}}, "monthly", false, false)
	case "switch":
		p.choice(c, "mfa", "Multi-factor authentication", "", "switch", "", false, false, false, true, false)
	case "choice":
		p.compute(c)
	case "group":
		p.notificationGroups(c)
	default:
		recipes := fieldCreateRecipes(p.recipe)
		selected := p.variants[p.recipe]
		for _, d := range recipes {
			if d.id == selected {
				p.drawCreate(c, h, native, d)
				return
			}
		}
	}
}
func (p *fieldCompositionPreviewState) profile(c *ui.Context) {
	fieldSet(c, "Profile", "Fill in your profile information.", false, func(child *ui.Context) {
		fieldCompositionGroup(child, func() {
			control.FieldWith(child.Key("profile-name"), control.FieldProps{Orientation: control.FieldResponsive, Required: true, Error: p.profileError}, func(f *control.FieldScope) {
				f.Content(child, func() {
					f.Label(child, control.FieldLabelProps{Text: "Name", Label: "Name *"}, nil)
					f.Description(child, "Provide your full name for identification")
				})
				f.Control(child, func() ui.Element {
					return control.Input(child, p.text("profile:name", ""), control.InputProps{Placeholder: "Evil Rabbit", Handle: &p.profileName, OnSubmit: func(string) { p.saveProfile(child) }}).Grow(1).MinWidth(0)
				})
			})
			control.FieldWith(child, control.FieldProps{Orientation: control.FieldResponsive}, func(_ *control.FieldScope) {
				if control.Button(child, "Submit", control.ButtonProps{Label: "Submit profile"}).Clicked() {
					p.saveProfile(child)
				}
				if control.Button(child, "Cancel", control.ButtonProps{Label: "Cancel profile", Variant: control.Outline}).Clicked() {
					*p.text("profile:name", "") = p.profileSaved
					p.profileError = ""
					p.status = "Profile draft restored"
				}
			})
		})
	})
}
func (p *fieldCompositionPreviewState) saveProfile(c *ui.Context) {
	value := strings.TrimSpace(*p.text("profile:name", ""))
	if value == "" {
		p.profileError = "Enter your name"
		p.profileName.Focus(c)
		return
	}
	p.profileError = ""
	p.profileSaved = value
	p.status = "Saved local profile: " + value
}
func (p *fieldCompositionPreviewState) compute(c *ui.Context) {
	fieldSet(c, "Compute Environment", "Select the compute environment for your cluster.", true, func(child *ui.Context) {
		selected := p.text("compute:environment", "kubernetes")
		ui.RadioGroup(child, func() {
			for _, option := range []struct{ value, label, description string }{{"kubernetes", "Kubernetes", "Run GPU workloads on a K8s cluster."}, {"vm", "Virtual Machine", "Access a cluster to run GPU workloads."}} {
				control.FieldWith(child.Key(option.value), control.FieldProps{Orientation: control.FieldHorizontal}, func(f *control.FieldScope) {
					f.Label(child, control.FieldLabelProps{Label: option.label, Choice: true, Selected: *selected == option.value}, func() {
						ui.Row(child).FillWidth().MinWidth(0).Gap(12).Children(func() {
							f.Content(child, func() { control.FieldTitle(child, option.label); f.Description(child, option.description) })
							f.Control(child, func() ui.Element { return control.Radio(child, selected, option.value, "", false) })
						})
					})
				})
			}
		})
	})
}
func (p *fieldCompositionPreviewState) desktopChoices(c *ui.Context) {
	fieldSet(c, "Show these items on the desktop", "Select the items you want to show on the desktop.", true, func(child *ui.Context) {
		ui.Column(child).FillWidth().Gap(12).Children(func() {
			for _, label := range []string{"Hard disks", "External disks", "CDs, DVDs, and iPods", "Connected servers"} {
				p.choice(child, label, label, "", "checkbox", "", false, false, false, false, false)
			}
		})
	})
	control.FieldSeparator(c, "")
	p.choice(c, "sync", "Sync Desktop & Documents folders", "Your Desktop & Documents folders are being synced with iCloud Drive. You can access them from other devices.", "checkbox", "", true, false, false, false, false)
}
func (p *fieldCompositionPreviewState) notificationGroups(c *ui.Context) {
	fieldSet(c, "Responses", "Get notified when ChatGPT responds to requests that take time, like research or image generation.", true, func(child *ui.Context) {
		p.choice(child, "responses", "Push notifications", "", "checkbox", "", true, true, false, false, false)
	})
	control.FieldSeparator(c, "")
	fieldSet(c, "Tasks", "Get notified when tasks you've created have updates.", true, func(child *ui.Context) {
		if control.Button(child, "Manage tasks", control.ButtonProps{Variant: control.Link, Size: control.Small}).Clicked() {
			p.status = "Opened local task preferences"
		}
		ui.Column(child).Gap(12).Children(func() {
			p.choice(child, "task-push", "Push notifications", "", "checkbox", "", false, false, false, false, false)
			p.choice(child, "task-email", "Email notifications", "", "checkbox", "", false, false, false, false, false)
		})
	})
}
func (p *fieldCompositionPreviewState) radioGroup(c *ui.Context, key, title, description string, options []control.Option, initial string, disabled, invalid bool) {
	selected := p.text(p.recipe+":"+key, initial)
	fieldSet(c, title, description, true, func(child *ui.Context) {
		ui.RadioGroup(child, func() {
			for _, option := range options {
				control.FieldWith(child.Key(option.Value), control.FieldProps{Orientation: control.FieldHorizontal, Disabled: disabled, Invalid: invalid}, func(f *control.FieldScope) {
					f.Control(child, func() ui.Element {
						return control.RadioWith(child, selected, option.Value, "", control.RadioProps{Disabled: disabled, Invalid: invalid})
					})
					f.Label(child, control.FieldLabelProps{Text: option.Label, Disabled: disabled}, nil)
				})
			}
		}).Gap(12)
	})
}

var fieldDepartments = []control.Option{{Value: "engineering", Label: "Engineering"}, {Value: "design", Label: "Design"}, {Value: "marketing", Label: "Marketing"}, {Value: "sales", Label: "Sales"}, {Value: "support", Label: "Customer Support"}, {Value: "hr", Label: "Human Resources"}, {Value: "finance", Label: "Finance"}, {Value: "operations", Label: "Operations"}}
var fieldCountries = []control.Option{{Value: "us", Label: "United States"}, {Value: "uk", Label: "United Kingdom"}, {Value: "ca", Label: "Canada"}}
var fieldBasicOptions = []control.Option{{Value: "option1", Label: "Option 1"}, {Value: "option2", Label: "Option 2"}, {Value: "option3", Label: "Option 3"}}

func fieldCreateRecipes(recipe string) []fieldRecipe {
	switch recipe {
	case "create-input":
		return []fieldRecipe{
			{id: "basic", label: "Basic Input", placeholder: "Enter text"}, {id: "description", label: "Input with Description", placeholder: "Enter your username", description: "Choose a unique username for your account."},
			{id: "email", label: "Email Address", placeholder: "email@example.com", description: "We'll never share your email with anyone.", before: true}, {id: "required", label: "Required Field", placeholder: "This field is required", description: "This field must be filled out.", required: true},
			{id: "disabled-input", label: "Disabled Input", placeholder: "Cannot edit", description: "This field is currently disabled.", disabled: true}, {id: "badge", label: "Input with Badge", placeholder: "Enter value", badge: true},
			{id: "invalid", label: "Invalid Input", placeholder: "This field has an error", description: "This field contains validation errors.", invalid: true}, {id: "disabled-field", label: "Disabled Field", placeholder: "Cannot edit", description: "This field is currently disabled.", disabled: true},
		}
	case "create-textarea":
		return []fieldRecipe{{id: "basic", label: "Basic Textarea", kind: "textarea", placeholder: "Enter your message"}, {id: "comments", label: "Comments", kind: "textarea", minHeight: 100, placeholder: "Share your thoughts...", description: "Maximum 500 characters allowed."}, {id: "bio", label: "Bio", kind: "textarea", minHeight: 120, placeholder: "I am a...", description: "Tell us about yourself in a few sentences.", before: true}, {id: "message", label: "Message", kind: "textarea", placeholder: "Enter your message", description: "Enter your message so it is long enough to test the layout."}, {id: "invalid", label: "Invalid Textarea", kind: "textarea", placeholder: "This field has an error", description: "This field contains validation errors.", invalid: true}, {id: "disabled", label: "Disabled Field", kind: "textarea", placeholder: "Cannot edit", description: "This field is currently disabled.", disabled: true}}
	case "create-select", "create-native":
		kind := "select"
		if recipe == "create-native" {
			kind = "native"
		}
		basic := "Basic Select"
		invalid := "Invalid Select"
		if kind == "native" {
			basic = "Basic Native Select"
			invalid = "Invalid Native Select"
		}
		out := []fieldRecipe{{id: "basic", label: basic, kind: kind, placeholder: "Choose an option", options: fieldBasicOptions}, {id: "country", label: "Country", kind: kind, placeholder: "Select your country", options: fieldCountries, description: "Select the country where you currently reside."}, {id: "timezone", label: "Timezone", kind: kind, placeholder: "Select timezone", options: []control.Option{{Value: "utc", Label: "UTC"}, {Value: "est", Label: "Eastern Time"}, {Value: "pst", Label: "Pacific Time"}}, description: "Choose your local timezone for accurate scheduling.", before: true}}
		if kind == "native" {
			out = append(out, fieldRecipe{id: "grouped", label: "Grouped Options", kind: kind, placeholder: "Select a region", description: "Native select with grouped options using optgroup.", options: []control.Option{{Value: "us", Label: "United States", Group: "North America"}, {Value: "ca", Label: "Canada", Group: "North America"}, {Value: "mx", Label: "Mexico", Group: "North America"}, {Value: "uk", Label: "United Kingdom", Group: "Europe"}, {Value: "fr", Label: "France", Group: "Europe"}, {Value: "de", Label: "Germany", Group: "Europe"}}})
		}
		return append(out, fieldRecipe{id: "invalid", label: invalid, kind: kind, placeholder: "This field has an error", description: "This field contains validation errors.", options: fieldBasicOptions, invalid: true}, fieldRecipe{id: "disabled", label: "Disabled Field", kind: kind, placeholder: "Cannot select", description: "This field is currently disabled.", options: fieldBasicOptions, disabled: true})
	case "create-checkbox":
		return []fieldRecipe{{id: "basic", label: "I agree to the terms and conditions", kind: "checkbox"}, {id: "right", label: "Accept terms and conditions", kind: "checkbox", right: true}, {id: "description", label: "Subscribe to newsletter", kind: "checkbox", description: "Receive weekly updates about new features and promotions."}, {id: "card", label: "Enable Touch ID", kind: "checkbox", description: "Enable Touch ID to quickly unlock your device.", choice: true}, {id: "preferences", label: "Preferences", kind: "preferences"}, {id: "invalid", label: "Invalid checkbox", kind: "checkbox", invalid: true}, {id: "disabled", label: "Disabled checkbox", kind: "checkbox", disabled: true}}
	case "create-radio":
		return []fieldRecipe{{id: "subscription", label: "Subscription Plan"}, {id: "battery", label: "Battery Level"}, {id: "description", label: "Enable Touch ID", description: "Enable Touch ID to quickly unlock your device."}, {id: "card", label: "Touch ID choice card", description: "Enable Touch ID to quickly unlock your device.", choice: true}, {id: "invalid", label: "Invalid Radio Group", invalid: true}, {id: "disabled", label: "Disabled Radio Group", disabled: true}}
	case "create-switch":
		return []fieldRecipe{{id: "airplane", label: "Airplane Mode", kind: "switch", description: "Turn on airplane mode to disable all connections.", right: true}, {id: "dark", label: "Dark Mode", kind: "switch", right: true}, {id: "marketing", label: "Marketing Emails", kind: "switch", description: "Receive emails about new products, features, and more."}, {id: "privacy", label: "Privacy Settings", kind: "privacy"}, {id: "invalid", label: "Invalid Switch", kind: "switch", description: "This switch has validation errors.", right: true, invalid: true}, {id: "disabled", label: "Disabled Switch", kind: "switch", description: "This switch is currently disabled.", right: true, disabled: true}}
	case "create-slider":
		return []fieldRecipe{{id: "volume", label: "Volume", kind: "slider", values: []float64{50}, max: 100, step: 1}, {id: "brightness", label: "Screen Brightness", kind: "slider", values: []float64{75}, max: 100, step: 5, description: "Current brightness"}, {id: "quality", label: "Video Quality", kind: "slider", values: []float64{720}, min: 360, max: 1080, step: 360, description: "Higher quality uses more bandwidth.", before: true}, {id: "temperature", label: "Temperature Range", kind: "slider", values: []float64{.3, .7}, max: 1, step: .1, description: "Range"}, {id: "price", label: "Price Range", kind: "slider", values: []float64{25, 75}, max: 100, step: 5, description: "Range"}, {id: "balance", label: "Color Balance", kind: "slider", values: []float64{10, 20, 70}, max: 100, step: 10, description: "Balance"}, {id: "invalid", label: "Invalid Slider", kind: "slider", values: []float64{30}, max: 100, step: 1, description: "This slider has validation errors.", invalid: true}, {id: "disabled", label: "Disabled Slider", kind: "slider", values: []float64{50}, max: 100, step: 1, description: "This slider is currently disabled.", disabled: true}}
	case "create-otp":
		return []fieldRecipe{{id: "basic", label: "Verification Code", kind: "otp", length: 6}, {id: "description", label: "Enter OTP", kind: "otp", length: 6, description: "Enter the 6-digit code sent to your email."}, {id: "separator", label: "Two-Factor Authentication", kind: "otp", length: 6, groups: []int{3, 3}, description: "Enter the code from your authenticator app."}, {id: "pin", label: "PIN Code", kind: "otp", length: 4, description: "Enter your 4-digit PIN (numbers only)."}, {id: "invalid", label: "Invalid OTP", kind: "otp", length: 6, description: "This OTP field contains validation errors.", invalid: true}, {id: "disabled", label: "Disabled OTP", kind: "otp", length: 6, description: "This OTP field is currently disabled.", disabled: true}}
	case "create-horizontal":
		return []fieldRecipe{{id: "username", label: "Username", placeholder: "johndoe", description: "Enter your preferred username."}, {id: "bio", label: "Bio", kind: "textarea", placeholder: "Tell us about yourself...", description: "Write a short description about yourself."}, {id: "notifications", label: "Email Notifications", kind: "switch", description: "Receive email updates about your account."}, {id: "fruit", label: "Favorite Fruit", kind: "select", placeholder: "Select a fruit", description: "Choose your favorite fruit.", options: []control.Option{{Value: "apple", Label: "Apple"}, {Value: "banana", Label: "Banana"}, {Value: "orange", Label: "Orange"}}}, {id: "country", label: "Country", kind: "native", placeholder: "Select a country", description: "Select your country.", options: fieldCountries}, {id: "volume", label: "Volume", kind: "slider", description: "Adjust the volume level.", values: []float64{50}, max: 100, step: 1}}
	}
	return nil
}
func (p *fieldCompositionPreviewState) drawCreate(c *ui.Context, h *desktop.Host, native bool, d fieldRecipe) {
	switch p.recipe {
	case "create-checkbox":
		if d.kind == "preferences" {
			fieldSet(c, "Preferences", "Select all that apply to customize your experience.", true, func(child *ui.Context) {
				control.FieldGroup(child, func() {
					for _, label := range []string{"Dark mode", "Compact view", "Enable notifications"} {
						p.choice(child, label, label, "", "checkbox", "", false, false, false, false, false)
					}
				}).Gap(12)
			})
		} else {
			p.choice(c, d.id, d.label, d.description, "checkbox", "", d.id == "basic", d.disabled, d.invalid, d.right, d.choice)
		}
	case "create-radio":
		switch d.id {
		case "subscription":
			p.radioGroup(c, d.id, d.label, "", []control.Option{{Value: "free", Label: "Free Plan"}, {Value: "pro", Label: "Pro Plan"}, {Value: "enterprise", Label: "Enterprise"}}, "free", false, false)
		case "battery":
			p.radioGroup(c, d.id, d.label, "Choose your preferred battery level.", []control.Option{{Value: "high", Label: "High"}, {Value: "medium", Label: "Medium"}, {Value: "low", Label: "Low"}}, "", false, false)
		case "invalid", "disabled":
			p.radioGroup(c, d.id, d.label, "", []control.Option{{Value: "one", Label: strings.Fields(d.label)[0] + " Option 1"}, {Value: "two", Label: strings.Fields(d.label)[0] + " Option 2"}}, "", d.disabled, d.invalid)
		default:
			ui.RadioGroup(c, func() {
				p.choice(c, d.id, "Enable Touch ID", d.description, "radio", "option1", false, false, false, false, d.choice)
			})
		}
	case "create-switch":
		if d.kind == "privacy" {
			control.FieldWith(c, control.FieldProps{}, func(f *control.FieldScope) {
				f.Label(c, control.FieldLabelProps{Text: d.label}, nil)
				f.Description(c, "Manage your privacy preferences.")
				p.choice(c, "profile", "Make profile visible to others", "", "switch", "", true, false, false, false, false)
				p.choice(c, "email", "Show email on profile", "", "switch", "", false, false, false, false, false)
			})
		} else {
			p.choice(c, d.id, d.label, d.description, "switch", "", false, d.disabled, d.invalid, d.right, false)
		}
	default:
		p.basic(c, h, native, d)
	}
}

type fieldPaymentLabels struct{ payment, secure, name, number, numberHelp, month, year, cvv, billing, billingHelp, shipping, comments, commentsPlaceholder, submit, cancel string }

func (p *fieldCompositionPreviewState) paymentLabels() fieldPaymentLabels {
	if p.recipe == "rtl" {
		switch p.language {
		case "ar":
			return fieldPaymentLabels{"طريقة الدفع", "جميع المعاملات آمنة ومشفرة", "الاسم على البطاقة", "رقم البطاقة", "أدخل رقم البطاقة المكون من 16 رقمًا", "الشهر", "السنة", "CVV", "عنوان الفوترة", "عنوان الفوترة المرتبط بطريقة الدفع الخاصة بك", "نفس عنوان الشحن", "تعليقات", "أضف أي تعليقات إضافية", "إرسال", "إلغاء"}
		case "he":
			return fieldPaymentLabels{"אמצעי תשלום", "כל העסקאות מאובטחות ומוצפנות", "שם על הכרטיס", "מספר כרטיס", "הזן את מספר הכרטיס בן 16 הספרות שלך", "חודש", "שנה", "CVV", "כתובת חיוב", "כתובת החיוב המשויכת לאמצעי התשלום שלך", "זהה לכתובת המשלוח", "הערות", "הוסף הערות נוספות", "שלח", "בטל"}
		}
	}
	return fieldPaymentLabels{"Payment Method", "All transactions are secure and encrypted", "Name on Card", "Card Number", "Enter your 16-digit card number", "Month", "Year", "CVV", "Billing Address", "The billing address associated with your payment method", "Same as shipping address", "Comments", "Add any additional comments", "Submit", "Cancel"}
}
func (p *fieldCompositionPreviewState) payment(c *ui.Context, h *desktop.Host, native bool) {
	labels := p.paymentLabels()
	fieldSet(c, labels.payment, labels.secure, false, func(child *ui.Context) {
		fieldCompositionGroup(child, func() {
			p.basic(child, h, native, fieldRecipe{id: "card-name", label: labels.name, placeholder: "Evil Rabbit", required: true})
			p.basic(child, h, native, fieldRecipe{id: "card-number", label: labels.number, placeholder: "1234 5678 9012 3456", description: labels.numberHelp, required: true})
			months := make([]control.Option, 12)
			years := make([]control.Option, 6)
			for i := range months {
				v := fmt.Sprintf("%02d", i+1)
				label := v
				if p.recipe == "rtl" && p.language == "ar" {
					label = strings.NewReplacer("0", "٠", "1", "١", "2", "٢", "3", "٣", "4", "٤", "5", "٥", "6", "٦", "7", "٧", "8", "٨", "9", "٩").Replace(v)
				}
				months[i] = control.Option{Value: v, Label: label}
			}
			for i := range years {
				v := fmt.Sprint(2024 + i)
				years[i] = control.Option{Value: v, Label: v}
			}
			ui.Row(child).FillWidth().MinWidth(0).Gap(16).Children(func() {
				for _, d := range []fieldRecipe{{id: "month", label: labels.month, kind: "select", placeholder: "MM", options: months}, {id: "year", label: labels.year, kind: "select", placeholder: "YYYY", options: years}, {id: "cvv", label: labels.cvv, placeholder: "123", required: true}} {
					ui.Column(child.Key(d.id)).Basis(0).Grow(1).MinWidth(0).Children(func() { p.basic(child, h, native, d) })
				}
			})
		})
	})
	control.FieldSeparator(c, "")
	fieldSet(c, labels.billing, labels.billingHelp, false, func(child *ui.Context) {
		p.choice(child, "shipping", labels.shipping, "", "checkbox", "", true, false, false, false, false)
	})
	p.basic(c, h, native, fieldRecipe{id: "comments", label: labels.comments, kind: "textarea", placeholder: labels.commentsPlaceholder})
	control.FieldWith(c, control.FieldProps{Orientation: control.FieldHorizontal}, func(_ *control.FieldScope) {
		submitLabel, cancelLabel := labels.submit, labels.cancel
		if p.recipe != "rtl" {
			submitLabel = "Submit payment fields"
			cancelLabel = "Cancel payment fields"
		}
		if control.Button(c, labels.submit, control.ButtonProps{Label: submitLabel}).Clicked() {
			read := func(key string) string { return *p.text(p.recipe+":"+key, "") }
			if strings.TrimSpace(read("card-name")) == "" || strings.TrimSpace(read("card-number")) == "" || strings.TrimSpace(read("cvv")) == "" {
				p.status = "Complete the required local payment fields"
			} else {
				p.paymentSaved[p.recipe] = fieldPaymentSnapshot{Name: read("card-name"), Number: read("card-number"), Month: read("month"), Year: read("year"), CVV: read("cvv"), Comments: read("comments"), Shipping: *p.checked(p.recipe+":shipping", true)}
				p.status = "Saved local payment example; no transaction was sent"
			}
		}
		if control.Button(c, labels.cancel, control.ButtonProps{Label: cancelLabel, Variant: control.Outline}).Clicked() {
			s, saved := p.paymentSaved[p.recipe]
			if !saved {
				s.Shipping = true
			}
			for key, value := range map[string]string{"card-name": s.Name, "card-number": s.Number, "month": s.Month, "year": s.Year, "cvv": s.CVV, "comments": s.Comments} {
				*p.text(p.recipe+":"+key, "") = value
			}
			*p.checked(p.recipe+":shipping", true) = s.Shipping
			p.status = "Payment draft restored locally"
		}
	})
}
