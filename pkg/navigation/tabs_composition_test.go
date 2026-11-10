package navigation_test

import (
	"image/color"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	nav "github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func tabsAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q", role, label)
	}
	return a
}
func TestTabsCompositionHeaderControlsExternalPanelAndNativeAX(t *testing.T) {
	value, query, draft := "home", "", "kept"
	clicks, changes := 0, 0
	var trigger, panel ui.Handle
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() {
			nav.TabsRoot(c, &value, nav.TabsRootProps{Label: "Workspace", OnValueChange: func(string) { changes++ }}, func(c *ui.Context) {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					nav.TabsList(c, nav.TabsListProps{}, func(c *ui.Context) {
						nav.TabsTrigger(c, "home", nav.TabsTriggerProps{Label: "Home"}).Bind(&trigger)
						nav.TabsTrigger(c, "settings", nav.TabsTriggerProps{Label: "Settings"})
					})
					control.Input(c, &query, control.InputProps{Label: "Header search"}).Width(140)
					if control.Button(c, "Header action", control.ButtonProps{Variant: control.Outline}).Clicked() {
						clicks++
					}
				})
				ui.Box(c).FillWidth().Height(100).Padding(12).Border(1, theme.TokensOf(c).Border).Role(ui.RoleGroup).Label("External panel").Children(func() {
					nav.TabsContent(c, "home", nav.TabsContentProps{}, func(c *ui.Context) { control.Input(c, &draft, control.InputProps{Label: "Home draft"}) }).Bind(&panel)
					nav.TabsContent(c, "settings", nav.TabsContentProps{}, func(c *ui.Context) { ui.Text(c, "Settings panel") })
				})
			}).FillWidth()
		})
	}, 700, 280, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	list := tabsAccess(t, r, ui.RoleTabList, "Workspace")
	external := tabsAccess(t, r, ui.RoleGroup, "External panel")
	search := tabsAccess(t, r, ui.RoleTextField, "Header search")
	if external.Bounds.Y < list.Bounds.Y+list.Bounds.H || search.Bounds.Y >= list.Bounds.Y+list.Bounds.H || trigger == (ui.Handle{}) || panel == (ui.Handle{}) {
		t.Fatal("slots did not produce independent real layout/handles")
	}
	tab := tabsAccess(t, r, ui.RoleTab, "Home")
	region := tabsAccess(t, r, ui.RoleRegion, "Home")
	if !slices.Equal(tab.Controls, []string{"Home"}) || !region.Focusable || region.RoleDescription != "tab panel" {
		t.Fatal("native panel relations missing")
	}
	r.Click("Header search")
	r.Type("query")
	r.Key(0, ui.KeyLeft)
	if query != "query" || value != "home" || !r.Focused("Header search") {
		t.Fatal("list captured unrelated input keys")
	}
	r.Click("Header action")
	if clicks != 1 || value != "home" {
		t.Fatal("header action became a tab")
	}
	r.Click("Settings")
	if value != "settings" || changes != 1 || !r.HasText("Settings panel") {
		t.Fatal("independent trigger did not activate panel")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Home draft"); ok {
		t.Fatal("inactive panel controls remained exposed")
	}
	r.Click("Home")
	if tabsAccess(t, r, ui.RoleTextField, "Home draft").Value != "kept" {
		t.Fatal("application draft lost on panel switch")
	}
}
func TestTabsCompositionManualDisabledOrientationAndRTL(t *testing.T) {
	value := "a"
	p := nav.TabsRootProps{Manual: true, NoLoop: true}
	changes := 0
	p.OnValueChange = func(string) { changes++ }
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Children(func() {
			nav.TabsRoot(c, &value, p, func(c *ui.Context) {
				nav.TabsList(c, nav.TabsListProps{Label: "Modes", Variant: nav.TabsLine}, func(c *ui.Context) {
					nav.TabsTrigger(c, "a", nav.TabsTriggerProps{Label: "First"})
					nav.TabsTrigger(c, "disabled", nav.TabsTriggerProps{Label: "Disabled", Disabled: true})
					nav.TabsTrigger(c, "b", nav.TabsTriggerProps{Label: "Last"})
				})
				nav.TabsContent(c, "a", nav.TabsContentProps{}, func(c *ui.Context) { ui.Text(c, "First panel") })
				nav.TabsContent(c, "b", nav.TabsContentProps{}, func(c *ui.Context) { ui.Text(c, "Last panel") })
			})
		})
	}, 500, 250)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyRight)
	if value != "a" || !r.Focused("Last") || changes != 0 {
		t.Fatal("manual navigation selected or failed to skip disabled")
	}
	r.Key(0, ui.KeySpace)
	if value != "b" || changes != 1 {
		t.Fatal("Space did not activate manual tab")
	}
	r.Key(0, ui.KeyRight)
	if !r.Focused("Last") {
		t.Fatal("NoLoop wrapped")
	}
	r.SetDirection(ui.RightToLeft)
	r.Key(0, ui.KeyRight)
	if !r.Focused("First") || value != "b" {
		t.Fatalf("RTL manual direction was not logical: value=%q first=%+v last=%+v", value, tabsAccess(t, r, ui.RoleTab, "First"), tabsAccess(t, r, ui.RoleTab, "Last"))
	}
	p.Vertical = true
	r.Frame()
	r.Key(0, ui.KeyLeft)
	if !r.Focused("First") {
		t.Fatal("vertical list used horizontal key")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	if value != "b" || !tabsAccess(t, r, ui.RoleTabList, "Modes").Vertical {
		t.Fatal("native vertical orientation/activation lost")
	}
	r.Click("Disabled")
	if value != "b" {
		t.Fatal("disabled trigger activated")
	}

	// Compound and legacy tabs apply source half-opacity exactly once, while
	// inherited disabled still blocks real input and uses the native dim.
	glyph := ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path fill="currentColor" d="M0 0h16v16H0z"/></svg>`))
	for _, legacy := range []bool{false, true} {
		explicit, inherited := false, false
		selected, index := "enabled", 0
		pixels := ui.NewTester(func(c *ui.Context) {
			tokens := theme.Preset("blue", false)
			tokens.Background, tokens.Muted = ui.RGB(0, 0, 0), ui.RGB(0, 0, 0)
			tokens.Foreground, tokens.MutedForeground = ui.RGB(255, 255, 255), ui.RGB(255, 255, 255)
			theme.Scope(c, tokens).Fill().Padding(20).Background(tokens.Background).Children(func() {
				ui.Column(c).Disabled(inherited).Children(func() {
					if legacy {
						nav.Tabs(c, &index, nav.TabsProps{Items: []nav.TabItem{{Label: "Enabled", Icon: glyph}, {Label: "Pixel disabled", Icon: glyph, Disabled: explicit}}}, nil)
					} else {
						nav.TabsRoot(c, &selected, nav.TabsRootProps{}, func(c *ui.Context) {
							nav.TabsList(c, nav.TabsListProps{}, func(c *ui.Context) {
								nav.TabsTrigger(c, "enabled", nav.TabsTriggerProps{Label: "Enabled", Icon: glyph})
								nav.TabsTrigger(c, "disabled", nav.TabsTriggerProps{Label: "Pixel disabled", Icon: glyph, Disabled: explicit})
							})
						})
					}
				})
			})
		}, 420, 100)
		check := func(want uint8) {
			t.Helper()
			a := tabsAccess(t, pixels, ui.RoleTab, "Pixel disabled")
			p := color.NRGBAModel.Convert(pixels.Image().At(int(a.Bounds.X+20), int(a.Bounds.Y+a.Bounds.H/2))).(color.NRGBA)
			if p.R != want || p.G != want || p.B != want || p.A != 255 {
				t.Fatalf("legacy=%v explicit=%v inherited=%v pixel=%v want=%d", legacy, explicit, inherited, p, want)
			}
		}
		check(255)
		explicit = true
		pixels.Frame()
		check(128)
		explicit, inherited = false, true
		pixels.Frame()
		check(128)
		a := tabsAccess(t, pixels, ui.RoleTab, "Pixel disabled")
		if !a.Disabled {
			t.Fatal("ancestor disabled failed to reach the actual tab")
		}
		pixels.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		pixels.Key(0, ui.KeyEnter)
		if selected != "enabled" || index != 0 {
			t.Fatal("ancestor disabled tab accepted input")
		}
	}
}
func TestTabsCompositionNestedSiblingScopesAndPanelBeforeList(t *testing.T) {
	outer, inner, sibling := "a", "x", "one"
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Gap(20).Children(func() {
			nav.TabsRoot(c, &outer, nav.TabsRootProps{Key: "outer"}, func(c *ui.Context) {
				nav.TabsContent(c, "a", nav.TabsContentProps{}, func(c *ui.Context) {
					nav.TabsRoot(c, &inner, nav.TabsRootProps{Key: "inner"}, func(c *ui.Context) {
						nav.TabsList(c, nav.TabsListProps{}, func(c *ui.Context) {
							nav.TabsTrigger(c, "x", nav.TabsTriggerProps{Label: "Inner X"})
							nav.TabsTrigger(c, "y", nav.TabsTriggerProps{Label: "Inner Y"})
						})
						nav.TabsContent(c, "x", nav.TabsContentProps{}, func(c *ui.Context) { ui.Text(c, "Inner X panel") })
						nav.TabsContent(c, "y", nav.TabsContentProps{}, func(c *ui.Context) { ui.Text(c, "Inner Y panel") })
					})
				})
				nav.TabsList(c, nav.TabsListProps{}, func(c *ui.Context) {
					nav.TabsTrigger(c, "a", nav.TabsTriggerProps{Label: "Outer A"})
					nav.TabsTrigger(c, "b", nav.TabsTriggerProps{Label: "Outer B"})
				})
				nav.TabsContent(c, "b", nav.TabsContentProps{}, func(c *ui.Context) { ui.Text(c, "Outer B panel") })
			})
			nav.TabsRoot(c, &sibling, nav.TabsRootProps{Key: "sibling"}, func(c *ui.Context) {
				nav.TabsList(c, nav.TabsListProps{}, func(c *ui.Context) {
					nav.TabsTrigger(c, "one", nav.TabsTriggerProps{Label: "Sibling One"})
					nav.TabsTrigger(c, "two", nav.TabsTriggerProps{Label: "Sibling Two"})
				})
				nav.TabsContent(c, "one", nav.TabsContentProps{}, func(c *ui.Context) { ui.Text(c, "Sibling panel") })
			})
		})
	}, 620, 460, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !slices.Equal(tabsAccess(t, r, ui.RoleTab, "Outer A").Controls, []string{"Outer A"}) {
		t.Fatal("panel before list was not linked")
	}
	r.Click("Inner Y")
	if outer != "a" || inner != "y" || sibling != "one" || !r.HasText("Inner Y panel") {
		t.Fatal("nested scope selected another root")
	}
	r.Click("Sibling Two")
	r.Click("Outer B")
	if inner != "y" || sibling != "two" || !r.HasText("Outer B panel") {
		t.Fatal("scope restoration lost sibling or outer state")
	}
}
func TestTabsCompositionIconLabelReorderAndIndicator(t *testing.T) {
	for _, reduced := range []bool{false, true} {
		t.Run(map[bool]string{false: "motion", true: "reduced"}[reduced], func(t *testing.T) {
			value := "home"
			order := []string{"home", "search"}
			labels := map[string]string{"home": "Home icon", "search": "Search icon"}
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Padding(20).Children(func() {
					nav.TabsRoot(c, &value, nav.TabsRootProps{}, func(c *ui.Context) {
						nav.TabsList(c, nav.TabsListProps{Variant: nav.TabsLine}, func(c *ui.Context) {
							for _, v := range order {
								icon := outlined.Home()
								if v == "search" {
									icon = outlined.Search()
								}
								nav.TabsTrigger(c, v, nav.TabsTriggerProps{Label: labels[v], Icon: icon, HideLabel: true})
							}
						})
						nav.TabsContent(c, value, nav.TabsContentProps{}, func(c *ui.Context) { ui.Text(c, "Icon content") })
					})
				})
			}, 450, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
			if tabsAccess(t, r, ui.RoleTab, "Home icon").Bounds.W > 42 || tabsAccess(t, r, ui.RoleTab, "Search icon").Bounds.W > 42 {
				t.Fatalf("HideLabel painted its accessible label: %v", r.Texts())
			}
			r.Click("Search icon")
			if value != "search" || !tabsAccess(t, r, ui.RoleTab, "Search icon").Selected {
				t.Fatal("icon-only trigger had no usable native name")
			}
			if !reduced && !r.FramePending() {
				t.Fatal("composed indicator did not animate")
			}
			r.Advance(400 * time.Millisecond)
			if r.FramePending() {
				t.Fatal("composed indicator did not settle")
			}
			slices.Reverse(order)
			r.Frame()
			if !r.Focused("Search icon") || value != "search" {
				t.Fatal("keyed reorder lost focused semantic trigger")
			}
		})
	}
}
