package desktop

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestFlyoutSelectPreservesScrollBeforeChoosingFont(t *testing.T) {
	backend := &fakeBackend{}
	host := NewHostWithBackend(backend)
	defer host.CloseAll()
	options := []control.Option{{Value: "", Label: "System default"}}
	for i := range 100 {
		options = append(options, control.Option{Value: fmt.Sprint(i), Label: fmt.Sprintf("Font %02d", i)})
	}
	selected, changes := "", 0
	parent := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		host.BeginFrame()
		defer host.EndFrame()
		if _, err := Select(c, host, "font", &selected, control.SelectProps{Label: "Application font", Options: options, OnChange: func(string) { changes++ }}); err != nil {
			t.Fatal(err)
		}
	}, 340, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := parent.Click("Application font"); err != nil {
		t.Fatal(err)
	}
	if len(backend.surfaces) != 1 {
		t.Fatalf("font selector opened %d surfaces", len(backend.surfaces))
	}
	child := ui.NewTesterWith(backend.surfaces[0].view, 260, 480, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, label string) ui.AccessInfo {
		t.Helper()
		a, ok := child.AccessByRole(role, label)
		if !ok {
			t.Fatalf("missing %q with role %v", label, role)
		}
		return a
	}
	child.Key(0, ui.KeyEnd)
	panel := access(ui.RoleListBox, "Application font").Bounds
	x, y := panel.X+panel.W/2, panel.Y+panel.H/2
	child.Move(x, y)
	child.Scroll(x, y, 0, -640)
	before := access(ui.RoleOption, "System default").Bounds.Y
	child.Move(x+1, y)
	// Rebuilding options in the parent must also keep this child and position.
	parent.Frame()
	child.Frame()
	if after := access(ui.RoleOption, "System default").Bounds.Y; math.Abs(float64(after-before)) > .1 {
		t.Fatalf("font list jumped after scrolling and hovering: first row y=%v, want %v", after, before)
	}
	if selected != "" || changes != 0 || !host.IsOpen("font") || len(backend.surfaces) != 1 {
		t.Fatal("browsing the font list changed the value or replaced the popup")
	}
	var target ui.AccessInfo
	for _, node := range child.AccessAll() {
		if node.Role == ui.RoleOption && !node.Offscreen && node.Bounds.Y <= y && node.Bounds.Y+node.Bounds.H > y {
			target = node
			break
		}
	}
	if target.Name == "" || target.Name == "System default" {
		t.Fatal("no font under the pointer after scrolling")
	}
	want := ""
	for _, option := range options {
		if option.Label == target.Name {
			want = option.Value
			break
		}
	}
	child.ClickAt(x+1, y)
	parent.Frame()
	if changes != 1 || host.IsOpen("font") || selected != want {
		t.Fatalf("scrolled font was not committed once: target=%q selected=%q changes=%d", target.Name, selected, changes)
	}
}
