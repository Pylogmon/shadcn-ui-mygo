package control

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestOTPUsesVegaSlotSize(t *testing.T) {
	value := ""
	r := tester(func(c *ui.Context) { InputOTP(c, &value, InputOTPProps{}) })
	field, ok := r.AccessByRole(ui.RoleTextField, "One-time code")
	if !ok || field.Bounds.H != 36 {
		t.Fatalf("Vega slot height should be 36 DIPs: %+v", field)
	}
}

func TestOTPOuterAlignmentPreservesNativeEditor(t *testing.T) {
	value, align, column := "1234", ui.Center, false
	var handle ui.Handle
	var editorID, handleID uint64
	var anchor, caret int
	var submitted string
	changed, changes, completed := false, 0, 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		parent := ui.Row(c).Size(500, 200).Gap(12).AlignItems(ui.Center).Label("OTP parent")
		if column {
			parent.Column()
		}
		parent.Children(func() {
			ui.Text(c, "Verification\ncode instructions\nwith several lines").Width(180)
			input := InputOTP(c, &value, InputOTPProps{
				AlignSelf: align, Groups: []int{3, 3}, Handle: &handle, Label: "Aligned code",
				OnChange: func(string) { changes++ }, OnComplete: func(string) { completed++ },
				SlotContent: func(c *ui.Context, p InputOTPSlotProps) {
					InputOTPSlot(c, p).Label(fmt.Sprintf("Aligned slot %d", p.Index))
				},
			})
			editorID, handleID = input.ID(), c.Resolve(handle).ID()
			anchor, caret = input.TextAnchorCaret()
			changed = changed || input.Changed()
			input.OnSubmit(func() { submitted = value })
		})
	}, 600, 300)
	checkLayout := func() {
		t.Helper()
		field, ok := r.AccessByRole(ui.RoleTextField, "Aligned code")
		parent, parentOK := r.Find("OTP parent")
		first, firstOK := r.Find("Aligned slot 0")
		last, lastOK := r.Find("Aligned slot 5")
		if !ok || !parentOK || !firstOK || !lastOK {
			t.Fatal("aligned OTP lost its editor or slot layout")
		}
		position, origin, available, size := field.Bounds.Y, parent.Y, parent.H, field.Bounds.H
		if column {
			position, origin, available, size = field.Bounds.X, parent.X, parent.W, field.Bounds.W
		}
		offset := float32(0)
		if align == ui.Center {
			offset = (available - size) / 2
		} else if align == ui.End {
			offset = available - size
		}
		if math.Abs(float64(position-origin-offset)) > .1 || first.X != field.Bounds.X || first.Y != field.Bounds.Y || last.X+last.W != field.Bounds.X+field.Bounds.W {
			t.Fatalf("editor and visual slots did not move together: editor=%+v first=%+v last=%+v parent=%+v", field.Bounds, first, last, parent)
		}
		if editorID == 0 || editorID != handleID {
			t.Fatalf("return value and Handle must identify the native editor: %d != %d", editorID, handleID)
		}
	}
	checkLayout()
	initialID := editorID
	handle.Focus()
	r.Frame()
	second, _ := r.Find("Aligned slot 1")
	r.ClickAt(second.X+second.W/2, second.Y+second.H/2)
	for _, next := range []struct {
		column bool
		align  ui.Align
	}{{false, ui.End}, {false, ui.Start}, {true, ui.Center}, {true, ui.End}} {
		column, align = next.column, next.align
		r.Frame()
		checkLayout()
		if !r.Focused("Aligned code") || editorID != initialID || anchor != 1 || caret != 2 {
			t.Fatalf("alignment remounted the editor or lost focus/selection: id=%d focus=%v selection=%d..%d", editorID, r.Focused("Aligned code"), anchor, caret)
		}
	}
	r.Type("X")
	if value != "1X34" || !changed || changes != 1 || completed != 0 {
		t.Fatalf("aligned editor lost native overwrite/change: value=%q changed=%v changes=%d complete=%d", value, changed, changes, completed)
	}
	r.Key(0, ui.KeyEnd)
	r.Type("56")
	r.Key(0, ui.KeyEnter)
	if value != "1X3456" || changes != 2 || completed != 1 || submitted != value {
		t.Fatalf("aligned editor lost completion or returned-element submit: value=%q changes=%d complete=%d submit=%q", value, changes, completed, submitted)
	}
}

