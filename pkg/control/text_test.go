package control

import (
	"image/color"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestInputUsesExactVegaHeightAndRowWidth(t *testing.T) {
	value := ""
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).LineHeight(2).Children(func() {
			ui.Row(c).Width(300).Children(func() { Input(c, &value, InputProps{Label: "Full field"}) })
		})
	}, 800, 300)
	field, ok := r.AccessByRole(ui.RoleTextField, "Full field")
	if !ok || field.Bounds.W != 300 || field.Bounds.H != 36 {
		t.Fatalf("input did not use full width/exact height: %+v", field.Bounds)
	}
}

func TestInputNativeLengthSelectionPasteUndoAndCallbacks(t *testing.T) {
	value := ""
	var changes, submissions []string
	r := tester(func(c *ui.Context) {
		Input(c, &value, InputProps{Label: "Limited", AutoFocus: true, MaxLength: 4,
			OnChange: func(s string) { changes = append(changes, s) }, OnSubmit: func(s string) { submissions = append(submissions, s) }})
	})
	r.Type("a😀b")
	if value != "a😀b" {
		t.Fatalf("native UTF-16 length limit split a scalar: %q", value)
	}
	before := len(changes)
	r.Type("c")
	if value != "a😀b" || len(changes) != before {
		t.Fatal("rejected edit changed value or fired a duplicate callback")
	}
	r.Key(ui.Shift, ui.KeyLeft)
	r.SetClipboard("CD")
	r.Key(ui.Cmd, ui.KeyV)
	if value != "a😀C" {
		t.Fatalf("paste did not honor room left by native selection: %q", value)
	}
	r.Key(ui.Cmd, ui.KeyZ)
	if value != "a😀b" {
		t.Fatalf("native undo did not restore selection replacement: %q", value)
	}
	r.Key(0, ui.KeyEnter)
	if !slices.Equal(submissions, []string{"a😀b"}) {
		t.Fatalf("submit callback = %v", submissions)
	}
	before = len(changes)
	value = "application-owned value longer than four units"
	r.Frame()
	if len(changes) != before || value != "application-owned value longer than four units" {
		t.Fatal("app replacement was truncated or reported as a native edit")
	}
}

func TestInputCompositionDoesNotSubmitUntilNativeCommit(t *testing.T) {
	value, submitted, changed := "", 0, 0
	r := tester(func(c *ui.Context) {
		Input(c, &value, InputProps{Label: "Composing", AutoFocus: true, MaxLength: 4,
			OnChange: func(string) { changed++ }, OnSubmit: func(string) { submitted++ }})
	})
	r.Compose("你好", 2)
	if value != "" || submitted != 0 || changed != 0 {
		t.Fatalf("marked text escaped into committed state or submitted early: value=%q submit=%d changes=%d", value, submitted, changed)
	}
	// The raw native key may choose/commit a candidate, but that gesture must
	// never also submit the form. A second Enter is an ordinary submission.
	r.Key(0, ui.KeyEnter)
	if value != "你好" || changed != 1 || submitted != 0 {
		t.Fatalf("native candidate commit = %q changes=%d submissions=%d", value, changed, submitted)
	}
	r.Key(0, ui.KeyEnter)
	if submitted != 1 {
		t.Fatal("committed native input did not submit")
	}
}

func TestInputReadOnlyPasswordSearchAndPastePolicy(t *testing.T) {
	readOnly, password, search := "Copy me", "fixture-secret", ""
	readOnlyPaste, pasteCalls := 0, 0
	r := tester(func(c *ui.Context) {
		Input(c, &readOnly, InputProps{Label: "Read only", ReadOnly: true, OnPaste: func(string) bool { readOnlyPaste++; return false }})
		Input(c, &password, InputProps{Label: "Password", Password: true})
		Input(c, &search, InputProps{Label: "Search files", Search: true, OnPaste: func(text string) bool { pasteCalls++; return strings.HasPrefix(text, "block") }})
	})
	click(t, r, "Read only")
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "Copy me" {
		t.Fatal("read-only field prevented native selection/copy")
	}
	r.SetClipboard("replacement")
	r.Key(ui.Cmd, ui.KeyV)
	r.Type("x")
	if readOnly != "Copy me" || readOnlyPaste != 0 {
		t.Fatal("read-only field accepted edits or paste callbacks")
	}
	secure, _ := r.AccessByRole(ui.RoleTextField, "Password")
	if !secure.Password || secure.Value != "" {
		t.Fatal("password did not use native protected text semantics")
	}
	click(t, r, "Password")
	r.SetClipboard("unchanged")
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "unchanged" {
		t.Fatal("password copied protected text")
	}
	info, _ := r.AccessByRole(ui.RoleTextField, "Search files")
	if !info.Search {
		t.Fatal("search did not expose the native search-field state")
	}
	click(t, r, "Search files")
	r.SetClipboard("blocked content")
	r.Command("paste")
	if search != "" || pasteCalls != 1 {
		t.Fatal("consumed paste changed native editor contents")
	}
	r.SetClipboard("accepted content")
	r.Command("paste")
	if search != "accepted content" || pasteCalls != 2 {
		t.Fatal("allowed native paste did not insert")
	}
}

