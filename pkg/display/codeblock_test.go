package display

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestCodeBlockCopiesCurrentSourceWithoutLineNumbers(t *testing.T) {
	source := "package main\n\nfunc main() {}\n"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Title: "main.go", Code: source, Language: "go", LineNumbers: true})
	}, 500, 300)
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if r.Clipboard() != source {
		t.Fatalf("clipboard = %q", r.Clipboard())
	}
	source = "// 新的源代码\n"
	r.Frame()
	if err := r.Click("Copied"); err != nil {
		t.Fatal(err)
	}
	if r.Clipboard() != source {
		t.Fatalf("stale clipboard = %q", r.Clipboard())
	}
}

func codeBlockField(t *testing.T, r *ui.Tester, label string) ui.AccessInfo {
	t.Helper()
	field, ok := r.AccessByRole(ui.RoleTextField, label)
	if !ok {
		t.Fatalf("native editor %q missing", label)
	}
	return field
}

func TestCodeBlockConfirmedCopyFailureRepeatAndAnnouncement(t *testing.T) {
	copies, failures := 0, 0
	source := "first"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Title: "Output", Code: source, Height: 60, OnCopy: func(value string) {
			copies++
			if value != source {
				t.Error("stale copied value")
			}
		}, OnCopyError: func(error) { failures++ }})
	}, 400, 180)
	r.FailClipboard(errors.New("clipboard busy"))
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if copies != 0 || failures != 1 {
		t.Fatalf("unconfirmed copy was reported: copies=%d failures=%d", copies, failures)
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Copied"); ok || len(r.Announcements()) != 0 {
		t.Fatal("failed clipboard write announced success")
	}
	r.FailClipboard(nil)
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if copies != 1 || r.Clipboard() != "first" || len(r.Announcements()) != 1 {
		t.Fatal("successful copy did not confirm once")
	}
	r.Advance(time.Second)
	source = "second"
	r.Frame()
	if err := r.Click("Copied"); err != nil {
		t.Fatal(err)
	}
	if copies != 2 || r.Clipboard() != "second" {
		t.Fatal("repeat copy did not use latest source")
	}
	// React's copied=true effect is not restarted by setting true again.
	r.Advance(1100 * time.Millisecond)
	if _, ok := r.AccessByRole(ui.RoleButton, "Copy code"); !ok {
		t.Fatal("repeat copy incorrectly extended the upstream timeout")
	}
	if r.FramePending() {
		t.Fatal("expired copy feedback kept an idle frame")
	}
}

func TestCodeBlockNativeEditorSurvivesWrapThemeSizeAndStreamAppend(t *testing.T) {
	state := CodeBlockState{}
	source := "first\nsecond"
	wrap, small, dark := false, false, false
	r := ui.NewTester(func(c *ui.Context) {
		tokens := theme.Light()
		if dark {
			tokens = theme.Dark()
		}
		theme.Scope(c, tokens).Children(func() {
			p := CodeBlockProps{Code: source, Title: "Stream", State: &state, Wrap: wrap, Height: 80}
			if small {
				p.Size = CodeBlockSmall
			}
			CodeBlock(c, p)
		})
	}, 400, 200)
	buffer := state.Editor.Buffer
	field := codeBlockField(t, r, "Stream")
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+10)
	state.Editor.Select(0, 5)
	r.Frame()
	wrap, small, dark = true, true, true
	source += "\nthird"
	r.Frame()
	r.Key(ui.Cmd, ui.KeyC)
	if state.Editor.Buffer != buffer || r.Clipboard() != "first" || !r.Focused("Stream") {
		t.Fatalf("reconfiguration replaced native state: clipboard=%q", r.Clipboard())
	}
	r.Type("mutate")
	r.Key(0, ui.KeyBackspace)
	r.Key(ui.Cmd, ui.KeyZ)
	if state.Editor.Buffer.String() != source {
		t.Fatal("read-only native code changed")
	}
}

