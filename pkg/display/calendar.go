package display

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Calendar retains MyGo's single-date convenience widget, including its fluent
// Browse modifier. Use CalendarWith, CalendarMultiple or CalendarRange when
// supplying date constraints, multiple months or a different selection mode.
func Calendar(c *ui.Context, date *time.Time) ui.Element {
	return ui.Calendar(c, date).Padding(12).Radius(theme.TokensOf(c).Radius).Border(1, theme.TokensOf(c).Border)
}

type DateMatcher func(time.Time) bool
type DateRange struct{ From, To time.Time }
type CalendarCaption int

const (
	CalendarCaptionLabel CalendarCaption = iota
	CalendarCaptionDropdown
	CalendarCaptionDropdownMonths
	CalendarCaptionDropdownYears
)

type CalendarDay struct {
	Date                                                                  time.Time
	Outside, Today, Selected, Disabled, RangeStart, RangeEnd, RangeMiddle bool
}
type CalendarProps struct {
	Label string
	// System defaults to Gregorian. Persian uses the fixed Jalali month model,
	// Farsi names, Eastern Arabic-Indic digits and Saturday-first weeks.
	System CalendarSystem
	// Nil inherits the parent direction for Gregorian and uses RTL for Persian.
	Direction *ui.Direction
	// Month is an optional caller-owned displayed month. DefaultMonth is used
	// only when the calendar is first constructed; browsing never selects a day.
	Month                                                       *time.Time
	DefaultMonth, StartMonth, EndMonth                          time.Time
	MinDate, MaxDate                                            time.Time
	Disabled, Hidden                                            DateMatcher
	Location                                                    *time.Location
	WeekStartsOn                                                *time.Weekday
	NumberOfMonths                                              int
	Caption                                                     CalendarCaption
	HideOutsideDays, FixedWeeks, WeekNumbers, DisableNavigation bool
	VerticalMonths, ReverseMonths, PagedNavigation              bool
	// Min/Max count selected days in multiple mode, nights in range mode.
	// Zero means no limit. Required prevents clearing a single/multiple value.
	Min, Max                  int
	Required, ExcludeDisabled bool
	// Single-date arrow navigation selects by default, matching Calendar.
	// Browse moves the active day only; Enter/Space then commits selection.
	Browse, AutoFocus bool
	DayContent        func(*ui.Context, CalendarDay)
	// CaptionSelect may provide a native or independently hosted month/year
	// picker. It runs in the calendar's own Context, never a parent's frame.
	CaptionSelect func(*ui.Context, *string, control.SelectProps) ui.Element
	Footer        func()
	OnMonthChange func(time.Time)
}

type calendarState struct {
	Month, Cursor      time.Time
	Signature          string
	Ready, FocusCursor bool
	System             CalendarSystem
}

func calendarLocation(p CalendarProps, date time.Time) *time.Location {
	if p.Location != nil {
		return p.Location
	}
	if !date.IsZero() {
		return date.Location()
	}
	return time.Local
}
func dateIn(d time.Time, loc *time.Location) time.Time {
	if d.IsZero() {
		return d
	}
	return d.In(loc)
}
func dayDate(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}
func monthDate(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
}
func dayKey(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format("2006-01-02")
}
func sameDate(a, b time.Time) bool { return dayKey(a) == dayKey(b) }
func compareDay(a, b time.Time) int {
	av := []int{a.Year(), int(a.Month()), a.Day()}
	bv := []int{b.Year(), int(b.Month()), b.Day()}
	for i := range av {
		if av[i] < bv[i] {
			return -1
		}
		if av[i] > bv[i] {
			return 1
		}
	}
	return 0
}
func withCalendarDay(original, day time.Time, loc *time.Location) time.Time {
	base := dateIn(original, loc)
	h, m, s := base.Clock()
	return time.Date(day.Year(), day.Month(), day.Day(), h, m, s, base.Nanosecond(), loc)
}
func disabledDay(day time.Time, p CalendarProps) bool {
	loc := day.Location()
	return !p.MinDate.IsZero() && compareDay(day, dateIn(p.MinDate, loc)) < 0 || !p.MaxDate.IsZero() && compareDay(day, dateIn(p.MaxDate, loc)) > 0 || p.Disabled != nil && p.Disabled(day)
}
func unselectableDate(day time.Time, p CalendarProps) bool {
	return day.IsZero() || !modelFor(p.System).valid(day) || disabledDay(day, p) || p.Hidden != nil && p.Hidden(day)
}
func calendarBounds(p CalendarProps, loc *time.Location) (time.Time, time.Time) {
	model := modelFor(p.System)
	lo, hi := dateIn(p.StartMonth, loc), dateIn(p.EndMonth, loc)
	if !p.MinDate.IsZero() && (lo.IsZero() || compareDay(model.start(dateIn(p.MinDate, loc)), model.start(lo)) > 0) {
		lo = p.MinDate.In(loc)
	}
	if !p.MaxDate.IsZero() && (hi.IsZero() || compareDay(model.start(dateIn(p.MaxDate, loc)), model.start(hi)) < 0) {
		hi = p.MaxDate.In(loc)
	}
	if !lo.IsZero() {
		lo = model.start(lo)
	}
	if !hi.IsZero() {
		hi = model.start(hi)
	}
	return lo, hi
}
func boundedMonth(month time.Time, p CalendarProps, loc *time.Location) time.Time {
	model := modelFor(p.System)
	month = model.start(month.In(loc))
	if p.System == CalendarPersian && !model.valid(month) {
		y, _, _ := model.parts(month)
		if y < 1 {
			month = persianTime(1, 1, 1, loc)
		} else {
			month = persianTime(maxPersianYear, 12, 1, loc)
		}
	}
	lo, hi := calendarBounds(p, loc)
	if !lo.IsZero() && month.Before(lo) {
		month = lo
	}
	if !hi.IsZero() && month.After(hi) {
		month = hi
	}
	return month
}
func calendarDaysBetween(a, b time.Time) int {
	// Civil days, rather than elapsed hours, keep range limits correct across DST.
	aa := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
	bb := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
	return int((bb.Unix() - aa.Unix()) / 86400)
}

