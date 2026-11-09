package display

import (
	"fmt"
	"image/color"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestItemDefaultVegaInsetsAndDescriptionClamp(t *testing.T) {
	r := displayTester(func(c *ui.Context) {
		Item(c, ItemProps{Title: "Document", Description: "A useful description.", Leading: func() {
			ui.Box(c).Size(16, 16).Label("Item media bounds")
		}})
	})
	media, ok := r.Find("Item media bounds")
	if !ok || media.X != 37 || media.Y != 37 {
		t.Fatalf("default item does not use Vega horizontal16/vertical14 plus media offset2: %+v", media)
	}
}

func itemAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	info, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q: %+v", role, label, r.AccessAll())
	}
	return info
}

func TestItemVariantSizeMediaCoverPixelsAndRTL(t *testing.T) {
	for _, size := range []ItemSize{ItemDefaultSize, ItemSmall, ItemExtraSmall} {
		for _, variant := range []ItemVariant{ItemDefault, ItemOutline, ItemMuted} {
			t.Run(fmt.Sprintf("size=%d/variant=%d", size, variant), func(t *testing.T) {
				var tokens *theme.Tokens
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					tokens = theme.TokensOf(c)
					ui.Column(c).Fill().Padding(20).Background(tokens.Background).Children(func() {
						Item(c, ItemProps{Label: "File item", Title: "Document", Description: "Two versions available.", Size: size, Variant: variant,
							Media:   &ItemMediaProps{Variant: ItemMediaImage, Image: avatarPicture(), Label: "Document cover"},
							Actions: func(child *ui.Context) { ui.Button(child, "Download") }})
					})
				}, 500, 250)
				root := itemAccess(t, r, ui.RoleGroup, "File item")
				media := itemAccess(t, r, ui.RoleImage, "Document cover")
				px, py, _ := itemMetrics(size)
				want := map[ItemSize]float32{ItemDefaultSize: 40, ItemSmall: 32, ItemExtraSmall: 24}[size]
				if media.Bounds.X != root.Bounds.X+px+1 || media.Bounds.Y != root.Bounds.Y+py+3 || media.Bounds.W != want || media.Bounds.H != want {
					t.Fatalf("media metrics root=%+v media=%+v", root.Bounds, media.Bounds)
				}
				left := r.Image().RGBAAt(int(media.Bounds.X+want/4), int(media.Bounds.Y+want/2))
				right := r.Image().RGBAAt(int(media.Bounds.X+3*want/4), int(media.Bounds.Y+want/2))
				if left != (color.RGBA{R: 220, A: 255}) || right != (color.RGBA{B: 220, A: 255}) {
					t.Fatalf("item image not painted using cover fit: left=%v right=%v", left, right)
				}
				if variant == ItemOutline {
					pixel := r.Image().RGBAAt(int(root.Bounds.X), int(root.Bounds.Y+root.Bounds.H/2))
					if pixel.R != tokens.Border.R || pixel.G != tokens.Border.G || pixel.B != tokens.Border.B {
						t.Fatalf("outline edge=%v token=%v", pixel, tokens.Border)
					}
				}
				if variant == ItemMuted {
					pixel := r.Image().RGBAAt(int(root.Bounds.X+3), int(root.Bounds.Y+root.Bounds.H/2))
					if pixel.R == tokens.Background.R && pixel.G == tokens.Background.G && pixel.B == tokens.Background.B {
						t.Fatal("muted variant did not paint its background")
					}
				}
				r.SetDirection(ui.RightToLeft)
				media = itemAccess(t, r, ui.RoleImage, "Document cover")
				action := itemAccess(t, r, ui.RoleButton, "Download")
				if media.Bounds.X <= action.Bounds.X {
					t.Fatalf("RTL media/action order not mirrored: media=%+v action=%+v", media.Bounds, action.Bounds)
				}
			})
		}
	}
}

