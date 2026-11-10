package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func tabsCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_TABS_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("tabs image %v/%v", err, closeErr)
	}
}

// One actual native-input flow checks the complete fixed source recipes. The
// injected menu backend and software PNGs do not prove OS/IME/AT acceptance.
func TestTabsCompositionsSourceWorkflow(t *testing.T) {
	p := &tabsCompositionPreviewState{}
	backend := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(backend)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Wrap().Children(func() {
				if ui.Button(c, "Leave tabs examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show tabs examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 760, 850, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, label string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, label)
		if !ok {
			t.Fatalf("missing native %v %q", role, label)
		}
		return a
	}
	click := func(role ui.Role, label string) {
		t.Helper()
		a := access(role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	choose := func(key string) {
		t.Helper()
		index := -1
		for i, recipe := range tabsSourceRecipes {
			if recipe.id == key {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatalf("unknown recipe %q", key)
		}
		if err := r.Click("Tabs configuration"); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(tabsSourceRecipes[index].label); err != nil {
				t.Fatal(err)
			}
		} else {
			popup := backend.surfaces[len(backend.surfaces)-1]
			popup.tester.Key(0, ui.KeyHome)
			for range index {
				popup.tester.Key(0, ui.KeyDown)
			}
			popup.tester.Key(0, ui.KeyEnter)
			if !popup.didClose {
				t.Fatal("styled configuration selector did not close")
			}
		}
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("configuration %q reached %q error=%q", key, p.recipe, p.popupError)
		}
	}
	if len(tabsSourceRecipes) != 20 || p.recipe != "demo" || !access(ui.RoleTab, "Overview").Selected || !r.HasText("You have 12 active projects and 3 pending tasks.") {
		t.Fatal("the complete source catalogue or docs dashboard default was lost")
	}
	click(ui.RoleTab, "Analytics")
	if p.drafts["demo"].value != "analytics" || p.drafts["demo"].changes != 1 || !r.HasText("Page views are up 25% compared to last month.") || access(ui.RoleTab, "Overview").Selected {
		t.Fatal("source tab selection/card did not update exactly once in the native frame")
	}
	r.Key(0, ui.KeyRight)
	r.Frame()
	if p.drafts["demo"].value != "reports" || p.drafts["demo"].changes != 2 || !r.HasText("You have 5 reports ready and available to export.") {
		t.Fatal("native source tab keyboard selection lost its panel")
	}
	tabsCompositionSnapshot(t, r, "tabs-dashboard-light")
	for _, recipe := range tabsSourceRecipes {
		choose(recipe.id)
	}
	choose("usage")
	click(ui.RoleTab, "Password")
	if !r.HasText("Change your password here.") || p.drafts["usage"].value != "password" {
		t.Fatal("the guide's distinct plain content was replaced by a Create panel")
	}
	choose("disabled")
	if !access(ui.RoleTab, "Disabled").Disabled {
		t.Fatal("source disabled did not reach the real tab")
	}
	click(ui.RoleTab, "Disabled")
	click(ui.RoleTab, "Home")
	r.Key(0, ui.KeyRight)
	r.Frame()
	if p.drafts["disabled"].value != "home" || p.drafts["disabled"].changes != 0 {
		t.Fatal("native navigation selected the disabled source trigger")
	}
	choose("vertical")
	click(ui.RoleTab, "Account")
	r.Key(0, ui.KeyDown)
	r.Frame()
	if p.drafts["vertical"].value != "password" || !access(ui.RoleTabList, "Source tabs main").Vertical {
		t.Fatal("the docs vertical source did not use real vertical native tabs")
	}
	r.Key(0, ui.KeyRight)
	r.Frame()
	if p.drafts["vertical"].value != "password" {
		t.Fatal("vertical tabs used a horizontal navigation key")
	}
	choose("create-comparison")
	second := access(ui.RoleTabList, "Source tabs line")
	found := false
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleTab && a.Name == "Analytics" && a.Bounds.X >= second.Bounds.X && a.Bounds.X+a.Bounds.W <= second.Bounds.X+second.Bounds.W+1 && a.Bounds.Y >= second.Bounds.Y && a.Bounds.Y+a.Bounds.H <= second.Bounds.Y+second.Bounds.H+1 {
			r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
			found = true
			break
		}
	}
	r.Frame()
	if !found || p.drafts["create-comparison"].value != "overview" || p.drafts["create-comparison"].second != "analytics" || p.drafts["create-comparison"].changes != 1 {
		t.Fatal("the source comparison's two actual tab roots were not independent")
	}
	choose("create-icon-only")
	for _, name := range []string{"Home", "Search", "Settings"} {
		if access(ui.RoleTab, name).Bounds.W > 60 {
			t.Fatal("source icon-only retained visible label width")
		}
	}
	click(ui.RoleTab, "Search")
	if p.drafts["create-icon-only"].value != "search" {
		t.Fatal("icon-only source lost its meaningful native target")
	}
	choose("create-line-disabled")
	click(ui.RoleTab, "Overview")
	r.Key(0, ui.KeyEnd)
	r.Frame()
	if p.drafts["create-line-disabled"].value != "analytics" || !access(ui.RoleTab, "Reports").Disabled {
		t.Fatal("line tabs did not skip the source's disabled last trigger")
	}
	choose("create-content")
	click(ui.RoleTab, "Notifications")
	if !r.HasText("Configure how you receive notifications and alerts.") {
		t.Fatal("Create source panel wording changed")
	}
	choose("create-vertical")
	click(ui.RoleTab, "Password")
	if !r.HasText("Update your password to keep your account secure. Use a strong password with a mix of letters, numbers, and symbols.") {
		t.Fatal("the longer Create vertical panel was collapsed into generic content")
	}
	dark = true
	r.Frame()
	tabsCompositionSnapshot(t, r, "tabs-vertical-dark")
	choose("create-input")
	field := access(ui.RoleTextField, "Search tabs")
	if field.Bounds.W != 176 {
		t.Fatalf("source header input width=%v want176", field.Bounds.W)
	}
	header, ok := r.Find("Tabs header viewport")
	action := access(ui.RoleButton, "Action")
	if !ok || action.Bounds.X+action.Bounds.W != header.X+header.W {
		t.Fatal("source ml-auto header actions did not align with the panel edge")
	}
	click(ui.RoleTextField, "Search tabs")
	r.Type("native query")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p.drafts["create-input"].query != "native query" || p.drafts["create-input"].actions != 0 {
		t.Fatal("source input was fake or invented an Enter action")
	}
	click(ui.RoleButton, "Action")
	if p.drafts["create-input"].actions != 1 || p.drafts["create-input"].savedQuery != "native query" || p.drafts["create-input"].savedTab != "overview" || p.action != "Local tabs action: Action" {
		t.Fatal("actual Action did not record its local source snapshot once")
	}
	click(ui.RoleTab, "Analytics")
	if access(ui.RoleTextField, "Search tabs").Value != "native query" || !r.HasText("Detailed analytics and insights about your data.") {
		t.Fatal("tab selection replaced the independent native header input")
	}
	tabsCompositionSnapshot(t, r, "tabs-input-dark")
	r.SetSize(360, 850)
	narrowHeader, ok := r.Find("Tabs header viewport")
	if !ok {
		t.Fatal("missing narrow source header viewport")
	}
	r.Scroll(narrowHeader.X+narrowHeader.W/2, narrowHeader.Y+narrowHeader.H/2, 600, 0)
	r.Frame()
	narrowAction := access(ui.RoleButton, "Action")
	if narrowAction.Bounds.X < narrowHeader.X || narrowAction.Bounds.X+narrowAction.Bounds.W > narrowHeader.X+narrowHeader.W+1 || access(ui.RoleTextField, "Search tabs").Bounds.W != 176 {
		t.Fatal("the narrow header did not retain/reveal its real source input and action")
	}
	click(ui.RoleButton, "Action")
	if p.drafts["create-input"].actions != 2 || p.drafts["create-input"].savedQuery != "native query" {
		t.Fatal("the horizontally revealed native Action did not consume its retained draft")
	}
	r.SetSize(760, 850)
	r.Frame()
	p.drafts["create-input"].value = "reports"
	r.Frame()
	if !r.HasText("Generate and view custom reports.") {
		t.Fatal("the source's untriggered Reports content declaration was lost")
	}
	choose("create-dropdown")
	click(ui.RoleMenuButton, "More options")
	if err := r.ChooseMenuItem("Export"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.action != "Local tabs menu: Export" || p.drafts["create-dropdown"].actions != 1 {
		t.Fatal("the native source menu did not dispatch one local intent")
	}
	native = false
	r.Frame()
	choose("create-dropdown")
	click(ui.RoleButton, "More options")
	popup := backend.surfaces[len(backend.surfaces)-1]
	popup.tester.Key(0, ui.KeyEnd)
	popup.tester.Key(0, ui.KeyEnter)
	r.Frame()
	if p.action != "Local tabs menu: Archive" || p.drafts["create-dropdown"].actions != 2 || !popup.didClose {
		t.Fatal("styled source menu did not use its independent child window")
	}
	native = true
	r.Frame()
	choose("rtl")
	if err := r.Click("Tabs language"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	r.SetSize(360, 850)
	click(ui.RoleTab, "אנליטיקה")
	if p.drafts["rtl"].value != "analytics" || !r.HasText("צפיות בדף עלו ב-25% בהשוואה לחודש שעבר.") {
		t.Fatal("the narrow translated source did not select its native Hebrew panel")
	}
	tabsCompositionSnapshot(t, r, "tabs-hebrew-narrow-dark")
	listViewport, ok := r.Find("Source tab list viewport")
	if !ok {
		t.Fatal("missing real translated source list viewport")
	}
	r.Scroll(listViewport.X+listViewport.W/2, listViewport.Y+listViewport.H/2, -400, 0)
	r.Frame()
	settings := access(ui.RoleTab, "הגדרות")
	if settings.Bounds.X < listViewport.X || settings.Bounds.X+settings.Bounds.W > listViewport.X+listViewport.W+1 {
		t.Fatal("horizontal native scrolling did not expose the complete last RTL trigger")
	}
	click(ui.RoleTab, "הגדרות")
	if p.drafts["rtl"].value != "settings" || !r.HasText("הגדר התראות, אבטחה וערכות נושא.") {
		t.Fatal("the horizontally revealed RTL native tab did not select its actual source panel")
	}
	if err := r.Click("Leave tabs examples"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleTab, "אנליטיקה"); ok {
		t.Fatal("leaving the page retained active source tabs")
	}
	if err := r.Click("Show tabs examples"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !access(ui.RoleTab, "הגדרות").Selected {
		t.Fatal("page return lost the independent native source selection")
	}
	r.SetSize(760, 850)
	native = false
	r.Frame()
	choose("demo")
	if p.drafts["demo"].value != "reports" || !access(ui.RoleTab, "Reports").Selected || p.drafts["create-input"].query != "native query" || p.drafts["create-dropdown"].actions != 2 {
		t.Fatal("recipe/style/page changes crossed independent application values")
	}
}
