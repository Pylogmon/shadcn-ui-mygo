package overlay

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestInWindowTooltipHoverOnlyKeepsDescriptionWithoutFocusOpening(t *testing.T) {
	hoverOnly := true
	r := ui.NewTester(view(func(c *ui.Context) {
		anchor := ui.Button(c, "Trigger")
		InWindowTooltip(c, anchor, "Helpful hint", TooltipProps{HoverOnly: hoverOnly})
	}), 600, 300)
	r.Key(0, ui.KeyTab)
	r.Advance(time.Second)
	trigger, ok := r.AccessByRole(ui.RoleButton, "Trigger")
	if !ok || !trigger.Focused || trigger.Description != "Helpful hint" || r.HasText("Helpful hint") {
		t.Fatalf("focus-only trigger lost its description or opened a tooltip: %+v", trigger)
	}
	r.Move(trigger.Bounds.X+trigger.Bounds.W/2, trigger.Bounds.Y+trigger.Bounds.H/2)
	r.Advance(time.Second)
	if !r.HasText("Helpful hint") {
		t.Fatal("hover did not open the tooltip")
	}
	r.Move(580, 280)
	r.Advance(time.Second)
	if r.HasText("Helpful hint") || !r.Focused("Trigger") {
		t.Fatal("focus retained the tooltip after the pointer left")
	}
	hoverOnly = false
	r.Frame()
	r.Advance(time.Second)
	if !r.HasText("Helpful hint") || !r.Focused("Trigger") {
		t.Fatal("default keyboard tooltip behavior was not restored")
	}
}