func TestOTPDefaultAcceptsAlphanumericWithoutAnExplicitPattern(t *testing.T) {
	value := ""
	r := tester(func(c *ui.Context) { InputOTP(c, &value, InputOTPProps{Length: 4}) })
	click(t, r, "One-time code")
	r.Type("a1B2")
	if value != "a1B2" {
		t.Fatalf("unrestricted source default discarded the code: %q", value)
	}
}

func TestOTPExplicitFiltersAndLegacyUnicodeLetters(t *testing.T) {
	for _, test := range []struct {
		name  string
		props InputOTPProps
		want  string
	}{
		{"digits", InputOTPProps{DigitsOnly: true}, "12"},
		{"legacy letters", InputOTPProps{Alphanumeric: true}, "a1界b2"},
		{"custom overrides preset", InputOTPProps{DigitsOnly: true, AllowRune: func(ch rune) bool { return ch == '界' || ch >= '0' && ch <= '9' }}, "1界2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := ""
			r := tester(func(c *ui.Context) { InputOTP(c, &value, test.props) })
			r.Key(0, ui.KeyTab)
			r.Type("a1界-b2")
			if value != test.want {
				t.Fatalf("filter=%q want=%q", value, test.want)
			}
		})
	}
}

func TestOTPCompletionOnlyTransitionsFromIncompleteToFull(t *testing.T) {
	value, calls := "12", 0
	r := tester(func(c *ui.Context) {
		InputOTP(c, &value, InputOTPProps{Length: 4, OnComplete: func(string) { calls++ }})
	})
	value = "1234"
	r.Frame()
	if calls != 1 {
		t.Fatalf("incomplete to full notified %d times", calls)
	}
	value = "5678"
	r.Frame()
	if calls != 1 {
		t.Fatalf("replacing a full code fired completion again: %d", calls)
	}
}

func TestOTPGroupGeometrySelectionAndNativeOverwrite(t *testing.T) {
	value := "123456"
	var slots []InputOTPSlotProps
	var a, z int
	r := tester(func(c *ui.Context) {
		slots = nil
		input := InputOTP(c, &value, InputOTPProps{Groups: []int{3, 3}, SlotContent: func(child *ui.Context, p InputOTPSlotProps) {
			slots = append(slots, p)
			InputOTPSlot(child, p).Label(fmt.Sprintf("Slot %d", p.Index))
		}})
		a, z = input.TextSelection()
	})
	field, _ := r.AccessByRole(ui.RoleTextField, "One-time code")
	if field.Bounds.W != 248 {
		t.Fatalf("grouped width=%g", field.Bounds.W)
	}
	third, _ := r.Find("Slot 2")
	fourth, _ := r.Find("Slot 3")
	if fourth.X-third.X != 68 {
		t.Fatalf("group separator spacing: third=%+v fourth=%+v", third, fourth)
	}
	r.ClickAt(third.X+18, third.Y+18)
	if a != 2 || z != 3 || !slots[2].Active || slots[5].Active {
		t.Fatalf("actual selection=%d..%d slots=%+v", a, z, slots)
	}
	r.Type("9")
	if value != "129456" {
		t.Fatalf("slot click did not overwrite native selection: %q", value)
	}
	if a != 3 || z != 4 {
		t.Fatalf("typing did not advance the native range: %d..%d", a, z)
	}
	r.Key(0, ui.KeyLeft)
	if a != 2 || z != 3 {
		t.Fatalf("Left skipped a native slot: %d..%d", a, z)
	}
	r.Key(0, ui.KeyHome)
	r.Key(ui.Shift, ui.KeyRight)
	if a != 0 || z != 2 || !slots[0].Active || !slots[1].Active || slots[2].Active {
		t.Fatalf("native range=%d..%d slots=%+v", a, z, slots)
	}
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "12" {
		t.Fatalf("native copy=%q", r.Clipboard())
	}
	r.Type("AB")
	if value != "AB9456" {
		t.Fatalf("native range replacement=%q", value)
	}
}

