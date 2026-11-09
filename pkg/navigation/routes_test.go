package navigation

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func routeTester(view func(*ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(12).Children(func() { view(c) })
	}, 900, 480)
}
func routeClick(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	if err := r.Click(label); err != nil {
		t.Fatal(err)
	}
}
func routeAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	node, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q: %+v", role, label, r.AccessAll())
	}
	return node
}

func TestBreadcrumbCurrentPageIsInertNativeLink(t *testing.T) {
	r := routeTester(func(c *ui.Context) {
		Breadcrumb(c, []BreadcrumbItem{{Key: "home", Label: "Home"}, {Key: "current", Label: "Current", URL: "https://example.com/current"}})
	})
	page := routeAccess(t, r, ui.RoleLink, "Current")
	if !page.Disabled || page.Focusable || page.Current != ui.CurrentPage {
		t.Fatalf("current crumb semantics=%+v", page)
	}
	r.ClickAt(page.Bounds.X+page.Bounds.W/2, page.Bounds.Y+page.Bounds.H/2)
	if len(r.OpenedURLs()) != 0 {
		t.Fatal("current crumb navigated")
	}
}

func TestBreadcrumbCompoundHelpersUseNativeURLAndEmptyScopes(t *testing.T) {
	show := false
	draft := "draft"
	r := routeTester(func(c *ui.Context) {
		BreadcrumbNav(c, "Custom trail", func(ctx *ui.Context) {
			BreadcrumbList(ctx, func(child *ui.Context) {
				BreadcrumbLink(child, "Docs", "https://example.com/docs")
				BreadcrumbSeparator(child, func(separator *ui.Context) {
					if show {
						ui.Text(separator, "/")
					}
				})
				BreadcrumbPage(child, "Current")
				BreadcrumbEllipsis(child)
			})
		})
		ui.TextInput(c, &draft).Label("Following custom trail")
	})
	routeClick(t, r, "Following custom trail")
	show = true
	r.Frame()
	if !r.Focused("Following custom trail") {
		t.Fatal("empty compound scope leaked identity")
	}
	routeClick(t, r, "Docs")
	if len(r.OpenedURLs()) != 1 || r.OpenedURLs()[0] != "https://example.com/docs" {
		t.Fatal("compound link did not use native URL action")
	}
	current := routeAccess(t, r, ui.RoleLink, "Current")
	if current.Current != ui.CurrentPage || !current.Disabled {
		t.Fatal("compound page metadata missing")
	}
	for _, node := range r.AccessAll() {
		if node.Name == "/" || node.Name == "…" {
			t.Fatal("decorative compound helper leaked AX")
		}
	}
}

func TestPaginationHugePageCountRendersBoundedWindow(t *testing.T) {
	page := math.MaxInt
	r := routeTester(func(c *ui.Context) { Pagination(c, &page, PaginationProps{Pages: math.MaxInt}) })
	current := routeAccess(t, r, ui.RoleButton, "Page "+strconv.Itoa(math.MaxInt))
	if current.Current != ui.CurrentPage || !routeAccess(t, r, ui.RoleButton, "Next ›").Disabled {
		t.Fatal("max int page lost current/boundary semantics")
	}
	count := 0
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleButton {
			count++
		}
	}
	if count > 7 {
		t.Fatalf("huge count built %d controls", count)
	}
	routeClick(t, r, "‹ Previous")
	if page != math.MaxInt-1 {
		t.Fatalf("previous overflowed: %d", page)
	}
}

func TestPaginationExplicitZeroCountsAndOversizedWindow(t *testing.T) {
	zero := 0
	page := 5
	invalid := false
	p := PaginationProps{Pages: 10, SiblingCount: &zero, BoundaryCount: &zero, MaxVisible: 1}
	r := routeTester(func(c *ui.Context) { PaginationWith(c, &page, p) })
	if _, ok := r.AccessByRole(ui.RoleButton, "Page 5"); !ok {
		t.Fatal("explicit zero windows hid current")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Page 1"); ok {
		t.Fatal("zero boundaries still rendered first page")
	}
	tooMany := math.MaxInt
	p.SiblingCount = &tooMany
	r = routeTester(func(c *ui.Context) { PaginationWith(c, &page, p) })
	invalid = routeAccess(t, r, ui.RoleNavigation, "Pagination").Invalid
	if !invalid || page != 5 {
		t.Fatal("oversized requested window was silently truncated or changed value")
	}
}

