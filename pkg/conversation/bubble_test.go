package conversation

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestBubbleDestructiveUsesSourceTintedSurface(t *testing.T) {
	tokens := theme.Light()
	tokens.Destructive = ui.RGB(220, 30, 50)
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Scope(c, tokens).Fill().Background(ui.RGB(255, 255, 255)).Children(func() {
			Bubble(c, BubbleProps{Variant: Destructive}, func() { ui.Text(c, "Failed") }).Role(ui.RoleGroup).Label("Bubble")
		})
	}, 500, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	n, _ := r.AccessByRole(ui.RoleGroup, "Bubble")
	got := r.Image().RGBAAt(int(n.Bounds.X+8), int(n.Bounds.Y+n.Bounds.H/2))
	want := color.RGBA{251, 232, 234, 255}
	if got != want {
		t.Fatalf("destructive source uses destructive/10 surface, got=%v want=%v", got, want)
	}
}
func TestBubbleReactionsOverlayWithoutChangingContentHeight(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Gap(60).Children(func() {
			Bubble(c, BubbleProps{}, func() { ui.Text(c, "Same text") }).Role(ui.RoleGroup).Label("Bare")
			Bubble(c, BubbleProps{Reactions: []Reaction{{Label: "Like", Count: 3, Selected: true}}}, func() { ui.Text(c, "Same text") }).Role(ui.RoleGroup).Label("Reacted")
		})
	}, 500, 300)
	bare, _ := r.AccessByRole(ui.RoleGroup, "Bare")
	reacted, _ := r.AccessByRole(ui.RoleGroup, "Reacted")
	if bare.Bounds.H != reacted.Bounds.H {
		t.Fatalf("reaction entered content flow: bare=%+v reacted=%+v", bare.Bounds, reacted.Bounds)
	}
	reaction, ok := r.AccessByRole(ui.RoleToggleButton, "Like 3")
	if !ok || !reaction.Checked {
		t.Fatal("selected reaction lacks real native pressed semantics")
	}
}