func TestOTPPasteTransformFilteringUndoAndRedo(t *testing.T) {
	value := "12"
	var changes []string
	completed := 0
	r := tester(func(c *ui.Context) {
		InputOTP(c, &value, InputOTPProps{Length: 6, DigitsOnly: true, PasteTransform: func(text string) string { return strings.NewReplacer("-", "", " ", "").Replace(text) }, OnChange: func(v string) { changes = append(changes, v) }, OnComplete: func(string) { completed++ }})
	})
	click(t, r, "One-time code")
	r.SetClipboard("3-4 56")
	r.Command("paste")
	if value != "123456" || completed != 1 || !slices.Equal(changes, []string{"123456"}) {
		t.Fatalf("paste normalization=%q complete=%d changes=%v", value, completed, changes)
	}
	r.Command("undo")
	if value != "12" {
		t.Fatalf("normalization destroyed native paste undo: %q", value)
	}
	r.Command("redo")
	if value != "123456" || completed != 2 {
		t.Fatalf("native redo=%q complete=%d", value, completed)
	}
}

func TestOTPCustomPatternRejectsWholeCandidateAndPreservesSelection(t *testing.T) {
	value := "AB"
	changes := 0
	var a, z int
	r := tester(func(c *ui.Context) {
		input := InputOTP(c, &value, InputOTPProps{Length: 4, Pattern: regexp.MustCompile(`^[A-Z0-9]*$`), OnChange: func(string) { changes++ }})
		a, z = input.TextSelection()
	})
	r.Key(0, ui.KeyTab)
	r.Type("!")
	if value != "AB" || changes != 0 || a != 2 || z != 2 {
		t.Fatalf("pattern rejection changed code/selection: %q count=%d selection=%d..%d", value, changes, a, z)
	}
	r.Type("12")
	if value != "AB12" || changes != 1 {
		t.Fatalf("valid pattern change=%q count=%d", value, changes)
	}
}

func TestOTPCompositionStaysNativeAndOnlyCompletesOnCommit(t *testing.T) {
	value := "AB"
	changes, completed := 0, 0
	var composing bool
	r := tester(func(c *ui.Context) {
		input := InputOTP(c, &value, InputOTPProps{Length: 4, OnChange: func(string) { changes++ }, OnComplete: func(string) { completed++ }})
		composing = input.Composing()
	})
	r.Key(0, ui.KeyTab)
	r.Compose("你好", 2)
	if value != "AB" || !composing || changes != 0 || completed != 0 {
		t.Fatalf("composition was published as committed code: %q compose=%v", value, composing)
	}
	if caret, ok := r.TextCaret(); !ok || caret.W <= 0 {
		t.Fatalf("hidden native editor lost input-method geometry: %+v, %v", caret, ok)
	}
	r.Type("你好")
	if value != "AB你好" || composing || changes != 1 || completed != 1 {
		t.Fatalf("composition commit=%q compose=%v change=%d complete=%d", value, composing, changes, completed)
	}
}

func TestOTPReadonlyDisabledAndSingleAccessibleEditor(t *testing.T) {
	value := "1234"
	p := InputOTPProps{Length: 4, ReadOnly: true}
	r := tester(func(c *ui.Context) { InputOTP(c, &value, p); ui.Button(c, "Next action") })
	click(t, r, "One-time code")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("5678")
	if value != "1234" {
		t.Fatal("readonly code changed")
	}
	field, _ := r.AccessByRole(ui.RoleTextField, "One-time code")
	if !field.ReadOnly {
		t.Fatal("readonly property missing from native editor")
	}
	editors := 0
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleTextField {
			editors++
		}
	}
	if editors != 1 {
		t.Fatalf("OTP exposed %d native editors", editors)
	}
	p.ReadOnly = false
	p.Disabled = true
	r.Frame()
	r.Key(0, ui.KeyTab)
	if !r.Focused("Next action") {
		t.Fatal("disabled OTP retained a keyboard stop")
	}
	field, _ = r.AccessByRole(ui.RoleTextField, "One-time code")
	if !field.Disabled {
		t.Fatal("disabled state not exposed")
	}
}

