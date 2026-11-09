package gallery

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Match the native backend's optional live-size/focus capabilities. Keeping
// these on the existing test surface also exercises the Gallery's actual Host.
func (s *creatorPopupSurface) SetSize(width, height int) {
	s.spec.Width, s.spec.Height = width, height
	if s.tester != nil {
		s.tester.SetSize(width, height)
	}
}
func (s *creatorPopupSurface) Focus() {
	if s.tester != nil {
		s.tester.Frame()
	}
}

func advancedPreviewTester(s *State, b *creatorPopupBackend) *ui.Tester {
	s.Popovers = desktop.NewHostWithBackend(b)
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		s.Popovers.BeginFrame()
		defer s.Popovers.EndFrame()
		ui.Column(c).Fill().Padding(24).Gap(20).Children(func() {
			switch s.Page {
			case "navigation":
				s.menubarPreview(c)
				s.routePanelsPreview(c)
			case "controls":
				s.sliderRangesPreview(c)
			case "display":
				s.workspacePreview(c)
			default:
				ui.Text(c, "Other preview")
			}
		})
	}, 1000, 700)
}

func advancedAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q: %+v", role, label, r.AccessAll())
	}
	return a
}
func advancedClick(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	if err := r.Click(label); err != nil {
		t.Fatal(err)
	}
}

func TestCreatorAdvancedMenubarPairsSwitchChooseAndCleanUp(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			s := New()
			s.Page, s.Config.NativeMenus = "navigation", native
			b := &creatorPopupBackend{}
			parent := advancedPreviewTester(s, b)
			advancedClick(t, parent, "File")
			if native {
				if len(b.surfaces) != 0 || !slices.Contains(parent.Menu(), "New project") || parent.HasText("New project") {
					t.Fatalf("native menubar did not route to system: surfaces=%d menu=%v", len(b.surfaces), parent.Menu())
				}
				if err := parent.ChooseMenuItem("Open project"); err != nil {
					t.Fatal(err)
				}
				if s.pendingMenuAction != "open" || len(parent.Menu()) != 0 {
					t.Fatalf("system action did not reach Gallery: %q menu=%v", s.pendingMenuAction, parent.Menu())
				}
			} else {
				if len(b.surfaces) != 1 || b.surfaces[0].spec.SystemPopover || parent.HasText("New project") {
					t.Fatal("themed menubar did not create an independent Flyout")
				}
				popup := b.surfaces[0]
				popup.tester.Advance(time.Second)
				if !popup.tester.HasText("New project") {
					t.Fatal("File content missing from child")
				}
				popup.tester.Key(0, ui.KeyRight)
				parent.Frame()
				popup.tester.Frame()
				if len(b.surfaces) != 1 || s.Menubar.Active != "edit" || !popup.tester.HasText("Undo") {
					t.Fatalf("sibling switch replaced viewport or retained old menu: surfaces=%d state=%+v texts=%v", len(b.surfaces), s.Menubar, popup.tester.Texts())
				}
				advancedClick(t, popup.tester, "Undo")
				parent.Frame()
				if s.pendingMenuAction != "undo" || s.Popovers.IsOpen("preview-menubar") {
					t.Fatalf("child result did not reach Gallery once: %q open=%v", s.pendingMenuAction, s.Popovers.IsOpen("preview-menubar"))
				}
				s.pendingMenuAction = ""
				parent.Frame()
				if s.pendingMenuAction != "" {
					t.Fatal("menubar action was repeated")
				}
				advancedClick(t, parent, "File")
				if !s.Popovers.IsOpen("preview-menubar") {
					t.Fatal("menubar did not reopen")
				}
				s.Page = "overview"
				parent.Frame()
				if s.Popovers.IsPresent("preview-menubar") || s.Menubar.Open {
					t.Fatal("leaving navigation retained its native child")
				}
			}
		})
	}
}

func TestCreatorAdvancedNavigationPairsChooseRoutesAndCleanUp(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			s := New()
			s.Page, s.Config.NativeMenus = "navigation", native
			b := &creatorPopupBackend{}
			parent := advancedPreviewTester(s, b)
			advancedClick(t, parent, "Documentation")
			if native {
				if len(b.surfaces) != 0 || !slices.Contains(parent.Menu(), "Themes") || parent.HasText("Colors and appearance that work together.") {
					t.Fatal("native route panel did not use system menu")
				}
				if err := parent.ChooseMenuItem("Themes"); err != nil {
					t.Fatal(err)
				}
				if s.navRoute != "themes" || !parent.HasText("Selected: themes") {
					t.Fatalf("native documentation route=%q", s.navRoute)
				}
				advancedClick(t, parent, "Components")
				if err := parent.ChooseMenuItem("Buttons"); err != nil {
					t.Fatal(err)
				}
				if s.navRoute != "buttons" || len(parent.Menu()) != 0 {
					t.Fatalf("native component route=%q menu=%v", s.navRoute, parent.Menu())
				}
			} else {
				if len(b.surfaces) != 1 || parent.HasText("Colors and appearance that work together.") {
					t.Fatal("documentation was not independently hosted")
				}
				popup := b.surfaces[0]
				popup.tester.Advance(time.Second)
				popup.tester.Key(0, ui.KeyRight)
				parent.Frame()
				popup.tester.Frame()
				parent.Advance(400 * time.Millisecond)
				popup.tester.Advance(400 * time.Millisecond)
				if _, ok := popup.tester.AccessByRole(ui.RoleButton, "Browse components"); !ok || len(b.surfaces) != 1 {
					t.Fatal("Gallery keyboard sibling switch was overridden by the pointer left on Documentation")
				}
				popup.tester.Key(0, ui.KeyLeft)
				parent.Frame()
				popup.tester.Frame()
				parent.Advance(400 * time.Millisecond)
				popup.tester.Advance(400 * time.Millisecond)
				advancedClick(t, popup.tester, "Themes")
				parent.Frame()
				if s.navRoute != "themes" || !parent.HasText("Selected: themes") || s.Popovers.IsOpen("preview-route-panels") {
					t.Fatalf("documentation route=%q open=%v", s.navRoute, s.Popovers.IsOpen("preview-route-panels"))
				}
				advancedClick(t, parent, "Components")
				popup = b.surfaces[len(b.surfaces)-1]
				popup.tester.Advance(time.Second)
				advancedClick(t, popup.tester, "Browse components")
				parent.Frame()
				if s.navRoute != "buttons" || !parent.HasText("Selected: buttons") {
					t.Fatalf("custom content choose did not reach Gallery: %q", s.navRoute)
				}
				advancedClick(t, parent, "Components")
				parent.Frame()
				if !s.Popovers.IsOpen("preview-route-panels") {
					t.Fatal("custom panel did not reopen")
				}
				s.Page = "overview"
				parent.Frame()
				if s.Popovers.IsPresent("preview-route-panels") {
					t.Fatal("leaving navigation kept independent route panel alive")
				}
			}
		})
	}
}

