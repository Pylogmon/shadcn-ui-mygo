package display

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestAlertVisualIconSlotCannotCreateHiddenKeyboardStop(t *testing.T) {
	r := displayTester(func(c *ui.Context) {
		AlertWith(c, AlertProps{Title: "Notice"}, AlertSlots{Icon: func(child *ui.Context) { ui.Button(child, "Hidden icon control") }})
		ui.Button(c, "Visible next action")
	})
	r.Key(0, ui.KeyTab)
	if !r.Focused("Visible next action") {
		t.Fatal("decorative icon slot retained an invisible keyboard target")
	}
}