func TestOTPGroupChangeKeepsEditorFocusAndNativeSelection(t *testing.T) {
	value := "123456"
	groups := []int{3, 3}
	var a, z int
	r := tester(func(c *ui.Context) {
		input := InputOTP(c, &value, InputOTPProps{Groups: groups})
		a, z = input.TextSelection()
	})
	click(t, r, "One-time code")
	r.Key(0, ui.KeyHome)
	groups = []int{2, 2, 2}
	r.Frame()
	if !r.Focused("One-time code") || a != 0 || z != 1 {
		t.Fatalf("group change remounted native editor: focused=%v selection=%d..%d", r.Focused("One-time code"), a, z)
	}
	r.Type("A")
	if value != "A23456" {
		t.Fatalf("group change lost native code: %q", value)
	}
}

func TestOTPReducedCaretAndInvalidGroupsDoNotIdle(t *testing.T) {
	value := ""
	r := ui.NewTesterWith(func(c *ui.Context) {
		InputOTP(c, &value, InputOTPProps{Length: 4, Groups: []int{2, 0, 2}, AutoFocus: true})
	}, 400, 120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	field, _ := r.AccessByRole(ui.RoleTextField, "One-time code")
	if !field.Invalid || field.Bounds.W != 144 || !r.Focused("One-time code") {
		t.Fatalf("invalid groups did not expose usable fallback: %+v", field)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("reduced-motion OTP kept requesting blink frames")
	}
}

func TestOTPEmptyVisualCallbacksDoNotLeakIdentity(t *testing.T) {
	value, text := "", "draft"
	show := false
	r := tester(func(c *ui.Context) {
		InputOTP(c, &value, InputOTPProps{Groups: []int{2, 2}, Separator: func(child *ui.Context, _ int) {
			if show {
				ui.Text(child, "–")
			}
		}, SlotContent: func(child *ui.Context, p InputOTPSlotProps) {
			if show {
				ui.Text(child, p.Char)
			}
		}})
		Input(c, &text, InputProps{Label: "Following editor"})
	})
	click(t, r, "Following editor")
	show = true
	r.Frame()
	if !r.Focused("Following editor") {
		t.Fatal("empty OTP callback leaked identity into following editor")
	}
}

func TestOTPCallbacksCanFormatRejectAndResetCallerState(t *testing.T) {
	t.Run("format", func(t *testing.T) {
		value, completed := "", ""
		r := tester(func(c *ui.Context) {
			InputOTP(c, &value, InputOTPProps{Length: 4, OnChange: func(code string) { value = strings.ToUpper(code) }, OnComplete: func(code string) { completed = code }})
		})
		r.Key(0, ui.KeyTab)
		r.Type("abcd")
		field, _ := r.AccessByRole(ui.RoleTextField, "One-time code")
		if value != "ABCD" || completed != "ABCD" || field.Value != "ABCD" {
			t.Fatalf("caller formatting was overwritten: value=%q complete=%q field=%+v", value, completed, field)
		}
	})
	t.Run("reject", func(t *testing.T) {
		value, completed := "", 0
		r := tester(func(c *ui.Context) {
			InputOTP(c, &value, InputOTPProps{Length: 4, OnChange: func(string) { value = "" }, OnComplete: func(string) { completed++ }})
		})
		r.Key(0, ui.KeyTab)
		r.Type("abcd")
		if value != "" || completed != 0 {
			t.Fatalf("caller rejection was overwritten: value=%q complete=%d", value, completed)
		}
	})
	t.Run("reset after completion", func(t *testing.T) {
		value := ""
		var completed []string
		r := tester(func(c *ui.Context) {
			InputOTP(c, &value, InputOTPProps{Length: 4, OnComplete: func(code string) { completed = append(completed, code); value = "" }})
		})
		r.Key(0, ui.KeyTab)
		r.Type("1234")
		r.Type("5678")
		field, _ := r.AccessByRole(ui.RoleTextField, "One-time code")
		if value != "" || field.Value != "" || !slices.Equal(completed, []string{"1234", "5678"}) {
			t.Fatalf("completion reset resurrected a code: value=%q field=%q codes=%v", value, field.Value, completed)
		}
	})
}

func TestOTPWindowBlurHidesActiveCaretAndStopsBlink(t *testing.T) {
	value := ""
	active := false
	r := tester(func(c *ui.Context) {
		InputOTP(c, &value, InputOTPProps{Length: 4, SlotContent: func(child *ui.Context, p InputOTPSlotProps) {
			if p.Index == 0 {
				active = p.Active
			}
			InputOTPSlot(child, p)
		}})
	})
	r.Key(0, ui.KeyTab)
	if !active {
		t.Fatal("focused OTP did not activate its append slot")
	}
	r.SetFocused(false)
	r.Advance(time.Second)
	if active || r.FramePending() {
		t.Fatal("blurred OTP kept a fake caret or timer alive")
	}
}

func TestOTPRTLPhysicalArrowsAndCapturedRangeAcrossGroups(t *testing.T) {
	value := "123456"
	var a, z int
	r := tester(func(c *ui.Context) {
		input := InputOTP(c, &value, InputOTPProps{Groups: []int{3, 3}, SlotContent: func(child *ui.Context, p InputOTPSlotProps) {
			InputOTPSlot(child, p).Label(fmt.Sprintf("RTL slot %d", p.Index))
		}})
		a, z = input.TextSelection()
	})
	r.SetDirection(ui.RightToLeft)
	first, _ := r.Find("RTL slot 0")
	last, _ := r.Find("RTL slot 5")
	if first.X <= last.X {
		t.Fatalf("RTL did not mirror visual slots: first=%+v last=%+v", first, last)
	}
	r.ClickAt(first.X+18, first.Y+18)
	r.Key(0, ui.KeyLeft)
	if a != 1 || z != 2 {
		t.Fatalf("RTL Left did not move to physical left slot: %d..%d", a, z)
	}
	r.Key(0, ui.KeyRight)
	if a != 0 || z != 1 {
		t.Fatalf("RTL Right did not return to first logical slot: %d..%d", a, z)
	}
	second, _ := r.Find("RTL slot 1")
	fifth, _ := r.Find("RTL slot 4")
	r.Press(second.X+18, second.Y+18)
	r.Move(fifth.X+18, fifth.Y+18)
	r.Release(fifth.X+18, fifth.Y+18)
	if a != 1 || z != 5 {
		t.Fatalf("captured group drag did not select logical range: %d..%d", a, z)
	}
	r.Type("AB")
	if value != "1AB6" {
		t.Fatalf("RTL captured range did not use native replacement: %q", value)
	}
}

func TestOTPCustomSlotGeometryKeepsNativeHitTestingAndFocus(t *testing.T) {
	value := "1234"
	size, gap := float32(48), float32(12)
	var a, z int
	r := tester(func(c *ui.Context) {
		input := InputOTP(c, &value, InputOTPProps{Groups: []int{2, 2}, SlotWidth: size, SlotHeight: size, Gap: &gap,
			SlotContent: func(child *ui.Context, p InputOTPSlotProps) {
				InputOTPSlot(child, p).Label(fmt.Sprintf("Large slot %d", p.Index))
			}})
		a, z = input.TextSelection()
	})
	field, _ := r.AccessByRole(ui.RoleTextField, "One-time code")
	second, _ := r.Find("Large slot 1")
	third, _ := r.Find("Large slot 2")
	if field.Bounds.W != 232 || field.Bounds.H != 48 || second.W != 48 || third.X-second.X != 88 {
		t.Fatalf("custom geometry field=%+v second=%+v third=%+v", field.Bounds, second, third)
	}
	r.ClickAt(second.X+24, second.Y+24)
	if a != 1 || z != 2 {
		t.Fatalf("custom-width native hit testing selected %d..%d", a, z)
	}
	size = 32
	r.Frame()
	if !r.Focused("One-time code") || a != 1 || z != 2 {
		t.Fatalf("resizing remounted editor or lost selection: %d..%d", a, z)
	}
	r.Type("A")
	if value != "1A34" {
		t.Fatalf("custom-width overwrite=%q", value)
	}
}

func TestOTPSelectedCompositionBackspaceAndTransformedPaste(t *testing.T) {
	value, changes := "1234", 0
	var a, z int
	r := tester(func(c *ui.Context) {
		input := InputOTP(c, &value, InputOTPProps{Length: 4, PasteTransform: strings.ToUpper, OnChange: func(string) { changes++ }})
		a, z = input.TextSelection()
	})
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyHome)
	r.Compose("你", 1)
	if value != "1234" || changes != 0 {
		t.Fatalf("preedit removed caller's selected digit: value=%q changes=%d", value, changes)
	}
	r.Type("你")
	if value != "你234" || changes != 1 || a != 1 || z != 2 {
		t.Fatalf("native selected composition commit=%q changes=%d range=%d..%d", value, changes, a, z)
	}
	r.Key(0, ui.KeyBackspace)
	if value != "你34" || a != 0 || z != 1 {
		t.Fatalf("native Backspace removal/previous range=%q %d..%d", value, a, z)
	}
	r.Key(0, ui.KeyHome)
	r.Key(ui.Shift, ui.KeyRight)
	r.SetClipboard("ab")
	r.Command("paste")
	if value != "AB4" || a != 3 || z != 3 {
		t.Fatalf("transformed paste did not replace native range: value=%q range=%d..%d", value, a, z)
	}
}

