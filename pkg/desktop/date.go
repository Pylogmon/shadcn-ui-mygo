package desktop

import (
	"errors"
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// DatePickerProps configures the paired independent-window date pickers.
// The native pair uses a native Popover container and a platform-themed Go
// calendar; MyGo does not expose a system date-grid widget. Calendar rules are
// identical in both styles.
type DatePickerProps struct {
	Placeholder, Label, Format string
	Disabled                   bool
	Calendar                   display.CalendarProps
	ShowTime                   bool
	TimeLabel                  string
	Presets                    []display.DatePreset
	RangePresets               []display.DateRangePreset
	// Popup controls the independent window. Structural calendar options and
	// controlled values refresh automatically; increment Popup.Revision when
	// mutable matcher/renderer closures change without structural data changes.
	Popup PopoverProps
	// Confirm defers publication until Apply. ShowTime implies Confirm so a
	// date can be chosen and its time edited before the window closes.
	Confirm                 bool
	ApplyLabel, CancelLabel string
	// Footer receives the CHILD context. Calendar.Footer is deliberately
	// rejected because a zero-argument callback may retain a parent frame.
	Footer        func(*ui.Context)
	OnChange      func(time.Time)
	OnRangeChange func(display.DateRange)
	OnCancel      func()
}

type desktopDateState struct {
	Open, Pending, RestoreFocus bool
	Committed, CancelReported   bool
	Source, Working             time.Time
	SourceRange, Range, Clocks  display.DateRange
	Revision                    uint64
	Options                     pickerOptions
}

// Structural configuration changes refresh an open child without causing
// idle redraws. Mutable matcher/content closures use Popup.Revision explicitly.
type pickerOptions struct {
	Placeholder, Label, Format, TimeLabel, ApplyLabel, CancelLabel string
	Disabled, ShowTime, Confirm                                    bool
	CalendarLabel                                                  string
	Month, DefaultMonth, StartMonth, EndMonth, MinDate, MaxDate    time.Time
	Location                                                       *time.Location
	WeekStartsOn                                                   int
	NumberOfMonths, Caption, Min, Max                              int
	HideOutside, FixedWeeks, WeekNumbers, DisableNavigation        bool
	Vertical, Reverse, Paged, Required, ExcludeDisabled            bool
	Presets, RangePresets                                          string
}

func pickerOptionsOf(p DatePickerProps) pickerOptions {
	q := p.Calendar
	week := -1
	if q.WeekStartsOn != nil {
		week = int(*q.WeekStartsOn)
	}
	var month time.Time
	if q.Month != nil {
		month = *q.Month
	}
	var presets, ranges []byte
	for _, preset := range p.Presets {
		presets = fmt.Appendf(presets, "%q/%v/%s/%t;", preset.Label, preset.Date, preset.Date.Location(), preset.Disabled)
	}
	for _, preset := range p.RangePresets {
		ranges = fmt.Appendf(ranges, "%q/%v/%s/%v/%s/%t;", preset.Label, preset.Range.From, preset.Range.From.Location(), preset.Range.To, preset.Range.To.Location(), preset.Disabled)
	}
	return pickerOptions{Placeholder: p.Placeholder, Label: p.Label, Format: p.Format, TimeLabel: p.TimeLabel, ApplyLabel: p.ApplyLabel, CancelLabel: p.CancelLabel,
		Disabled: p.Disabled, ShowTime: p.ShowTime, Confirm: p.Confirm, CalendarLabel: q.Label,
		Month: month, DefaultMonth: q.DefaultMonth, StartMonth: q.StartMonth, EndMonth: q.EndMonth, MinDate: q.MinDate, MaxDate: q.MaxDate, Location: q.Location, WeekStartsOn: week,
		NumberOfMonths: q.NumberOfMonths, Caption: int(q.Caption), Min: q.Min, Max: q.Max,
		HideOutside: q.HideOutsideDays, FixedWeeks: q.FixedWeeks, WeekNumbers: q.WeekNumbers, DisableNavigation: q.DisableNavigation,
		Vertical: q.VerticalMonths, Reverse: q.ReverseMonths, Paged: q.PagedNavigation, Required: q.Required, ExcludeDisabled: q.ExcludeDisabled,
		Presets: string(presets), RangePresets: string(ranges)}
}

// DatePicker opens the uniform shadcn calendar in an independent Flyout.
// A selected day commits once. Confirm/ShowTime keep a cancellable draft.
func DatePicker(c *ui.Context, h *Host, key string, date *time.Time, p DatePickerProps) (ui.Element, error) {
	return datePicker(c, h, key, date, nil, p, false)
}

// NativeDatePicker keeps the child window's platform theme and native popup
// exterior, while sharing the constrained Go calendar behavior with DatePicker.
func NativeDatePicker(c *ui.Context, h *Host, key string, date *time.Time, p DatePickerProps) (ui.Element, error) {
	return datePicker(c, h, key, date, nil, p, true)
}

// DateRangePicker isolates an incomplete range in the child. The caller's
// value changes only after both endpoints are valid, or after explicit Apply.
func DateRangePicker(c *ui.Context, h *Host, key string, date *display.DateRange, p DatePickerProps) (ui.Element, error) {
	return datePicker(c, h, key, nil, date, p, false)
}

func NativeDateRangePicker(c *ui.Context, h *Host, key string, date *display.DateRange, p DatePickerProps) (ui.Element, error) {
	return datePicker(c, h, key, nil, date, p, true)
}

func datePicker(c *ui.Context, h *Host, key string, date *time.Time, dates *display.DateRange, p DatePickerProps, native bool) (ui.Element, error) {
	if c == nil || !c.Valid() || key == "" || date == nil && dates == nil {
		return ui.Element{}, errors.New("desktop: date picker needs a live context, stable key and caller-owned value")
	}
	if p.Calendar.Footer != nil {
		return ui.Element{}, errors.New("desktop: Calendar.Footer cannot cross window contexts; use DatePickerProps.Footer(child)")
	}
	rangeMode := dates != nil
	root := ui.Column(c.Key("date-picker:" + key))
	s := ui.Local(root, "date-picker", func() desktopDateState { return desktopDateState{} })
	if options := pickerOptionsOf(p); options != s.Options {
		s.Options = options
		s.Revision++
	}
	loc := pickerLocation(p.Calendar, date, dates)
	if date != nil && !pickerEqual(s.Source, *date) {
		s.Source, s.Working = *date, *date
		s.Revision++
	}
	if dates != nil && !pickerRangeEqual(s.SourceRange, *dates) {
		s.SourceRange, s.Range, s.Clocks = *dates, *dates, *dates
		s.Revision++
	}
	var anchor ui.Element
	var popupErr error
	root.Children(func() {
		text := pickerText(date, dates, p, loc)
		label := p.Label
		if label == "" {
			label = text
		}
		if native {
			anchor = ui.Button(c, text).Disabled(p.Disabled).Label(label)
		} else {
			anchor = control.Button(c, text, control.ButtonProps{Variant: control.Outline, Disabled: p.Disabled, Label: label})
		}
		anchor.Value(text).HasPopup(ui.PopupDialog)
		if anchor.Clicked() && !p.Disabled {
			s.Open = !s.Open
			if s.Open {
				s.Committed, s.CancelReported = false, false
				if date != nil {
					s.Source, s.Working = *date, *date
				} else {
					s.SourceRange, s.Range, s.Clocks = *dates, *dates, *dates
				}
				s.Revision++
			}
		}
		if p.Disabled {
			s.Open = false
		}
		anchor.Expanded(s.Open)
		if s.RestoreFocus {
			anchor.Focus()
			s.RestoreFocus = false
		}
		if s.Pending {
			anchor.SetChanged()
			root.SetChanged()
			s.Pending = false
		}
		if h == nil {
			if s.Open {
				s.Open = false
				anchor.Expanded(false)
				popupErr = errors.New("desktop: opening a date picker requires a parent-bound Host")
			}
			return
		}
		popup := p.Popup
		popup.ManualTrigger = true
		popup.FitContent = true
		popup.Revision += s.Revision
		if popup.Label == "" {
			popup.Label = label
		}
		if popup.Width == 0 {
			popup.Width = 300
			if rangeMode {
				popup.Width = 600
			}
		}
		if popup.ContentPadding == nil {
			padding := float32(4)
			popup.ContentPadding = &padding
		}
		dismissed := popup.OnDismiss
		popup.OnDismiss = func() {
			s.Open = false
			if !s.Committed && !s.CancelReported {
				s.CancelReported = true
				if p.OnCancel != nil {
					p.OnCancel()
				}
			}
			if dismissed != nil {
				dismissed()
			}
		}
		commit := func() {
			changed := false
			if date != nil {
				if !pickerValidDate(s.Working, p.Calendar, loc) {
					return
				}
				changed = !pickerEqual(*date, s.Working)
				*date, s.Source = s.Working, s.Working
				if changed && p.OnChange != nil {
					p.OnChange(*date)
				}
			} else {
				if !pickerValidRange(s.Range, p.Calendar, loc) {
					return
				}
				changed = !pickerRangeEqual(*dates, s.Range)
				*dates, s.SourceRange = s.Range, s.Range
				if changed && p.OnRangeChange != nil {
					p.OnRangeChange(*dates)
				}
			}
			s.Committed, s.Open, s.RestoreFocus = true, false, true
			s.Pending = s.Pending || changed
		}
		cancel := func() {
			s.Open, s.RestoreFocus = false, true
			if !s.CancelReported && !s.Committed {
				s.CancelReported = true
				if p.OnCancel != nil {
					p.OnCancel()
				}
			}
		}
		content := func(child *ui.Context) {
			build := func() { datePickerContent(child, s, p, loc, rangeMode, native, commit, cancel) }
			if native {
				theme.NativeScope(child).Children(build)
			} else {
				build()
			}
		}
		if native {
			popupErr = NativePopover(c, h, key, anchor, &s.Open, popup, content)
		} else {
			popupErr = Popover(c, h, key, anchor, &s.Open, popup, content)
		}
	})
	return anchor, popupErr
}

func datePickerContent(c *ui.Context, s *desktopDateState, p DatePickerProps, loc *time.Location, rangeMode, native bool, commit, cancel func()) {
	confirm := p.Confirm || p.ShowTime
	button := func(c *ui.Context, text string, disabled bool, primary bool) ui.Element {
		if native {
			return ui.Button(c, text).Disabled(disabled)
		}
		variant := control.Ghost
		if primary {
			variant = control.Default
		}
		return control.Button(c, text, control.ButtonProps{Variant: variant, Size: control.Small, Disabled: disabled})
	}
	props := p.Calendar
	props.Location, props.Browse, props.AutoFocus = loc, true, true
	props.Footer = nil
	if native && props.CaptionSelect == nil {
		props.CaptionSelect = control.NativeSelect
	}
	if p.Footer != nil {
		props.Footer = func() { p.Footer(c) }
	}
	if rangeMode && props.NumberOfMonths <= 0 {
		props.NumberOfMonths = 2
	}
	ui.Column(c).Gap(8).Children(func() {
		if rangeMode && len(p.RangePresets) > 0 {
			ui.Row(c).Gap(4).Wrap().Padding(4).Children(func() {
				for i, preset := range p.RangePresets {
					value := display.DateRange{From: pickerIn(preset.Range.From, loc), To: pickerIn(preset.Range.To, loc)}
					if button(c.Key(i), preset.Label, preset.Disabled || !pickerValidRange(value, p.Calendar, loc), false).Clicked() {
						s.Range, s.Clocks = value, value
						if !confirm {
							commit()
						}
					}
				}
			})
		} else if !rangeMode && len(p.Presets) > 0 {
			ui.Row(c).Gap(4).Wrap().Padding(4).Children(func() {
				for i, preset := range p.Presets {
					value := pickerWithDay(s.Working, preset.Date, loc)
					if button(c.Key(i), preset.Label, preset.Disabled || preset.Date.IsZero() || !pickerValidDate(value, p.Calendar, loc), false).Clicked() {
						s.Working = value
						if !confirm {
							commit()
						}
					}
				}
			})
		}
		if rangeMode {
			previous := s.Range
			if display.CalendarRange(c, &s.Range, props).Changed() {
				next := s.Range
				if p.ShowTime {
					next.From = pickerWithDay(s.Clocks.From, next.From, loc)
					next.To = pickerWithDay(s.Clocks.To, next.To, loc)
				}
				if !pickerValidDraftRange(next, p.Calendar, loc) {
					s.Range = previous
				} else {
					s.Range = next
					if !next.From.IsZero() {
						s.Clocks.From = next.From
					}
					if !next.To.IsZero() {
						s.Clocks.To = next.To
					}
					if !confirm && pickerValidRange(s.Range, p.Calendar, loc) {
						commit()
					}
				}
			}
		} else {
			previous := s.Working
			if display.CalendarWith(c, &s.Working, props).Changed() {
				if !pickerValidDate(s.Working, p.Calendar, loc) {
					s.Working = previous
				} else if !confirm {
					commit()
				}
			}
		}
		if p.ShowTime {
			ui.Row(c).Gap(12).Padding(8).Children(func() {
				if rangeMode {
					pickerTime(c.Key("from-time"), &s.Range.From, "Start time", p.Calendar, loc, func(value time.Time) bool {
						next := s.Range
						next.From = value
						if !pickerValidDraftRange(next, p.Calendar, loc) {
							return false
						}
						s.Clocks.From = value
						return true
					})
					pickerTime(c.Key("to-time"), &s.Range.To, "End time", p.Calendar, loc, func(value time.Time) bool {
						next := s.Range
						next.To = value
						if !pickerValidDraftRange(next, p.Calendar, loc) {
							return false
						}
						s.Clocks.To = value
						return true
					})
				} else {
					label := p.TimeLabel
					if label == "" {
						label = "Time"
					}
					pickerTime(c, &s.Working, label, p.Calendar, loc, nil)
				}
			})
		}
		if confirm {
			ui.Row(c).Gap(8).Padding(8).Justify(ui.End).Children(func() {
				cancelText, applyText := p.CancelLabel, p.ApplyLabel
				if cancelText == "" {
					cancelText = "Cancel"
				}
				if applyText == "" {
					applyText = "Apply"
				}
				if button(c, cancelText, false, false).Clicked() {
					cancel()
				}
				valid := pickerValidDate(s.Working, p.Calendar, loc)
				if rangeMode {
					valid = pickerValidRange(s.Range, p.Calendar, loc)
				}
				if button(c, applyText, !valid, true).Clicked() {
					commit()
				}
			})
		}
	})
}

func pickerTime(c *ui.Context, value *time.Time, label string, p display.CalendarProps, loc *time.Location, accept func(time.Time) bool) {
	ui.Column(c).Gap(4).Children(func() {
		ui.Text(c, label).FontSize(12)
		working := pickerIn(*value, loc)
		if working.IsZero() {
			working = c.Now().In(loc)
		}
		if ui.TimeInput(c, &working).Label(label).Disabled(value.IsZero()).Changed() && pickerValidDate(working, p, loc) && (accept == nil || accept(working)) {
			*value = working
		}
	})
}

func pickerText(date *time.Time, dates *display.DateRange, p DatePickerProps, loc *time.Location) string {
	format := p.Format
	if format == "" {
		format = "Jan 2, 2006"
	}
	text := p.Placeholder
	if dates != nil {
		if text == "" {
			text = "Pick a date range"
		}
		if !dates.From.IsZero() {
			text = dates.From.In(loc).Format(format)
			if !dates.To.IsZero() {
				text += " – " + dates.To.In(loc).Format(format)
			}
		}
	} else {
		if text == "" {
			text = "Pick a date"
		}
		if !date.IsZero() {
			text = date.In(loc).Format(format)
		}
	}
	return text
}
func pickerLocation(p display.CalendarProps, date *time.Time, dates *display.DateRange) *time.Location {
	if p.Location != nil {
		return p.Location
	}
	if date != nil && !date.IsZero() {
		return date.Location()
	}
	if dates != nil && !dates.From.IsZero() {
		return dates.From.Location()
	}
	return time.Local
}
func pickerIn(date time.Time, loc *time.Location) time.Time {
	if date.IsZero() {
		return date
	}
	return date.In(loc)
}
func pickerWithDay(clock, day time.Time, loc *time.Location) time.Time {
	if day.IsZero() {
		return day
	}
	clock, day = pickerIn(clock, loc), day.In(loc)
	h, m, s := clock.Clock()
	return time.Date(day.Year(), day.Month(), day.Day(), h, m, s, clock.Nanosecond(), loc)
}
func pickerCivil(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}
func pickerAllowed(date time.Time, p display.CalendarProps, loc *time.Location) bool {
	if date.IsZero() {
		return false
	}
	date = date.In(loc)
	day := pickerCivil(date)
	return (p.MinDate.IsZero() || !day.Before(pickerCivil(p.MinDate.In(loc)))) &&
		(p.MaxDate.IsZero() || !day.After(pickerCivil(p.MaxDate.In(loc)))) &&
		(p.Disabled == nil || !p.Disabled(date)) && (p.Hidden == nil || !p.Hidden(date))
}
func pickerValidDate(date time.Time, p display.CalendarProps, loc *time.Location) bool {
	if date.IsZero() {
		return !p.Required
	}
	return pickerAllowed(date, p, loc)
}
func pickerValidDraftRange(value display.DateRange, p display.CalendarProps, loc *time.Location) bool {
	if value.From.IsZero() {
		return value.To.IsZero() && !p.Required
	}
	if !pickerAllowed(value.From, p, loc) {
		return false
	}
	if value.To.IsZero() {
		return true
	}
	return pickerValidRange(value, p, loc)
}
func pickerValidRange(value display.DateRange, p display.CalendarProps, loc *time.Location) bool {
	if value.From.IsZero() || value.To.IsZero() || !pickerAllowed(value.From, p, loc) || !pickerAllowed(value.To, p, loc) {
		return false
	}
	from, to := value.From.In(loc), value.To.In(loc)
	if to.Before(from) {
		return false
	}
	nights := int((pickerCivil(to).Unix() - pickerCivil(from).Unix()) / 86400)
	if nights < max(0, p.Min) || p.Max > 0 && nights > p.Max {
		return false
	}
	if p.ExcludeDisabled {
		for day := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc); !pickerCivil(day).After(pickerCivil(to)); day = day.AddDate(0, 0, 1) {
			if p.Disabled != nil && p.Disabled(day) {
				return false
			}
		}
	}
	return true
}
func pickerEqual(a, b time.Time) bool {
	return a.Equal(b) && (a.IsZero() || a.Location() == b.Location())
}
func pickerRangeEqual(a, b display.DateRange) bool {
	return pickerEqual(a.From, b.From) && pickerEqual(a.To, b.To)
}