func TestCreatorAdvancedWorkspaceAndSliderStateSurvivesPageChanges(t *testing.T) {
	s := New()
	s.Page = "controls"
	b := &creatorPopupBackend{}
	parent := advancedPreviewTester(s, b)
	thumb := advancedAccess(t, parent, ui.RoleSlider, "Price minimum")
	parent.ClickAt(thumb.Bounds.X+thumb.Bounds.W/2, thumb.Bounds.Y+thumb.Bounds.H/2)
	parent.Key(0, ui.KeyRight)
	if s.SliderRange != [2]float64{30, 75} {
		t.Fatalf("price thumb not bound to Gallery state: %v", s.SliderRange)
	}
	thumb = advancedAccess(t, parent, ui.RoleSlider, "Bass")
	parent.ClickAt(thumb.Bounds.X+8, thumb.Bounds.Y+8)
	parent.Key(0, ui.KeyUp)
	if s.SliderBands[0] != 25 {
		t.Fatalf("equalizer thumb not bound to Gallery state: %v", s.SliderBands)
	}
	thumb = advancedAccess(t, parent, ui.RoleSlider, "Output level")
	parent.ClickAt(thumb.Bounds.X+8, thumb.Bounds.Y+8)
	parent.Key(0, ui.KeyDown)
	if s.VerticalLevel != 55 {
		t.Fatalf("vertical output not bound to Gallery state: %v", s.VerticalLevel)
	}
	s.Page = "display"
	parent.Frame()
	divider := advancedAccess(t, parent, ui.RoleSplitter, "Resize Files")
	x, y := divider.Bounds.X+divider.Bounds.W/2, divider.Bounds.Y+divider.Bounds.H/2
	files := s.WorkspaceLayout["files"]
	source := s.EditorLayout["source"]
	parent.Press(x, y)
	parent.Move(x+60, y)
	parent.Release(x+60, y)
	parent.Frame()
	if s.WorkspaceLayout["files"] <= files || math.Abs(s.EditorLayout["source"]-source) > 1e-6 {
		t.Fatalf("outer/nested layout binding incorrect: workspace=%v editor=%v", s.WorkspaceLayout, s.EditorLayout)
	}
	outer := make(display.ResizableLayout)
	for k, v := range s.WorkspaceLayout {
		outer[k] = v
	}
	divider = advancedAccess(t, parent, ui.RoleSplitter, "Resize Source")
	x, y = divider.Bounds.X+divider.Bounds.W/2, divider.Bounds.Y+divider.Bounds.H/2
	parent.Press(x, y)
	parent.Move(x, y+35)
	parent.Release(x, y+35)
	parent.Frame()
	if s.EditorLayout["source"] <= source {
		t.Fatalf("nested vertical splitter not bound: %v", s.EditorLayout)
	}
	for k, v := range outer {
		if math.Abs(s.WorkspaceLayout[k]-v) > 1e-6 {
			t.Fatal("nested resize changed outer layout")
		}
	}
	divider = advancedAccess(t, parent, ui.RoleSplitter, "Resize Files")
	parent.ClickAt(divider.Bounds.X+divider.Bounds.W/2, divider.Bounds.Y+20)
	parent.Key(0, ui.KeyEnter)
	parent.Frame()
	if s.WorkspaceLayout["files"] != 0 {
		t.Fatalf("file-pane collapse not bound: %v", s.WorkspaceLayout)
	}
	inner := s.EditorLayout["source"]
	s.Page = "controls"
	parent.Frame()
	if s.SliderRange != [2]float64{30, 75} || s.SliderBands[0] != 25 || s.VerticalLevel != 55 || advancedAccess(t, parent, ui.RoleSlider, "Price minimum").Now != 30 {
		t.Fatal("page switch reset controlled sliders")
	}
	s.Page = "display"
	parent.Frame()
	if s.WorkspaceLayout["files"] != 0 || math.Abs(s.EditorLayout["source"]-inner) > 1e-6 {
		t.Fatal("page switch reset workspace layout")
	}
	parent.Advance(time.Second)
	if parent.FramePending() {
		t.Fatal("settled workspace scheduled idle redraws")
	}
}
