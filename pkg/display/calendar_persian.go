package display

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
)

// CalendarSystem selects a civil month model. Values remain ordinary time.Time
// instants in the selected Location in both systems.
type CalendarSystem uint8

const (
	CalendarGregorian CalendarSystem = iota
	// CalendarPersian matches date-fns-jalali 4.1.0-0's 33-year arithmetic
	// Persian/Jalali model. It is not an Islamic lunar Hijri calendar.
	CalendarPersian
)

// PersianDate is a civil Persian date. Positive years 1..275138 are complete
// years within the fixed upstream JavaScript Date range. The model deliberately
// follows the pinned arithmetic algorithm rather than an astronomical calendar.
type PersianDate struct{ Year, Month, Day int }

const maxPersianYear = 275138
const persianEpoch int64 = 1948320

var persianMonthDays = [...]int64{0, 31, 62, 93, 124, 155, 186, 216, 246, 276, 306, 336}
var persianMonths = [...]string{"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور", "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند"}
var persianWeekdays = [...]string{"یک‌شنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنج‌شنبه", "جمعه", "شنبه"}
var persianShortWeekdays = [...]string{"۱ش", "۲ش", "۳ش", "۴ش", "۵ش", "ج", "ش"}

// The conversion formulas below are adapted from date-fns-jalali 4.1.0-0,
// _lib/jalali.js (MIT). Its original and license are archived in testdata.
// Go integer division, like the source's div(), truncates toward zero.
func normalizePersianMonth(year, month int) (int, int) {
	month--
	if month < 0 {
		old := month
		month = ((month % 12) + 12) % 12
		year -= (month - old) / 12
	}
	if month > 11 {
		year += month / 12
		month %= 12
	}
	return year, month + 1
}
func persianDayNumber(year, month, day int) int64 {
	year, month = normalizePersianMonth(year, month)
	y := int64(year)
	return persianEpoch - 1 + 365*(y-1) + (8*y+21)/33 + persianMonthDays[month-1] + int64(day)
}
func gregorianDayNumber(year, month, day int) int64 {
	year, month = normalizePersianMonth(year, month)
	y, m, d := int64(year), int64(month), int64(day)
	return 1461*(y+4800+(m-14)/12)/4 + 367*(m-2-12*((m-14)/12))/12 - 3*((y+4900+(m-14)/12)/100)/4 + d - 32075
}
func persianFromDayNumber(number int64) PersianDate {
	days := number - persianEpoch
	year := 1 + (33*days+3)/12053
	dayOfYear := days - (365*(year-1) + (8*year+21)/33)
	if dayOfYear < 0 {
		year--
		dayOfYear = days - (365*(year-1) + (8*year+21)/33)
	}
	month := dayOfYear / 31
	if dayOfYear >= 216 {
		month = (dayOfYear - 6) / 30
	}
	if month < 0 || month > 11 {
		return PersianDate{}
	}
	return PersianDate{int(year), int(month) + 1, int(dayOfYear - persianMonthDays[month] + 1)}
}
func gregorianFromDayNumber(number int64) (int, int, int) {
	l := number + 68569
	n := 4 * l / 146097
	l -= (146097*n + 3) / 4
	i := 4000 * (l + 1) / 1461001
	l = l - 1461*i/4 + 31
	j := 80 * l / 2447
	day := l - 2447*j/80
	l = j / 11
	return int(100*(n-49) + i + l), int(j + 2 - 12*l), int(day)
}
func persianTime(year, month, day int, loc *time.Location) time.Time {
	y, m, d := gregorianFromDayNumber(persianDayNumber(year, month, day))
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
}

// PersianDateOf reads the date's local civil fields without changing its instant
// or location. A zero time is an unselected value in the calendar APIs.
func PersianDateOf(date time.Time) (PersianDate, error) {
	if date.IsZero() || date.Year() < 622 || date.Year() > 275760 {
		return PersianDate{}, fmt.Errorf("display: date is outside the supported Persian calendar")
	}
	p := persianFromDayNumber(gregorianDayNumber(date.Year(), int(date.Month()), date.Day()))
	if p.Year < 1 || p.Year > maxPersianYear {
		return PersianDate{}, fmt.Errorf("display: Persian year must be between 1 and %d", maxPersianYear)
	}
	return p, nil
}

// Time converts a validated Persian date to local midnight. A nil location uses
// time.Local. Invalid days are rejected instead of silently overflowing a month.
func (p PersianDate) Time(location *time.Location) (time.Time, error) {
	if p.Year < 1 || p.Year > maxPersianYear || p.Month < 1 || p.Month > 12 {
		return time.Time{}, fmt.Errorf("display: invalid Persian year/month")
	}
	length := int(persianDayNumber(p.Year, p.Month+1, 1) - persianDayNumber(p.Year, p.Month, 1))
	if p.Day < 1 || p.Day > length {
		return time.Time{}, fmt.Errorf("display: invalid Persian day")
	}
	if location == nil {
		location = time.Local
	}
	date := persianTime(p.Year, p.Month, p.Day, location)
	// A civil date skipped by a time-zone transition must not become another
	// selected day. Ordinary DST folds/gaps retain time.Time's native policy.
	actual, err := PersianDateOf(date)
	if err != nil || actual != p {
		return time.Time{}, fmt.Errorf("display: Persian date does not exist in the requested location")
	}
	return date, nil
}

