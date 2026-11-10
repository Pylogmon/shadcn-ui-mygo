package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func itemReveal(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	for range 8 {
		a, found := r.AccessByRole(role, label)
		if role == ui.RoleAuto {
			a, found = r.Access(label)
		}
		v, viewport := r.AccessByRole(ui.RoleRegion, "Item workspace")
		if !found || !viewport {
			t.Fatalf("missing item control %v/%s", role, label)
		}
		b := a.Bounds
		if b.W <= 0 || b.H <= 0 || b.X < v.Bounds.X-1 || b.X+b.W > v.Bounds.X+v.Bounds.W+1 {
			t.Fatalf("unbounded item control %s: %v in %v", label, b, v.Bounds)
		}
		if b.Y >= v.Bounds.Y && b.Y+b.H <= v.Bounds.Y+v.Bounds.H {
			return a
		}
		r.Scroll(v.Bounds.X+v.Bounds.W-8, v.Bounds.Y+v.Bounds.H/2, 0, b.Y+b.H/2-v.Bounds.Y-v.Bounds.H/2)
	}
	t.Fatalf("could not reveal item control %s", label)
	return ui.AccessInfo{}
}
func itemClick(t *testing.T, r *ui.Tester, role ui.Role, label string) {
	t.Helper()
	a := itemReveal(t, r, role, label)
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
}
func itemSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_ITEM_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("item snapshot %v/%v", err, closeErr)
	}
}

