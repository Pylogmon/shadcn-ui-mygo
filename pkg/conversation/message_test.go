package conversation

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestMessageSourceContentKeepsFullAvailableWidth(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			Message(c, MessageProps{}, func() {
				Bubble(c, BubbleProps{}, func() { ui.Text(c, strings.Repeat("Native message ", 30)) }).Role(ui.RoleGroup).Label("Face")
			})
		})
	}, 400, 500)
	face := bubbleNode(t, r, ui.RoleGroup, "Face")
	if math.Abs(float64(face.Bounds.W-320)) > .1 {
		t.Fatalf("source full-width message content was additionally capped: bubble width=%g want320", face.Bounds.W)
	}
}

func TestMessageReversePreservesSourceVisualAndKeyboardOrderInRTL(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			draft := ""
			calls := 0
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(20).Children(func() {
					MessageWith(c, MessageProps{Align: End, Label: "End message"}, MessageSlots{Avatar: func(child *ui.Context) { control.Button(child, "Avatar action", control.ButtonProps{}).Size(32, 32) }, Content: func(child *ui.Context) {
						control.Input(child, &draft, control.InputProps{Label: "Message draft"}).Width(180)
					}, Footer: func(child *ui.Context) {
						if control.Button(child, "Copy", control.ButtonProps{Size: control.ExtraSmall}).Clicked() {
							calls++
						}
					}})
				})
			}, 400, 220)
			r.SetDirection(direction)
			avatar := bubbleNode(t, r, ui.RoleButton, "Avatar action").Bounds
			if direction == ui.LeftToRight && avatar.X+avatar.W != 380 || direction == ui.RightToLeft && avatar.X != 20 {
				t.Fatalf("row-reverse visual side lost: %v", avatar)
			}
			r.Key(0, ui.KeyTab)
			if !r.Focused("Avatar action") {
				t.Fatal("visual row reverse also reversed source Tab order")
			}
			r.Key(0, ui.KeyTab)
			if !r.Focused("Message draft") {
				t.Fatal("native Tab did not reach content after avatar")
			}
			r.Type("typed")
			r.Key(0, ui.KeyTab)
			if !r.Focused("Copy") {
				t.Fatal("native Tab did not reach footer after content")
			}
			r.Key(0, ui.KeySpace)
			if draft != "typed" || calls != 1 {
				t.Fatal("message composition lost actual native input/actions")
			}
		})
	}
}

func TestMessageContentAlignsOnlyDeclaredSlotsAndRestoresBubbleScope(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			MessageWith(c, MessageProps{Align: End}, MessageSlots{Content: func(child *ui.Context) {
				ui.Box(child).Size(40, 20).Role(ui.RoleImage).Label("Raw child")
				MessageSlot(child, ui.Box(child).Size(40, 20).Role(ui.RoleImage).Label("Slot child"))
				Bubble(child, BubbleProps{}, func() { ui.Text(child, "Scoped") }).Role(ui.RoleGroup).Label("Scoped bubble")
			}})
			Bubble(c, BubbleProps{}, func() { ui.Text(c, "Outside") }).Role(ui.RoleGroup).Label("Outside bubble")
		})
	}, 400, 240)
	raw := bubbleNode(t, r, ui.RoleImage, "Raw child").Bounds
	slot := bubbleNode(t, r, ui.RoleImage, "Slot child").Bounds
	inside := bubbleNode(t, r, ui.RoleGroup, "Scoped bubble").Bounds
	outside := bubbleNode(t, r, ui.RoleGroup, "Outside bubble").Bounds
	if raw.X != 0 || slot.X+slot.W != 400 || inside.X+inside.W != 400 || outside.X != 0 {
		t.Fatalf("direct slot/source bubble alignment leaked: raw=%v slot=%v inside=%v outside=%v", raw, slot, inside, outside)
	}
}