type calendarModel struct{ system CalendarSystem }

func modelFor(system CalendarSystem) calendarModel {
	if system != CalendarGregorian && system != CalendarPersian {
		panic("display: unknown CalendarSystem")
	}
	return calendarModel{system}
}
func (m calendarModel) parts(date time.Time) (int, int, int) {
	if m.system == CalendarGregorian {
		return date.Year(), int(date.Month()), date.Day()
	}
	p := persianFromDayNumber(gregorianDayNumber(date.Year(), int(date.Month()), date.Day()))
	return p.Year, p.Month, p.Day
}
func (m calendarModel) valid(date time.Time) bool {
	if m.system == CalendarGregorian || date.IsZero() {
		return true
	}
	_, err := PersianDateOf(date)
	return err == nil
}
func (m calendarModel) date(year, month, day int, loc *time.Location) time.Time {
	if m.system == CalendarGregorian {
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
	}
	return persianTime(year, month, day, loc)
}
func (m calendarModel) start(date time.Time) time.Time {
	if date.IsZero() {
		return date
	}
	y, month, _ := m.parts(date)
	return m.date(y, month, 1, date.Location())
}
func (m calendarModel) add(date time.Time, months int) time.Time {
	if m.system == CalendarGregorian {
		return date.AddDate(0, months, 0)
	}
	y, month, day := m.parts(date)
	return m.date(y, month+months, day, date.Location())
}
func (m calendarModel) move(date time.Time, months int) time.Time {
	y, month, day := m.parts(date)
	y, month = normalizePersianMonth(y, month+months)
	start := m.date(y, month, 1, date.Location())
	last := m.add(start, 1).AddDate(0, 0, -1)
	_, _, limit := m.parts(last)
	return m.date(y, month, min(day, limit), date.Location())
}
func (m calendarModel) difference(a, b time.Time) int {
	ay, am, _ := m.parts(a)
	by, bm, _ := m.parts(b)
	return (ay-by)*12 + am - bm
}
func (m calendarModel) label(locale string, date time.Time, caption bool) string {
	if m.system == CalendarGregorian {
		skeleton := "yMMMMd"
		if caption {
			skeleton = "yMMMM"
		}
		return i18n.FormatDate(locale, date, i18n.DateOptions{Skeleton: skeleton})
	}
	y, month, day := m.parts(date)
	if month < 1 || month > 12 {
		return ""
	}
	if caption {
		return persianMonths[month-1] + " " + persianNumber(y)
	}
	return persianWeekdays[date.Weekday()] + " " + persianNumber(day) + "-ام " + persianMonths[month-1] + " " + persianNumber(y)
}
func persianNumber(value int) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return '۰' + r - '0'
		}
		return r
	}, strconv.Itoa(value))
}
func (m calendarModel) dayNumber(date time.Time) string {
	_, _, day := m.parts(date)
	if m.system == CalendarPersian {
		return persianNumber(day)
	}
	return strconv.Itoa(day)
}
func (m calendarModel) months(locale string) []string {
	if m.system == CalendarPersian {
		return persianMonths[:]
	}
	names := i18n.Months(locale, i18n.Wide, true)
	return names[:]
}
func (m calendarModel) weekdays(locale string) []string {
	if m.system == CalendarPersian {
		return persianShortWeekdays[:]
	}
	names := i18n.Weekdays(locale, i18n.Short, true)
	return names[:]
}
func (m calendarModel) firstWeekday(c *ui.Context, p CalendarProps) time.Weekday {
	if p.WeekStartsOn != nil {
		return *p.WeekStartsOn
	}
	if m.system == CalendarPersian {
		return time.Saturday
	}
	return c.Region().FirstWeekday
}
func (m calendarModel) week(date time.Time, first time.Weekday) string {
	if m.system == CalendarGregorian {
		_, week := date.ISOWeek()
		return strconv.Itoa(week)
	}
	year, _, _ := m.parts(date)
	startOfWeek := func(d time.Time) time.Time { return d.AddDate(0, 0, -((int(d.Weekday()) - int(first) + 7) % 7)) }
	start := startOfWeek(m.date(year, 1, 1, date.Location()))
	next := startOfWeek(m.date(year+1, 1, 1, date.Location()))
	if !date.Before(next) {
		return persianNumber(1)
	}
	return persianNumber(calendarDaysBetween(start, date)/7 + 1)
}

func (m calendarModel) yearNumber(year int) string {
	if m.system == CalendarPersian {
		return persianNumber(year)
	}
	return strconv.Itoa(year)
}