func TestPaginationWindowArithmeticAndExplicitBoundaries(t *testing.T) {
	seed := uint64(9347)
	for i := 0; i < 1200; i++ {
		seed = seed*6364136223846793005 + 1
		pages := int(seed%10000) + 1
		if i%7 == 0 {
			pages = math.MaxInt - int(seed%1000)
		}
		page := int(seed%uint64(pages)) + 1
		siblings, boundaries := int(seed%7), int((seed>>4)%5)
		numbers, ok := paginationNumbers(pages, page, siblings, boundaries, 100)
		if !ok || !slices.Contains(numbers, page) {
			t.Fatalf("lost current: pages=%d page=%d nums=%v", pages, page, numbers)
		}
		for j, n := range numbers {
			if n < 1 || n > pages || j > 0 && n <= numbers[j-1] {
				t.Fatalf("invalid bounded ascending window: %v", numbers)
			}
		}
		if boundaries > 0 && (!slices.Contains(numbers, 1) || !slices.Contains(numbers, pages)) {
			t.Fatal("boundary disappeared")
		}
	}
}

func TestPaginationNativeURLRouterAndSameEventCallback(t *testing.T) {
	page, calls, changes := 3, 0, 0
	router := ui.NewRouter("/notes?page=3")
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		router.View(c, func(route *ui.Route) {
			ui.Column(c).Fill().Padding(20).Children(func() {
				if PaginationWith(c, &page, PaginationProps{Pages: 8, URL: func(n int) string { return fmt.Sprintf("?page=%d", n) }, OnPageChange: func(int) { calls++ }}).Changed() {
					changes++
				}
			})
		})
	}, 900, 250)
	routeClick(t, r, "Page 4")
	if page != 4 || router.Query("page") != "4" || calls != 1 || changes != 1 || len(r.OpenedURLs()) != 0 {
		t.Fatalf("native router/event mismatch page=%d route=%s callbacks=%d changed=%d urls=%v", page, router.Location(), calls, changes, r.OpenedURLs())
	}
	r.Frame()
	if calls != 1 || changes != 1 {
		t.Fatal("page event repeated on an idle frame")
	}
}

func TestPaginationLinksUseEnterOnlyAndDisabledAncestry(t *testing.T) {
	page, calls := 1, 0
	disabled := false
	r := routeTester(func(c *ui.Context) {
		ui.Column(c).Disabled(disabled).Children(func() {
			PaginationWith(c, &page, PaginationProps{Pages: 3, URL: func(n int) string { return fmt.Sprintf("https://example.com/?page=%d", n) }, PageContent: func(ctx *ui.Context, n int, _ bool) { ui.Text(ctx, strconv.Itoa(n)) }, OnPageChange: func(int) { calls++ }})
		})
	})
	link := routeAccess(t, r, ui.RoleLink, "Page 2")
	r.ClickAt(link.Bounds.X+link.Bounds.W/2, link.Bounds.Y+link.Bounds.H/2)
	if page != 2 || calls != 1 || len(r.OpenedURLs()) != 1 {
		t.Fatal("actual native link did not activate")
	}
	r.Key(0, ui.KeySpace)
	if len(r.OpenedURLs()) != 1 {
		t.Fatal("space activated native link")
	}
	r.Key(0, ui.KeyEnter)
	if len(r.OpenedURLs()) != 2 || calls != 1 {
		t.Fatal("Enter did not navigate without duplicate page change")
	}
	disabled = true
	r.Frame()
	link = routeAccess(t, r, ui.RoleLink, "Page 3")
	r.ClickAt(link.Bounds.X+link.Bounds.W/2, link.Bounds.Y+link.Bounds.H/2)
	if page != 2 || len(r.OpenedURLs()) != 2 {
		t.Fatal("disabled ancestor did not suppress URL action")
	}
}

func TestPaginationCompactRTLAndCustomEmptyContentIdentity(t *testing.T) {
	page := 2
	show := false
	text := "Draft"
	r := routeTester(func(c *ui.Context) {
		PaginationWith(c, &page, PaginationProps{Pages: 8, Compact: true, PageContent: func(ctx *ui.Context, n int, _ bool) {
			if show {
				ui.Text(ctx, fmt.Sprint(n))
			}
		}})
		ui.TextInput(c, &text).Label("After pages")
	})
	routeClick(t, r, "After pages")
	show = true
	r.Frame()
	if !r.Focused("After pages") {
		t.Fatal("empty page callback leaked identity to following editor")
	}
	previous := routeAccess(t, r, ui.RoleButton, "Go to previous page")
	if previous.Bounds.W != 36 || r.HasText("Previous") {
		t.Fatal("compact previous did not use icon-only width")
	}
	r.SetDirection(ui.RightToLeft)
	previous = routeAccess(t, r, ui.RoleButton, "Go to previous page")
	next := routeAccess(t, r, ui.RoleButton, "Go to next page")
	if previous.Bounds.X <= next.Bounds.X {
		t.Fatal("RTL pagination order was not mirrored")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("pagination did not settle")
	}
}