func TestCodeBlockNativeWrapAndBoundedAutoHeight(t *testing.T) {
	state := CodeBlockState{}
	wrap := false
	source := strings.Repeat("a long native wrapped line ", 20)
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(280).Children(func() {
			CodeBlockWith(c, CodeBlockProps{Code: source, Label: "Wrapped", State: &state, Wrap: wrap, MaxHeight: 100, HideCopy: true}, CodeBlockSlots{})
		})
	}, 400, 240)
	plain := codeBlockField(t, r, "Wrapped").Bounds.H
	wrap = true
	r.Frame()
	r.Frame()
	r.Frame()
	field := codeBlockField(t, r, "Wrapped")
	if plain >= field.Bounds.H || field.Bounds.H != 100 {
		t.Fatalf("native wrapped auto height: plain=%g wrapped=%g", plain, field.Bounds.H)
	}
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+10)
	r.Key(ui.Cmd, ui.KeyEnd)
	if state.Editor.ScrollY() <= 0 {
		t.Fatal("bounded wrap did not scroll native caret into view")
	}
	state.Editor.Select(0, state.Editor.Buffer.Len())
	r.Frame()
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != source {
		t.Fatal("wrapped copy inserted visual newlines")
	}
}

func TestCodeBlockVariantSizePixelsAndEmptyContentGeometry(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, small := range []bool{false, true} {
			t.Run(fmt.Sprintf("dark=%v/small=%v", dark, small), func(t *testing.T) {
				tokens := theme.Light()
				if dark {
					tokens = theme.Dark()
				}
				parent := ui.RGB(50, 90, 150)
				r := ui.NewTester(func(c *ui.Context) {
					theme.Scope(c, tokens).Fill().Background(parent).Children(func() {
						p := CodeBlockProps{Code: "", Label: "Empty code", HideCopy: true}
						if small {
							p.Size = CodeBlockSmall
						}
						ui.Column(c).Width(300).Padding(20).Children(func() { CodeBlockWith(c, p, CodeBlockSlots{}) })
					})
				}, 400, 200)
				field := codeBlockField(t, r, "Empty code")
				wantHeight := float32(14 * 1.625)
				if small {
					wantHeight = 12 * 1.625
				}
				if math.Abs(float64(field.Bounds.H-wantHeight)) >= 1 {
					t.Fatalf("empty native line height=%g want=%g", field.Bounds.H, wantHeight)
				}
				pixel := r.Image().RGBAAt(int(field.Bounds.X+field.Bounds.W-20), int(field.Bounds.Y+5))
				want := color.RGBA{R: parent.R, G: parent.G, B: parent.B, A: 255}
				if dark {
					bg := tokens.Input.Alpha(.3)
					alpha := float32(bg.A) / 255
					want.R = uint8(float32(bg.R)*alpha + float32(parent.R)*(1-alpha) + .5)
					want.G = uint8(float32(bg.G)*alpha + float32(parent.G)*(1-alpha) + .5)
					want.B = uint8(float32(bg.B)*alpha + float32(parent.B)*(1-alpha) + .5)
				}
				if math.Abs(float64(pixel.R)-float64(want.R)) > 2 || math.Abs(float64(pixel.G)-float64(want.G)) > 2 || math.Abs(float64(pixel.B)-float64(want.B)) > 2 {
					t.Fatalf("native editor painted opaque background: got=%v want=%v", pixel, want)
				}
			})
		}
	}
}

func TestCodeBlockLanguageChangePreservesNativeSelection(t *testing.T) {
	language := "text"
	source := "alpha\nbeta\n"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Title: "sample.txt", Code: source, Language: language})
	}, 500, 300)
	field, ok := r.AccessByRole(ui.RoleTextField, "sample.txt")
	if !ok {
		t.Fatal("native editor missing")
	}
	r.ClickAt(field.Bounds.X+field.Bounds.W/2, field.Bounds.Y+20)
	r.Key(ui.Cmd, ui.KeyA)
	if !r.Focused("sample.txt") {
		t.Fatal("native editor did not focus before language change")
	}
	language = "go"
	r.Frame()
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != source || !r.Focused("sample.txt") {
		t.Fatalf("language change lost native selection/focus: copied=%q focus=%v", r.Clipboard(), r.Focused("sample.txt"))
	}
}

