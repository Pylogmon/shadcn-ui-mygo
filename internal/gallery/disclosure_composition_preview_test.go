package gallery

import (
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func disclosureCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_DISCLOSURE_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("disclosure snapshot %v/%v", err, closeErr)
	}
}

// A single application flow visits the source layouts. The final page owner
// separately checks actual Gallery search/navigation; this owner is a module.
func TestDisclosureCompositionsSourceWorkflowAndNativeDrafts(t *testing.T) {
	p := &disclosureCompositionPreviewState{}
	backend := &creatorPopupBackend{}
	host := desktop.NewHostWithBackend(backend)
	native, dark, visible := true, false, true
	var anchor, caret int
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			host.BeginFrame()
			defer host.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave disclosures").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show disclosures").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, host, native)
				if e := c.Resolve(p.radiusHandles[2]); e.Valid() {
					anchor, caret = e.TextSelection()
				}
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 1120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, label string) ui.AccessInfo {
		t.Helper()
		return compoundAccess(t, r, role, label)
	}
	click := func(role ui.Role, label string) {
		t.Helper()
		a := access(role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	choose := func(label string) {
		t.Helper()
		click(ui.RoleMenuButton, "Disclosure configuration")
		if err := r.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	near := func(a, b float32) bool { return math.Abs(float64(a-b)) <= .1 }
	if p.recipe != "settings" || p.popupError != "" {
		t.Fatal("the initial source layout or quiet unopened Host changed")
	}
	x, y := access(ui.RoleTextField, "Radius X"), access(ui.RoleTextField, "Radius Y")
	expand := access(ui.RoleDisclosure, "Expand radius settings")
	if !near(x.Bounds.Y, y.Bounds.Y) || !near(expand.Bounds.Y, x.Bounds.Y) || expand.Bounds.X <= y.Bounds.X+y.Bounds.W || expand.Expanded {
		t.Fatal("radius inputs are not a constant grid beside an independent expander")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Expanded Radius X"); ok {
		t.Fatal("closed settings exposed its additional native editor")
	}
	compoundType(t, r, "Radius X", "12")
	r.Key(0, ui.KeyTab)
	if !r.Focused("Radius Y") {
		t.Fatal("the ordinary radius inputs lost source Tab order")
	}
	r.Key(0, ui.KeyTab)
	if !access(ui.RoleDisclosure, "Expand radius settings").Focused {
		t.Fatal("a closed content slot intercepted Tab before the expander")
	}
	r.Key(0, ui.KeySpace)
	r.Frame()
	x, y = access(ui.RoleTextField, "Radius X"), access(ui.RoleTextField, "Radius Y")
	x2, y2 := access(ui.RoleTextField, "Expanded Radius X"), access(ui.RoleTextField, "Expanded Radius Y")
	if !near(x2.Bounds.Y, y2.Bounds.Y) || !near(x.Bounds.X, x2.Bounds.X) || !near(y.Bounds.X, y2.Bounds.X) || !near(x.Bounds.W, x2.Bounds.W) || !near(x2.Bounds.Y-x.Bounds.Y-x.Bounds.H, 8) {
		t.Fatal("Content-before-Trigger did not share the source two-column grid")
	}
	click(ui.RoleTextField, "Radius Y")
	r.Key(0, ui.KeyTab)
	if !r.Focused("Expanded Radius X") {
		t.Fatal("deferred content construction changed native document Tab order")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("6.25")
	r.Key(ui.Cmd, ui.KeyA)
	r.Frame()
	before := p.radiusHandles[2]
	if anchor != 0 || caret != 4 {
		t.Fatalf("additional radius did not keep the actual editor selection: %d/%d", anchor, caret)
	}
	*p.open("settings", false) = false
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleTextField, "Expanded Radius X"); ok || !access(ui.RoleDisclosure, "Expand radius settings").Focused {
		t.Fatal("external close did not hide the panel and restore its real trigger")
	}
	click(ui.RoleDisclosure, "Expand radius settings")
	if p.radiusHandles[2] != before || p.radius[2] != "6.25" || anchor != 0 || caret != 4 {
		t.Fatal("reopening retained radius content lost its editor identity or selection")
	}
	disclosureCompositionSnapshot(t, r, "disclosure-settings-light")
	choose("Order details")
	if !r.HasText("Shipped") || r.HasText("100 Market St, San Francisco") {
		t.Fatal("the Order source status did not remain outside the closed panel")
	}
	click(ui.RoleDisclosure, "Toggle order details")
	if !r.HasText("100 Market St, San Francisco") || len(access(ui.RoleDisclosure, "Toggle order details").Controls) != 1 {
		t.Fatal("Order did not reveal its associated shipping details")
	}
	click(ui.RoleDisclosure, "Toggle order details")
	if !r.HasText("Shipped") || r.HasText("100 Market St, San Francisco") {
		t.Fatal("closing Order hid the source's persistent status or retained its body")
	}
	choose("Product details")
	click(ui.RoleDisclosure, "Product details")
	click(ui.RoleButton, "Learn More")
	if p.action != "Local product details preview" {
		t.Fatal("source product action did not use its native button")
	}
	choose("File tree")
	click(ui.RoleDisclosure, "Folder components")
	click(ui.RoleDisclosure, "Folder components/ui")
	click(ui.RoleButton, "Open components/ui/button.tsx")
	if p.selectedFile != "components/ui/button.tsx" {
		t.Fatal("nested source explorer did not preserve its selected application path")
	}
	click(ui.RoleTab, "Outline")
	if p.treeTab != "settings" || !access(ui.RoleTab, "Outline").Selected {
		t.Fatal("source Explorer/Outline tabs did not preserve actual native selection")
	}
	click(ui.RoleDisclosure, "Folder components")
	if _, ok := r.AccessByRole(ui.RoleButton, "Open components/ui/button.tsx"); ok {
		t.Fatal("a collapsed folder exposed the nested file action")
	}
	click(ui.RoleDisclosure, "Folder components")
	access(ui.RoleButton, "Open components/ui/button.tsx")
	dark = true
	r.Frame()
	disclosureCompositionSnapshot(t, r, "disclosure-file-tree-dark")
	choose("Create: Subscription card")
	click(ui.RoleLink, "Annual billing is available")
	if p.action != "Local annual-billing information" {
		t.Fatal("rich source paragraph did not retain its actual native link")
	}
	click(ui.RoleButton, "View plans")
	if p.action != "Local subscription-plan preview" {
		t.Fatal("source Card's independent action did not run")
	}
	click(ui.RoleDisclosure, "How does billing work?")
	if !*p.open("create-card:plans", false) || !*p.open("create-card:billing", false) {
		t.Fatal("Create card did not retain its source multiple/default-open state")
	}
	disclosureCompositionSnapshot(t, r, "disclosure-subscription-card-dark")
	choose("Account questions")
	click(ui.RoleDisclosure, "Can I change my subscription plan?")
	if *p.open("accordion-basic:item-1", false) || !*p.open("accordion-basic:item-2", false) {
		t.Fatal("single Accordion did not close the previous source question")
	}
	choose("Preference questions")
	click(ui.RoleDisclosure, "Privacy & Security")
	if !*p.open("accordion-multiple:notifications", false) || !*p.open("accordion-multiple:privacy", false) {
		t.Fatal("source multiple Accordion incorrectly closed its first section")
	}
	choose("Create: Long multiple accordion")
	for _, q := range disclosureQuestions["create-multiple"] {
		click(ui.RoleDisclosure, q.title)
		if !*p.open("create-multiple:"+q.key, false) {
			t.Fatal("long source question did not open")
		}
	}
	choose("Create: Disabled accordion")
	if !access(ui.RoleDisclosure, "Premium feature information").Disabled {
		t.Fatal("source premium section lost native disabled semantics")
	}
	click(ui.RoleDisclosure, "Can I access my account history?")
	r.Key(0, ui.KeyDown)
	if !access(ui.RoleDisclosure, "How do I update my email address?").Focused {
		t.Fatal("source header keyboard navigation did not skip the disabled section")
	}
	for _, config := range []string{"Shipping questions", "Joined billing questions", "Disabled account question", "Subscription card", "Create: Basic accordion", "Create: Separate borders"} {
		choose(config)
		if len(disclosureQuestions[p.recipe]) == 0 {
			t.Fatal("source Accordion configuration has no native question collection")
		}
	}
	choose("Translated account questions")
	if !r.HasText(disclosureTranslations["ar"][0].title) {
		t.Fatal("Arabic source questions missing")
	}
	choose("Translated order")
	click(ui.RoleMenuButton, "Disclosure language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	click(ui.RoleDisclosure, "Toggle order details")
	r.SetSize(420, 1120)
	if !r.HasText("2x אוזניות סטודיו") {
		t.Fatal("Hebrew order details did not retain their translated content")
	}
	disclosureCompositionSnapshot(t, r, "disclosure-order-hebrew-dark")
	compoundClick(t, r, "Leave disclosures")
	compoundClick(t, r, "Show disclosures")
	choose("Radius settings")
	if access(ui.RoleTextField, "Radius X").Value != "12" || access(ui.RoleTextField, "Expanded Radius X").Value != "6.25" {
		t.Fatal("source radius drafts did not survive recipe and page changes")
	}
	native = false
	r.Frame()
	compoundClick(t, r, "Disclosure configuration")
	if len(backend.surfaces) == 0 {
		t.Fatal("styled source configuration did not use an independent window")
	}
	surface := backend.surfaces[len(backend.surfaces)-1]
	if surface.spec.SystemPopover {
		t.Fatal("styled source configuration used a system popup")
	}
	surface.tester.Type("product")
	surface.tester.Key(0, ui.KeyEnter)
	r.Frame()
	if p.recipe != "product" || p.popupError != "" {
		t.Fatalf("styled source selector did not return current data: %q %q", p.recipe, p.popupError)
	}
}
