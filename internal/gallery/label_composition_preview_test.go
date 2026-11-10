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

func labelCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	if dir := os.Getenv("GALLERY_LABEL_COMPOSITION_SNAPSHOT_DIR"); dir != "" {
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
			t.Fatalf("Label image %v/%v", err, closeErr)
		}
	}
}

// One real-input flow checks native label relationships, click/focus/selection,
// disabled targets and the documentation's related full Field composition.
func TestLabelCompositionsSourceWorkflow(t *testing.T) {
	p := &labelCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Wrap().Children(func() {
				if ui.Button(c, "Leave Label examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show Label examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 850, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, name)
		if !ok {
			t.Fatalf("missing real native %v/%q", role, name)
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
		for i, o := range labelCompositionRecipes {
			if o.Value == key {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatal(key)
		}
		if err := r.Click("Label configuration"); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(labelCompositionRecipes[index].Label); err != nil {
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
				t.Fatal("styled Label configuration did not close")
			}
		}
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("recipe %s reached %s/%s", key, p.recipe, p.popupError)
		}
	}
	if len(labelCompositionRecipes) != 9 || p.recipe != "demo" {
		t.Fatal("incomplete Label source configurations")
	}
	caption := "Accept terms and conditions"
	checkbox := access(ui.RoleCheckBox, caption)
	if checkbox.Checked || checkbox.Name != caption {
		t.Fatal("source plain Label did not associate with the real native Checkbox")
	}
	if access(ui.RoleText, caption).Focusable {
		t.Fatal("a source label became an extra keyboard target")
	}
	click(ui.RoleText, caption)
	if !*p.checks["demo:terms"] || p.changes["demo"] != 1 {
		t.Fatal("native label click failed one checkbox change")
	}
	r.Key(0, ui.KeySpace)
	r.Frame()
	if *p.checks["demo:terms"] || p.changes["demo"] != 2 {
		t.Fatal("label activation did not hand native keyboard focus to the real checkbox")
	}
	labelCompositionSnapshot(t, r, "label-checkbox-light")
	for _, recipe := range labelCompositionRecipes {
		choose(recipe.Value)
	}
	choose("create-input")
	input := access(ui.RoleTextField, "Username")
	label := access(ui.RoleText, "Username")
	if input.Name != "Username" || label.Bounds.Y+label.Bounds.H > input.Bounds.Y || input.Placeholder != "Username" {
		t.Fatal("source before-target Label order/association was lost")
	}
	click(ui.RoleText, "Username")
	r.Type("native username")
	r.Key(ui.Cmd, ui.KeyA)
	r.Frame()
	before := p.changes["create-input"]
	click(ui.RoleText, "Username")
	if !r.Focused("Username") || *p.texts["create-input:value"] != "native username" || p.changes["create-input"] != before || p.selections["create-input"] != [2]int{0, len("native username")} {
		t.Fatal("source Label focus lost the native editor/selection or fabricated an edit")
	}
	choose("create-disabled")
	if !access(ui.RoleTextField, "Disabled").Disabled {
		t.Fatal("source disabled Label recipe lost its actual disabled native editor")
	}
	click(ui.RoleText, "Disabled")
	r.Type("rejected")
	r.Frame()
	if *p.texts["create-disabled:value"] != "" || p.changes["create-disabled"] != 0 {
		t.Fatal("disabled label focused an editable fake or accepted text")
	}
	dark = true
	r.Frame()
	labelCompositionSnapshot(t, r, "label-disabled-dark")
	choose("create-textarea")
	if !access(ui.RoleTextField, "Message").Multiline {
		t.Fatal("source textarea Label targeted a single-line editor")
	}
	click(ui.RoleText, "Message")
	r.Type("First line")
	r.Key(0, ui.KeyEnter)
	r.Type("Second line")
	if *p.texts["create-textarea:value"] != "First line\nSecond line" || !r.Focused("Message") {
		t.Fatal("source Label failed native multiline editing")
	}
	choose("usage")
	access(ui.RoleText, "Your email address")
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleTextField {
			t.Fatal("the bare guide Label invented an unprovided editor")
		}
	}
	choose("field-usage")
	click(ui.RoleText, "Your email address")
	r.Type("guide@example.com")
	if *p.texts["field-usage:value"] != "guide@example.com" {
		t.Fatal("guide FieldLabel failed the real native target")
	}
	r.SetSize(620, 1500)
	r.Frame()
	choose("field-example")
	dark = false
	r.Frame()
	if !access(ui.RoleCheckBox, "Same as shipping address").Checked || !r.HasText("All transactions are secure and encrypted") || !r.HasText("The billing address associated with your payment method") {
		t.Fatal("related Field demo lost its original content/default")
	}
	click(ui.RoleButton, "Submit")
	if p.paymentSubmits != 0 || p.paymentSaved != nil {
		t.Fatal("native Field form skipped source required values")
	}
	for _, field := range []struct{ label, value string }{{"Name on Card", "Evil Rabbit"}, {"Card Number", "1"}, {"CVV", "1"}, {"Comments", "local comments"}} {
		click(ui.RoleText, field.label)
		r.Type(field.value)
	}
	for _, selectField := range []struct{ label, value string }{{"Month", "03"}, {"Year", "2026"}} {
		click(ui.RoleMenuButton, selectField.label)
		if err := r.ChooseMenuItem(selectField.value); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	click(ui.RoleText, "Same as shipping address")
	click(ui.RoleButton, "Submit")
	if p.paymentSubmits != 1 || p.paymentSaved["name"] != "Evil Rabbit" || p.paymentSaved["number"] != "1" || p.paymentSaved["month"] != "03" || p.paymentSaved["year"] != "2026" || p.paymentSaved["comments"] != "local comments" || p.paymentSaved["shipping"] != "false" {
		t.Fatal("the related complete Field example lost a native label target or local snapshot")
	}
	labelCompositionSnapshot(t, r, "label-payment-field-light")
	click(ui.RoleButton, "Cancel")
	if p.action != "Local Label action: Cancel" || *p.texts["field-example:name"] != "Evil Rabbit" {
		t.Fatal("source Cancel invented a reset operation")
	}
	choose("rtl")
	click(ui.RoleMenuButton, "Label language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	r.SetSize(360, 850)
	dark = true
	r.Frame()
	caption = labelSourceTerms("he")
	checkbox = access(ui.RoleCheckBox, caption)
	label = access(ui.RoleText, caption)
	if checkbox.Bounds.X <= label.Bounds.X || label.Bounds.X < 20 || label.Bounds.X+label.Bounds.W > 340 {
		t.Fatal("native translated Label lost RTL placement or overflowed the narrow card")
	}
	click(ui.RoleText, caption)
	if !*p.checks["rtl:terms"] || p.changes["rtl"] != 1 {
		t.Fatal("translated native Label failed one activation")
	}
	labelCompositionSnapshot(t, r, "label-hebrew-narrow-dark")
	if err := r.Click("Leave Label examples"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.AccessByRole(ui.RoleCheckBox, caption); ok {
		t.Fatal("leaving source page retained its native target")
	}
	if err := r.Click("Show Label examples"); err != nil {
		t.Fatal(err)
	}
	if !access(ui.RoleCheckBox, caption).Checked {
		t.Fatal("page return lost the independent translated checkbox")
	}
	r.SetSize(620, 850)
	native = false
	r.Frame()
	choose("create-input")
	if access(ui.RoleTextField, "Username").Value != "native username" || *p.texts["create-textarea:value"] != "First line\nSecond line" || p.paymentSubmits != 1 {
		t.Fatal("style/recipe/page changes crossed independent Label drafts")
	}
}
