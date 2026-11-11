package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func inputSourceSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_INPUT_COMPOSITION_SNAPSHOT_DIR")
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
	cerr := f.Close()
	if err != nil || cerr != nil {
		t.Fatalf("image %v/%v", err, cerr)
	}
}

// One public native-input flow uses current-frame editors and value-only picker
// injection. Software images do not establish OS picker or platform IME/AT proof.
func TestInputCompositionsSourceWorkflow(t *testing.T) {
	p := &inputCompositionPreviewState{}
	b := &creatorPopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Role(ui.RoleRegion).Label("Input workspace").Children(func() {
				ui.Column(c).FillWidth().MinWidth(0).Children(func() {
					if visible {
						p.view(c, nil, h, native)
					} else {
						ui.Text(c, "Another application page")
					}
				})
			})
		})
	}, 620, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.SetRegion(i18n.Region{Locale: "en-GB", HourCycle: 24, FirstWeekday: time.Monday})
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		for range 8 {
			a, ok := r.AccessByRole(role, name)
			v, vok := r.AccessByRole(ui.RoleRegion, "Input workspace")
			if !ok || !vok {
				t.Fatalf("missing native source target %v/%s", role, name)
			}
			if a.Bounds.Y >= v.Bounds.Y && a.Bounds.Y+a.Bounds.H <= v.Bounds.Y+v.Bounds.H {
				if a.Bounds.W <= 0 || a.Bounds.X < v.Bounds.X-1 || a.Bounds.X+a.Bounds.W > v.Bounds.X+v.Bounds.W+1 {
					t.Fatalf("source target exceeds viewport: %s %v in %v", name, a.Bounds, v.Bounds)
				}
				return a
			}
			r.Scroll(v.Bounds.X+v.Bounds.W-12, v.Bounds.Y+v.Bounds.H/2, 0, a.Bounds.Y+a.Bounds.H/2-v.Bounds.Y-v.Bounds.H/2)
		}
		t.Fatalf("cannot reveal %s", name)
		return ui.AccessInfo{}
	}
	click := func(role ui.Role, name string) {
		t.Helper()
		a := access(role, name)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	replace := func(name, text string) {
		t.Helper()
		click(ui.RoleTextField, name)
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(text)
		r.Frame()
	}
	choose := func(key string) {
		t.Helper()
		role := ui.RoleMenuButton
		if !native {
			role = ui.RoleComboBox
		}
		click(role, "Input configuration")
		if native {
			for _, o := range inputCompositionRecipes {
				if o.Value == key {
					if err := r.ChooseMenuItem(o.Label); err != nil {
						t.Fatal(err)
					}
					break
				}
			}
		} else {
			s := b.surfaces[len(b.surfaces)-1]
			s.tester.Key(0, ui.KeyHome)
			for _, o := range inputCompositionRecipes {
				if o.Value == key {
					break
				}
				s.tester.Key(0, ui.KeyDown)
			}
			s.tester.Key(0, ui.KeyEnter)
			if !s.didClose {
				t.Fatal("styled selector did not close after choice")
			}
		}
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("configuration %q got %q error %q", key, p.recipe, p.popupError)
		}
	}
	if p.recipe != "demo" || len(inputCompositionRecipes) != 25 || !access(ui.RoleTextField, "API Key").Password {
		t.Fatal("source API-key default or complete configurations lost")
	}
	replace("API Key", "sk-native-secret")
	if *p.drafts["demo"].texts["api"] != "sk-native-secret" || p.action != "Local input edit: API Key" {
		t.Fatal("actual protected native editor did not update its model")
	}
	for _, o := range inputCompositionRecipes {
		choose(o.Value)
	}
	choose("disabled")
	a := access(ui.RoleTextField, "Email")
	if !a.Disabled {
		t.Fatal("disabled field did not reach actual native editor")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Type("blocked")
	if *p.draft().text("email") != "" {
		t.Fatal("disabled source accepted input")
	}
	choose("invalid")
	replace("Invalid Input", "editable invalid draft")
	if !access(ui.RoleTextField, "Invalid Input").Invalid || *p.draft().text("invalid") != "editable invalid draft" {
		t.Fatal("invalid source was incorrectly disabled")
	}
	choose("fieldgroup")
	replace("Name", "Jordan")
	replace("Email", "bad-email")
	click(ui.RoleButton, "Submit")
	if p.draft().error == "" || p.draft().saved != nil {
		t.Fatal("source email form accepted malformed value")
	}
	replace("Email", "jordan@example.com")
	click(ui.RoleButton, "Submit")
	if p.draft().saved["name"] != "Jordan" || p.draft().saved["email"] != "jordan@example.com" {
		t.Fatal("real source form did not save local values")
	}
	click(ui.RoleButton, "Reset")
	if *p.draft().text("name") != "" || *p.draft().text("email") != "" || p.draft().saved != nil {
		t.Fatal("source reset did not restore empty defaults")
	}
	choose("form")
	click(ui.RoleButton, "Submit")
	if p.draft().error != "Name is required." || !r.Focused("Name") {
		t.Fatal("docs required Name was not enforced")
	}
	replace("Name", "Evil Rabbit")
	replace("Email", "rabbit@example.com")
	replace("Phone", "+1 555 123 4567")
	replace("Address", "123 Main St")
	click(ui.RoleMenuButton, "Country")
	if err := r.ChooseMenuItem("Canada"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	click(ui.RoleButton, "Submit")
	if p.draft().saved["country"] != "ca" || p.draft().saved["name"] != "Evil Rabbit" {
		t.Fatal("source form select and text values did not share the model")
	}
	inputSourceSnapshot(t, r, "input-contact-form-light")
	choose("create-form")
	click(ui.RoleButton, "Submit")
	if p.draft().error != "" || p.draft().submitted != 1 {
		t.Fatal("Create form inherited the docs-only required Name rule")
	}
	choose("button-group")
	replace("Search", "native grouped query")
	r.Key(0, ui.KeyEnter)
	if p.draft().submitted != 0 {
		t.Fatal("the plain source Button Group editor invented an Enter search action")
	}
	click(ui.RoleButton, "Search")
	if p.draft().saved["query"] != "native grouped query" {
		t.Fatal("joined group action did not use its real editor")
	}
	choose("input-group")
	replace("Website URL", "native.example")
	if *p.draft().text("website") != "native.example" {
		t.Fatal("addon group did not bind its editor")
	}
	if err := r.Click("https://"); err != nil {
		t.Fatal(err)
	}
	r.Type(".org")
	if *p.draft().text("website") != "native.example.org" {
		t.Fatal("prefix addon did not focus the actual editor")
	}
	choose("create-select")
	replace("Amount", "125")
	click(ui.RoleMenuButton, "Currency")
	if err := r.ChooseMenuItem("EUR"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.draft().choice != "eur" || *p.draft().text("amount") != "125" {
		t.Fatal("currency/source amount drafts lost")
	}
	choose("create-native-select")
	click(ui.RoleMenuButton, "Phone country code")
	if err := r.ChooseMenuItem("+44"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.draft().choice != "+44" {
		t.Fatal("source native select did not use its real route")
	}
	choose("create-types")
	if a := access(ui.RoleTextField, "Number"); a.Value != "" || a.Placeholder != "123" || !p.draft().numberEmpty {
		t.Fatal("source numeric placeholder was replaced by a forced zero")
	}
	replace("Number", "42.5")
	r.Key(0, ui.KeyUp)
	if p.draft().number != 43.5 || p.draft().numberEmpty {
		t.Fatal("numeric native input failed edit/step")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(0, ui.KeyBackspace)
	r.Frame()
	if !p.draft().numberEmpty || p.draft().number != 43.5 || access(ui.RoleTextField, "Number").Value != "" {
		t.Fatal("numeric clear lost empty presence or the last valid number")
	}
	beforeNumberChanges := p.draft().changes
	replace("Number", "0")
	if p.draft().numberEmpty || p.draft().number != 0 || p.draft().changes != beforeNumberChanges+1 {
		t.Fatal("present zero was confused with empty or counted twice")
	}

	beforeDate := p.draft().date
	click(ui.RolePopUpButton, "Date")
	r.Key(0, ui.KeyRight)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p.draft().date.Format("2006-01-02") != beforeDate.AddDate(0, 0, 1).Format("2006-01-02") {
		t.Fatal("native date popup selection did not change its typed model to the next day")
	}
	click(ui.RoleStepper, "Time hours")
	r.Key(0, ui.KeyUp)
	if p.draft().clock.Hour() != 1 {
		t.Fatal("native typed time segment did not change")
	}
	dark = true
	r.SetSize(620, 1200)
	r.Frame()
	access(ui.RoleTextField, "Password")
	inputSourceSnapshot(t, r, "input-typed-controls-dark")
	r.SetSize(620, 900)
	picked := make(chan mygo.OpenDialogOptions, 1)
	p.picker = func(opts mygo.OpenDialogOptions) ([]string, error) {
		picked <- opts
		return []string{"/tmp/source-picture.png"}, nil
	}
	choose("file")
	click(ui.RoleButton, "Picture")
	deadline := time.Now().Add(time.Second)
	for len(p.draft().files) == 0 && time.Now().Before(deadline) {
		r.Frame()
		time.Sleep(time.Millisecond)
	}
	if len(p.draft().files) != 1 || p.draft().files[0] != "/tmp/source-picture.png" {
		t.Fatal("injected native picker result did not reach source model")
	}
	select {
	case <-picked:
	default:
		t.Fatal("file input fabricated a text selection without calling picker")
	}
	choose("grid")
	r.SetSize(360, 900)
	replace("First Name", "Jordan")
	replace("Last Name", "Lee")
	if *p.draft().text("first") != "Jordan" || *p.draft().text("last") != "Lee" {
		t.Fatal("narrow two-column native editing failed")
	}
	inputSourceSnapshot(t, r, "input-grid-narrow-dark")
	r.SetSize(620, 900)
	choose("rtl")
	click(ui.RoleMenuButton, "Input language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	replace("מפתח API", "sk-hebrew")
	inputSourceSnapshot(t, r, "input-hebrew-dark")
	choose("demo")
	if *p.draft().text("api") != "sk-native-secret" {
		t.Fatal("source configuration transition mixed secret drafts")
	}
	visible = false
	r.Frame()
	visible = true
	r.Frame()
	if *p.draft().text("api") != "sk-native-secret" || !access(ui.RoleTextField, "API Key").Password {
		t.Fatal("hide/return lost protected model")
	}
	native = false
	r.Frame()
	choose("create-basic")
	replace("Email", "flyout@example.com")
	if *p.draft().text("email") != "flyout@example.com" {
		t.Fatal("styled source route failed real input")
	}
}
