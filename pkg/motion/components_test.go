package motion_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestComponentsApplyReducedMotionDuringRunningFeedback(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprintf("dark=%v", dark), func(t *testing.T) {
			tokens := theme.Light()
			if dark {
				tokens = theme.Dark()
			}
			invalid, selected, page := false, 0, 1
			text, code := "Value", "1234"
			view := func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
					ui.Column(c).Fill().Padding(16).Gap(16).Children(func() {
						buttonProps, badgeVariant := control.ButtonProps{}, display.BadgeDefault
						if invalid {
							buttonProps, badgeVariant = control.ButtonProps{Variant: control.Soft, Color: &tokens.Destructive}, display.BadgeDestructive
						}
						control.Button(c, "Action", buttonProps)
						display.Badge(c, "Status", badgeVariant)
						control.Input(c, &text, control.InputProps{Label: "Name", Invalid: invalid})
						control.InputOTP(c, &code, control.InputOTPProps{Length: 4, Invalid: invalid})
						navigation.Tabs(c, &selected, navigation.TabsProps{Variant: navigation.TabsLine, Items: []navigation.TabItem{{Label: "First"}, {Label: "Second"}}}, nil)
						navigation.PaginationWith(c, &page, navigation.PaginationProps{Pages: 3})
					})
				})
			}
			r := ui.NewTester(view, 500, 380)
			invalid, selected, page = true, 1, 2
			r.Frame()
			r.Advance(30 * time.Millisecond)
			if !r.FramePending() {
				t.Fatal("component feedback did not begin animating")
			}
			r.SetPreferences(ui.Preferences{ReduceMotion: true})
			reference := ui.NewTesterWith(view, 500, 380, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if !bytes.Equal(r.Image().Pix, reference.Image().Pix) {
				t.Fatal("running components did not immediately match their static reduced-motion rendering")
			}
			if r.FramePending() {
				t.Fatal("running component feedback kept scheduling reduced-motion frames")
			}
			r.SetPreferences(ui.Preferences{})
			invalid, selected, page = false, 0, 1
			r.Frame()
			if !r.FramePending() {
				t.Fatal("component feedback did not resume after restoring animation")
			}
			r.Advance(time.Second)
			if r.FramePending() {
				t.Fatal("restored component feedback did not settle")
			}
		})
	}
}
