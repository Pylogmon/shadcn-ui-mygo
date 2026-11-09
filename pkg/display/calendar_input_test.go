package display

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
)

func TestDatePickerInputNativeParseDraftBrowseAndSelection(t *testing.T) {
	loc := time.FixedZone("Recipe", 8*3600)
	value := time.Date(2025, 6, 1, 14, 35, 10, 123, loc)
	state := DatePickerInputState{Text: "June 01, 2025"}
	version, parses, changes, submits := 1, []int{}, 0, 0
	var parts DatePickerInputParts
	var selection [2]int
	r := displayTester(func(c *ui.Context) {
		v := version
		control.Field(c, control.FieldProps{Label: "Subscription date", Error: state.Error}, func() ui.Element {
			parts = DatePickerInput(c, &value, &state, DatePickerInputProps{Input: control.InputProps{Label: "Date query", OnSubmit: func(string) { submits++ }}, TriggerLabel: "Choose subscription date", Calendar: CalendarProps{Location: loc}, Parse: func(text string) (time.Time, error) {
				parses = append(parses, v)
				date, err := time.ParseInLocation("January 02, 2006", text, loc)
				if err != nil {
					return time.Time{}, fmt.Errorf("Use June 12, 2025")
				}
				return time.Date(date.Year(), date.Month(), date.Day(), 14, 35, 10, 123, loc), nil
			}, OnChange: func(time.Time) { changes++ }})
			a, b := parts.Editor.TextSelection()
			selection = [2]int{a, b}
			return parts.Root
		})
	})
	clickAccessible(t, r, ui.RoleTextField, "Subscription date")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("June 12, 2025")
	r.Frame()
	if value.Day() != 12 || value.Hour() != 14 || value.Nanosecond() != 123 || value.Location() != loc || state.Error != "" || changes != 1 || len(parses) != 1 || parses[0] != 1 {
		t.Fatalf("valid parse value%v error%q callbacks%v/%d", value, state.Error, parses, changes)
	}
	before, month := value, state.Month
	version = 2
	r.Frame()
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("not a date")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if value != before || state.Month != month || state.Text != "not a date" || state.Error == "" || changes != 1 || submits != 1 || parses[len(parses)-1] != 2 {
		t.Fatalf("invalid draft/date/submit %+v value%v callbacks%v/%d submit%d", state, value, parses, changes, submits)
	}
	if n, _ := r.AccessByRole(ui.RoleTextField, "Subscription date"); !n.Invalid {
		t.Fatal("parse error was not exposed on actual native editor")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("June 12, 2025")
	r.Frame()
	caret := selection
	r.Key(0, ui.KeyDown)
	r.Frame()
	if !state.Open {
		t.Fatal("ArrowDown did not open Calendar")
	}
	clickAccessible(t, r, ui.RoleButton, "Next month")
	if state.Text != "June 12, 2025" || value != before || selection != caret {
		t.Fatalf("browsing changed date/text/native selection: %+v selection%v/%v", state, selection, caret)
	}
	clickAccessible(t, r, ui.RoleButton, "July 15, 2025")
	if value.Month() != time.July || value.Day() != 15 || value.Hour() != 14 || value.Nanosecond() != 123 || value.Location() != loc || state.Text != "July 15, 2025" || state.Open || state.Error != "" || changes != 2 {
		t.Fatalf("calendar publication value%v state%+v changes%d", value, state, changes)
	}
	if !r.Focused("Choose subscription date") {
		t.Fatal("calendar choice did not restore native trigger focus")
	}
}

func TestDatePickerInputPersianParseAndNativeComposition(t *testing.T) {
	value, _ := (PersianDate{1404, 3, 22}).Time(time.UTC)
	state := DatePickerInputState{Text: "1404/03/22"}
	parses, submits := 0, 0
	var inputs []string
	r := displayTester(func(c *ui.Context) {
		DatePickerInput(c, &value, &state, DatePickerInputProps{Input: control.InputProps{Label: "Persian input", OnSubmit: func(string) { submits++ }}, Calendar: CalendarProps{System: CalendarPersian, Location: time.UTC}, TriggerLabel: "Choose Persian date", Parse: func(text string) (time.Time, error) {
			parses++
			inputs = append(inputs, text)
			p := PersianDate{}
			if n, _ := fmt.Sscanf(text, "%d/%d/%d", &p.Year, &p.Month, &p.Day); n != 3 {
				return time.Time{}, fmt.Errorf("Use 1404/03/22")
			}
			return p.Time(time.UTC)
		}, Format: func(date time.Time) string {
			if date.IsZero() {
				return ""
			}
			p, _ := PersianDateOf(date)
			return fmt.Sprintf("%04d/%02d/%02d", p.Year, p.Month, p.Day)
		}})
	})
	clickAccessible(t, r, ui.RoleTextField, "Persian input")
	r.Key(ui.Cmd, ui.KeyA)
	r.Compose("1404/03/23", 10)
	r.Frame()
	if parses != 0 || submits != 0 {
		t.Fatalf("preedit parsed/submitted: %d/%d", parses, submits)
	}
	r.Type("1404/03/23")
	r.Frame()
	if p, _ := PersianDateOf(value); p != (PersianDate{1404, 3, 23}) || parses != 1 {
		t.Fatalf("native commit %+v parses%d text%q error%q inputs%v", p, parses, state.Text, state.Error, inputs)
	}
	// Native synthetic Enter commits the marked range, without submitting.
	// Sending InputText after that would be a separate insertion.
	r.Key(ui.Cmd, ui.KeyA)
	r.Compose("1404/03/23", 10)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if submits != 0 {
		t.Fatal("composition Enter submitted the editor")
	}

	r.Key(0, ui.KeyDown)
	r.Frame()
	if !state.Open || !r.HasText("خرداد ۱۴۰۴") {
		t.Fatal("parsed real date did not open Persian month")
	}
	model := modelFor(CalendarPersian)
	target, _ := (PersianDate{1404, 3, 24}).Time(time.UTC)
	clickAccessible(t, r, ui.RoleButton, model.label("en-US", target, false))
	if state.Text != "1404/03/24" || state.Open || value != target {
		t.Fatalf("Persian selection did not format/commit %+v %v", state, value)
	}
}
