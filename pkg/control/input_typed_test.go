package control_test

import (
	"errors"
	"math"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func typedTester(view func(*ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Padding(20).Gap(12).Children(func() { view(c) }) }, 700, 600)
}
func typedClick(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	if err := r.Click(name); err != nil {
		t.Fatal(err)
	}
	r.Frame()
}
func typedRoleClick(t *testing.T, r *ui.Tester, role ui.Role, name string) {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing role %v %q", role, name)
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
}
func numberText(t *testing.T, r *ui.Tester, name string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(ui.RoleTextField, name)
	if !ok {
		t.Fatalf("missing real editor %q: %+v", name, r.AccessAll())
	}
	return a
}
func replaceNumber(r *ui.Tester, text string) { r.Key(ui.Cmd, ui.KeyA); r.Type(text); r.Frame() }

func TestNumberInputTypedDraftSelectionSubmitAndBlur(t *testing.T) {
	n := 12.
	changes, submits := []float64{}, []float64{}
	r := typedTester(func(c *ui.Context) {
		control.NumberInput(c, &n, control.NumberInputProps{Label: "Amount", AutoFocus: true, OnChange: func(v float64) { changes = append(changes, v) }, OnSubmit: func(v float64) { submits = append(submits, v) }})
		control.Button(c, "Outside", control.ButtonProps{})
	})
	replaceNumber(r, "-")
	if n != 12 || numberText(t, r, "Amount").Value != "-" || !numberText(t, r, "Amount").Invalid || len(changes) != 0 {
		t.Fatalf("incomplete draft affected value: %v %+v %v", n, numberText(t, r, "Amount"), changes)
	}
	r.Type("1.50")
	if n != -1.5 || numberText(t, r, "Amount").Value != "-1.50" {
		t.Fatalf("numeric draft lost native formatting: %v %+v", n, numberText(t, r, "Amount"))
	}
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !slices.Equal(submits, []float64{-1.5}) || numberText(t, r, "Amount").Value != "-1.5" {
		t.Fatalf("submit %v %+v", submits, numberText(t, r, "Amount"))
	}
	replaceNumber(r, "4e")
	if n != -1.5 {
		t.Fatal("partial exponent was committed")
	}
	typedClick(t, r, "Outside")
	if numberText(t, r, "Amount").Value != "-1.5" || numberText(t, r, "Amount").Invalid {
		t.Fatal("blur failed to restore last numeric value")
	}
	n = 42
	r.Frame()
	if numberText(t, r, "Amount").Value != "42" || len(submits) != 1 {
		t.Fatal("application state replacement failed")
	}
}

