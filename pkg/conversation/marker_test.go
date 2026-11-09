package conversation

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestMarkerDefaultIsSourceStaticWithoutLiveStatus(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); Marker(c, "Today", MarkerProps{}) }, 400, 180)
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleStatus || node.Role == ui.RoleButton || node.Role == ui.RoleLink {
			t.Fatal("source default marker became a live or interactive control")
		}
	}
	if !r.HasText("Today") {
		t.Fatal("static marker text missing")
	}
}

func markerAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	node, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing role%d %q", role, name)
	}
	return node
}

func TestMarkerSourceDefaultAndBorderUse14pxInheritedFontPixels(t *testing.T) {
	for _, variant := range []MarkerVariant{Inline, Bordered} {
		t.Run(fmt.Sprint(variant), func(t *testing.T) {
			render := func(marker bool) *ui.Tester {
				return ui.NewTester(func(c *ui.Context) {
					tokens := theme.Light()
					tokens.MutedForeground = ui.RGB(20, 90, 160)
					native := tokens.Theme()
					native.Font = "monospace"
					ui.Themed(c, native).Fill().FontWeight(700).Children(func() {
						if marker {
							Marker(c, "Native text", MarkerProps{Variant: variant})
						} else {
							row := ui.Row(c).FillWidth().MinHeight(16).Gap(8).AlignItems(ui.Center).FontSize(14).LineHeight(10.0 / 7).TextColor(tokens.MutedForeground)
							if variant == Bordered {
								row.BorderWidth(0, 0, 1, 0).BorderColor(tokens.Border).Padding(0, 0, 8, 0)
							}
							row.Children(func() { ui.Text(c, "Native text") })
						}
					})
				}, 400, 140)
			}
			actual, reference := render(true), render(false)
			if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
				t.Fatal("marker replaced source typography or native theme/token inheritance")
			}
		})
	}
}

func TestMarkerEmptyRootAndDecorativeIconUseSourceMinimum(t *testing.T) {
	for _, icon := range []bool{false, true} {
		t.Run(fmt.Sprint(icon), func(t *testing.T) {
			var bounds ui.Rect
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				MarkerWith(c, MarkerProps{}, func(child *ui.Context) {
					if icon {
						MarkerIcon(child, traceChevronIcon)
					}
				}).Draw(func(_ *ui.Painter, r ui.Rect) { bounds = r })
			}, 300, 100)
			if bounds.H != 16 || bounds.W != 300 {
				t.Fatalf("source minimum/root width lost: %v", bounds)
			}
			for _, node := range r.AccessAll() {
				if node.Role == ui.RoleImage || node.Focusable {
					t.Fatal("decorative marker icon became an AX/input control")
				}
			}
		})
	}
}

func TestMarkerSeparatorCentersContentAndPaintsPhysicalLinesInRTL(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			tokens := theme.Light()
			tokens.Border = ui.RGB(210, 20, 80)
			var content, root ui.Rect
			r := ui.NewTester(func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
					MarkerWith(c, MarkerProps{Variant: Separator}, func(child *ui.Context) {
						MarkerContent(child, func(text *ui.Context) { MarkerText(text, "Worked for42s") }).Draw(func(_ *ui.Painter, r ui.Rect) { content = r })
					}).Draw(func(_ *ui.Painter, r ui.Rect) { root = r })
				})
			}, 400, 140)
			r.SetDirection(direction)
			if math.Abs(float64(content.X+content.W/2-200)) > .1 {
				t.Fatalf("source separator content not centered: %v", content)
			}
			image := r.Image()
			y := int(root.Y + root.H/2)
			want := color.RGBA{tokens.Border.R, tokens.Border.G, tokens.Border.B, 255}
			if image.RGBAAt(10, y) != want || image.RGBAAt(390, y) != want || root.H != 20 {
				t.Fatal("source separator lines/minimum text geometry missing")
			}
		})
	}
}