func bubbleTestView(content func(*ui.Context)) func(*ui.Context) {
	return func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(40).Gap(50).AlignItems(ui.Start).Children(func() { content(c) })
	}
}
func bubbleNode(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	n, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing %v %q: %+v", role, name, r.AccessAll())
	}
	return n
}
func TestBubbleCompoundSeparatesFacesAndSupportsNativeButtonLink(t *testing.T) {
	calls := 0
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		BubbleGroup(c, func() {
			Bubble(c, BubbleProps{Trigger: &BubbleTrigger{Label: "Button bubble", OnActivate: func() { calls++ }}}, func() { ui.Text(c, "Button face") })
			BubbleWith(c, BubbleProps{Variant: Secondary}, BubbleSlots{Content: func(c *ui.Context, p BubbleContentProps) {
				p.Trigger = &BubbleTrigger{Kind: BubbleLink, Label: "Link bubble", URL: "https://example.com/bubble", OnActivate: func() { calls++ }}
				BubbleContent(c, p, func() { ui.Text(c, "Link face") })
			}})
		})
	}), 500, 320, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Button bubble"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeySpace)
	r.Key(0, ui.KeyEnter)
	if calls != 3 {
		t.Fatalf("native button delivery %d", calls)
	}
	if err := r.Click("Link bubble"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeySpace)
	if calls != 4 || len(r.OpenedURLs()) != 1 {
		t.Fatalf("native link Space activated: %d %v", calls, r.OpenedURLs())
	}
	r.Key(0, ui.KeyEnter)
	if calls != 5 || len(r.OpenedURLs()) != 2 {
		t.Fatalf("native link Enter/URL missing: %d %v", calls, r.OpenedURLs())
	}
}
func TestBubbleVariantsAndInteractiveHoverPixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, variant := range []BubbleVariant{Primary, Secondary, Muted, Tinted, Outline, Ghost, Destructive} {
			t.Run(fmt.Sprintf("dark=%v/variant=%d", dark, variant), func(t *testing.T) {
				tokens := theme.Light()
				if dark {
					tokens = theme.Dark()
				}
				tokens.Primary = theme.OKLCH(.6, .1, 250)
				tokens.Secondary = ui.RGB(255, 255, 255)
				tokens.Muted = ui.RGB(255, 255, 255)
				tokens.Foreground = ui.RGB(0, 0, 0)
				tokens.Background = ui.RGB(255, 255, 255)
				r := ui.NewTesterWith(func(c *ui.Context) {
					theme.Scope(c, tokens).Fill().Background(ui.RGB(255, 255, 255)).Padding(40).Children(func() {
						Bubble(c, BubbleProps{Variant: variant, Trigger: &BubbleTrigger{Label: "Face"}}, func() { ui.Box(c).Size(100, 30) })
					})
				}, 500, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				b := bubbleNode(t, r, ui.RoleButton, "Face").Bounds
				sample := func() color.RGBA { return r.Image().RGBAAt(int(b.X+8), int(b.Y+b.H/2)) }
				expected := func(hover bool) ui.Color {
					switch variant {
					case Primary:
						if hover {
							return theme.Mix(ui.RGB(255, 255, 255), tokens.Primary, .8)
						}
						return tokens.Primary
					case Secondary, Muted:
						if hover {
							return theme.OKLCH(.95, 0, 0)
						}
						return ui.RGB(255, 255, 255)
					case Tinted:
						l, c := .93, .04
						if dark {
							l = .3
						}
						if hover {
							l, c = .88, .05
							if dark {
								l = .35
							}
						}
						return theme.OKLCH(l, c, 250)
					case Outline:
						if hover && dark {
							return theme.Mix(ui.RGB(255, 255, 255), tokens.Input, .3)
						}
						return ui.RGB(255, 255, 255)
					case Ghost:
						if hover && dark {
							return theme.Mix(ui.RGB(255, 255, 255), tokens.Muted, .5)
						}
						return ui.RGB(255, 255, 255)
					default:
						a := float32(.1)
						if dark {
							a = .2
						}
						if hover {
							a += .1
						}
						return theme.Mix(ui.RGB(255, 255, 255), tokens.Destructive, a)
					}
				}
				for _, hover := range []bool{false, true} {
					if hover {
						r.Move(b.X+b.W/2, b.Y+b.H/2)
					}
					got, want := sample(), expected(hover)
					if abs8(got.R, want.R) > 1 || abs8(got.G, want.G) > 1 || abs8(got.B, want.B) > 1 {
						t.Fatalf("hover=%v got=%v want=%v", hover, got, want)
					}
				}
			})
		}
	}
}
func abs8(a, b uint8) int { return abs(int(a) - int(b)) }
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
func TestBubbleReactionsFourPlacementsRemainPhysicalInRTL(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		for _, side := range []BubbleReactionSide{BubbleReactionBottom, BubbleReactionTop} {
			for _, align := range []BubbleReactionAlign{BubbleReactionEnd, BubbleReactionStart} {
				t.Run(fmt.Sprintf("%d/%d/%d", direction, side, align), func(t *testing.T) {
					r := ui.NewTester(bubbleTestView(func(c *ui.Context) {
						ui.Box(c).Direction(direction).Width(300).Children(func() {
							BubbleWith(c, BubbleProps{}, BubbleSlots{Content: func(c *ui.Context, p BubbleContentProps) {
								BubbleContent(c, p, func() { ui.Text(c, "Face") }).Size(180, 60)
							}, Reactions: func(c *ui.Context) {
								BubbleReactions(c, BubbleReactionsProps{Side: side, Align: align, Label: "Reaction pill"}, func() { ui.Text(c, "👍") })
							}}).Role(ui.RoleGroup).Label("Outer")
						})
					}), 500, 260)
					outer := bubbleNode(t, r, ui.RoleGroup, "Outer").Bounds
					reaction := bubbleNode(t, r, ui.RoleImage, "Reaction pill").Bounds
					wantY := outer.Y + outer.H - reaction.H*.25
					if side == BubbleReactionTop {
						wantY = outer.Y - reaction.H*.75
					}
					wantX := outer.X + outer.W - 12 - reaction.W
					if align == BubbleReactionStart {
						wantX = outer.X + 12
					}
					if math.Abs(float64(reaction.X-wantX)) > .1 || math.Abs(float64(reaction.Y-wantY)) > .1 {
						t.Fatalf("physical reaction placement outer=%+v pill=%+v want=(%v,%v)", outer, reaction, wantX, wantY)
					}
				})
			}
		}
	}
}
func TestBubbleKeyedReactionsKeepFocusAndNativePressedState(t *testing.T) {
	selected := false
	reverse := false
	count := 1
	calls := 0
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		reactions := []Reaction{{Key: "like", Label: "Like", Count: count, Selected: selected, OnSelectedChange: func(on bool) { selected = on; calls++ }}, {Key: "fire", Label: "Fire"}}
		if reverse {
			slices.Reverse(reactions)
		}
		Bubble(c, BubbleProps{Key: "message", Reactions: reactions}, func() { ui.Text(c, "Text") })
	}), 500, 260, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Like 1"); err != nil {
		t.Fatal(err)
	}
	if !selected || calls != 1 || !bubbleNode(t, r, ui.RoleToggleButton, "Like 1").Checked {
		t.Fatal("controlled pressed state not published")
	}
	reverse = true
	count = 7
	r.Frame()
	if !r.Focused("Like 7") {
		t.Fatal("reaction reorder/count update lost native focus")
	}
	r.Key(0, ui.KeySpace)
	if selected || calls != 2 || bubbleNode(t, r, ui.RoleToggleButton, "Like 7").Checked {
		t.Fatal("native keyboard toggle delivery failed")
	}
}
func TestBubbleWrappingGhostWidthAndFiniteCap(t *testing.T) {
	text := strings.Repeat("unbroken", 100)
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Gap(20).Children(func() {
			Bubble(c, BubbleProps{}, func() { ui.Text(c, text) }).Role(ui.RoleGroup).Label("Capped")
			Bubble(c, BubbleProps{Variant: Ghost}, func() { ui.Text(c, text) }).Role(ui.RoleGroup).Label("Ghost")
		})
	}, 400, 800, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	capped, ghost := bubbleNode(t, r, ui.RoleGroup, "Capped").Bounds, bubbleNode(t, r, ui.RoleGroup, "Ghost").Bounds
	if capped.W > 320.1 || capped.H < 80 || ghost.W < 399 || ghost.W > 400.1 || ghost.H < 50 {
		t.Fatalf("source max-width/word wrapping missing: capped=%+v ghost=%+v", capped, ghost)
	}
	for _, limit := range []float32{150, -1, float32(math.NaN()), float32(math.Inf(1))} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			rr := ui.NewTester(bubbleTestView(func(c *ui.Context) {
				Bubble(c, BubbleProps{MaxWidth: limit}, func() { ui.Text(c, text) }).Role(ui.RoleGroup).Label("Finite")
			}), 400, 900)
			n := bubbleNode(t, rr, ui.RoleGroup, "Finite")
			if n.Bounds.W <= 0 || n.Bounds.W > 256.1 || !finiteBubbleRect(n.Bounds) {
				t.Fatalf("bad finite fallback %+v", n)
			}
			if limit == 150 && n.Bounds.W > 150.1 {
				t.Fatal("DIP cap ignored")
			}
			if limit != 150 && !n.Invalid {
				t.Fatal("invalid width not exposed")
			}
		})
	}
}
func finiteBubbleRect(r ui.Rect) bool {
	for _, v := range []float32{r.X, r.Y, r.W, r.H} {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return false
		}
	}
	return true
}