func TestItemDescriptionClampsTwoLinesAndActionsWrap(t *testing.T) {
	for _, size := range []ItemSize{ItemDefaultSize, ItemSmall, ItemExtraSmall} {
		description := strings.Repeat("A detailed item description. ", 20)
		r := ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			ui.Column(c).Width(200).Padding(8).Children(func() {
				Item(c, ItemProps{Label: "Wrapped item", Title: strings.Repeat("Long title ", 20), Description: description, Size: size,
					Actions: func(child *ui.Context) { ui.Button(child, "Long action").Width(130) }})
			})
		}, 230, 300)
		desc, ok := r.Find(description)
		if !ok || desc.H > 42 || desc.W <= 0 {
			t.Fatalf("description did not clamp to two lines size=%d: %+v", size, desc)
		}
		root := itemAccess(t, r, ui.RoleGroup, "Wrapped item")
		action := itemAccess(t, r, ui.RoleButton, "Long action")
		if action.Bounds.Y <= desc.Y || action.Bounds.X+action.Bounds.W > root.Bounds.X+root.Bounds.W {
			t.Fatalf("narrow item did not wrap action inside item: root=%+v desc=%+v action=%+v", root.Bounds, desc, action.Bounds)
		}
	}
}

func TestItemPrimaryButtonHasIndependentActionsHeaderFooterAndEditor(t *testing.T) {
	activated, changed, shares, header, footer := 0, 0, 0, 0, 0
	draft := "Draft"
	p := ItemProps{Title: "Open project", Description: "Read the project", Primary: &ItemPrimary{OnActivate: func() { activated++ }},
		Header: func(child *ui.Context) {
			if ui.Button(child, "Pin").Clicked() {
				header++
			}
		},
		Actions: func(child *ui.Context) {
			if ui.Button(child, "Share").Clicked() {
				shares++
			}
			ui.TextInput(child, &draft).Label("Rename project").Width(120)
		},
		Footer: func(child *ui.Context) {
			if ui.Button(child, "Archive").Clicked() {
				footer++
			}
		},
	}
	r := displayTester(func(c *ui.Context) {
		if Item(c, p).Changed() {
			changed++
		}
	})
	for _, label := range []string{"Pin", "Share", "Archive"} {
		displayClick(t, r, label)
	}
	if activated != 0 || changed != 0 || header != 1 || shares != 1 || footer != 1 {
		t.Fatalf("secondary controls activated parent: primary=%d changed=%d header=%d share=%d footer=%d", activated, changed, header, shares, footer)
	}
	displayClick(t, r, "Rename project")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Edited")
	if draft != "Edited" || activated != 0 {
		t.Fatal("action editor lost native input or activated primary")
	}
	displayClick(t, r, "Open project")
	r.Key(0, ui.KeySpace)
	r.Key(0, ui.KeyEnter)
	if activated != 3 || changed != 3 {
		t.Fatalf("native primary button event counts: activated=%d changed=%d", activated, changed)
	}
	itemAccess(t, r, ui.RoleButton, "Open project")
	p.Disabled = true
	r.Frame()
	if !itemAccess(t, r, ui.RoleButton, "Share").Disabled || !itemAccess(t, r, ui.RoleTextField, "Rename project").Disabled {
		t.Fatal("disabled item left sibling actions interactive")
	}
	r.Key(0, ui.KeyEnter)
	if activated != 3 {
		t.Fatal("disabled item primary still activated")
	}
	p.Disabled, p.Primary.Disabled = false, true
	r.Frame()
	displayClick(t, r, "Share")
	if shares != 2 || !itemAccess(t, r, ui.RoleButton, "Open project").Disabled {
		t.Fatal("primary-only disabled state disabled independent action")
	}
}

func TestItemPrimaryLinkUsesNativeURLAndRouterEnterOnly(t *testing.T) {
	activated, changed := 0, 0
	r := displayTester(func(c *ui.Context) {
		if Item(c, ItemProps{Title: "Documentation", Description: "Learn more", Primary: &ItemPrimary{Kind: ItemPrimaryLink, URL: "https://example.com/docs", OnActivate: func() { activated++ }}}).Changed() {
			changed++
		}
	})
	displayClick(t, r, "Documentation")
	r.Key(0, ui.KeySpace)
	if activated != 1 || changed != 1 || len(r.OpenedURLs()) != 1 {
		t.Fatalf("Space activated native link: calls=%d changes=%d URLs=%v", activated, changed, r.OpenedURLs())
	}
	r.Key(0, ui.KeyEnter)
	if activated != 2 || changed != 2 || len(r.OpenedURLs()) != 2 {
		t.Fatalf("Enter lost URL/action delivery: %d %d %v", activated, changed, r.OpenedURLs())
	}
	itemAccess(t, r, ui.RoleLink, "Documentation")
	router := ui.NewRouter("/start")
	rr := displayTester(func(c *ui.Context) {
		router.View(c, func(route *ui.Route) {
			if route.Match("/start") {
				Item(c, ItemProps{Title: "Open settings", Primary: &ItemPrimary{Kind: ItemPrimaryLink, URL: "/settings"}})
			} else {
				ui.Text(c, "Settings route")
			}
		})
	})
	displayClick(t, rr, "Open settings")
	if !rr.HasText("Settings route") || len(rr.OpenedURLs()) != 0 {
		t.Fatalf("link bypassed native router: texts=%v URLs=%v", rr.Texts(), rr.OpenedURLs())
	}
}

