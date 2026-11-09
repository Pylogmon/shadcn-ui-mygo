package gallery

import (
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func recipeTester(s *State, module string) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		ui.Scroll(c).Fill().Children(func() {
			ui.Column(c).FillWidth().Padding(24).Gap(20).Children(func() {
				if s.Page == "other" {
					ui.Text(c, "Another preview")
					return
				}
				switch module {
				case "choices":
					s.recipeChoicesPreview(c)
				case "progress":
					s.recipeProgressPreview(c)
				case "loading":
					s.recipeAvatarPreview(c)
					s.recipeSkeletonPreview(c)
				}
			})
		})
	}, 900, 820, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func recipeAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	v, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing %s (role %v)", name, role)
	}
	return v
}
func recipeSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_RECIPE_SNAPSHOT_DIR")
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
	if err = png.Encode(f, r.Image()); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestCreatorRecipeRadioCardsAndGridKeepNativeSelection(t *testing.T) {
	s := New()
	originalPlan := s.Plan
	r := recipeTester(s, "choices")
	if !recipeAccess(t, r, ui.RoleRadio, "Plus").Checked || !recipeAccess(t, r, ui.RoleRadio, "Medium").Checked {
		t.Fatal("source defaults were not selected")
	}
	pro := recipeAccess(t, r, ui.RoleRadio, "Pro")
	// Activate the outer card edge rather than its indicator/text.
	r.ClickAt(pro.Bounds.X+pro.Bounds.W-5, pro.Bounds.Y+pro.Bounds.H/2)
	if s.recipes.plan != "pro" || !r.HasText("Selected plan: Pro") {
		t.Fatal("whole card did not select its native radio")
	}
	r.Key(0, ui.KeyDown)
	if s.recipes.plan != "enterprise" || !recipeAccess(t, r, ui.RoleRadio, "Enterprise").Checked {
		t.Fatal("native radio navigation did not update the application model")
	}
	medium := recipeAccess(t, r, ui.RoleRadio, "Medium")
	r.ClickAt(medium.Bounds.X+medium.Bounds.W/2, medium.Bounds.Y+medium.Bounds.H/2)
	r.Key(0, ui.KeyDown)
	if s.recipes.size != "large" || !r.HasText("Selected size: Large") {
		t.Fatal("visual grid lost ordinary native radio navigation")
	}
	recipeAccess(t, r, ui.RoleRadioGroup, "Workspace plan cards")
	recipeAccess(t, r, ui.RoleRadioGroup, "Preview size grid")
	recipeSnapshot(t, r, "radio-choices")
	s.Page = "other"
	r.Frame()
	s.Page = "controls"
	r.Frame()
	if !recipeAccess(t, r, ui.RoleRadio, "Enterprise").Checked || !recipeAccess(t, r, ui.RoleRadio, "Large").Checked || s.Plan != originalPlan {
		t.Fatal("page remount reset recipe choices or changed the existing Plan model")
	}
}
func TestCreatorRecipeProgressSharesItsActualSliderValue(t *testing.T) {
	s := New()
	r := recipeTester(s, "progress")
	fixed := recipeAccess(t, r, ui.RoleProgress, "Upload progress")
	moving := recipeAccess(t, r, ui.RoleProgress, "Transfer progress")
	if math.IsNaN(fixed.Now) || math.Abs(fixed.Now-.56) > .00001 || fixed.Value != "56%" || moving.Now != .5 || !r.HasText("50%") {
		t.Fatal("label/value source defaults are not real native progress")
	}
	recipeSnapshot(t, r, "progress")
	slider := recipeAccess(t, r, ui.RoleSlider, "Adjust transfer progress")
	r.ClickAt(slider.Bounds.X+slider.Bounds.W/2, slider.Bounds.Y+slider.Bounds.H/2)
	r.Key(0, ui.KeyEnd)
	moving = recipeAccess(t, r, ui.RoleProgress, "Transfer progress")
	if s.recipes.progress != 100 || moving.Now != 1 || moving.Value != "100%" || !r.HasText("100%") {
		t.Fatal("slider and visible/accessible progress did not share one value")
	}
	r.Key(0, ui.KeyHome)
	moving = recipeAccess(t, r, ui.RoleProgress, "Transfer progress")
	if moving.Now != 0 || moving.Value != "0%" || !r.HasText("0%") || recipeAccess(t, r, ui.RoleProgress, "Upload progress").Value != "56%" {
		t.Fatal("zero progress or independent fixed source example changed")
	}
}
func TestCreatorRecipeAvatarAndSkeletonComposeWithoutFakeFields(t *testing.T) {
	s := New()
	r := recipeTester(s, "loading")
	recipeAccess(t, r, ui.RoleImage, "Sofia Davis portrait")
	if r.HasText("SD") || !s.recipes.grayscale {
		t.Fatal("grayscale image did not use the decoded Avatar route")
	}
	recipeSnapshot(t, r, "avatar-gray-and-skeletons")
	toggle := recipeAccess(t, r, ui.RoleSwitch, "Grayscale portrait")
	r.ClickAt(toggle.Bounds.X+toggle.Bounds.W/2, toggle.Bounds.Y+toggle.Bounds.H/2)
	if s.recipes.grayscale || !r.HasText("Color portrait") {
		t.Fatal("portrait switch did not change application-owned image choice")
	}
	recipeSnapshot(t, r, "avatar-color-and-skeletons")
	pause := recipeAccess(t, r, ui.RoleSwitch, "Pause skeleton animations")
	r.ClickAt(pause.Bounds.X+pause.Bounds.W/2, pause.Bounds.Y+pause.Bounds.H/2)
	if !s.loadingStatic || !recipeAccess(t, r, ui.RoleGroup, "Form loading preview").Busy || !recipeAccess(t, r, ui.RoleGroup, "Table loading preview").Busy {
		t.Fatal("placeholder loading state or shared pause model was lost")
	}
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleTextField || a.Role == ui.RoleButton {
			t.Fatal("visual form/table skeletons exposed fake editable or submit controls")
		}
	}
}