func TestOTPVisibleCaretFocusAndInvalidPixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprintf("dark=%v", dark), func(t *testing.T) {
			tokens := theme.Light()
			if dark {
				tokens = theme.Dark()
			}
			value, invalid := "", false
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
					ui.Column(c).Padding(20).Children(func() { InputOTP(c, &value, InputOTPProps{Length: 4, Invalid: invalid}) })
				})
			}, 220, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			field, _ := r.AccessByRole(ui.RoleTextField, "One-time code")
			before := r.Image().RGBAAt(int(field.Bounds.X+18), int(field.Bounds.Y+18))
			r.Key(0, ui.KeyTab)
			caret := r.Image().RGBAAt(int(field.Bounds.X+18), int(field.Bounds.Y+18))
			if caret == before {
				t.Fatalf("focused empty slot lacks visible native-selection caret: before=%v caret=%v foreground=%v", before, caret, tokens.Foreground)
			}
			focusRing := r.Image().RGBAAt(int(field.Bounds.X-2), int(field.Bounds.Y+18))
			invalid = true
			r.Frame()
			invalidRing := r.Image().RGBAAt(int(field.Bounds.X-2), int(field.Bounds.Y+18))
			border := r.Image().RGBAAt(int(field.Bounds.X), int(field.Bounds.Y+18))
			if invalidRing == focusRing || border.R != tokens.Destructive.R || border.G != tokens.Destructive.G || border.B != tokens.Destructive.B {
				t.Fatalf("invalid pixels did not replace focus treatment: focus=%v invalid=%v border=%v destructive=%v", focusRing, invalidRing, border, tokens.Destructive)
			}
			r.SetFocused(false)
			if pixel := r.Image().RGBAAt(int(field.Bounds.X+18), int(field.Bounds.Y+18)); pixel == caret {
				t.Fatalf("blur left the painted caret visible: %v", pixel)
			}
		})
	}
}

