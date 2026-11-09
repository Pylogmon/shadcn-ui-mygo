package gallery

import (
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func inputGroupAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok || a.Offscreen {
		t.Fatalf("missing native control %v %q", role, name)
	}
	return a
}
func inputGroupClick(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	a, ok := r.Access(name)
	if !ok || a.Offscreen {
		t.Fatalf("missing action %q", name)
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
}
func inputGroupType(t *testing.T, r *ui.Tester, name, text string) {
	t.Helper()
	a := inputGroupAccess(t, r, ui.RoleTextField, name)
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type(text)
	r.Frame()
}
func inputGroupSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_INPUT_GROUP_SNAPSHOT_DIR")
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
		t.Fatalf("snapshot %v/%v", err, closeErr)
	}
}
func inputGroupChoose(t *testing.T, r *ui.Tester, b *creatorPopupBackend, trigger, choice string) {
	t.Helper()
	before := len(b.surfaces)
	inputGroupClick(t, r, trigger)
	if len(b.surfaces) != before+1 {
		t.Fatalf("%s did not open an independent Flyout", trigger)
	}
	child := b.surfaces[len(b.surfaces)-1]
	child.tester.Advance(time.Second)
	if _, ok := child.tester.AccessByRole(ui.RoleMenuItem, choice); !ok {
		t.Fatal("missing child menu item", choice)
	}
	if err := child.tester.Click(choice); err != nil {
		t.Fatal(err)
	}
	r.Frame()
}

func TestInputGroupPreviewNativeWorkflow(t *testing.T) {
	p := &inputGroupPreviewState{}
	b := &creatorPopupBackend{}
	h := desktop.NewHostWithBackend(b)
	dark, native, show := false, false, true
	var selection [2]int
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Column().Fill().Padding(24).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			if show {
				p.view(c, h, native)
				if e := c.Resolve(p.draftHandle); e.Valid() {
					selection[0], selection[1] = e.TextSelection()
				}
			} else {
				ui.Text(c, "Another Gallery page")
			}
		})
	}, 1040, 960, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	// The real asynchronous clipboard acknowledgment, including refusal, drives
	// status. No completion callback is called directly by this test.
	r.SetClipboard("kept clipboard")
	r.FailClipboard(errors.New("clipboard temporarily locked"))
	inputGroupClick(t, r, "Copy link")
	if r.Clipboard() != "kept clipboard" || !strings.Contains(p.feedback, "Could not copy link") || p.copyKind != "" {
		t.Fatal("failed clipboard write claimed success")
	}
	r.FailClipboard(nil)
	inputGroupClick(t, r, "Copy link")
	if r.Clipboard() != p.link || p.feedback != "Copied link." || p.copyKind != "link" {
		t.Fatal("clipboard acknowledgment did not update copy state")
	}
	inputGroupClick(t, r, "Favorite site toggle")
	if !p.favorite {
		t.Fatal("favorite button did not change app state")
	}
	inputGroupType(t, r, "Search query", "native")
	inputGroupChoose(t, r, b, "Search target", "Blog Posts")
	captionCount := 0
	for _, text := range r.Texts() {
		if text == "Blog Posts" {
			captionCount++
		}
	}
	if captionCount != 1 {
		t.Fatalf("search trigger repeated its visible caption %d times", captionCount)
	}
	inputGroupClick(t, r, "Search")
	if p.target != "Blog Posts" || !slices.Equal(p.results, []string{"Building native forms"}) {
		t.Fatalf("target/search action not meaningful: %q %v", p.target, p.results)
	}
	inputGroupType(t, r, "Draft file name", "drafts/example.js")
	inputGroupChoose(t, r, b, "Draft file actions", "Settings")
	inputGroupClick(t, r, "Monospace preview")
	if !p.settingsOpen || p.monospace {
		t.Fatal("file settings did not control preview")
	}
	inputGroupChoose(t, r, b, "Draft file actions", "Copy file name")
	if r.Clipboard() != "drafts/example.js" || p.feedback != "Copied file name." {
		t.Fatal("file action did not copy current filename")
	}
	inputGroupChoose(t, r, b, "Draft file actions", "Show file details")
	if !p.detailsOpen {
		t.Fatal("file detail action is inert")
	}
	inputGroupType(t, r, "Draft text", "alpha\nbeta 😀")
	r.Key(ui.Shift, ui.KeyLeft)
	before := selection
	if before[0] == before[1] {
		t.Fatal("native textarea selection guard was empty")
	}
	inputGroupClick(t, r, "Preview draft")
	if !p.previewed || p.preview != p.draft || selection != before || !r.HasText("Draft preview") || !r.HasText("alpha\nbeta 😀") {
		t.Fatalf("preview replaced editor or lost real selection: %v→%v", before, selection)
	}
	inputGroupClick(t, r, "Copy draft")
	if r.Clipboard() != p.draft || p.feedback != "Copied draft." || selection != before {
		t.Fatal("block addon copy lost editor text/selection or clipboard feedback")
	}
	inputGroupSnapshot(t, r, "input-groups-light")
	inputGroupClick(t, r, "Reset draft")
	if p.draft != inputGroupInitialDraft || p.previewed || inputGroupAccess(t, r, ui.RoleTextField, "Draft text").Value != inputGroupInitialDraft {
		t.Fatal("reset did not update actual native textarea")
	}
	inputGroupClick(t, r, "Site information")
	info := b.surfaces[len(b.surfaces)-1]
	info.tester.Advance(time.Second)
	if !info.tester.HasText("Example address") || !info.tester.HasText("https://example.com") {
		t.Fatal("site info did not use separate current child content")
	}
	// A single switch to the OS route also verifies mode change closes the old
	// child, without rerunning the complete workflow for each implementation.
	native = true
	r.Frame()
	if !info.didClose || p.infoOpen {
		t.Fatal("native route switch leaked old Flyout")
	}
	count := len(b.surfaces)
	inputGroupClick(t, r, "Search target")
	if len(b.surfaces) != count || !slices.Contains(r.Menu(), "Changelog") {
		t.Fatal("Native route did not register OS menu")
	}
	if err := r.ChooseMenuItem("Changelog"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.target != "Changelog" {
		t.Fatal("OS menu choice did not reach model")
	}
	inputGroupClick(t, r, "Preview draft")
	dark = true
	r.Frame()
	inputGroupSnapshot(t, r, "input-groups-dark")
	show = false
	r.Frame()
	show = true
	r.Frame()
	if p.draft != inputGroupInitialDraft || p.fileName != "drafts/example.js" || !p.favorite || p.target != "Changelog" {
		t.Fatal("page change discarded app-owned values")
	}
}

func TestInputGroupPreviewMissingHostReportsPopupError(t *testing.T) {
	p := &inputGroupPreviewState{}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(24).Children(func() { p.view(c, nil, false) })
	}, 1040, 960, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if p.popupError != "" || p.infoOpen {
		t.Fatal("missing Host reported an initial popup error before activation")
	}
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleAlert {
			t.Fatal("initial inactive popup exposed an alert")
		}
	}
	inputGroupClick(t, r, "Site information")
	if p.infoOpen || !strings.Contains(p.popupError, "site information") {
		t.Fatal("missing Host info activation was silent or left popup open")
	}
	inputGroupClick(t, r, "Search target")
	if p.targetOpen || p.popupError == "" {
		t.Fatal("missing Host silently opened a parent-bounded fallback")
	}
	if a, ok := r.AccessByRole(ui.RoleAlert, p.popupError); !ok || a.Name == "" {
		t.Fatal("missing Host error not exposed")
	}
}
