package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func otpCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_OTP_COMPOSITION_SNAPSHOT_DIR")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, r.Image())
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("OTP image %v/%v", err, closeErr)
	}
}

// One native-input workflow covers source slots and application drafts. Its
// software images and injected selectors do not establish OS/AT/IME acceptance.
func TestOTPCompositionsSourceWorkflow(t *testing.T) {
	p := &otpCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Wrap().Children(func() {
				if ui.Button(c, "Leave OTP examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show OTP examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, h, native)
			}
		})
	}, 620, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, name)
		if !ok {
			t.Fatalf("missing actual native %v %q", role, name)
		}
		return a
	}
	click := func(role ui.Role, name string) {
		t.Helper()
		a := access(role, name)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	choose := func(key string) {
		t.Helper()
		index := -1
		for i, option := range otpCompositionRecipes {
			if option.Value == key {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatalf("unknown OTP source recipe %q", key)
		}
		if err := r.Click("Input OTP configuration"); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(otpCompositionRecipes[index].Label); err != nil {
				t.Fatal(err)
			}
		} else {
			s := b.surfaces[len(b.surfaces)-1]
			s.tester.Key(0, ui.KeyHome)
			for range index {
				s.tester.Key(0, ui.KeyDown)
			}
			s.tester.Key(0, ui.KeyEnter)
			if !s.didClose {
				t.Fatal("styled OTP selector failed to complete its actual action")
			}
		}
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("OTP source selection %q reached %q error=%s", key, p.recipe, p.popupError)
		}
	}
	replace := func(name, text string) {
		t.Helper()
		click(ui.RoleTextField, name)
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(text)
		r.Frame()
	}
	if p.recipe != "demo" || len(otpCompositionRecipes) != 18 || *p.drafts["demo"] != "123456" || access(ui.RoleTextField, "One-time code").Bounds.W != 216 {
		t.Fatal("docs populated default or single six-slot geometry was lost")
	}
	count := 0
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleTextField {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("OTP slots introduced %d native editors", count)
	}
	click(ui.RoleTextField, "One-time code")
	r.Key(ui.Cmd, ui.KeyA)
	r.Frame()
	if p.selections["demo"] != [2]int{0, 6} {
		t.Fatal("the actual native code failed to select all six slots")
	}
	r.Type("aB9")
	r.SetClipboard("345")
	r.Key(ui.Cmd, ui.KeyV)
	if *p.drafts["demo"] != "aB9345" || p.changes["demo"] == 0 {
		t.Fatal("unrestricted source default did not retain native replacement/paste")
	}
	r.Key(ui.Cmd, ui.KeyZ)
	if *p.drafts["demo"] != "aB9" {
		t.Fatal("native paste undo did not restore the preceding source draft")
	}
	for _, recipe := range otpCompositionRecipes {
		choose(recipe.Value)
	}
	choose("pattern")
	click(ui.RoleTextField, "Digits Only")
	r.Type("a")
	if *p.draft() != "" {
		t.Fatal("digits pattern accepted a non-digit")
	}
	r.Type("123456")
	if *p.draft() != "123456" {
		t.Fatal("the original anchored digit pattern rejected valid native input")
	}
	choose("alphanumeric")
	replace("One-time code", "aB3")
	r.Type("界")
	if *p.draft() != "aB3" {
		t.Fatal("the source ASCII alphanumeric pattern was widened to Unicode letters")
	}
	r.Type("9")
	if *p.draft() != "aB39" {
		t.Fatal("the original ASCII alphanumeric pattern lost accepted native text")
	}
	choose("four")
	replace("One-time code", "56789")
	if *p.draft() != "5678" || access(ui.RoleTextField, "One-time code").Bounds.W != 144 {
		t.Fatal("source four-digit limit or single group changed")
	}
	choose("disabled")
	if !access(ui.RoleTextField, "One-time code").Disabled {
		t.Fatal("source disabled state did not reach the native editor")
	}
	replace("One-time code", "999999")
	if *p.draft() != "123456" || p.changes["disabled"] != 0 {
		t.Fatal("source disabled native editor accepted text")
	}
	choose("controlled")
	if _, ok := r.Find("Enter your one-time password."); !ok {
		t.Fatal("source controlled empty-state text is missing")
	}
	replace("One-time code", "9z2")
	if _, ok := r.Find("You entered: 9z2"); !ok {
		t.Fatal("source controlled readout failed to track the actual editor")
	}
	r.Key(0, ui.KeyEnter)
	if len(p.actions) != 0 || len(p.submitted) != 0 {
		t.Fatal("controlled source invented an Enter verification action")
	}
	otpCompositionSnapshot(t, r, "otp-controlled-light")
	choose("create-invalid")
	invalid := access(ui.RoleTextField, "Invalid State")
	if !invalid.Invalid || !strings.Contains(invalid.Description, "Example showing the invalid error state.") || !strings.Contains(strings.Join(invalid.ErrorMessage, " "), "Invalid code. Please try again.") || *p.draft() != "000000" || invalid.Bounds.W != 280 {
		t.Fatalf("Create invalid defaults/three pairs/Field associations lost: %+v", invalid)
	}
	replace("Invalid State", "123456")
	if *p.draft() != "123456" {
		t.Fatal("invalid source incorrectly disabled its native editor")
	}
	dark = true
	r.Frame()
	otpCompositionSnapshot(t, r, "otp-invalid-dark")
	choose("form")
	form := access(ui.RoleTextField, "Verification code")
	if form.Bounds.W != 312 || form.Bounds.H != 48 || !strings.Contains(form.Description, "Required") {
		t.Fatalf("docs form explicit44x48 slots/mx2 separator/required changed: %+v", form)
	}
	replace("Verification code", "AB1234")
	r.Key(0, ui.KeyEnter)
	if len(p.actions) != 0 {
		t.Fatal("source form invented an Enter handler")
	}
	click(ui.RoleButton, "Verify")
	if p.submitted["form"] != "AB1234" || p.actions["Verify"] != 1 || *p.draft() != "AB1234" {
		t.Fatal("actual Verify action lost the local source draft snapshot")
	}
	click(ui.RoleButton, "Resend Code")
	if p.actions["Resend Code"] != 1 || *p.draft() != "AB1234" {
		t.Fatal("actual Resend action repeated or reset the source editor")
	}
	click(ui.RoleLink, "Contact support")
	r.Key(0, ui.KeySpace)
	if p.actions["Contact support"] != 1 || len(r.OpenedURLs()) != 0 {
		t.Fatal("source hash link lost native link semantics or opened an OS browser")
	}
	choose("create-form")
	form = access(ui.RoleTextField, "Verification code")
	if form.Bounds.W != 320 || form.Bounds.H != 64 || *p.draft() != "" {
		t.Fatal("Create Vega form did not keep48x64 slots and separate empty draft")
	}
	replace("Verification code", "765432")
	otpCompositionSnapshot(t, r, "otp-create-form-dark")
	r.SetSize(360, 900)
	viewport := access(ui.RoleRegion, "Verification code slots")
	beforeScroll := access(ui.RoleTextField, "Verification code").Bounds.X
	r.Scroll(viewport.Bounds.X+viewport.Bounds.W/2, viewport.Bounds.Y+viewport.Bounds.H/2, 400, 0)
	r.Frame()
	form = access(ui.RoleTextField, "Verification code")
	if form.Bounds.X >= beforeScroll || form.Bounds.X+form.Bounds.W > viewport.Bounds.X+viewport.Bounds.W+1 {
		t.Fatal("the narrow form did not expose the actual last native slot through horizontal scrolling")
	}
	r.ClickAt(form.Bounds.X+form.Bounds.W-24, form.Bounds.Y+form.Bounds.H/2)
	r.Type("9")
	if *p.draft() != "765439" {
		t.Fatal("the revealed final form slot was not reachable for actual native replacement")
	}
	r.SetSize(620, 900)
	choose("rtl")
	if err := r.Click("Input OTP language"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	r.SetSize(360, 900)
	rtl := access(ui.RoleTextField, "קוד אימות")
	if rtl.Bounds.X < 20 || rtl.Bounds.X+rtl.Bounds.W > 340 || *p.draft() != "123456" {
		t.Fatal("translated RTL native editor overflowed its narrow parent or lost the source default")
	}
	replace("קוד אימות", "654321")
	otpCompositionSnapshot(t, r, "otp-hebrew-narrow-dark")
	if err := r.Click("Leave OTP examples"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "קוד אימות"); ok {
		t.Fatal("leaving source examples retained the actual editor")
	}
	if err := r.Click("Show OTP examples"); err != nil {
		t.Fatal(err)
	}
	if access(ui.RoleTextField, "קוד אימות").Value != "654321" {
		t.Fatal("page return lost the independent translated native draft")
	}
	r.SetSize(620, 900)
	native = false
	r.Frame()
	choose("demo")
	if *p.draft() != "aB9" || p.submitted["form"] != "AB1234" || p.changes["disabled"] != 0 {
		t.Fatal("source/style/page transitions mixed independent application drafts")
	}
}
