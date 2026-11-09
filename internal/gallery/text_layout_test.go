package gallery

import (
	"bytes"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func textLayoutTester(s *State, module string) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		s.Theme.Apply(c)
		if module != "loading" {
			c.DisableAnimations()
		}
		ui.Column(c).Fill().Padding(20).Gap(20).Children(func() {
			if s.Page != "text-layout" {
				ui.Text(c, "Another page")
				return
			}
			switch module {
			case "code":
				s.codePreview(c)
			case "loading":
				s.loadingPreview(c)
			case "scroll":
				s.scrollPreview(c)
			default:
				s.typographyPreview(c)
			}
		})
	}, 1100, 800)
}

func TestCreatorTextLayoutPageIsReachableFromFuzzyCommand(t *testing.T) {
	s := New()
	r := ui.NewTester(s.View, 1280, 900)
	advancedClick(t, r, "Search components")
	r.Advance(200 * time.Millisecond)
	r.Type("txtlyt")
	r.Key(0, ui.KeyEnter)
	r.Advance(200 * time.Millisecond)
	if s.Page != "text-layout" || s.Command || !r.HasText("Words, code and space.") {
		t.Fatalf("new preview was not registered in page commands: page=%q open=%v", s.Page, s.Command)
	}
}

func TestCreatorTextLayoutComponentsAreSearchableByName(t *testing.T) {
	for _, query := range []string{"Typography", "Kbd", "CodeBlock", "ScrollArea", "AspectRatio"} {
		t.Run(query, func(t *testing.T) {
			s := New()
			r := ui.NewTester(s.View, 1280, 900)
			advancedClick(t, r, "Search components")
			r.Advance(200 * time.Millisecond)
			r.Type(query)
			r.Key(0, ui.KeyEnter)
			r.Advance(200 * time.Millisecond)
			if s.Page != "text-layout" || s.Command {
				t.Fatalf("component name %q did not reach its actual preview: page=%q open=%v", query, s.Page, s.Command)
			}
		})
	}
}

func TestCreatorCodeOptionsKeepNativeBufferAndCopyOnlySource(t *testing.T) {
	s := New()
	s.Page = "text-layout"
	r := textLayoutTester(s, "code")
	field := advancedAccess(t, r, ui.RoleTextField, "Source preview")
	if !field.ReadOnly || !field.Multiline {
		t.Fatal("code preview lost its actual read-only native editor")
	}
	buffer := s.codeExample.Editor.Buffer
	advancedClick(t, r, "Plain text")
	advancedClick(t, r, "Wrap code")
	advancedClick(t, r, "Compact code")
	advancedClick(t, r, "Highlight changes")
	if s.codeLanguage != "text" || !s.codeWrap || !s.codeCompact || !s.codeDiff || s.codeExample.Editor.Buffer != buffer {
		t.Fatal("code options replaced the native buffer or did not change app-owned props")
	}
	if buffer.Language() != nil || !s.codeExample.Editor.SoftWrap || s.codeExample.Editor.Font.Size != 12 {
		t.Fatal("code options did not reach the actual native editor")
	}
	advancedClick(t, r, "Copy example")
	r.Frame()
	if r.Clipboard() != previewCode || s.codeCopied != "Example copied" {
		t.Fatalf("copy included line numbers/markers or did not confirm completion: copied=%q status=%q", r.Clipboard(), s.codeCopied)
	}
	s.Page = "content"
	r.Frame()
	s.Page = "text-layout"
	r.Frame()
	if s.codeExample.Editor.Buffer != buffer || s.codeLanguage != "text" || !s.codeWrap || !s.codeDiff {
		t.Fatal("owner removal lost code buffer or options")
	}
}

