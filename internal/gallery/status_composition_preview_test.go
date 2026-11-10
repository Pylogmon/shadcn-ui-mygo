package gallery

import (
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func statusReveal(t *testing.T, r *ui.Tester, label string) ui.AccessInfo {
	t.Helper()
	for range 8 {
		a, found := r.Access(label)
		viewport, hasViewport := r.AccessByRole(ui.RoleRegion, "Status workspace")
		if !found || !hasViewport {
			t.Fatalf("missing status control %q", label)
		}
		b, v := a.Bounds, viewport.Bounds
		if b.W <= 0 || b.H <= 0 || b.X < v.X-1 || b.X+b.W > v.X+v.W+1 {
			t.Fatalf("unbounded status control %q: %v in %v", label, b, v)
		}
		if b.Y >= v.Y && b.Y+b.H <= v.Y+v.H {
			return a
		}
		r.Scroll(v.X+v.W-8, v.Y+v.H/2, 0, b.Y+b.H/2-v.Y-v.H/2)
	}
	t.Fatalf("could not reveal status control %q", label)
	return ui.AccessInfo{}
}

func statusClick(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	a := statusReveal(t, r, label)
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
}

func statusSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_STATUS_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("status snapshot: %v/%v", err, closeErr)
	}
}

// One application workflow exercises the fixed source configurations. Its
// public native Tester/Backend are not a real OS window or screen reader.
func TestStatusCompositionsSourceActionsAndSearchWorkflow(t *testing.T) {
	p := &statusCompositionPreviewState{}
	backend := &creatorPopupBackend{}
	host := desktop.NewHostWithBackend(backend)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			host.BeginFrame()
			defer host.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave status examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show status examples").Clicked() {
					visible = true
				}
			})
			ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Role(ui.RoleRegion).Label("Status workspace").Children(func() {
				ui.Column(c).FillWidth().MinWidth(0).Gap(20).Children(func() {
					if visible {
						p.view(c, host, native)
					} else {
						ui.Text(c, "Another application page")
					}
				})
			})
		})
	}, 620, 1120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	choose := func(selector, label string) {
		t.Helper()
		statusClick(t, r, selector)
		if err := r.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	if p.popupError != "" || p.alertRecipe != "basic" || p.emptyRecipe != "projects" || !r.HasText("Account updated successfully") {
		t.Fatal("initial source examples or quiet unopened selectors are missing")
	}
	// Every Alert recipe is built through an actual native menu choice, with
	// its native alert remaining bounded even for the longest prose.
	for _, recipe := range statusAlertRecipes {
		if recipe.Value == "rtl" {
			continue
		}
		choose("Alert configuration", recipe.Label)
		a, found := r.AccessByRole(ui.RoleAlert, "")
		if !found || a.Bounds.W <= 0 || a.Bounds.W > 512.1 || a.Bounds.H <= 0 {
			t.Fatalf("source Alert %s is missing or unbounded: %v", recipe.Value, a.Bounds)
		}
	}
	choose("Alert configuration", "Icon and two inline links")
	statusClick(t, r, "second link")
	if p.alertAction != "Local alert link: second" {
		t.Fatal("the second native inline link did not activate independently")
	}
	link, found := r.AccessByRole(ui.RoleLink, "But it has a link")
	if !found || link.Bounds.H <= 0 {
		t.Fatal("rich description's first real native link is missing")
	}
	statusSnapshot(t, r, "status-alert-links-light")
	choose("Alert configuration", "Destructive payment checklist")
	list := compoundAccess(t, r, ui.RoleList, "Payment recovery checklist")
	item := compoundAccess(t, r, ui.RoleListItem, "Verify billing address")
	if item.PosInSet != 3 || item.SetSize != 3 || item.Bounds.Y < list.Bounds.Y {
		t.Fatal("source payment list lost native positions")
	}
	statusClick(t, r, "billing information")
	if p.alertAction != "Local alert link: billing" {
		t.Fatal("destructive inline billing link is inert")
	}
	choose("Alert configuration", "Enable dark mode action")
	statusClick(t, r, "Enable")
	if !p.enabled || p.alertAction != "Dark mode preference enabled locally" {
		t.Fatal("enable action did not update its local preference")
	}
	choose("Alert configuration", "Undo action")
	statusClick(t, r, "Undo")
	if !p.undone || p.alertAction != "Local spam action undone" {
		t.Fatal("Undo did not update the local action state")
	}
	choose("Alert configuration", "Amber subscription notice")
	assertAmber := func(shade int) {
		t.Helper()
		a, found := r.AccessByRole(ui.RoleAlert, "")
		if !found {
			t.Fatal("amber Alert missing")
		}
		pixel := color.NRGBAModel.Convert(r.Image().At(int(a.Bounds.X+a.Bounds.W/2), int(a.Bounds.Y+5))).(color.NRGBA)
		want := statusAmber(shade)
		if pixel.R != want.R || pixel.G != want.G || pixel.B != want.B {
			t.Fatalf("amber native surface = %v, want shade%d %v", pixel, shade, want)
		}
	}
	assertAmber(50)
	dark = true
	r.Frame()
	assertAmber(950)
	choose("Alert configuration", "Translated RTL alerts")
	choose("Alert language", "עברית")
	if !r.HasText("התשלום בוצע בהצלחה") || !r.HasText("תכונה חדשה זמינה") {
		t.Fatal("translated alert pair is missing")
	}
	choose("Alert configuration", "Account updated")
	dark = false
	r.Frame()
	choose("Empty configuration", "Search input group")
	assertSearchWidth := func(percent float32) {
		t.Helper()
		editor := statusReveal(t, r, "Search missing pages")
		empty, found := r.AccessByRole(ui.RoleGroup, "Source empty")
		if !found {
			t.Fatal("source Empty is missing")
		}
		// Inspect the actual enclosing native InputGroup, including its search
		// and Kbd addons, rather than the narrower editor or a text glyph box.
		var group ui.AccessInfo
		for _, item := range r.AccessAll() {
			b, e := item.Bounds, editor.Bounds
			if item.Role == ui.RoleGroup && b.H <= e.H+4 && b.W > e.W && b.X <= e.X && b.X+b.W >= e.X+e.W && b.Y <= e.Y && b.Y+b.H >= e.Y+e.H {
				if group.Bounds.W == 0 || b.W < group.Bounds.W {
					group = item
				}
			}
		}
		// Source Empty has p-12 and max-w-sm content; the docs media query is
		// sm:w-3/4, while Create's two search examples use unqualified w-3/4.
		contentWidth := min(float32(384), empty.Bounds.W-96)
		want := contentWidth * percent / 100
		if group.Bounds.W <= 0 || math.Abs(float64(group.Bounds.W-want)) > .2 {
			t.Errorf("source %s InputGroup width = %.2f, want %.2f (%.0f%% of %.2f)", p.emptyRecipe, group.Bounds.W, want, percent, contentWidth)
		}
	}
	assertSearchWidth(100)
	r.SetSize(640, 1120)
	assertSearchWidth(75)
	r.SetSize(620, 1120)
	statusClick(t, r, "Search missing pages")
	r.Type("component docs")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if *p.query("search") != "component docs" || p.emptyAction != "Local page search: component docs" || !r.Focused("Search missing pages") {
		t.Fatal("real Empty InputGroup editor did not retain and submit its draft")
	}
	statusClick(t, r, "Contact support")
	if p.emptyAction != "Local help action: Contact support" {
		t.Fatal("native support link is inert")
	}
	statusSnapshot(t, r, "status-empty-search-light")
	choose("Empty configuration", "Create: Bordered search")
	assertSearchWidth(75)
	statusClick(t, r, "Search missing pages")
	r.Type("separate query")
	if *p.query("border-search") != "separate query" || *p.query("search") != "component docs" {
		t.Fatal("source search configurations shared an editor model")
	}
	choose("Empty configuration", "Muted notifications")
	statusClick(t, r, "Refresh")
	statusClick(t, r, "Refresh")
	if p.refreshes != 2 || p.emptyAction != "Local notification refresh: 2" {
		t.Fatal("notification refresh did not report current local state")
	}
	choose("Empty configuration", "Offline user")
	avatar := statusReveal(t, r, "@shadcn")
	if avatar.Role != ui.RoleImage || avatar.Bounds.W != 48 || avatar.Bounds.H != 48 {
		t.Fatal("source offline avatar is not a 48DIP real image")
	}
	statusClick(t, r, "Leave Message")
	if p.emptyAction != "Local action: Leave Message" {
		t.Fatal("offline action is inert")
	}
	choose("Empty configuration", "No team members")
	dark = true
	r.Frame()
	for _, handle := range []string{"@shadcn", "@maxleiter", "@evilrabbit"} {
		a := statusReveal(t, r, handle)
		if a.Role != ui.RoleImage || a.Bounds.W != 48 {
			t.Fatal("source avatar group lost its original image or diameter")
		}
	}
	statusClick(t, r, "Invite Members")
	if p.emptyAction != "Local action: Invite Members" {
		t.Fatal("invitation action is inert")
	}
	statusSnapshot(t, r, "status-empty-team-dark")
	for _, recipe := range statusEmptyRecipes {
		if recipe.Value == "rtl" {
			continue
		}
		choose("Empty configuration", recipe.Label)
		if recipe.Value == "muted-search" {
			assertSearchWidth(75)
		}
		a, found := r.AccessByRole(ui.RoleGroup, "Source empty")
		if !found || a.Bounds.W <= 0 || a.Bounds.H <= 0 || a.Bounds.W > 580.1 {
			t.Fatalf("source Empty %s is not bounded: %v", recipe.Value, a.Bounds)
		}
	}
	choose("Empty configuration", "Create: Basic projects")
	statusClick(t, r, "Create project")
	if _, found := r.AccessByRole(ui.RoleLink, "Create project"); !found || p.emptyAction != "Local action: Create project" {
		t.Fatal("source linked button did not retain real native Link behavior")
	}
	statusClick(t, r, "Import project")
	if p.emptyAction != "Local action: Import project" {
		t.Fatal("the adjacent project action is not independent")
	}
	choose("Empty configuration", "Dashed cloud storage")
	statusClick(t, r, "Upload Files")
	if p.emptyAction != "Local action: Upload Files" {
		t.Fatal("upload preview action is inert")
	}
	choose("Empty configuration", "Create: No posts")
	statusReveal(t, r, "creating your first post")
	// A wrapped inline link's rectangular union can have empty space at its
	// center. Use its real Tab/Enter path rather than clicking that whitespace.
	r.Key(0, ui.KeyTab)
	r.Frame()
	if !r.Focused("creating your first post") {
		t.Fatal("source inline post link did not follow its selector in native Tab order")
	}
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p.emptyAction != "Local action: creating your first post" {
		t.Fatal("inline post link is inert")
	}
	choose("Empty configuration", "Translated RTL projects")
	choose("Empty language", "עברית")
	r.SetSize(500, 1120)
	statusClick(t, r, "צור פרויקט")
	if p.emptyAction != "Local action: צור פרויקט" {
		t.Fatal("translated project action is inert")
	}
	statusSnapshot(t, r, "status-empty-rtl-dark")
	compoundClick(t, r, "Leave status examples")
	compoundClick(t, r, "Show status examples")
	choose("Empty configuration", "Search input group")
	if value := statusReveal(t, r, "Search missing pages").Value; value != "component docs" {
		t.Fatalf("page return lost the actual source search draft: %q", value)
	}
	if len(r.OpenedURLs()) != 0 || p.popupError != "" {
		t.Fatal("placeholder links opened an external URL or a closed selector reported an error")
	}
	// The styled selector's collection is a genuine independent child Context.
	native = false
	r.Frame()
	statusClick(t, r, "Alert configuration")
	if len(backend.surfaces) == 0 {
		t.Fatal("styled selector did not open an independent surface")
	}
	surface := backend.surfaces[len(backend.surfaces)-1]
	if surface.spec.SystemPopover {
		t.Fatal("styled selector used the native system container")
	}
	surface.tester.Type("long title")
	surface.tester.Key(0, ui.KeyEnter)
	r.Frame()
	if p.alertRecipe != "long-title" || p.popupError != "" || !strings.Contains(p.emptyAction, "צור") {
		t.Fatal("independent selector did not return the current source recipe or preserve unrelated action state")
	}
}