func CalendarWith(c *ui.Context, date *time.Time, p CalendarProps) ui.Element {
	if date == nil {
		panic("display: CalendarWith requires a date pointer")
	}
	loc := calendarLocation(p, *date)
	signature := func() string { return dayKey(dateIn(*date, loc)) }
	return calendarView(c, p, dateIn(*date, loc), signature, func(day time.Time) bool {
		if sameDate(dateIn(*date, loc), day) {
			if p.Required {
				return false
			}
			*date = time.Time{}
			return true
		}
		*date = withCalendarDay(*date, day, loc)
		return true
	}, func(day time.Time) CalendarDay { return CalendarDay{Selected: sameDate(dateIn(*date, loc), day)} }, false)
}
func CalendarMultiple(c *ui.Context, dates *[]time.Time, p CalendarProps) ui.Element {
	if dates == nil {
		panic("display: CalendarMultiple requires a dates pointer")
	}
	first := time.Time{}
	if len(*dates) > 0 {
		first = (*dates)[0]
	}
	loc := calendarLocation(p, first)
	signature := func() string {
		keys := make([]string, len(*dates))
		for i, d := range *dates {
			keys[i] = dayKey(dateIn(d, loc))
		}
		return strings.Join(keys, ",")
	}
	return calendarView(c, p, dateIn(first, loc), signature, func(day time.Time) bool {
		at := slices.IndexFunc(*dates, func(d time.Time) bool { return sameDate(dateIn(d, loc), day) })
		if at >= 0 {
			minimum := max(0, p.Min)
			if p.Required {
				minimum = max(1, minimum)
			}
			if len(*dates) <= minimum {
				return false
			}
			*dates = slices.Delete(slices.Clone(*dates), at, at+1)
			return true
		}
		if p.Max > 0 && len(*dates) >= p.Max {
			return false
		}
		*dates = append(slices.Clone(*dates), dayDate(day))
		return true
	}, func(day time.Time) CalendarDay {
		return CalendarDay{Selected: slices.ContainsFunc(*dates, func(d time.Time) bool { return sameDate(dateIn(d, loc), day) })}
	}, true)
}
func CalendarRange(c *ui.Context, date *DateRange, p CalendarProps) ui.Element {
	if date == nil {
		panic("display: CalendarRange requires a range pointer")
	}
	loc := calendarLocation(p, date.From)
	signature := func() string { return dayKey(dateIn(date.From, loc)) + ":" + dayKey(dateIn(date.To, loc)) }
	return calendarView(c, p, dateIn(date.From, loc), signature, func(day time.Time) bool {
		if date.From.IsZero() || !date.To.IsZero() {
			*date = DateRange{From: dayDate(day)}
			return true
		}
		from, to := dayDate(date.From.In(loc)), dayDate(day)
		if from.After(to) {
			from, to = to, from
		}
		nights := calendarDaysBetween(from, to)
		if nights < max(0, p.Min) || p.Max > 0 && nights > p.Max {
			return false
		}
		if p.ExcludeDisabled {
			for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
				if disabledDay(d, p) {
					*date = DateRange{From: dayDate(day)}
					return true
				}
			}
		}
		*date = DateRange{From: from, To: to}
		return true
	}, func(day time.Time) CalendarDay {
		from, to := dateIn(date.From, loc), dateIn(date.To, loc)
		start, end := sameDate(from, day), sameDate(to, day)
		middle := !from.IsZero() && !to.IsZero() && compareDay(day, from) > 0 && compareDay(day, to) < 0
		return CalendarDay{Selected: start || end || middle, RangeStart: start, RangeEnd: end, RangeMiddle: middle}
	}, true)
}

