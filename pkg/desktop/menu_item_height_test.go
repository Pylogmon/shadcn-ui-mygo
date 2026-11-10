package desktop_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type itemHeightBackend struct {
	surfaces []*itemHeightSurface
	sizes    int
}

type itemHeightSurface struct {
	backend     *itemHeightBackend
	spec        desktop.Spec
	tester      *ui.Tester
	closed      func()
	closes      int
	invalidates int
}

func (b *itemHeightBackend) Open(spec desktop.Spec, view func(*ui.Context), closed func()) (desktop.Surface, error) {
	s := &itemHeightSurface{backend: b, spec: spec, closed: closed}
	b.surfaces = append(b.surfaces, s)
	s.tester = ui.NewTesterWith(view, spec.Width, spec.Height, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	return s, nil
}
func (b *itemHeightBackend) InvalidateParent()   {}
func (s *itemHeightSurface) SetAnchor(r ui.Rect) { s.spec.Anchor = r }
func (s *itemHeightSurface) Invalidate()         { s.invalidates++; s.tester.Frame() }
func (s *itemHeightSurface) SetSize(w, h int) {
	s.backend.sizes++
	s.spec.Width, s.spec.Height = w, h
	s.tester.SetSize(w, h)
}
func (s *itemHeightSurface) Close() {
	if s.closes == 0 {
		s.closes++
		s.closed()
	}
}

// One independent-window workflow exercises rich rows, a live height change,
// long-list reveal and one accepted selection. The native OS menu path is not
// involved: its row metrics remain owned by the operating system.
func TestMenuCustomItemHeightIndependentWindowWorkflow(t *testing.T) {
	b := &itemHeightBackend{}
	h := desktop.NewHostWithBackend(b)
	state, open := desktop.MenuState{}, false
	itemHeight := float32(48)
	revision := uint64(0)
	items := []overlay.MenuItem{{Key: "one", Label: "Person one"}, {Key: "two", Label: "Person two"}, {Key: "three", Label: "Person three"}}
	chosen, returns, mounted := "", 0, 0
	parent := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		anchor := ui.Button(c, "People menu")
		value, err := desktop.Menu(c, h, "people", anchor, &open, &state, desktop.MenuProps{
			Label: "People", Width: 260, MaxHeight: 224, ItemHeight: itemHeight, Revision: revision, Items: items,
			ItemContent: func(child *ui.Context, item overlay.MenuItem) {
				ui.Local(child.Root(), "height-workflow-mounted", func() bool { mounted++; return true })
				ui.Row(child).FillWidth().MinWidth(0).Padding(6, 0).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Box(child).Size(26, 26).Shrink(0).Background(ui.RGB(80, 100, 140)).AccessHidden(true)
					ui.Column(child).Grow(1).MinWidth(0).Children(func() {
						ui.Text(child, item.Label).FontSize(14)
						ui.Text(child, item.Key+"@example.invalid").FontSize(12)
					})
				})
			},
			OnChoose: func(key string) { chosen = key },
		})
		if err != nil {
			t.Fatal(err)
		}
		if value != "" {
			returns++
		}
	}, 340, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := parent.Click("People menu"); err != nil {
		t.Fatal(err)
	}
	if len(b.surfaces) != 1 || !open || parent.HasText("Person one") {
		t.Fatal("custom menu was not built in an independent child")
	}
	surface := b.surfaces[0]
	child := surface.tester
	access := func(role ui.Role, label string) ui.AccessInfo {
		t.Helper()
		a, ok := child.AccessByRole(role, label)
		if !ok {
			t.Fatalf("missing %q with role %v", label, role)
		}
		return a
	}
	checkRows := func(height float32) {
		t.Helper()
		var previous ui.Rect
		for _, item := range items {
			if item.Separator {
				continue
			}
			row := access(ui.RoleMenuItem, item.Label).Bounds
			if math.Abs(float64(row.H-height)) > .1 || previous.H > 0 && row.Y < previous.Y+previous.H-.1 {
				t.Fatalf("rich rows overlap or ignore height %.1f: previous=%v row=%v", height, previous, row)
			}
			caption, found := child.Find(item.Key + "@example.invalid")
			if !found || caption.Y < row.Y || caption.Y+caption.H > row.Y+row.H+.1 {
				t.Fatalf("rich caption overflows its row: %v in %v", caption, row)
			}
			previous = row
		}
	}
	checkRows(48)
	if surface.spec.Height != 154 || mounted != 1 {
		t.Fatalf("custom root budget or initial child identity wrong: height=%d mounts=%d", surface.spec.Height, mounted)
	}
	itemHeight = 64
	parent.Frame()
	checkRows(64)
	if len(b.surfaces) != 1 || b.sizes != 1 || surface.spec.Height != 202 || mounted != 1 {
		t.Fatalf("live row height recreated or failed to resize its child: surfaces=%d sizes=%d height=%d mounts=%d", len(b.surfaces), b.sizes, surface.spec.Height, mounted)
	}
	parent.Frame()
	if b.sizes != 1 {
		t.Fatal("stable row height scheduled another resize")
	}
	items = make([]overlay.MenuItem, 14)
	for i := range items {
		items[i] = overlay.MenuItem{Key: fmt.Sprintf("person-%02d", i), Label: fmt.Sprintf("Person %02d", i)}
	}
	items[0].Group = "Members"
	items[4] = overlay.MenuItem{Key: "divider", Separator: true}
	revision++
	parent.Frame()
	panel := access(ui.RoleMenu, "People").Bounds
	if panel.H > 224.1 || surface.spec.Height > 224 {
		t.Fatalf("long custom menu exceeds its outer height budget: panel=%v spec=%+v", panel, surface.spec)
	}
	child.Key(0, ui.KeyEnd)
	last := access(ui.RoleMenuItem, "Person 13")
	panel = access(ui.RoleMenu, "People").Bounds
	if state.Highlighted[0] != "person-13" || last.Offscreen || last.Bounds.H != 64 || last.Bounds.Y+last.Bounds.H > panel.Y+panel.H+.1 {
		t.Fatalf("End failed to reveal the actual 64DIP last row: %v in %v, highlight=%v", last, panel, state.Highlighted)
	}
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	parent.Frame()
	if chosen != "person-13" || returns != 1 || open || h.IsOpen("people") || !parent.Focused("People menu") {
		t.Fatalf("custom row choice did not close/return once: chosen=%q returns=%d open=%v", chosen, returns, open)
	}
	// Invalid optional dimensions preserve the existing menu's default metrics.
	itemHeight = float32(math.NaN())
	items = []overlay.MenuItem{{Key: "fallback", Label: "Default row"}}
	revision++
	if err := parent.Click("People menu"); err != nil {
		t.Fatal(err)
	}
	child = b.surfaces[len(b.surfaces)-1].tester
	if row := access(ui.RoleMenuItem, "Default row"); row.Bounds.H != 32 {
		t.Fatalf("non-finite optional height did not keep default metrics: %v", row)
	}
	// Namespaces belong to the explicit menu key, even when two menu owners
	// share one anchor. Different row sizes must not fight over one revision.
	sharedBackend := &itemHeightBackend{}
	sharedHost := desktop.NewHostWithBackend(sharedBackend)
	firstOpen, secondOpen := true, true
	firstState, secondState := desktop.MenuState{}, desktop.MenuState{}
	shared := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		sharedHost.BeginFrame()
		defer sharedHost.EndFrame()
		anchor := ui.Button(c, "Shared anchor")
		for _, owner := range []struct {
			key    string
			height float32
			open   *bool
			state  *desktop.MenuState
		}{{"first", 48, &firstOpen, &firstState}, {"second", 64, &secondOpen, &secondState}} {
			if _, err := desktop.Menu(c, sharedHost, owner.key, anchor, owner.open, owner.state, desktop.MenuProps{
				Label: owner.key, Items: []overlay.MenuItem{{Key: "one", Label: "Shared choice"}}, ItemHeight: owner.height,
			}); err != nil {
				t.Fatal(err)
			}
		}
	}, 340, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	shared.Frame()
	if len(sharedBackend.surfaces) != 2 {
		t.Fatal("shared-anchor owners did not open independently")
	}
	before := []int{sharedBackend.surfaces[0].invalidates, sharedBackend.surfaces[1].invalidates}
	shared.Frame()
	shared.Frame()
	if sharedBackend.surfaces[0].invalidates != before[0] || sharedBackend.surfaces[1].invalidates != before[1] {
		t.Fatal("different menu keys on one anchor continuously refreshed each other's item height")
	}
}