func TestBubbleStreamAppendAndKeyedPrependPreserveNativeEditorSelection(t *testing.T) {
	text := "Short message"
	draft := "draft"
	prepend := false
	initializations := 0
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		BubbleGroup(c, func() {
			if prepend {
				Bubble(c, BubbleProps{Key: "history", Variant: Muted}, func() { ui.Text(c, "Older history") })
			}
			Bubble(c, BubbleProps{Key: "live", Variant: Secondary}, func() {
				ui.Text(c, text)
				input := control.Input(c, &draft, control.InputProps{Label: "Reply editor"})
				ui.Local(input, "mount", func() int { initializations++; return 0 })
			})
		})
	}), 600, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Reply editor"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	text += strings.Repeat(" appended stream content", 20)
	prepend = true
	r.Frame()
	if !r.Focused("Reply editor") || initializations != 1 {
		t.Fatalf("stream/prepend remounted native editor: focus=%v mounts=%d", r.Focused("Reply editor"), initializations)
	}
	r.Type("replacement")
	if draft != "replacement" {
		t.Fatalf("native selection lost during stream update: %q", draft)
	}
}
func TestBubbleDisabledNativeTriggersAndControlsDoNotActivate(t *testing.T) {
	for _, kind := range []BubbleTriggerKind{BubbleButton, BubbleLink} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			disabled := false
			calls := 0
			r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
				Bubble(c, BubbleProps{Disabled: disabled, Trigger: &BubbleTrigger{Kind: kind, Label: "Target", URL: "https://example.com/disabled", OnActivate: func() { calls++ }}}, func() { ui.Text(c, "Target") })
			}), 500, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if err := r.Click("Target"); err != nil {
				t.Fatal(err)
			}
			calls = 0
			disabled = true
			r.Frame()
			n := bubbleNode(t, r, map[BubbleTriggerKind]ui.Role{BubbleButton: ui.RoleButton, BubbleLink: ui.RoleLink}[kind], "Target")
			if !n.Disabled {
				t.Fatal("disabled ancestor not exposed")
			}
			before := len(r.OpenedURLs())
			r.Key(0, ui.KeyEnter)
			r.Key(0, ui.KeySpace)
			r.ClickAt(n.Bounds.X+n.Bounds.W/2, n.Bounds.Y+n.Bounds.H/2)
			if calls != 0 || len(r.OpenedURLs()) != before {
				t.Fatal("disabled native trigger activated")
			}
		})
	}
	draft := ""
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		Bubble(c, BubbleProps{Disabled: true}, func() { control.Input(c, &draft, control.InputProps{Label: "Disabled reply"}) })
	}), 500, 220, ui.TesterOptions{})
	node := bubbleNode(t, r, ui.RoleTextField, "Disabled reply")
	if !node.Disabled {
		t.Fatal("plain face lost inherited disabled state")
	}
	r.ClickAt(node.Bounds.X+8, node.Bounds.Y+8)
	r.Type("blocked")
	if draft != "" {
		t.Fatal("disabled child accepted native typing")
	}
}
func TestBubbleStartEndAlignmentIsLogicalAndGroupsUseSourceGap(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Direction(direction).Children(func() {
					BubbleGroup(c, func() {
						Bubble(c, BubbleProps{}, func() { ui.Text(c, "Start") }).Role(ui.RoleGroup).Label("Start bubble")
						Bubble(c, BubbleProps{Align: End}, func() { ui.Text(c, "End") }).Role(ui.RoleGroup).Label("End bubble")
					}).FillWidth()
				})
			}, 400, 220)
			start, end := bubbleNode(t, r, ui.RoleGroup, "Start bubble").Bounds, bubbleNode(t, r, ui.RoleGroup, "End bubble").Bounds
			if math.Abs(float64(end.Y-start.Y-start.H-8)) > .1 {
				t.Fatalf("group gap differs from source: start=%+v end=%+v", start, end)
			}
			if direction == ui.LeftToRight {
				if start.X != 0 || end.X+end.W != 400 {
					t.Fatalf("logical alignment %+v %+v", start, end)
				}
			} else {
				if end.X != 0 || start.X+start.W != 400 {
					t.Fatalf("RTL logical alignment %+v %+v", start, end)
				}
			}
		})
	}
}
func TestBubbleDisclosureIsControlledAndNativeKeyboardAccessible(t *testing.T) {
	expanded := false
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		Bubble(c, BubbleProps{Expanded: &expanded, Summary: "Preview"}, func() { ui.Text(c, "Complete message") })
	}), 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !r.HasText("Preview") || r.HasText("Complete message") {
		t.Fatal("initial controlled disclosure incorrect")
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeySpace)
	if !expanded || !r.HasText("Complete message") || r.HasText("Preview") {
		t.Fatal("native disclosure activation failed")
	}
	if !bubbleNode(t, r, ui.RoleButton, "Show less").Expanded {
		t.Fatal("disclosure AX state missing")
	}
	r.Key(0, ui.KeyEnter)
	if expanded || !r.HasText("Preview") {
		t.Fatal("native disclosure collapse failed")
	}
}
func TestBubbleReducedMotionIsPixelStableAndIdle(t *testing.T) {
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		Bubble(c, BubbleProps{Variant: Tinted, Trigger: &BubbleTrigger{Label: "Reduced"}}, func() { ui.Text(c, "Tinted button") })
	}), 500, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	b := bubbleNode(t, r, ui.RoleButton, "Reduced").Bounds
	r.Move(b.X+b.W/2, b.Y+b.H/2)
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(time.Second)
	if !bytes.Equal(before, r.Image().Pix) || r.FramePending() {
		t.Fatal("reduced bubble hover kept animating")
	}
	r.SetPreferences(ui.Preferences{})
	r.Move(0, 0)
	r.Advance(40 * time.Millisecond)
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	snap := append([]byte(nil), r.Image().Pix...)
	r.Advance(time.Second)
	if !bytes.Equal(snap, r.Image().Pix) || r.FramePending() {
		t.Fatal("changing reduced preference did not settle transition")
	}
}
func TestBubbleMessageAncestorEndAppliesWithoutLeakingToNextMessage(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			withBubbleMessageAlignment(c, End, func() {
				Bubble(c, BubbleProps{}, func() { ui.Text(c, "From end") }).Role(ui.RoleGroup).Label("End scoped")
			})
			withBubbleMessageAlignment(c, Start, func() {
				Bubble(c, BubbleProps{}, func() { ui.Text(c, "From start") }).Role(ui.RoleGroup).Label("Start scoped")
			})
			Bubble(c, BubbleProps{}, func() { ui.Text(c, "Outside") }).Role(ui.RoleGroup).Label("Outside scoped")
		})
	}, 400, 220)
	a, b, d := bubbleNode(t, r, ui.RoleGroup, "End scoped").Bounds, bubbleNode(t, r, ui.RoleGroup, "Start scoped").Bounds, bubbleNode(t, r, ui.RoleGroup, "Outside scoped").Bounds
	if a.X+a.W != 400 || b.X != 0 || d.X != 0 {
		t.Fatalf("source message alignment scope leaked %+v %+v %+v", a, b, d)
	}
}

