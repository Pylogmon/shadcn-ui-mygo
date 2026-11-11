package desktop

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

func TestContextHostIsWindowLocalAndExplicitHostWins(t *testing.T) {
	leftBackend, rightBackend := &fakeBackend{}, &fakeBackend{}
	left, right := NewHostWithBackend(leftBackend), NewHostWithBackend(rightBackend)
	inherited, explicit := "one", "one"
	bound := left
	menuOpen := true
	menuState := MenuState{}
	parent := ui.NewTesterWith(func(c *ui.Context) {
		BindHost(c, bound)
		if HostOf(c) != bound {
			t.Fatal("Context binding was lost")
		}
		left.BeginFrame()
		right.BeginFrame()
		defer left.EndFrame()
		defer right.EndFrame()
		ui.Column(c).Gap(8).Children(func() {
			if _, err := Select(c, nil, "inherited", &inherited, control.SelectProps{Label: "Inherited", Options: []control.Option{{Value: "one"}, {Value: "two"}}}); err != nil {
				t.Fatal(err)
			}
			if _, err := Select(c, right, "explicit", &explicit, control.SelectProps{Label: "Explicit", Options: []control.Option{{Value: "one"}, {Value: "two"}}}); err != nil {
				t.Fatal(err)
			}
			anchor := control.Button(c, "Hint", control.ButtonProps{})
			if err := Tooltip(c, nil, "hint", anchor, "Context tooltip", TooltipProps{HoverOnly: true}); err != nil {
				t.Fatal(err)
			}
			anchor = control.Button(c, "Actions", control.ButtonProps{})
			if _, err := Menu(c, nil, "actions", anchor, &menuOpen, &menuState, MenuProps{Items: []overlay.MenuItem{{Key: "save", Label: "Save"}}}); err != nil {
				t.Fatal(err)
			}
		})
	}, 400, 360, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !left.IsOpen("actions") || right.IsOpen("actions") {
		t.Fatal("Menu did not inherit the context host")
	}
	if err := parent.Click("Inherited"); err != nil {
		t.Fatal(err)
	}
	if !left.IsOpen("inherited") || right.IsOpen("inherited") {
		t.Fatal("Select did not inherit the context host")
	}
	if err := parent.Click("Explicit"); err != nil {
		t.Fatal(err)
	}
	if !right.IsOpen("explicit") || left.IsOpen("explicit") {
		t.Fatal("explicit host did not override the context")
	}
	hint, _ := parent.Find("Hint")
	parent.Move(hint.X+hint.W/2, hint.Y+hint.H/2)
	parent.Advance(time.Second)
	if !left.IsOpen("hint") || right.IsOpen("hint") {
		t.Fatal("Tooltip did not inherit the context host")
	}
	// Even an independently rendered child must not inherit its owner's Host:
	// its anchors use a different window's coordinates and lifetime.
	ui.NewTester(func(c *ui.Context) {
		leftBackend.surfaces[0].view(c)
		if HostOf(c) != nil {
			t.Fatal("owner binding leaked into an independent native child")
		}
	}, 280, 160)
	ui.NewTester(func(c *ui.Context) {
		if HostOf(c) != nil {
			t.Fatal("host binding leaked into a second window")
		}
		BindHost(c, right)
		if HostOf(c) != right {
			t.Fatal("second window could not bind its own host")
		}
		BindHost(c, nil)
		if HostOf(c) != nil {
			t.Fatal("nil did not clear the window binding")
		}
	}, 200, 100)
	parent.Frame()
	if !left.IsOpen("actions") || HostOf(nil) != nil {
		t.Fatal("another window altered the first binding")
	}
}
