package control_test

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
)

// One real-editor flow distinguishes an absent value from zero, follows native
// clearing/undo/composition, and keeps the existing typed range/readonly rules.
func TestNumberInputControlledEmptyNativeWorkflow(t *testing.T) {
	n, empty := 0., true
	lo, hi := 0., 10.
	readOnly, disabled := false, false
	rejectPresent, redirectStorage := false, false
	var presence []bool
	var changes, submits []float64
	r := typedTester(func(c *ui.Context) {
		control.Field(c, control.FieldProps{Label: "Optional amount", Disabled: disabled}, func() ui.Element {
			return control.NumberInput(c, &n, control.NumberInputProps{Empty: &empty, Placeholder: "123", AutoFocus: true, Min: &lo, Max: &hi, Step: 2, ReadOnly: readOnly,
				OnEmptyChange: func(value bool) {
					presence = append(presence, value)
					if !value && rejectPresent {
						empty = true
						if redirectStorage {
							n = 4
						}
					}
				},
				OnChange: func(value float64) { changes = append(changes, value) },
				OnSubmit: func(value float64) { submits = append(submits, value) },
			})
		})
		control.Button(c, "Outside optional amount", control.ButtonProps{})
	})
	field := func() ui.AccessInfo { return numberText(t, r, "Optional amount") }
	clear := func() { r.Key(ui.Cmd, ui.KeyA); r.Key(0, ui.KeyBackspace) }
	if !empty || n != 0 || field().Value != "" || field().Placeholder != "123" || field().Invalid || len(changes) != 0 {
		t.Fatalf("empty did not remain distinct from zero: %v %v %+v", empty, n, field())
	}
	r.Type("0")
	if empty || n != 0 || !slices.Equal(presence, []bool{false}) || !slices.Equal(changes, []float64{0}) {
		t.Fatalf("zero did not commit a present value once: %v %v %v %v", empty, n, presence, changes)
	}
	clear()
	if !empty || field().Value != "" || len(changes) != 1 || !slices.Equal(presence, []bool{false, true}) {
		t.Fatal("native clearing fabricated a zero value or missed absence")
	}
	r.Type("-")
	if !empty || !field().Invalid || field().Value != "-" {
		t.Fatal("incomplete numeric draft was published as a value")
	}
	typedClick(t, r, "Outside optional amount")
	if !empty || field().Value != "" || field().Invalid {
		t.Fatal("blur did not restore the last absent value")
	}
	typedClick(t, r, "Optional amount")
	r.Key(0, ui.KeyEnter)
	if len(submits) != 0 || !empty || field().Value != "" {
		t.Fatal("empty Enter fabricated a numeric submit")
	}
	r.Key(0, ui.KeyUp)
	if empty || n != 2 || field().Value != "2" || !slices.Equal(changes, []float64{0, 2}) {
		t.Fatal("stepping from absence did not start from zero and commit once")
	}
	replaceNumber(r, "9")
	r.Key(0, ui.KeyUp)
	if n != 10 || empty {
		t.Fatal("present optional value ignored the existing bound/step contract")
	}
	replaceNumber(r, "99")
	r.Key(0, ui.KeyEnter)
	if n != 10 || field().Value != "10" || len(submits) != 0 {
		t.Fatal("invalid optional submit did not restore the last present value")
	}
	r.Key(0, ui.KeyEnter)
	if !slices.Equal(submits, []float64{10}) {
		t.Fatal("valid present submit was not delivered once")
	}
	empty = true
	before := len(presence)
	r.Frame()
	if field().Value != "" || len(presence) != before || !r.Focused("Optional amount") {
		t.Fatal("controlled absence lost the editor/focus or emitted a user callback")
	}
	readOnly = true
	r.Frame()
	r.Type("3")
	r.Key(0, ui.KeyUp)
	if !empty || n != 10 || field().Value != "" {
		t.Fatal("readonly optional input accepted text/step")
	}
	readOnly, disabled = false, true
	r.Frame()
	r.Key(0, ui.KeyUp)
	if !empty || n != 10 || !field().Disabled {
		t.Fatal("disabled ancestor accepted optional numeric input")
	}
	disabled, empty, n = false, false, 6
	r.Frame()
	typedClick(t, r, "Optional amount")
	r.Key(ui.Cmd, ui.KeyA)
	before = len(presence)
	r.Compose("8", 1)
	if empty || n != 6 || len(presence) != before {
		t.Fatal("marked replacement transiently cleared or committed the optional model")
	}
	r.Key(0, ui.KeyEnter)
	if empty || n != 8 || len(submits) != 1 {
		t.Fatal("candidate Enter did not commit without submitting")
	}
	clear()
	if !empty || n != 8 {
		t.Fatal("clearing failed to retain the previous numeric storage")
	}
	r.Key(ui.Cmd, ui.KeyZ)
	if empty || n != 8 || field().Value != "8" {
		t.Fatal("native undo did not restore present numeric state")
	}

	// Rejecting presence keeps the old storage unless the callback explicitly
	// replaces it, and neither rejection publishes a numeric change or submit.
	n, empty, rejectPresent = 9, true, true
	r.Frame()
	before, beforeChanges, beforeSubmits := len(presence), len(changes), len(submits)
	r.Type("0")
	r.Frame()
	if !empty || n != 9 || field().Value != "" || field().Invalid || len(presence) != before+1 || presence[before] || len(changes) != beforeChanges || len(submits) != beforeSubmits {
		t.Fatal("rejected presence silently replaced storage or published a value")
	}
	redirectStorage = true
	r.Type("0")
	r.Frame()
	if !empty || n != 4 || field().Value != "" || len(presence) != before+2 || presence[before+1] || len(changes) != beforeChanges || len(submits) != beforeSubmits {
		t.Fatal("rejected presence discarded callback storage or repeated a callback")
	}
}
