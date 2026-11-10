package desktop_test

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// One independent-window flow verifies per-menu sizing, child events before a
// parent reanchor, and the same native editor surviving a width-only change.
// The injected backend exposes SetSize; it does not simulate an OS fitter.
func TestMenubarPerMenuWidthIndependentWindowWorkflow(t *testing.T) {
	b := &itemHeightBackend{}
	h := desktop.NewHostWithBackend(b)
	state := desktop.MenubarState{}
	rtl, native := false, false
	note, chosen := "", ""
	mounted := 0
	selection := [2]int{}
	var editorBounds ui.Rect
	var menuErr error
	p := desktop.MenubarProps{Width: 224, KeepOpenOnSelect: true, Items: []overlay.MenubarItem{
		{Key: "file", Label: "File", Width: 260, Items: []overlay.MenuItem{{Key: "open", Label: "Open file"}}},
		{Key: "view", Label: "View", Width: 176, Items: []overlay.MenuItem{{Key: "note", Label: "Note"}}},
		{Key: "tools", Label: "Tools", Items: []overlay.MenuItem{{Key: "find", Label: "Find"}}},
	}}
	p.ItemContent = func(c *ui.Context, item overlay.MenuItem) {
		if item.Key == "note" {
			editor := ui.TextInputBase(c, &note).Label("Menu note").FillWidth().Height(24).Basis(24).Grow(0)
			ui.Local(editor, "width-editor-mount", func() bool { mounted++; return true })
			editorBounds = editor.LayoutBounds()
			selection[0], selection[1] = editor.TextSelection()
		} else {
			ui.Text(c, item.Label).SingleLine()
		}
	}
	parent := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c).Padding(20).Direction(direction).Children(func() {
			var value string
			if native {
				value, menuErr = desktop.NativeMenubar(c, p.Items)
			} else {
				value, menuErr = desktop.Menubar(c, h, "width-bar", &state, p)
			}
			if value != "" {
				chosen = value
			}
		})
	}, 500, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := parent.Click("File"); err != nil {
		t.Fatal(err)
	}
	if menuErr != nil || len(b.surfaces) != 1 || !state.Open {
		t.Fatalf("independent menu did not open: %v", menuErr)
	}
	surface := b.surfaces[0]
	child := surface.tester
	panelWidth := func(label string, want float32) {
		t.Helper()
		a, ok := child.AccessByRole(ui.RoleMenu, label)
		if !ok || math.Abs(float64(a.Bounds.W-want)) > .1 {
			t.Fatalf("%s panel width=%v want%.1f", label, a.Bounds, want)
		}
	}
	panelWidth("File", 260)
	if surface.spec.Width != 260 {
		t.Fatalf("initial native width=%d want260", surface.spec.Width)
	}
	child.Key(0, ui.KeyRight)
	panelWidth("View", 176)
	if state.Active != "view" || surface.spec.Width != 260 {
		t.Fatal("child sibling event required or fabricated a parent resize frame")
	}
	parent.Frame()
	if len(b.surfaces) != 1 || surface.closes != 0 || surface.spec.Width != 176 {
		t.Fatal("sibling reanchor replaced the surface or ignored its width")
	}
	child.Advance(150 * time.Millisecond)
	viewMounts := mounted
	field, ok := child.Find("Menu note")
	if !ok {
		t.Fatalf("custom content did not construct its actual editor: mounted=%d bounds=%v", mounted, editorBounds)
	}
	child.ClickAt(field.X+field.W/2, field.Y+field.H/2)
	child.Type("native-width")
	child.Key(ui.Cmd, ui.KeyA)
	child.Frame()
	p.Items[1].Width = 192
	parent.Frame()
	if surface.spec.Width != 192 || len(b.surfaces) != 1 || surface.closes != 0 || mounted != viewMounts || note != "native-width" || !child.Focused("Menu note") || selection != [2]int{0, len(note)} {
		t.Fatalf("width refresh lost editor identity or selection: spec=%d mounts=%d note=%q selection=%v", surface.spec.Width, mounted, note, selection)
	}
	panelWidth("View", 192)
	// Leave the editor using the menu's public focus target, then switch to an
	// unspecified menu whose width comes from the bar's existing fallback.
	child.Key(0, ui.KeyEscape)
	parent.Frame()
	if err := parent.Click("Tools"); err != nil {
		t.Fatal(err)
	}
	surface = b.surfaces[len(b.surfaces)-1]
	child = surface.tester
	panelWidth("Tools", 224)
	p.Width = 0
	parent.Frame()
	panelWidth("Tools", 144)
	if surface.spec.Width != 144 {
		t.Fatal("the old default144 width was not retained")
	}
	p.Items[2].Width, p.Width = -1, 208
	parent.Frame()
	panelWidth("Tools", 208)
	// Direction changes mirror sibling navigation without changing the width
	// contract. The child constructs View at192 before parent refreshes Tools.
	rtl = true
	parent.Frame()
	child.Key(0, ui.KeyRight)
	panelWidth("View", 192)
	parent.Frame()
	if surface.spec.Width != 192 || state.Active != "view" {
		t.Fatal("RTL sibling sizing or parent resize ignored the current menu")
	}
	// Every candidate must be finite, including an inactive sibling. Rejecting
	// invalid dimensions closes the existing surface safely, without a new one.
	count := len(b.surfaces)
	for _, invalid := range []float32{float32(math.NaN()), float32(math.Inf(1))} {
		p.Items[0].Width = invalid
		parent.Frame()
		if menuErr == nil || state.Open || surface.closes != 1 || len(b.surfaces) != count {
			t.Fatal("non-finite candidate width was accepted or reopened a surface")
		}
	}
	p.Items[0].Width, p.Width = 260, float32(math.NaN())
	parent.Frame()
	if menuErr == nil {
		t.Fatal("an item override hid an invalid bar fallback")
	}
	// Native menu tracking ignores styled dimensions, even non-finite visual
	// overrides; it still delivers the platform menu's actual selected action.
	native = true
	p.Items[0].Width = float32(math.NaN())
	parent.Frame()
	if menuErr != nil {
		t.Fatal(menuErr)
	}
	if err := parent.Click("File"); err != nil {
		t.Fatal(err)
	}
	if err := parent.ChooseMenuItem("Open file"); err != nil {
		t.Fatal(err)
	}
	parent.Frame()
	if chosen != "open" || len(b.surfaces) != count {
		t.Fatal("native menu did not retain system tracking and action delivery")
	}
}
