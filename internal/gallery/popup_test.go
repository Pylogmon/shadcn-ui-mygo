package gallery

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
)

type creatorPopupBackend struct {
	surfaces    []*creatorPopupSurface
	invalidates int
}
type creatorPopupSurface struct {
	spec     desktop.Spec
	view     func(*ui.Context)
	tester   *ui.Tester
	closed   func()
	didClose bool
}

func (b *creatorPopupBackend) Open(spec desktop.Spec, view func(*ui.Context), closed func()) (desktop.Surface, error) {
	s := &creatorPopupSurface{spec: spec, view: view, closed: closed}
	b.surfaces = append(b.surfaces, s)
	s.tester = ui.NewTester(view, max(1, spec.Width), max(1, spec.Height))
	return s, nil
}
func (b *creatorPopupBackend) InvalidateParent()   { b.invalidates++ }
func (s *creatorPopupSurface) SetAnchor(r ui.Rect) { s.spec.Anchor = r }
func (s *creatorPopupSurface) Invalidate() {
	if s.tester != nil {
		s.tester.Frame()
	}
}
func (s *creatorPopupSurface) Close() {
	if !s.didClose {
		s.didClose = true
		s.closed()
	}
}

func TestCreatorSecondaryDateAndHoverPopupsUseDesktopHost(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			s := New()
			s.Config.NativeMenus = native
			s.Page = "display"
			s.Date = time.Date(2026, time.November, 9, 0, 0, 0, 0, time.UTC)
			b := &creatorPopupBackend{}
			s.Popovers = desktop.NewHostWithBackend(b)
			r := ui.NewTester(s.View, 1280, 900)
			if err := r.Click("Project date"); err != nil {
				t.Fatal(err)
			}
			if len(b.surfaces) != 1 {
				t.Fatalf("date surfaces=%d", len(b.surfaces))
			}
			popup := b.surfaces[0]
			if popup.spec.SystemPopover != native {
				t.Fatalf("date pair wrong platform container %+v", popup.spec)
			}
			if r.HasText("November 2026") && r.HasText("Apply") {
				t.Fatal("date popup content leaked into parent tree")
			}
			popup.tester.Key(0, ui.KeyRight)
			popup.tester.Key(0, ui.KeyEnter)
			r.Frame()
			if s.Date.Day() != 10 {
				t.Fatalf("child choice not reflected in Gallery: %v", s.Date)
			}
			if s.Popovers.IsOpen("preview-project-date") {
				t.Fatal("date popup did not close")
			}
			s.Page = "overlays"
			r.Frame()
			trigger, ok := r.Find("Hover for a tip")
			if !ok {
				t.Fatal("tooltip trigger missing")
			}
			r.Move(trigger.X+trigger.W/2, trigger.Y+trigger.H/2)
			if len(b.surfaces) != 2 {
				t.Fatalf("tooltip surfaces=%d", len(b.surfaces))
			}
			tip := b.surfaces[1]
			if !tip.spec.IgnoreMouseEvents || tip.spec.Focusable {
				t.Fatalf("tooltip catches pointer/key input %+v", tip.spec)
			}
			if !tip.tester.HasText("Small details make a difference.") {
				t.Fatal("tooltip did not render in child window")
			}
			tip.tester.SetSize(tip.spec.MaxWidth, 100)
			// Find returns visual bounds, including the flyout's entry scale.
			tip.tester.Advance(200 * time.Millisecond)
			if !tip.tester.HasText("Search components") {
				t.Fatal("rich tooltip omitted the actual application action")
			}
			if !native {
				if key, ok := tip.tester.Find("K"); !ok || key.H != 20 {
					t.Fatal("styled tooltip did not build a native Kbd")
				}
			}
			r.Key(ui.Cmd, ui.KeyK)
			r.Frame()
			if !s.Command {
				t.Fatal("tooltip advertised a shortcut without an application action")
			}
			if _, ok := r.AccessByRole(ui.RoleComboBox, "Search commands"); !ok {
				t.Fatal("advertised search shortcut did not open the real command input")
			}
			s.Page = "overview"
			r.Frame()
			if s.Popovers.IsPresent("preview-tooltip") {
				t.Fatal("hidden preview kept detached tooltip")
			}
		})
	}
}

