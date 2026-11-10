package gallery

import (
	"bytes"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func switchCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_SWITCH_COMPOSITION_SNAPSHOT_DIR")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, r.Image())
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("switch image %v/%v", err, closeErr)
	}
}

// One public native-input workflow follows the fixed source compositions. The
// injected selector backend and software images are not real OS/AT acceptance.
func TestSwitchCompositionsSourceWorkflow(t *testing.T) {
	p := &switchCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 800, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	choose := func(selector, label string) {
		t.Helper()
		if err := r.Click(selector); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(label); err != nil {
				t.Fatal(err)
			}
		} else {
			surface := b.surfaces[len(b.surfaces)-1]
			surface.tester.Key(0, ui.KeyHome)
			for _, option := range switchCompositionRecipes {
				if option.Label == label {
					break
				}
				surface.tester.Key(0, ui.KeyDown)
			}
			surface.tester.Key(0, ui.KeyEnter)
			if !surface.didClose {
				t.Fatal("the source styled selector did not close after committing")
			}
		}
		r.Frame()
		if p.popupError != "" {
			t.Fatal(p.popupError)
		}
	}
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, name)
		if !ok {
			t.Fatalf("missing source target %v/%s", role, name)
		}
		return a
	}
	clickSwitch := func(label string) {
		t.Helper()
		a := access(ui.RoleSwitch, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	if p.recipe != "demo" || len(switchCompositionRecipes) != 12 || access(ui.RoleSwitch, "Airplane Mode").Checked {
		t.Fatal("the source default or complete collection changed")
	}
	clickSwitch("Airplane Mode")
	if !*p.values["demo:airplane"] || p.changes != 1 || p.action != "Local switch action: Airplane Mode" {
		t.Fatal("the real source switch did not activate once")
	}
	r.Key(0, ui.KeySpace)
	if *p.values["demo:airplane"] || p.changes != 2 {
		t.Fatal("native Space did not toggle the source model once")
	}
	r.SetPreferences(ui.Preferences{})
	r.Key(0, ui.KeySpace)
	r.Advance(20 * time.Millisecond)
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(250 * time.Millisecond)
	if bytes.Equal(before, r.Image().Pix) || p.changes != 3 || !*p.values["demo:airplane"] {
		t.Fatal("the actual source switch was disconnected from native thumb motion")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	for _, recipe := range switchCompositionRecipes {
		choose("Switch configuration", recipe.Label)
		if p.recipe != recipe.Value {
			t.Fatalf("source configuration %s built %s", recipe.Value, p.recipe)
		}
	}
	choose("Switch configuration", "Description (docs)")
	a := access(ui.RoleSwitch, "Share across devices")
	if a.Checked || !slices.Contains(a.DescribedBy, switchSourceDescription) {
		t.Fatal("the preceding source label/description lost its native target association")
	}
	if err := r.Click("Share across devices"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !*p.values["description:share"] || p.changes != 4 {
		t.Fatal("the source Field label did not toggle its actual native switch once")
	}
	description, ok := r.Find(switchSourceDescription)
	if !ok {
		t.Fatal("the full description was not rendered")
	}
	r.ClickAt(description.X+description.W/2, description.Y+description.H/2)
	r.Frame()
	if p.changes != 4 {
		t.Fatal("a plain description incorrectly became a choice-card label")
	}
	switchCompositionSnapshot(t, r, "switch-description-light")
	choose("Switch configuration", "Choice Cards")
	if access(ui.RoleSwitch, "Share across devices").Checked || !access(ui.RoleSwitch, "Enable notifications").Checked {
		t.Fatal("the independent source card defaults changed")
	}
	description, _ = r.Find(switchSourceDescription)
	r.ClickAt(description.X+4, description.Y+description.H/2)
	r.Frame()
	if !*p.values["choice-card:share"] || p.changes != 5 || p.lastKey != "choice-card:share" {
		t.Fatal("clicking the whole card's description did not toggle the actual switch once")
	}
	clickSwitch("Enable notifications")
	if *p.values["choice-card:notifications"] || p.changes != 6 {
		t.Fatal("the card wrapper double-activated a direct switch click")
	}
	dark = true
	r.Frame()
	switchCompositionSnapshot(t, r, "switch-choice-cards-dark")
	choose("Switch configuration", "Create: Description Choice Card")
	if access(ui.RoleSwitch, "Share across devices").Checked {
		t.Fatal("Create's choice-card state shared the distinct docs draft")
	}
	description, _ = r.Find(switchSourceDescription)
	r.ClickAt(description.X+4, description.Y+description.H/2)
	r.Frame()
	if !*p.values["create-description:share"] || p.changes != 7 {
		t.Fatal("Create's description did not keep its whole-card activation")
	}
	choose("Switch configuration", "Create: Bluetooth Label")
	if !access(ui.RoleSwitch, "Bluetooth").Checked {
		t.Fatal("source Bluetooth defaultChecked was lost")
	}
	choose("Switch configuration", "Disabled (docs)")
	disabled := access(ui.RoleSwitch, "Disabled")
	if !disabled.Disabled || disabled.Checked || disabled.Focusable {
		t.Fatal("the source disabled state was not native")
	}
	r.ClickAt(disabled.Bounds.X+2, disabled.Bounds.Y+disabled.Bounds.H/2)
	r.Key(0, ui.KeySpace)
	if *p.values["disabled:disabled"] || p.changes != 7 {
		t.Fatal("a disabled source switch changed")
	}
	choose("Switch configuration", "Create: Disabled States")
	for _, expected := range []struct {
		label   string
		checked bool
	}{{"Disabled (Unchecked)", false}, {"Disabled (Checked)", true}} {
		a = access(ui.RoleSwitch, expected.label)
		if !a.Disabled || a.Checked != expected.checked || a.Focusable {
			t.Fatal("Create's two distinct disabled defaults were lost")
		}
	}
	choose("Switch configuration", "Invalid Terms")
	a = access(ui.RoleSwitch, "Accept terms and conditions")
	if !a.Invalid || a.Disabled || !slices.Contains(a.DescribedBy, "You must accept the terms and conditions to continue.") {
		t.Fatal("source invalid styling disabled a valid switch or lost its description")
	}
	clickSwitch("Accept terms and conditions")
	if !*p.values["invalid:terms"] || !access(ui.RoleSwitch, "Accept terms and conditions").Invalid {
		t.Fatal("invalid remained uneditable or lost invalid state after activation")
	}
	choose("Switch configuration", "Sizes (docs)")
	small, normal := access(ui.RoleSwitch, "Small"), access(ui.RoleSwitch, "Default")
	if math.Abs(float64(small.Bounds.W-24)) > .1 || math.Abs(float64(small.Bounds.H-14)) > .1 || math.Abs(float64(normal.Bounds.W-32)) > .1 || math.Abs(float64(normal.Bounds.H-18.4)) > .1 {
		t.Fatalf("source track sizes differ small=%v default=%v", small.Bounds, normal.Bounds)
	}
	// Field rows and native caption rows have different intrinsic font line
	// boxes. Compare their actual surrounding row edges, not track centers.
	rowBounds := func(target ui.AccessInfo) ui.Rect {
		bounds := target.Bounds
		for _, candidate := range r.AccessAll() {
			b := candidate.Bounds
			if candidate.Role == ui.RoleGroup && candidate.Name != "Switch compositions" && b.X <= target.Bounds.X && b.X+b.W >= target.Bounds.X+target.Bounds.W && b.Y <= target.Bounds.Y && b.Y+b.H >= target.Bounds.Y+target.Bounds.H && (bounds == target.Bounds || b.H < bounds.H) {
				bounds = b
			}
		}
		return bounds
	}
	smallRow, normalRow := rowBounds(small), rowBounds(normal)
	docsGap := normalRow.Y - smallRow.Y - smallRow.H
	choose("Switch configuration", "Create: Sizes")
	small, normal = access(ui.RoleSwitch, "Small"), access(ui.RoleSwitch, "Default")
	createGap := normal.Bounds.Y - small.Bounds.Y - small.Bounds.H
	if math.Abs(float64(docsGap-28)) > .1 || math.Abs(float64(createGap-48)) > .1 {
		t.Fatalf("source group gap28 vs gap48 collapsed: docs=%g Create=%g", docsGap, createGap)
	}
	dark = false
	r.Frame()
	switchCompositionSnapshot(t, r, "switch-create-sizes-light")
	choose("Switch configuration", "Translated Switch")
	choose("Switch language", "עברית")
	label := "שיתוף בין מכשירים"
	a = access(ui.RoleSwitch, label)
	caption, ok := r.AccessByRole(ui.RoleText, label)
	if !ok || a.Bounds.X >= caption.Bounds.X || a.Checked {
		t.Fatal("the source RTL content/control ordering or default was lost")
	}
	clickSwitch(label)
	if !*p.values["rtl:share"] || p.action != "Local switch action: "+label {
		t.Fatal("translated source action did not update its independent model")
	}
	dark = true
	r.Frame()
	switchCompositionSnapshot(t, r, "switch-hebrew-dark")
	r.SetSize(280, 800)
	card := access(ui.RoleGroup, "Switch compositions").Bounds
	a = access(ui.RoleSwitch, label)
	if a.Bounds.X < card.X || a.Bounds.X+a.Bounds.W > card.X+card.W || a.Bounds.Y+a.Bounds.H > card.Y+card.H {
		t.Fatal("the real source switch escaped the narrow card")
	}
	r.SetSize(620, 800)
	native = false
	r.Frame()
	choose("Switch configuration", "Airplane Mode (docs)")
	if !access(ui.RoleSwitch, "Airplane Mode").Checked || p.changes != 9 {
		t.Fatal("changing the selector route lost the source switch draft or replayed a callback")
	}
	if err := r.Click("Switch configuration"); err != nil {
		t.Fatal(err)
	}
	surface := b.surfaces[len(b.surfaces)-1]
	visible = false
	r.Frame()
	if !surface.didClose {
		t.Fatal("leaving source content retained its selector window")
	}
	visible = true
	r.Frame()
	if p.recipe != "demo" || !access(ui.RoleSwitch, "Airplane Mode").Checked || p.changes != 9 {
		t.Fatal("page return lost application state or replayed old input")
	}
}