func TestMarkerNativeRootButtonAndAnchorKeepKeyboardAndDisabledSemantics(t *testing.T) {
	for _, kind := range []MarkerTriggerKind{MarkerButton, MarkerAnchor} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			disabled := false
			calls := 0
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				MarkerWith(c, MarkerProps{Disabled: disabled, Trigger: &MarkerTrigger{Kind: kind, Label: "Marker action", URL: "https://example.com/marker", OnActivate: func() { calls++ }}}, func(child *ui.Context) {
					MarkerIcon(child, traceChevronIcon)
					MarkerContent(child, func(text *ui.Context) { MarkerText(text, "Marker action") })
				})
			}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if err := r.Click("Marker action"); err != nil {
				t.Fatal(err)
			}
			r.Key(0, ui.KeyEnter)
			r.Key(0, ui.KeySpace)
			want := 3
			if kind == MarkerAnchor {
				want = 2
				if len(r.OpenedURLs()) != 2 {
					t.Fatal("actual marker anchor did not open native URL exactly twice")
				}
			}
			if calls != want {
				t.Fatalf("native delivery=%d want%d", calls, want)
			}
			disabled = true
			r.Frame()
			before := calls
			urls := len(r.OpenedURLs())
			r.Key(0, ui.KeyEnter)
			r.Key(0, ui.KeySpace)
			role := ui.RoleButton
			if kind == MarkerAnchor {
				role = ui.RoleLink
			}
			node := markerAccess(t, r, role, "Marker action")
			r.ClickAt(node.Bounds.X+10, node.Bounds.Y+node.Bounds.H/2)
			if !node.Disabled || calls != before || len(r.OpenedURLs()) != urls {
				t.Fatal("disabled marker activated through native input")
			}
		})
	}
}

func TestMarkerStatusIsExplicitAndLeavesLiveAnnouncementsToNativeRole(t *testing.T) {
	text := "Checking logs"
	r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); Marker(c, text, MarkerProps{Status: true}) }, 400, 120)
	markerAccess(t, r, ui.RoleStatus, text)
	r.Announcements()
	text = "Logs checked"
	r.Frame()
	markerAccess(t, r, ui.RoleStatus, text)
	// Tester.Announcements only observes Context.Announce; actual RoleStatus
	// updates are delivered by native accessibility, without a second explicit
	// announcement that would read the same text twice with a screen reader.
	r.Frame()
	if got := r.Announcements(); len(got) != 0 {
		t.Fatalf("idle status repeated announcement: %q", got)
	}
}

func TestMarkerShimmerUsesNativeGlyphsAndStopsWhenUnset(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			active := true
			r := ui.NewTester(func(c *ui.Context) {
				tokens := theme.Light()
				if dark {
					tokens = theme.Dark()
				}
				theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() { Marker(c, "Native 字形 working", MarkerProps{Shimmer: active, Status: true}) })
			}, 400, 140)
			before := append([]byte(nil), r.Image().Pix...)
			r.Advance(time.Second)
			if bytes.Equal(before, r.Image().Pix) || !r.FramePending() {
				t.Fatal("marker shimmer did not affect actual native glyphs")
			}
			markerAccess(t, r, ui.RoleStatus, "Native 字形 working")
			active = false
			r.Frame()
			stable := append([]byte(nil), r.Image().Pix...)
			r.Advance(3 * time.Second)
			if !bytes.Equal(stable, r.Image().Pix) || r.FramePending() {
				t.Fatal("inactive marker retained shimmer work")
			}
		})
	}
}

func TestMarkerArbitraryNativeContentComposesWithAccordionAndDrawer(t *testing.T) {
	open, drawer := false, false
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		display.CollapsibleWith(c, &open, display.CollapsibleProps{Label: "Accordion marker", Trigger: func(trigger ui.Element) { trigger.Children(func() { Marker(c, "Worked for42s", MarkerProps{}) }) }, Content: func() { ui.Text(c, "Native accordion details") }})
		MarkerWith(c, MarkerProps{Variant: Separator}, func(child *ui.Context) {
			if control.Button(child, "Explore files", control.ButtonProps{Variant: control.Outline}).Clicked() {
				drawer = true
			}
		})
		overlay.Drawer(c, &drawer, overlay.DrawerProps{Title: "File activity", Side: overlay.Right}, func() { ui.Text(c, "Four files") })
	}, 400, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Accordion marker"); err != nil {
		t.Fatal(err)
	}
	if !open || !r.HasText("Native accordion details") {
		t.Fatal("passive marker interfered with native disclosure input")
	}
	if err := r.Click("Explore files"); err != nil {
		t.Fatal(err)
	}
	markerAccess(t, r, ui.RoleDialog, "File activity")
	r.Key(0, ui.KeyEscape)
	if !r.Focused("Explore files") {
		t.Fatal("native drawer did not return focus to custom marker content")
	}
}