func TestCodeBlockCopyFeedbackExpiresWithoutRepeatingClipboardAction(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Code: "copy me", Height: 80})
	}, 500, 200)
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Copied"); !ok {
		t.Fatal("copy action did not expose the source's copied feedback")
	}
	r.SetClipboard("outside clipboard")
	r.Advance(2100 * time.Millisecond)
	if _, ok := r.AccessByRole(ui.RoleButton, "Copy code"); !ok || r.Clipboard() != "outside clipboard" {
		t.Fatal("copy feedback did not expire or replayed clipboard write")
	}
}

func TestCodeBlockRejectsNonFiniteHeight(t *testing.T) {
	for _, height := range []float32{float32(math.Inf(1)), float32(math.NaN())} {
		r := ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			CodeBlock(c, CodeBlockProps{Title: "Finite code", Code: "ok", Height: height})
		}, 500, 300)
		field, ok := r.AccessByRole(ui.RoleTextField, "Finite code")
		if !ok || math.IsInf(float64(field.Bounds.H), 0) || math.IsNaN(float64(field.Bounds.H)) || field.Bounds.H <= 0 {
			t.Fatalf("nonfinite code height escaped to native bounds: %+v", field)
		}
	}
}

func TestCodeBlockDiffLinesAreFullWidthAndNeverEnterNativeCopy(t *testing.T) {
	state := CodeBlockState{}
	source := "same\nremoved\nadded\nhighlighted"
	var tokens *theme.Tokens
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		tokens = theme.TokensOf(c)
		ui.Column(c).Padding(20).Width(380).Children(func() {
			CodeBlockWith(c, CodeBlockProps{Code: source, Label: "Diff", HideCopy: true, LineNumbers: true, StartLine: 995, State: &state,
				Lines: []CodeBlockLine{{Index: 1, Variant: CodeLineRemoved}, {Index: 2, Variant: CodeLineAdded}, {Index: 3, Variant: CodeLineHighlighted}}}, CodeBlockSlots{})
		})
	}, 500, 260)
	field := codeBlockField(t, r, "Diff")
	if !field.ReadOnly || !field.Multiline {
		t.Fatalf("native readonly text semantics missing: %+v", field)
	}
	// The source line background bleeds across padding and its decorative gutter.
	x := 22
	y := int(field.Bounds.Y + float32(state.Editor.LineTop(2)) + 5)
	got := r.Image().RGBAAt(x, y)
	want := color.RGBA{R: tokens.Muted.R, G: tokens.Muted.G, B: tokens.Muted.B, A: 255}
	if got != want {
		t.Fatalf("added row did not cover the full code surface: pixel=%v want=%v", got, want)
	}
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+10)
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != source {
		t.Fatalf("gutter numbers or signs entered native selection: %q", r.Clipboard())
	}
	for _, node := range r.AccessAll() {
		if node.Name == "995" || node.Name == "+" || node.Name == "−" {
			t.Fatal("decorative gutter leaked into accessibility")
		}
	}
}

func TestCodeBlockLanguageAliasesAndUnknownFallbackPaintActualTokens(t *testing.T) {
	state := CodeBlockState{}
	language := "text"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Title: "Language", Code: "func main() { println(42) }", Language: language, State: &state, HideCopy: true, Height: 60})
	}, 440, 150)
	field := codeBlockField(t, r, "Language")
	colored := func() int {
		count := 0
		im := r.Image()
		for y := int(field.Bounds.Y); y < int(field.Bounds.Y+30); y++ {
			for x := int(field.Bounds.X); x < int(field.Bounds.X+field.Bounds.W); x++ {
				p := im.RGBAAt(x, y)
				if int(p.R)-int(p.G) > 30 || int(p.B)-int(p.G) > 30 {
					count++
				}
			}
		}
		return count
	}
	plain := colored()
	buffer := state.Editor.Buffer
	language = "golang"
	r.Frame()
	if colored() <= plain+10 || state.Editor.Buffer != buffer {
		t.Fatal("language alias did not color the original native buffer")
	}
	language = "not-a-supported-language"
	r.Frame()
	if colored() != plain || state.Editor.Buffer.Language() != nil {
		t.Fatal("unknown language did not return to plain native text")
	}
}

