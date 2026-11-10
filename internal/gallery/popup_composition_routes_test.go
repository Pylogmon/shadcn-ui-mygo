package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
)

func TestCreatorPopupCompositionChildEditingAndPageReturn(t *testing.T) {
	s := New()
	s.Config.NativeMenus, s.Config.ReduceMotion = true, true
	b := &creatorPopupBackend{}
	s.Popovers = desktop.NewHostWithBackend(b)
	r := ui.NewTester(s.View, 500, 900)
	r.SetDirection(ui.RightToLeft)
	navigate := func(query, page string) {
		t.Helper()
		r.Key(ui.Cmd, ui.KeyK)
		r.Frame()
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(query)
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if s.Page != page || s.Command {
			t.Fatalf("component search %q reached %q instead of %q", query, s.Page, page)
		}
	}
	open := func() *creatorPopupSurface {
		t.Helper()
		a := revealPreviewPart(t, r, "Overlays", ui.RoleButton, "Open dimensions popover")
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
		if len(b.surfaces) == 0 {
			t.Fatal("the source Popover did not open an independent window")
		}
		surface := b.surfaces[len(b.surfaces)-1]
		if surface.didClose || !surface.spec.SystemPopover {
			t.Fatal("the global Native setting did not reach the source Popover")
		}
		return surface
	}
	navigate("Popover", "overlays")
	surface := open()
	if _, found := r.AccessByRole(ui.RoleTextField, "Width"); found {
		t.Fatal("the source child editor leaked into the parent window")
	}
	field, ok := surface.tester.AccessByRole(ui.RoleTextField, "Width")
	if !ok {
		t.Fatal("the native child did not construct its actual Width editor")
	}
	surface.tester.ClickAt(field.Bounds.X+field.Bounds.W/2, field.Bounds.Y+field.Bounds.H/2)
	surface.tester.Key(ui.Cmd, ui.KeyA)
	surface.tester.Type("80%")
	r.Frame()
	if s.popupCompositions.dimensions[0] != "80%" || s.popupCompositions.popupError != "" {
		t.Fatal("the actual child edit did not reach the source application model")
	}
	// Inject the backend's native-close notification; this is not a claim about
	// NSPopover tracking or a real operating-system Escape key path.
	surface.Close()
	r.Frame()
	navigate("Item", "content")
	navigate("Popover", "overlays")
	surface = open()
	field, ok = surface.tester.AccessByRole(ui.RoleTextField, "Width")
	if !ok || field.Value != "80%" || s.popupCompositions.dimensions[0] != "80%" {
		t.Fatal("the independent source draft did not survive page navigation and reopening")
	}
	surface.Close()
	r.Frame()
}