func calendarView(c *ui.Context, p CalendarProps, selected time.Time, signature func() string, choose func(time.Time) bool, selection func(time.Time) CalendarDay, multiple bool) ui.Element {
	model := modelFor(p.System)
	t := theme.TokensOf(c)
	loc := calendarLocation(p, selected)
	n := max(1, p.NumberOfMonths)
	lo, hi := calendarBounds(p, loc)
	if !lo.IsZero() && !hi.IsZero() && !hi.Before(lo) {
		n = min(n, model.difference(hi, lo)+1)
	}
	boundView := func(month time.Time) time.Time {
		month = boundedMonth(month, p, loc)
		if !hi.IsZero() {
			latest := model.add(hi, -n+1)
			if !lo.IsZero() && latest.Before(lo) {
				latest = lo
			}
			if month.After(latest) {
				month = latest
			}
		}
		return month
	}
	root := ui.Column(c).AlignSelf(ui.Start).Gap(12).Padding(12).Radius(t.Radius).Border(1, t.Border).Background(t.Background).TextColor(t.Foreground).Role(ui.RoleGroup)
	label := p.Label
	if label == "" {
		label = "Calendar"
	}
	root.Label(label)
	if p.Direction != nil {
		root.Direction(*p.Direction)
	} else if p.System == CalendarPersian {
		root.Direction(ui.RightToLeft)
	}
	if !model.valid(selected) {
		return root.Invalid(true).Description("Date is outside the supported Persian calendar").Children(func() { ui.Text(c, "Unsupported Persian date") })
	}
	st := ui.Local(root, "calendar", func() calendarState { return calendarState{} })
	if st.Ready && st.System != p.System {
		st.Month = boundView(st.Month)
		st.System = p.System
	}
	if !st.Ready {
		st.System = p.System
		base := selected
		if base.IsZero() {
			base = c.Now().In(loc)
		}
		st.Cursor = dayDate(base)
		st.Month = p.DefaultMonth
		if st.Month.IsZero() {
			st.Month = base
		}
		st.Month = boundView(st.Month)
		if st.Cursor.Before(st.Month) || !st.Cursor.Before(model.add(st.Month, n)) {
			st.Cursor = st.Month
		}
		// Focus a selectable day when an empty/default month is constrained.
		if disabledDay(st.Cursor, p) || p.Hidden != nil && p.Hidden(st.Cursor) {
			for day := st.Month; day.Before(model.add(st.Month, n)); day = day.AddDate(0, 0, 1) {
				if !disabledDay(day, p) && (p.Hidden == nil || !p.Hidden(day)) {
					st.Cursor = day
					break
				}
			}
		}
		st.Signature = signature()
		st.Ready = true
	}
	if st.Signature != signature() {
		st.Signature = signature()
		if !selected.IsZero() {
			st.Cursor = dayDate(selected)
			st.Month = boundView(selected)
		}
	}
	if p.Month != nil {
		if !p.Month.IsZero() {
			st.Month = boundView(*p.Month)
		}
		*p.Month = st.Month
	}
	// Caption buttons and controlled Month changes browse without selecting.
	// Keep the grid's active descendant inside the visible months so keyboard
	// activation cannot select a stale date from the month just left behind.
	relocateCursor := func() {
		end := model.add(st.Month, n)
		if !st.Cursor.Before(st.Month) && st.Cursor.Before(end) && !unselectableDate(st.Cursor, p) {
			return
		}
		target := st.Month
		if !st.Cursor.Before(end) {
			target = model.add(st.Month, n-1)
		}
		last := model.add(target, 1).AddDate(0, 0, -1)
		year, month, _ := model.parts(target)
		_, _, cursorDay := model.parts(st.Cursor)
		_, _, lastDay := model.parts(last)
		candidate := model.date(year, month, min(max(1, cursorDay), lastDay), loc)
		for day := candidate; day.Before(end); day = day.AddDate(0, 0, 1) {
			if !unselectableDate(day, p) {
				st.Cursor = day
				return
			}
		}
		for day := candidate.AddDate(0, 0, -1); !day.Before(st.Month); day = day.AddDate(0, 0, -1) {
			if !unselectableDate(day, p) {
				st.Cursor = day
				return
			}
		}
		st.Cursor = candidate
	}
	relocateCursor()
	setMonth := func(month time.Time) {
		next := boundView(month)
		if !sameDate(next, st.Month) {
			st.Month = next
			relocateCursor()
			if p.Month != nil {
				*p.Month = next
			}
			if p.OnMonthChange != nil {
				p.OnMonthChange(next)
			}
		}
	}
	if p.Month != nil && p.Month.IsZero() {
		*p.Month = st.Month
	}
	commit := func(day time.Time) {
		if disabledDay(day, p) || p.Hidden != nil && p.Hidden(day) {
			return
		}
		if choose(day) {
			st.Signature = signature()
			root.SetChanged()
		}
		st.Cursor = dayDate(day)
		st.FocusCursor = true
	}
	moveCursor := func(day time.Time, delta int) {
		lo, hi := calendarBounds(p, loc)
		attempt := dayDate(day)
		if !model.valid(attempt) {
			return
		}
		if p.DisableNavigation && (attempt.Before(st.Month) || !attempt.Before(model.add(st.Month, n))) {
			return
		}
		for tries := 0; tries < 3660; tries++ {
			if !model.valid(attempt) {
				return
			}
			if p.DisableNavigation && (attempt.Before(st.Month) || !attempt.Before(model.add(st.Month, n))) {
				return
			}
			if !lo.IsZero() && model.start(attempt).Before(lo) || !hi.IsZero() && model.start(attempt).After(hi) {
				return
			}
			if model.valid(attempt) && !disabledDay(attempt, p) && (p.Hidden == nil || !p.Hidden(attempt)) {
				break
			}
			attempt = attempt.AddDate(0, 0, delta)
			if tries == 3659 {
				return
			}
		}
		st.Cursor = attempt
		st.FocusCursor = true
		if model.start(attempt).Before(st.Month) {
			setMonth(attempt)
		} else if model.start(attempt).After(model.add(st.Month, n-1)) {
			setMonth(model.add(attempt, -n+1))
		}
		if !p.Browse && !multiple && !sameDate(dateIn(selected, loc), attempt) {
			commit(attempt)
		}
	}
	root.Children(func() {

		lo, hi := calendarBounds(p, loc)
		step := 1
		if p.PagedNavigation {
			step = n
		}
		prev := ui.ButtonBase(c.Key("previous")).Absolute().Top(12).Start(12).ZIndex(1).Size(28, 28).Radius(t.RadiusSM()).Label("Previous month").Disabled(p.DisableNavigation || !model.valid(model.add(st.Month, -step)) || !lo.IsZero() && !st.Month.After(lo)).Children(func() {
			glyph := "‹"
			if c.Direction() == ui.RightToLeft {
				glyph = "›"
			}
			ui.Text(c, glyph)
		})
		if prev.Hovered() {
			prev.Background(t.Accent)
		}
		if prev.Clicked() {
			setMonth(model.add(st.Month, -step))
		}
		next := ui.ButtonBase(c.Key("next")).Absolute().Top(12).End(12).ZIndex(1).Size(28, 28).Radius(t.RadiusSM()).Label("Next month").Disabled(p.DisableNavigation || !model.valid(model.add(st.Month, n)) || !hi.IsZero() && !model.add(st.Month, n-1).Before(hi)).Children(func() {
			glyph := "›"
			if c.Direction() == ui.RightToLeft {
				glyph = "‹"
			}
			ui.Text(c, glyph)
		})
		if next.Hovered() {
			next.Background(t.Accent)
		}
		if next.Clicked() {
			setMonth(model.add(st.Month, step))
		}

		months := ui.Row(c).Gap(16).AlignItems(ui.Start)
		if p.VerticalMonths {
			months.Column()
		}
		months.Children(func() {
			for order := range n {
				index := order
				if p.ReverseMonths {
					index = n - 1 - order
				}
				month := model.add(st.Month, index)
				ctx := c.Key(month.Format("2006-01"))
				grid := ui.Column(ctx).Gap(4).Role(ui.RoleGrid).ColumnCount(7).Multiselectable(multiple).Focusable().Label(model.label(c.Locale(), month, true))
				if p.AutoFocus && model.start(st.Cursor).Equal(month) {
					grid.AutoFocus()
				}
				if st.FocusCursor && model.start(st.Cursor).Equal(month) {
					grid.Focus()
					st.FocusCursor = false
				}
				if grid.Focused() {
					back, forth := ui.KeyLeft, ui.KeyRight
					if c.Direction() == ui.RightToLeft {
						back, forth = forth, back
					}
					cur := st.Cursor
					switch {
					case grid.Shortcut(0, back):
						moveCursor(cur.AddDate(0, 0, -1), -1)
					case grid.Shortcut(0, forth):
						moveCursor(cur.AddDate(0, 0, 1), 1)
					case grid.Shortcut(0, ui.KeyUp):
						moveCursor(cur.AddDate(0, 0, -7), -1)
					case grid.Shortcut(0, ui.KeyDown):
						moveCursor(cur.AddDate(0, 0, 7), 1)
					case grid.Shortcut(0, ui.KeyHome):
						moveCursor(model.start(cur), 1)
					case grid.Shortcut(0, ui.KeyEnd):
						moveCursor(model.add(model.start(cur), 1).AddDate(0, 0, -1), -1)
					case grid.Shortcut(0, ui.KeyPageUp):
						if !p.DisableNavigation {
							moveCursor(model.move(cur, -1), -1)
						}
					case grid.Shortcut(0, ui.KeyPageDown):
						if !p.DisableNavigation {
							moveCursor(model.move(cur, 1), 1)
						}
					case grid.Shortcut(ui.Shift, ui.KeyPageUp):
						if !p.DisableNavigation {
							moveCursor(model.move(cur, -12), -1)
						}
					case grid.Shortcut(ui.Shift, ui.KeyPageDown):
						if !p.DisableNavigation {
							moveCursor(model.move(cur, 12), 1)
						}
					case grid.Shortcut(0, ui.KeyEnter), grid.Shortcut(0, ui.KeySpace):
						commit(cur)
					}
				}
				grid.Children(func() {
					if p.Caption != CalendarCaptionLabel {
						calendarCaption(ctx, month, index, p, loc, setMonth)
					} else {
						ui.Text(ctx, model.label(c.Locale(), month, true)).TextAlign(ui.Center).FontSize(14).FontWeight(500).Height(28)
					}
					first := model.firstWeekday(c, p)
					names := model.weekdays(c.Locale())
					ui.Row(ctx).Role(ui.RoleRow).Children(func() {
						if p.WeekNumbers {
							ui.Text(ctx, "#").Width(28).TextAlign(ui.Center).Role(ui.RoleColumnHeader)
						}
						for i := range 7 {
							ui.Text(ctx, names[(int(first)+i)%7]).Width(32).FontSize(12).TextColor(t.MutedForeground).TextAlign(ui.Center).Role(ui.RoleColumnHeader)
						}
					})
					start := month.AddDate(0, 0, -((int(month.Weekday()) - int(first) + 7) % 7))
					weeks := (calendarDaysBetween(start, model.add(month, 1).AddDate(0, 0, -1)) + 7) / 7
					if p.FixedWeeks {
						weeks = 6
					}
					grid.RowCount(weeks + 1)
					if p.WeekNumbers {
						grid.ColumnCount(8)
					}
					for week := range weeks {
						ui.Row(ctx.Key(week)).Role(ui.RoleRow).RowIndex(week + 2).Children(func() {
							if p.WeekNumbers {
								num := model.week(start.AddDate(0, 0, week*7), first)
								ui.Text(ctx, num).Width(28).FontSize(12).TextColor(t.MutedForeground).TextAlign(ui.Center).Role(ui.RoleRowHeader)
							}
							for weekday := range 7 {
								day := start.AddDate(0, 0, week*7+weekday)
								d := selection(day)
								d.Date = day
								d.Outside = !model.start(day).Equal(month)
								d.Disabled = disabledDay(day, p)
								d.Today = sameDate(day, c.Now().In(loc))
								// A date already rendered in a neighboring month is decorative here.
								duplicate := d.Outside && !model.start(day).Before(st.Month) && !model.start(day).After(model.add(st.Month, n-1))
								if !model.valid(day) || p.Hidden != nil && p.Hidden(day) || p.HideOutsideDays && d.Outside || duplicate {
									ui.Box(ctx.Key(dayKey(day))).Size(32, 32).AccessHidden(true)
									continue
								}
								column := weekday + 1
								if p.WeekNumbers {
									column++
								}
								calendarDayButton(ctx, grid, st, p, d, column, commit)
							}
						})
					}
				})
			}
		})
		if p.Footer != nil {
			p.Footer()
		}
	})
	return root
}
func calendarMoveMonth(day time.Time, delta int) time.Time {
	last := time.Date(day.Year(), day.Month()+time.Month(delta)+1, 0, 0, 0, 0, 0, day.Location())
	return time.Date(last.Year(), last.Month(), min(day.Day(), last.Day()), 0, 0, 0, 0, day.Location())
}

