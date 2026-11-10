package gallery

import (
	"encoding/json"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/fonts"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type fontGalleryBackend struct {
	creatorPopupBackend
	roles theme.FontRoles
}

func (b *fontGalleryBackend) Open(spec desktop.Spec, view func(*ui.Context), closed func()) (desktop.Surface, error) {
	surface, err := b.creatorPopupBackend.Open(spec, func(c *ui.Context) {
		view(c)
		b.roles = theme.FontsOf(c)
	}, closed)
	if err == nil && spec.FitContent && spec.MaxHeight > spec.Height {
		// Tester has no native FitContent sizing negotiation. Give this menu
		// its declared bounded canvas; do not mistake the provisional 200-DIP
		// Spec.Height for a measured OS Flyout or claim platform sizing here.
		b.surfaces[len(b.surfaces)-1].tester.SetSize(max(1, spec.Width), spec.MaxHeight)
	}
	return surface, err
}

func galleryFontInk(im image.Image, rect ui.Rect) int {
	left, right := int(rect.X+rect.W+1), int(rect.X)-1
	for y := int(rect.Y); y < int(rect.Y+rect.H+1); y++ {
		for x := int(rect.X); x < int(rect.X+rect.W+1); x++ {
			r, g, b, a := im.At(x, y).RGBA()
			if a != 0 && r < 32000 && g < 32000 && b < 32000 {
				left, right = min(left, x), max(right, x)
			}
		}
	}
	return max(0, right-left+1)
}

func fontGallerySnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_FONT_SNAPSHOT_DIR")
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
		t.Fatalf("font snapshot: %v/%v", err, closeErr)
	}
}

