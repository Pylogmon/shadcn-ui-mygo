package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func buttonSourceSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_BUTTON_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("button image %v/%v", err, closeErr)
	}
}

// One native public-input workflow uses source configurations and real button /
// link targets. Tester images and its injected Flyout backend are not evidence
// of operating-system windows, assistive technology or platform text rasterizers.
func TestButtonCompositionsSourceWorkflow(t *testing.T) {
	p := &buttonCompositionPreviewState{}
	b := &creatorPopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("neutral", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave button examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show button examples").Clicked() {
					visible = true
				}
			})
			ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Role(ui.RoleRegion).Label("Button workspace").Children(func() {
				ui.Column(c).FillWidth().MinWidth(0).Children(func() {
					if visible {
						p.view(c, h, native)
					} else {
						ui.Text(c, "Another application page")
					}
				})
			})
		})
	}, 620, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		for range 8 {
			a, ok := r.AccessByRole(role, name)
			v, viewport := r.AccessByRole(ui.RoleRegion, "Button workspace")
			if !ok || !viewport {
				t.Fatalf("missing source target %v/%s", role, name)
			}
			if a.Bounds.Y >= v.Bounds.Y && a.Bounds.Y+a.Bounds.H <= v.Bounds.Y+v.Bounds.H {
				if a.Bounds.W <= 0 || a.Bounds.X < v.Bounds.X-1 || a.Bounds.X+a.Bounds.W > v.Bounds.X+v.Bounds.W+1 {
					t.Fatalf("source target %s exceeds workspace: %v in %v", name, a.Bounds, v.Bounds)
				}
				return a
			}
			r.Scroll(v.Bounds.X+v.Bounds.W-8, v.Bounds.Y+v.Bounds.H/2, 0, a.Bounds.Y+a.Bounds.H/2-v.Bounds.Y-v.Bounds.H/2)
		}
		t.Fatalf("cannot reveal source target %s", name)
		return ui.AccessInfo{}
	}
	click := func(role ui.Role, name string) {
		t.Helper()
		a := access(role, name)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	choose := func(value string) {
		t.Helper()
		index := -1
		for i, o := range buttonCompositionRecipes {
			if o.Value == value {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatalf("unknown source recipe %s", value)
		}
		role := ui.RoleMenuButton
		if !native {
			role = ui.RoleButton
		}
		click(role, "Button configuration")
		if native {
			if err := r.ChooseMenuItem(buttonCompositionRecipes[index].Label); err != nil {
				t.Fatal(err)
			}
		} else {
			child := b.surfaces[len(b.surfaces)-1].tester
			child.Key(0, ui.KeyHome)
			for range index {
				child.Key(0, ui.KeyDown)
			}
			child.Key(0, ui.KeyEnter)
		}
		r.Frame()
		if p.recipe != value || p.popupError != "" {
			t.Fatalf("source configuration %s built %s error=%s", value, p.recipe, p.popupError)
		}
	}
	if p.recipe != "demo" {
		t.Fatal("docs demo is not the source default")
	}
	click(ui.RoleButton, "Button")
	r.Key(0, ui.KeySpace)
	r.Key(0, ui.KeyEnter)
	if p.action != "Local button action: Button" || p.counts["demo"] != 3 {
		t.Fatal("native pointer/Space/Enter did not activate the source button exactly once each")
	}
	for _, recipe := range buttonCompositionRecipes {
		choose(recipe.Value)
	}
	choose("size")
	for _, s := range []struct {
		label  string
		height float32
	}{{"Extra Small", 24}, {"Small", 32}, {"Default", 36}, {"Large", 40}} {
		a := access(ui.RoleButton, s.label)
		if a.Bounds.H != s.height {
			t.Fatalf("source size %s height=%g want%g", s.label, a.Bounds.H, s.height)
		}
	}
	click(ui.RoleButton, "Submit (small)")
	if p.counts["size"] != 1 {
		t.Fatal("icon-only source button did not activate")
	}
	choose("create-variants")
	plainOutline := access(ui.RoleButton, "Outline (small)").Bounds.W
	plainSecondary := access(ui.RoleButton, "Secondary (small)").Bounds.W
	choose("create-icon-right")
	if delta := access(ui.RoleButton, "Outline (small)").Bounds.W - plainOutline; delta < 19.9 || delta > 20.1 {
		t.Fatalf("unmarked Small Outline icon width delta=%g want20", delta)
	}
	if delta := access(ui.RoleButton, "Secondary (small)").Bounds.W - plainSecondary; delta < 15.9 || delta > 16.1 {
		t.Fatalf("inline-end Small icon width delta=%g want16", delta)
	}
	buttonSourceSnapshot(t, r, "button-create-icon-right-light")
	click(ui.RoleButton, "Outline (large)")
	if p.action != "Local button action: Outline (large)" {
		t.Fatal("source icon matrix button was inert")
	}
	choose("create-invalid")
	invalid := access(ui.RoleButton, "Default (extra small)")
	if !invalid.Invalid || invalid.Disabled {
		t.Fatal("aria-invalid source became disabled or lost its semantic state")
	}
	click(ui.RoleButton, "Default (extra small)")
	if p.counts["create-invalid"] != 1 {
		t.Fatal("invalid source button could not be activated")
	}
	dark = true
	r.Frame()
	choose("spinner")
	for _, label := range []string{"Generating", "Downloading"} {
		a := access(ui.RoleButton, label)
		if !a.Disabled || !a.Busy || a.Focusable {
			t.Fatal("loading source lost disabled/busy semantics or remained focusable")
		}
		click(ui.RoleButton, label)
	}
	if p.counts["spinner"] != 0 {
		t.Fatal("disabled loading source accepted an action")
	}
	buttonSourceSnapshot(t, r, "button-loading-dark")
	// A high-contrast witness separates actual opacity from antialiased glyphs
	// and translucent borders. The old explicitly-disabled source drew at64;
	// source0.5 must draw128, preserving loading-only and inherited behavior.
	opacity := ui.NewTester(func(c *ui.Context) {
		ui.Row(c).Fill().Gap(8).Background(ui.RGB(0, 0, 0)).Children(func() {
			control.Button(c, "", control.ButtonProps{Label: "Disabled opacity", Disabled: true}).Size(40, 40).Background(ui.RGB(255, 255, 255)).Border(0, ui.Transparent)
			control.Button(c, "", control.ButtonProps{Label: "Loading opacity", Loading: true}).Size(40, 40).Background(ui.RGB(255, 255, 255)).Border(0, ui.Transparent)
			ui.Row(c).Disabled(true).Children(func() {
				control.Button(c, "", control.ButtonProps{Label: "Inherited opacity"}).Size(40, 40).Background(ui.RGB(255, 255, 255)).Border(0, ui.Transparent)
			})
		})
	}, 180, 60)
	for _, name := range []string{"Disabled opacity", "Loading opacity", "Inherited opacity"} {
		a, ok := opacity.AccessByRole(ui.RoleButton, name)
		if !ok {
			t.Fatal(name)
		}
		pixel := opacity.Image().RGBAAt(int(a.Bounds.X+5), int(a.Bounds.Y+a.Bounds.H/2))
		if pixel.R < 127 || pixel.R > 128 || pixel.G != pixel.R || pixel.B != pixel.R {
			t.Fatalf("%s pixel=%v want one half-opacity layer", name, pixel)
		}
	}
	dark = false
	r.Frame()
	choose("render")
	click(ui.RoleLink, "Login")
	if p.action != "Local button link: Login" || p.counts["render"] != 1 || len(r.OpenedURLs()) != 0 {
		t.Fatal("actual native hash link did not stay in its router")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Login"); ok {
		t.Fatal("rendered link retained an extra button target")
	}
	r.Key(0, ui.KeySpace)
	if p.counts["render"] != 1 {
		t.Fatal("Space activated a native Link")
	}
	r.Key(0, ui.KeyEnter)
	if p.counts["render"] != 2 || len(r.OpenedURLs()) != 0 {
		t.Fatal("Enter did not activate the true source hash Link exactly once")
	}
	login := access(ui.RoleLink, "Login").Bounds
	ink := 0
	for y := int(login.Y + 6); y < int(login.Y+login.H-6); y++ {
		for x := int(login.X + 6); x < int(login.X+login.W-6); x++ {
			q := r.Image().RGBAAt(x, y)
			if q.R < 100 && q.G < 100 && q.B < 100 {
				ink++
			}
		}
	}
	if ink < 8 {
		t.Fatal("native Link target background hid the source Login face")
	}
	buttonSourceSnapshot(t, r, "button-render-link-light")
	choose("create-examples")
	click(ui.RoleButton, "Cancel")
	click(ui.RoleLink, "Link")
	if p.action != "Local button link: Link" || p.counts["create-examples"] != 2 || len(r.OpenedURLs()) != 0 {
		t.Fatal("Create's default anchor or independent native action group was lost")
	}
	choose("link")
	click(ui.RoleButton, "Link")
	link := access(ui.RoleButton, "Link").Bounds
	linkInk := func() int {
		n := 0
		for y := int(link.Y + 2); y < int(link.Y+link.H-2); y++ {
			for x := int(link.X + 2); x < int(link.X+link.W-2); x++ {
				q := r.Image().RGBAAt(x, y)
				if q.R < 100 && q.G < 100 && q.B < 100 {
					n++
				}
			}
		}
		return n
	}
	r.Move(0, 0)
	r.Frame()
	quiet := linkInk()
	r.Move(link.X+link.W/2, link.Y+link.H/2)
	r.Frame()
	hovered := linkInk()
	if hovered <= quiet+5 {
		t.Fatalf("source Link variant hover omitted visible underline: %d→%d", quiet, hovered)
	}
	r.Key(0, ui.KeySpace)
	if p.counts["link"] != 2 {
		t.Fatal("the link visual variant lost its source Button semantics")
	}
	choose("rtl")
	click(ui.RoleMenuButton, "Button language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	click(ui.RoleButton, "שלח")
	if p.action != "Local button action: שלח" || p.language != "he" {
		t.Fatal("translated source action lost its verbatim text")
	}
	if !access(ui.RoleButton, "טוען").Disabled {
		t.Fatal("translated loading button became interactive")
	}
	dark = true
	r.Frame()
	buttonSourceSnapshot(t, r, "button-rtl-hebrew-dark")
	// The global presentation selector changes the source configuration picker;
	// actual source actions remain native MyGo buttons in both modes.
	native = false
	r.Frame()
	choose("with-icon")
	click(ui.RoleButton, "Fork")
	if p.action != "Local button action: Fork" {
		t.Fatal("Flyout configuration selection lost the current source native button")
	}
	choose("demo")
	if p.counts["demo"] != 3 || p.action != "Local button action: Button" {
		t.Fatal("recipe changes replaced the independent source action history")
	}
	if err := r.Click("Leave button examples"); err != nil {
		t.Fatal(err)
	}
	if err := r.Click("Show button examples"); err != nil {
		t.Fatal(err)
	}
	if p.recipe != "demo" || p.counts["demo"] != 3 || p.action != "Local button action: Button" {
		t.Fatal("source state did not survive leaving and returning")
	}
}