func TestMessageHeaderFooterSourcePaddingAndGhostDeclaration(t *testing.T) {
	for _, ghost := range []bool{false, true} {
		t.Run(fmt.Sprint(ghost), func(t *testing.T) {
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				MessageWith(c, MessageProps{Ghost: ghost}, MessageSlots{Header: func(child *ui.Context) { ui.Text(child, "Header") }, Content: func(child *ui.Context) { ui.Box(child).Height(40) }, Footer: func(child *ui.Context) { ui.Text(child, "Footer") }})
			}, 400, 240)
			header := bubbleNode(t, r, ui.RoleText, "Header").Bounds
			footer := bubbleNode(t, r, ui.RoleText, "Footer").Bounds
			x := float32(12)
			if ghost {
				x = 0
			}
			if header.X != x || footer.X != x || header.H != 16 || footer.H != 16 {
				t.Fatalf("source metadata padding/text-xs geometry missing: header=%v footer=%v", header, footer)
			}
			if math.Abs(float64(footer.Y-header.Y-header.H-60)) > .1 {
				t.Fatalf("content source gap10 missing: header=%v footer=%v", header, footer)
			}
		})
	}
}

func TestMessageEmptyAvatarReservesSourceWidthAndGroupGap(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		MessageGroup(c, func(group *ui.Context) {
			for i := 0; i < 2; i++ {
				MessageWith(group.Key(i), MessageProps{Label: fmt.Sprintf("Row %d", i)}, MessageSlots{Avatar: func(child *ui.Context) {
					if i == 1 {
						ui.Box(child).Size(32, 32)
					}
				}, Content: func(child *ui.Context) {
					Bubble(child, BubbleProps{}, func() { ui.Text(child, "Text") }).Role(ui.RoleGroup).Label(fmt.Sprintf("Face %d", i))
				}})
			}
		})
	}, 400, 220)
	first := bubbleNode(t, r, ui.RoleGroup, "Face 0").Bounds
	second := bubbleNode(t, r, ui.RoleGroup, "Face 1").Bounds
	a := bubbleNode(t, r, ui.RoleArticle, "Row 0").Bounds
	b := bubbleNode(t, r, ui.RoleArticle, "Row 1").Bounds
	if first.X != 40 || second.X != 40 || b.Y-a.Y-a.H != 8 {
		t.Fatalf("empty avatar/group source spacing missing: first=%v second=%v rows=%v/%v", first, second, a, b)
	}
}

func TestMessageKeyedPrependAndEmptyHeaderKeepNativeIMEAndSelection(t *testing.T) {
	prepend, header := false, false
	draft := "draft"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		MessageGroup(c, func(group *ui.Context) {
			if prepend {
				Message(group, MessageProps{Key: "earlier"}, func() { ui.Text(group, "Earlier") })
			}
			slots := MessageSlots{Content: func(child *ui.Context) {
				control.Input(child, &draft, control.InputProps{Label: "Stable draft"}).Width(180)
			}}
			if header {
				slots.Header = func(*ui.Context) {}
			}
			MessageWith(group, MessageProps{Key: "live"}, slots)
		})
	}, 400, 240)
	if err := r.Click("Stable draft"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	prepend, header = true, true
	r.Frame()
	r.Type("replacement")
	if draft != "replacement" || !r.Focused("Stable draft") {
		t.Fatal("keyed prepend/header changed native editor identity or selection")
	}
	r.Compose("文", 1)
	header = false
	r.Frame()
	r.Type("文")
	if draft != "replacement文" {
		t.Fatalf("empty slot toggle lost native IME composition: %q", draft)
	}
}

