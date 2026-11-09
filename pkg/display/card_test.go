package display

import (
	"image/color"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestCompoundCardPaddedSlotsActionAndRTL(t *testing.T) {
	clicks := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(280).Padding(20).Children(func() {
			CardWith(c, CardProps{Label: "Meeting notes", HeaderBorder: true, FooterBorder: true}, CardSlots{
				Title: "A long title wraps beside its action", Description: "Last updated today",
				Action: func(child *ui.Context) {
					if control.Button(child, "Edit", control.ButtonProps{Variant: control.Ghost}).Clicked() {
						clicks++
					}
				},
				Content: func(child *ui.Context) { ui.Box(child).Height(28).FillWidth().Label("Card body") },
				Footer:  func(child *ui.Context) { control.Button(child, "Save", control.ButtonProps{}) },
			})
		})
	}, 320, 480)
	card, ok := r.AccessByRole(ui.RoleGroup, "Meeting notes")
	if !ok {
		t.Fatal("named card lacks group semantics")
	}
	body, _ := r.Find("Card body")
	edit, _ := r.AccessByRole(ui.RoleButton, "Edit")
	save, _ := r.AccessByRole(ui.RoleButton, "Save")
	if body.X-card.Bounds.X != 25 || (card.Bounds.X+card.Bounds.W)-(body.X+body.W) != 25 || save.Bounds.X != body.X || (edit.Bounds.X+edit.Bounds.W) != (body.X+body.W) {
		t.Fatalf("slot insets card=%+v body=%+v edit=%+v save=%+v", card.Bounds, body, edit.Bounds, save.Bounds)
	}
	if body.Y < (edit.Bounds.Y+edit.Bounds.H)+24 || save.Bounds.Y < (body.Y+body.H)+48 {
		t.Fatalf("header/footer dividers lost vertical spacing: edit=%+v body=%+v footer=%+v", edit.Bounds, body, save.Bounds)
	}
	displayClick(t, r, "Edit")
	if clicks != 1 {
		t.Fatalf("action clicked %d times", clicks)
	}
	r.SetDirection(ui.RightToLeft)
	edit, _ = r.AccessByRole(ui.RoleButton, "Edit")
	card, _ = r.AccessByRole(ui.RoleGroup, "Meeting notes")
	if edit.Bounds.X != card.Bounds.X+25 {
		t.Fatalf("RTL action did not move to logical end: card=%+v action=%+v", card.Bounds, edit.Bounds)
	}
}

func TestCompoundCardSizesAndFiniteSpacing(t *testing.T) {
	for _, example := range []struct {
		size    CardSize
		spacing float32
		want    float32
	}{{CardDefault, 0, 24}, {CardSmall, 0, 16}, {CardSmall, 32, 32}, {CardDefault, float32(math.NaN()), 24}, {CardSmall, float32(math.Inf(1)), 16}} {
		r := displayTester(func(c *ui.Context) {
			CardWith(c, CardProps{Label: "Spacing", Size: example.size, Spacing: example.spacing}, CardSlots{
				Content: func(child *ui.Context) { ui.Box(child).FillWidth().Height(20).Label("Content bounds") },
			})
		})
		card, _ := r.AccessByRole(ui.RoleGroup, "Spacing")
		body, _ := r.Find("Content bounds")
		if body.X-card.Bounds.X != example.want+1 || body.Y-card.Bounds.Y != example.want+1 || (card.Bounds.Y+card.Bounds.H)-(body.Y+body.H) != example.want+1 {
			t.Fatalf("size=%v spacing=%v: card=%+v body=%+v", example.size, example.spacing, card.Bounds, body)
		}
	}
}

func TestCompoundCardMediaBleedClipAndBottomEdge(t *testing.T) {
	for _, bottom := range []bool{false, true} {
		r := displayTester(func(c *ui.Context) {
			CardWith(c, CardProps{Label: "Media card", ContentBleed: true, MediaBottom: bottom}, CardSlots{
				Media: func(child *ui.Context) {
					ui.Box(child).FillWidth().Height(80).Background(ui.RGB(240, 50, 50)).Label("Media bounds")
				},
				Content: func(child *ui.Context) {
					ui.Box(child).FillWidth().Height(24).Background(ui.RGB(40, 50, 240)).Label("Bleed bounds")
				},
			})
		})
		card, _ := r.AccessByRole(ui.RoleGroup, "Media card")
		media, _ := r.Find("Media bounds")
		body, _ := r.Find("Bleed bounds")
		if media.X != card.Bounds.X+1 || body.X != media.X || media.W != body.W {
			t.Fatalf("edge-to-edge widths card=%+v media=%+v body=%+v", card.Bounds, media, body)
		}
		if !bottom && media.Y != card.Bounds.Y+1 || bottom && (media.Y+media.H) != (card.Bounds.Y+card.Bounds.H)-1 {
			t.Fatalf("media not flush with its edge bottom=%v card=%+v media=%+v", bottom, card.Bounds, media)
		}
		y := media.Y + 30
		if got := r.Image().RGBAAt(int(media.X+2), int(y)); got != (color.RGBA{R: 240, G: 50, B: 50, A: 255}) {
			t.Fatalf("media did not actually paint the side edge: %v", got)
		}
		if got := r.Image().RGBAAt(int(media.X+media.W/2), int(media.Y+10)); got.R != 240 || got.B != 50 {
			t.Fatalf("media paint=%v", got)
		}
	}
}

func TestCompoundCardOptionalHeaderPreservesBodyLocalStateAndFocus(t *testing.T) {
	show := false
	r := displayTester(func(c *ui.Context) {
		slots := CardSlots{Content: func(child *ui.Context) {
			owner := ui.Column(child)
			text := ui.Local(owner, "draft", func() string { return "Draft" })
			owner.Children(func() { control.Input(child, text, control.InputProps{Label: "Card draft"}) })
		}}
		if show {
			slots.Header = func(child *ui.Context) { ui.Text(child, "Optional header") }
		}
		CardWith(c, CardProps{}, slots)
	})
	displayClick(t, r, "Card draft")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Edited")
	show = true
	r.Frame()
	if !r.Focused("Card draft") {
		t.Fatal("header insertion lost editor focus")
	}
	input, ok := r.AccessByRole(ui.RoleTextField, "Card draft")
	if !ok || input.Value != "Edited" {
		t.Fatalf("header insertion remounted body editor: %+v", input)
	}
}

func TestCompoundCardEmptySlotsDoNotLeakIdentityToFollowingEditor(t *testing.T) {
	for _, slot := range []string{"header", "media", "content", "footer"} {
		t.Run(slot, func(t *testing.T) {
			show, text := false, "Outside draft"
			r := displayTester(func(c *ui.Context) {
				slots := CardSlots{}
				build := func(child *ui.Context) {
					if show {
						ui.Text(child, "Optional content")
					}
				}
				switch slot {
				case "header":
					slots.Header = build
				case "media":
					slots.Media = build
				case "content":
					slots.Content = build
				case "footer":
					slots.Footer = build
				}
				CardWith(c, CardProps{}, slots)
				control.Input(c, &text, control.InputProps{Label: "Following editor"})
			})
			displayClick(t, r, "Following editor")
			show = true
			r.Frame()
			if !r.Focused("Following editor") {
				t.Fatal("empty slot leaked an unconsumed key into the following editor")
			}
		})
	}
}