func TestCodeBlockLargeFileNativeScrollAndHorizontalReveal(t *testing.T) {
	state := CodeBlockState{}
	source := strings.Repeat("native line\n", 100_000) + strings.Repeat("a", 300) + "_THE_END"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(300).Children(func() {
			CodeBlock(c, CodeBlockProps{Code: source, Title: "Large file", State: &state, Height: 90, HideCopy: true, LineNumbers: true})
		})
	}, 400, 220)
	buffer := state.Editor.Buffer
	state.Editor.GoTo(editor.Pos{Line: 100_000})
	r.Frame()
	if state.Editor.ScrollY() < 1_000_000 || state.Editor.Buffer != buffer {
		t.Fatal("large native file did not preserve buffer and reveal distant lines")
	}
	field := codeBlockField(t, r, "Large file")
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+10)
	state.Editor.GoTo(editor.Pos{Line: 100_000})
	r.Frame()
	before := append([]byte(nil), r.Image().Pix...)
	r.Key(ui.Cmd, ui.KeyEnd)
	if state.Editor.Cursor() != state.Editor.Buffer.Len() {
		t.Fatal("native end key did not reach distant source end")
	}
	different := 0
	after := r.Image()
	for y := int(field.Bounds.Y); y < int(field.Bounds.Y+30); y++ {
		for x := int(field.Bounds.X + 30); x < int(field.Bounds.X+field.Bounds.W-30); x++ {
			i := after.PixOffset(x, y)
			if !bytes.Equal(before[i:i+4], after.Pix[i:i+4]) {
				different++
			}
		}
	}
	if different < 20 {
		t.Fatal("native horizontal reveal did not actually change visible text pixels")
	}
}

func TestCodeBlockEmptyCompoundSlotsPreserveEditorAndFollowingInput(t *testing.T) {
	for _, slot := range []string{"header", "actions", "footer"} {
		t.Run(slot, func(t *testing.T) {
			show := false
			value := "draft"
			state := CodeBlockState{}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Children(func() {
					content := func(child *ui.Context) {
						if show {
							ui.Text(child, "Appeared")
						}
					}
					slots := CodeBlockSlots{}
					switch slot {
					case "header":
						slots.Header = content
					case "actions":
						slots.Actions = content
					case "footer":
						slots.Footer = content
					}
					CodeBlockWith(c, CodeBlockProps{Code: "sample", Label: "Stable code", HideCopy: true, State: &state}, slots)
					ui.TextInput(c, &value).Label("Following")
				})
			}, 500, 260)
			if err := r.Click("Following"); err != nil {
				t.Fatal(err)
			}
			buffer := state.Editor.Buffer
			show = true
			r.Frame()
			if !r.Focused("Following") || state.Editor.Buffer != buffer {
				t.Fatal("empty compound callback changed native identity")
			}
		})
	}
}

func TestCodeBlockRTLHeaderMirrorsButNativeCodeRemainsLTR(t *testing.T) {
	rtl := false
	state := CodeBlockState{}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		root := ui.Column(c).Width(360).Padding(20)
		if rtl {
			root.Direction(ui.RightToLeft)
		}
		root.Children(func() {
			CodeBlock(c, CodeBlockProps{Title: "file.go", Code: "package main", LineNumbers: true, State: &state, Height: 60})
		})
	}, 420, 180)
	left := codeBlockField(t, r, "file.go").Bounds
	copyLTR, ok := r.AccessByRole(ui.RoleButton, "Copy code")
	if !ok {
		t.Fatal("copy button missing")
	}
	rtl = true
	r.Frame()
	right := codeBlockField(t, r, "file.go").Bounds
	copyRTL, _ := r.AccessByRole(ui.RoleButton, "Copy code")
	if left.X != right.X || copyRTL.Bounds.X >= copyLTR.Bounds.X {
		t.Fatalf("RTL code/header geometry: editor=%+v/%+v copy=%+v/%+v", left, right, copyLTR.Bounds, copyRTL.Bounds)
	}
	if state.Editor.Buffer.String() != "package main" {
		t.Fatal("RTL changed source order")
	}
}

