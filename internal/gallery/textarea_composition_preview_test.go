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

func textareaCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_TEXTAREA_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("textarea image %v/%v", err, closeErr)
	}
}

// One public native-input workflow checks source editors and their application
// drafts. Software Tester/Flyout images do not establish OS/AT/IME acceptance.
func TestTextareaCompositionsSourceWorkflow(t *testing.T) {
	p := &textareaCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Wrap().Children(func() {
				if ui.Button(c, "Leave textarea examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show textarea examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 800, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(ui.RoleTextField, name)
		if !ok {
			t.Fatalf("missing native multiline editor %q", name)
		}
		// Pinned MyGo returns from disabled AX nodes before editor metadata.
		// Keep the real disabled input; do not replace it with a fake AX node.
		if !a.Disabled && !a.Multiline {
			t.Fatalf("source textarea %q is not a native multiline field", name)
		}
		return a
	}
	choose := func(key string) {
		t.Helper()
		index := -1
		for i, option := range textareaCompositionRecipes {
			if option.Value == key {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatalf("unknown source recipe %q", key)
		}
		if err := r.Click("Textarea configuration"); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(textareaCompositionRecipes[index].Label); err != nil {
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
				t.Fatal("styled configuration selector did not close after selection")
			}
		}
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("configuration %q reached %q error=%s", key, p.recipe, p.popupError)
		}
	}
	clickEditor := func(name string) {
		t.Helper()
		a := access(name)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	}
	const placeholder = "Type your message here."
	if p.recipe != "demo" || len(textareaCompositionRecipes) != 11 || access(placeholder).Bounds.W != 320 || access(placeholder).Bounds.H < 64 {
		t.Fatal("source default, width or minimum native height changed")
	}
	clickEditor(placeholder)
	r.Type("First line")
	r.Key(0, ui.KeyEnter)
	r.Type("Second line")
	if *p.drafts["demo"] != "First line\nSecond line" || p.sends != 0 {
		t.Fatal("Enter failed to insert a real newline or invented a send action")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Frame()
	if p.selections["demo"] != [2]int{0, len(*p.drafts["demo"])} {
		t.Fatal("the actual native textarea did not retain its selection")
	}
	r.SetClipboard("Pasted\nmessage")
	r.Key(ui.Cmd, ui.KeyV)
	if *p.drafts["demo"] != "Pasted\nmessage" {
		t.Fatal("native multiline paste did not replace the selected text")
	}
	r.Key(ui.Cmd, ui.KeyZ)
	if *p.drafts["demo"] != "First line\nSecond line" {
		t.Fatal("native undo did not restore the selected multiline draft")
	}
	for _, recipe := range textareaCompositionRecipes {
		choose(recipe.Value)
	}
	choose("field")
	editor := access("Message")
	description, ok := r.Find("Enter your message below.")
	if !ok || description.Y+description.H > editor.Bounds.Y || !strings.Contains(editor.Description, "Enter your message below.") {
		t.Fatal("the docs Field description did not precede and describe its real editor")
	}
	label, ok := r.Find("Message")
	if !ok {
		t.Fatal("the source Field label did not associate with the actual editor")
	}
	r.ClickAt(label.X+label.W/2, label.Y+label.H/2)
	if !access("Message").Focused {
		t.Fatal("the real source Label did not focus its native editor")
	}
	r.Type("Field draft")
	textareaCompositionSnapshot(t, r, "textarea-field-light")
	choose("disabled")
	if !access("Message").Disabled {
		t.Fatal("the source disabled Field failed to disable its actual textarea")
	}
	clickEditor("Message")
	r.Type("Rejected")
	if *p.drafts["disabled"] != "" || p.changes["disabled"] != 0 {
		t.Fatal("the disabled native textarea accepted input")
	}
	choose("create-disabled")
	if !access("Message").Disabled {
		t.Fatal("Create's disabled editor lost its native disabled state")
	}
	choose("invalid")
	if !access("Message").Invalid || access("Message").Disabled {
		t.Fatal("invalid state did not reach the editor independently of disabled")
	}
	clickEditor("Message")
	r.Type("Editable error draft")
	if *p.drafts["invalid"] != "Editable error draft" {
		t.Fatal("the invalid source editor incorrectly blocked editing")
	}
	dark = true
	r.Frame()
	textareaCompositionSnapshot(t, r, "textarea-invalid-dark")
	choose("create-description")
	six := access("Message")
	description, ok = r.Find("Type your message and press enter to send.")
	if !ok || description.Y < six.Bounds.Y+six.Bounds.H || six.Bounds.H <= 120 {
		t.Fatal("Create's declared six rows and following description were lost")
	}
	clickEditor("Message")
	r.Type("No implicit send")
	r.Key(0, ui.KeyEnter)
	if *p.drafts["create-description"] != "No implicit send\n" || p.sends != 0 {
		t.Fatal("description copy was incorrectly treated as an Enter submit handler")
	}
	textareaCompositionSnapshot(t, r, "textarea-six-rows-dark")
	choose("button")
	clickEditor(placeholder)
	r.Type("Send this\nkeep my draft")
	send, ok := r.AccessByRole(ui.RoleButton, "Send message")
	if !ok {
		t.Fatal("source Send message did not expose its actual native button")
	}
	r.ClickAt(send.Bounds.X+send.Bounds.W/2, send.Bounds.Y+send.Bounds.H/2)
	if p.sends != 1 || p.submitted["button"] != "Send this\nkeep my draft" || *p.drafts["button"] != p.submitted["button"] || p.action != "Local textarea action: Send message" {
		t.Fatalf("source button snapshot sends=%d submitted=%q draft=%q action=%q", p.sends, p.submitted["button"], *p.drafts["button"], p.action)
	}
	choose("rtl")
	if err := r.Click("Textarea language"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	r.SetSize(360, 800)
	feedback := access("משוב")
	if feedback.Bounds.W > 300 || feedback.Bounds.X < 20 || feedback.Bounds.X+feedback.Bounds.W > 340 || feedback.Bounds.H >= six.Bounds.H {
		t.Fatal("the four-row RTL source editor overflowed its narrow parent or used six rows")
	}
	clickEditor("משוב")
	r.Type("משוב מקומי\nשורה שנייה")
	textareaCompositionSnapshot(t, r, "textarea-hebrew-dark")
	if err := r.Click("Leave textarea examples"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "משוב"); ok {
		t.Fatal("leaving the source page retained a native editor")
	}
	if err := r.Click("Show textarea examples"); err != nil {
		t.Fatal(err)
	}
	if access("משוב").Value != *p.drafts["rtl"] {
		t.Fatal("page return lost the independent RTL native draft")
	}
	r.SetSize(620, 800)
	native = false
	r.Frame()
	choose("demo")
	if access(placeholder).Value != "First line\nSecond line" || p.sends != 1 || p.submitted["button"] != "Send this\nkeep my draft" {
		t.Fatal("changing selector style or source recipe crossed independent application drafts")
	}
}
