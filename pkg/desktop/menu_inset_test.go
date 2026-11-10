package desktop_test

import (
	"image"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// One public, independent-window flow checks logical padding, both selection
// glyphs with independent business icons, and the mixed-menu close contract.
func TestMenuInsetSelectionIndependentWindowWorkflow(t *testing.T) {
	b := &itemHeightBackend{}
	h := desktop.NewHostWithBackend(b)
	open, checked, rtl := false, true, false
	keepOpen := false
	state := desktop.MenuState{}
	calls := 0
	chosen := ""
	items := []overlay.MenuItem{
		{Key: "normal", Label: "Normal", Group: "Actions"},
		{Key: "inset", Label: "Inset", Inset: true},
		{Key: "check", Label: "Email", Checked: &checked, Icon: outlined.Mail(), Inset: true, Group: "Appearance", GroupInset: true},
		{Key: "radio", Label: "Card", Radio: true, Selected: true, Icon: outlined.CreditCard()},
		{Key: "more", Label: "More", Inset: true, Children: []overlay.MenuItem{{Key: "child", Label: "Child"}}},
		{Key: "done", Label: "Done"},
	}
	parent := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c).Direction(direction).Children(func() {
			anchor := ui.Button(c, "Actions")
			_, err := desktop.Menu(c, h, "inset", anchor, &open, &state, desktop.MenuProps{Label: "Inset menu", Width: 260, Items: items, KeepOpenOnSelect: keepOpen, ItemContent: func(child *ui.Context, item overlay.MenuItem) { ui.Text(child, item.Label+" body").SingleLine() }, OnChoose: func(k string) { calls++; chosen = k }})
			if err != nil {
				t.Fatal(err)
			}
		})
	}, 340, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := parent.Click("Actions"); err != nil {
		t.Fatal(err)
	}
	child := b.surfaces[0].tester
	text := func(label string) ui.Rect {
		t.Helper()
		r, ok := child.Find(label + " body")
		if !ok {
			t.Fatalf("missing body %s", label)
		}
		return r
	}
	normal, inset := text("Normal"), text("Inset")
	if math.Abs(float64(inset.X-normal.X-24)) > .1 {
		t.Fatalf("logical start delta normal=%v inset=%v", normal, inset)
	}
	email, ok := child.AccessByRole(ui.RoleMenuItemCheckBox, "Email")
	if !ok || !email.Checked {
		t.Fatal("checkbox semantics lost")
	}
	card, ok := child.AccessByRole(ui.RoleMenuItemRadio, "Card")
	if !ok || !card.Checked {
		t.Fatal("radio semantics lost")
	}
	// Selected indicators must coexist with business icons on opposite ends.
	ink := func(img *image.RGBA, x, y int) int {
		t.Helper()
		n := 0
		for yy := y; yy < y+16; yy++ {
			for xx := x; xx < x+16; xx++ {
				r, g, b, a := img.At(xx, yy).RGBA()
				if a > 0 && r < 45000 && g < 45000 && b < 45000 {
					n++
				}
			}
		}
		return n
	}
	for _, a := range []ui.AccessInfo{email, card} {
		start := 8
		if a.Name == "Email" {
			start = 32
		}
		if ink(child.Image(), int(a.Bounds.X)+start, int(a.Bounds.Y+(a.Bounds.H-16)/2)) == 0 || ink(child.Image(), int(a.Bounds.X+a.Bounds.W)-24, int(a.Bounds.Y+(a.Bounds.H-16)/2)) == 0 {
			t.Fatalf("business icon or ending check missing in %s", a.Name)
		}
	}
	caption, _ := child.Find("Appearance")
	captionInkStart := func() int {
		t.Helper()
		for x := int(caption.X) + 2; x < int(caption.X+caption.W)-2; x++ {
			for y := int(caption.Y) + 3; y < int(caption.Y+caption.H)-3; y++ {
				r, g, b, a := child.Image().At(x, y).RGBA()
				if a > 0 && r < 45000 && g < 45000 && b < 45000 {
					return x
				}
			}
		}
		t.Fatal("missing group caption ink")
		return 0
	}
	before := captionInkStart()
	items[2].GroupInset = false
	parent.Frame()
	child.Frame()
	after := captionInkStart()
	if before-after != 24 {
		t.Fatalf("group caption logical inset delta=%d want24", before-after)
	}

	// Checkbox selection keeps this mixed menu open and dispatches once.
	child.Key(0, ui.KeyHome)
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	if checked || !open || calls != 1 || chosen != "check" {
		t.Fatalf("mixed checkbox calls=%d chosen=%s checked=%v open=%v", calls, chosen, checked, open)
	}
	// A radio selection also stays open.
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	if !open || calls != 2 || chosen != "radio" {
		t.Fatal("radio did not dispatch once while remaining open")
	}
	// A normal action closes by default.
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	if open || calls != 3 || chosen != "done" {
		t.Fatalf("action did not close once calls=%d chosen=%s open=%v", calls, chosen, open)
	}
	// Direction is read by the child and logical inset moves to the right.
	rtl = true
	parent.Frame()
	if err := parent.Click("Actions"); err != nil {
		t.Fatal(err)
	}
	child = b.surfaces[len(b.surfaces)-1].tester
	normal, inset = text("Normal"), text("Inset")
	if math.Abs(float64((normal.X+normal.W)-(inset.X+inset.W)-24)) > .1 {
		t.Fatalf("RTL logical end did not mirror normal=%v inset=%v", normal, inset)
	}
	// Explicit per-item close overrides the checkbox default.
	yes := true
	items[2].CloseOnSelect = &yes
	parent.Frame()
	child.Frame()
	child.Key(0, ui.KeyHome)
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	if open || calls != 4 {
		t.Fatal("per-item CloseOnSelect was ignored")
	}
	no := false
	items[5].CloseOnSelect = &no
	parent.Frame()
	if err := parent.Click("Actions"); err != nil {
		t.Fatal(err)
	}
	child = b.surfaces[len(b.surfaces)-1].tester
	child.Key(0, ui.KeyEnd)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	if !open || calls != 5 {
		t.Fatal("explicit false did not retain an action")
	}
	keepOpen = true
	parent.Frame()
	child.Frame()
	child.Key(0, ui.KeyHome)
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	if !open || calls != 6 {
		t.Fatal("global keep-open did not override explicit row close")
	}
	child.Key(0, ui.KeyEscape)
	parent.Frame()

}