func calendarDayButton(c *ui.Context, grid ui.Element, st *calendarState, p CalendarProps, d CalendarDay, column int, commit func(time.Time)) {
	t := theme.TokensOf(c)
	model := modelFor(p.System)
	label := model.label(c.Locale(), d.Date, false)
	slot := ui.Box(c.Key(dayKey(d.Date))).Size(32, 32).Role(ui.RoleCell).Label(label).AccessSelected(d.Selected).AccessDisabled(d.Disabled).ColumnIndex(column)
	if d.RangeStart || d.RangeEnd || d.RangeMiddle {
		slot.Background(t.Accent).Clip()
		firstColumn, lastColumn := 1, 7
		if p.WeekNumbers {
			firstColumn++
			lastColumn++
		}
		left, right := d.RangeStart || d.RangeMiddle && column == firstColumn, d.RangeEnd || d.RangeMiddle && column == lastColumn
		if c.Direction() == ui.RightToLeft {
			left, right = right, left
		}
		if left && !right {
			slot.Radius(t.RadiusMD(), 0, 0, t.RadiusMD())
		} else if right && !left {
			slot.Radius(0, t.RadiusMD(), t.RadiusMD(), 0)
		} else if left && right {
			slot.Radius(t.RadiusMD())
		}
	}
	slot.Children(func() {
		cell := motion.ApplyTransition(c, ui.ButtonBase(c.Key("button")).Size(32, 32).Radius(t.RadiusMD()).FontSize(14).TabStop(false).Label(label).Disabled(d.Disabled).Checked(d.Selected))
		if d.Selected {
			cell.Background(t.Primary).TextColor(t.PrimaryForeground)
		} else if d.Today || cell.Hovered() {
			cell.Background(t.Accent).TextColor(t.AccentForeground)
		}
		if d.RangeMiddle {
			cell.Radius(0).Background(t.Accent).TextColor(t.AccentForeground)
		}
		if d.Outside {
			cell.Opacity(.5)
		}
		if d.Disabled {
			cell.Opacity(.35).Cursor(ui.CursorNotAllowed)
		}
		if d.Today {
			cell.Current(ui.CurrentDate)
		}
		if sameDate(d.Date, st.Cursor) {
			grid.ActiveDescendant(slot)
			if grid.FocusVisible() {
				cell.Border(1, t.Ring)
			}
		}
		cell.Children(func() {
			if p.DayContent != nil {
				p.DayContent(c, d)
			} else {
				ui.Text(c, model.dayNumber(d.Date)).FontFeatures("tnum")
			}
		})
		if cell.Clicked() {
			commit(d.Date)
			grid.Focus()
		}
	})
}