var bubbleSolidGlyph = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 10"><rect width="10" height="10" fill="currentColor"/></svg>`))

func TestBubbleQuietVariantsInheritParentTextColor(t *testing.T) {
	for _, variant := range []BubbleVariant{Muted, Outline, Ghost} {
		t.Run(fmt.Sprint(variant), func(t *testing.T) {
			r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
				ui.Column(c).TextColor(ui.RGB(30, 100, 170)).Children(func() {
					Bubble(c, BubbleProps{Variant: variant}, func() { ui.Icon(c, bubbleSolidGlyph).Size(10, 10).Role(ui.RoleImage).Label("Inherited color") })
				})
			}), 500, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			b := bubbleNode(t, r, ui.RoleImage, "Inherited color").Bounds
			got := r.Image().RGBAAt(int(b.X+b.W/2), int(b.Y+b.H/2))
			if got != (color.RGBA{30, 100, 170, 255}) {
				t.Fatalf("source variant overwrote inherited foreground %v", got)
			}
		})
	}
}

func TestBubbleTailwindOpacityMultipliesExistingTokenAlpha(t *testing.T) {
	tokens := theme.Dark()
	tokens.Background = ui.RGB(20, 40, 60)
	tokens.Input = ui.RGBA(220, 180, 100, .2)
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
			Bubble(c, BubbleProps{Variant: Outline, Trigger: &BubbleTrigger{Label: "Alpha face"}}, func() { ui.Box(c).Size(100, 30) })
		})
	}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	b := bubbleNode(t, r, ui.RoleButton, "Alpha face").Bounds
	r.Move(b.X+b.W/2, b.Y+b.H/2)
	got := r.Image().RGBAAt(int(b.X+8), int(b.Y+b.H/2))
	want := tokens.Input.Alpha(.3).Over(tokens.Background)
	if abs8(got.R, want.R) > 1 || abs8(got.G, want.G) > 1 || abs8(got.B, want.B) > 1 {
		t.Fatalf("Tailwind input/30 must multiply token alpha: got=%v want=%v", got, want)
	}
}

