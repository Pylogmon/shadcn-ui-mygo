package desktop

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestNumberInputStepperUsesBoundHostTooltipAndClearsBinding(t *testing.T) {
	backend, otherBackend := &fakeBackend{}, &fakeBackend{}
	host, otherHost := NewHostWithBackend(backend), NewHostWithBackend(otherBackend)
	bound, mounted := host, true
	n := 3.
	parent := ui.NewTesterWith(func(c *ui.Context) {
		c.SetTheme(theme.Preset("orange", false).Theme())
		BindHost(c, bound)
		host.BeginFrame()
		defer host.EndFrame()
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			if mounted {
				control.NumberInput(c, &n, control.NumberInputProps{Label: "Width", Stepper: &control.NumberInputStepperProps{
					Increment: control.ButtonProps{Label: "Wider", Variant: control.Ghost, Size: control.IconExtraSmall}, IncrementTooltip: "Increase stroke width",
				}}).Width(180)
			}
			control.Button(c, "Outside", control.ButtonProps{})
		})
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	hover := func(r *ui.Tester, name string) {
		t.Helper()
		a, ok := r.AccessByRole(ui.RoleButton, name)
		if !ok {
			t.Fatal("missing tooltip trigger", name)
		}
		r.Move(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	}
	hover(parent, "Wider")
	if len(backend.surfaces) != 1 || len(otherBackend.surfaces) != 0 {
		t.Fatal("stepper did not immediately use its own window's native tooltip host")
	}
	surface := backend.surfaces[0]
	if surface.spec.Focusable || !surface.spec.IgnoreMouseEvents || !surface.spec.ShowInactive {
		t.Fatalf("stepper tooltip does not share desktop tooltip input behavior: %+v", surface.spec)
	}
	child := ui.NewTesterWith(surface.view, 220, 80, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if _, ok := child.AccessByRole(ui.RoleTooltip, "Increase stroke width"); !ok || parent.HasText("Increase stroke width") {
		t.Fatal("stepper did not render exactly one tooltip in the independent child")
	}
	// The inverse tooltip surface must follow desktop.Tooltip, even when the
	// owner's accent is orange. It must not use the old Primary-colored panel.
	tokens := theme.Preset("orange", false)
	pixel := child.Image().RGBAAt(15, 15)
	if pixel.R != tokens.Foreground.R || pixel.G != tokens.Foreground.G || pixel.B != tokens.Foreground.B {
		t.Fatalf("stepper tooltip kept a separate accent-colored appearance: %v", pixel)
	}
	if err := parent.Click("Wider"); err != nil {
		t.Fatal(err)
	}
	if n != 4 || !parent.Focused("Width") || surface.closes != 1 {
		t.Fatal("step click failed to close tooltip, commit or restore editor focus")
	}
	parent.Move(1, 1)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyTab)
	if !parent.Focused("Wider") || len(backend.surfaces) != 2 {
		t.Fatal("keyboard focus did not open the same native tooltip route")
	}
	parent.Key(0, ui.KeyEscape)
	parent.Frame()
	if backend.surfaces[1].closes != 1 || len(backend.surfaces) != 2 {
		t.Fatal("Escape failed to close and suppress the stepper tooltip")
	}
	if err := parent.Click("Outside"); err != nil {
		t.Fatal(err)
	}
	hover(parent, "Wider")
	if len(backend.surfaces) != 3 {
		t.Fatal("tooltip did not reopen after leaving its trigger")
	}
	mounted = false
	parent.Frame()
	if backend.surfaces[2].closes != 1 || len(host.entries) != 0 {
		t.Fatal("unmounted number input retained a native tooltip or callback")
	}
	// A separate window owns its own binding. Binding it must not redirect
	// the first window's compound-control hints.
	other := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		BindHost(c, otherHost)
		otherHost.BeginFrame()
		defer otherHost.EndFrame()
		control.NumberInput(c, &n, control.NumberInputProps{Label: "Other", Stepper: &control.NumberInputStepperProps{
			Increment: control.ButtonProps{Label: "Other wider"}, IncrementTooltip: "Other hint",
		}})
	}, 240, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	hover(other, "Other wider")
	if len(otherBackend.surfaces) != 1 || len(backend.surfaces) != 3 {
		t.Fatal("compound tooltip host leaked between windows")
	}
	bound, mounted = nil, true
	parent.Frame()
	hover(parent, "Wider")
	parent.Advance(time.Second)
	if !parent.HasText("Increase stroke width") || len(backend.surfaces) != 3 {
		t.Fatal("clearing BindHost did not restore the in-window fallback")
	}
}

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