func calendarCaption(c *ui.Context, month time.Time, index int, p CalendarProps, loc *time.Location, setMonth func(time.Time)) {
	model := modelFor(p.System)
	year, monthIndex, _ := model.parts(month)
	selectCaption := p.CaptionSelect
	if selectCaption == nil {
		selectCaption = control.Select
	}
	months := model.months(c.Locale())
	monthValue := strconv.Itoa(monthIndex)
	yearValue := strconv.Itoa(year)
	lo, hi := calendarBounds(p, loc)
	firstYear, lastYear := year-100, year+100
	if !lo.IsZero() {
		firstYear, _, _ = model.parts(lo)
	}
	if !hi.IsZero() {
		lastYear, _, _ = model.parts(hi)
	}
	if p.System == CalendarPersian {
		firstYear = max(1, firstYear)
		lastYear = min(maxPersianYear, lastYear)
	}
	ui.Row(c).Gap(4).PaddingX(28).Children(func() {
		options := make([]control.Option, 12)
		for i := range options {
			candidate := model.date(year, i+1, 1, loc)
			options[i] = control.Option{Value: strconv.Itoa(i + 1), Label: months[i], Disabled: !lo.IsZero() && candidate.Before(lo) || !hi.IsZero() && candidate.After(hi)}
		}
		if p.Caption != CalendarCaptionDropdownYears {
			if selectCaption(c, &monthValue, control.SelectProps{Label: "Month", Options: options, Disabled: p.DisableNavigation}).Height(28).MinHeight(28).PaddingX(4).FontSize(12).Changed() {
				m, _ := strconv.Atoi(monthValue)
				setMonth(model.date(year, m-index, 1, loc))
			}
		} else {
			ui.Text(c, months[monthIndex-1]).FontSize(14).FontWeight(500)
		}
		options = make([]control.Option, 0, max(0, lastYear-firstYear+1))
		for year := firstYear; year <= lastYear; year++ {
			options = append(options, control.Option{Value: strconv.Itoa(year), Label: model.yearNumber(year)})
		}
		if p.Caption != CalendarCaptionDropdownMonths {
			if selectCaption(c, &yearValue, control.SelectProps{Label: "Year", Options: options, Disabled: p.DisableNavigation}).Height(28).MinHeight(28).PaddingX(4).FontSize(12).Changed() {
				year, _ := strconv.Atoi(yearValue)
				setMonth(model.date(year, monthIndex-index, 1, loc))
			}
		} else {
			ui.Text(c, model.yearNumber(year)).FontSize(14).FontWeight(500)
		}
	})
}

