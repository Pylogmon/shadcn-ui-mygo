package desktop

import (
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type menuMaterialBackend struct{ surface *menuMaterialSurface }
type menuMaterialSurface struct {
	presenceSurface
	materialChanges int
}

func (b *menuMaterialBackend) Open(spec Spec, view func(*ui.Context), closed func()) (Surface, error) {
	b.surface = &menuMaterialSurface{presenceSurface: presenceSurface{spec: spec, view: view, onClosed: closed}}
	return b.surface, nil
}
func (b *menuMaterialBackend) InvalidateParent() {}
func (s *menuMaterialSurface) SetVibrancy(value mygo.Vibrancy) {
	s.spec.Vibrancy = value
	s.materialChanges++
}
func TestMenuPresetHostCopiesSnapshotRequestsMaterialAndKeepsLiveSurface(t *testing.T) {
	backend := &menuMaterialBackend{}
	host := NewHostWithBackend(backend)
	open := true
	state := MenuState{}
	colorMode := theme.MenuInvertedTranslucent
	var seen theme.MenuSnapshot
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		theme.MenuScope(c, theme.Preset("blue", false), theme.MenuOptions{Color: colorMode, Accent: theme.MenuBold, InvertedTokens: theme.Preset("blue", true)}).Children(func() {
			host.BeginFrame()
			defer host.EndFrame()
			anchor := control.Button(c, "Open preset", control.ButtonProps{})
			_, err := Menu(c, host, "preset", anchor, &open, &state, MenuProps{Items: []overlay.MenuItem{{Key: "more", Label: "More", Children: []overlay.MenuItem{{Key: "child", Label: "Child"}}}}, ItemContent: func(c *ui.Context, item overlay.MenuItem) { seen = theme.MenuOf(c); ui.Text(c, item.Label) }})
			if err != nil {
				panic(err)
			}
		})
	}, 400, 260)
	first := backend.surface
	if first.spec.Vibrancy != mygo.VibrancyMenu || first.spec.SystemPopover {
		t.Fatalf("did not request plain Flyout material: %+v", first.spec)
	}
	child := ui.NewTester(first.view, 400, 200)
	if !seen.Tokens.Dark || seen.Animated() || seen.Accent != theme.MenuSubtle {
		t.Fatalf("lost parent value snapshot: %+v", seen)
	}
	if _, ok := child.AccessByRole(ui.RoleMenuItem, "More"); !ok {
		t.Fatal("translucent Flyout did not expose items immediately")
	}
	child.Key(0, ui.KeyRight)
	if _, ok := child.AccessByRole(ui.RoleMenuItem, "Child"); !ok {
		t.Fatal("source no-animation profile did not reach submenu")
	}
	child.Key(0, ui.KeyLeft)
	if _, ok := child.AccessByRole(ui.RoleMenuItem, "Child"); ok {
		t.Fatal("translucent submenu retained an exit")
	}
	colorMode = theme.MenuInverted
	parent.Frame()
	child.Frame()
	if backend.surface != first || first.materialChanges != 1 || first.spec.Vibrancy != mygo.VibrancyNone {
		t.Fatal("material change remounted child instead of using live public capability")
	}
	colorMode = theme.MenuInvertedTranslucent
	parent.Frame()
	child.Frame()
	open = false
	parent.Frame()
	if first.closes != 1 || host.IsPresent("preset") {
		t.Fatalf("no-animation profile retained native exit: closes%d present%v", first.closes, host.IsPresent("preset"))
	}
}

func TestMenuMaterialTintAndSubmenuGapHaveNoOpaqueBacking(t *testing.T) {
	backend := &menuMaterialBackend{}
	host := NewHostWithBackend(backend)
	open := true
	state := MenuState{}
	ui.NewTester(func(c *ui.Context) {
		theme.MenuScope(c, theme.Preset("neutral", true), theme.MenuOptions{Color: theme.MenuDefaultTranslucent}).Children(func() {
			host.BeginFrame()
			defer host.EndFrame()
			anchor := control.Button(c, "Open", control.ButtonProps{})
			_, err := Menu(c, host, "material", anchor, &open, &state, MenuProps{Items: []overlay.MenuItem{{Key: "more", Label: "More", Children: []overlay.MenuItem{{Key: "child", Label: "Child"}}}}})
			if err != nil {
				t.Fatal(err)
			}
		})
	}, 400, 200)
	child := ui.NewTester(backend.surface.view, 400, 200)
	child.SetVibrancy(true)
	child.Key(0, ui.KeyRight)
	panel, ok := child.AccessByRole(ui.RoleMenu, "Menu")
	if !ok {
		t.Fatal("root menu missing")
	}
	x, y := int(panel.Bounds.X+panel.Bounds.W/2), int(panel.Bounds.Y+2)
	if alpha := child.Image().RGBAAt(x, y).A; alpha < 177 || alpha > 180 {
		t.Fatalf("70%% menu tint is covered by another background: alpha=%d", alpha)
	}
	if alpha := child.Image().RGBAAt(int(panel.Bounds.X+panel.Bounds.W+2), 20).A; alpha != 0 {
		t.Fatalf("submenu gap has an opaque theme wrapper: alpha=%d", alpha)
	}
	child.SetPreferences(ui.Preferences{ReduceTransparency: true})
	if alpha := child.Image().RGBAAt(x, y).A; alpha != 255 {
		t.Fatalf("reduced transparency lost its opaque fallback: alpha=%d", alpha)
	}
}