func TestInputNativeTextKeyCanConsumeEnterWithoutSubmitting(t *testing.T) {
	value, submitted := "draft", 0
	r := tester(func(c *ui.Context) {
		Input(c, &value, InputProps{Label: "Command editor", AutoFocus: true, OnSubmit: func(string) { submitted++ }, OnTextKey: func(key ui.TextKey) bool {
			if key.Key == ui.KeyEnter && key.Mods == 0 {
				key.SetText(strings.ToUpper(key.Text()))
				return true
			}
			return false
		}})
	})
	r.Key(0, ui.KeyEnter)
	if value != "DRAFT" || submitted != 0 {
		t.Fatalf("pre-editor key handling failed: value=%q submit=%d", value, submitted)
	}
}

func TestInputDisabledAncestorDoesNotAutofocusAndSizesScale(t *testing.T) {
	value, outside := "", ""
	r := tester(func(c *ui.Context) {
		Input(c, &outside, InputProps{Label: "Outside", AutoFocus: true})
		ui.Column(c).Disabled(true).Children(func() { Input(c, &value, InputProps{Label: "Hidden focus", AutoFocus: true}) })
	})
	if !r.Focused("Outside") {
		t.Fatal("disabled descendant stole explicit autofocus")
	}
	for _, size := range []Size{ExtraSmall, Small, DefaultSize, Large} {
		r := ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			Input(c, &value, InputProps{Label: "Sized field", Size: size})
		}, 800, 200)
		x, _ := r.AccessByRole(ui.RoleTextField, "Sized field")
		want := map[Size]float32{ExtraSmall: 24, Small: 32, DefaultSize: 36, Large: 40}[size]
		if x.Bounds.H != want {
			t.Fatalf("size %v height=%g want=%g", size, x.Bounds.H, want)
		}
		r.SetPreferences(ui.Preferences{TextScale: 2})
		x, _ = r.AccessByRole(ui.RoleTextField, "Sized field")
		if err := r.Click("Sized field"); err != nil {
			t.Fatal(err)
		}
		r.Type("Sample")
		caret, active := r.TextCaret()
		if !active || caret.Y < x.Bounds.Y+1 || caret.Y+caret.H > x.Bounds.Y+x.Bounds.H-1 {
			t.Fatalf("native large text caret clipped into fixed control: field=%+v caret=%+v", x.Bounds, caret)
		}
	}
}

func TestTextareaNativeRowsWrapScrollAndCaretRemainCoherent(t *testing.T) {
	value := strings.Repeat("A native line\n", 15) + "Last line"
	rows := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(300).Padding(20).Children(func() {
			Textarea(c, &value, TextareaProps{Label: "Bounded notes", MinRows: 2, MaxRows: 4, Rows: rows, AutoFocus: true})
		})
	}, 800, 500)
	field, _ := r.AccessByRole(ui.RoleTextField, "Bounded notes")
	if !field.Multiline || field.Bounds.H != 102 {
		t.Fatalf("native line bound did not constrain height: %+v", field)
	}
	r.Key(ui.Cmd, ui.KeyEnd)
	caret, ok := r.TextCaret()
	if !ok || caret.Y < field.Bounds.Y || caret.Y+caret.H > field.Bounds.Y+field.Bounds.H {
		t.Fatalf("native caret was not scrolled into the constrained textarea: %+v", caret)
	}
	rows = 6
	r.Frame()
	field, _ = r.AccessByRole(ui.RoleTextField, "Bounded notes")
	if field.Bounds.H != 144 || !r.Focused("Bounded notes") {
		t.Fatal("row-count change replaced editor or lost its content size")
	}
	info, _ := r.AccessByRole(ui.RoleTextField, "Bounded notes")
	if info.Value != value {
		t.Fatal("row-count change mutated controlled content")
	}
}

