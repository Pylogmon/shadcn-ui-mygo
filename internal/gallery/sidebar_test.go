package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
)

func TestCreatorSidebarResponsiveSelectionCollapseAndDraft(t *testing.T) {
	s := New()
	s.Popovers = desktop.NewHostWithBackend(&creatorPopupBackend{})
	r := ui.NewTester(func(c *ui.Context) {
		s.Theme.Apply(c)
		s.Popovers.BeginFrame()
		defer s.Popovers.EndFrame()
		ui.Column(c).Fill().Padding(24).Children(func() { s.sidebarPreview(c) })
	}, 1000, 500)
	advancedClick(t, r, "Inbox")
	if s.sidebarRoute != "inbox" {
		t.Fatal("sidebar did not return its selected route")
	}
	advancedClick(t, r, "Workspace owner")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Native team")
	advancedClick(t, r, "Toggle sidebar")
	r.Advance(250 * time.Millisecond)
	if !s.Sidebar.Collapsed || s.Name != "Native team" {
		t.Fatal("collapsing sidebar lost draft or collapse state")
	}
	r.SetSize(640, 600)
	r.Frame()
	if !s.Sidebar.IsMobile() {
		t.Fatal("embedded sidebar did not react to its narrower container")
	}
	advancedClick(t, r, "Toggle sidebar")
	r.Advance(250 * time.Millisecond)
	if !s.Sidebar.MobileOpen || !s.Sidebar.Collapsed {
		t.Fatal("mobile visibility overwrote desktop collapse state")
	}
	advancedClick(t, r, "Settings")
	r.Advance(300 * time.Millisecond)
	if s.Sidebar.MobileOpen || s.sidebarRoute != "settings" || s.Name != "Native team" {
		t.Fatal("mobile choice did not close/commit while retaining editor state")
	}
	if !r.Focused("Toggle sidebar") {
		t.Fatal("mobile navigation did not return focus to its trigger")
	}
}
