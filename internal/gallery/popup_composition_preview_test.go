package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type sourcePopupBackend struct {
	surfaces []*sourcePopupSurface
}
type sourcePopupSurface struct {
	spec     desktop.Spec
	tester   *ui.Tester
	closed   func()
	didClose bool
	font     string
}

func (b *sourcePopupBackend) Open(spec desktop.Spec, view func(*ui.Context), closed func()) (desktop.Surface, error) {
	s := &sourcePopupSurface{spec: spec, closed: closed}
	b.surfaces = append(b.surfaces, s)
	platform := *ui.DarkTheme()
	platform.Font, platform.FontSize = "monospace", 17
	s.tester = ui.NewTesterWith(func(c *ui.Context) {
		c.SetTheme(&platform)
		view(c)
		s.font = c.Theme().Font
	}, max(1, spec.Width), max(1, spec.Height), ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}, Dark: true})
	return s, nil
}
func (b *sourcePopupBackend) InvalidateParent()   {}
func (s *sourcePopupSurface) SetAnchor(r ui.Rect) { s.spec.Anchor = r }
func (s *sourcePopupSurface) Invalidate()         { s.tester.Frame() }
func (s *sourcePopupSurface) SetSize(width, height int) {
	s.spec.Width, s.spec.Height = width, height
	if s.tester != nil {
		s.tester.SetSize(width, height)
	}
}
func (s *sourcePopupSurface) IsFocused() bool { return false }
func (s *sourcePopupSurface) Close() {
	if !s.didClose {
		s.didClose = true
		s.closed()
	}
}

func popupCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_POPUP_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("popup image %v/%v", err, closeErr)
	}
}