func TestItemGroupKeysKeepActionEditorStateFocusAndListSemantics(t *testing.T) {
	starts := map[string]int{}
	action := func(key string) func(*ui.Context) {
		return func(c *ui.Context) {
			owner := ui.Box(c)
			draft := ui.Local(owner, "draft", func() string { starts[key]++; return "Draft " + key })
			owner.Children(func() { ui.TextInput(c, draft).Label("Editor " + key).Width(130) })
		}
	}
	p := ItemGroupProps{Label: "Files", Separators: true, Items: []ItemProps{
		{Key: "remaining", Title: "Alpha", Actions: action("Alpha")},
		{Key: "separator", Title: "Beta", Size: ItemExtraSmall, Actions: action("Beta")},
	}}
	r := displayTester(func(c *ui.Context) { ItemGroup(c, p) })
	itemAccess(t, r, ui.RoleList, "Files")
	a := itemAccess(t, r, ui.RoleListItem, "Alpha")
	b := itemAccess(t, r, ui.RoleListItem, "Beta")
	if a.PosInSet != 1 || a.SetSize != 2 || b.PosInSet != 2 {
		t.Fatalf("list positions alpha=%+v beta=%+v", a, b)
	}
	displayClick(t, r, "Editor Alpha")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Saved draft")
	slices.Reverse(p.Items)
	r.Frame()
	if !r.Focused("Editor Alpha") || itemAccess(t, r, ui.RoleTextField, "Editor Alpha").Value != "Saved draft" || starts["Alpha"] != 1 || starts["Beta"] != 1 {
		t.Fatalf("reorder remounted state/focus: starts=%v focus=%v input=%+v", starts, r.Focused("Editor Alpha"), itemAccess(t, r, ui.RoleTextField, "Editor Alpha"))
	}
	a = itemAccess(t, r, ui.RoleListItem, "Alpha")
	if a.PosInSet != 2 {
		t.Fatal("reorder left stale native list position")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("static item group kept requesting frames")
	}
}

func TestItemEmptyCallbacksDoNotLeakKeyToFollowingEditor(t *testing.T) {
	for _, slot := range []string{"header", "media", "content", "actions", "footer"} {
		t.Run(slot, func(t *testing.T) {
			show, draft := false, "Outside draft"
			r := displayTester(func(c *ui.Context) {
				build := func(child *ui.Context) {
					if show {
						ui.Text(child, "Optional content")
					}
				}
				p := ItemProps{}
				switch slot {
				case "header":
					p.Header = build
				case "media":
					p.Media = &ItemMediaProps{Content: build}
				case "content":
					p.Content = build
				case "actions":
					p.Actions = build
				case "footer":
					p.Footer = build
				}
				Item(c, p)
				ui.TextInput(c, &draft).Label("Following editor")
			})
			displayClick(t, r, "Following editor")
			show = true
			r.Frame()
			if !r.Focused("Following editor") {
				t.Fatal("empty callback leaked identity into following editor")
			}
		})
	}
}