func TestBubbleNativeLinkFluentColorOverrideAndNaturalAccessibleName(t *testing.T) {
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		BubbleWith(c, BubbleProps{Variant: Muted}, BubbleSlots{Content: func(c *ui.Context, p BubbleContentProps) {
			p.Trigger = &BubbleTrigger{Kind: BubbleLink, URL: "https://example.com", TextOnly: true}
			BubbleContent(c, p, func() { ui.Text(c, "Natural link name") }).TextColor(ui.RGB(30, 100, 170))
		}})
	}), 500, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	n := bubbleNode(t, r, ui.RoleLink, "Natural link name")
	if n.Bounds.W <= 0 || !n.Focusable {
		t.Fatal("native link lost natural name/focus")
	}
	image := r.Image()
	colored := 0
	for y := int(n.Bounds.Y); y < int(n.Bounds.Y+n.Bounds.H); y++ {
		for x := int(n.Bounds.X); x < int(n.Bounds.X+n.Bounds.W); x++ {
			p := image.RGBAAt(x, y)
			if p.R == 30 && p.G == 100 && p.B == 170 {
				colored++
			}
		}
	}
	if colored == 0 {
		t.Fatal("native link rejected explicit fluent foreground")
	}
}

func TestBubbleLinkUsesNativeRouter(t *testing.T) {
	router := ui.NewRouter("/chat")
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		router.View(c, func(route *ui.Route) {
			if route.Match("/chat") {
				Bubble(c, BubbleProps{Trigger: &BubbleTrigger{Kind: BubbleLink, URL: "/settings", Label: "Open settings"}}, func() { ui.Text(c, "Open settings") })
			} else {
				ui.Text(c, "Settings route")
			}
		})
	}), 500, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Open settings"); err != nil {
		t.Fatal(err)
	}
	if !r.HasText("Settings route") || len(r.OpenedURLs()) != 0 {
		t.Fatalf("bubble link bypassed native router: %v %v", r.Texts(), r.OpenedURLs())
	}
}
func TestBubbleReactionCardRingPaintsOutsideFaceAndDisabledKeepsPressedState(t *testing.T) {
	tokens := theme.Light()
	tokens.Card = ui.RGB(20, 90, 160)
	tokens.Muted = ui.RGB(100, 160, 80)
	calls := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Scope(c, tokens).Fill().Padding(40).Children(func() {
			BubbleWith(c, BubbleProps{Disabled: true}, BubbleSlots{Content: func(c *ui.Context, p BubbleContentProps) {
				BubbleContent(c, p, func() { ui.Text(c, "Text") }).Size(180, 60)
			}, Reactions: func(c *ui.Context) {
				BubbleReactions(c, BubbleReactionsProps{Label: "Reaction", Interactive: true}, func() { BubbleReaction(c, Reaction{Label: "Like", Selected: true, OnClick: func() { calls++ }}) })
			}})
		})
	}, 500, 260, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	n := bubbleNode(t, r, ui.RoleGroup, "Reaction")
	p := r.Image().RGBAAt(int(n.Bounds.X-2), int(n.Bounds.Y+n.Bounds.H/2))
	if p != (color.RGBA{20, 90, 160, 255}) {
		t.Fatalf("reaction card ring was clipped or changed flow: %v", p)
	}
	b := bubbleNode(t, r, ui.RoleToggleButton, "Like")
	if !b.Disabled || !b.Checked {
		t.Fatal("disabled reaction lost inherited disabled/pressed semantics")
	}
	r.ClickAt(b.Bounds.X+b.Bounds.W/2, b.Bounds.Y+b.Bounds.H/2)
	r.Key(0, ui.KeySpace)
	if calls != 0 {
		t.Fatal("disabled reaction activated")
	}
}