func TestCodeBlockCopyUnmountCancelsFeedbackAndRemountDoesNotReplay(t *testing.T) {
	show := true
	calls := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		if show {
			CodeBlock(c.Key("code"), CodeBlockProps{Title: "Unmount", Code: "original", Height: 60, OnCopy: func(string) { calls++ }})
		}
	}, 400, 180)
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	show = false
	r.Frame()
	if r.FramePending() {
		t.Fatal("unmounted copy left its expiry timer alive")
	}
	r.SetClipboard("external")
	r.Advance(3 * time.Second)
	show = true
	r.Frame()
	if calls != 1 || r.Clipboard() != "external" {
		t.Fatal("remounted copy replayed an old external action")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Copy code"); !ok {
		t.Fatal("remounted copy kept stale copied feedback")
	}
}

func TestCodeBlockNumberOverflowAndNativeDisabledAncestry(t *testing.T) {
	value := "first\nsecond"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Disabled(true).Children(func() {
			CodeBlock(c, CodeBlockProps{Code: value, Title: "Guarded", LineNumbers: true, StartLine: math.MaxInt, Height: 60})
		})
	}, 500, 200)
	field := codeBlockField(t, r, "Guarded")
	if !field.Invalid || !field.Disabled || !codeFinite(field.Bounds.W) {
		t.Fatalf("invalid numbering or disabled ancestor lost native state: %+v", field)
	}
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if r.Clipboard() != "" {
		t.Fatal("disabled copy wrote clipboard")
	}
}