func TestOTPUnmountDropsNativeCompositionAndCallbacks(t *testing.T) {
	value, visible := "AB", true
	changes, completed := 0, 0
	composing := false
	r := tester(func(c *ui.Context) {
		if visible {
			input := InputOTP(c, &value, InputOTPProps{Length: 4, AutoFocus: true, OnChange: func(string) { changes++ }, OnComplete: func(string) { completed++ }})
			composing = input.Composing()
		}
	})
	r.Compose("你好", 2)
	if !composing || value != "AB" {
		t.Fatalf("preedit=%q composing=%v", value, composing)
	}
	visible = false
	r.Frame()
	r.Type("界")
	if active, _ := r.TextInputActive(); active {
		t.Fatal("unmounted native editor retained text services")
	}
	if value != "AB" || changes != 0 || completed != 0 {
		t.Fatalf("unmounted composition dispatched: %q changes=%d complete=%d", value, changes, completed)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("unmounted editor retained caret timers")
	}
	visible = true
	r.Frame()
	if composing {
		t.Fatal("remounted editor inherited old marked text")
	}
	r.Type("12")
	if value != "AB12" || changes != 1 || completed != 1 {
		t.Fatalf("remount=%q changes=%d complete=%d", value, changes, completed)
	}
	visible = false
	r.Frame()
	visible = true
	r.Frame()
	if changes != 1 || completed != 1 {
		t.Fatal("initial full code fired on remount")
	}
}