func TestCreatorArticleDetailsAndShortcutUseRealActions(t *testing.T) {
	s := New()
	s.Page = "text-layout"
	r := textLayoutTester(s, "article")
	if _, exists := r.AccessByRole(ui.RoleTable, "Document structure"); exists {
		t.Fatal("closed article details exposed a table")
	}
	advancedClick(t, r, "Show article details")
	if !s.articleDetails {
		t.Fatal("article details button did not update app state")
	}
	advancedAccess(t, r, ui.RoleArticle, "Preview article")
	advancedAccess(t, r, ui.RoleTable, "Document structure")
	advancedAccess(t, r, ui.RoleList, "Writing principles")
	advancedClick(t, r, "Typography guide")
	if urls := r.OpenedURLs(); len(urls) != 1 || urls[0] != "https://ui.shadcn.com/docs/components/typography" {
		t.Fatalf("native inline link did not open its actual URL: %v", urls)
	}
	advancedClick(t, r, "Search with shortcut")
	if !s.Command {
		t.Fatal("keycap composition blocked its real native search button")
	}
}

func TestCreatorLoadingToggleRetainsDraftAndBecomesIdle(t *testing.T) {
	s := New()
	s.Page = "text-layout"
	r := textLayoutTester(s, "loading")
	initial := append([]byte(nil), r.Image().Pix...)
	r.Advance(250 * time.Millisecond)
	if bytes.Equal(initial, r.Image().Pix) || !r.FramePending() {
		t.Fatal("loading preview did not start with actual animated pixels")
	}
	advancedClick(t, r, "Workspace lookup")
	r.Type("Native draft")
	advancedClick(t, r, "Pause loading animations")
	if !s.loadingStatic || s.layoutDraft != "Native draft" {
		t.Fatal("loading controls lost draft or did not pause")
	}
	advancedAccess(t, r, ui.RoleProgress, "Preview loading indicator")
	if advancedAccess(t, r, ui.RoleButton, "Sync workspace").Busy {
		t.Fatal("paused preview left its action busy")
	}
	r.Advance(250 * time.Millisecond)
	image := append([]byte(nil), r.Image().Pix...)
	r.Advance(2 * time.Second)
	if !bytes.Equal(image, r.Image().Pix) || r.FramePending() {
		t.Fatal("paused loading preview still changed or scheduled work")
	}
	s.Page = "content"
	r.Frame()
	s.Page = "text-layout"
	r.Frame()
	if advancedAccess(t, r, ui.RoleTextField, "Workspace lookup").Value != "Native draft" || !s.loadingStatic {
		t.Fatal("leaving the preview lost the caller-owned loading state")
	}
}

func TestCreatorNotesScrollForRealAndRetainPlaceAcrossStylesAndPages(t *testing.T) {
	s := New()
	s.Page = "text-layout"
	r := textLayoutTester(s, "scroll")
	first, ok := r.Find("Note 01 · Make room for your next idea.")
	if !ok || s.articleScroll.MaxY <= 0 {
		t.Fatal("scroll preview did not have actual overflowing native content")
	}
	advancedClick(t, r, "Scroll to end")
	lastPlace := s.articleScroll.Y
	after, _ := r.Find("Note 01 · Make room for your next idea.")
	if lastPlace <= 0 || lastPlace != s.articleScroll.MaxY || after.Y >= first.Y {
		t.Fatal("end action changed an indicator without moving native content")
	}
	advancedClick(t, r, "Use MyGo scrollbars")
	if !s.nativeScrollbars || s.articleScroll.Y != lastPlace {
		t.Fatal("scrollbar style change reset the accepted native offset")
	}
	s.Page = "content"
	r.Frame()
	s.Page = "text-layout"
	r.Frame()
	if s.articleScroll.Y != lastPlace {
		t.Fatal("leaving the preview lost its native scroll position")
	}
	advancedAccess(t, r, ui.RoleImage, "Portrait illustration")
	advancedClick(t, r, "Scroll to start")
	if s.articleScroll.Y != 0 {
		t.Fatal("start action did not restore the native origin")
	}
}