func TestCreatorHoverCardRefreshesSharedValueAcrossStylesAndClosesOnNavigation(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			s := New()
			s.Page, s.Config.NativeMenus = "overlays", native
			b := &creatorPopupBackend{}
			s.Popovers = desktop.NewHostWithBackend(b)
			parent := ui.NewTester(s.View, 1280, 900)
			anchor, ok := parent.Find("Preview author")
			if !ok {
				t.Fatal("author trigger missing")
			}
			parent.Move(anchor.X+anchor.W/2, anchor.Y+anchor.H/2)
			parent.Advance(701 * time.Millisecond)
			if len(b.surfaces) != 1 {
				t.Fatalf("hover surfaces=%d", len(b.surfaces))
			}
			popup := b.surfaces[0]
			for _, mode := range []bool{native, !native, native} {
				p := s.Config
				p.NativeMenus = mode
				if err := s.ApplyPreset(p); err != nil {
					t.Fatal(err)
				}
				parent.Frame()
				popup.tester.Advance(time.Second)
				if len(b.surfaces) != 1 || !s.Popovers.IsOpen("preview-author") {
					t.Fatal("style change replaced or closed the inactive hover window")
				}
				if popup.tester.HasText("SD") == mode {
					t.Fatalf("native=%v did not refresh the matching content", mode)
				}
				s.Marketing = !s.Marketing
				parent.Frame()
				checkbox, ok := popup.tester.AccessByRole(ui.RoleCheckBox, "Follow project updates")
				if !ok || checkbox.Checked != s.Marketing {
					t.Fatalf("native=%v child retained old shared value: %+v", mode, checkbox)
				}
				if parent.HasText("Follow project updates") {
					t.Fatal("hover content leaked into the parent tree")
				}
			}
			s.Page = "overview"
			parent.Frame()
			if s.Popovers.IsPresent("preview-author") || !popup.didClose {
				t.Fatal("leaving the page retained the detached hover window")
			}
		})
	}
}

func TestCreatorPopoverRefreshesSharedValueAcrossStylesAndClosesOnNavigation(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			s := New()
			s.Page, s.Config.NativeMenus = "overlays", native
			b := &creatorPopupBackend{}
			s.Popovers = desktop.NewHostWithBackend(b)
			parent := ui.NewTester(s.View, 1280, 900)
			if err := parent.Click("Popover"); err != nil {
				t.Fatal(err)
			}
			for i, mode := range []bool{native, !native, native} {
				p := s.Config
				p.NativeMenus = mode
				if err := s.ApplyPreset(p); err != nil {
					t.Fatal(err)
				}
				parent.Frame()
				popup := b.surfaces[len(b.surfaces)-1]
				popup.tester.Advance(time.Second)
				if popup.spec.SystemPopover != mode || !s.Popovers.IsOpen("preview-popover") {
					t.Fatalf("native=%v did not use the matching open container", mode)
				}
				s.Name = fmt.Sprintf("External name %d", i)
				parent.Frame()
				input, ok := popup.tester.AccessByRole(ui.RoleTextField, "Layer name")
				if !ok || input.Value != s.Name {
					t.Fatalf("native=%v child retained old name: %+v", mode, input)
				}
			}
			s.Page = "overview"
			parent.Frame()
			if s.Popover || s.Popovers.IsPresent("preview-popover") {
				t.Fatal("leaving the page retained the controlled popover")
			}
			for _, popup := range b.surfaces {
				if !popup.didClose {
					t.Fatal("style change or navigation left a detached surface")
				}
			}
		})
	}
}

func TestCreatorInspectorUsesIndependentShadcnAndSystemSurfaces(t *testing.T) {
	for _, native := range []bool{false, true} {
		s := New()
		s.Config.NativeMenus = native
		b := &creatorPopupBackend{}
		s.Popovers = desktop.NewHostWithBackend(b)
		r := ui.NewTester(s.View, 1280, 900)
		if err := r.Click("Inspect theme"); err != nil {
			t.Fatal(err)
		}
		if len(b.surfaces) != 1 {
			t.Fatalf("native=%v surfaces=%d", native, len(b.surfaces))
		}
		popup := b.surfaces[0]
		if popup.spec.SystemPopover != native {
			t.Fatalf("wrong presentation native=%v spec=%+v", native, popup.spec)
		}
		if !popup.tester.HasText("Theme tokens") || r.HasText("Theme tokens") {
			t.Fatal("inspector did not build in an independent child context")
		}
		if popup.spec.Width < 300 {
			t.Fatal("independent popup was constrained to customizer card width")
		}
		p := s.Config
		p.Accent = "blue"
		if err := s.ApplyPreset(p); err != nil {
			t.Fatal(err)
		}
		r.Frame()
		color := s.Config.Tokens(false).Primary
		label := fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B)
		if !popup.tester.HasText(label) {
			t.Fatalf("open inspector did not refresh after design change: %v", popup.tester.Texts())
		}
		if err := popup.tester.Click("Copy preset"); err != nil {
			t.Fatal(err)
		}
		if p, err := ParsePreset(popup.tester.Clipboard()); err != nil || p != s.Config {
			t.Fatalf("child clipboard output=%+v err=%v", p, err)
		}
		s.Popovers.CloseAll()
		r.Frame()
		if s.tokenOpen || !popup.didClose {
			t.Fatal("closing popup did not synchronize controlled state")
		}
	}
}