func TestOTPUnrestrictedInputAndCallbackOrder(t *testing.T) {
	value := ""
	var events []string
	r := tester(func(c *ui.Context) {
		InputOTP(c, &value, InputOTPProps{Length: 4, OnChange: func(code string) {
			events = append(events, fmt.Sprintf("change:%s:%s", code, value))
			value = strings.ToUpper(code)
		}, OnComplete: func(code string) { events = append(events, fmt.Sprintf("complete:%s:%s", code, value)) }})
	})
	r.Key(0, ui.KeyTab)
	r.Type("a!界?")
	if value != "A!界?" || !slices.Equal(events, []string{"change:a!界?:a!界?", "complete:A!界?:A!界?"}) {
		t.Fatalf("value=%q callbacks=%v", value, events)
	}
	events = nil
	value = "AB"
	r.Frame()
	value = "C#你！"
	r.Frame()
	if !slices.Equal(events, []string{"complete:C#你！:C#你！"}) {
		t.Fatalf("external update callbacks=%v", events)
	}
}

func TestOTPCompositionCancelRestoresSelectedText(t *testing.T) {
	value, changes := "1234", 0
	var a, z int
	r := tester(func(c *ui.Context) {
		input := InputOTP(c, &value, InputOTPProps{Length: 4, OnChange: func(string) { changes++ }})
		a, z = input.TextSelection()
	})
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyHome)
	r.Compose("你", 1)
	r.Compose("", 0)
	if value != "1234" || changes != 0 || a != 0 || z != 1 {
		t.Fatalf("cancel=%q change=%d selection=%d..%d", value, changes, a, z)
	}
	r.Type("A")
	if value != "A234" || changes != 1 {
		t.Fatalf("cancel left invalid native selection: %q changes=%d", value, changes)
	}
}

func TestOTPClearedPreeditKeepsRangeForFollowingCommitAndUndo(t *testing.T) {
	for _, index := range []int{0, 3} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			value, changes := "1234", 0
			var a, z int
			r := tester(func(c *ui.Context) {
				input := InputOTP(c, &value, InputOTPProps{Length: 4, OnChange: func(string) { changes++ }, SlotContent: func(child *ui.Context, p InputOTPSlotProps) {
					InputOTPSlot(child, p).Label(fmt.Sprintf("Composition slot %d", p.Index))
				}})
				a, z = input.TextSelection()
			})
			bounds, _ := r.Find(fmt.Sprintf("Composition slot %d", index))
			r.ClickAt(bounds.X+18, bounds.Y+18)
			r.Compose("你", 1)
			r.Compose("", 0)
			if value != "1234" || changes != 0 || a != index || z != index+1 {
				t.Fatalf("cleared mark published deletion/lost range: %q changes=%d range=%d..%d", value, changes, a, z)
			}
			r.Type("你")
			want := []rune("1234")
			want[index] = '你'
			if value != string(want) || changes != 1 {
				t.Fatalf("commit=%q want=%q changes=%d", value, string(want), changes)
			}
			r.Command("undo")
			if value != "1234" {
				t.Fatalf("native composition undo=%q", value)
			}
		})
	}
}

func TestOTPRejectsOverflowingAndUnboundedSlotGroups(t *testing.T) {
	for _, props := range []InputOTPProps{
		{Length: math.MaxInt},
		{Groups: []int{math.MaxInt, math.MaxInt, 3}},
		{Groups: []int{2, -1, 2}},
	} {
		n, groups, valid := otpGroups(props)
		if valid || n <= 0 || n > 4096 || len(groups) != 1 || groups[0] != n {
			t.Fatalf("unsafe slot budget accepted: props=%+v n=%d groups=%v valid=%v", props, n, groups, valid)
		}
	}
}

