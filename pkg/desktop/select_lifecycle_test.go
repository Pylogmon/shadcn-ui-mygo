package desktop

import (
	"errors"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
)

func TestSelectHostAndKeyTransferRetiresAnimatedOwnerImmediately(t *testing.T) {
	leftBackend, rightBackend := &presenceBackend{}, &presenceBackend{}
	left, right := NewHostWithBackend(leftBackend), NewHostWithBackend(rightBackend)
	host, key := left, "first"
	selected, callbacks := "one", 0
	parent := ui.NewTester(func(c *ui.Context) {
		BindHost(c, host)
		// Explicit rebinding cleanup must also work without a frame sweep.
		if _, err := Select(c, nil, key, &selected, control.SelectProps{Label: "Choice", Options: []control.Option{{Value: "one"}, {Value: "two"}}, OnChange: func(string) { callbacks++ }}); err != nil {
			t.Fatal(err)
		}
	}, 360, 180)
	if err := parent.Click("Choice"); err != nil {
		t.Fatal(err)
	}
	first := leftBackend.surfaces[0]
	child := ui.NewTester(first.view, 280, 200)
	child.Advance(time.Second)
	parent.Frame()
	host, key = right, "second"
	parent.Frame()
	if left.IsPresent("first") || first.closes != 1 {
		t.Fatal("rebound Select retained its obsolete animated surface")
	}
	if err := parent.Click("Choice"); err != nil {
		t.Fatal(err)
	}
	first.onClosed()
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	if !right.IsOpen("second") || callbacks != 0 || selected != "one" {
		t.Fatal("old native owner changed the new host, key or selection")
	}
	right.CloseAll()
}

func TestSelectContextHostRebindingAndStaleNativeEvents(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	first, second := "one", "one"
	selected := &first
	visible := true
	changes, callbacks, currentCallback := 0, 0, 1
	var callbackVersion int
	parent := ui.NewTesterWith(func(c *ui.Context) {
		BindHost(c, h)
		h.BeginFrame()
		defer h.EndFrame()
		if visible {
			version := currentCallback
			e, err := Select(c, nil, "choice", selected, control.SelectProps{Label: "Choice", Options: []control.Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}}, OnChange: func(string) {
				callbacks++
				callbackVersion = version
			}})
			if err != nil {
				t.Fatal(err)
			}
			if e.Changed() {
				changes++
			}
		}
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	open := func() (*fakeSurface, *ui.Tester) {
		t.Helper()
		if err := parent.Click("Choice"); err != nil {
			t.Fatal(err)
		}
		if !h.IsOpen("choice") {
			t.Fatal("context-bound select did not open")
		}
		surface := b.surfaces[len(b.surfaces)-1]
		return surface, ui.NewTesterWith(surface.view, 280, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	}
	stale, child := open()
	child.Key(0, ui.KeyEscape)
	parent.Frame()
	if first != "one" || callbacks != 0 || changes != 0 || h.IsOpen("choice") {
		t.Fatal("Escape fabricated a selection change")
	}
	_, child = open()
	stale.dismiss()
	if !h.IsOpen("choice") {
		t.Fatal("late native close dismissed a replacement")
	}
	currentCallback = 2
	parent.Frame()
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	if first != "two" || callbacks != 1 || callbackVersion != 2 {
		t.Fatal("child used a stale callback or missed immediate notification")
	}
	parent.Frame()
	if changes != 1 {
		t.Fatal("Changed did not report the asynchronous choice once")
	}
	_, child = open()
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	if callbacks != 1 || changes != 1 {
		t.Fatal("same selection emitted a value change")
	}
	stale, child = open()
	selected = &second
	parent.Frame()
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	stale.dismiss()
	if second != "one" || first != "two" || callbacks != 1 || h.IsOpen("choice") {
		t.Fatal("rebound control accepted an obsolete popup event")
	}
	stale, child = open()
	visible = false
	parent.Frame()
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	stale.dismiss()
	if h.IsPresent("choice") || stale.closes != 1 || callbacks != 1 || changes != 1 || second != "one" {
		t.Fatal("removed select leaked its popup or delivered a stale choice")
	}
	visible = true
	parent.Frame()
	_, child = open()
	h.CloseAll()
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	if callbacks != 1 || changes != 1 || second != "one" {
		t.Fatal("closed owner accepted child input")
	}
}

func TestSelectBackendFailureAndImmediateDismissDoNotChangeValue(t *testing.T) {
	for _, immediate := range []bool{false, true} {
		b := &fakeBackend{closeDuringOpen: immediate}
		if !immediate {
			b.err = errors.New("test open failure")
		}
		h := NewHostWithBackend(b)
		selected, callbacks, changes := "one", 0, 0
		sawError := false
		parent := ui.NewTester(func(c *ui.Context) {
			BindHost(c, h)
			h.BeginFrame()
			defer h.EndFrame()
			e, err := Select(c, nil, "failure", &selected, control.SelectProps{Label: "Choice", Options: []control.Option{{Value: "one"}, {Value: "two"}}, OnChange: func(string) { callbacks++ }})
			sawError = sawError || err != nil
			if e.Changed() {
				changes++
			}
		}, 300, 140)
		if err := parent.Click("Choice"); err != nil {
			t.Fatal(err)
		}
		parent.Frame()
		if selected != "one" || callbacks != 0 || changes != 0 || h.IsOpen("failure") || sawError == immediate {
			t.Fatalf("failed/dismissed popup changed state: immediate=%v error=%v selected=%q", immediate, sawError, selected)
		}
	}
}