func TestFieldErrorStylesTextAndCompoundEditorWithoutLeakingScope(t *testing.T) {
	value, other, message := "", "", "Required field"
	var tokens *theme.Tokens
	r := tester(func(c *ui.Context) {
		tokens = theme.TokensOf(c)
		Field(c, FieldProps{Label: "Address", Error: message}, func() ui.Element {
			return InputGroup(c, InputGroupProps{Prefix: "https://"}, func() ui.Element { return Input(c, &value, InputProps{}) }).Label("Address group")
		})
		Input(c, &other, InputProps{Label: "Unrelated"})
	})
	input, _ := r.AccessByRole(ui.RoleTextField, "Address")
	if !input.Invalid || len(input.ErrorMessage) != 1 {
		t.Fatal("field did not expose native error metadata")
	}
	group, _ := r.Find("Address group")
	unrelated, _ := r.Find("Unrelated")
	groupPixel := r.Image().RGBAAt(int(group.X), int(group.Y+group.H/2))
	otherPixel := r.Image().RGBAAt(int(unrelated.X), int(unrelated.Y+unrelated.H/2))
	near := func(a, b uint8) bool { delta := int(a) - int(b); return delta >= -2 && delta <= 2 }
	if !near(groupPixel.R, tokens.Destructive.R) || !near(groupPixel.G, tokens.Destructive.G) || !near(otherPixel.R, tokens.Input.R) || !near(otherPixel.G, tokens.Input.G) {
		t.Fatalf("borders groupR=%d want=%d otherRGB=%d,%d,%d want=%d,%d,%d", groupPixel.R, tokens.Destructive.R, otherPixel.R, otherPixel.G, otherPixel.B, tokens.Input.R, tokens.Input.G, tokens.Input.B)
	}
	message = ""
	r.Frame()
	input, _ = r.AccessByRole(ui.RoleTextField, "Address")
	if input.Invalid || len(input.ErrorMessage) != 0 {
		t.Fatal("cleared field error left stale native state")
	}
}

func TestInputGroupPreservesEditorDisabledInvalidAndCallbacks(t *testing.T) {
	value, changes := "initial", 0
	r := tester(func(c *ui.Context) {
		InputGroup(c, InputGroupProps{Label: "Inner editor"}, func() ui.Element {
			return Input(c, &value, InputProps{Disabled: true, Invalid: true, OnChange: func(string) { changes++ }})
		})
	})
	info, _ := r.AccessByRole(ui.RoleTextField, "Inner editor")
	if !info.Disabled || !info.Invalid {
		t.Fatalf("compound wrapper overrode editor state: disabled=%v invalid=%v", info.Disabled, info.Invalid)
	}
	click(t, r, "Inner editor")
	r.Type("edited")
	if value != "initial" || changes != 0 {
		t.Fatal("disabled compound editor accepted native edits")
	}
}

func TestInputUsesTransparentLightAndSoftDarkSurface(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			value := ""
			r := ui.NewTester(func(c *ui.Context) {
				tokens := theme.Light()
				if dark {
					tokens = theme.Dark()
				}
				theme.Scope(c, tokens).Fill().Children(func() {
					ui.Column(c).Fill().Padding(20).Gap(12).Background(ui.RGB(160, 190, 220)).Children(func() {
						Input(c, &value, InputProps{Label: "Surface"}).Width(200)
						bg := ui.Transparent
						if dark {
							bg = tokens.Input.Alpha(.3)
						}
						ui.Box(c).Size(200, 36).Background(bg).Label("Reference surface")
					})
				})
			}, 800, 250)
			field, _ := r.Find("Surface")
			reference, _ := r.Find("Reference surface")
			got := r.Image().RGBAAt(int(field.X+100), int(field.Y+8))
			want := r.Image().RGBAAt(int(reference.X+100), int(reference.Y+8))
			if got != want || got == (color.RGBA{}) {
				t.Fatalf("input surface got %v want %v", got, want)
			}
		})
	}
}

func TestTextareaDefaultHasVegaMinimumAndGrowsBeyondEightLines(t *testing.T) {
	value := ""
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(400).Padding(20).Children(func() { Textarea(c, &value, TextareaProps{Label: "Growing notes"}) })
	}, 800, 700)
	field, _ := r.AccessByRole(ui.RoleTextField, "Growing notes")
	if field.Bounds.H != 64 {
		t.Fatalf("empty textarea minimum=%g, want 64", field.Bounds.H)
	}
	value = strings.Repeat("A line\n", 11) + "Last line"
	r.Frame()
	field, _ = r.AccessByRole(ui.RoleTextField, "Growing notes")
	if field.Bounds.H <= 200 {
		t.Fatalf("default content-sizing textarea remained capped at eight rows: %+v", field.Bounds)
	}
}
