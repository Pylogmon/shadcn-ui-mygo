package gallery

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestCreatorPresetComposesIndependentThemesAndFonts(t *testing.T) {
	p := DefaultPreset()
	p.Base = "mauve"
	p.Accent = "blue"
	p.Chart = "orange"
	p.Radius = 0
	p.BodyFont = "mono"
	p.HeadingFont = "serif"
	for _, dark := range []bool{false, true} {
		tokens := p.Tokens(dark)
		base, accent, chart := theme.Preset("mauve", dark), theme.Preset("blue", dark), theme.Preset("orange", dark)
		if tokens.Background != base.Background || tokens.Muted != base.Muted || tokens.Primary != accent.Primary || tokens.PrimaryForeground != accent.PrimaryForeground || tokens.Chart1 != chart.Chart1 || tokens.Chart5 != chart.Chart5 || tokens.Radius != 0 {
			t.Fatalf("composed tokens mismatch: %+v", tokens)
		}
		if theme.Preset("mauve", dark).Primary == accent.Primary {
			t.Fatal("preset composition mutated shared base")
		}
	}
	if p.BodyFamily() != "monospace" || !strings.Contains(p.HeadingFamily(), "Georgia") {
		t.Fatalf("font selection has no real native family mapping: %q / %q", p.BodyFamily(), p.HeadingFamily())
	}
	p.HeadingFont = "same"
	if p.HeadingFamily() != p.BodyFamily() {
		t.Fatal("same-as-body does not inherit body font")
	}
}

func TestCreatorImportValidatesBeforeChangingState(t *testing.T) {
	s := New()
	before := s.Config
	invalid := []string{`{}`, strings.Replace(before.JSON(), `"version": 1`, `"version": 2`, 1), strings.Replace(before.JSON(), `"base": "neutral"`, `"base": "unknown"`, 1), strings.Replace(before.JSON(), `"radius": 10`, `"radius": -1`, 1), strings.Replace(before.JSON(), `"bodyFont": "system"`, `"bodyFont": "unavailable-web-font"`, 1), strings.Replace(before.JSON(), `"version": 1`, `"extra": true, "version": 1`, 1), before.JSON() + before.JSON(), strings.Repeat("x", 16385)}
	for _, source := range invalid {
		if err := s.ImportPreset(source); err == nil {
			t.Fatalf("accepted invalid preset %q", source)
		}
		if s.Config != before || s.CanUndo() {
			t.Fatal("invalid import changed configuration/history")
		}
	}
	p := before
	p.Base = "olive"
	p.Accent = "violet"
	p.Appearance = theme.DarkMode
	if err := s.ImportPreset(p.JSON()); err != nil {
		t.Fatal(err)
	}
	if s.Config != p || s.Theme.Mode != theme.DarkMode || !s.CanUndo() {
		t.Fatalf("valid import failed: %+v", s.Config)
	}
	parsed, err := ParsePreset(s.ExportPreset())
	if err != nil || parsed != p {
		t.Fatalf("round trip=%+v err=%v", parsed, err)
	}
}

func TestCreatorUndoRedoRandomAndResetPreservePreviewState(t *testing.T) {
	s := New()
	original := s.Config
	s.Name = "A persistent form value"
	s.Randomize(11)
	randomized := s.Config
	if randomized == original || randomized.Validate() != nil {
		t.Fatalf("randomization=%+v", randomized)
	}
	if !s.Undo() || s.Config != original || !s.Redo() || s.Config != randomized {
		t.Fatal("undo/redo did not restore exact designs")
	}
	s.Reset()
	if s.Config != original || s.Name != "A persistent form value" {
		t.Fatal("reset altered form state or failed design reset")
	}
	s.Undo()
	next := s.Config
	next.Radius = 32
	if err := s.ApplyPreset(next); err != nil {
		t.Fatal(err)
	}
	if s.CanRedo() {
		t.Fatal("new design did not clear redo branch")
	}
	for i := range 105 {
		next = s.Config
		next.Radius = float32(i % 32)
		_ = s.ApplyPreset(next)
	}
	if len(s.undo) > 100 {
		t.Fatal("unbounded design history")
	}
}

func TestCreatorNativeChoiceChangesPreviewAndCanUndo(t *testing.T) {
	s := New()
	s.Config.NativeMenus = true
	r := ui.NewTester(s.View, 1280, 900)
	field, ok := r.AccessByRole(ui.RoleMenuButton, "Theme")
	if !ok {
		t.Fatal("native theme selector missing")
	}
	r.ClickAt(field.Bounds.X+field.Bounds.W/2, field.Bounds.Y+field.Bounds.H/2)
	if err := r.ChooseMenuItem("Blue"); err != nil {
		t.Fatal(err)
	}
	if s.Config.Accent != "blue" || s.Config.Tokens(false).Primary != theme.Preset("blue", false).Primary {
		t.Fatalf("native menu did not change real tokens: %+v", s.Config)
	}
	if !s.CanUndo() {
		t.Fatal("native menu change did not enter history")
	}
	if err := r.Click("Undo design change"); err != nil {
		t.Fatal(err)
	}
	if s.Config.Accent != "neutral" {
		t.Fatal("UI undo did not restore accent")
	}
}

func TestCreatorClipboardImportAndErrorRecovery(t *testing.T) {
	s := New()
	r := ui.NewTester(s.View, 1280, 900)
	if err := r.Click("Copy preset"); err != nil {
		t.Fatal(err)
	}
	if copied, err := ParsePreset(r.Clipboard()); err != nil || copied != s.Config {
		t.Fatalf("clipboard preset=%+v err=%v", copied, err)
	}
	if err := r.Click("Open preset"); err != nil {
		t.Fatal(err)
	}
	s.importText = `{"version":99}`
	if err := r.Click("Apply preset"); err != nil {
		t.Fatal(err)
	}
	if !s.importOpen || s.importError == "" {
		t.Fatal("bad import did not preserve dialog with validation feedback")
	}
	p := s.Config
	p.Base = "stone"
	p.Chart = "red"
	p.BodyFont = "mono"
	r.SetClipboard(p.JSON())
	if err := r.Click("Paste clipboard"); err != nil {
		t.Fatal(err)
	}
	if err := r.Click("Apply preset"); err != nil {
		t.Fatal(err)
	}
	if s.importOpen || s.Config != p {
		t.Fatalf("valid import did not apply: %+v", s.Config)
	}
}

func TestCreatorResponsiveLayoutKeepsCustomizerAndLivePreviewVisible(t *testing.T) {
	for _, size := range [][2]int{{1280, 900}, {800, 900}, {520, 900}} {
		for _, mode := range []theme.Mode{theme.LightMode, theme.DarkMode} {
			s := New()
			p := s.Config
			p.Appearance = mode
			_ = s.ApplyPreset(p)
			r := ui.NewTester(s.View, size[0], size[1])
			for _, label := range []string{"Design system", "Cards", "Base color"} {
				b, ok := r.Find(label)
				if !ok || b.X < 0 || b.Y < 0 || b.X+b.W > float32(size[0])+1 || b.Y+b.H > float32(size[1])+1 {
					t.Fatalf("%v %s clipped %q %+v", size, mode, label, b)
				}
			}
			if r.Image().Bounds().Dx() != size[0] {
				t.Fatal("wrong responsive render size")
			}
		}
	}
}
