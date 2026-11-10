package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func radioCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_RADIO_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("radio image %v/%v", err, closeErr)
	}
}

// One public-input workflow exercises source choices, one-option and whole-
// group disabling, Field relationships, actual card activation and native
// exclusivity before an additional application frame. The injected popup
// backend and software images are not operating-system or AT acceptance.
func TestRadioCompositionsSourceWorkflow(t *testing.T) {
	p := &radioCompositionPreviewState{}
	b := &creatorPopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 800, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, label string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, label)
		if !ok {
			t.Fatalf("missing source target %v/%s", role, label)
		}
		return a
	}
	choose := func(selector, label string) {
		t.Helper()
		if err := r.Click(selector); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(label); err != nil {
				t.Fatal(err)
			}
		} else {
			s := b.surfaces[len(b.surfaces)-1]
			s.tester.Key(0, ui.KeyHome)
			for _, option := range radioCompositionRecipes {
				if option.Label == label {
					break
				}
				s.tester.Key(0, ui.KeyDown)
			}
			s.tester.Key(0, ui.KeyEnter)
		}
		r.Frame()
		if p.popupError != "" {
			t.Fatal(p.popupError)
		}
	}
	clickRadio := func(label string) {
		t.Helper()
		a := access(ui.RoleRadio, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	}
	unique := func(label string) {
		t.Helper()
		count := 0
		for _, a := range r.AccessAll() {
			if a.Role == ui.RoleRadio && a.Checked {
				count++
				if a.Name != label {
					t.Fatalf("stale selected radio %s while choosing %s", a.Name, label)
				}
			}
		}
		if count != 1 {
			t.Fatalf("native group has %d checked radios, want1", count)
		}
	}
	if p.recipe != "demo" || !access(ui.RoleRadio, "Comfortable").Checked {
		t.Fatal("source default Comfortable was not selected")
	}
	for _, recipe := range radioCompositionRecipes {
		choose("Radio Group configuration", recipe.Label)
		if p.recipe != recipe.Value {
			t.Fatalf("source selection %s built %s", recipe.Value, p.recipe)
		}
	}
	choose("Radio Group configuration", "Spacing")
	clickRadio("Compact")
	unique("Compact") // No explicit extra Frame may hide transient double checks.
	if *p.values["demo"] != "compact" || p.changes["demo"] != 1 || p.action != "Local radio choice: Compact" {
		t.Fatal("pointer activation was not a single native source change")
	}
	r.Key(0, ui.KeyUp)
	unique("Comfortable")
	r.Key(0, ui.KeySpace)
	if *p.values["demo"] != "comfortable" || p.changes["demo"] != 2 {
		t.Fatal("Space toggled an already-selected radio or replayed a callback")
	}
	choose("Radio Group configuration", "One disabled option")
	if !access(ui.RoleRadio, "Disabled").Disabled || !access(ui.RoleRadio, "Option 2").Checked || access(ui.RoleRadio, "Option 3").Disabled {
		t.Fatal("docs one-disabled source was changed to a disabled group")
	}
	clickRadio("Disabled")
	if *p.values["disabled"] != "option2" || p.changes["disabled"] != 0 {
		t.Fatal("disabled native radio changed its model")
	}
	clickRadio("Option 2")
	r.Key(0, ui.KeyDown)
	unique("Option 3")
	r.Key(0, ui.KeyDown)
	unique("Option 2")
	choose("Radio Group configuration", "Create: Disabled Group")
	for _, name := range []string{"Option 1", "Option 2", "Option 3"} {
		if !access(ui.RoleRadio, name).Disabled {
			t.Fatal("Create disabled-group semantics lost")
		}
	}
	clickRadio("Option 1")
	if *p.values["create-disabled"] != "option2" || p.changes["create-disabled"] != 0 {
		t.Fatal("disabled whole group accepted an activation")
	}
	choose("Radio Group configuration", "Invalid notification preferences")
	if !r.HasText("Choose how you want to receive notifications.") {
		t.Fatal("source fieldset description missing")
	}
	for _, name := range []string{"Email only", "SMS only", "Both Email & SMS"} {
		if !access(ui.RoleRadio, name).Invalid {
			t.Fatal("source Invalid did not reach the actual radio")
		}
	}
	clickRadio("SMS only")
	unique("SMS only")
	choose("Radio Group configuration", "Subscription plan fieldset")
	clickRadio("Yearly ($99.99/year)")
	unique("Yearly ($99.99/year)")
	if *p.values["create-fieldset"] != "medium" {
		t.Fatal("subscription and battery fieldset models were coupled")
	}
	radioCompositionSnapshot(t, r, "radio-subscription-light")
	choose("Radio Group configuration", "Plan choice cards")
	card := access(ui.RoleGroup, "Pro")
	r.ClickAt(card.Bounds.X+card.Bounds.W-4, card.Bounds.Y+card.Bounds.H/2)
	unique("Pro")
	if *p.values["choice-card"] != "pro" || p.changes["choice-card"] != 1 || !r.HasText("For growing businesses.") || access(ui.RoleRadio, "Pro").Description != "For growing businesses." {
		t.Fatal("whole source card or its native description association failed")
	}
	radioCompositionSnapshot(t, r, "radio-plan-cards-light")
	r.Key(0, ui.KeyDown)
	unique("Enterprise")
	dark = true
	r.Frame()
	choose("Radio Group configuration", "Create: Grid Layout")
	clickRadio("Small")
	r.Key(0, ui.KeyRight)
	unique("Medium")
	r.Key(0, ui.KeyDown)
	unique("Large")
	if *p.values["create-grid"] != "large" || *p.values["choice-card"] != "enterprise" || *p.values["fieldset"] != "yearly" {
		t.Fatal("grid navigation lost independent source values")
	}
	radioCompositionSnapshot(t, r, "radio-grid-dark")
	choose("Radio Group configuration", "Translated spacing descriptions")
	choose("Radio Group language", "עברית")
	clickRadio("ברירת מחדל")
	r.Key(0, ui.KeyLeft)
	unique("נוח")
	if *p.values["rtl"] != "comfortable" || access(ui.RoleRadio, "נוח").Description != "יותר מקום בין האלמנטים." {
		t.Fatal("RTL arrow navigation or source description changed")
	}
	radioCompositionSnapshot(t, r, "radio-hebrew-dark")
	visible = false
	r.Frame()
	visible = true
	r.Frame()
	unique("נוח")
	native = false
	r.Frame()
	choose("Radio Group configuration", "Create: Grid Layout")
	if len(b.surfaces) == 0 || h.IsOpen("source-radio-config") || *p.values["create-grid"] != "large" {
		t.Fatal("styled selector did not close on select or source model was reset")
	}
	unique("Large")
}
