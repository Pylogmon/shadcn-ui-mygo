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

func dialogCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	if dir := os.Getenv("GALLERY_DIALOG_COMPOSITION_SNAPSHOT_DIR"); dir != "" {
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
			t.Fatalf("dialog image %v / %v", err, closeErr)
		}
	}
}

// One real-input workflow covers the complete source catalogue and application
// drafts. The injected tooltip/menu Backend does not prove OS/AT/IME behavior.
func TestDialogCompositionsSourceWorkflow(t *testing.T) {
	p := &dialogCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Wrap().Gap(8).Children(func() {
				if ui.Button(c, "Leave dialog examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show dialog examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 760, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, name)
		if !ok {
			t.Fatalf("missing native role=%v name=%q", role, name)
		}
		return a
	}
	click := func(role ui.Role, name string) {
		t.Helper()
		a := access(role, name)
		if a.Offscreen {
			t.Fatalf("native %q is outside the actual viewport: %+v", name, a.Bounds)
		}
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	closeModal := func() {
		t.Helper()
		r.Key(0, ui.KeyEscape)
		r.Frame()
		r.Frame()
		if p.draft().open {
			t.Fatal("native Escape failed to dismiss source modal")
		}
	}
	choose := func(key string) {
		t.Helper()
		if p.drafts[p.recipe] != nil && p.drafts[p.recipe].open {
			closeModal()
		}
		index := -1
		for i, o := range dialogCompositionRecipes {
			if o.Value == key {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatalf("unknown dialog recipe %q", key)
		}
		click(ui.RoleMenuButton, "Dialog configuration")
		if native {
			if err := r.ChooseMenuItem(dialogCompositionRecipes[index].Label); err != nil {
				t.Fatal(err)
			}
		} else {
			s := b.surfaces[len(b.surfaces)-1]
			s.tester.Key(0, ui.KeyHome)
			for range index {
				s.tester.Key(0, ui.KeyDown)
			}
			s.tester.Key(0, ui.KeyEnter)
			if !s.didClose {
				t.Fatal("styled source configuration selector stayed open")
			}
		}
		r.Frame()
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("selected recipe=%s wanted=%s error=%s", p.recipe, key, p.popupError)
		}
	}
	// The selector's own native role changes with its presentation route.
	chooseStyled := func(key string) {
		t.Helper()
		index := -1
		for i, o := range dialogCompositionRecipes {
			if o.Value == key {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatal("missing recipe")
		}
		click(ui.RoleButton, "Dialog configuration")
		s := b.surfaces[len(b.surfaces)-1]
		s.tester.Key(0, ui.KeyHome)
		for range index {
			s.tester.Key(0, ui.KeyDown)
		}
		s.tester.Key(0, ui.KeyEnter)
		r.Frame()
		r.Frame()
		if p.recipe != key || !s.didClose || p.popupError != "" {
			t.Fatal("styled source selector failed its real child selection")
		}
	}
	open := func(label, title string) {
		t.Helper()
		click(ui.RoleButton, label)
		r.Frame()
		a := access(ui.RoleDialog, title)
		if !p.draft().open || a.Bounds.X < 15.5 || a.Bounds.Y < 15.5 {
			t.Fatalf("source native modal did not fit: %+v", a.Bounds)
		}
	}
	replace := func(name, text string) {
		t.Helper()
		click(ui.RoleTextField, name)
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(text)
		r.Frame()
	}
	if len(dialogCompositionRecipes) != 12 || p.recipe != "demo" {
		t.Fatal("the source dialog catalogue is incomplete")
	}
	open("Open Dialog", "Edit profile")
	if !access(ui.RoleTextField, "Name").Focused || access(ui.RoleDialog, "Edit profile").Bounds.W != 384 {
		t.Fatal("docs384 dialog or explicit first editor focus was lost")
	}
	if title, label := access(ui.RoleHeading, "Edit profile"), access(ui.RoleText, "Name"); title.Bounds.H > label.Bounds.H+.5 {
		t.Fatalf("source14DIP title did not match its native14DIP label metric: title%+v label%+v", title.Bounds, label.Bounds)
	}
	replace("Name", "Local source profile")
	replace("Username", "@native-profile")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p.draft().saves != 0 || !p.draft().open {
		t.Fatal("source form acquired an invented Enter submit/dismiss handler")
	}
	click(ui.RoleButton, "Save changes")
	if p.draft().saves != 1 || p.draft().saved["name"] != "Local source profile" || p.draft().saved["username"] != "@native-profile" || !p.draft().open {
		t.Fatal("real source Save failed local snapshot or invented dialog-close behavior")
	}
	dialogCompositionSnapshot(t, r, "dialog-profile-light")
	click(ui.RoleButton, "Cancel")
	r.Frame()
	r.Frame()
	if p.draft().open || !access(ui.RoleButton, "Open Dialog").Focused {
		t.Fatal("Cancel did not close and return focus to its actual source opener")
	}
	// Every original source configuration is independently reachable by input.
	for _, example := range []struct{ key, trigger, title string }{
		{"usage", "Open", "Are you absolutely sure?"}, {"share", "Share", "Share link"}, {"no-close", "No Close Button", "No Close Button"}, {"sticky", "Sticky Footer", "Sticky Footer"}, {"scrollable", "Scrollable Content", "Scrollable Content"}, {"rtl", "فتح الحوار", "تعديل الملف الشخصي"}, {"create-form", "Edit Profile", "Edit profile"}, {"create-scrollable", "Scrollable Content", "Scrollable Content"}, {"create-sticky", "Sticky Footer", "Scrollable Content"}, {"create-no-close", "No Close Button", "No Close Button"}, {"create-chat", "Chat Settings", "Chat Settings"},
	} {
		choose(example.key)
		open(example.trigger, example.title)
		closeModal()
	}
	choose("share")
	open("Share", "Share link")
	if !access(ui.RoleTextField, "Link").ReadOnly {
		t.Fatal("source shared link lost native read-only semantics")
	}
	replace("Link", "rejected edit")
	if p.draft().link != "https://ui.shadcn.com/docs/installation" {
		t.Fatal("readonly share source accepted editing")
	}
	click(ui.RoleButton, "Close")
	r.Frame()
	r.Frame()
	if !access(ui.RoleButton, "Share").Focused {
		t.Fatal("custom footer Close did not restore Share opener")
	}
	choose("no-close")
	open("No Close Button", "No Close Button")
	if _, ok := r.AccessByRole(ui.RoleButton, "Close dialog"); ok {
		t.Fatal("source showCloseButton=false exposed a corner close")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Close"); ok {
		t.Fatal("docs no-close example invented the Create-only footer")
	}
	r.ClickAt(3, 3)
	r.Frame()
	r.Frame()
	if p.draft().open {
		t.Fatal("no-close source incorrectly disabled normal backdrop dismissal")
	}
	choose("create-no-close")
	open("No Close Button", "No Close Button")
	access(ui.RoleButton, "Close")
	if _, ok := r.AccessByRole(ui.RoleButton, "Close dialog"); ok {
		t.Fatal("Create no-close example gained a corner close")
	}
	click(ui.RoleButton, "Close")
	r.Frame()
	r.Frame()
	choose("sticky")
	open("Sticky Footer", "Sticky Footer")
	scroll := access(ui.RoleRegion, "Dialog source scroll")
	footer := access(ui.RoleButton, "Close")
	heading := access(ui.RoleHeading, "Sticky Footer")
	if scroll.Bounds.H > 451 || footer.Bounds.Y < scroll.Bounds.Y+scroll.Bounds.H {
		t.Fatal("docs50vh scroll viewport or pinned footer was lost")
	}
	r.Scroll(scroll.Bounds.X+scroll.Bounds.W/2, scroll.Bounds.Y+scroll.Bounds.H/2, 0, 9000)
	r.Frame()
	if access(ui.RoleButton, "Close").Bounds.Y != footer.Bounds.Y || access(ui.RoleHeading, "Sticky Footer").Bounds.Y != heading.Bounds.Y {
		t.Fatal("long source content scrolled the fixed header/footer")
	}
	var paragraphs []ui.AccessInfo
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleText && a.Name == dialogSourceLorem {
			paragraphs = append(paragraphs, a)
		}
	}
	if len(paragraphs) != 10 || paragraphs[9].Offscreen {
		t.Fatal("source ten full paragraphs were truncated or unreachable")
	}
	dark = true
	r.Frame()
	dialogCompositionSnapshot(t, r, "dialog-sticky-dark")
	click(ui.RoleButton, "Close")
	r.Frame()
	r.Frame()
	choose("create-sticky")
	open("Sticky Footer", "Scrollable Content")
	if a := access(ui.RoleRegion, "Dialog source scroll"); a.Bounds.H <= scroll.Bounds.H {
		t.Fatal("Create70vh body was collapsed to the docs50vh viewport")
	}
	closeModal()
	choose("create-chat")
	r.SetSize(1000, 1100)
	r.Frame()
	open("Chat Settings", "Chat Settings")
	if p.draft().chatTheme != "system" || p.draft().accent != "default" || p.draft().spoken != "en" || p.draft().voice != "samantha" || len(dialogChatLanguages) != 18 || len(dialogChatVoices) != 8 {
		t.Fatal("source complete settings options/defaults were collapsed")
	}
	access(ui.RoleTabList, "Chat settings tabs")
	for _, choice := range []struct{ label, item string }{{"Theme", "Dark"}, {"Accent Color", "Purple"}, {"Spoken Language", "Malayalam"}, {"Voice", "Daniel"}} {
		click(ui.RoleMenuButton, choice.label)
		if err := r.ChooseMenuItem(choice.item); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	if p.draft().chatTheme != "dark" || p.draft().accent != "purple" || p.draft().spoken != "ml" || p.draft().voice != "daniel" {
		t.Fatal("real chat settings selections did not share the controlled source model")
	}
	click(ui.RoleTab, "Notifications")
	var pushes []ui.AccessInfo
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleCheckBox && a.Name == "Push notifications" {
			pushes = append(pushes, a)
		}
	}
	if len(pushes) != 2 || !pushes[0].Disabled || !pushes[0].Checked || pushes[1].Disabled || pushes[1].Checked {
		t.Fatal("Responses disabled checked default and independent Tasks checkbox were lost")
	}
	r.ClickAt(pushes[1].Bounds.X+pushes[1].Bounds.W/2, pushes[1].Bounds.Y+pushes[1].Bounds.H/2)
	r.Frame()
	click(ui.RoleCheckBox, "Email notifications")
	if !p.draft().taskPush || !p.draft().taskEmail {
		t.Fatal("Tasks native checkbox states were not independently editable")
	}
	click(ui.RoleLink, "Manage tasks")
	if p.action != "Local dialog action: Manage tasks" || len(r.OpenedURLs()) != 0 {
		t.Fatal("source hash link escaped local routing or failed its real intent")
	}
	click(ui.RoleTab, "Personalization")
	replace("Nickname", "Broski Native")
	replace("More about you", "I build native apps.\nKeep the source multiline draft.")
	if !access(ui.RoleTextField, "More about you").Multiline || access(ui.RoleTextField, "More about you").Bounds.H < 120 || !p.draft().custom {
		t.Fatal("source personalization textarea/customization default was lost")
	}
	click(ui.RoleSwitch, "Enable customizations")
	if p.draft().custom {
		t.Fatal("source customizations real switch failed")
	}
	customCaptions := 0
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleText && a.Name == "Enable customizations" {
			customCaptions++
		}
	}
	if customCaptions != 1 {
		t.Fatal("the Field control invented a duplicate visible Switch caption")
	}
	info := access(ui.RoleButton, "Nickname help")
	if !strings.Contains(info.Description, "Used to identify you in the chat.") {
		t.Fatal("source rich tooltip failed the actual button description")
	}
	r.Move(info.Bounds.X+info.Bounds.W/2, info.Bounds.Y+info.Bounds.H/2)
	r.Frame()
	r.Advance(time.Millisecond)
	if len(b.surfaces) == 0 {
		t.Fatal("nickname help did not open its actual independent tooltip")
	}
	tip := b.surfaces[len(b.surfaces)-1]
	if !tip.spec.IgnoreMouseEvents || !tip.tester.HasText("Used to identify you in the chat.") || !tip.tester.HasText("N") {
		t.Fatal("source rich tooltip lost pointer transparency or original Kbd content")
	}
	r.Move(1, 1)
	r.Frame()
	r.Advance(time.Millisecond)
	dialogCompositionSnapshot(t, r, "dialog-chat-personalization-dark")
	click(ui.RoleTab, "Security")
	click(ui.RoleSwitch, "Multi-factor authentication")
	click(ui.RoleButton, "Log Out")
	click(ui.RoleButton, "Log Out All")
	if !p.draft().twoFactor || p.action != "Local dialog action: Log Out All" || !p.draft().open {
		t.Fatal("source security controls lost local intent or invented account/dismiss behavior")
	}
	r.SetSize(360, 780)
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleTabList, "Chat settings tabs"); ok {
		t.Fatal("narrow chat settings retained the desktop tablist")
	}
	click(ui.RoleMenuButton, "Chat settings section")
	if err := r.ChooseMenuItem("Personalization"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.draft().tab != "personalization" || p.draft().nickname != "Broski Native" || !strings.Contains(p.draft().about, "source multiline") {
		t.Fatal("responsive native section selection lost personalized drafts")
	}
	click(ui.RoleButton, "Close dialog")
	r.Frame()
	r.Frame()
	choose("rtl")
	click(ui.RoleMenuButton, "Dialog language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	open("פתח דיאלוג", "ערוך פרופיל")
	replace("שם", "פרופיל מקומי")
	modal := access(ui.RoleDialog, "ערוך פרופיל")
	if modal.Bounds.W > 328.5 {
		t.Fatal("native RTL dialog overflowed its narrow window")
	}
	corner := access(ui.RoleButton, "Close dialog")
	if corner.Bounds.W != 32 || corner.Bounds.X+corner.Bounds.W/2 >= modal.Bounds.X+modal.Bounds.W/2 {
		t.Fatal("the real source close target overlapped the RTL title instead of its logical end")
	}
	if access(ui.RoleButton, "שמור שינויים").Bounds.W < modal.Bounds.W-50.5 || access(ui.RoleButton, "בטל").Bounds.W < modal.Bounds.W-50.5 {
		t.Fatal("narrow source footer buttons did not stretch in their reversed column")
	}
	dialogCompositionSnapshot(t, r, "dialog-hebrew-narrow-dark")
	closeModal()
	visible = false
	r.Frame()
	visible = true
	r.Frame()
	native = false
	r.SetSize(760, 900)
	r.Frame()
	chooseStyled("demo")
	open("Open Dialog", "Edit profile")
	if access(ui.RoleTextField, "Name").Value != "Local source profile" || p.draft().saves != 1 || p.draft().saved["username"] != "@native-profile" {
		t.Fatal("page/style/recipe changes lost independent source drafts/snapshots")
	}
	closeModal()
}