// One source workflow uses independent native Contexts. An injected public
// Backend and real ui.Tester input do not claim an operating-system matrix.
func TestPopupCompositionsIndependentSourceWorkflow(t *testing.T) {
	p := &popupCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(16).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave popup examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show popup examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, label string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, label)
		if !ok {
			t.Fatalf("missing %q role %v", label, role)
		}
		return a
	}
	click := func(label string) {
		t.Helper()
		if err := r.Click(label); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	choose := func(selector, label string) {
		t.Helper()
		click(selector)
		if err := r.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	latest := func() *sourcePopupSurface {
		t.Helper()
		if len(b.surfaces) == 0 {
			t.Fatal("no independent native surface")
		}
		return b.surfaces[len(b.surfaces)-1]
	}
	closeChild := func(s *sourcePopupSurface) { t.Helper(); s.Close(); r.Frame() }
	if p.popupError != "" || p.popoverRecipe != "dimensions" {
		t.Fatalf("initial quiet state: %q", p.popupError)
	}
	click("Open dimensions popover")
	s := latest()
	if r.HasText("Max. width") || !s.tester.HasText("Max. width") || !s.spec.SystemPopover || s.spec.Placement != mygo.PlacementBottom {
		t.Fatal("source dimensions did not use an independent centered native window")
	}
	if s.font != "monospace" {
		t.Fatalf("Native lost its own platform font: %q", s.font)
	}
	if err := s.tester.Click("Width"); err != nil {
		t.Fatal(err)
	}
	s.tester.Key(ui.Cmd, ui.KeyA)
	s.tester.Type("75%")
	if p.dimensions[0] != "75%" {
		t.Fatal("child editor did not update application draft")
	}
	dialog, ok := s.tester.AccessByRole(ui.RoleDialog, "Dimensions")
	if !ok || dialog.Name != "Dimensions" || len(dialog.DescribedBy) != 1 || dialog.DescribedBy[0] != "Set the dimensions for the layer." {
		t.Fatalf("actual child dialog relationships: %+v", dialog)
	}
	closeChild(s)
	if !r.Focused("Open dimensions popover") {
		t.Fatal("accepted native close did not return to actual trigger")
	}
	click("Open dimensions popover")
	s = latest()
	if a, ok := s.tester.AccessByRole(ui.RoleTextField, "Width"); !ok || a.Value != "75%" {
		t.Fatal("reopen lost native draft")
	}
	closeChild(s)
	choose("Popover configuration", "Basic header")
	click("Open basic popover")
	s = latest()
	if s.spec.Width != 288 || s.spec.Placement != mygo.PlacementBottomStart || !s.tester.HasText("Dimensions") {
		t.Fatal("basic source width/align/header missing")
	}
	closeChild(s)
	choose("Popover configuration", "Dimensions form")
	click("Open dimensions form")
	s = latest()
	if err := s.tester.Click("Height"); err != nil {
		t.Fatal(err)
	}
	s.tester.Key(ui.Cmd, ui.KeyA)
	s.tester.Type("40px")
	if p.form[1] != "40px" || p.dimensions[2] != "25px" || s.spec.Width != 256 {
		t.Fatal("two-field draft leaked into four-field source")
	}
	popupCompositionSnapshot(t, s.tester, "popup-native-form-dark")
	closeChild(s)
	choose("Popover configuration", "Three alignments")
	for _, pair := range []struct {
		label     string
		placement mygo.Placement
	}{{"Start", mygo.PlacementBottomStart}, {"Center", mygo.PlacementBottom}, {"End", mygo.PlacementBottomEnd}} {
		click(pair.label)
		s = latest()
		if s.spec.Placement != pair.placement || s.spec.Width != 160 || !s.tester.HasText("Aligned to "+stringLower(pair.label)) {
			t.Fatalf("alignment %s: %+v", pair.label, s.spec)
		}
		closeChild(s)
	}
	choose("Popover configuration", "Six sides")
	for _, side := range popupSides {
		click(popupSideLabel(side, "en"))
		s = latest()
		if s.spec.Placement != popupPlacement(side, ui.LeftToRight) || s.spec.Gap != 4 {
			t.Fatalf("side %s: %+v", side, s.spec)
		}
		closeChild(s)
	}
	choose("Popover configuration", "Translated dimensions")
	choose("Popup language", "עברית")
	r.SetDirection(ui.RightToLeft)
	click(popupSideLabel("inline-start", "he"))
	s = latest()
	if s.spec.Placement != mygo.PlacementRight || !s.tester.HasText("מימדים") {
		t.Fatal("logical RTL popover or current translated title missing")
	}
	popupCompositionSnapshot(t, s.tester, "popup-hebrew-header-dark")
	closeChild(s)
	r.SetDirection(ui.LeftToRight)
	choose("Popover configuration", "Popover in Dialog")
	click("Open popover dialog")
	if !p.popoverDialog {
		t.Fatal("source parent Dialog did not open")
	}
	click("Open popover in dialog")
	s = latest()
	if !s.tester.HasText("Popover in Dialog") || !p.popoverDialog {
		t.Fatal("nested popup replaced its actual parent Dialog")
	}
	closeChild(s)
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if p.popoverDialog {
		t.Fatal("Dialog did not close after child dismissal")
	}
	// Pointer timing is source-specific; keyboard focus deliberately remains the
	// existing native immediate accessibility route.
	profile := access(ui.RoleButton, "Hover Here")
	r.Move(profile.Bounds.X+profile.Bounds.W/2, profile.Bounds.Y+profile.Bounds.H/2)
	r.Advance(9 * time.Millisecond)
	if h.IsOpen("source-hover:profile") {
		t.Fatal("profile ignored its 10ms pointer delay")
	}
	r.Advance(2 * time.Millisecond)
	s = latest()
	if !h.IsOpen("source-hover:profile") || !s.tester.HasText("@nextjs") || s.spec.SystemPopover || !s.spec.ShowInactive {
		t.Fatal("source profile/native inactive Flyout missing")
	}
	popupCompositionSnapshot(t, s.tester, "popup-nextjs-profile-dark")
	r.Move(0, 0)
	r.Advance(99 * time.Millisecond)
	if !h.IsOpen("source-hover:profile") {
		t.Fatal("profile closed before source100ms")
	}
	r.Advance(2 * time.Millisecond)
	r.Frame()
	if h.IsOpen("source-hover:profile") {
		t.Fatal("profile did not close after source100ms")
	}
	choose("Hover card configuration", "Four physical sides")
	for _, side := range []string{"left", "top", "bottom", "right"} {
		a := access(ui.RoleButton, popupSideLabel(side, "en"))
		r.Move(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Advance(99 * time.Millisecond)
		if h.IsOpen("source-hover:" + side) {
			t.Fatal("physical side ignored100ms")
		}
		r.Advance(2 * time.Millisecond)
		s = latest()
		if s.spec.Placement != popupPlacement(side, ui.LeftToRight) || !s.tester.HasText("Hover Card") {
			t.Fatalf("hover side %s missing", side)
		}
		r.Move(0, 0)
		r.Advance(101 * time.Millisecond)
		r.Frame()
	}
	choose("Hover card configuration", "Six physical and logical sides")
	for _, side := range popupSides {
		a := access(ui.RoleButton, popupSideLabel(side, "en"))
		r.Move(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Advance(101 * time.Millisecond)
		s = latest()
		if s.spec.Placement != popupPlacement(side, ui.LeftToRight) {
			t.Fatal("Create logical side missing")
		}
		r.Move(0, 0)
		r.Advance(101 * time.Millisecond)
		r.Frame()
	}
	choose("Hover card configuration", "Hover card in Dialog")
	click("Open hover card dialog")
	if !p.hoverDialog {
		t.Fatal("hover parent Dialog absent")
	}
	a := access(ui.RoleButton, "Hover me")
	r.Move(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Advance(101 * time.Millisecond)
	s = latest()
	if !s.tester.HasText("This hover card appears inside a dialog. Hover over the button to see it.") || !p.hoverDialog {
		t.Fatal("nested hover content or parent Dialog lost")
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if p.hoverDialog {
		t.Fatal("nested hover parent did not close")
	}
	choose("Hover card configuration", "Translated product")
	choose("Popup language", "עברית")
	r.SetDirection(ui.RightToLeft)
	dark = true
	native = false
	r.Frame()
	a = access(ui.RoleButton, popupSideLabel("inline-end", "he"))
	r.Move(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Advance(11 * time.Millisecond)
	s = latest()
	if s.spec.Placement != mygo.PlacementLeft || s.spec.SystemPopover || s.font == "monospace" || !s.tester.HasText("אוזניות אלחוטיות") || !s.tester.HasText("99.99 $") {
		t.Fatal("styled RTL translated product did not use current direction/theme")
	}
	popupCompositionSnapshot(t, s.tester, "popup-product-hebrew-dark")
	click("Leave popup examples")
	if h.IsPresent("source-hover:inline-end") {
		t.Fatal("hidden application page retained an independent child")
	}
	click("Show popup examples")
	if p.dimensions[0] != "75%" || p.form[1] != "40px" || p.popupError != "" {
		t.Fatal("page return lost independent draft or reported an error")
	}
}

func stringLower(text string) string {
	if text == "Start" {
		return "start"
	}
	if text == "Center" {
		return "center"
	}
	return "end"
}
