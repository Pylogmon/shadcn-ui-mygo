package display

import (
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func disclosureAccess(t *testing.T, r *ui.Tester, label string) ui.AccessInfo {
	t.Helper()
	for _, info := range r.AccessAll() {
		if info.Role == ui.RoleDisclosure && strings.HasPrefix(info.Name, label) {
			return info
		}
	}
	t.Fatalf("missing disclosure %q: %+v", label, r.AccessAll())
	return ui.AccessInfo{}
}

func TestAccordionSkipsDisabledHeaderAndReportsChanges(t *testing.T) {
	first, unavailable, last := false, false, false
	changes := 0
	r := displayTester(func(c *ui.Context) {
		if Accordion(c, AccordionProps{Items: []AccordionItem{
			{Key: "first", Title: "First", Open: &first},
			{Key: "off", Title: "Unavailable", Open: &unavailable, Disabled: true},
			{Key: "last", Title: "Last", Open: &last},
		}}).Changed() {
			changes++
		}
	})
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyDown)
	if !disclosureAccess(t, r, "Last").Focused {
		t.Fatal("Down did not skip the disabled header")
	}
	r.Key(0, ui.KeySpace)
	if !last || changes != 1 {
		t.Fatalf("activation last=%v changes=%d", last, changes)
	}
	r.Frame()
	if changes != 1 {
		t.Fatal("activation reported more than once")
	}
}

func TestDisclosurePanelAssociationAndClosedContentIsInert(t *testing.T) {
	open := true
	value := ""
	r := displayTester(func(c *ui.Context) {
		Collapsible(c, "Details", &open, func() { ui.TextInput(c, &value).Label("Secret").AutoFocus() })
	})
	trigger := disclosureAccess(t, r, "Details")
	if len(trigger.Controls) == 0 {
		t.Fatal("disclosure has no controlled panel association")
	}
	open = false
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleTextField, "Secret"); ok {
		t.Fatal("closing panel still exposes an interactive editor")
	}
	if !disclosureAccess(t, r, "Details").Focused {
		t.Fatal("closing focused content did not return focus to trigger")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("closed disclosure has pending idle frames")
	}
}

func TestAccordionSingleKeyIdentityAndHeadingVariants(t *testing.T) {
	first, second, disabled := true, false, false
	items := []AccordionItem{
		{Key: "first", Title: "First", Open: &first, Content: func() {}},
		{Key: "second", Title: "Second", Open: &second},
		{Key: "off", Title: "Disabled", Open: &disabled, Disabled: true},
	}
	var events [][]string
	r := displayTester(func(c *ui.Context) {
		Accordion(c, AccordionProps{Items: items, Single: true, Bordered: true, HeadingLevel: 2, Loop: true,
			OnChange: func(keys []string) { events = append(events, slices.Clone(keys)) }})
	})
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeySpace)
	if first || !second || len(events) != 1 || !slices.Equal(events[0], []string{"second"}) {
		t.Fatalf("single commit first=%v second=%v events=%v", first, second, events)
	}
	// Reordering with the same stable keys must retain the focused section.
	items[0], items[1] = items[1], items[0]
	r.Frame()
	if !disclosureAccess(t, r, "Second").Focused {
		t.Fatal("reordering lost stable header focus")
	}
	r.Key(0, ui.KeyUp)
	if !disclosureAccess(t, r, "First").Focused {
		t.Fatal("loop did not skip the disabled last header")
	}
	headings := 0
	for _, info := range r.AccessAll() {
		if info.Role == ui.RoleHeading {
			headings++
			if info.Level != 2 {
				t.Fatalf("heading level=%d", info.Level)
			}
		}
	}
	if headings != 3 {
		t.Fatalf("heading count=%d", headings)
	}
	first, second = true, true
	r.Frame()
	if !first || second || len(events) != 1 {
		t.Fatal("external controlled normalization did not prefer the newly opened section or reported a user event")
	}
}