type DatePickerProps struct {
	Placeholder, Label string
	Disabled           bool
	Format             string
	Calendar           CalendarProps
	// ShowTime composes MyGo's native segmented time editor. Choosing a day
	// keeps the current time; changing time preserves the selected civil date.
	ShowTime     bool
	TimeLabel    string
	Presets      []DatePreset
	RangePresets []DateRangePreset
}
type DatePreset struct {
	Label    string
	Date     time.Time
	Disabled bool
}
type DateRangePreset struct {
	Label    string
	Range    DateRange
	Disabled bool
}
type datePickerState struct {
	Open    bool
	Working time.Time
	Range   DateRange
	Clocks  DateRange
}

func DatePicker(c *ui.Context, date *time.Time, p DatePickerProps) ui.Element {
	if date == nil {
		panic("display: DatePicker requires a date pointer")
	}
	t := theme.TokensOf(c)
	root := ui.Column(c).Gap(8)
	state := ui.Local(root, "date-picker", func() datePickerState { return datePickerState{} })
	root.Children(func() {
		text := p.Placeholder
		if text == "" {
			text = "Pick a date"
		}
		format := p.Format
		if format == "" {
			format = "Jan 2, 2006"
		}
		loc := calendarLocation(p.Calendar, *date)
		if !date.IsZero() {
			text = date.In(loc).Format(format)
		}
		label := p.Label
		if label == "" {
			label = text
		}
		anchor := ui.ButtonBase(c).Padding(8, 12).Height(36).Radius(t.RadiusMD()).Border(1, t.Input).Disabled(p.Disabled).Label(label).Children(func() { ui.Text(c, text) })
		if anchor.Clicked() {
			state.Open = !state.Open
			if state.Open {
				state.Working = *date
			}
		}
		if p.Disabled {
			state.Open = false
			anchor.Opacity(.5)
		}
		anchor.Expanded(state.Open)
		panel := ui.PopoverBase(c, anchor, &state.Open, func(panel ui.Element) {
			panel.Background(t.Popover).TextColor(t.PopoverForeground).Padding(4).Margin(4, 0, 0, 0).Border(1, t.Border).Radius(t.Radius).Transition(motion.Popup(c)).Children(func() {
				if len(p.Presets) > 0 {
					ui.Row(c).Gap(4).Padding(8).Children(func() {
						for _, preset := range p.Presets {
							day := dateIn(preset.Date, loc)
							candidate := withCalendarDay(*date, day, loc)
							if ui.ButtonBase(c.Key(preset.Label)).Padding(6, 8).Radius(t.RadiusSM()).Disabled(preset.Disabled || day.IsZero() || unselectableDate(candidate, p.Calendar)).Children(func() { ui.Text(c, preset.Label) }).Clicked() {
								*date = candidate
								state.Open = false
								root.SetChanged()
								anchor.Focus()
							}
						}
					})
				}
				props := p.Calendar
				props.Browse = true
				props.AutoFocus = true
				if CalendarWith(c, &state.Working, props).Changed() {
					if !state.Working.IsZero() && unselectableDate(state.Working, p.Calendar) {
						state.Working = *date
						return
					}
					*date = state.Working
					state.Open = false
					root.SetChanged()
					anchor.Focus()
				}
			})
		})
		if panel.Dismissed() {
			anchor.Focus()
		}
		if p.ShowTime {
			label := p.TimeLabel
			if label == "" {
				label = "Time"
			}
			control.Label(c, label)
			working := date.In(loc)
			if date.IsZero() {
				working = c.Now().In(loc)
			}
			if ui.TimeInput(c, &working).Label(label).Disabled(p.Disabled || date.IsZero()).Invalid(!date.IsZero() && unselectableDate(working, p.Calendar)).Changed() && !unselectableDate(working, p.Calendar) {
				*date = working
				root.SetChanged()
			}
		}
	})
	return root
}

