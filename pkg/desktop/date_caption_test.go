package desktop

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

func TestNativeDatePickerCaptionUsesSystemMenus(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	value := desktopCivil(2026, time.November, 9)
	initial := value
	parent := ui.NewTester(func(c *ui.Context) {
		h.BeginFrame()
		defer h.EndFrame()
		_, err := NativeDatePicker(c, h, "native-caption", &value, DatePickerProps{
			Label:    "Native date",
			Calendar: display.CalendarProps{Caption: display.CalendarCaptionDropdown},
		})
		if err != nil {
			t.Fatal(err)
		}
	}, 420, 160)
	if err := parent.Click("Native date"); err != nil {
		t.Fatal(err)
	}
	child := pickerChild(t, b)
	pickerClickRole(t, child, ui.RoleMenuButton, "Month")
	if !slices.Contains(child.Menu(), "December") || child.HasText("January") {
		t.Fatalf("caption did not use an independent system menu: menu=%v texts=%v", child.Menu(), child.Texts())
	}
	if err := child.ChooseMenuItem("December"); err != nil {
		t.Fatal(err)
	}
	if value != initial {
		t.Fatalf("browsing committed a date: %v", value)
	}
	month, ok := child.AccessByRole(ui.RoleMenuButton, "Month")
	if !ok || month.Value != "December" {
		t.Fatalf("month did not change: %+v", month)
	}
	pickerClickRole(t, child, ui.RoleMenuButton, "Year")
	if !slices.Contains(child.Menu(), "2027") {
		t.Fatalf("year did not use the system menu: %v", child.Menu())
	}
	if err := child.ChooseMenuItem("2027"); err != nil {
		t.Fatal(err)
	}
	if value != initial {
		t.Fatalf("year browsing committed a date: %v", value)
	}
}