func TestCodeBlockFloatingCopySurfaceAndDestructiveFooterPixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprintf("dark=%v", dark), func(t *testing.T) {
			tokens := theme.Light()
			if dark {
				tokens = theme.Dark()
			}
			glyph := ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><path fill="currentColor" d="M0 0h10v10H0z"/></svg>`))
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Background(ui.RGB(50, 90, 150)).Children(func() {
					ui.Column(c).Padding(20).Gap(16).Children(func() {
						CodeBlockWith(c, CodeBlockProps{Code: "floating", Label: "Floating", Height: 60}, CodeBlockSlots{})
						CodeBlockWith(c, CodeBlockProps{Code: "", Variant: CodeBlockDestructive, HideCopy: true}, CodeBlockSlots{Footer: func(child *ui.Context) { ui.Icon(child, glyph).Size(12, 12).Label("Footer glyph") }})
					})
				})
			}, 450, 260, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			block, _ := r.Find("Floating")
			r.Move(block.X+20, block.Y+block.H-20)
			button, ok := r.AccessByRole(ui.RoleButton, "Copy code")
			if !ok {
				t.Fatal("floating copy missing")
			}
			face := r.Image().RGBAAt(int(button.Bounds.X+2), int(button.Bounds.Y+button.Bounds.H/2))
			if face.R != tokens.Background.R || face.G != tokens.Background.G || face.B != tokens.Background.B {
				t.Fatalf("floating action is transparent: %v", face)
			}
			bounds, _ := r.Find("Footer glyph")
			pixel := r.Image().RGBAAt(int(bounds.X+6), int(bounds.Y+6))
			if pixel.R != tokens.MutedForeground.R || pixel.G != tokens.MutedForeground.G || pixel.B != tokens.MutedForeground.B {
				t.Fatalf("destructive footer lost source muted color: %v", pixel)
			}
		})
	}
}

func TestCodeBlockTitleResetsInheritedMonospaceToNativeBodyFont(t *testing.T) {
	text := "MMMMiiii"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Font("monospace").Children(func() {
			CodeBlock(c, CodeBlockProps{Title: text, Label: "Code region", Code: "x", HideCopy: true, Height: 50})
		})
	}, 400, 180)
	reference := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Text(c, text).Font(c.Theme().Font).FontSize(14).FontWeight(500).AlignSelf(ui.Start)
	}, 400, 80)
	actual, ok := r.Find(text)
	want, exists := reference.Find(text)
	if !ok || !exists || math.Abs(float64(actual.W-want.W)) > .1 {
		t.Fatalf("title retained mono ancestor: actual=%+v body=%+v", actual, want)
	}
}

func TestCodeBlockReentrantConfirmedCopyUpdatesControlledCodeImmediately(t *testing.T) {
	state := CodeBlockState{}
	source := "original"
	calls := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Title: "Reentrant", Code: source, State: &state, Height: 60, OnCopy: func(string) { calls++; source = "replacement" }})
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || state.Editor.Buffer.String() != "replacement" {
		t.Fatalf("confirmed callback app update was not rendered immediately: calls=%d native=%q", calls, state.Editor.Buffer.String())
	}
}

func TestCodeBlockUTF8PrefixEditsPreserveNativeSelectedWord(t *testing.T) {
	state := CodeBlockState{}
	source := "界 alpha\nbeta target\nlast"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Code: source, Title: "UTF8", State: &state, Height: 70, HideCopy: true})
	}, 420, 190)
	field := codeBlockField(t, r, "UTF8")
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+10)
	start := strings.Index(source, "beta")
	state.Editor.Select(start, start+4)
	r.Frame()
	source = "新的前缀\n" + source
	r.Frame()
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "beta" {
		t.Fatalf("prefix insertion shifted native selection: %q", r.Clipboard())
	}
	source = "新" + strings.TrimPrefix(source, "新的")
	r.Frame()
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "beta" {
		t.Fatalf("UTF8 prefix replacement split native selection: %q", r.Clipboard())
	}
}

func TestCodeBlockNativeAppearanceShortcutsDoNotOverrideControlledWrapAndNumbers(t *testing.T) {
	for _, binding := range editor.DefaultBindings() {
		if binding.Command != editor.ToggleSoftWrap && binding.Command != editor.ToggleLineNumbers {
			continue
		}
		t.Run(fmt.Sprint(binding.Command), func(t *testing.T) {
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Width(260).Children(func() {
					CodeBlock(c, CodeBlockProps{Title: "Controlled", Code: strings.Repeat("long code ", 20), LineNumbers: true, Height: 75, HideCopy: true})
				})
			}, 320, 170)
			field := codeBlockField(t, r, "Controlled")
			r.ClickAt(field.Bounds.X+20, field.Bounds.Y+10)
			before := append([]byte(nil), r.Image().Pix...)
			r.Key(binding.Mods, binding.Key)
			if binding.ThenKey != 0 {
				r.Key(binding.ThenMods, binding.ThenKey)
			}
			if !bytes.Equal(before, r.Image().Pix) {
				t.Fatal("native editor shortcut overrode the controlled code presentation")
			}
		})
	}
}

func TestCodeBlockFailedCopyReentrantStatusRendersBeforeOwnerContent(t *testing.T) {
	message := "Ready"
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Children(func() {
			ui.Text(c, message)
			CodeBlock(c, CodeBlockProps{Title: "Copy error", Code: "source", Height: 60, OnCopyError: func(error) { message = "Clipboard failed" }})
		})
	}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.FailClipboard(errors.New("busy"))
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if !r.HasText("Clipboard failed") {
		t.Fatalf("failed copy callback only updated data, stale rendered status=%q", message)
	}
}

func TestCodeBlockViewerBindingsPreserveCustomNativeCommandsWithoutGrowth(t *testing.T) {
	state := CodeBlockState{}
	state.Editor.Bindings = []editor.Binding{{Mods: ui.Cmd | ui.Shift, Key: ui.Key8, Command: editor.SelectAll}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Title: "Bindings", Code: "select native text", State: &state, Height: 60, HideCopy: true})
	}, 400, 180)
	count := len(state.Editor.Bindings)
	for range 30 {
		r.Frame()
	}
	if len(state.Editor.Bindings) != count {
		t.Fatal("controlled-view bindings accumulated each frame")
	}
	field := codeBlockField(t, r, "Bindings")
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+10)
	r.Key(ui.Cmd|ui.Shift, ui.Key8)
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "select native text" {
		t.Fatal("viewer guard replaced an unrelated native app binding")
	}
}
