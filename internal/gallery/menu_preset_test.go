package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"testing"
)

func TestCreatorMenuPresetsNormalizeExportUndoAndRestoreSolidAccent(t *testing.T) {
	s := New()
	s.Config.NativeMenus = true
	r := ui.NewTester(s.View, 1280, 1000)
	choose := func(label, item string) {
		t.Helper()
		info, ok := r.AccessByRole(ui.RoleMenuButton, label)
		if !ok {
			t.Fatalf("missing creator menu %s", label)
		}
		r.ClickAt(info.Bounds.X+8, info.Bounds.Y+info.Bounds.H/2)
		if err := r.ChooseMenuItem(item); err != nil {
			t.Fatal(err)
		}
	}
	choose("Menu accent", "Bold")
	if s.Config.MenuAccent != theme.MenuBold || s.Config.Tokens(false).Accent != s.Config.Tokens(false).Primary {
		t.Fatal("bold source theme transformation missing")
	}
	choose("Menu color", "Default translucent")
	if s.Config.MenuColor != theme.MenuDefaultTranslucent || s.Config.MenuAccent != theme.MenuSubtle {
		t.Fatalf("translucent normalization=%+v", s.Config)
	}
	parsed, err := ParsePreset(s.ExportPreset())
	if err != nil || parsed != s.Config {
		t.Fatalf("new fields did not round trip: %v", err)
	}
	choose("Menu color", "Inverted")
	if s.Config.MenuAccent != theme.MenuBold || s.Config.MenuColor != theme.MenuInverted {
		t.Fatal("return to solid did not restore preferred bold accent")
	}
	if !s.Undo() || s.Config.MenuColor != theme.MenuDefaultTranslucent || s.Config.MenuAccent != theme.MenuSubtle {
		t.Fatal("undo lost effective normalized configuration")
	}
	if !s.Redo() || s.Config.MenuColor != theme.MenuInverted || s.Config.MenuAccent != theme.MenuBold {
		t.Fatal("redo lost solid preference")
	}
}