func TestOTPDerivedGeometryStaysFiniteAndInvalid(t *testing.T) {
	maxGap := float32(math.MaxFloat32)
	for _, props := range []InputOTPProps{
		{Groups: []int{3, 3}, Gap: &maxGap},
		{Length: 4, SlotWidth: math.MaxFloat32},
	} {
		value := ""
		r := tester(func(c *ui.Context) { InputOTP(c, &value, props) })
		field, ok := r.AccessByRole(ui.RoleTextField, "One-time code")
		if !ok || !field.Invalid || math.IsInf(float64(field.Bounds.W), 0) || math.IsNaN(float64(field.Bounds.W)) || field.Bounds.W > 4096*36+4095*32 {
			t.Fatalf("unsafe derived native bounds: %+v", field)
		}
	}
}

func TestOTPPasteTransformNotifiesWhenNativeTextWasUnchanged(t *testing.T) {
	for _, raw := range []string{"a", ""} {
		t.Run(fmt.Sprintf("raw=%q", raw), func(t *testing.T) {
			value, changes, changed := "abcd", 0, false
			r := tester(func(c *ui.Context) {
				input := InputOTP(c, &value, InputOTPProps{Length: 4, PasteTransform: func(text string) string {
					if text == "" {
						return "X"
					}
					return strings.ToUpper(text)
				}, OnChange: func(string) { changes++ }})
				changed = changed || input.Changed()
			})
			r.Key(0, ui.KeyTab)
			r.Key(0, ui.KeyHome)
			r.SetClipboard(raw)
			r.Command("paste")
			want := "Abcd"
			if raw == "" {
				want = "Xbcd"
			}
			if value != want || changes != 1 || !changed {
				t.Fatalf("transform=%q want=%q changes=%d changed=%v", value, want, changes, changed)
			}
			field, ok := r.AccessByRole(ui.RoleTextField, "One-time code")
			if !ok || field.Value != want {
				t.Fatalf("transform did not reach native editor: %+v", field)
			}
			r.Advance(time.Second)
			if r.FramePending() || changes != 1 {
				t.Fatal("transformed paste retained frames or repeated its callback")
			}
		})
	}
}

func TestOTPTransformedMiddlePasteSelectsEndAsUpstream142(t *testing.T) {
	value, changes := "ABCD", 0
	var a, z int
	r := tester(func(c *ui.Context) {
		input := InputOTP(c, &value, InputOTPProps{Length: 4, PasteTransform: strings.ToUpper, OnChange: func(string) { changes++ }, SlotContent: func(child *ui.Context, p InputOTPSlotProps) {
			InputOTPSlot(child, p).Label(fmt.Sprintf("Paste slot %d", p.Index))
		}})
		a, z = input.TextSelection()
	})
	second, _ := r.Find("Paste slot 1")
	r.ClickAt(second.X+18, second.Y+18)
	r.SetClipboard("x")
	r.Command("paste")
	// The pinned source input.tsx:341–346 selects min(length,maxLength-1)
	// through length after transformed paste, regardless of replaced slot.
	if value != "AXCD" || changes != 1 || a != 3 || z != 4 {
		t.Fatalf("upstream transformed paste contract=%q changes=%d range=%d..%d", value, changes, a, z)
	}
	r.Command("undo")
	if value != "ABCD" {
		t.Fatalf("native transformed-paste undo=%q", value)
	}
}

func TestOTPFieldInvalidStylesSlotsWithoutLeakingToFollowingEditor(t *testing.T) {
	value, outside := "", ""
	r := tester(func(c *ui.Context) {
		Field(c, FieldProps{Label: "Required code", Error: "Enter the code"}, func() ui.Element {
			return InputOTP(c, &value, InputOTPProps{Length: 4})
		})
		InputOTP(c, &outside, InputOTPProps{Length: 4, Label: "Outside code"})
	})
	field, ok := r.AccessByRole(ui.RoleTextField, "Required code")
	clean, otherOK := r.AccessByRole(ui.RoleTextField, "Outside code")
	if !ok || !otherOK || !field.Invalid || clean.Invalid {
		t.Fatalf("field invalid scope=%+v outside=%+v", field, clean)
	}
	pixel := r.Image().RGBAAt(int(field.Bounds.X+18), int(field.Bounds.Y-2))
	if pixel.R <= pixel.G+15 {
		t.Fatalf("Field error did not style OTP slots: %v", pixel)
	}
}
