package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type appState struct {
	appearance                                         theme.Controller
	dark, dropdown                                     bool
	system                                             string
	location                                           *time.Location
	selected, month, exactDate, naturalDate, reference time.Time
	exact, natural                                     display.DatePickerInputState
	feedback                                           string
}

func newState() *appState {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		panic(err)
	}
	reference := time.Date(2025, 6, 10, 14, 35, 0, 0, loc)
	selected := reference.AddDate(0, 0, 2)
	return &appState{appearance: theme.Controller{Preset: "blue", Mode: theme.LightMode}, system: "persian", location: loc, selected: selected, month: selected, exactDate: selected, naturalDate: selected, reference: reference, exact: display.DatePickerInputState{Text: formatDate(selected)}, natural: display.DatePickerInputState{Text: "In 2 days"}}
}
func formatDate(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format("January 02, 2006")
}
func (s *appState) parseExact(text string) (time.Time, error) {
	for _, layout := range []string{"January 02, 2006", "2006-01-02"} {
		if date, err := time.ParseInLocation(layout, strings.TrimSpace(text), s.location); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("Use June 12, 2025 or 2025-06-12.")
}

// This deliberately small application parser demonstrates the public callback.
// It neither depends on nor claims the complete chrono-node language grammar.
func (s *appState) parseRelative(text string) (time.Time, error) {
	query := strings.ToLower(strings.TrimSpace(text))
	switch query {
	case "tomorrow":
		return s.reference.AddDate(0, 0, 1), nil
	case "next week":
		return s.reference.AddDate(0, 0, 7), nil
	}
	words := strings.Fields(query)
	if len(words) == 3 && words[0] == "in" && (words[2] == "day" || words[2] == "days") {
		if count, err := strconv.Atoi(words[1]); err == nil && count >= 0 && count <= 365 {
			return s.reference.AddDate(0, 0, count), nil
		}
	}
	if value, err := s.parseExact(text); err == nil {
		return value, nil
	}
	return time.Time{}, fmt.Errorf("Try tomorrow, next week, in 2 days, or an exact date.")
}
func (s *appState) describe(date time.Time) string {
	if date.IsZero() {
		return "No date selected"
	}
	p, err := display.PersianDateOf(date)
	if err != nil {
		return date.Format(time.RFC3339)
	}
	return fmt.Sprintf("%s · Jalali %04d/%02d/%02d", date.Format("2006-01-02 15:04 MST"), p.Year, p.Month, p.Day)
}
func (s *appState) view(c *ui.Context) {
	s.appearance.Mode = theme.LightMode
	if s.dark {
		s.appearance.Mode = theme.DarkMode
	}
	s.appearance.Apply(c)
	t := theme.TokensOf(c)
	ui.Scroll(c).Fill().Background(t.Background).Children(func() {
		ui.Column(c).FillWidth().MaxWidth(1000).MarginX(ui.Auto).Padding(24).Gap(20).Children(func() {
			display.Typography(c, "Calendar and editable dates", display.H2)
			ui.Text(c, "A Persian month model and application-defined date parsers share real dates and local time.").TextColor(t.MutedForeground)
			ui.Row(c).Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
				control.Switch(c, &s.dark, control.SwitchProps{Label: "Dark appearance"})
				control.NativeSelect(c, &s.system, control.SelectProps{Label: "Calendar model", Options: []control.Option{{Value: "persian", Label: "Persian / Jalali"}, {Value: "gregorian", Label: "Gregorian"}}}).Width(200)
				control.Switch(c, &s.dropdown, control.SwitchProps{Label: "Month and year menus"})
			})
			display.Card(c, func() {
				ui.Column(c).Gap(16).Children(func() {
					display.Typography(c, "One instant, two civil calendars", display.H3)
					ui.Row(c).Wrap().Gap(28).AlignItems(ui.Start).Children(func() {
						props := display.CalendarProps{Label: "Civil calendar", Location: s.location, Month: &s.month, Browse: true}
						if s.system == "persian" {
							props.System = display.CalendarPersian
						}
						if s.dropdown {
							props.Caption = display.CalendarCaptionDropdown
							props.StartMonth = time.Date(2020, 1, 1, 0, 0, 0, 0, s.location)
							props.EndMonth = time.Date(2030, 12, 1, 0, 0, 0, 0, s.location)
						}
						display.CalendarWith(c, &s.selected, props)
						ui.Column(c).Gap(10).MaxWidth(440).Children(func() {
							ui.Text(c, "Selected date").FontWeight(500)
							ui.Text(c, s.describe(s.selected)).Role(ui.RoleStatus)
							ui.Text(c, "Changing the calendar or browsing a month keeps the selected instant. Choosing a day keeps its local clock.").TextColor(t.MutedForeground)
						})
					})
				})
			})
			display.Card(c, func() {
				ui.Column(c).Gap(16).Children(func() {
					display.Typography(c, "Editable date recipes", display.H3)
					ui.Row(c).Wrap().Gap(28).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Width(390).Gap(8).Children(func() {
							control.Field(c, control.FieldProps{Label: "Subscription date", Description: "Type June 12, 2025 or 2025-06-12; ArrowDown opens the calendar.", Error: s.exact.Error}, func() ui.Element {
								return display.DatePickerInput(c, &s.exactDate, &s.exact, display.DatePickerInputProps{Input: control.InputProps{Label: "Subscription query", Placeholder: "June 01, 2025"}, TriggerLabel: "Choose subscription date", Calendar: display.CalendarProps{Location: s.location}, Parse: s.parseExact, Format: formatDate, OnChange: func(date time.Time) { s.feedback = "Subscription: " + s.describe(date) }}).Root
							})
							ui.Text(c, s.describe(s.exactDate)).TextColor(t.MutedForeground)
						})
						ui.Column(c).Width(390).Gap(8).Children(func() {
							gap, offset := float32(8), float32(0)
							control.Field(c, control.FieldProps{Label: "Schedule date", Description: "Try tomorrow, next week or in 2 days. Relative dates use June 10, 2025.", Error: s.natural.Error}, func() ui.Element {
								return display.DatePickerInput(c, &s.naturalDate, &s.natural, display.DatePickerInputProps{Input: control.InputProps{Label: "Schedule query", Placeholder: "Tomorrow or next week"}, TriggerLabel: "Choose schedule date", Calendar: display.CalendarProps{Location: s.location, Caption: display.CalendarCaptionDropdown}, SideOffset: &gap, AlignOffset: &offset, Parse: s.parseRelative, Format: formatDate, OnChange: func(date time.Time) { s.feedback = "Schedule: " + s.describe(date) }}).Root
							})
							ui.Text(c, s.describe(s.naturalDate)).TextColor(t.MutedForeground)
						})
					})
					ui.Text(c, s.feedback).Role(ui.RoleStatus)
				})
			})
		})
	})
}
func main() {
	snapshot := flag.String("snapshot", "", "write a native PNG instead of opening a window")
	dark := flag.Bool("dark", false, "use dark appearance")
	open := flag.Bool("open", false, "show the editable-date calendar")
	flag.Parse()
	s := newState()
	s.dark = *dark
	s.exact.Open = *open
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		err = png.Encode(f, ui.Render(func(c *ui.Context) { c.DisableAnimations(); s.view(c) }, 1000, 1000, 1))
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Calendar and editable dates", Width: 1000, Height: 1000, Content: ui.View(s.view)}).Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
