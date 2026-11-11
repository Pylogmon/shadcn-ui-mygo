package gallery

import (
	"bytes"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func badgeCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_BADGE_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("badge image %v/%v", err, closeErr)
	}
}

// This one public-input workflow selects every source configuration, then uses
// actual Link, loading, palette, scroll and RTL content. Tester is software
// evidence, not a real OS/screen-reader acceptance matrix.
func TestBadgeCompositionsSourceWorkflow(t *testing.T) {
	p := &badgeCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
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
			s := b.surfaces[len(b.surfaces)-1]
			// This injected Backend has no platform FitContent loop. The source
			// selector is bounded, and keyboard selection needs no fake resizing.
			s.tester.Key(0, ui.KeyHome)
			for _, option := range badgeCompositionRecipes {
				if option.Label == label {
					break
				}
				s.tester.Key(0, ui.KeyDown)
			}
			s.tester.Key(0, ui.KeyEnter)
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
			t.Fatalf("missing source role %v/%s", role, name)
		}
		return a
	}
	if p.recipe != "demo" || len(badgeCompositionRecipes) != 14 {
		t.Fatal("source default or full collection missing")
	}
	if access(ui.RoleText, "Badge").Focusable {
		t.Fatal("a source span became an interactive badge")
	}
	checkDanger := func(name string) {
		t.Helper()
		a := access(ui.RoleText, name)
		pixel := r.Image().RGBAAt(int(a.Bounds.X+a.Bounds.W/2), int(a.Bounds.Y+3))
		if pixel.R <= pixel.G || pixel.R <= pixel.B {
			t.Fatalf("source %s lost its danger color: %v", name, pixel)
		}
	}
	checkDanger("Destructive")
	for _, recipe := range badgeCompositionRecipes {
		choose("Badge configuration", recipe.Label)
		if p.recipe != recipe.Value {
			t.Fatalf("source selection %q built %q", recipe.Value, p.recipe)
		}
	}
	choose("Badge configuration", "Create: Six Variants")
	access(ui.RoleText, "Link")
	if _, ok := r.AccessByRole(ui.RoleLink, "Link"); ok {
		t.Fatal("the source's visual Link variant was made navigable without href")
	}
	choose("Badge configuration", "Open Link")
	link := access(ui.RoleLink, "Open Link")
	// Click the visual end icon's area, then Enter on the same native Link.
	r.ClickAt(link.Bounds.X+link.Bounds.W-6, link.Bounds.Y+link.Bounds.H/2)
	r.Frame()
	if p.action != "Local badge action: Open Link" || p.activations != 1 || !r.Focused("Open Link") {
		t.Fatal("decorative source icon intercepted the real native link")
	}
	r.Key(0, ui.KeySpace)
	if p.activations != 1 {
		t.Fatal("source native Link accepted Space as a button")
	}
	r.Key(0, ui.KeyEnter)
	if p.activations != 2 || p.router.Path() != "/badge-examples" || len(r.OpenedURLs()) != 0 {
		t.Fatal("source hash href did not remain a local native router action")
	}
	choose("Badge configuration", "Create: asChild Links")
	for i, variant := range []display.BadgeVariant{display.BadgeDefault, display.BadgeSecondary, display.BadgeSoft, display.BadgeGhost} {
		links := []ui.AccessInfo{}
		for _, a := range r.AccessAll() {
			if a.Role == ui.RoleLink && a.Name == "Link" {
				links = append(links, a)
			}
		}
		if len(links) != 4 {
			t.Fatal("Create asChild did not preserve its four actual links")
		}
		a := links[i]
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
		if p.activations != i+3 || p.lastVariant != variant || len(r.OpenedURLs()) != 0 {
			t.Fatal("Create variant link did not activate exactly once through its router")
		}
	}
	choose("Badge configuration", "Deleting and Generating")
	if !access(ui.RoleText, "Deleting").Busy || !access(ui.RoleText, "Generating").Busy || access(ui.RoleText, "Deleting").Focusable {
		t.Fatal("static loading badges lost loading semantics or gained a button")
	}
	r.Move(0, 0)
	r.SetPreferences(ui.Preferences{})
	r.Advance(time.Second)
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(250 * time.Millisecond)
	if bytes.Equal(before, r.Image().Pix) || !r.FramePending() {
		t.Fatal("the source's leading and trailing spinner content did not animate")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	dark = true
	r.Frame()
	checkDanger("Deleting")
	badgeCompositionSnapshot(t, r, "badge-loading-dark")
	choose("Badge configuration", "Create: Custom Colors")
	checkColors := func() {
		t.Helper()
		blues := []ui.AccessInfo{}
		for _, a := range r.AccessAll() {
			if a.Role == ui.RoleText && a.Name == "Blue" {
				blues = append(blues, a)
			}
		}
		if len(blues) != 2 {
			t.Fatal("Create's solid and pastel Blue were collapsed")
		}
		for i, shade := range []int{600, 50} {
			if dark && i == 1 {
				shade = 950
			}
			want, _ := theme.Palette("blue", shade)
			a := blues[i].Bounds
			pixel := r.Image().RGBAAt(int(a.X+a.W/2), int(a.Y+a.H-3))
			if pixel != (color.RGBA{R: want.R, G: want.G, B: want.B, A: 255}) {
				t.Fatalf("source palette shade %d pixel=%v want=%v", shade, pixel, want)
			}
		}
	}
	checkColors()
	badgeCompositionSnapshot(t, r, "badge-colors-dark")
	dark = false
	r.Frame()
	checkColors()
	badgeCompositionSnapshot(t, r, "badge-colors-light")
	choose("Badge configuration", "Create: Long Text")
	r.SetSize(260, 700)
	if p.longScroll.MaxX <= 0 {
		t.Fatal("the source nowrap badge was clipped instead of becoming horizontally reachable")
	}
	viewport := access(ui.RoleRegion, "Long badge viewport")
	r.Scroll(viewport.Bounds.X+viewport.Bounds.W/2, viewport.Bounds.Y+viewport.Bounds.H/2, p.longScroll.MaxX, 0)
	if p.longScroll.X <= 0 || !r.HasText("A badge with a lot of text to see how it wraps") {
		t.Fatal("the complete original long badge could not be reached by native scrolling")
	}
	r.SetSize(620, 700)
	choose("Badge configuration", "Translated Badges")
	choose("Badge language", "עברית")
	dark = true
	r.Frame()
	verified := access(ui.RoleText, "מאומת")
	bookmark := access(ui.RoleText, "סימנייה")
	if verified.Bounds.X <= bookmark.Bounds.X || verified.Focusable || bookmark.Focusable {
		t.Fatal("Hebrew source order or inert badge semantics changed")
	}
	badgeCompositionSnapshot(t, r, "badge-hebrew-dark")
	visible = false
	r.Frame()
	visible, native = true, false
	r.Frame()
	choose("Badge configuration", "Open Link")
	if p.recipe != "link" || p.language != "he" || p.action != "Local badge action: Link" || p.activations != 6 || p.popupError != "" || p.router.Path() != "/badge-examples" {
		t.Fatal("source model or native-router state was lost across hide/style/recipe changes")
	}
	if len(b.surfaces) == 0 || b.surfaces[0].spec.SystemPopover || h.IsOpen("source-badge-config") {
		t.Fatal("styled source selector was not an independent Flyout that closes on choice")
	}
}
