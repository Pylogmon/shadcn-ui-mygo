package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
)

type exampleSurface struct{ view func(*ui.Context) }

func (*exampleSurface) SetAnchor(ui.Rect) {}
func (*exampleSurface) Invalidate()       {}
func (*exampleSurface) Close()            {}

type exampleBackend struct{ surfaces []*exampleSurface }

func (b *exampleBackend) Open(_ desktop.Spec, view func(*ui.Context), _ func()) (desktop.Surface, error) {
	s := &exampleSurface{view}
	b.surfaces = append(b.surfaces, s)
	return s, nil
}
func (*exampleBackend) InvalidateParent() {}

func TestExampleDirectNativeSelectionAndOptionalQuantity(t *testing.T) {
	a := newApplication()
	b := &exampleBackend{}
	a.host = desktop.NewHostWithBackend(b)
	r := ui.NewTesterWith(a.view, 820, 760, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Plan"); err != nil {
		t.Fatal(err)
	}
	child := ui.NewTesterWith(b.surfaces[0].view, 300, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	if a.plan != "team" || a.status != "Selected: Team subscription" {
		t.Fatal("example did not update business state directly from the child")
	}
	r.Frame()
	if err := r.Click("Increase quantity"); err != nil {
		t.Fatal(err)
	}
	if a.quantity != 2.5 {
		t.Fatal("example's icon stepper is inert")
	}
	a.readOnly = true
	r.Frame()
	if button, ok := r.AccessByRole(ui.RoleButton, "Increase quantity"); !ok || !button.Disabled {
		t.Fatal("example did not demonstrate readonly boundary behavior")
	}
	if err := r.Click("Actions"); err != nil {
		t.Fatal(err)
	}
	child = ui.NewTesterWith(b.surfaces[len(b.surfaces)-1].view, 300, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	child.Key(0, ui.KeyEnter)
	r.Frame()
	if !a.empty || a.quantity != 2.5 {
		t.Fatal("application clear action did not retain optional numeric storage")
	}
}
