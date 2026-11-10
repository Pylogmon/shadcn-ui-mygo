package gallery

import (
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func sliderCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_SLIDER_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("slider image %v/%v", err, closeErr)
	}
}

// A single public-input flow exercises the source numeric controls and their
// independent application drafts. Software images do not establish OS/AT input.
func TestSliderCompositionsSourceWorkflow(t *testing.T) {
	p := &sliderCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Wrap().Children(func() {
				if ui.Button(c, "Leave slider examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show slider examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, name)
		if !ok {
			t.Fatalf("missing native %v %q", role, name)
		}
		return a
	}
	choose := func(key string) {
		t.Helper()
		index := -1
		for i, option := range sliderCompositionRecipes {
			if option.Value == key {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatalf("unknown source recipe %q", key)
		}
		if err := r.Click("Slider configuration"); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(sliderCompositionRecipes[index].Label); err != nil {
				t.Fatal(err)
			}
		} else {
			surface := b.surfaces[len(b.surfaces)-1]
			surface.tester.Key(0, ui.KeyHome)
			for range index {
				surface.tester.Key(0, ui.KeyDown)
			}
			surface.tester.Key(0, ui.KeyEnter)
			if !surface.didClose {
				t.Fatal("styled slider selector failed to close")
			}
		}
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("recipe %q reached %q error=%s", key, p.recipe, p.popupError)
		}
	}
	clickThumb := func(name string) {
		t.Helper()
		a := access(ui.RoleSlider, name)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	}
	group := access(ui.RoleGroup, "Slider")
	thumb := access(ui.RoleSlider, "Slider")
	if p.recipe != "demo" || len(sliderCompositionRecipes) != 13 || thumb.Now != 75 || thumb.Step != 1 || group.Bounds.W != 320 || !thumb.Horizontal {
		t.Fatal("docs source default, domain or 320-DIP width changed")
	}
	r.ClickAt(group.Bounds.X+8+(group.Bounds.W-16)*.25, group.Bounds.Y+group.Bounds.H/2)
	r.Key(0, ui.KeyRight)
	r.Frame()
	if p.drafts["demo"].values[0] != 26 || p.drafts["demo"].changes != 2 || !access(ui.RoleSlider, "Slider").Focused {
		t.Fatal("source slider pointer/keyboard input did not update once per value change")
	}
	for _, recipe := range sliderCompositionRecipes {
		choose(recipe.Value)
	}
	choose("create-basic")
	if access(ui.RoleSlider, "Slider").Now != 50 || access(ui.RoleGroup, "Slider").Bounds.W <= 320 {
		t.Fatal("Create Basic inherited the docs value or maximum width")
	}
	choose("range")
	if p.draft().values[0] != 25 || p.draft().values[1] != 50 || access(ui.RoleSlider, "Slider minimum").Step != 5 {
		t.Fatal("source range defaults or step changed")
	}
	clickThumb("Slider minimum")
	r.Key(0, ui.KeyRight)
	if !slices.Equal(p.draft().values, []float64{30, 50}) {
		t.Fatal("range input did not move the actual lower thumb by five")
	}
	r.Key(0, ui.KeyEnd)
	if !slices.Equal(p.draft().values, []float64{50, 50}) {
		t.Fatal("source range crossing boundary did not keep ordered endpoints")
	}
	r.Key(0, ui.KeyLeft)
	r.Frame()
	sliderCompositionSnapshot(t, r, "slider-range-light")
	choose("multiple")
	clickThumb("Slider Thumb 2")
	r.Key(0, ui.KeyRight)
	if !slices.Equal(p.draft().values, []float64{10, 30, 70}) || access(ui.RoleSlider, "Slider Thumb 2").Step != 10 {
		t.Fatal("native multiple thumb input lost its independent slot or step")
	}
	choose("disabled")
	disabledChanges := p.draft().changes
	if !access(ui.RoleSlider, "Slider").Disabled {
		t.Fatal("source disabled did not reach the native thumb")
	}
	clickThumb("Slider")
	r.Key(0, ui.KeyRight)
	if p.draft().values[0] != 50 || p.draft().changes != disabledChanges {
		t.Fatal("disabled source accepted pointer or keyboard input")
	}
	choose("vertical")
	first, second := access(ui.RoleSlider, "Vertical slider 1"), access(ui.RoleSlider, "Vertical slider 2")
	firstGroup, secondGroup := access(ui.RoleGroup, "Vertical slider 1"), access(ui.RoleGroup, "Vertical slider 2")
	if !first.Vertical || !second.Vertical || first.Now != 50 || second.Now != 25 || firstGroup.Bounds.H != 160 || secondGroup.Bounds.H != 160 || secondGroup.Bounds.X-firstGroup.Bounds.X != 44 {
		t.Fatal("vertical source pair lost its values, h-40 height or gap-6 spacing")
	}
	clickThumb("Vertical slider 2")
	r.Key(0, ui.KeyUp)
	if p.draft().vertical != [2]float64{50, 26} {
		t.Fatal("vertical keyboard input changed the wrong native slider")
	}
	r.ClickAt(firstGroup.Bounds.X+firstGroup.Bounds.W/2, firstGroup.Bounds.Y)
	if p.draft().vertical != [2]float64{100, 26} {
		t.Fatal("vertical native pointer input did not reach its upper endpoint")
	}
	dark = true
	r.Frame()
	sliderCompositionSnapshot(t, r, "slider-vertical-dark")
	choose("controlled")
	if text, ok := r.Find("0.3, 0.7"); !ok || text.W == 0 {
		t.Fatal("source controlled temperature readout is missing")
	}
	clickThumb("Temperature minimum")
	r.Key(0, ui.KeyRight)
	r.Frame()
	if math.Abs(p.draft().values[0]-.4) > 1e-9 || math.Abs(p.draft().values[1]-.7) > 1e-9 || access(ui.RoleSlider, "Temperature minimum").Step != .1 {
		t.Fatal("fractional controlled source did not retain its real numeric domain")
	}
	if _, ok := r.Find("0.4, 0.7"); !ok {
		t.Fatal("controlled source readout did not follow native thumb input")
	}
	sliderCompositionSnapshot(t, r, "slider-controlled-dark")
	choose("create-controlled")
	if math.Abs(p.draft().values[0]-.3) > 1e-9 || math.Abs(p.draft().values[1]-.7) > 1e-9 {
		t.Fatal("Create controlled slider shared docs storage")
	}
	choose("rtl")
	if err := r.Click("Slider language"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	r.SetSize(360, 700)
	group = access(ui.RoleGroup, "Slider")
	if group.Bounds.X < 20 || group.Bounds.X+group.Bounds.W > 340 || group.Bounds.W >= 320 {
		t.Fatal("the RTL native slider overflowed its narrow parent")
	}
	clickThumb("Slider")
	r.Key(0, ui.KeyRight)
	if p.draft().values[0] != 74 {
		t.Fatal("RTL Right did not move the source slider toward its lower value")
	}
	r.Key(0, ui.KeyUp)
	r.ClickAt(group.Bounds.X, group.Bounds.Y+group.Bounds.H/2)
	r.Frame()
	if p.draft().values[0] != 100 {
		t.Fatal("RTL native pointer geometry did not put the upper endpoint on the left")
	}
	sliderCompositionSnapshot(t, r, "slider-hebrew-narrow-dark")
	if err := r.Click("Leave slider examples"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.AccessByRole(ui.RoleSlider, "Slider"); ok {
		t.Fatal("source native thumb stayed mounted on another page")
	}
	if err := r.Click("Show slider examples"); err != nil {
		t.Fatal(err)
	}
	if access(ui.RoleSlider, "Slider").Now != 100 || p.draft().values[0] != 100 {
		t.Fatal("source hide/page return lost the independent RTL value")
	}
	r.SetSize(620, 700)
	native = false
	r.Frame()
	choose("demo")
	if p.draft().values[0] != 26 || access(ui.RoleSlider, "Slider").Now != 26 || p.drafts["range"].values[0] != 45 || math.Abs(p.drafts["controlled"].values[1]-.7) > 1e-9 {
		t.Fatal("changing selector style/source recipe crossed independent slider drafts")
	}
}