func TestCollapsibleRetainedContentCustomTriggerAndDisabled(t *testing.T) {
	open, disabled, created, changes := false, false, 0, 0
	r := displayTester(func(c *ui.Context) {
		if CollapsibleWith(c, &open, CollapsibleProps{Label: "Options", KeepMounted: true, Disabled: disabled,
			Trigger: func(tr ui.Element) { tr.Size(120, 32).Children(func() { ui.Text(c, "Show options") }) },
			Content: func() {
				e := ui.Column(c)
				n := ui.Local(e, "counter", func() int { created++; return 0 })
				e.Children(func() {
					if ui.Button(c, "Increment").Clicked() {
						*n++
					}
					ui.Textf(c, "Count %d", *n)
				})
			}, OnOpenChange: func(bool) { changes++ }}).Changed() && changes == 0 {
			t.Fatal("Changed and OnOpenChange disagreed")
		}
	})
	if _, ok := r.AccessByRole(ui.RoleButton, "Increment"); ok {
		t.Fatal("retained closed content is accessible")
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeySpace)
	r.Advance(300 * time.Millisecond)
	if err := r.Click("Increment"); err != nil {
		t.Fatal(err)
	}
	if !r.HasText("Count 1") {
		t.Fatal("retained content did not receive its own input")
	}
	open = false
	r.Frame()
	r.Advance(300 * time.Millisecond)
	open = true
	r.Frame()
	r.Advance(300 * time.Millisecond)
	if created != 1 || !r.HasText("Count 1") || changes != 1 {
		t.Fatalf("local state remounted or external value reported changes: created=%d changes=%d", created, changes)
	}
	disabled = true
	r.Frame()
	info := disclosureAccess(t, r, "Options")
	r.ClickAt(info.Bounds.X+5, info.Bounds.Y+5)
	if !open || !disclosureAccess(t, r, "Options").Disabled || changes != 1 {
		t.Fatal("disabled custom disclosure still toggled")
	}
}

func TestDisclosureAnimationReversalAndReducedMotion(t *testing.T) {
	for _, reduced := range []bool{false, true} {
		t.Run(map[bool]string{false: "animated", true: "reduced"}[reduced], func(t *testing.T) {
			open := true
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).FillWidth().Padding(16).Children(func() {
					Collapsible(c, "Details", &open, func() { ui.Box(c).Height(120).FillWidth().Background(ui.RGB(230, 0, 0)) })
					ui.Text(c, "Following content")
				})
			}, 360, 320, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
			marker := func() float32 {
				b, ok := r.Find("Following content")
				if !ok {
					t.Fatal("following marker missing")
				}
				return b.Y
			}
			start := marker()
			open = false
			r.Frame()
			if reduced {
				if marker() >= start-100 || r.FramePending() {
					t.Fatal("reduced disclosure did not close immediately and settle")
				}
				return
			}
			r.Advance(80 * time.Millisecond)
			// Sample at the exact current time before retargeting; Advance's
			// last refresh can precede its final timestamp by a fraction frame.
			r.Frame()
			middle := marker()
			if middle >= start || middle <= start-136 {
				t.Fatalf("no actual closing layout: start=%v middle=%v", start, middle)
			}
			open = true
			r.Frame()
			if math.Abs(float64(marker()-middle)) > 1 {
				t.Fatalf("reversal jumped: before=%v after=%v", middle, marker())
			}
			r.Advance(80 * time.Millisecond)
			if marker() <= middle {
				t.Fatal("reversal did not grow back toward the open height")
			}
			r.Advance(500 * time.Millisecond)
			if math.Abs(float64(marker()-start)) > 1 || r.FramePending() {
				t.Fatal("disclosure did not settle on its original layout")
			}
		})
	}
}

func TestReducedDisclosureRestoresOnlyItsOwnContentFocus(t *testing.T) {
	open, value := true, ""
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).FillWidth().Children(func() {
			Collapsible(c, "Details", &open, func() { ui.TextInput(c, &value).Label("Editor").AutoFocus() })
			ui.Button(c, "Other action")
		})
	}, 360, 320, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !r.Focused("Editor") {
		t.Fatal("initial editor is not focused")
	}
	open = false
	r.Frame()
	if !r.Focused("Details") {
		t.Fatal("instant removal stranded content focus")
	}
	open = true
	r.Frame()
	if err := r.Click("Other action"); err != nil {
		t.Fatal(err)
	}
	open = false
	r.Frame()
	if !r.Focused("Other action") {
		t.Fatal("closure stole another control's focus")
	}
}

func TestClosedRetainedDisclosureDoesNotStealInitialFocus(t *testing.T) {
	open, value := false, ""
	r := displayTester(func(c *ui.Context) {
		ui.Button(c, "Outside").AutoFocus()
		CollapsibleWith(c, &open, CollapsibleProps{Label: "Details", KeepMounted: true, Content: func() {
			ui.TextInput(c, &value).Label("Hidden editor").AutoFocus()
		}})
	})
	if !r.Focused("Outside") {
		t.Fatal("retained closed content captured initial focus")
	}
	r.Type("unexpected")
	if value != "" {
		t.Fatal("closed retained editor accepted keyboard input")
	}
	open = true
	r.Frame()
	r.Advance(300 * time.Millisecond)
	if _, ok := r.AccessByRole(ui.RoleTextField, "Hidden editor"); !ok {
		t.Fatal("retained editor did not become accessible on opening")
	}
}