func TestPaginationPreviousNextKeepNativeFocusAcrossTargets(t *testing.T) {
	for _, linked := range []bool{false, true} {
		t.Run(fmt.Sprint(linked), func(t *testing.T) {
			page := 2
			p := PaginationProps{Pages: 8}
			if linked {
				p.URL = func(n int) string { return fmt.Sprintf("https://example.com/?page=%d", n) }
			}
			r := routeTester(func(c *ui.Context) { PaginationWith(c, &page, p) })
			routeClick(t, r, "Go to next page")
			if page != 3 || !r.Focused("Go to next page") {
				t.Fatalf("next activation replaced focused target: page=%d", page)
			}
			r.Key(0, ui.KeyEnter)
			if page != 4 || !r.Focused("Go to next page") {
				t.Fatal("Enter could not repeat next navigation")
			}
			routeClick(t, r, "Go to previous page")
			r.Key(0, ui.KeyEnter)
			if page != 2 || !r.Focused("Go to previous page") {
				t.Fatal("previous target did not preserve keyboard activation")
			}
		})
	}
}

func TestPaginationCurrentVegaBorderAndRTLArrowPixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			page := 2
			var reference ui.Rect
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Scope(c, theme.Preset("neutral", dark)).Children(func() {
					tokens := theme.TokensOf(c)
					ui.Column(c).Fill().Padding(20).Gap(12).Background(tokens.Background).Children(func() {
						PaginationWith(c, &page, PaginationProps{Pages: 8, Compact: true})
						ui.Box(c).Size(36, 36).Border(1, tokens.Border).Background(tokens.Background).Draw(func(_ *ui.Painter, b ui.Rect) { reference = b })
					})
				})
			}, 700, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			current := routeAccess(t, r, ui.RoleButton, "Page 2").Bounds
			if current.W != 36 || current.H != 36 {
				t.Fatalf("Vega icon metrics=%+v", current)
			}
			pixel := func(rect ui.Rect) color.RGBA { return r.Image().RGBAAt(int(rect.X+rect.W/2), int(rect.Y)) }
			if pixel(current) != pixel(reference) {
				t.Fatalf("current border does not actually paint themed outline: %v vs %v", pixel(current), pixel(reference))
			}
			r.SetDirection(ui.RightToLeft)
			left, _ := r.Find("‹")
			right, _ := r.Find("›")
			if right.X <= left.X {
				t.Fatal("RTL previous/next arrow glyphs did not flip")
			}
		})
	}
}

func TestPaginationLongPageTextDoesNotPaintOutsideNumericControl(t *testing.T) {
	page, zero := math.MaxInt, 0
	var background ui.Color
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		background = theme.TokensOf(c).Background
		ui.Column(c).Fill().Padding(20).Background(background).Children(func() {
			PaginationWith(c, &page, PaginationProps{Pages: math.MaxInt, SiblingCount: &zero, BoundaryCount: &zero, HidePreviousNext: true})
		})
	}, 500, 160)
	box := routeAccess(t, r, ui.RoleButton, "Page "+strconv.Itoa(math.MaxInt)).Bounds
	im := r.Image()
	want := color.RGBA{R: background.R, G: background.G, B: background.B, A: background.A}
	for y := int(box.Y); y < int(box.Y+box.H); y++ {
		for x := int(box.X) - 50; x < int(box.X+box.W)+50; x++ {
			if x >= int(box.X)-2 && x <= int(box.X+box.W)+2 {
				continue
			}
			if got := im.RGBAAt(x, y); got != want {
				t.Fatalf("large page text painted outside control at%d,%d: %v", x, y, got)
			}
		}
	}
}

func routeCrumbs() []BreadcrumbItem {
	return []BreadcrumbItem{{Key: "home", Label: "Home"}, {Key: "docs", Label: "Documentation"}, {Key: "components", Label: "Components"}, {Key: "data", Label: "Data"}, {Key: "current", Label: "Current"}}
}

