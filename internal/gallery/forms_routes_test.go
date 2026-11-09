package gallery

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
)

func formsRoutesTester(s *State, b *creatorPopupBackend) *ui.Tester {
	s.Popovers = desktop.NewHostWithBackend(b)
	return ui.NewTester(func(c *ui.Context) {
		s.Theme.Apply(c)
		s.Popovers.BeginFrame()
		defer s.Popovers.EndFrame()
		ui.Column(c).Fill().Padding(24).Gap(20).Children(func() {
			switch s.Page {
			case "controls":
				s.otpPreview(c)
				s.choiceVariantsPreview(c)
			case "navigation":
				s.breadcrumbPreview(c)
				s.paginationPreview(c)
			default:
				ui.Text(c, "Another preview")
			}
		})
	}, 1100, 1000)
}

func TestCreatorFormCodesAndChoicesKeepControlledStateAcrossPages(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := formsRoutesTester(s, &creatorPopupBackend{})
	advancedClick(t, r, "Verification code")
	r.SetClipboard("123-456")
	r.Key(ui.Cmd, ui.KeyV)
	if s.OTP != "123456" || s.formStatus != "Verification code complete" || !r.HasText(s.formStatus) {
		t.Fatalf("numeric code demo did not filter and complete: %q status=%q", s.OTP, s.formStatus)
	}
	advancedClick(t, r, "Invite code")
	r.SetClipboard("go-123")
	r.Key(ui.Cmd, ui.KeyV)
	if s.inviteCode != "GO-123" {
		t.Fatalf("unrestricted paste-transform demo = %q", s.inviteCode)
	}
	if !advancedAccess(t, r, ui.RoleCheckBox, "All permissions").Mixed {
		t.Fatal("partial permissions did not expose native mixed state")
	}
	advancedClick(t, r, "All permissions")
	if !s.readAccess || !s.writeAccess {
		t.Fatal("select all did not change both child permissions")
	}
	advancedClick(t, r, "Edit projects")
	if s.writeAccess || !advancedAccess(t, r, ui.RoleCheckBox, "All permissions").Mixed {
		t.Fatal("child edit did not restore mixed parent state")
	}
	advancedClick(t, r, "Quiet notifications")
	advancedClick(t, r, "Italic")
	if !s.quietNotifications || !slices.Contains(s.formatting, "bold") || !slices.Contains(s.formatting, "italic") {
		t.Fatal("native switch/toggle group did not update caller-owned state")
	}
	s.Page = "overview"
	r.Frame()
	s.Page = "controls"
	r.Frame()
	if advancedAccess(t, r, ui.RoleTextField, "Invite code").Value != "GO-123" || s.OTP != "123456" || !s.quietNotifications || !slices.Contains(s.formatting, "italic") {
		t.Fatal("leaving Forms lost draft or selection state")
	}
}

func TestCreatorBreadcrumbOverflowUsesSelectedMenuFamily(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			s := New()
			s.Page, s.Config.NativeMenus = "navigation", native
			b := &creatorPopupBackend{}
			r := formsRoutesTester(s, b)
			advancedClick(t, r, "Show parent routes")
			if native {
				if len(b.surfaces) != 0 || !slices.Contains(r.Menu(), "Forms") || r.HasText("Forms") {
					t.Fatal("breadcrumb native family did not use the OS menu")
				}
				if err := r.ChooseMenuItem("Forms"); err != nil {
					t.Fatal(err)
				}
			} else {
				if len(b.surfaces) != 1 || b.surfaces[0].spec.SystemPopover || r.HasText("Forms") {
					t.Fatal("breadcrumb themed family did not use an independent Flyout")
				}
				popup := b.surfaces[0].tester
				popup.Advance(time.Second)
				advancedClick(t, popup, "Forms")
				r.Frame()
			}
			if s.breadcrumbSelection != "Selected parent: Forms" || len(r.OpenedURLs()) != 0 {
				t.Fatalf("breadcrumb result = %q urls=%v", s.breadcrumbSelection, r.OpenedURLs())
			}
			advancedClick(t, r, "Show parent routes")
			s.Page = "overview"
			r.Frame()
			for _, surface := range b.surfaces {
				if !surface.didClose {
					t.Fatal("leaving Navigation retained an overflow window")
				}
			}
		})
	}
}

func TestCreatorPaginationRepeatsKeysAndHandlesDistantBoundary(t *testing.T) {
	s := New()
	s.Page = "navigation"
	r := formsRoutesTester(s, &creatorPopupBackend{})
	advancedClick(t, r, "Go to next page")
	r.Key(0, ui.KeyEnter)
	r.Key(0, ui.KeyEnter)
	if s.PageNumber != 4 || !r.Focused("Go to next page") || !r.HasText("Page 4 of 1,000,000") {
		t.Fatal("pagination demo lost focus or did not report repeated key changes")
	}
	s.PageNumber = 1_000_000
	r.Frame()
	if !advancedAccess(t, r, ui.RoleButton, "Go to next page").Disabled || advancedAccess(t, r, ui.RoleButton, "Page 1000000").Current != ui.CurrentPage {
		t.Fatal("distant last page lost its boundary/current semantics")
	}
	advancedClick(t, r, "Page 1")
	if s.PageNumber != 1 || !r.HasText("Page 1 of 1,000,000") {
		t.Fatal("bounded window failed to jump back to the first page")
	}
}

func TestCreatorBioNativeEditorAndResizeSurvivePageChanges(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := ui.NewTester(s.View, 1280, 900)
	advancedClick(t, r, "Bio")
	r.Type("Native bio draft")
	advancedClick(t, r, "Resize bio")
	r.Key(0, ui.KeyDown)
	field := advancedAccess(t, r, ui.RoleTextField, "Bio")
	if s.Bio != "Native bio draft" || s.bioResize.Height <= 64 || field.Bounds.H != s.bioResize.Height {
		t.Fatalf("bio editor/resize did not compose: draft=%q height=%g field=%+v", s.Bio, s.bioResize.Height, field)
	}
	height := s.bioResize.Height
	s.Page = "content"
	r.Frame()
	s.Page = "controls"
	r.Frame()
	field = advancedAccess(t, r, ui.RoleTextField, "Bio")
	if field.Value != "Native bio draft" || field.Bounds.H != height || s.bioResize.Height != height {
		t.Fatal("leaving Forms reset the caller-owned bio draft or height")
	}
}
