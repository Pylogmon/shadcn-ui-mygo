package overlay

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func view(fn func(*ui.Context)) func(*ui.Context) {
	return func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(24).Gap(12).AlignItems(ui.Start).Children(func() { fn(c) })
	}
}

func TestDropdownKeyboardSkipsDisabledAndRestoresFocus(t *testing.T) {
	open, checked := false, false
	chosen := ""
	tt := ui.NewTester(view(func(c *ui.Context) {
		value := DropdownMenu(c, &open, MenuProps{Items: []MenuItem{{Key: "disabled", Label: "Unavailable", Disabled: true}, {Key: "pin", Label: "Pinned", Checked: &checked}, {Key: "sort", Label: "Sort", Children: []MenuItem{{Key: "name", Label: "Name"}, {Key: "date", Label: "Date"}}}}}, func(b ui.Element) { button(c, b, "Actions", false) })
		if value != "" {
			chosen = value
		}
	}), 480, 400)
	if err := tt.Click("Actions"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if !checked || chosen != "pin" || open {
		t.Fatalf("checked=%v, chosen=%q, open=%v", checked, chosen, open)
	}
	if !tt.Focused("Actions") {
		t.Fatal("focus did not return to the trigger")
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if chosen != "date" || open {
		t.Fatalf("nested menu chose %q, open=%v", chosen, open)
	}
}

func TestContextMenuSecondaryClick(t *testing.T) {
	chosen := ""
	tt := ui.NewTester(view(func(c *ui.Context) {
		target := ui.Box(c).Padding(16).Focusable().Label("Document").Children(func() { ui.Text(c, "Document") })
		if key := ContextMenu(c, target, MenuProps{Items: []MenuItem{{Key: "rename", Label: "Rename"}}}); key != "" {
			chosen = key
		}
	}), 400, 300)
	if err := tt.RightClick("Document"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Rename"); err != nil {
		t.Fatal(err)
	}
	if chosen != "rename" {
		t.Fatalf("chose %q", chosen)
	}
}

func TestDialogEscapeAndFocusRestoration(t *testing.T) {
	open := false
	tt := ui.NewTester(view(func(c *ui.Context) {
		if ui.Button(c, "Open dialog").Clicked() {
			open = true
		}
		Dialog(c, &open, DialogProps{Title: "Edit profile", Description: "Your public details"}, func() { ui.Button(c, "Save profile") })
	}), 640, 480)
	if err := tt.Click("Open dialog"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	if info, ok := tt.AccessByRole(ui.RoleDialog, "Edit profile"); !ok || info.Description != "Your public details" {
		t.Fatalf("dialog accessibility=%+v, found=%v", info, ok)
	}
	if err := tt.Click("Save profile"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEscape)
	if open {
		t.Fatal("Escape did not close")
	}
	tt.Advance(time.Second)
	if _, ok := tt.AccessByRole(ui.RoleDialog, "Edit profile"); ok {
		t.Fatal("dialog remained after exit")
	}
	if !tt.Focused("Open dialog") {
		t.Fatal("focus did not return to opening button")
	}
}

func TestAlertDialogRequiresExplicitResponse(t *testing.T) {
	open := true
	tt := ui.NewTester(view(func(c *ui.Context) {
		AlertDialog(c, &open, AlertDialogProps{Title: "Delete account?", Description: "This cannot be undone", ActionLabel: "Delete"})
	}), 640, 480)
	tt.Advance(time.Second)
	if _, ok := tt.AccessByRole(ui.RoleAlertDialog, "Delete account?"); !ok {
		t.Fatal("missing alert-dialog role")
	}
	tt.ClickAt(3, 3)
	if !open {
		t.Fatal("backdrop dismissed alert dialog")
	}
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	if open {
		t.Fatal("Cancel did not close alert dialog")
	}
}

func TestCommandFiltersKeywordsAndSkipsDisabled(t *testing.T) {
	query, chosen := "", ""
	tt := ui.NewTester(view(func(c *ui.Context) {
		if key := Command(c, &query, CommandProps{Width: 360, AutoFocus: true, Items: []CommandItem{{Value: "locked", Label: "Print locked", Disabled: true}, {Value: "print", Label: "Print document", Keywords: []string{"paper"}}, {Value: "settings", Label: "Settings"}}}); key != "" {
			chosen = key
		}
	}), 480, 400)
	tt.Type("paper")
	if !tt.HasText("Print document") || tt.HasText("Settings") {
		t.Fatalf("wrong filtered results: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	if chosen != "print" {
		t.Fatalf("Enter chose %q", chosen)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("nonsense")
	if !tt.HasText("No results found.") {
		t.Fatal("missing empty state")
	}
}

func TestToastActionAndClose(t *testing.T) {
	undo := false
	tt := ui.NewTester(view(func(c *ui.Context) {
		if ui.Button(c, "Notify").Clicked() {
			Toast(c, ui.Toast{Title: "Deleted", Action: "Undo", OnAction: func() { undo = true }, Timeout: -1})
		}
		Toaster(c)
	}), 640, 480)
	if err := tt.Click("Notify"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	if !undo {
		t.Fatal("toast action did not execute")
	}
	tt.Advance(time.Second)
	if tt.HasText("Deleted") {
		t.Fatal("action did not dismiss toast")
	}
}
