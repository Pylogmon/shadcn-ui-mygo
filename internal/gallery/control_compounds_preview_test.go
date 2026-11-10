package gallery

import (
	"errors"
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

func compoundAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, label)
	if !ok || a.Offscreen || a.Bounds.W <= 0 || a.Bounds.H <= 0 {
		t.Fatalf("missing visible compound control %q role=%v", label, role)
	}
	return a
}

func compoundClick(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	a, ok := r.Access(label)
	if !ok || a.Offscreen || a.Bounds.W <= 0 || a.Bounds.H <= 0 {
		t.Fatalf("missing compound action %q", label)
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
}

func compoundType(t *testing.T, r *ui.Tester, label, value string) {
	t.Helper()
	compoundClick(t, r, label)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type(value)
	r.Frame()
}

func compoundSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_CONTROL_COMPOUND_SNAPSHOT_DIR")
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
		t.Fatalf("compound snapshot: %v/%v", err, closeErr)
	}
}

// One application flow traverses existing native controls and the source
// compositions. This local owner/page host is not the complete State.View.
func TestControlCompoundsSourceWorkflowAndPageDrafts(t *testing.T) {
	p := &controlCompoundsPreviewState{}
	b := &creatorPopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	page := "groups"
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(24).Gap(18).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Show groups").Clicked() {
					visible, page = true, "groups"
				}
				if ui.Button(c, "Show avatar").Clicked() {
					visible, page = true, "avatar"
				}
				if ui.Button(c, "Leave compounds").Clicked() {
					visible = false
				}
			})
			if !visible {
				ui.Text(c, "Another application page")
			} else if page == "avatar" {
				p.avatar(c, h, native)
			} else {
				p.buttonGroups(c, h, native)
			}
		})
	}, 1050, 850, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	configure := func(label string) {
		t.Helper()
		compoundClick(t, r, "Button group configuration")
		if err := r.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	if p.error != "" || p.label != "personal" || !r.HasText("Current label: personal") {
		t.Fatal("source mail defaults missing or inactive menu reported an error")
	}
	compoundClick(t, r, "Mail options")
	if err := r.ChooseMenuItem("Label As...", "Work"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.label != "work" || p.action != "Local mail action: work" {
		t.Fatal("native nested radio choice did not update the source label model")
	}

	configure("Voice input group")
	compoundType(t, r, "Grouped message", "Ship the native UI")
	compoundClick(t, r, "Voice Mode")
	if !p.voice || !compoundAccess(t, r, ui.RoleTextField, "Grouped message").Disabled || p.message != "Ship the native UI" {
		t.Fatal("voice mode did not disable the real editor while retaining its draft")
	}
	compoundClick(t, r, "Voice Mode")
	if compoundAccess(t, r, ui.RoleTextField, "Grouped message").Value != "Ship the native UI" {
		t.Fatal("restoring text mode lost the native input's draft")
	}
	configure("Copy and paste")
	r.SetClipboard("unchanged")
	r.FailClipboard(errors.New("clipboard refusal"))
	compoundClick(t, r, "Copy grouped draft")
	if r.Clipboard() != "unchanged" || p.copied != "" || !strings.Contains(p.action, "Could not copy") {
		t.Fatal("copy refusal was reported as success")
	}
	r.FailClipboard(nil)
	compoundClick(t, r, "Copy grouped draft")
	if r.Clipboard() != "Ship the native UI" || p.copied != "Ship the native UI" {
		t.Fatal("the acknowledged copy did not use the actual grouped draft")
	}
	r.SetClipboard("Pasted native draft")
	compoundClick(t, r, "Paste grouped draft")
	configure("Nested message input")
	if compoundAccess(t, r, ui.RoleTextField, "Grouped message").Value != "Pasted native draft" {
		t.Fatal("nested and pill source configurations did not share their application draft")
	}

	configure("Currency and amount")
	compoundClick(t, r, "Grouped currency")
	if err := r.ChooseMenuItem("€ Euro"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	compoundType(t, r, "Grouped amount", "12.50")
	compoundClick(t, r, "Preview grouped amount")
	if p.currency != "€" || p.amount != "12.50" || p.action != "Local amount preview: €12.50" {
		t.Fatal("the joined currency/editor/action did not read one current model")
	}
	compoundSnapshot(t, r, "control-compounds-amount-light")
	configure("Search input")
	compoundType(t, r, "Grouped search", "native controls")
	compoundClick(t, r, "Search grouped text")
	if p.action != "Local search: native controls" {
		t.Fatal("joined search action did not read the native editor")
	}
	configure("Split button")
	compoundClick(t, r, "Add grouped item")
	if p.additions != 1 {
		t.Fatal("source split plus action was not independently clickable")
	}
	configure("Vertical controls")
	compoundClick(t, r, "Increase media level")
	compoundClick(t, r, "Decrease media level")
	if p.mediaLevel != 50 || !compoundAccess(t, r, ui.RoleGroup, "Media controls").Vertical {
		t.Fatal("vertical native group lost its actual actions or orientation")
	}
	configure("Small, default and large")
	for _, label := range []string{"Small button group", "Default button group", "Large button group"} {
		compoundAccess(t, r, ui.RoleGroup, label)
	}
	configure("Text and label")
	compoundType(t, r, "Grouped name", "Taylor")
	compoundClick(t, r, "Show groups")
	compoundClick(t, r, "Name")
	if !r.Focused("Grouped name") {
		t.Fatal("the native grouped label did not focus its editor")
	}

	configure("Task popover")
	compoundClick(t, r, "Open grouped task")
	if !h.IsOpen("compound-task") || len(b.surfaces) == 0 {
		t.Fatal("source task did not open a real independent child surface")
	}
	child := b.surfaces[len(b.surfaces)-1]
	if !child.spec.SystemPopover || r.HasText("Start a new task with Copilot") {
		t.Fatal("native task container was not separate from the parent tree")
	}
	compoundType(t, child.tester, "Task Description", "Keep the task draft locally.")
	if p.task != "Keep the task draft locally." {
		t.Fatal("the native child editor did not update the application task")
	}
	// Native Escape/outside dismissal belongs to the window backend. Deliver
	// that backend's actual closed callback; Tester has no NSPopover tracking.
	child.Close()
	r.Frame()
	if h.IsOpen("compound-task") {
		t.Fatal("the native closed callback left the grouped task window open")
	}
	configure("Follow with menu")
	compoundClick(t, r, "Follow conversation")
	compoundClick(t, r, "Conversation options")
	if err := r.ChooseMenuItem("Mute Conversation"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !p.following || p.action != "Local conversation action: mute" {
		t.Fatal("source follow action and popup action were not independent")
	}

	configure("Translated RTL actions")
	if !r.HasText("أرشفة") || !r.HasText("تأجيل") {
		t.Fatal("the fixed RTL configuration did not use its actual translated labels")
	}
	compoundClick(t, r, "Button group language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !r.HasText("ארכיון") || p.label != "work" {
		t.Fatal("language change lost the source radio state")
	}
	dark = true
	r.SetSize(500, 850)
	r.SetDirection(ui.RightToLeft)
	compoundSnapshot(t, r, "control-compounds-mail-dark-rtl")
	compoundClick(t, r, "Leave compounds")
	compoundClick(t, r, "Show groups")
	if p.task != "Keep the task draft locally." || p.message != "Pasted native draft" || p.name != "Taylor" || p.currency != "€" || p.amount != "12.50" {
		t.Fatal("leaving the owner page reset a source configuration's application draft")
	}

	compoundClick(t, r, "Show avatar")
	compoundClick(t, r, "Open avatar menu")
	if err := r.ChooseMenuItem("Profile"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.avatarAction != "Local account preview: profile" {
		t.Fatal("native avatar menu did not return its actual selection")
	}
	native = false
	r.Frame()
	before := len(b.surfaces)
	compoundClick(t, r, "Open avatar menu")
	if len(b.surfaces) != before+1 {
		t.Fatal("styled avatar menu did not create its own native Flyout")
	}
	child = b.surfaces[len(b.surfaces)-1]
	child.tester.Advance(time.Second)
	compoundAccess(t, child.tester, ui.RoleMenuItem, "Settings")
	if err := child.tester.Click("Settings"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.avatarAction != "Local account preview: settings" || p.error != "" {
		t.Fatalf("styled avatar selection was stale: action=%q error=%q", p.avatarAction, p.error)
	}
	compoundSnapshot(t, r, "control-compounds-avatar-dark-rtl")
}
