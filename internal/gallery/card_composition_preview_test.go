package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func cardSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_CARD_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("card snapshot %v/%v", err, closeErr)
	}
}

// One application workflow visits the source compositions. This is the module
// owner, not the complete Gallery State.View or a real operating-system window.
func TestCardCompositionsSourceWorkflowAndDrafts(t *testing.T) {
	p := &cardCompositionPreviewState{}
	backend := &creatorPopupBackend{}
	host := desktop.NewHostWithBackend(backend)
	native, dark, visible := true, false, true
	palette := "blue"
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset(palette, dark)
		theme.Scope(c, tokens).Fill().Background(tokens.Background).TextColor(tokens.Foreground).Padding(20).Gap(12).Children(func() {
			host.BeginFrame()
			defer host.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave cards").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show cards").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, host, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	choose := func(label string) {
		t.Helper()
		compoundClick(t, r, "Card configuration")
		if err := r.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	assertCard := func() {
		t.Helper()
		a := compoundAccess(t, r, ui.RoleGroup, "Source card")
		if a.Bounds.W > 384.1 || a.Bounds.X < 0 || a.Bounds.X+a.Bounds.W > 620.1 {
			t.Fatalf("unbounded source card %+v", a.Bounds)
		}
	}
	if p.error != "" || p.recipe != "default" || !r.HasText("Default Card") {
		t.Fatal("initial source card or quiet unopened host missing")
	}
	compoundClick(t, r, "Run card action")
	if p.action != "Local card action" {
		t.Fatal("default source action inert")
	}
	choose("Small size")
	assertCard()
	if !r.HasText("Small Card") {
		t.Fatal("small source content missing")
	}
	for _, label := range []string{"Header with border", "Header with border (small)", "Footer with border", "Footer with border (small)"} {
		choose(label)
		assertCard()
		if strings.HasPrefix(label, "Footer") {
			compoundClick(t, r, "Run bordered footer action")
			if p.action != "Local bordered footer action" {
				t.Fatal("bordered footer action inert")
			}
		}
	}
	choose("Custom spacing")
	before := compoundAccess(t, r, ui.RoleGroup, "Source card").Bounds
	compoundClick(t, r, "32px")
	after := compoundAccess(t, r, ui.RoleGroup, "Source card").Bounds
	if p.spacingValue() != 32 || after.H <= before.H+30 {
		t.Fatalf("spacing did not affect actual card geometry: %v -> %v", before, after)
	}
	compoundClick(t, r, "32px")
	if p.spacingValue() != 32 {
		t.Fatal("single spacing selection allowed an empty choice")
	}
	choose("Login")
	compoundClick(t, r, "Login")
	if p.login[0].submitted || p.login[0].error == "" {
		t.Fatal("empty login claimed a local success")
	}
	compoundType(t, r, "Email *", "native@example.com")
	compoundType(t, r, "Password", "correct horse")
	secure := compoundAccess(t, r, ui.RoleTextField, "Password")
	if !secure.Password || secure.Value != "" || p.login[0].password != "correct horse" {
		t.Fatal("source password is not a genuine secure editor")
	}
	compoundClick(t, r, "Login")
	if !p.login[0].submitted || p.login[0].error != "" || p.action != "Local login preview: native@example.com" || strings.Contains(p.action, p.login[0].password) {
		t.Fatal("login did not read current local input or exposed the password")
	}
	compoundClick(t, r, "Forgot your password?")
	if p.action != "Local password-recovery preview" {
		t.Fatal("separate recovery action inert")
	}
	compoundClick(t, r, "Login with Google")
	if p.action != "Local Google sign-in preview" {
		t.Fatal("provider action inert")
	}
	compoundClick(t, r, "Sign up")
	if p.action != "Local sign-up preview" {
		t.Fatal("footer signup action inert")
	}
	choose("Login with header action")
	if compoundAccess(t, r, ui.RoleTextField, "Email *").Value != "" {
		t.Fatal("independent source login examples shared a draft")
	}
	compoundClick(t, r, "Sign Up")
	if p.action != "Local sign-up preview" {
		t.Fatal("header action inert")
	}
	choose("Login with custom spacing")
	compoundType(t, r, "Email *", "spacing@example.com")
	compoundType(t, r, "Password", "private draft")
	compoundClick(t, r, "20px")
	if compoundAccess(t, r, ui.RoleTextField, "Email *").Value != "spacing@example.com" || p.login[2].password != "private draft" {
		t.Fatal("spacing change replaced native editor values")
	}
	cardSnapshot(t, r, "card-login-spacing-light")
	choose("Content edge to edge")
	if p.terms.MaxY <= 0 {
		t.Fatal("source terms are not in a bounded scroll viewport")
	}
	termsBefore := p.terms.Y
	p.terms.ScrollTo(0, p.terms.MaxY, ui.Animation{})
	r.Frame()
	if p.terms.Y <= termsBefore {
		t.Fatal("source terms could not scroll")
	}
	compoundClick(t, r, "Accept")
	if p.termsDecision != "Accept" || p.action != "Local agreement decision: Accept" {
		t.Fatal("accepted agreement did not update local model")
	}
	compoundClick(t, r, "Decline")
	if p.termsDecision != "Decline" {
		t.Fatal("decline action not independent")
	}
	choose("Meeting notes")
	compoundClick(t, r, "Transcribe")
	if !p.transcribed || !r.HasText("New analytics widgets for daily/weekly metrics") {
		t.Fatal("source note text/action missing")
	}
	compoundAccess(t, r, ui.RoleList, "Meeting agenda")
	agendaItem := compoundAccess(t, r, ui.RoleListItem, "New analytics widgets for daily/weekly metrics")
	if agendaItem.PosInSet != 1 || agendaItem.SetSize != 5 {
		t.Fatal("ordered agenda lost its native item positions")
	}
	compoundAccess(t, r, ui.RoleGroup, "Meeting participants")
	for _, name := range []string{"@shadcn", "@maxleiter", "@evilrabbit"} {
		compoundAccess(t, r, ui.RoleImage, name)
	}
	cardSnapshot(t, r, "card-meeting-notes-light")
	choose("With image")
	image := compoundAccess(t, r, ui.RoleImage, "Photo by mymind on Unsplash")
	if image.Bounds.W/image.Bounds.H < 1.77 || image.Bounds.W/image.Bounds.H > 1.79 {
		t.Fatalf("source landscape is not 16:9 %+v", image.Bounds)
	}
	compoundClick(t, r, "Add landscape")
	cover := p.cover
	palette, dark = "violet", true
	r.Frame()
	if p.cover == cover || p.coverColor != theme.Preset(palette, dark).Primary {
		t.Fatal("source color blend did not follow current palette")
	}
	cardSnapshot(t, r, "card-landscape-dark")
	choose("With image (small)")
	compoundClick(t, r, "Add landscape")
	if p.landscapes != 2 {
		t.Fatal("image sizes did not retain the same application action model")
	}
	choose("Featured event")
	compoundAccess(t, r, ui.RoleImage, "Event cover")
	if !r.HasText("Featured") {
		t.Fatal("source event action badge missing")
	}
	compoundClick(t, r, "View Event")
	if p.action != "Local event preview: Design systems meetup" {
		t.Fatal("event footer action inert")
	}
	choose("Scheduled reports")
	compoundAccess(t, r, ui.RoleList, "Scheduled report features")
	reportItem := compoundAccess(t, r, ui.RoleListItem, "Choose a schedule (daily, or weekly).")
	if reportItem.PosInSet != 1 || reportItem.SetSize != 3 {
		t.Fatal("report feature list lost its native item positions")
	}
	if compoundAccess(t, r, ui.RoleGroup, "Source card").Bounds.W > 320.1 {
		t.Fatal("small docs card exceeds max-w-xs")
	}
	compoundClick(t, r, "Set up scheduled reports")
	if p.action != "Local scheduled-report setup" {
		t.Fatal("report setup action inert")
	}
	compoundClick(t, r, "See what's new")
	if p.action != "Local report release notes" {
		t.Fatal("report secondary action inert")
	}
	choose("Translated RTL login")
	r.SetSize(500, 900)
	compoundClick(t, r, "Card language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	compoundType(t, r, "אימייל *", "rtl@example.com")
	compoundType(t, r, "סיסמה", "rtl secret")
	compoundClick(t, r, "התחבר")
	if !p.login[3].submitted || p.login[3].email != "rtl@example.com" {
		t.Fatal("translated RTL login lost actual editors")
	}
	cardSnapshot(t, r, "card-rtl-hebrew-dark")
	compoundClick(t, r, "Leave cards")
	compoundClick(t, r, "Show cards")
	if compoundAccess(t, r, ui.RoleTextField, "אימייל *").Value != "rtl@example.com" {
		t.Fatal("page return lost independent RTL draft")
	}
	choose("Login")
	if compoundAccess(t, r, ui.RoleTextField, "Email *").Value != "native@example.com" || p.login[0].password != "correct horse" {
		t.Fatal("returning to Create login lost its draft")
	}
	// Route switch uses the same source selection but opens an independent,
	// themed child. The collection is not painted into the parent Context.
	native = false
	r.Frame()
	compoundClick(t, r, "Card configuration")
	if len(backend.surfaces) == 0 {
		t.Fatal("styled configuration did not use Host")
	}
	surface := backend.surfaces[len(backend.surfaces)-1]
	if surface.spec.SystemPopover {
		t.Fatal("styled recipe selector used a system popup")
	}
	surface.tester.Type("meeting")
	surface.tester.Key(0, ui.KeyEnter)
	r.Frame()
	if p.recipe != "notes" || p.error != "" {
		t.Fatalf("independent styled selector did not return current recipe: %q error %q", p.recipe, p.error)
	}
}
