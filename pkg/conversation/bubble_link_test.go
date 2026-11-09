package conversation

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestBubbleAnchorWithNativeBoxDoesNotPanic(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		BubbleContent(c, BubbleContentProps{Trigger: &BubbleTrigger{Kind: BubbleLink, Label: "Compound anchor", URL: "https://example.com"}}, func() {
			ui.Box(c).Size(16, 16)
			ui.Text(c, "Compound anchor")
		})
	}, 400, 200)
}

func TestBubbleCompoundLinkKeepsNativeHandleWholeFaceActivationAndFluentName(t *testing.T) {
	var handle ui.Handle
	calls := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() {
			BubbleContent(c, BubbleContentProps{Variant: Secondary, Trigger: &BubbleTrigger{
				Kind: BubbleLink, URL: "https://example.com/compound", Label: "Original name", Handle: &handle,
				OnActivate: func() { calls++ },
			}}, func() {
				ui.Row(c).Gap(6).Children(func() {
					ui.Icon(c, attachmentDefaultIcon("check")).Size(16, 16)
					ui.Box(c).Children(func() { ui.Text(c, "Rich native content") })
				})
			}).Label("Current compound name")
		})
	}, 420, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	link := bubbleNode(t, r, ui.RoleLink, "Current compound name")
	if !link.Focusable || link.Bounds.W < 100 || link.Bounds.H < 30 {
		t.Fatal("compound face lost its real native link target")
	}
	// Native Link occupies the padding as well as the glyph region.
	r.ClickAt(link.Bounds.X+2, link.Bounds.Y+link.Bounds.H/2)
	if calls != 1 || len(r.OpenedURLs()) != 1 {
		t.Fatal("padding did not activate the actual native URL")
	}
	handle.Focus()
	r.Frame()
	if !r.Focused("Current compound name") {
		t.Fatal("Handle did not bind the actual native target")
	}
	r.Key(0, ui.KeySpace)
	if calls != 1 || len(r.OpenedURLs()) != 1 {
		t.Fatal("compound link acquired Button Space behavior")
	}
	r.Key(0, ui.KeyEnter)
	if calls != 2 || len(r.OpenedURLs()) != 2 {
		t.Fatal("compound link lost native Enter activation")
	}
	for _, node := range r.AccessAll() {
		if node.Name == "Rich native content" || node.Role == ui.RoleImage {
			t.Fatal("inert compound visuals became duplicate AX targets")
		}
	}
}

func TestBubbleCompoundLinkInheritsQuietSourceColorWithoutNativeUnderline(t *testing.T) {
	for _, variant := range []BubbleVariant{Muted, Outline, Ghost} {
		for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
			t.Run(fmt.Sprintf("%d/%d", variant, direction), func(t *testing.T) {
				render := func(link bool) *ui.Tester {
					return ui.NewTesterWith(func(c *ui.Context) {
						theme.Apply(c)
						ui.Column(c).Fill().Padding(20).Font("monospace").FontWeight(700).TextColor(ui.RGB(40, 90, 160)).Children(func() {
							props := BubbleContentProps{Variant: variant}
							if link {
								props.Trigger = &BubbleTrigger{Kind: BubbleLink, URL: "https://example.com", Label: "Compound"}
							}
							BubbleContent(c, props, func() { ui.Text(c, "Inherited native text") })
						})
					}, 420, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				}
				actual, reference := render(true), render(false)
				actual.SetDirection(direction)
				reference.SetDirection(direction)
				if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
					t.Fatal("compound anchor changed source glyph styles before hover")
				}
			})
		}
	}
}

func TestBubbleCompoundLinkUnnamedVisualUsesURLFallback(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		BubbleContent(c, BubbleContentProps{Trigger: &BubbleTrigger{Kind: BubbleLink, URL: "https://example.com/preview"}}, func() {
			ui.Box(c).Size(60, 24).Background(ui.RGB(30, 80, 160))
		})
	}, 200, 100)
	bubbleNode(t, r, ui.RoleLink, "https://example.com/preview")
}