func TestCreatorMissingPopupHostReportsErrorWithoutFakeWindowContent(t *testing.T) {
	s := New()
	r := ui.NewTester(s.View, 1280, 900)
	if err := r.Click("Inspect theme"); err != nil {
		t.Fatal(err)
	}
	if !r.HasText("Could not open popup") || r.HasText("Theme tokens") {
		t.Fatalf("missing host was hidden or rendered parent-bounded content: %v", r.Texts())
	}
}

func TestCreatorFlyoutThemeChoiceChangesLivePreviewAndDismisses(t *testing.T) {
	s := New()
	b := &creatorPopupBackend{}
	s.Popovers = desktop.NewHostWithBackend(b)
	r := ui.NewTester(s.View, 1280, 900)
	field, ok := r.AccessByRole(ui.RoleMenuButton, "Theme")
	if !ok {
		field, ok = r.AccessByRole(ui.RoleButton, "Theme")
	}
	if !ok {
		t.Fatal("theme selector missing")
	}
	r.ClickAt(field.Bounds.X+field.Bounds.W/2, field.Bounds.Y+field.Bounds.H/2)
	if len(b.surfaces) != 1 || b.surfaces[0].spec.SystemPopover {
		t.Fatal("theme choice did not use a shadcn Flyout")
	}
	child := b.surfaces[0].tester
	if r.HasText("Blue") {
		t.Fatal("Flyout choices painted inside parent window")
	}
	if err := child.Click("Blue"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	r.Advance(time.Second)
	if s.Config.Accent != "blue" || !s.CanUndo() || !b.surfaces[0].didClose {
		t.Fatalf("choice did not commit/close: %+v", s.Config)
	}
	button, ok := r.AccessByRole(ui.RoleButton, "Create account")
	if !ok {
		t.Fatal("live preview action missing")
	}
	color := s.Config.Tokens(false).Primary
	pixel := r.Image().RGBAAt(int(button.Bounds.X+5), int(button.Bounds.Y+button.Bounds.H/2))
	if pixel.R != color.R || pixel.G != color.G || pixel.B != color.B {
		t.Fatalf("preview pixels did not update to selected theme: %v want %+v", pixel, color)
	}
}

func TestCreatorMenuStyleSwitchesBetweenFlyoutAndSystemAdapters(t *testing.T) {
	s := New()
	b := &creatorPopupBackend{}
	s.Popovers = desktop.NewHostWithBackend(b)
	r := ui.NewTester(s.View, 1280, 900)
	field, ok := r.AccessByRole(ui.RoleMenuButton, "Menu style")
	if !ok {
		field, ok = r.AccessByRole(ui.RoleButton, "Menu style")
	}
	if !ok {
		t.Fatal("menu style selector missing")
	}
	r.ClickAt(field.Bounds.X+field.Bounds.W/2, field.Bounds.Y+field.Bounds.H/2)
	if len(b.surfaces) != 1 {
		t.Fatalf("style list did not open Flyout: %d", len(b.surfaces))
	}
	if err := b.surfaces[0].tester.Click("System native"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !s.Config.NativeMenus {
		t.Fatal("style choice did not change popup adapter")
	}
	field, ok = r.AccessByRole(ui.RoleMenuButton, "Theme")
	if !ok {
		t.Fatal("native Theme selector missing")
	}
	r.ClickAt(field.Bounds.X+field.Bounds.W/2, field.Bounds.Y+field.Bounds.H/2)
	if len(r.Menu()) == 0 || len(b.surfaces) != 1 {
		t.Fatal("system choice did not use OS menu route")
	}
	if err := r.ChooseMenuItem("Violet"); err != nil {
		t.Fatal(err)
	}
	if s.Config.Accent != "violet" {
		t.Fatal("native adapter did not write design state")
	}
}

func TestCreatorPreviewMenuUsesIndependentFlyoutAndQueuesParentAction(t *testing.T) {
	s := New()
	s.Page = "overlays"
	b := &creatorPopupBackend{}
	s.Popovers = desktop.NewHostWithBackend(b)
	r := ui.NewTester(s.View, 1280, 900)
	if err := r.Click("Open menu"); err != nil {
		t.Fatal(err)
	}
	if len(b.surfaces) != 1 || r.HasText("Log out") {
		t.Fatal("preview menu did not leave parent window")
	}
	if err := b.surfaces[0].tester.Click("Profile"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !r.HasText("Menu action") || !b.surfaces[0].didClose {
		t.Fatal("Flyout action did not close and notify parent preview")
	}
}
