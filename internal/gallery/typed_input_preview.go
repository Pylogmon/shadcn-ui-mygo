package gallery

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// The Gallery owns this model across page changes. The optional picker is a
// value-only application service; nil selects the real operating-system dialog.
type typedInputPreviewState struct {
	ready               bool
	amount              float64
	date                time.Time
	draft               display.DatePickerInputState
	calendar            string
	paths               []string
	fileError, feedback string
	saved               string
	picker              control.FilePicker
}

func bookingDateText(date time.Time) string { return date.Format("January 02, 2006") }

func (p *typedInputPreviewState) initialize() {
	if p.ready {
		return
	}
	p.ready = true
	p.amount = 125.50
	p.date = time.Date(2026, 11, 10, 15, 30, 0, 0, time.Local)
	p.calendar = "gregorian"
	p.restoreDateDraft()
}

func (p *typedInputPreviewState) restoreDateDraft() {
	p.draft.Text, p.draft.Month, p.draft.Error = bookingDateText(p.date), p.date, ""
}

func (p *typedInputPreviewState) parseDate(text string) (time.Time, error) {
	for _, layout := range []string{"January 02, 2006", "2006-01-02"} {
		day, err := time.ParseInLocation(layout, strings.TrimSpace(text), p.date.Location())
		if err == nil {
			h, m, s := p.date.Clock()
			return time.Date(day.Year(), day.Month(), day.Day(), h, m, s, p.date.Nanosecond(), p.date.Location()), nil
		}
	}
	return time.Time{}, fmt.Errorf("Use November 12, 2026 or 2026-11-12.")
}

func (p *typedInputPreviewState) summary() string {
	return fmt.Sprintf("$%.2f · %s · %d attachment(s)", p.amount, p.date.Format("2006-01-02 15:04 MST"), len(p.paths))
}

func (p *typedInputPreviewState) view(c *ui.Context, window *mygo.Window) {
	p.initialize()
	t := theme.TokensOf(c)
	display.Card(c.Key("booking-inputs"), func() {
		ui.Column(c).FillWidth().Gap(18).Children(func() {
			display.Typography(c, tr(c, "Plan a booking"), display.H3)
			ui.Text(c, tr(c, "Set a budget and departure, attach an itinerary, then save your booking.")).TextColor(t.MutedForeground)
			ui.Row(c).FillWidth().Wrap().Gap(24).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Basis(0).Grow(1).MinWidth(260).Gap(16).Children(func() {
					control.Field(c, control.FieldProps{Label: tr(c, "Booking amount"), Description: tr(c, "USD; use the arrow keys to adjust by 0.50.")}, func() ui.Element {
						return control.InputGroup(c, control.InputGroupProps{Prefix: "$", Suffix: "USD"}, func() ui.Element {
							minimum := float64(0)
							return control.NumberInput(c, &p.amount, control.NumberInputProps{Min: &minimum, Step: .5, OnSubmit: func(float64) { p.feedback = fmt.Sprintf("Budget updated to $%.2f.", p.amount) }})
						})
					})
					control.Field(c, control.FieldProps{Label: tr(c, "Booking date")}, func() ui.Element {
						return control.DateInput(c, &p.date, control.DateInputProps{OnChange: func(time.Time) {
							p.restoreDateDraft()
							p.feedback = "Departure date updated."
						}})
					})
					control.Field(c, control.FieldProps{Label: tr(c, "Booking time"), Description: tr(c, "Uses your local hour format.")}, func() ui.Element {
						return control.TimeInput(c, &p.date, control.TimeInputProps{OnChange: func(time.Time) {
							p.restoreDateDraft()
							p.feedback = "Departure time updated."
						}})
					})
					control.Field(c, control.FieldProps{Label: tr(c, "Booking attachment"), Description: tr(c, "Choose documents from your computer."), Error: tr(c, p.fileError)}, func() ui.Element {
						return control.FileInput(c, &p.paths, control.FileInputProps{
							Parent: window, Picker: p.picker, Title: tr(c, "Attach booking documents"), ChooseLabel: "Choose documents", Multiple: true,
							Filters: []mygo.FileFilter{{Name: "Documents", Extensions: []string{"txt", "pdf", "csv"}}},
							OnChange: func(paths []string) {
								p.fileError = ""
								p.feedback = fmt.Sprintf("Attached %d document(s).", len(paths))
							},
							OnError: func(error) {
								p.fileError = "The documents could not be selected. Try again."
								p.feedback = p.fileError
							},
						})
					})
				})
				ui.Column(c).Basis(0).Grow(1).MinWidth(260).Gap(16).Children(func() {
					control.Field(c, control.FieldProps{Label: tr(c, "Departure calendar"), Description: tr(c, "Choose how the popup displays the same departure date.")}, func() ui.Element {
						return control.NativeSelect(c, &p.calendar, control.SelectProps{Options: localizedOptions(c, []control.Option{{Value: "gregorian", Label: "Gregorian"}, {Value: "persian", Label: "Persian / Jalali"}})})
					})
					control.Field(c, control.FieldProps{Label: tr(c, "Departure query"), Description: tr(c, "Type an exact date or press ArrowDown to browse."), Error: tr(c, p.draft.Error)}, func() ui.Element {
						system := display.CalendarGregorian
						if p.calendar == "persian" {
							system = display.CalendarPersian
						}
						return display.DatePickerInput(c, &p.date, &p.draft, display.DatePickerInputProps{
							Input: control.InputProps{Placeholder: "2026-11-12", OnSubmit: func(string) {
								if p.draft.Error == "" {
									p.feedback = "Departure date confirmed."
								}
							}},
							TriggerLabel: "Choose departure date", Calendar: display.CalendarProps{System: system, Location: p.date.Location()}, Parse: p.parseDate, Format: bookingDateText,
							OnChange: func(time.Time) { p.feedback = "Departure date updated." },
						}).Root
					})
					if control.Button(c, tr(c, "Restore departure"), control.ButtonProps{Variant: control.Outline, Disabled: p.draft.Text == bookingDateText(p.date) && p.draft.Error == ""}).Clicked() {
						p.restoreDateDraft()
						p.draft.Open = false
						p.feedback = "Restored the accepted departure date."
					}
					if day, err := display.PersianDateOf(p.date); err == nil {
						ui.Text(c, fmt.Sprintf("Departure: %s\nJalali: %04d/%02d/%02d", p.date.Format("2006-01-02 15:04 MST"), day.Year, day.Month, day.Day)).TextColor(t.MutedForeground)
					}
				})
			})
			ui.Text(c, tr(c, "Current booking: ")+p.summary()).FontWeight(500)
			if len(p.paths) > 0 {
				names := make([]string, len(p.paths))
				for i, path := range p.paths {
					names[i] = filepath.Base(path)
				}
				ui.Text(c, tr(c, "Documents: ")+strings.Join(names, ", ")).TextColor(t.MutedForeground)
			}
			if control.Button(c, tr(c, "Save booking"), control.ButtonProps{Disabled: p.draft.Error != ""}).Clicked() {
				p.saved = p.summary()
				p.feedback = "Booking saved."
			}
			if p.saved != "" {
				ui.Text(c, tr(c, "Booking saved: ")+p.saved).Role(ui.RoleStatus)
			}
			if p.feedback != "" {
				ui.Text(c, p.feedback).Role(ui.RoleStatus)
			}
		})
	}).Role(ui.RoleGroup).Label(tr(c, "Booking planner"))
}