func TestNumberInputCompositionCandidateEnterAndNativeUndo(t *testing.T) {
	n := 1.
	submits := 0
	r := typedTester(func(c *ui.Context) {
		control.NumberInput(c, &n, control.NumberInputProps{Label: "Composing amount", AutoFocus: true, OnSubmit: func(float64) { submits++ }})
	})
	r.Key(ui.Cmd, ui.KeyA)
	r.Compose("25", 2)
	if n != 1 || submits != 0 {
		t.Fatal("marked numeric text committed early")
	}
	r.Key(0, ui.KeyEnter)
	if n != 25 || submits != 0 {
		t.Fatalf("candidate Enter also submitted: %v %d", n, submits)
	}
	r.Key(0, ui.KeyEnter)
	if submits != 1 {
		t.Fatal("ordinary committed Enter did not submit")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("84")
	r.Key(ui.Cmd, ui.KeyZ)
	if n != 25 {
		t.Fatalf("native undo did not restore numeric model: %v", n)
	}
}

func TestNumberInputBoundsSteppingReadOnlyAndFieldGroupTarget(t *testing.T) {
	n := 2.
	lo, hi := 0., 3.
	readOnly, disabled := false, false
	r := typedTester(func(c *ui.Context) {
		control.Field(c, control.FieldProps{Label: "Price", Description: "Unit price", Error: "Check price", Disabled: disabled}, func() ui.Element {
			return control.InputGroup(c, control.InputGroupProps{Prefix: "USD"}, func() ui.Element {
				return control.NumberInput(c, &n, control.NumberInputProps{Min: &lo, Max: &hi, Step: .5, ReadOnly: readOnly})
			})
		})
	})
	a := numberText(t, r, "Price")
	if !a.Invalid || !slices.Contains(a.ErrorMessage, "Check price") || !slices.Contains(a.DescribedBy, "Unit price") || a.Min != 0 || a.Max != 3 || a.Step != .5 {
		t.Fatalf("metadata missed actual native editor: %+v", a)
	}
	typedClick(t, r, "USD")
	r.Key(0, ui.KeyUp)
	if n != 2.5 {
		t.Fatalf("addon did not focus/step editor: %v", n)
	}
	r.Key(0, ui.KeyUp)
	r.Key(0, ui.KeyUp)
	if n != 3 {
		t.Fatal("step exceeded bound")
	}
	replaceNumber(r, "9")
	if n != 3 {
		t.Fatal("out-of-range draft committed")
	}
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if numberText(t, r, "Price").Value != "3" {
		t.Fatal("invalid submit not canonicalized")
	}
	readOnly = true
	r.Frame()
	r.Key(0, ui.KeyDown)
	replaceNumber(r, "1")
	if n != 3 || !numberText(t, r, "Price").ReadOnly {
		t.Fatal("read-only accepted native edits")
	}
	disabled = true
	readOnly = false
	r.Frame()
	r.Key(0, ui.KeyDown)
	if n != 3 || !numberText(t, r, "Price").Disabled {
		t.Fatal("disabled ancestor accepted numeric step")
	}
}

func TestDateInputNativePopupKeyboardCancelAndCivilTime(t *testing.T) {
	loc := time.FixedZone("fixture", 9*3600)
	date := time.Date(2026, 11, 10, 15, 30, 12, 123, loc)
	before := date
	changes := 0
	r := typedTester(func(c *ui.Context) {
		control.Field(c, control.FieldProps{Label: "Date", Description: "Choose a day"}, func() ui.Element {
			return control.DateInput(c, &date, control.DateInputProps{OnChange: func(time.Time) { changes++ }})
		})
	})
	a, ok := r.AccessByRole(ui.RolePopUpButton, "Date")
	if !ok || a.Bounds.H != 36 || !slices.Contains(a.DescribedBy, "Choose a day") {
		t.Fatalf("date field %+v", a)
	}
	typedRoleClick(t, r, ui.RolePopUpButton, "Date")
	r.Key(0, ui.KeyEscape)
	if !date.Equal(before) || changes != 0 {
		t.Fatal("cancel changed date")
	}
	typedRoleClick(t, r, ui.RolePopUpButton, "Date")
	typedClick(t, r, "November 12, 2026")
	if date.Day() != 12 || date.Hour() != 15 || date.Minute() != 30 || date.Second() != 12 || date.Nanosecond() != 123 || date.Location() != loc || changes != 1 {
		t.Fatalf("date selection lost civil time %v (%d)", date, changes)
	}
	typedRoleClick(t, r, ui.RolePopUpButton, "Date")
	r.Key(0, ui.KeyRight)
	r.Key(0, ui.KeyEnter)
	if date.Day() != 13 || changes != 2 {
		t.Fatalf("native date keyboard did not select %v %d", date, changes)
	}
}

func TestTimeInputNativeSegmentsFieldNamesAndDisabled(t *testing.T) {
	loc := time.FixedZone("fixture", 8*3600)
	v := time.Date(2026, 11, 10, 7, 30, 0, 0, loc)
	changes := 0
	disabled := false
	r := typedTester(func(c *ui.Context) {
		control.Field(c, control.FieldProps{Label: "Alarm", Description: "Local time", Disabled: disabled, Error: "Check time"}, func() ui.Element {
			return control.TimeInput(c, &v, control.TimeInputProps{OnChange: func(time.Time) { changes++ }})
		})
	})
	r.SetRegion(i18n.Region{Locale: "en-GB", HourCycle: 24, FirstWeekday: time.Monday})
	a, ok := r.AccessByRole(ui.RoleStepper, "Alarm hours")
	if !ok || a.Now != 7 || a.Max != 23 {
		t.Fatalf("field label did not reach segment name: %+v", r.AccessAll())
	}
	typedRoleClick(t, r, ui.RoleStepper, "Alarm hours")
	r.Key(0, ui.KeyUp)
	if v.Hour() != 8 || changes != 1 {
		t.Fatalf("native Up %v %d", v, changes)
	}
	r.Key(0, ui.Key1)
	r.Key(0, ui.Key5)
	r.Key(0, ui.Key4)
	r.Key(0, ui.Key5)
	if v.Hour() != 15 || v.Minute() != 45 || v.Day() != 10 || v.Location() != loc {
		t.Fatalf("segments lost typed/native semantics: %v", v)
	}
	group, ok := r.AccessByRole(ui.RoleGroup, "Alarm")
	if !ok || !group.Invalid || !slices.Contains(group.ErrorMessage, "Check time") || !slices.Contains(group.DescribedBy, "Local time") {
		t.Fatalf("native group metadata %+v", group)
	}
	disabled = true
	r.Frame()
	before := v
	r.Key(0, ui.KeyUp)
	if v != before {
		t.Fatal("disabled segment changed time")
	}
}

func waitFile(t *testing.T, r *ui.Tester, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !done() && time.Now().Before(deadline) {
		r.Frame()
		time.Sleep(time.Millisecond)
	}
	if !done() {
		t.Fatal("asynchronous file result did not arrive")
	}
}
func TestFileInputUserActivationAsyncCopyCancelAndError(t *testing.T) {
	paths := []string{"/old/previous.txt"}
	var calls atomic.Int32
	answers := make(chan fileAnswer, 4)
	launched := make(chan mygo.OpenDialogOptions, 4)
	changes, errs := 0, 0
	filters := []mygo.FileFilter{{Name: "Text", Extensions: []string{"txt"}}}
	pick := func(p mygo.OpenDialogOptions) ([]string, error) {
		calls.Add(1)
		launched <- p
		a := <-answers
		return a.paths, a.err
	}
	r := typedTester(func(c *ui.Context) {
		control.Field(c, control.FieldProps{Label: "Picture", Description: "Select an attachment"}, func() ui.Element {
			return control.FileInput(c, &paths, control.FileInputProps{Picker: pick, Filters: filters, Multiple: true, OnChange: func(p []string) { changes++; p[0] = "callback-owned" }, OnError: func(error) { errs++ }})
		})
	})
	r.Frame()
	if calls.Load() != 0 {
		t.Fatal("picker opened during render")
	}
	typedRoleClick(t, r, ui.RoleButton, "Picture")
	var opts mygo.OpenDialogOptions
	select {
	case opts = <-launched:
	case <-time.After(time.Second):
		t.Fatal("click did not launch native picker")
	}
	if !opts.Multiple || opts.Filters[0].Extensions[0] != "txt" {
		t.Fatalf("dialog opts %+v", opts)
	}
	filters[0].Extensions[0] = "mutated"
	if opts.Filters[0].Extensions[0] != "txt" {
		t.Fatal("dialog retained filter slice")
	}
	a, _ := r.AccessByRole(ui.RoleButton, "Picture")
	if !a.Busy || !a.Disabled {
		t.Fatalf("pending picker not busy/disabled %+v", a)
	}
	answers <- fileAnswer{paths: []string{"/tmp/a.txt", "/tmp/b.txt"}}
	waitFile(t, r, func() bool { return changes == 1 })
	if !slices.Equal(paths, []string{"/tmp/a.txt", "/tmp/b.txt"}) || !r.HasText("a.txt, b.txt") {
		t.Fatalf("result/callback copy or filenames %v", paths)
	}
	typedRoleClick(t, r, ui.RoleButton, "Picture")
	<-launched
	answers <- fileAnswer{}
	waitFile(t, r, func() bool { a, _ := r.AccessByRole(ui.RoleButton, "Picture"); return !a.Busy })
	if changes != 1 || len(paths) != 2 {
		t.Fatal("cancel cleared selection")
	}
	typedRoleClick(t, r, ui.RoleButton, "Picture")
	<-launched
	answers <- fileAnswer{err: errors.New("permission denied")}
	waitFile(t, r, func() bool { return errs == 1 })
	a, _ = r.AccessByRole(ui.RoleButton, "Picture")
	if !a.Invalid || !strings.Contains(a.Description, "permission denied") {
		t.Fatalf("picker error missed native field %+v", a)
	}
}

type fileAnswer struct {
	paths []string
	err   error
}

func TestFileInputUnmountAndRebindDiscardStaleWork(t *testing.T) {
	first, second := []string{"first.txt"}, []string{"second.txt"}
	current := &first
	shown := true
	calls := 0
	start := make(chan struct{}, 2)
	answer := make(chan fileAnswer, 2)
	returned := make(chan struct{}, 2)
	pick := func(mygo.OpenDialogOptions) ([]string, error) {
		start <- struct{}{}
		a := <-answer
		defer func() { returned <- struct{}{} }()
		return a.paths, a.err
	}
	r := typedTester(func(c *ui.Context) {
		if shown {
			control.FileInput(c, current, control.FileInputProps{Label: "Attachment", Picker: pick, OnChange: func([]string) { calls++ }, OnError: func(error) { calls++ }})
		}
	})
	typedClick(t, r, "Attachment")
	<-start
	current = &second
	r.Frame()
	answer <- fileAnswer{paths: []string{"stale-rebound.txt"}}
	<-returned
	r.Frame()
	if calls != 0 || second[0] != "second.txt" || first[0] != "first.txt" {
		t.Fatal("rebound owner consumed stale file result")
	}
	typedClick(t, r, "Attachment")
	<-start
	shown = false
	r.Frame()
	answer <- fileAnswer{err: errors.New("stale-unmounted")}
	<-returned
	r.Frame()
	shown = true
	r.Frame()
	if calls != 0 || second[0] != "second.txt" {
		t.Fatal("removed control consumed stale callback")
	}
	typedClick(t, r, "Attachment")
	<-start
	answer <- fileAnswer{paths: []string{"fresh.txt"}}
	<-returned
	waitFile(t, r, func() bool { return calls == 1 })
	if second[0] != "fresh.txt" {
		t.Fatal("remount did not accept fresh picker")
	}
}

func TestNumberInputInvalidConfigurationRecoveryAndEditorHandle(t *testing.T) {
	n, lo, hi := 2., 3., 1.
	step := -1.
	changes := 0
	var handle ui.Handle
	r := typedTester(func(c *ui.Context) {
		if control.NumberInput(c, &n, control.NumberInputProps{Label: "Configured number", Min: &lo, Max: &hi, Step: step, Handle: &handle}).Changed() {
			changes++
		}
		if control.Button(c, "Focus number", control.ButtonProps{}).Clicked() {
			handle.Focus(c)
		}
	})
	a := numberText(t, r, "Configured number")
	if !a.Invalid || !a.ReadOnly {
		t.Fatal("invalid numeric configuration not explicit")
	}
	lo, hi, step = 0, 10, 1
	r.Frame()
	typedClick(t, r, "Focus number")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("6")
	if n != 6 || numberText(t, r, "Configured number").Invalid {
		t.Fatal("recovered configuration/real editor Handle not functional")
	}
	before := changes
	n = math.NaN()
	r.Frame()
	r.Frame()
	if changes != before || !numberText(t, r, "Configured number").Invalid {
		t.Fatal("nonfinite model fabricated an input change or lost invalid state")
	}
	n = 5
	r.Frame()
	if numberText(t, r, "Configured number").Invalid {
		t.Fatal("finite application replacement did not recover")
	}
}

func TestFileInputKeyboardActivationSingleSelectionAndCurrentCallback(t *testing.T) {
	paths := []string{}
	launch := make(chan struct{}, 1)
	result := make(chan []string, 1)
	callback := "old"
	got := ""
	picker := func(mygo.OpenDialogOptions) ([]string, error) { launch <- struct{}{}; return <-result, nil }
	r := typedTester(func(c *ui.Context) {
		name := callback
		control.FileInput(c, &paths, control.FileInputProps{Label: "Keyboard attachment", Picker: picker, OnChange: func([]string) { got = name }})
	})
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeySpace)
	select {
	case <-launch:
	case <-time.After(time.Second):
		t.Fatal("keyboard did not activate file input")
	}
	callback = "current"
	r.Frame()
	result <- []string{"one.txt", "two.txt"}
	waitFile(t, r, func() bool { return got != "" })
	if got != "current" || !slices.Equal(paths, []string{"one.txt"}) {
		t.Fatalf("single-file/current callback semantics %q %v", got, paths)
	}
}