func TestMarkerBorderUsesOnlySourceBottomBorderAndPadding(t *testing.T) {
	var bounds ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Marker(c, "Border marker", MarkerProps{Variant: Bordered}).Draw(func(_ *ui.Painter, r ui.Rect) { bounds = r })
	}, 400, 180)
	text, _ := r.Find("Border marker")
	if bounds.H != 29 || text.X != 0 {
		t.Fatalf("source border variant should be20+8+1 tall without inset: bounds=%v text=%v", bounds, text)
	}
}

func TestMarkerKeyedPrependEmptyIconAndContentKeepNativeEditorIdentity(t *testing.T) {
	prepend, icon := false, false
	draft := "draft"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).FillWidth().Children(func() {
			if prepend {
				Marker(c, "Earlier", MarkerProps{Key: "earlier"})
			}
			MarkerWith(c, MarkerProps{Key: "live"}, func(child *ui.Context) {
				if icon {
					MarkerIcon(child.Key("icon"), nil, MarkerIconProps{Content: func(*ui.Context) {}})
				}
				MarkerContent(child.Key("content"), func(text *ui.Context) {
					control.Input(text, &draft, control.InputProps{Label: "Marker draft"}).Width(180)
				})
			})
		})
	}, 400, 200)
	if err := r.Click("Marker draft"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	prepend, icon = true, true
	r.Frame()
	r.Type("replacement")
	if draft != "replacement" || !r.Focused("Marker draft") {
		t.Fatal("marker keyed prepend or empty icon changed native selection")
	}
	r.Compose("文", 1)
	icon = false
	r.Frame()
	r.Type("文")
	if draft != "replacement文" {
		t.Fatalf("marker slot toggle lost native IME composition: %q", draft)
	}
}

func TestMarkerLongTextWrapsAndUsesNativeSelectionInBothDirections(t *testing.T) {
	text := strings.Repeat("Long native text 字形 ", 8)
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			var root, content ui.Rect
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				MarkerWith(c, MarkerProps{}, func(child *ui.Context) {
					MarkerIcon(child, traceChevronIcon)
					MarkerContent(child, func(textContext *ui.Context) { MarkerText(textContext, text).Selectable() }).Draw(func(_ *ui.Painter, r ui.Rect) { content = r })
				}).Draw(func(_ *ui.Painter, r ui.Rect) { root = r })
			}, 190, 300)
			r.SetDirection(direction)
			if root.W != 190 || root.H <= 20 || content.W > 166 || content.X < 0 || content.X+content.W > 190.1 {
				t.Fatalf("long marker did not wrap inside native parent: root=%v content=%v", root, content)
			}
			if err := r.Click(text); err != nil {
				t.Fatal(err)
			}
			r.Key(ui.Cmd, ui.KeyA)
			r.Key(ui.Cmd, ui.KeyC)
			if r.Clipboard() != text {
				t.Fatal("marker replaced native selectable text")
			}
		})
	}
}

func TestMarkerAnchorsKeepNativeRouterAndWholeCardTargetWithNaturalAndFluentName(t *testing.T) {
	for _, fluent := range []bool{false, true} {
		for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
			t.Run(fmt.Sprint(fluent, direction), func(t *testing.T) {
				router := ui.NewRouter("/activity")
				var handle ui.Handle
				r := ui.NewTesterWith(func(c *ui.Context) {
					theme.Apply(c)
					router.View(c, func(route *ui.Route) {
						if !route.Match("/activity") {
							ui.Text(c, "Native settings page")
							return
						}
						marker := MarkerWith(c, MarkerProps{Key: "route", Trigger: &MarkerTrigger{Kind: MarkerAnchor, URL: "/settings", Handle: &handle}}, func(child *ui.Context) {
							MarkerIcon(child, nil, MarkerIconProps{Content: func(icon *ui.Context) {
								MarkerText(icon, "Decorative")
							}})
							MarkerContent(child, func(content *ui.Context) {
								MarkerText(content, "Open settings")
								ui.Box(content).Size(50, 10)
							})
						})
						if fluent {
							marker.Label("Fluent settings")
						}
					})
				}, 400, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				r.SetDirection(direction)
				name := "Open settings"
				if fluent {
					name = "Fluent settings"
				}
				link := markerAccess(t, r, ui.RoleLink, name)
				if link.Bounds.W != 400 || !link.Focusable {
					t.Fatalf("anchor target lost full card/focus: %+v", link)
				}
				handle.Focus()
				r.Frame()
				if !r.Focused(name) {
					t.Fatal("marker Handle did not bind actual native Link")
				}
				r.Key(0, ui.KeySpace)
				if !r.HasText("Open settings") {
					t.Fatal("Space activated anchor")
				}
				r.ClickAt(link.Bounds.X+link.Bounds.W-2, link.Bounds.Y+link.Bounds.H/2)
				if !r.HasText("Native settings page") || len(r.OpenedURLs()) != 0 {
					t.Fatal("full-card anchor bypassed actual native router")
				}
			})
		}
	}
}

