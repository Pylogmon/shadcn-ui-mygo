// Run the fixed input examples using native text, calendar, time and file controls.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type appState struct {
	appearance                                                      theme.Controller
	section                                                         string
	email, username, name, password, phone, website, query, address string
	currency, country, dialCode                                     string
	number                                                          float64
	date, clock                                                     time.Time
	paths                                                           []string
	status                                                          string
	attempted                                                       bool
	picker                                                          control.FilePicker
	window                                                          *mygo.Window
}

func newState(dark bool, section string) *appState {
	s := &appState{appearance: theme.Controller{Preset: "neutral", Mode: theme.LightMode}, section: section, currency: "usd", country: "us", dialCode: "+1", number: 123, date: time.Date(2026, 11, 10, 0, 0, 0, 0, time.Local), clock: time.Date(2026, 11, 10, 15, 30, 0, 0, time.Local)}
	if dark {
		s.appearance.Mode = theme.DarkMode
	}
	return s
}
func emailError(value string) string {
	if strings.TrimSpace(value) == "" {
		return "Enter your email address."
	}
	address, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil || address.Address != strings.TrimSpace(value) {
		return "Enter a valid email address."
	}
	return ""
}
func websiteError(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "Enter a full http or https URL."
	}
	return ""
}
func (s *appState) view(c *ui.Context) {
	s.appearance.Apply(c)
	t := theme.TokensOf(c)
	ui.Column(c).Fill().Padding(24).Gap(16).Background(t.Background).TextColor(t.Foreground).Children(func() {
		display.Typography(c, "Input examples", display.H2)
		ui.Text(c, "Edit the fields, choose a date and time, or attach a file.").TextColor(t.MutedForeground)
		ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
			control.NativeSelect(c, &s.section, control.SelectProps{Label: "Input example", Options: []control.Option{{Value: "types", Label: "Input Types"}, {Value: "states", Label: "Basic and Field states"}, {Value: "composition", Label: "With controls"}, {Value: "form", Label: "Form"}}}).Width(230)
			dark := s.appearance.Mode == theme.DarkMode
			if control.Switch(c, &dark, control.SwitchProps{Label: "Dark appearance"}).Changed() {
				if dark {
					s.appearance.Mode = theme.DarkMode
				} else {
					s.appearance.Mode = theme.LightMode
				}
			}
		})
		ui.Scroll(c).Fill().Children(func() {
			ui.Column(c).FillWidth().MaxWidth(740).Gap(20).Children(func() {
				switch s.section {
				case "states":
					s.states(c)
				case "composition":
					s.composition(c)
				case "form":
					s.form(c)
				default:
					s.types(c)
				}
				if s.status != "" {
					ui.Text(c, s.status).Role(ui.RoleStatus)
				}
			})
		})
	})
}
func field(c *ui.Context, label string, editor func() ui.Element) {
	control.Field(c, control.FieldProps{Label: label}, editor)
}
func (s *appState) types(c *ui.Context) {
	display.Typography(c, "Input Types", display.H3)
	ui.Row(c).FillWidth().Gap(24).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Basis(0).Grow(1).MinWidth(0).Gap(20).Children(func() {
			field(c, "Password", func() ui.Element {
				return control.Input(c, &s.password, control.InputProps{Password: true, Placeholder: "Password"})
			})
			field(c, "Phone", func() ui.Element {
				return control.Input(c, &s.phone, control.InputProps{Placeholder: "+1 (555) 123-4567"})
			})
			control.Field(c, control.FieldProps{Label: "URL", Error: websiteError(s.website)}, func() ui.Element {
				return control.Input(c, &s.website, control.InputProps{Placeholder: "https://example.com"})
			})
			field(c, "Search", func() ui.Element {
				return control.Input(c, &s.query, control.InputProps{Search: true, Placeholder: "Search", OnSubmit: func(v string) { s.status = "Search: " + v }})
			})
		})
		ui.Column(c).Basis(0).Grow(1).MinWidth(0).Gap(20).Children(func() {
			field(c, "Number", func() ui.Element {
				return control.NumberInput(c, &s.number, control.NumberInputProps{Placeholder: "123"})
			})
			field(c, "Date", func() ui.Element { return control.DateInput(c, &s.date, control.DateInputProps{}) })
			field(c, "Time", func() ui.Element { return control.TimeInput(c, &s.clock, control.TimeInputProps{}) })
			control.Field(c, control.FieldProps{Label: "File", Description: "Choose an attachment."}, func() ui.Element {
				return control.FileInput(c, &s.paths, control.FileInputProps{Parent: s.window, Picker: s.picker, OnError: func(error) { s.status = "The file could not be opened. Try again." }})
			})
		})
	})
	if control.Button(c, "Review values", control.ButtonProps{Variant: control.Outline}).Clicked() {
		s.status = fmt.Sprintf("Number %g · %s · %s · %d attachment(s)", s.number, s.date.Format("2006-01-02"), s.clock.Format("15:04"), len(s.paths))
	}
}
func (s *appState) states(c *ui.Context) {
	display.Typography(c, "Basic and Field states", display.H3)
	control.Input(c, &s.email, control.InputProps{Label: "Basic email", Placeholder: "Email"})
	control.Input(c, &s.username, control.InputProps{Label: "Invalid input", Placeholder: "Error", Invalid: true})
	field(c, "Email", func() ui.Element {
		return control.Input(c, &s.email, control.InputProps{Placeholder: "name@example.com"})
	})
	control.Field(c, control.FieldProps{Label: "Username", Description: "Choose a unique username for your account."}, func() ui.Element {
		return control.Input(c, &s.username, control.InputProps{Placeholder: "Enter your username"})
	})
	field(c, "Disabled email", func() ui.Element {
		return control.Input(c, &s.email, control.InputProps{Placeholder: "Email", Disabled: true})
	})
}
func (s *appState) composition(c *ui.Context) {
	display.Typography(c, "With controls", display.H3)
	control.Field(c, control.FieldProps{Label: "Amount", Description: "Enter an amount and choose its currency."}, func() ui.Element {
		return control.InputGroup(c, control.InputGroupProps{Prefix: "$", InlineEnd: func() {
			control.NativeSelect(c, &s.currency, control.SelectProps{Label: "Currency", Options: []control.Option{{Value: "usd", Label: "USD"}, {Value: "eur", Label: "EUR"}, {Value: "gbp", Label: "GBP"}}}).Width(84)
		}}, func() ui.Element { return control.NumberInput(c, &s.number, control.NumberInputProps{Step: .5}) })
	})
	ui.Row(c).Gap(8).Children(func() {
		control.Input(c, &s.query, control.InputProps{Label: "Search text", Search: true, Placeholder: "Search...", OnSubmit: func(v string) { s.status = "Search: " + v }})
		if control.Button(c, "Search", control.ButtonProps{}).Clicked() {
			s.status = "Search: " + s.query
		}
	})
	ui.Row(c).Gap(8).Children(func() {
		control.Input(c, &s.phone, control.InputProps{Label: "Phone number", Placeholder: "(555) 123-4567"})
		control.NativeSelect(c, &s.dialCode, control.SelectProps{Label: "Country code", Options: []control.Option{{Value: "+1", Label: "+1"}, {Value: "+44", Label: "+44"}, {Value: "+46", Label: "+46"}}}).Width(90)
	})
	control.Field(c, control.FieldProps{Label: "Picture", Description: "Select a picture to upload."}, func() ui.Element {
		return control.FileInput(c, &s.paths, control.FileInputProps{Picker: s.picker, Parent: s.window, Filters: []mygo.FileFilter{{Name: "Images", Extensions: []string{"png", "jpg", "jpeg"}}}})
	})
}
func (s *appState) form(c *ui.Context) {
	display.Typography(c, "Form", display.H3)
	control.FieldGroup(c, func() {
		nameError, emailErr := "", ""
		if s.attempted {
			if strings.TrimSpace(s.name) == "" {
				nameError = "Enter your name."
			}
			emailErr = emailError(s.email)
		}
		control.Field(c, control.FieldProps{Label: "Name", Error: nameError}, func() ui.Element { return control.Input(c, &s.name, control.InputProps{Placeholder: "John Doe"}) })
		control.Field(c, control.FieldProps{Label: "Email", Description: "We'll never share your email with anyone.", Error: emailErr}, func() ui.Element {
			return control.Input(c, &s.email, control.InputProps{Placeholder: "john@example.com"})
		})
		ui.Row(c).Gap(16).Children(func() {
			field(c, "Phone", func() ui.Element {
				return control.Input(c, &s.phone, control.InputProps{Placeholder: "+1 (555) 123-4567"})
			})
			field(c, "Country", func() ui.Element {
				return control.NativeSelect(c, &s.country, control.SelectProps{Options: []control.Option{{Value: "us", Label: "United States"}, {Value: "uk", Label: "United Kingdom"}, {Value: "ca", Label: "Canada"}}})
			})
		})
		field(c, "Address", func() ui.Element { return control.Input(c, &s.address, control.InputProps{Placeholder: "123 Main St"}) })
		ui.Row(c).Gap(8).Children(func() {
			if control.Button(c, "Cancel", control.ButtonProps{Variant: control.Outline}).Clicked() {
				s.name = ""
				s.email = ""
				s.phone = ""
				s.address = ""
				s.attempted = false
				s.status = "Form cleared."
			}
			if control.Button(c, "Submit", control.ButtonProps{}).Clicked() {
				s.attempted = true
				if strings.TrimSpace(s.name) == "" || emailError(s.email) != "" {
					s.status = "Check the highlighted fields."
					c.Invalidate()
				} else {
					s.status = "Submitted for " + s.name + " · " + s.email
				}
			}
		})
	})
}
func main() {
	snapshot := flag.String("snapshot", "", "render example as PNG")
	dark := flag.Bool("dark", false, "dark appearance")
	section := flag.String("section", "types", "types, states, composition, or form")
	flag.Parse()
	s := newState(*dark, *section)
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		if err = png.Encode(f, ui.Render(s.view, 880, 720, 1)); err != nil {
			log.Fatal(err)
		}
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		s.window = mygo.NewWindow(mygo.WindowOptions{Title: "Input examples", Width: 880, Height: 720, Content: ui.View(s.view)})
		s.window.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
