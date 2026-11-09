package overlay

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestNativeDropdownSubmenuCheckedAndDisabled(t *testing.T) {
	checked := false
	chosen := ""
	items := []MenuItem{{Key: "locked", Label: "Locked", Disabled: true}, {Key: "pin", Label: "Pinned", Checked: &checked}, {Key: "export", Label: "Export", Children: []MenuItem{{Key: "pdf", Label: "PDF"}}}, {Separator: true}, {Key: "remove", Label: "Remove", Destructive: true}}
	r := ui.NewTester(view(func(c *ui.Context) {
		if v := NativeDropdownMenu(c, MenuProps{Items: items}, func(b ui.Element) { button(c, b, "Native actions", false) }); v != "" {
			chosen = v
		}
	}), 220, 120)
	if err := r.Click("Native actions"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(r.Menu(), "Export") {
		t.Fatalf("native menu %v", r.Menu())
	}
	if r.HasText("Pinned") {
		t.Fatal("system menu was rendered into the parent's UI tree")
	}
	if err := r.ChooseMenuItem("Locked"); err == nil {
		t.Fatal("disabled native action accepted")
	}
	if err := r.ChooseMenuItem("Pinned"); err != nil {
		t.Fatal(err)
	}
	if !checked || chosen != "pin" {
		t.Fatalf("checked=%v chosen=%q", checked, chosen)
	}
	if err := r.Click("Native actions"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("Export", "PDF"); err != nil {
		t.Fatal(err)
	}
	if chosen != "pdf" {
		t.Fatal(chosen)
	}
}
func TestNativeContextAndMenubar(t *testing.T) {
	chosen := ""
	r := ui.NewTester(view(func(c *ui.Context) {
		b := ui.Box(c).Size(120, 32).Focusable().Label("Document")
		if v := NativeContextMenu(c, b, MenuProps{Items: []MenuItem{{Key: "rename", Label: "Rename"}}}); v != "" {
			chosen = v
		}
		if v := NativeMenubar(c, []MenubarItem{{Label: "File", Items: []MenuItem{{Key: "new", Label: "New document"}}}}); v != "" {
			chosen = v
		}
	}), 360, 200)
	if err := r.RightClick("Document"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("Rename"); err != nil {
		t.Fatal(err)
	}
	if chosen != "rename" {
		t.Fatal(chosen)
	}
	if err := r.Click("File"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("New document"); err != nil {
		t.Fatal(err)
	}
	if chosen != "new" {
		t.Fatal(chosen)
	}
}