// One actual Gallery flow spans font controls, a native draft, preset history,
// independent styled content and the platform route. It is not a font matrix.
func TestCreatorNamedFontsKeepNativeDraftHistoryAndWindowRoles(t *testing.T) {
	s := New()
	s.Page, s.Config.NativeMenus, s.Config.ReduceMotion, s.Config.Appearance = "controls", true, true, theme.LightMode
	if s.Config.BodyFont != "inter" || s.Config.HeadingFont != theme.FontInherit {
		t.Fatal("fixed Vega default fonts were not selected")
	}
	b := &fontGalleryBackend{}
	s.Popovers = desktop.NewHostWithBackend(b)
	var selection [2]int
	r := ui.NewTester(func(c *ui.Context) {
		s.View(c)
		if editor := c.Resolve(s.inputGroups.draftHandle); editor.Valid() {
			selection[0], selection[1] = editor.TextSelection()
		}
	}, 1280, 900)
	choose := func(label, value string) {
		t.Helper()
		a := advancedAccess(t, r, ui.RoleMenuButton, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		if err := r.ChooseMenuItem(value); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	// The actual Font menu exposes every fixed family under its readable name.
	a := advancedAccess(t, r, ui.RoleMenuButton, "Font")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	for _, font := range fonts.Catalog() {
		if !slices.Contains(r.Menu(), font.Name) || !slices.Contains(bodyFonts, font.ID) || !slices.Contains(headingFonts, font.ID) {
			t.Fatalf("source font unavailable in the native choices: %s", font.Name)
		}
	}
	if err := r.ChooseMenuItem("Inter"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	a = revealPreviewPart(t, r, "Forms & controls", ui.RoleTextField, "Draft text")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("alpha beta")
	r.Key(ui.Shift, ui.KeyLeft)
	r.Key(ui.Shift, ui.KeyLeft)
	before := selection
	if before[0] == before[1] {
		t.Fatal("native draft selection fixture was empty")
	}
	choose("Font", "JetBrains Mono")
	choose("Heading", "Playfair Display")
	if s.Config.BodyFont != "jetbrains-mono" || s.Config.HeadingFont != "playfair-display" || s.fontError != "" || s.inputGroups.draft != "alpha beta" || selection != before {
		t.Fatalf("font controls lost their actual model/editor: %+v selection=%v/%v", s.Config, before, selection)
	}
	registered := map[string]bool{}
	for _, font := range ui.FontFamilies() {
		registered[font.Name] = font.Registered
	}
	if !registered["Inter"] || !registered["JetBrains Mono"] || !registered["Playfair Display"] {
		t.Fatal("a selected family was merely named without native registration")
	}
	selected := s.Config
	parsed, err := ParsePreset(s.ExportPreset())
	if err != nil || parsed != selected {
		t.Fatalf("named font preset did not round trip: %+v %v", parsed, err)
	}
	// Equal named catalogue fonts persist as inherit, so later body changes flow.
	choose("Heading", "JetBrains Mono")
	if s.Config.HeadingFont != theme.FontInherit || s.Config.FontRoles().Heading != theme.FontInherit || selection != before {
		t.Fatal("equal body/heading was frozen to a concrete previous family")
	}
	if err := r.Click("Undo design change"); err != nil {
		t.Fatal(err)
	}
	if s.Config != selected || selection != before {
		t.Fatal("UI undo did not restore the exact independent font roles/editor")
	}
	if err := r.Click("Redo design change"); err != nil {
		t.Fatal(err)
	}
	if s.Config.HeadingFont != theme.FontInherit || selection != before {
		t.Fatal("UI redo did not restore inherited headings and the draft selection")
	}
	if !s.Undo() {
		t.Fatal("no font change to restore")
	}
	s.Page = "overview"
	r.Frame()
	actualTitle := advancedAccess(t, r, ui.RoleText, "Create an account")
	actualDescription := advancedAccess(t, r, ui.RoleText, "Enter your email below to create your account.")
	titleInk, bodyInk := galleryFontInk(r.Image(), actualTitle.Bounds), galleryFontInk(r.Image(), actualDescription.Bounds)
	ref := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(20).Gap(12).Background(ui.RGB(255, 255, 255)).TextColor(ui.RGB(0, 0, 0)).Children(func() {
			ui.Text(c, "Create an account").Font("Playfair Display").FontSize(18).FontWeight(600).Label("Independent heading reference")
			ui.Text(c, "Create an account").Font("JetBrains Mono").FontSize(18).FontWeight(600).Label("Independent body reference")
		})
	}, 500, 120)
	headingRef := galleryFontInk(ref.Image(), advancedAccess(t, ref, ui.RoleText, "Independent heading reference").Bounds)
	monoRef := galleryFontInk(ref.Image(), advancedAccess(t, ref, ui.RoleText, "Independent body reference").Bounds)
	if titleInk == 0 || math.Abs(float64(titleInk-headingRef)) > 2 || titleInk == monoRef {
		t.Fatalf("actual Gallery heading pixels did not use the selected font: heading=%d reference=%d mono=%d", titleInk, headingRef, monoRef)
	}
	fontGallerySnapshot(t, r, "font-gallery-mono-playfair")
	choose("Heading", "Inherit")
	title := advancedAccess(t, r, ui.RoleText, "Create an account")
	description := advancedAccess(t, r, ui.RoleText, "Enter your email below to create your account.")
	if math.Abs(float64(galleryFontInk(r.Image(), title.Bounds)-monoRef)) > 2 || galleryFontInk(r.Image(), description.Bounds) != bodyInk {
		t.Fatal("inherited title did not use body pixels or changed the independent body")
	}
	choose("Heading", "System")
	if s.Config.FontRoles().Heading != theme.FontSystem || s.Config.BodyFamily() != "JetBrains Mono" {
		t.Fatal("explicit system heading was conflated with body inheritance")
	}
	// Native categories predate the source catalogue. In particular, an
	// explicit system heading must survive a matching system body and later
	// body changes; only equal named catalogue fonts normalize to inherit.
	choose("Font", "System")
	choose("Heading", "System")
	systemPair, err := ParsePreset(s.ExportPreset())
	if err != nil || systemPair.BodyFont != "system" || systemPair.HeadingFont != "system" {
		t.Fatalf("explicit system pair was rewritten as inheritance: %+v err=%v", systemPair, err)
	}
	choose("Font", "Inter")
	systemHeading, err := ParsePreset(s.ExportPreset())
	if err != nil || systemHeading.BodyFont != "inter" || systemHeading.HeadingFont != "system" || systemHeading.FontRoles().Heading != theme.FontSystem {
		t.Fatalf("changing the body replaced an explicit system heading: %+v err=%v", systemHeading, err)
	}
	legacy := s.Config
	legacy.BodyFont, legacy.HeadingFont = "serif", "same"
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	compat, err := ParsePreset(string(raw))
	if err != nil || compat.HeadingFont != theme.FontInherit || compat.BodyFont != "serif" || compat.HeadingFamily() != compat.BodyFamily() {
		t.Fatalf("legacy Version 1 font choices were not preserved: %+v %v", compat, err)
	}
	styled := selected
	styled.NativeMenus = false
	if err := s.ApplyPreset(styled); err != nil {
		t.Fatal(err)
	}
	s.Page = "overlays"
	r.Frame()
	advancedClick(t, r, "Popover")
	if len(b.surfaces) != 1 || b.surfaces[0].spec.SystemPopover || b.roles != selected.FontRoles() {
		t.Fatalf("actual styled popup lost Gallery font roles: %d %+v", len(b.surfaces), b.roles)
	}
	child := b.surfaces[0]
	fontGallerySnapshot(t, child.tester, "font-gallery-styled-flyout")
	styled.BodyFont = "inter"
	if err := s.ApplyPreset(styled); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if len(b.surfaces) != 1 || b.roles != styled.FontRoles() || b.surfaces[0] != child {
		t.Fatal("font-only preset update replaced the existing Flyout or lost its roles")
	}
	child.tester.Key(0, ui.KeyEscape)
	r.Frame()
	r.Advance(time.Second)
	styled.NativeMenus = true
	if err := s.ApplyPreset(styled); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	advancedClick(t, r, "Popover")
	if len(b.surfaces) != 2 || !b.surfaces[1].spec.SystemPopover || b.roles != (theme.FontRoles{Heading: theme.FontInherit}) {
		t.Fatalf("Native popup did not preserve its platform font: count=%d roles=%+v", len(b.surfaces), b.roles)
	}
	b.surfaces[1].tester.Key(0, ui.KeyEscape)
	r.Frame()
	styled.NativeMenus = false
	if err := s.ApplyPreset(styled); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	trigger, found := r.AccessByRole(ui.RoleMenuButton, "Font")
	if !found {
		trigger = advancedAccess(t, r, ui.RoleButton, "Font")
	}
	r.ClickAt(trigger.Bounds.X+trigger.Bounds.W/2, trigger.Bounds.Y+trigger.Bounds.H/2)
	if len(b.surfaces) != 3 || b.surfaces[2].spec.SystemPopover {
		t.Fatal("the full font picker did not use the selected styled window route")
	}
	picker := b.surfaces[2].tester
	picker.Advance(time.Second)
	picker.Key(0, ui.KeyEnd)
	// Three retained native-category entries follow the last source font.
	for range 3 {
		picker.Key(0, ui.KeyUp)
	}
	picker.Advance(time.Second)
	lastFont, present := picker.Access("Instrument Serif")
	fontGallerySnapshot(t, picker, "font-gallery-scroll-picker")
	if !present || lastFont.Offscreen || lastFont.Bounds.Y < 0 || lastFont.Bounds.Y+lastFont.Bounds.H > float32(picker.Image().Bounds().Dy())+1 {
		t.Fatalf("the final named font was not revealed in the bounded picker: present=%v node=%+v spec=%+v image=%v", present, lastFont, b.surfaces[2].spec, picker.Image().Bounds())
	}
	picker.Key(0, ui.KeyEnter)
	r.Frame()
	if s.Config.BodyFont != "instrument-serif" || s.Config.BodyFamily() != "Instrument Serif" || s.fontError != "" || s.inputGroups.draft != "alpha beta" {
		t.Fatal("the scrolled source font choice did not register/apply or lost the draft")
	}
}