func TestMarkerPersistentRootAndInlineLinkUnderlineHasSourceThreeDIPOffset(t *testing.T) {
	for _, rootLink := range []bool{false, true} {
		t.Run(fmt.Sprint(rootLink), func(t *testing.T) {
			render := func(marker bool, offset float32) *ui.Tester {
				return ui.NewTester(func(c *ui.Context) {
					tokens := theme.Light()
					theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
						if marker && rootLink {
							Marker(c, "Source underline", MarkerProps{Trigger: &MarkerTrigger{Kind: MarkerAnchor, URL: "https://example.com"}})
							return
						}
						row := ui.Row(c).FillWidth().MinHeight(16).Gap(8).AlignItems(ui.Center).FontSize(14).LineHeight(10.0 / 7).TextColor(tokens.MutedForeground)
						if rootLink {
							row.Underline().UnderlineOffset(offset)
						}
						row.Children(func() {
							if rootLink {
								ui.Text(c, "Source underline")
							} else if marker {
								MarkerLink(c, "Source underline", "https://example.com")
							} else {
								ui.Link(c, "Source underline", "https://example.com").TextColor(tokens.MutedForeground).Underline().UnderlineOffset(offset)
							}
						})
					})
				}, 300, 100)
			}
			actual, reference, wrong := render(true, 3), render(false, 3), render(false, 1)
			if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
				t.Fatal("persistent native marker underline differs from explicit3DIP reference before hover")
			}
			if bytes.Equal(reference.Image().Pix, wrong.Image().Pix) {
				t.Fatal("underline assertion failed to distinguish a wrong offset")
			}
		})
	}
}

func TestMarkerMixedNativeRichTextLinksAndVerticalSlotsRemainRealCompositions(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).FillWidth().Gap(16).Children(func() {
			MarkerWith(c, MarkerProps{Variant: Separator}, func(child *ui.Context) {
				MarkerContent(child, func(content *ui.Context) {
					ui.RichText(content).Children(func() {
						ui.Text(content, "Preview ")
						MarkerLink(content, "example.com", "https://example.com/preview")
						ui.Text(content, " is ready.").FontWeight(500)
					})
				})
			})
			MarkerWith(c, MarkerProps{}, func(child *ui.Context) {
				MarkerIcon(child, nil, MarkerIconProps{Content: func(icon *ui.Context) { display.Spinner(icon) }})
				MarkerContent(child, func(content *ui.Context) { MarkerText(content, "Review completed") })
			}).Column().AlignItems(ui.Center).Role(ui.RoleGroup).Label("Vertical marker")
			Marker(c, "Checking files", MarkerProps{Busy: true})
		})
	}, 420, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("example.com"); err != nil {
		t.Fatal(err)
	}
	if urls := r.OpenedURLs(); len(urls) != 1 || urls[0] != "https://example.com/preview" {
		t.Fatalf("rich inline marker link lost native target: %v", urls)
	}
	vertical := markerAccess(t, r, ui.RoleGroup, "Vertical marker").Bounds
	text, ok := r.Find("Review completed")
	if !ok || vertical.H != 44 || math.Abs(float64(text.X+text.W/2-210)) > .1 || text.Y-vertical.Y != 24 {
		t.Fatalf("source vertical consumer override failed: root=%v text=%v", vertical, text)
	}
	if !markerAccess(t, r, ui.RoleStatus, "Checking files").Busy {
		t.Fatal("explicit Busy marker lost native busy semantics")
	}
}