func TestBreadcrumbNativeLinksRouterAndCustomSeparator(t *testing.T) {
	router := ui.NewRouter("/notes/42")
	chosen := ""
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		router.View(c, func(route *ui.Route) {
			ui.Column(c).Fill().Padding(20).Children(func() {
				value, err := BreadcrumbWith(c, nil, "route", []BreadcrumbItem{{Key: "edit", Label: "Edit note", URL: "edit"}, {Key: "current", Label: "Current"}}, BreadcrumbProps{Separator: func(child *ui.Context) { ui.Text(child, "/") }})
				if err != nil {
					t.Fatal(err)
				}
				if value != "" {
					chosen = value
				}
			})
		})
	}, 600, 200)
	routeClick(t, r, "Edit note")
	if router.Path() != "/notes/edit" || chosen != "edit" || !r.HasText("/") || len(r.OpenedURLs()) != 0 {
		t.Fatal("breadcrumb native link bypassed router or custom separator")
	}
}

func TestBreadcrumbThemedOverflowIsIndependentAndChoosesOnce(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	items := routeCrumbs()
	chosen := ""
	calls := 0
	r := routeTester(func(c *ui.Context) {
		h.BeginFrame()
		defer h.EndFrame()
		value, err := BreadcrumbWith(c, h, "trail", items, BreadcrumbProps{MaxItems: 3, OnSelect: func(BreadcrumbItem) { calls++ }})
		if err != nil {
			t.Fatal(err)
		}
		if value != "" {
			chosen = value
		}
	})
	if r.HasText("Documentation") || r.HasText("Components") {
		t.Fatal("overflow ancestors painted into parent")
	}
	routeClick(t, r, "Show breadcrumb ancestors")
	if len(b.surfaces) != 1 || b.surfaces[0].spec.SystemPopover {
		t.Fatal("themed overflow did not use independent plain Flyout")
	}
	child := b.surfaces[0].child()
	if !child.HasText("Components") {
		t.Fatal("overflow content did not build in child context")
	}
	routeClick(t, child, "Components")
	r.Frame()
	if chosen != "components" || calls != 1 {
		t.Fatalf("overflow result=%q calls=%d", chosen, calls)
	}
	r.Frame()
	if calls != 1 {
		t.Fatal("overflow choice repeated")
	}
}

func TestBreadcrumbNativeOverflowAlwaysUsesOSMenus(t *testing.T) {
	chosen := ""
	calls := 0
	r := routeTester(func(c *ui.Context) {
		c.SetMenuPresentation(ui.PopoverMenus)
		value, err := NativeBreadcrumbWith(c, "native", routeCrumbs(), BreadcrumbProps{MaxItems: 3, OnSelect: func(BreadcrumbItem) { calls++ }})
		if err != nil {
			t.Fatal(err)
		}
		if c.MenuPresentation() != ui.PopoverMenus {
			t.Fatal("OS menu policy leaked")
		}
		if value != "" {
			chosen = value
		}
	})
	routeClick(t, r, "Show breadcrumb ancestors")
	if strings.Join(r.Menu(), ",") != "Documentation,Components,Data" || r.HasText("Documentation") {
		t.Fatalf("not OS menu: %v", r.Menu())
	}
	if err := r.ChooseMenuItem("Data"); err != nil {
		t.Fatal(err)
	}
	if chosen != "data" || calls != 1 {
		t.Fatal("native menu did not choose once")
	}
}

func TestBreadcrumbPendingOverflowRechecksCurrentRoute(t *testing.T) {
	for _, change := range []string{"url", "disabled", "removed", "current"} {
		t.Run(change, func(t *testing.T) {
			b := &panelBackend{}
			h := desktop.NewHostWithBackend(b)
			items := routeCrumbs()
			items[2].URL = "https://old.example.com/"
			calls := 0
			chosen := ""
			r := routeTester(func(c *ui.Context) {
				h.BeginFrame()
				defer h.EndFrame()
				v, err := BreadcrumbWith(c, h, "pending", items, BreadcrumbProps{MaxItems: 3, OnSelect: func(BreadcrumbItem) { calls++ }})
				if err != nil {
					t.Fatal(err)
				}
				if v != "" {
					chosen = v
				}
			})
			routeClick(t, r, "Show breadcrumb ancestors")
			child := b.surfaces[0].child()
			routeClick(t, child, "Components")
			switch change {
			case "url":
				items[2].URL = "https://new.example.com/"
			case "disabled":
				items[2].Disabled = true
			case "removed":
				items = append(items[:2], items[3:]...)
			case "current":
				items = items[:3]
			}
			r.Frame()
			if change == "url" {
				if calls != 1 || chosen != "components" || len(r.OpenedURLs()) != 1 || r.OpenedURLs()[0] != "https://new.example.com/" {
					t.Fatalf("stale URL used: chosen=%s calls=%d urls=%v", chosen, calls, r.OpenedURLs())
				}
			} else if calls != 0 || chosen != "" || len(r.OpenedURLs()) != 0 {
				t.Fatalf("stale route dispatched: chosen=%s calls=%d urls=%v", chosen, calls, r.OpenedURLs())
			}
		})
	}
}