func DateRangePicker(c *ui.Context, date *DateRange, p DatePickerProps) ui.Element {
	if date == nil {
		panic("display: DateRangePicker requires a range pointer")
	}
	t := theme.TokensOf(c)
	root := ui.Column(c)
	state := ui.Local(root, "date-range-picker", func() datePickerState { return datePickerState{} })
	root.Children(func() {
		text := p.Placeholder
		if text == "" {
			text = "Pick a date range"
		}
		format := p.Format
		if format == "" {
			format = "Jan 2, 2006"
		}
		loc := calendarLocation(p.Calendar, date.From)
		if !date.From.IsZero() {
			text = date.From.In(loc).Format(format)
			if !date.To.IsZero() {
				text += " – " + date.To.In(loc).Format(format)
			}
		}
		label := p.Label
		if label == "" {
			label = text
		}
		anchor := ui.ButtonBase(c).Padding(8, 12).Height(36).Radius(t.RadiusMD()).Border(1, t.Input).Disabled(p.Disabled).Label(label).Children(func() { ui.Text(c, text) })
		if anchor.Clicked() {
			state.Open = !state.Open
			if state.Open {
				state.Range = *date
				state.Clocks = *date
			}
		}
		if p.Disabled {
			state.Open = false
			anchor.Opacity(.5)
		}
		anchor.Expanded(state.Open)
		panel := ui.PopoverBase(c, anchor, &state.Open, func(panel ui.Element) {
			panel.Background(t.Popover).TextColor(t.PopoverForeground).Padding(4).Margin(4, 0, 0, 0).Border(1, t.Border).Radius(t.Radius).Transition(motion.Popup(c)).Children(func() {
				if len(p.RangePresets) > 0 {
					ui.Row(c).Gap(4).Padding(8).Children(func() {
						for _, preset := range p.RangePresets {
							r := DateRange{From: dateIn(preset.Range.From, loc), To: dateIn(preset.Range.To, loc)}
							invalid := r.To.Before(r.From) || unselectableDate(r.From, p.Calendar) || unselectableDate(r.To, p.Calendar)
							nights := calendarDaysBetween(r.From, r.To)
							invalid = invalid || nights < max(0, p.Calendar.Min) || p.Calendar.Max > 0 && nights > p.Calendar.Max
							if !invalid && p.Calendar.ExcludeDisabled {
								for day := r.From; !day.After(r.To); day = day.AddDate(0, 0, 1) {
									if disabledDay(day, p.Calendar) {
										invalid = true
										break
									}
								}
							}
							if ui.ButtonBase(c.Key(preset.Label)).Padding(6, 8).Radius(t.RadiusSM()).Disabled(preset.Disabled || invalid).Children(func() { ui.Text(c, preset.Label) }).Clicked() {
								*date = r
								state.Range = r
								state.Open = false
								root.SetChanged()
								anchor.Focus()
							}
						}
					})
				}
				props := p.Calendar
				props.Browse = true
				props.AutoFocus = true
				if props.NumberOfMonths <= 0 {
					props.NumberOfMonths = 2
				}
				if CalendarRange(c, &state.Range, props).Changed() {
					next := state.Range
					if p.ShowTime {
						if !next.From.IsZero() {
							next.From = withCalendarDay(state.Clocks.From, next.From, loc)
						}
						if !next.To.IsZero() {
							next.To = withCalendarDay(state.Clocks.To, next.To, loc)
						}
					}
					*date = next
					state.Range = next
					root.SetChanged()
					if !date.From.IsZero() && !date.To.IsZero() {
						state.Open = false
						anchor.Focus()
					}
				}
			})
		})
		if panel.Dismissed() {
			anchor.Focus()
		}
		if p.ShowTime {
			ui.Row(c).Gap(12).Children(func() {
				for index, label := range []string{"Start time", "End time"} {
					value := &date.From
					if index == 1 {
						value = &date.To
					}
					working := dateIn(*value, loc)
					ui.Column(c.Key(index)).Gap(4).Children(func() {
						control.Label(c, label)
						if ui.TimeInput(c, &working).Label(label).Disabled(p.Disabled || value.IsZero()).Invalid(!value.IsZero() && unselectableDate(working, p.Calendar)).Changed() && !unselectableDate(working, p.Calendar) {
							candidate := *date
							if index == 0 {
								candidate.From = working
							} else {
								candidate.To = working
							}
							if !candidate.From.IsZero() && !candidate.To.IsZero() && candidate.To.Before(candidate.From) {
								return
							}
							*value = working
							state.Range = *date
							root.SetChanged()
						}
					})
				}
			})
		}
	})
	return root
}

// String uses a stable date-only representation for logging and form summaries.
func (r DateRange) String() string {
	if r.From.IsZero() {
		return ""
	}
	if r.To.IsZero() {
		return dayKey(r.From)
	}
	return fmt.Sprintf("%s – %s", dayKey(r.From), dayKey(r.To))
}