// One workflow uses real native targets and menu selections, not snapshots
// substituted for interaction. It is a public Tester/Backend, not a real OS.
func TestItemCompositionsSourceWorkflow(t *testing.T) {
	p := &itemCompositionPreviewState{}
	backend := &creatorPopupBackend{}
	host := desktop.NewHostWithBackend(backend)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			host.BeginFrame()
			defer host.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave item examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show item examples").Clicked() {
					visible = true
				}
			})
			ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Role(ui.RoleRegion).Label("Item workspace").Children(func() {
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
	choose := func(label string) {
		t.Helper()
		itemClick(t, r, ui.RoleMenuButton, "Item configuration")
		if err := r.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	clickTrailingVisual := func(label string, padding float32) {
		t.Helper()
		a := itemReveal(t, r, ui.RoleListItem, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W-padding-8, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	if p.recipe != "demo" || p.popupError != "" {
		t.Fatal("initial source recipe or quiet closed selector missing")
	}
	itemClick(t, r, ui.RoleButton, "Run basic item action")
	if p.action != "Local item action: Basic Item" {
		t.Fatal("independent source action was inert")
	}
	clickTrailingVisual("Your profile has been verified.", 12)
	if p.action != "Local item link: Your profile has been verified." {
		t.Error("verified source link did not include its trailing decorative icon")
	}
	choose("Native item links")
	for _, label := range []string{"Visit our documentation", "External resource"} {
		clickTrailingVisual(label, 16)
		if p.action != "Local item link: "+label {
			t.Errorf("source link %s did not include its trailing decorative icon", label)
		}
	}
	// All 35 source configurations are selected through the actual menu.
	for _, recipe := range itemCompositionRecipes {
		if recipe.Value == "rtl" || recipe.Value == "dropdown" {
			continue
		}
		choose(recipe.Label)
		if p.recipe != recipe.Value || p.popupError != "" {
			t.Fatalf("source recipe %s did not build", recipe.Value)
		}
		if recipe.Value != "header" {
			if _, found := r.AccessByRole(ui.RoleList, ""); !found {
				t.Fatalf("source collection %s lost native list semantics", recipe.Value)
			}
		}
	}
	choose("Create: Default")
	itemClick(t, r, ui.RoleButton, "Action for Title + Description + Button")
	if p.action != "Local item action: Title + Description + Button" {
		t.Fatal("compound trailing action was inert")
	}
	itemClick(t, r, ui.RoleButton, "Confirm multiple actions")
	if p.action != "Local item action: Confirm multiple actions" {
		t.Fatal("second action target did not activate independently")
	}
	itemSnapshot(t, r, "item-create-actions-light")
	choose("Create: Outline links")
	itemClick(t, r, ui.RoleLink, "With Actions (Link)")
	if p.action != "Local item link: With Actions (Link)" {
		t.Fatal("native composed link did not activate")
	}
	itemClick(t, r, ui.RoleButton, "Share linked item")
	if p.action != "Local item action: Share linked item" {
		t.Fatal("Share also dispatched the primary link")
	}
	choose("Music with cover images")
	for i, song := range itemSongs {
		a := itemReveal(t, r, ui.RoleListItem, song.title)
		link := itemReveal(t, r, ui.RoleLink, song.title)
		if a.PosInSet != i+1 || a.SetSize != 3 || !strings.Contains(link.Description, song.duration) || !strings.Contains(link.Description, song.album) {
			t.Fatal("source track list positions, album or duration semantics missing")
		}
		itemClick(t, r, ui.RoleLink, song.title)
		if p.action != "Local item link: "+song.title {
			t.Fatal("track native link is inert")
		}
	}
	itemSnapshot(t, r, "item-music-light")
	// Inspect actual native image pixels inside the third track's 40DIP media.
	track := itemReveal(t, r, ui.RoleLink, "Digital Rain")
	grayPixels := 0
	for y := 8; y < 32; y++ {
		for x := 8; x < 32; x++ {
			rr, gg, bb, _ := r.Image().At(int(track.Bounds.X)+x, int(track.Bounds.Y)+y).RGBA()
			if rr != gg || gg != bb {
				t.Fatal("source grayscale cover retained colored pixels")
			}
			if rr < 60000 {
				grayPixels++
			}
		}
	}
	if grayPixels < 20 {
		t.Fatal("actual native source cover did not draw")
	}
	choose("User and team avatars")
	if _, found := r.AccessByRole(ui.RoleImage, "@shadcn"); found {
		t.Fatal("sm:hidden source avatar was visible below640")
	}
	r.SetSize(640, 1120)
	if _, found := r.AccessByRole(ui.RoleImage, "@shadcn"); !found {
		t.Fatal("source avatar did not appear at640")
	}
	itemClick(t, r, ui.RoleButton, "Invite Evil Rabbit")
	if p.action != "Local item action: Invite Evil Rabbit" {
		t.Fatal("avatar trailing action was inert")
	}
	choose("Model image headers")
	for _, name := range []string{"v0-1.5-sm", "v0-1.5-lg", "v0-2.0-mini"} {
		a := itemReveal(t, r, ui.RoleImage, name)
		if a.Bounds.W <= 0 || a.Bounds.H <= 0 || a.Bounds.W-a.Bounds.H > .2 || a.Bounds.H-a.Bounds.W > .2 {
			t.Fatal("source header image is not square")
		}
	}
	itemSnapshot(t, r, "item-model-headers-light")
	choose("People dropdown")
	itemClick(t, r, ui.RoleMenuButton, "Select item person")
	if err := r.ChooseMenuItem("maxleiter · maxleiter@vercel.com"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.selectedPerson != "maxleiter" || p.action != "Local person selection: maxleiter" {
		t.Fatal("system text menu did not return selected person")
	}
	dark, native = true, false
	r.Frame()
	itemClick(t, r, ui.RoleAuto, "Select item person")
	if len(backend.surfaces) == 0 {
		t.Fatal("styled people menu did not use independent Flyout")
	}
	child := backend.surfaces[len(backend.surfaces)-1]
	child.tester.Advance(time.Second)
	if r.HasText("evilrabbit@vercel.com") || !child.tester.HasText("evilrabbit@vercel.com") {
		t.Fatal("rich Item menu leaked to parent or omitted actual child context content")
	}
	previousBottom := float32(0)
	for _, person := range itemPeople {
		row, found := child.tester.AccessByRole(ui.RoleMenuItem, person.name+" · "+person.email)
		// The actual26DIP source avatar plus p-2 needs42DIP before the
		// longer two-line caption is considered. A fixed32DIP target cannot
		// contain this source content. Menu children are intentionally inert AX.
		if !found || row.Bounds.H < 26+16 || row.Bounds.Y < previousBottom-.1 {
			t.Fatalf("rich source row cannot contain avatar/padding or overlaps: %s %v", person.name, row.Bounds)
		}
		previousBottom = row.Bounds.Y + row.Bounds.H
	}
	itemSnapshot(t, child.tester, "item-people-menu-dark")
	menuItem, found := child.tester.AccessByRole(ui.RoleMenuItem, "evilrabbit · evilrabbit@vercel.com")
	if !found {
		t.Fatal("the rich menu target omitted the source name or email")
	}
	child.tester.ClickAt(menuItem.Bounds.X+menuItem.Bounds.W/2, menuItem.Bounds.Y+menuItem.Bounds.H/2)
	r.Frame()
	if p.selectedPerson != "evilrabbit" || p.popupError != "" {
		t.Fatal("styled rich menu did not return current item")
	}
	native = true
	r.Frame()
	choose("Translated RTL items")
	itemClick(t, r, ui.RoleMenuButton, "Item language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	r.SetSize(420, 1120)
	itemClick(t, r, ui.RoleButton, "פעולה")
	if p.action != "Local item action: פריט בסיסי" {
		t.Fatal("translated action did not preserve its native language")
	}
	itemSnapshot(t, r, "item-rtl-hebrew-dark")
	if len(r.OpenedURLs()) != 0 {
		t.Fatal("placeholder source links launched an unrelated external URL")
	}
	if err := r.Click("Leave item examples"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if err := r.Click("Show item examples"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.recipe != "rtl" || p.language != "he" || p.selectedPerson != "evilrabbit" {
		t.Fatal("page navigation lost independent Item application state")
	}
}
