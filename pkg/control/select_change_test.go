package control

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSelectConvenienceCallbacksMatchChangedAndAllowRedirection(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "listbox", true: "system-menu"}[native], func(t *testing.T) {
			selected, callbacks, changes := "one", 0, 0
			redirect := false
			r := tester(func(c *ui.Context) {
				props := SelectProps{Label: "Plan", Options: []Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}}, OnChange: func(value string) {
					if selected != value {
						t.Fatal("callback preceded storage update")
					}
					callbacks++
					if redirect {
						selected = "one"
					}
				}}
				build := Select
				if native {
					build = NativeSelect
				}
				if build(c, &selected, props).Changed() {
					changes++
				}
			})
			chooseTwo := func() {
				t.Helper()
				if err := r.Click("Plan"); err != nil {
					t.Fatal(err)
				}
				if native {
					if err := r.ChooseMenuItem("Two"); err != nil {
						t.Fatal(err)
					}
				} else {
					r.Key(0, ui.KeyEnd)
					r.Key(0, ui.KeyEnter)
				}
				r.Frame()
			}
			chooseTwo()
			if selected != "two" || callbacks != 1 || changes != 1 {
				t.Fatalf("callback/Changed mismatch: selected=%q callbacks=%d changes=%d", selected, callbacks, changes)
			}
			chooseTwo()
			if callbacks != 1 || changes != 1 {
				t.Fatal("same choice emitted another change")
			}
			selected = "one"
			r.Frame()
			if callbacks != 1 || changes != 1 {
				t.Fatal("external model synchronization emitted a user change")
			}
			redirect = true
			chooseTwo()
			if selected != "one" || callbacks != 2 || changes != 2 {
				t.Fatal("synchronous callback redirection was lost or notified twice")
			}
		})
	}
}