func TestBubbleStandaloneReactionHelperHonorsExplicitKey(t *testing.T) {
	reverse := false
	r := ui.NewTesterWith(bubbleTestView(func(c *ui.Context) {
		BubbleWith(c, BubbleProps{}, BubbleSlots{Content: func(c *ui.Context, p BubbleContentProps) { BubbleContent(c, p, func() { ui.Text(c, "Text") }) }, Reactions: func(c *ui.Context) {
			BubbleReactions(c, BubbleReactionsProps{Interactive: true}, func() {
				items := []Reaction{{Key: "like", Label: "Like"}, {Key: "fire", Label: "Fire"}}
				if reverse {
					slices.Reverse(items)
				}
				for _, item := range items {
					BubbleReaction(c, item)
				}
			})
		}})
	}), 500, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Like"); err != nil {
		t.Fatal(err)
	}
	reverse = true
	r.Frame()
	if !r.Focused("Like") {
		t.Fatal("standalone reaction helper ignored explicit key on reorder")
	}
}

func TestBubbleFacesInheritNativeParentFontWeight(t *testing.T) {
	// Neither native ButtonBase nor Link assigns a weight. Like the source's
	// plain/button/anchor faces, all three inherit it from their parent.
	for _, kind := range []int{-1, int(BubbleButton), int(BubbleLink)} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			render := func(reference bool) *ui.Tester {
				return ui.NewTesterWith(func(c *ui.Context) {
					theme.Apply(c)
					ui.Column(c).Fill().FontWeight(700).Children(func() {
						p := BubbleProps{Variant: Ghost}
						if kind >= 0 {
							p.Trigger = &BubbleTrigger{Kind: BubbleTriggerKind(kind), Label: "Face", URL: "https://example.com/font"}
						}
						BubbleWith(c, p, BubbleSlots{Content: func(child *ui.Context, faceProps BubbleContentProps) {
							face := BubbleContent(child, faceProps, func() { ui.Text(child, "MMMM native text") })
							if reference {
								face.FontWeight(700)
							}
						}})
					})
				}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			}
			actual, reference := render(false), render(true)
			if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
				t.Fatal("source face overrode inherited parent font weight")
			}
		})
	}
}

