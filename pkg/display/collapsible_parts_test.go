package display_test

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// One native workflow composes the source Settings layout, a separate Order
// status and Accordion styling. It uses real native editors/disclosures rather
// than a second checkbox or a button simulating a disclosure.
func TestCollapsiblePartsIndependentLayoutAndScopedContent(t *testing.T) {
	open, disabled, reduce := false, false, false
	values := [4]string{"0", "0", "0", "0"}
	scene, changes, frame, builtAt, builds := "settings", 0, 0, -1, 0
	var escaped *display.CollapsibleScope
	var detail ui.Handle
	var selection [2]int
	var declaredTokens, builtTokens theme.Tokens
	var declaredFonts, builtFonts theme.FontRoles
	var declaredMenu, builtMenu theme.MenuSnapshot
	var builtDirection ui.Direction
	local, localDark := theme.Preset("rose", false), theme.Preset("rose", true)
	local.FontSize = 17
	first, second, legacyOpen := true, false, false
	accordionCallbacks := 0
	r := ui.NewTester(func(c *ui.Context) {
		frame++
		if reduce {
			c.DisableAnimations()
		}
		theme.Scope(c, theme.Light()).Fill().Padding(20).Gap(12).Children(func() {
			switch scene {
			case "settings":
				display.CollapsibleRoot(c, &open, display.CollapsibleRootProps{Label: "Radius settings", Disabled: disabled, KeepMounted: true, OnOpenChange: func(bool) { changes++ }}, func(s *display.CollapsibleScope) {
					escaped = s
					ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Start).Children(func() {
						theme.FontScope(c, theme.FontRoles{Body: "serif", Heading: "monospace"}).Grow(1).MinWidth(0).Children(func() {
							theme.MenuScope(c, local, theme.MenuOptions{Color: theme.MenuInverted, Accent: theme.MenuBold, InvertedTokens: localDark}).Children(func() {
								ui.Column(c).FillWidth().MinWidth(0).Gap(8).Font("monospace").FontSize(23).TextColor(ui.RGB(173, 34, 105)).Direction(ui.RightToLeft).Children(func() {
									ui.Text(c, "MWMW 123").Label("Inheritance reference").AlignSelf(ui.Start)
									ui.Grid(c).FillWidth().Columns(2).Gap(8).Children(func() {
										control.Input(c.Key("x"), &values[0], control.InputProps{Label: "Radius X"})
										control.Input(c.Key("y"), &values[1], control.InputProps{Label: "Radius Y"})
										declaredTokens, declaredFonts, declaredMenu = *theme.TokensOf(c), theme.FontsOf(c), theme.MenuOf(c)
										before := builds
										parts := s.Content(c, display.CollapsibleContentProps{Label: "Radius detail panel"}, func(child *ui.Context) {
											if builtAt == frame {
												t.Fatal("Content was built twice in one current Root build")
											}
											builtAt, builds = frame, builds+1
											builtTokens, builtFonts, builtMenu, builtDirection = *theme.TokensOf(child), theme.FontsOf(child), theme.MenuOf(child), child.Direction()
											control.Input(child.Key("detail-x"), &values[2], control.InputProps{Label: "Radius X detail", Handle: &detail})
											control.Input(child.Key("detail-y"), &values[3], control.InputProps{Label: "Radius Y detail"})
											ui.Column(child).ColumnSpan(2).AlignItems(ui.Start).Children(func() {
												ui.Text(child, "MWMW 123").Label("Deferred inheritance")
											})
											if target := child.Resolve(detail); target.Valid() {
												selection[0], selection[1] = target.TextSelection()
											}
										})
										parts.Root.ColumnSpan(2)
										parts.Content.Grid().Columns(2).Gap(8)
										if builds != before {
											t.Fatal("Content executed before the later Trigger")
										}
									})
								})
							})
						})
						s.Trigger(c, display.CollapsibleTriggerProps{Label: "Expand radius settings", Style: func(tr ui.Element) {
							control.StyleButton(c, tr, "", control.ButtonProps{Variant: control.Outline, Size: control.IconSmall})
						}}, func(child *ui.Context) { ui.Text(child, "+").AccessHidden(true) })
					})
				})
			case "order":
				display.CollapsibleRoot(c, &open, display.CollapsibleRootProps{Label: "Order #4189"}, func(s *display.CollapsibleScope) {
					ui.Row(c).FillWidth().Justify(ui.SpaceBetween).Children(func() {
						ui.Text(c, "Order #4189")
						s.Trigger(c, display.CollapsibleTriggerProps{Label: "Toggle order details"}, nil)
					})
					ui.Text(c, "Status: Shipped")
					s.Content(c, display.CollapsibleContentProps{}, func(child *ui.Context) { ui.Text(child, "100 Market St, San Francisco") })
				})
			case "accordion":
				display.Accordion(c, display.AccordionProps{Single: true, Items: []display.AccordionItem{
					{Key: "one", Title: "Styled first", Open: &first, Content: func() { ui.Text(c, "Styled body text") }},
					{Key: "two", Title: "Styled second", Open: &second},
				}, StyleItem: func(child *ui.Context, parts display.AccordionItemParts) {
					accordionCallbacks++
					parts.Trigger.PaddingX(16)
					if parts.Content.Valid() {
						parts.Content.PaddingX(16)
					}
				}}).Border(1, theme.TokensOf(c).Border)
			case "legacy":
				display.Collapsible(c.Key("legacy"), "Legacy details", &legacyOpen, func() {
					ui.Box(c).FillWidth().Height(120)
				})
			}
		})
	}, 560, 600)
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, name)
		if !ok {
			t.Fatalf("missing native %v %q", role, name)
		}
		return a
	}
	click := func(name string) {
		t.Helper()
		if err := r.Click(name); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	if builds != 0 {
		t.Fatal("initially closed content was eagerly mounted")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Radius X detail"); ok {
		t.Fatal("initially closed editor was accessible")
	}
	if access(ui.RoleTextField, "Radius X").Bounds.H < 30 || access(ui.RoleTextField, "Radius Y").Bounds.H < 30 {
		t.Fatal("always-visible editors were compressed into the trigger")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("a retained Scope stayed usable after its build")
			}
		}()
		escaped.Open()
	}()
	click("Radius X")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("12")
	click("Expand radius settings")
	r.Advance(100 * time.Millisecond)
	enteringHeight := access(ui.RoleRegion, "Radius detail panel").Bounds.H
	r.Advance(150 * time.Millisecond)
	if full := access(ui.RoleRegion, "Radius detail panel").Bounds.H; enteringHeight <= 0 || enteringHeight >= full {
		t.Fatalf("native disclosure skipped its measured intermediate height: %.2f -> %.2f", enteringHeight, full)
	}
	if !open || changes != 1 || values[0] != "12" || builds == 0 {
		t.Fatal("independent native trigger did not preserve/edit the two visible inputs")
	}
	if declaredTokens != builtTokens || declaredFonts != builtFonts || declaredMenu != builtMenu || builtDirection != ui.RightToLeft {
		t.Fatal("deferred content lost its declaration's current theme/font/menu/direction")
	}
	ref, deferred := access(ui.RoleText, "Inheritance reference"), access(ui.RoleText, "Deferred inheritance")
	if math.Abs(float64(ref.Bounds.W-deferred.Bounds.W)) > .1 || math.Abs(float64(ref.Bounds.H-deferred.Bounds.H)) > .1 {
		t.Fatalf("theme reentry replaced ordinary fluent text inheritance: %v vs %v", ref.Bounds, deferred.Bounds)
	}
	trigger := access(ui.RoleDisclosure, "Expand radius settings")
	if len(trigger.Controls) != 1 || trigger.Controls[0] != "Radius detail panel" {
		t.Fatalf("native trigger did not control its independent panel: %+v", trigger.Controls)
	}
	click("Radius X detail")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("1234")
	r.Key(ui.Shift, ui.KeyLeft)
	beforeSelection := selection
	click("Expand radius settings")
	if open || changes != 2 {
		t.Fatal("native close was not reported exactly once")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Radius X detail"); ok {
		t.Fatal("closing content was not immediately excluded from AX")
	}
	r.Advance(250 * time.Millisecond)
	if !r.Focused("Expand radius settings") || r.FramePending() {
		t.Fatal("closed retained content lost trigger focus or scheduled idle frames")
	}
	click("Expand radius settings")
	r.Advance(250 * time.Millisecond)
	if values[2] != "1234" || selection != beforeSelection {
		t.Fatalf("reopening discarded the real editor draft/selection: %q %v vs %v", values[2], selection, beforeSelection)
	}
	disabled = true
	r.Frame()
	click("Expand radius settings")
	if !open || changes != 3 {
		t.Fatal("disabled trigger changed the disclosure")
	}
	disabled, reduce, open = false, true, false
	r.Frame()
	if access(ui.RoleTextField, "Radius X").Value != "12" || r.FramePending() {
		t.Fatal("reduced external close changed the always-visible input or stayed animated")
	}
	scene = "order"
	r.Frame()
	if !r.HasText("Status: Shipped") || r.HasText("100 Market St, San Francisco") {
		t.Fatal("order status was hidden inside the closed panel")
	}
	click("Toggle order details")
	if !r.HasText("Status: Shipped") || !r.HasText("100 Market St, San Francisco") {
		t.Fatal("order content did not open independently of its status")
	}
	scene = "accordion"
	r.Frame()
	a, text := access(ui.RoleDisclosure, "Styled first"), access(ui.RoleText, "Styled body text")
	if math.Abs(float64(text.Bounds.X-a.Bounds.X-16)) > .1 || accordionCallbacks < 2 {
		t.Fatalf("current Accordion item style did not reach the actual panel: %v vs %v", text.Bounds, a.Bounds)
	}
	click("Styled first")
	r.Key(0, ui.KeyDown)
	if !r.Focused("Styled second") {
		t.Fatal("styling the items split native Accordion header navigation")
	}
	// The compatibility entry uses the same height measurement boundary. Its
	// first lazy opening must animate too, rather than wait for an unclipped Draw.
	scene, reduce = "legacy", false
	r.Frame()
	click("Legacy details")
	r.Advance(100 * time.Millisecond)
	legacyMiddle := access(ui.RoleRegion, "Legacy details").Bounds.H
	r.Advance(150 * time.Millisecond)
	if full := access(ui.RoleRegion, "Legacy details").Bounds.H; legacyMiddle <= 0 || legacyMiddle >= full {
		t.Fatalf("legacy lazy disclosure skipped intermediate height: %.2f -> %.2f", legacyMiddle, full)
	}
}
