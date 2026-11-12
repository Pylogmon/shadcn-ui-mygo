package desktop

import (
	"runtime"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestWindowsStyledSurfacesKeepPaintedBordersDuringExit(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows styled flyouts replace native window outlines")
	}
	for _, dark := range []bool{false, true} {
		for _, family := range []string{"popover", "hover-card", "menu", "combobox"} {
			t.Run(family+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				backend := &presenceBackend{}
				host := NewHostWithBackend(backend)
				open, state := true, MenuState{}
				combo := control.ComboboxState[string]{Open: true}
				parent := ui.NewTester(func(c *ui.Context) {
					c.SetTheme(theme.Preset("neutral", dark).Theme())
					anchor := control.Button(c, "Open", control.ButtonProps{})
					var err error
					switch family {
					case "popover":
						err = Popover(c, host, "panel", anchor, &open, PopoverProps{Label: "Panel"}, func(child *ui.Context) { ui.Text(child, "Content") })
					case "hover-card":
						err = HoverCard(c, host, "panel", anchor, HoverCardProps{Label: "Panel"}, func(child *ui.Context) { ui.Text(child, "Content") })
					case "menu":
						_, err = Menu(c, host, "panel", anchor, &open, &state, MenuProps{Label: "Panel", Items: []overlay.MenuItem{{Key: "one", Label: "One"}}})
					case "combobox":
						_, err = Combobox(c, host, "panel", &combo, ComboboxProps{Label: "Panel"}, func(*ui.Context) control.ComboboxRootProps[string] {
							return control.ComboboxRootProps[string]{Items: []string{"One"}, Adapter: control.ComboboxAdapter[string]{Key: func(value string) string { return value }}}
						})
					}
					if err != nil {
						t.Fatal(err)
					}
				}, 400, 200)
				if family == "hover-card" {
					parent.Key(0, ui.KeyTab)
				}
				if len(backend.surfaces) != 1 {
					t.Fatalf("opened %d surfaces", len(backend.surfaces))
				}
				child := ui.NewTester(backend.surfaces[0].view, 320, 200)
				child.Advance(300 * time.Millisecond)
				role, label := ui.RoleDialog, "Panel"
				switch family {
				case "hover-card":
					role = ui.RoleGroup
				case "menu":
					role = ui.RoleMenu
				case "combobox":
					role, label = ui.RoleGroup, "Panel content"
				}
				panel, ok := child.AccessByRole(role, label)
				if !ok {
					t.Fatal("painted panel missing")
				}
				x, y := int(panel.Bounds.X+panel.Bounds.W/2), int(panel.Bounds.Y)
				if family == "hover-card" {
					// The accessible preview group is inside the host's panel padding.
					x, y = child.Image().Bounds().Dx()/2, 0
				}
				edge, inside := child.Image().RGBAAt(x, y), child.Image().RGBAAt(x, y+3)
				if edge.A != 255 || inside.A != 255 || edge == inside {
					t.Fatalf("panel blends into its surface without a painted border: edge=%v inside=%v", edge, inside)
				}
				shown := paintedAlpha(child.Image())
				host.Close("panel")
				child.Frame()
				child.Advance(50 * time.Millisecond)
				if leaving := paintedAlpha(child.Image()); leaving == 0 || leaving >= shown {
					t.Fatal("painted border and surface did not fade together")
				}
				child.Advance(150 * time.Millisecond)
				if paintedAlpha(child.Image()) != 0 {
					t.Fatal("border remained after its panel disappeared")
				}
			})
		}
	}
}