func TestBubbleAdditionalDIPCapPreservesSourcePercentage(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			Bubble(c, BubbleProps{MaxWidth: 600}, func() { ui.Text(c, strings.Repeat("long unbroken message ", 100)) }).Role(ui.RoleGroup).Label("Capped")
		})
	}, 400, 1000, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if width := bubbleNode(t, r, ui.RoleGroup, "Capped").Bounds.W; width > 320.1 {
		t.Fatalf("DIP cap removed source80percent maximum: %g", width)
	}
}

func TestBubblePendingNativePressCancellationDoesNotActivate(t *testing.T) {
	for _, kind := range []BubbleTriggerKind{BubbleButton, BubbleLink} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			disabled := false
			calls := 0
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				Bubble(c, BubbleProps{Disabled: disabled, Trigger: &BubbleTrigger{Kind: kind, Label: "Card", URL: "https://example.com/card", OnActivate: func() { calls++ }}}, func() { ui.Text(c, "Card") })
			}, 400, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			card, ok := r.Find("Card")
			if !ok {
				t.Fatal("missing card")
			}
			x, y := card.X+card.W/2, card.Y+card.H/2
			r.Press(x, y)
			disabled = true
			r.Frame()
			r.Release(x, y)
			if calls != 0 || len(r.OpenedURLs()) != 0 {
				t.Fatal("disabled pending press activated")
			}
			disabled = false
			r.Frame()
			r.Press(x, y)
			r.CancelPress()
			r.Release(x, y)
			if calls != 0 || len(r.OpenedURLs()) != 0 {
				t.Fatal("canceled native press activated")
			}
		})
	}
}

func TestBubbleNarrowLongTextReactionKeepsNativeInputOutsideFace(t *testing.T) {
	calls := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(30).Children(func() {
			Bubble(c, BubbleProps{Variant: Secondary, MaxWidth: 170, Reactions: []Reaction{{Key: "reaction", Label: "Helpful", Count: 17, OnClick: func() { calls++ }}}}, func() { ui.Text(c, strings.Repeat("native content ", 25)) }).Role(ui.RoleGroup).Label("Face")
		})
	}, 260, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	face := bubbleNode(t, r, ui.RoleGroup, "Face")
	reaction := bubbleNode(t, r, ui.RoleToggleButton, "Helpful 17")
	if face.Bounds.W > 170.1 || reaction.Bounds.Y+reaction.Bounds.H <= face.Bounds.Y+face.Bounds.H {
		t.Fatal("long-text reaction did not keep source overflow placement")
	}
	if err := r.Click("Helpful 17"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeySpace)
	if calls != 2 {
		t.Fatalf("outside-face native reaction did not deliver pointer/keyboard once: %d", calls)
	}
}
