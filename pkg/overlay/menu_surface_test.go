package overlay

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestInWindowMenuThemeWrappersPreserveTranslucentBackdrop(t *testing.T) {
	for _, family := range []string{"dropdown", "context", "submenu"} {
		t.Run(family, func(t *testing.T) {
			open := false
			items := []MenuItem{{Key: "leaf", Label: "Leaf"}}
			if family == "submenu" {
				items = []MenuItem{{Key: "more", Label: "More", Children: items}}
			}
			r := ui.NewTester(func(c *ui.Context) {
				c.Root().Background(ui.RGB(0, 70, 220))
				c.SetMenuPresentation(ui.PopoverMenus)
				theme.MenuScope(c, theme.Preset("neutral", true), theme.MenuOptions{Color: theme.MenuDefaultTranslucent}).Background(ui.Transparent).Fill().Padding(20).Children(func() {
					if family == "context" {
						target := ui.Box(c).Size(120, 32).Focusable().Label("Document")
						InWindowContextMenu(c, target, MenuProps{Items: items})
					} else {
						InWindowDropdownMenu(c, &open, MenuProps{Items: items}, func(b ui.Element) { button(c, b, "Actions", false) })
					}
				})
			}, 600, 300)
			if family == "context" {
				if err := r.RightClick("Document"); err != nil {
					t.Fatal(err)
				}
			} else if err := r.Click("Actions"); err != nil {
				t.Fatal(err)
			}
			if family == "submenu" {
				r.Key(0, ui.KeyDown)
				r.Key(0, ui.KeyRight)
			}
			item, ok := r.AccessByRole(ui.RoleMenuItem, "Leaf")
			if !ok {
				t.Fatal("menu content missing")
			}
			pixel := r.Image().RGBAAt(int(item.Bounds.X+item.Bounds.W-24), int(item.Bounds.Y+item.Bounds.H/2))
			if int(pixel.B)-int(pixel.R) < 20 {
				t.Fatalf("theme wrapper covered the blue backdrop: %v", pixel)
			}
		})
	}
}
