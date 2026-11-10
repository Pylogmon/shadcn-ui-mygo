package gallery

import (
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func kbdCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	if dir := os.Getenv("GALLERY_KBD_COMPOSITION_SNAPSHOT_DIR"); dir != "" {
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
			t.Fatalf("Kbd snapshot %v/%v", err, closeErr)
		}
	}
}

// One public native-input workflow exercises static keycaps in actual buttons,
// native editors and independently rendered tooltip children.
func TestKbdCompositionsSourceWorkflow(t *testing.T) {
	p := &kbdCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Wrap().Children(func() {
				if ui.Button(c, "Leave Kbd examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show Kbd examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
			ui.Button(c, "After Kbd examples")
		})
	}, 620, 850, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, name)
		if !ok {
			t.Fatalf("missing actual native %v/%q", role, name)
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
		for i, o := range kbdCompositionRecipes {
			if o.Value == key {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatal(key)
		}
		if err := r.Click("Kbd configuration"); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(kbdCompositionRecipes[index].Label); err != nil {
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
				t.Fatal("styled Kbd configuration did not close")
			}
		}
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("recipe %s reached %s/%s", key, p.recipe, p.popupError)
		}
	}
	if len(kbdCompositionRecipes) != 17 || p.recipe != "demo" {
		t.Fatal("incomplete fixed Kbd source catalogue")
	}
	for _, name := range []string{"⌘", "⇧", "⌥", "⌃", "Ctrl", "B"} {
		a := access(ui.RoleText, name)
		if a.Focusable || a.Disabled || a.Bounds.H != 20 || a.Bounds.W < 20 {
			t.Fatalf("source keycap is not inert20DIP native text: %+v", a)
		}
	}
	command, controlSymbol := access(ui.RoleText, "⌘"), access(ui.RoleText, "⌃")
	controlWord, letterB := access(ui.RoleText, "Ctrl"), access(ui.RoleText, "B")
	centeringDelta := (command.Bounds.X + controlSymbol.Bounds.X + controlSymbol.Bounds.W - controlWord.Bounds.X - letterB.Bounds.X - letterB.Bounds.W) / 2
	if centeringDelta > 1 || centeringDelta < -1 {
		t.Fatal("the source's two different key groups were not independently centered")
	}
	kbdCompositionSnapshot(t, r, "kbd-modifiers-light")
	for _, recipe := range kbdCompositionRecipes {
		choose(recipe.Value)
	}
	choose("group")
	if !r.HasText("Use") || !r.HasText("Ctrl + B") || !r.HasText("Ctrl + K") || !r.HasText("to open the command palette") {
		t.Fatal("source grouped instruction was collapsed")
	}
	r.Key(ui.Cmd, ui.KeyK)
	if p.accepted != 0 || len(b.surfaces) != 0 {
		t.Fatal("static keycaps installed an application shortcut")
	}
	choose("button")
	access(ui.RoleButton, "Accept")
	cap, hasCap := r.Find("⏎")
	if !hasCap {
		t.Fatal("the real Accept button omitted its inert return keycap")
	}
	r.ClickAt(cap.X+cap.W/2, cap.Y+cap.H/2)
	r.Key(0, ui.KeySpace)
	r.Frame()
	if p.accepted != 2 || p.action != "Local Kbd action: Accept" {
		t.Fatal("inert keycap blocked or duplicated its real parent button")
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("After Kbd examples") {
		t.Fatal("inert keycap added a Tab target")
	}
	choose("input-group")
	click(ui.RoleTextField, "Kbd source input")
	r.Type("native query")
	r.Key(ui.Cmd, ui.KeyA)
	r.SetClipboard("pasted query")
	r.Key(ui.Cmd, ui.KeyV)
	if *p.drafts["input-group"] != "pasted query" {
		t.Fatal("Kbd addon replaced the real native editor")
	}
	r.Key(ui.Cmd, ui.KeyZ)
	if *p.drafts["input-group"] != "native query" {
		t.Fatal("native paste undo was lost in the Kbd input group")
	}
	r.Key(0, ui.KeyRight)
	click(ui.RoleText, "K")
	r.Type("!")
	r.Key(ui.Cmd, ui.KeyK)
	if *p.drafts["input-group"] != "native query!" || !r.Focused("Kbd source input") || p.accepted != 2 {
		t.Fatalf("inert Kbd addon value=%q focused=%v accepted=%d", *p.drafts["input-group"], r.Focused("Kbd source input"), p.accepted)
	}
	choose("create-input")
	if access(ui.RoleTextField, "Kbd source input").Placeholder != "" || !r.HasText("Space") {
		t.Fatal("Create input inherited the docs-only search placeholder")
	}
	click(ui.RoleTextField, "Kbd source input")
	r.Type("independent draft")
	choose("create-icons")
	for _, name := range []string{"Dashed circle", "Arrow left", "Arrow right"} {
		if access(ui.RoleText, name).Focusable {
			t.Fatal("decorative key icon became interactive")
		}
	}
	choose("create-icon-text")
	if !r.HasText("Left") || !r.HasText("Voice Enabled") {
		t.Fatal("source icon-and-text keycaps were truncated")
	}
	dark = true
	r.Frame()
	kbdCompositionSnapshot(t, r, "kbd-icon-text-dark")
	choose("create-samp")
	if access(ui.RoleText, "File").Bounds.H != 20 {
		t.Fatal("native monospace sample lost the keycap box")
	}
	choose("tooltip")
	save := access(ui.RoleButton, "Save")
	r.Move(save.Bounds.X+save.Bounds.W/2, save.Bounds.Y+save.Bounds.H/2)
	r.Frame()
	if len(b.surfaces) == 0 {
		t.Fatal("native tooltip route did not open a separate child")
	}
	s := b.surfaces[len(b.surfaces)-1]
	if s.spec.Focusable || !s.spec.IgnoreMouseEvents || !s.spec.ShowInactive || r.HasText("Save Changes") || !s.tester.HasText("Save Changes") {
		t.Fatal("rich Kbd tooltip did not use an inert independent child")
	}
	if a, ok := s.tester.Find("S"); !ok || a.H != 20 {
		t.Fatal("tooltip child omitted its actual keycap visuals")
	}
	for _, a := range s.tester.AccessAll() {
		if a.Focusable {
			t.Fatal("the tooltip keycap exposed a focus target")
		}
	}
	print := access(ui.RoleButton, "Print")
	r.Move(print.Bounds.X+print.Bounds.W/2, print.Bounds.Y+print.Bounds.H/2)
	r.Frame()
	s = b.surfaces[len(b.surfaces)-1]
	if r.HasText("Print Document") || !s.tester.HasText("Print Document") {
		t.Fatal("Print did not build its separate rich tooltip")
	}
	for _, key := range []string{"Ctrl", "P"} {
		if cap, ok := s.tester.Find(key); !ok || cap.H != 20 {
			t.Fatal("Print tooltip lost its original grouped keys")
		}
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if !s.didClose {
		t.Fatal("the real parent Escape did not dismiss its inert tooltip")
	}
	r.Move(5, 840)
	r.Frame()
	native = false
	r.Frame()
	choose("tooltip")
	save = access(ui.RoleButton, "Save")
	r.Move(save.Bounds.X+save.Bounds.W/2, save.Bounds.Y+save.Bounds.H/2)
	r.Frame()
	s = b.surfaces[len(b.surfaces)-1]
	if !s.spec.FitContent || s.spec.Focusable {
		t.Fatal("styled tooltip lost intrinsic fitting or took focus")
	}
	// This injected Backend does not emulate the operating system's FitContent
	// loop. Expand the software canvas to inspect the actual rich child safely.
	s.tester.SetSize(s.spec.Width, 80)
	s.tester.Frame()
	key, ok := s.tester.Find("S")
	if !ok {
		t.Fatal("styled tooltip keycap missing")
	}
	tokens := theme.Preset("blue", true)
	fg, bg := tokens.Background.Alpha(.1), tokens.Foreground
	blend := func(b, f uint8) uint8 { return uint8((int(f)*int(fg.A) + int(b)*(255-int(fg.A)) + 127) / 255) }
	pixel := color.NRGBAModel.Convert(s.tester.Image().At(int(key.X+1), int(key.Y+key.H/2))).(color.NRGBA)
	if pixel.R != blend(bg.R, fg.R) || pixel.G != blend(bg.G, fg.G) || pixel.B != blend(bg.B, fg.B) {
		t.Fatalf("styled source keycap lost inverse tooltip tint: %v", pixel)
	}
	kbdCompositionSnapshot(t, s.tester, "kbd-tooltip-dark")
	r.Move(5, 840)
	r.Frame()
	choose("create-tooltip")
	save = access(ui.RoleButton, "Save changes")
	if save.Bounds.W != 32 || save.Bounds.H != 32 {
		t.Fatal("Create tooltip omitted its original small icon button")
	}
	r.Move(save.Bounds.X+save.Bounds.W/2, save.Bounds.Y+save.Bounds.H/2)
	r.Frame()
	s = b.surfaces[len(b.surfaces)-1]
	if !s.tester.HasText("Save Changes") {
		t.Fatal("Create icon trigger omitted its real rich tooltip")
	}
	r.Move(5, 840)
	r.Frame()
	choose("rtl")
	if err := r.Click("Kbd language"); err != nil {
		t.Fatal(err)
	}
	s = b.surfaces[len(b.surfaces)-1]
	s.tester.Key(0, ui.KeyHome)
	s.tester.Key(0, ui.KeyDown)
	s.tester.Key(0, ui.KeyEnter)
	r.Frame()
	if p.language != "he" {
		t.Fatal("the actual language selector failed Hebrew")
	}
	r.SetSize(360, 850)
	r.Frame()
	command, shift := access(ui.RoleText, "⌘"), access(ui.RoleText, "⇧")
	if command.Bounds.X <= shift.Bounds.X || command.Bounds.Y != shift.Bounds.Y {
		t.Fatal("RTL key groups lost logical source order")
	}
	kbdCompositionSnapshot(t, r, "kbd-hebrew-narrow-dark")
	if err := r.Click("Leave Kbd examples"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.AccessByRole(ui.RoleText, "⌘"); ok {
		t.Fatal("leaving Kbd source page retained native keycaps")
	}
	if err := r.Click("Show Kbd examples"); err != nil {
		t.Fatal(err)
	}
	r.SetSize(620, 850)
	r.Frame()
	choose("input-group")
	if access(ui.RoleTextField, "Kbd source input").Value != "native query!" || *p.drafts["create-input"] != "independent draft" || p.accepted != 2 {
		t.Fatal("source/page/style changes mixed independent Kbd application drafts")
	}
}
