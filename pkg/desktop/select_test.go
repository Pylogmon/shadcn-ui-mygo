package desktop

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestFlyoutSelectCommitsFromIndependentWindow(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	selected := "one"
	changes := 0
	callbacks := 0
	disabled := false
	var selectErr error
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(16).Children(func() {
			var anchor ui.Element
			anchor, selectErr = Select(c, h, "plan", &selected, control.SelectProps{Label: "Plan", Disabled: disabled, Options: []control.Option{{Value: "one", Label: "One"}, {Value: "blocked", Label: "Unavailable", Disabled: true}, {Value: "two", Label: "Two"}}, ItemContent: func(child *ui.Context, o control.Option) { ui.Text(child, "Custom "+o.Label) }, OnChange: func(value string) {
				if selected != value {
					t.Fatal("OnChange ran before updating selected storage")
				}
				callbacks++
			}})
			if anchor.Changed() {
				changes++
			}
		})
	}, 340, 180)
	if selectErr != nil {
		t.Fatal(selectErr)
	}
	if err := parent.Click("Plan"); err != nil {
		t.Fatal(err)
	}
	if selectErr != nil || len(b.surfaces) != 1 {
		t.Fatalf("surface=%d err=%v", len(b.surfaces), selectErr)
	}
	if parent.HasText("Custom Two") {
		t.Fatal("Flyout options painted into the parent tree")
	}
	child := ui.NewTester(b.surfaces[0].view, 260, 230)
	if !child.HasText("Custom Two") {
		t.Fatal("custom option renderer was dropped")
	}
	if a, ok := parent.AccessByRole(ui.RoleComboBox, "Plan"); !ok || a.HasPopup != ui.PopupListBox || !a.Expanded {
		t.Fatalf("select trigger did not own listbox semantics: %+v", a)
	}
	if _, ok := child.AccessByRole(ui.RoleListBox, "Plan"); !ok {
		t.Fatal("select popup did not own listbox semantics")
	}
	if a, ok := child.AccessByRole(ui.RoleOption, "One"); !ok || !a.Selected {
		t.Fatalf("selected option semantics missing: %+v", a)
	}
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	if selected != "two" || callbacks != 1 || changes != 0 {
		t.Fatalf("async selection did not notify before parent rebuild: selected=%q callbacks=%d changes=%d", selected, callbacks, changes)
	}
	parent.Frame()
	if selected != "two" || changes != 1 {
		t.Fatalf("selected=%q change events=%d", selected, changes)
	}
	if h.IsOpen("plan") {
		t.Fatal("select stayed open after commit")
	}
	parent.Frame()
	if changes != 1 || callbacks != 1 {
		t.Fatal("selection change was emitted twice")
	}
	disabled = true
	parent.Frame()
	if err := parent.Click("Plan"); err == nil && h.IsOpen("plan") {
		t.Fatal("disabled selector opened")
	}
}

func TestFlyoutSelectDoesNotSilentlyFallbackWithoutHost(t *testing.T) {
	selected := "one"
	var err error
	sawOpenError := false
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		_, err = Select(c, nil, "no-host", &selected, control.SelectProps{Label: "Choice", Options: []control.Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}}})
		if err != nil {
			sawOpenError = true
		}
	}, 300, 150)
	if err != nil {
		t.Fatalf("closed selector needs no native window: %v", err)
	}
	if clickErr := r.Click("Choice"); clickErr != nil {
		t.Fatal(clickErr)
	}
	if !sawOpenError {
		t.Fatal("missing parent Host did not report an open error")
	}
	if r.HasText("Two") {
		t.Fatal("missing Host silently drew in-window options")
	}
}