func TestBreadcrumbStructureUpdatesOpenChildAndOwnerRemovalDropsLateResult(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	items := routeCrumbs()
	visible := true
	maxItems := 3
	calls := 0
	r := routeTester(func(c *ui.Context) {
		h.BeginFrame()
		defer h.EndFrame()
		if visible {
			_, err := BreadcrumbWith(c, h, "owner", items, BreadcrumbProps{MaxItems: maxItems, OnSelect: func(BreadcrumbItem) { calls++ }})
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	routeClick(t, r, "Show breadcrumb ancestors")
	surface := b.surfaces[0]
	child := surface.child()
	items[2].Disabled = true
	items[1].Label = "Updated docs"
	r.Frame()
	if !child.HasText("Updated docs") || !routeAccess(t, child, ui.RoleMenuItem, "Components").Disabled {
		t.Fatal("changed structure did not refresh child")
	}
	maxItems = 4
	r.Frame()
	if child.HasText("Data") {
		t.Fatal("changed overflow window did not refresh child")
	}
	lateClosed := surface.closed
	visible = false
	r.Frame()
	child.Key(0, ui.KeyEnter)
	visible = true
	r.Frame()
	routeClick(t, r, "Show breadcrumb ancestors")
	if len(b.surfaces) != 2 {
		t.Fatalf("owner removal did not release first surface: %d", len(b.surfaces))
	}
	lateClosed()
	r.Frame()
	if b.surfaces[1].closes != 0 || !routeAccess(t, r, ui.RoleButton, "Show breadcrumb ancestors").Expanded {
		t.Fatal("late old close affected remounted owner")
	}
	if calls != 0 {
		t.Fatal("unmounted owner received old result")
	}
}

func TestBreadcrumbStableKeyReorderRetainsNativeFocus(t *testing.T) {
	items := []BreadcrumbItem{{Key: "alpha", Label: "Alpha"}, {Key: "beta", Label: "Beta"}, {Key: "current", Label: "Current"}}
	r := routeTester(func(c *ui.Context) { Breadcrumb(c, items) })
	routeClick(t, r, "Beta")
	if !r.Focused("Beta") {
		t.Fatal("native crumb did not receive focus")
	}
	items[0], items[1] = items[1], items[0]
	r.Frame()
	if !r.Focused("Beta") {
		t.Fatal("reordered crumb lost stable keyed focus")
	}
	r.SetDirection(ui.RightToLeft)
	alpha := routeAccess(t, r, ui.RoleButton, "Alpha")
	beta := routeAccess(t, r, ui.RoleButton, "Beta")
	if beta.Bounds.X <= alpha.Bounds.X {
		t.Fatal("RTL breadcrumb order was not mirrored")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("breadcrumb kept idle frames")
	}
}

func TestBreadcrumbMissingHostAndEmptyCallbacksDoNotFallbackOrLeak(t *testing.T) {
	show := false
	value := "draft"
	errors := 0
	var lastErr error
	r := routeTester(func(c *ui.Context) {
		items := routeCrumbs()
		items[0].Content = func(child *ui.Context) {
			if show {
				ui.Text(child, "Visual home")
			}
		}
		_, err := BreadcrumbWith(c, nil, "no-host", items, BreadcrumbProps{MaxItems: 3, Separator: func(child *ui.Context) {
			if show {
				ui.Text(child, "/")
			}
		}})
		if err != nil {
			lastErr = err
			errors++
		}
		ui.TextInput(c, &value).Label("After trail")
	})
	routeClick(t, r, "After trail")
	show = true
	r.Frame()
	if !r.Focused("After trail") {
		t.Fatal("empty crumb/separator callback leaked a key")
	}
	routeClick(t, r, "Show breadcrumb ancestors")
	if lastErr == nil || errors != 1 || r.HasText("Documentation") {
		t.Fatal("opening without Host did not report explicit error")
	}
	r.Frame()
	if errors != 1 {
		t.Fatal("missing Host entered an error/redraw loop")
	}
}