func TestItemNativeFocusRingInvalidAndReducedMotionDoNotIdle(t *testing.T) {
	activated := 0
	var tokens *theme.Tokens
	var primary ui.Handle
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		tokens = theme.TokensOf(c)
		ui.Column(c).Fill().Padding(20).Background(tokens.Background).Children(func() {
			Item(c, ItemProps{Label: "Keyboard item", Title: "Open", Invalid: true,
				Primary: &ItemPrimary{Handle: &primary, OnActivate: func() { activated++ }}})
		})
	}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Key(0, ui.KeyTab)
	a := itemAccess(t, r, ui.RoleButton, "Keyboard item")
	if !a.Focused || !a.Invalid {
		t.Fatalf("primary did not expose native focus/invalid state: %+v", a)
	}
	root := itemAccess(t, r, ui.RoleGroup, "Keyboard item")
	pixel := r.Image().RGBAAt(int(root.Bounds.X-2), int(root.Bounds.Y+root.Bounds.H/2))
	if pixel.R == tokens.Background.R && pixel.G == tokens.Background.G && pixel.B == tokens.Background.B {
		t.Fatal("keyboard focus did not paint the Item outer focus ring")
	}
	r.Move(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Press(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.CancelPress()
	if activated != 0 || r.FramePending() {
		t.Fatalf("cancelled or reduced-motion item kept an action/animation: activated=%d idle=%v", activated, r.FramePending())
	}
}

func TestItemOverwideActionsScrollInsideSurfaceAndRemainReachable(t *testing.T) {
	clicks := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(180).Padding(8).Children(func() {
			Item(c, ItemProps{Label: "Narrow item", Title: "Document", Actions: func(child *ui.Context) {
				button := ui.Button(child, "Wide action").Width(190)
				button.Children(func() { ui.Box(child).Absolute().Right(0).Top(0).Bottom(0).Width(10).Background(ui.RGB(0, 0, 255)) })
				if button.Clicked() {
					clicks++
				}
			}})
		})
	}, 240, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	root := itemAccess(t, r, ui.RoleGroup, "Narrow item")
	action := itemAccess(t, r, ui.RoleButton, "Wide action")
	var viewport ui.Rect
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleScroll {
			viewport = node.Bounds
			break
		}
	}
	// Native accessibility preserves a scrolled child's full content bounds.
	// Its viewport, pixels and pointer clipping establish the visible limit.
	if viewport.W <= 0 || viewport.X < root.Bounds.X || viewport.X+viewport.W > root.Bounds.X+root.Bounds.W {
		t.Fatalf("fixed action viewport escaped the surface: frame=%+v viewport=%+v", root.Bounds, viewport)
	}
	bluePixels := func() int {
		im, count := r.Image(), 0
		for y := 0; y < im.Bounds().Dy(); y++ {
			for x := 0; x < im.Bounds().Dx(); x++ {
				if im.RGBAAt(x, y) == (color.RGBA{B: 255, A: 255}) {
					count++
				}
			}
		}
		return count
	}
	before := bluePixels()
	r.Scroll(viewport.X+viewport.W/2, viewport.Y+viewport.H/2, 100, 0)
	r.Advance(time.Second)
	if after := bluePixels(); after <= before {
		t.Fatalf("bounded actions were clipped without native scrolling: before=%d after=%d", before, after)
	}
	for y := 0; y < r.Image().Bounds().Dy(); y++ {
		for x := int(root.Bounds.X + root.Bounds.W); x < r.Image().Bounds().Dx(); x++ {
			if r.Image().RGBAAt(x, y) == (color.RGBA{B: 255, A: 255}) {
				t.Fatal("fixed action painted beyond its bounded viewport")
			}
		}
	}
	r.ClickAt(viewport.X+viewport.W+5, viewport.Y+viewport.H/2)
	if clicks != 0 {
		t.Fatal("fixed action accepted a hit outside its scrolling viewport")
	}
	r.ClickAt(viewport.X+viewport.W/2, viewport.Y+viewport.H/2)
	if clicks != 1 {
		t.Fatal("scrolling fixed action made it unreachable")
	}
	if !r.Focused("Wide action") || action.Bounds.W != 190 {
		t.Fatal("bounded viewport replaced the actual native action control")
	}
	r.Key(0, ui.KeySpace)
	if clicks != 2 {
		t.Fatal("bounded action lost native keyboard activation")
	}
}

func TestItemNamedCustomMediaKeepsControlsAndDecorativeMediaSkipsStops(t *testing.T) {
	for _, decorative := range []bool{false, true} {
		r := displayTester(func(c *ui.Context) {
			ui.Button(c, "Before media")
			ItemMedia(c, ItemMediaProps{Label: "Team illustration", Decorative: decorative, Content: func(child *ui.Context) {
				ui.Button(child, "Illustration action")
			}})
			ui.Button(c, "After media")
		})
		r.Key(0, ui.KeyTab)
		r.Key(0, ui.KeyTab)
		if decorative && !r.Focused("After media") || !decorative && !r.Focused("Illustration action") {
			t.Fatal("named/decorative media did not preserve its intended keyboard contract")
		}
		if _, ok := r.AccessByRole(ui.RoleButton, "Illustration action"); ok == decorative {
			t.Fatal("named/decorative media hid a live control or exposed decoration")
		}
		_, named := r.AccessByRole(ui.RoleGroup, "Team illustration")
		if named == decorative {
			t.Fatal("media did not expose its intended named/decorative semantics")
		}
	}
}
