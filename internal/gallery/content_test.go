package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestCreatorContentActionsAndInputsKeepSeparateState(t *testing.T) {
	s := New()
	s.Page = "content"
	r := ui.NewTester(func(c *ui.Context) {
		s.Theme.Apply(c)
		ui.Column(c).Fill().Padding(24).Gap(20).Children(func() {
			if s.Page == "content" {
				s.contents(c)
			} else {
				ui.Text(c, "Another preview")
			}
		})
	}, 1100, 1000)
	advancedClick(t, r, "Pin report")
	if !s.contentPinned || s.contentAction != "" {
		t.Fatal("item action activated its independent primary region")
	}
	advancedClick(t, r, "Open quarterly report")
	if s.contentAction != "Opened quarterly report" || !s.contentPinned {
		t.Fatal("primary action lost its notification or sibling state")
	}
	advancedClick(t, r, "Undo")
	if s.contentAction != "Changes restored" {
		t.Fatal("alert action did not deliver its event")
	}
	advancedClick(t, r, "Dismiss update")
	if !s.contentDismissed || r.HasText("A new update is available") || !r.HasText("Restore notice") {
		t.Fatal("dismissed notice still participated in the layout")
	}
	advancedClick(t, r, "New project name")
	r.Type("Native workspace")
	advancedClick(t, r, "Create project")
	if s.contentAction != "Created project: Native workspace" {
		t.Fatal("empty-state editor and action did not compose")
	}
	advancedClick(t, r, "Remove filter")
	if s.contentAction != "Filter removed" {
		t.Fatal("button badge did not publish its native activation")
	}
	advancedClick(t, r, "Badge documentation")
	if urls := r.OpenedURLs(); len(urls) != 1 || urls[0] != "https://ui.shadcn.com/docs/components/badge" {
		t.Fatalf("link badge did not use native URL action: %v", urls)
	}
	s.Page = "overview"
	r.Frame()
	s.Page = "content"
	r.Frame()
	input, ok := r.AccessByRole(ui.RoleTextField, "New project name")
	if !ok || input.Value != "Native workspace" || !s.contentPinned || !s.contentDismissed {
		t.Fatal("preview page changes lost caller-owned draft or actions")
	}
}

func TestCreatorNewContentPageIsReachableByFuzzyCommand(t *testing.T) {
	s := New()
	r := ui.NewTester(s.View, 1280, 900)
	advancedClick(t, r, "Search components")
	r.Advance(200 * time.Millisecond)
	r.Type("cntfdbck")
	r.Key(0, ui.KeyEnter)
	r.Advance(200 * time.Millisecond)
	if s.Page != "content" || s.Command || !r.HasText("Content, with context.") {
		t.Fatalf("page registry/search routing diverged: page=%q open=%v", s.Page, s.Command)
	}
	if _, ok := r.AccessByRole(ui.RoleList, "Workspace files"); !ok {
		t.Fatal("new content preview did not expose its native item group")
	}
}
