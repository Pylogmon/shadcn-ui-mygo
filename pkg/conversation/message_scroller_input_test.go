package conversation

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestMessageScrollerCompoundButtonsNativeFocusAXAndPreventDefault(t *testing.T) {
	var s MessageScrollerState
	zero := float32(0)
	prevent := true
	calls := 0
	viewportFocused := false
	r := ui.NewTesterWith(func(c *ui.Context) {
		MessageScrollerWith(c, &s, MessageScrollerProviderProps{AutoScroll: true, DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
			MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
				MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero, Label: "Transcript"}, func(c *ui.Context) {
					for i := 0; i < 8; i++ {
						MessageScrollerItem(c, MessageScrollerItemProps{ID: fmt.Sprint(i)}, func(c *ui.Context) { ui.Text(c, fmt.Sprint(i)).Height(80) })
					}
				})
			})
			MessageScrollerButton(c.Key("start"), MessageScrollerButtonProps{Direction: MessageScrollerStart, Instant: true})
			MessageScrollerButton(c.Key("end"), MessageScrollerButtonProps{Instant: true, OnClick: func() bool { calls++; return !prevent }})
		})
		viewportFocused = s.Scroll.Handle.Focused(c)
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if _, ok := r.AccessByRole(ui.RoleRegion, "Messages"); !ok {
		t.Fatal("native region absent")
	}
	if _, ok := r.AccessByRole(ui.RoleLog, "Transcript"); !ok {
		t.Fatal("native log absent")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Scroll to start"); ok {
		t.Fatal("inactive start exposed to accessibility")
	}
	end, ok := r.AccessByRole(ui.RoleButton, "Scroll to end")
	if !ok {
		t.Fatal("active end absent")
	}
	if end.Bounds.W != 32 || end.Bounds.H != 32 || end.Bounds.X != 144 || end.Bounds.Y != 152 {
		t.Fatalf("button geometry=%v", end.Bounds)
	}
	if err := r.Click("Scroll to end"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || s.Scroll.Y != 0 {
		t.Fatalf("prevent calls%v y%v", calls, s.Scroll.Y)
	}
	prevent = false
	if err := r.Click("Scroll to end"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || s.Scroll.Y != 440 || !viewportFocused {
		t.Fatalf("click calls%v y%v viewportfocus%v", calls, s.Scroll.Y, viewportFocused)
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Scroll to end"); ok {
		t.Fatal("caught-up end still exposed")
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerCompoundSmoothInputCancellationAndRunningReducedMotion(t *testing.T) {
	for _, action := range []string{"wheel", "key", "reduce", "hidden"} {
		t.Run(action, func(t *testing.T) {
			var s MessageScrollerState
			messages := scrollerRows(30)
			p := MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}
			r := scrollerFixture(t, &s, &messages, &p)
			r.SetPreferences(ui.Preferences{})
			if !s.ScrollToEnd(MessageScrollerScrollOptions{Smooth: true}) {
				t.Fatal("end request rejected")
			}
			r.Frame()
			r.Advance(50 * time.Millisecond)
			if s.Scroll.Y <= 0 || s.Scroll.Y >= 2200 || !s.Scroll.Scrolling() {
				t.Fatalf("not a real native glide y=%v moving=%v", s.Scroll.Y, s.Scroll.Scrolling())
			}
			switch action {
			case "wheel":
				r.Scroll(100, 100, 0, -50)
				if s.Mode() != MessageScrollerFree {
					t.Fatalf("wheel retained mode%v", s.Mode())
				}
				y := s.Scroll.Y
				r.Advance(time.Second)
				r.Frame()
				if math.Abs(float64(s.Scroll.Y-y)) > .5 || s.Scroll.Scrolling() {
					t.Fatalf("wheel did not cancel y%v→%v moving%v", y, s.Scroll.Y, s.Scroll.Scrolling())
				}
			case "key":
				s.Scroll.Handle.Focus()
				r.Frame()
				r.Key(0, ui.KeyHome)
				if s.Scroll.Y != 0 || s.Scroll.Scrolling() || s.Mode() != MessageScrollerFree {
					t.Fatalf("Home y%v moving%v mode%v", s.Scroll.Y, s.Scroll.Scrolling(), s.Mode())
				}
			case "reduce":
				r.SetPreferences(ui.Preferences{ReduceMotion: true})
				if s.Scroll.Y != 2200 || s.Scroll.Scrolling() {
					t.Fatalf("running reduce y%v moving%v", s.Scroll.Y, s.Scroll.Scrolling())
				}
			case "hidden":
				r.SetWindowHidden(true)
				r.Advance(time.Second)
				r.SetWindowHidden(false)
				r.Frame()
				r.Advance(time.Second)
				if s.Scroll.Y != 2200 || s.Scroll.Scrolling() {
					t.Fatalf("resume y%v moving%v", s.Scroll.Y, s.Scroll.Scrolling())
				}
			}
			scrollerIdle(t, r)
		})
	}
}

func TestMessageScrollerCompoundEditorSelectionIMEAndKeyedReorder(t *testing.T) {
	var s MessageScrollerState
	zero := float32(0)
	ids := []string{"one", "two", "three"}
	text := "original"
	focused := false
	selectionStart, selectionEnd := 0, 0
	var editor ui.Handle
	r := ui.NewTesterWith(func(c *ui.Context) {
		MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
			MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
				MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
					for _, id := range ids {
						MessageScrollerItem(c, MessageScrollerItemProps{ID: id}, func(c *ui.Context) {
							if id == "two" {
								ui.TextInput(c, &text).Label("Message editor").Bind(&editor).Height(50)
							} else {
								text := ui.Text(c, id).Selectable().Height(50)
								if id == "one" {
									selectionStart, selectionEnd = text.TextSelection()
								}
							}
						})
					}
				})
			})
		})
		focused = editor.Focused(c)
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Message editor"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Compose("中文", 2)
	ids = []string{"three", "two", "one"}
	r.Frame()
	if !focused {
		t.Fatal("stable message lost native editor focus")
	}
	r.Type("替换")
	if text != "替换" {
		t.Fatalf("keyed native selection/IME lost: %q", text)
	}
	if err := r.Click("one"); err != nil {
		t.Fatal(err)
	}
	row, _ := r.Find("one")
	r.Press(row.X+2, row.Y+20)
	r.Move(row.X+30, row.Y+20)
	r.Release(row.X+30, row.Y+20)
	if selectionEnd <= selectionStart {
		t.Fatalf("native message text selection absent %d:%d", selectionStart, selectionEnd)
	}
	scrollerIdle(t, r)
}