func TestMessageCompoundPreservesNativeFontAndSemanticMetadata(t *testing.T) {
	render := func(message bool) *ui.Tester {
		return ui.NewTester(func(c *ui.Context) {
			tokens := theme.Light()
			tokens.MutedForeground = ui.RGB(20, 80, 150)
			native := tokens.Theme()
			native.Font = "monospace"
			ui.Themed(c, native).Fill().FontWeight(700).Children(func() {
				if message {
					MessageWith(c, MessageProps{}, MessageSlots{Header: func(child *ui.Context) { ui.Text(child, "Metadata") }})
				} else {
					ui.Row(c).PaddingX(12).FontSize(12).LineHeight(4.0 / 3).FontWeight(500).TextColor(tokens.MutedForeground).Children(func() { ui.Text(c, "Metadata") })
				}
			})
		}, 400, 140)
	}
	actual, reference := render(true), render(false)
	if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
		t.Fatal("message metadata lost native theme font, semantic color, or source font-medium")
	}
}

func TestMessageNativeAttachmentSlotAndFooterURLRemainRealActions(t *testing.T) {
	opened := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		MessageWith(c, MessageProps{Align: End, Label: "Files"}, MessageSlots{Content: func(child *ui.Context) {
			MessageSlot(child, Attachment(child, AttachmentProps{Name: "report.pdf", SizeBytes: 1024, OnOpen: func() { opened++ }})).Role(ui.RoleGroup).Label("Attached file")
		}, Footer: func(child *ui.Context) { ui.Link(child, "Open details", "https://example.com/message") }})
	}, 500, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	file := bubbleNode(t, r, ui.RoleGroup, "Attached file").Bounds
	if math.Abs(float64(file.X+file.W-500)) > .1 {
		t.Fatalf("declared attachment slot did not inherit source end alignment: %v", file)
	}
	if err := r.Click("Open"); err != nil {
		t.Fatal(err)
	}
	if opened != 1 {
		t.Fatal("message attachment native open action did not execute once")
	}
	if err := r.Click("Open details"); err != nil {
		t.Fatal(err)
	}
	if urls := r.OpenedURLs(); len(urls) != 1 || urls[0] != "https://example.com/message" {
		t.Fatalf("footer replaced actual native URL behavior: %v", urls)
	}
}

func TestMessageCustomRootChildrenAvoidAutomaticContentAssembly(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		MessageWith(c, MessageProps{Sender: "Semantic sender", Label: "Raw message"}, MessageSlots{Children: func(child *ui.Context) {
			ui.Row(child).FillWidth().Height(20).Role(ui.RoleStatus).Label("Checking logs")
		}})
	}, 400, 180)
	message := bubbleNode(t, r, ui.RoleArticle, "Raw message").Bounds
	status := bubbleNode(t, r, ui.RoleStatus, "Checking logs").Bounds
	if message.W != 400 || message.H != 20 || status.W != 400 || r.HasText("Semantic sender") {
		t.Fatalf("custom children gained an implicit header/content wrapper: message=%v status=%v", message, status)
	}
}

func TestMessageContentUsesNativeSelectableText(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		MessageWith(c, MessageProps{}, MessageSlots{Content: func(child *ui.Context) { ui.Text(child, "Selectable native transcript").Selectable() }})
	}, 400, 180)
	if err := r.Click("Selectable native transcript"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "Selectable native transcript" {
		t.Fatalf("message wrapper broke native transcript selection: %q", r.Clipboard())
	}
}

func TestMessageSourceFooterMovesAvatarUp32DIP(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			Message(c, MessageProps{Sender: "Sofia", Avatar: func() { ui.Box(c).Size(32, 32).Role(ui.RoleImage).Label("Avatar") }, Footer: func() { ui.Text(c, "Delivered") }}, func() { ui.Box(c).Size(150, 70) }).Label("Message")
		})
	}, 400, 240)
	message := bubbleNode(t, r, ui.RoleArticle, "Message")
	avatar := bubbleNode(t, r, ui.RoleImage, "Avatar")
	if delta := message.Bounds.Y + message.Bounds.H - avatar.Bounds.Y - avatar.Bounds.H; math.Abs(float64(delta-32)) > .1 {
		t.Fatalf("source footer avatar translation=%g want32", delta)
	}
}
