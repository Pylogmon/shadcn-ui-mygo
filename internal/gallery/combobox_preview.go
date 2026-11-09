package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type comboboxPreviewState struct {
	initialized                                                                      bool
	framework, country, city, plan, region, requiredCountry                          control.ComboboxValue
	frameworkQuery, countryQuery, cityQuery, technologyQuery, planQuery, regionQuery string
	requiredCountryQuery                                                             string
	technologies                                                                     []string
	frameworks, countries, cities, technologyOptions                                 []control.Option
}

func (s *comboboxPreviewState) initialize() {
	if s.initialized {
		return
	}
	s.initialized = true
	s.framework = control.ComboboxValue{Value: "mygo", Valid: true}
	s.frameworks = []control.Option{
		{Value: "mygo", Label: "MyGo"}, {Value: "fyne", Label: "Fyne"},
		{Value: "gio", Label: "Gio"}, {Value: "wails", Label: "Wails"},
	}
	s.countries = []control.Option{
		{Value: "jp", Label: "Japan", Description: "Tokyo · Asia", Group: "Asia"},
		{Value: "sg", Label: "Singapore", Description: "Singapore · Asia", Group: "Asia"},
		{Separator: true},
		{Value: "de", Label: "Germany", Description: "Berlin · Europe", Group: "Europe"},
		{Value: "fr", Label: "France", Description: "Paris · Europe", Group: "Europe"},
		{Value: "no", Label: "Norway", Description: "Oslo · Currently unavailable", Group: "Europe", Disabled: true},
	}
	s.cities = []control.Option{
		{Value: "london", Label: "London"}, {Value: "new-york", Label: "New York"},
		{Value: "san-francisco", Label: "San Francisco"}, {Value: "tokyo", Label: "Tokyo"},
	}
	s.technologies = []string{"go", "typescript"}
	s.technologyOptions = []control.Option{
		{Value: "go", Label: "Go"}, {Value: "typescript", Label: "TypeScript"},
		{Value: "rust", Label: "Rust"}, {Value: "swift", Label: "Swift"},
		{Value: "kotlin", Label: "Kotlin"},
	}
	s.plan = control.ComboboxValue{Value: "pro", Valid: true}
	s.region = control.ComboboxValue{Value: "eu", Valid: true}
}

func (s *State) comboboxPreview(c *ui.Context) {
	p := &s.comboboxes
	p.initialize()
	columns(c, func() {
		display.Card(c.Key("combobox-basic"), func() {
			s.cardHeader(c, "Combobox", "Search a framework, choose one, or clear your selection.")
			control.Field(c, control.FieldProps{Label: "Framework search"}, func() ui.Element {
				return control.ComboboxWith(c, &p.framework, &p.frameworkQuery, control.ComboboxWithProps{
					Label: "Framework search", Placeholder: "Choose a framework", ShowClear: true, Options: p.frameworks,
				})
			})
			ui.Text(c, "Selected: "+comboboxChoiceLabel(p.frameworks, p.framework)).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
		}).Role(ui.RoleGroup).Label("Framework selection")
		display.Card(c.Key("combobox-groups"), func() {
			s.cardHeader(c, "Grouped suggestions", "Find a destination by country or city.")
			control.Field(c, control.FieldProps{Label: "Destination country"}, func() ui.Element {
				return control.ComboboxWith(c, &p.country, &p.countryQuery, control.ComboboxWithProps{
					Label: "Destination country", Placeholder: "Search destinations", EmptyText: "No destinations found.",
					ShowClear: true, AutoHighlight: true, Options: p.countries,
					Filter: func(option control.Option, query string) bool {
						return strings.Contains(strings.ToLower(option.Label+" "+option.Description), strings.ToLower(strings.TrimSpace(query)))
					},
					ItemContent: func(child *ui.Context, option control.Option) {
						ui.Column(child).MinWidth(0).Gap(2).Children(func() {
							ui.Text(child, option.Label).FontSize(14).SingleLine()
							ui.Text(child, option.Description).FontSize(12).TextColor(theme.TokensOf(child).MutedForeground).SingleLine()
						})
					},
				})
			})
		}).Role(ui.RoleGroup).Label("Destination selection")
	}, func() {
		display.Card(c.Key("combobox-popup"), func() {
			s.cardHeader(c, "Search in a popup", "Choose an office with the search field inside the menu.")
			control.Field(c, control.FieldProps{Label: "Office city"}, func() ui.Element {
				return control.ComboboxWith(c, &p.city, &p.cityQuery, control.ComboboxWithProps{
					Label: "Office city", Placeholder: "Select an office", InputPlaceholder: "Search cities…",
					PopupInput: true, AutoHighlight: true, Options: p.cities,
				})
			})
		}).Role(ui.RoleGroup).Label("Office selection")
		display.Card(c.Key("combobox-multiple"), func() {
			s.cardHeader(c, "Multiple selection", "Add technologies or use the arrow keys to move between selected items.")
			control.Field(c, control.FieldProps{Label: "Project technologies"}, func() ui.Element {
				return control.MultiComboboxWith(c, &p.technologies, &p.technologyQuery, control.ComboboxWithProps{
					Label: "Project technologies", Placeholder: "Add technology", Options: p.technologyOptions, AutoHighlight: true,
				})
			})
		}).Role(ui.RoleGroup).Label("Technology selection")
	})
	display.Card(c.Key("combobox-states"), func() {
		s.cardHeader(c, "Selection states", "Required choices and editing permissions.")
		columns(c, func() {
			control.Field(c, control.FieldProps{Label: "Billing plan", Description: "Read only"}, func() ui.Element {
				return control.ComboboxWith(c, &p.plan, &p.planQuery, control.ComboboxWithProps{
					Label: "Billing plan", ReadOnly: true, Options: []control.Option{{Value: "free", Label: "Free"}, {Value: "pro", Label: "Pro"}},
				})
			})
		}, func() {
			control.Field(c, control.FieldProps{Label: "Managed region", Description: "Managed by your organization"}, func() ui.Element {
				return control.ComboboxWith(c, &p.region, &p.regionQuery, control.ComboboxWithProps{
					Label: "Managed region", Disabled: true, Options: []control.Option{{Value: "eu", Label: "Europe"}, {Value: "us", Label: "United States"}},
				})
			})
		}, func() {
			errorText := ""
			if !p.requiredCountry.Valid {
				errorText = "Choose a destination."
			}
			control.Field(c, control.FieldProps{Label: "Required destination", Required: true, Error: errorText}, func() ui.Element {
				return control.ComboboxWith(c, &p.requiredCountry, &p.requiredCountryQuery, control.ComboboxWithProps{
					Label: "Required destination", Placeholder: "Select a country", Invalid: errorText != "", Options: p.countries,
				})
			})
		})
	}).Role(ui.RoleGroup).Label("Selection permissions")
}

func comboboxChoiceLabel(options []control.Option, selected control.ComboboxValue) string {
	if selected.Valid {
		for _, option := range options {
			if !option.Separator && option.Value == selected.Value {
				return option.Label
			}
		}
		return selected.Value
	}
	return "None"
}
