package desktop

import (
	"image"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func paintedAlpha(img *image.RGBA) uint64 {
	var total uint64
	for i := 3; i < len(img.Pix); i += 4 {
		total += uint64(img.Pix[i])
	}
	return total
}

// Exercise the public entry points: a transparent native window must contain
// only the animated surface, including on its first and final render passes.
func TestFlyoutFamiliesCompositeOnlyTheirVisibleSurface(t *testing.T) {
	for _, family := range []string{"popover", "menu", "dropdown", "context", "menubar", "select", "combobox", "date", "date-range", "tooltip", "hover-card"} {
		t.Run(family, func(t *testing.T) {
			backend := &presenceBackend{}
			host := NewHostWithBackend(backend)
			open := true
			menu := MenuState{}
			bar := MenubarState{Open: true, Active: "file"}
			combo := control.ComboboxState[string]{Open: true}
			value := "one"
			date := desktopCivil(2026, 11, 9)
			dates := display.DateRange{}
			items := []overlay.MenuItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}}
			parent := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(theme.Preset("neutral", true).Theme())
				host.BeginFrame()
				defer host.EndFrame()
				var err error
				switch family {
				case "select":
					_, err = Select(c, host, "surface", &value, control.SelectProps{Label: "Open", Options: []control.Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}}})
				case "combobox":
					_, err = Combobox(c, host, "surface", &combo, ComboboxProps{Label: "Open"}, func(*ui.Context) control.ComboboxRootProps[string] {
						return control.ComboboxRootProps[string]{Items: []string{"One", "Two"}, Adapter: control.ComboboxAdapter[string]{Key: func(v string) string { return v }}}
					})
				case "date":
					_, err = DatePicker(c, host, "surface", &date, DatePickerProps{Label: "Open"})
				case "date-range":
					_, err = DateRangePicker(c, host, "surface", &dates, DatePickerProps{Label: "Open"})
				case "menubar":
					_, err = Menubar(c, host, "surface", &bar, MenubarProps{Items: []overlay.MenubarItem{{Key: "file", Label: "File", Items: items}}})
				default:
					anchor := control.Button(c, "Open", control.ButtonProps{})
					switch family {
					case "popover":
						err = Popover(c, host, "surface", anchor, &open, PopoverProps{}, func(child *ui.Context) { ui.Text(child, "Content") })
					case "menu":
						_, err = Menu(c, host, "surface", anchor, &open, &menu, MenuProps{Items: items})
					case "dropdown":
						_, err = DropdownMenu(c, host, "surface", anchor, &open, &menu, MenuProps{Items: items})
					case "context":
						_, err = ContextMenu(c, host, "surface", anchor, &open, &menu, MenuProps{Items: items})
					case "tooltip":
						err = Tooltip(c, host, "surface", anchor, "Content", TooltipProps{})
					case "hover-card":
						err = HoverCard(c, host, "surface", anchor, HoverCardProps{}, func(child *ui.Context) { ui.Text(child, "Content") })
					}
				}
				if err != nil {
					t.Fatal(err)
				}
			}, 800, 200)
			switch family {
			case "select", "date", "date-range":
				if err := parent.Click("Open"); err != nil {
					t.Fatal(err)
				}
			case "tooltip", "hover-card":
				parent.Key(0, ui.KeyTab)
			}
			if len(backend.surfaces) != 1 {
				t.Fatalf("opened %d surfaces", len(backend.surfaces))
			}
			surface := backend.surfaces[0]
			child := ui.NewTester(surface.view, 680, 420)
			if alpha := paintedAlpha(child.Image()); alpha != 0 {
				t.Errorf("invisible entry painted an opaque backing: total alpha=%d", alpha)
			}
			child.Advance(30 * time.Millisecond)
			enter := paintedAlpha(child.Image())
			child.Advance(250 * time.Millisecond)
			shown := paintedAlpha(child.Image())
			if enter == 0 || enter >= shown || child.FramePending() {
				t.Errorf("entry did not progress and settle: entering=%d shown=%d pending=%v", enter, shown, child.FramePending())
			}
			host.Close("surface")
			child.Frame()
			child.Advance(35 * time.Millisecond)
			leaving := paintedAlpha(child.Image())
			if leaving == 0 || leaving >= shown || surface.closes != 0 {
				t.Errorf("exit did not fade the whole surface: leaving=%d shown=%d closes=%d", leaving, shown, surface.closes)
			}
			child.Advance(150 * time.Millisecond)
			if alpha := paintedAlpha(child.Image()); alpha != 0 || surface.closes != 1 || host.IsPresent("surface") {
				t.Errorf("exit completion flashed pixels: alpha=%d closes=%d", alpha, surface.closes)
			}
			// Native destruction is deferred until after painting. A queued render
			// from an obsolete generation must remain clear during that interval.
			child.Frame()
			if alpha := paintedAlpha(child.Image()); alpha != 0 || child.FramePending() {
				t.Errorf("obsolete child flashed or kept rendering: alpha=%d pending=%v", alpha, child.FramePending())
			}
		})
	}
}

func TestNativePopoverMaterialHasNoPaintedBacking(t *testing.T) {
	for _, dark := range []bool{false, true} {
		backend := &fakeBackend{}
		host := NewHostWithBackend(backend)
		open := true
		ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			anchor := control.Button(c, "Open", control.ButtonProps{})
			if err := NativePopover(c, host, "native", anchor, &open, PopoverProps{}, nil); err != nil {
				t.Fatal(err)
			}
		}, 400, 200)
		child := ui.NewTesterWith(backend.surfaces[0].view, 288, 200, ui.TesterOptions{Dark: dark})
		if alpha := child.Image().RGBAAt(20, 20).A; alpha != 255 {
			t.Fatalf("platform fallback has no background: alpha=%d", alpha)
		}
		child.SetVibrancy(true)
		if alpha := paintedAlpha(child.Image()); alpha != 0 {
			t.Errorf("system material is covered in dark=%v: total alpha=%d", dark, alpha)
		}
	}
}

func TestFlyoutEntryKeepsItsFirstContentTree(t *testing.T) {
	backend := &presenceBackend{}
	host := NewHostWithBackend(backend)
	open := true
	mounts := 0
	draft := "Draft"
	ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		anchor := control.Button(c, "Open", control.ButtonProps{})
		_ = Popover(c, host, "editor", anchor, &open, PopoverProps{FitContent: true}, func(child *ui.Context) {
			editor := ui.TextInput(child, &draft).Label("Draft")
			ui.Local(editor, "editor-lifetime", func() bool { mounts++; return true })
		})
	}, 400, 200)
	child := ui.NewTester(backend.surfaces[0].view, 288, 200)
	child.Advance(300 * time.Millisecond)
	if mounts != 1 {
		t.Fatalf("entry replaced a mounted native editor %d times", mounts-1)
	}
}

func TestNativeFlyoutFamiliesKeepThemeWrappersClear(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, family := range []string{"date", "date-range", "combobox", "hover-card", "tooltip"} {
			t.Run(family+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				backend := &presenceBackend{}
				host := NewHostWithBackend(backend)
				date := desktopCivil(2026, 11, 9)
				dates := display.DateRange{}
				combo := control.ComboboxState[string]{Open: true}
				parent := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					host.BeginFrame()
					defer host.EndFrame()
					var err error
					switch family {
					case "date":
						_, err = NativeDatePicker(c, host, "native", &date, DatePickerProps{Label: "Open"})
					case "date-range":
						_, err = NativeDateRangePicker(c, host, "native", &dates, DatePickerProps{Label: "Open"})
					case "combobox":
						_, err = NativeCombobox(c, host, "native", &combo, ComboboxProps{Label: "Open"}, func(*ui.Context) control.ComboboxRootProps[string] {
							return control.ComboboxRootProps[string]{Items: []string{"One", "Two"}, Adapter: control.ComboboxAdapter[string]{Key: func(v string) string { return v }}}
						})
					default:
						anchor := control.Button(c, "Open", control.ButtonProps{})
						if family == "tooltip" {
							err = NativeTooltip(c, host, "native", anchor, "Content", TooltipProps{})
						} else {
							err = NativeHoverCard(c, host, "native", anchor, HoverCardProps{}, func(child *ui.Context) { ui.Text(child, "Content") })
						}
					}
					if err != nil {
						t.Fatal(err)
					}
				}, 800, 200)
				switch family {
				case "date", "date-range":
					if err := parent.Click("Open"); err != nil {
						t.Fatal(err)
					}
				case "hover-card", "tooltip":
					parent.Key(0, ui.KeyTab)
				}
				if len(backend.surfaces) != 1 {
					t.Fatalf("opened %d native surfaces", len(backend.surfaces))
				}
				surface := backend.surfaces[0]
				child := ui.NewTesterWith(surface.view, 680, 420, ui.TesterOptions{Dark: dark})
				child.SetVibrancy(true)
				x, y := 1, 20
				if family == "date" || family == "date-range" {
					calendar, ok := child.AccessByRole(ui.RoleGroup, "Calendar")
					if !ok {
						t.Fatal("native calendar missing")
					}
					x, y = int(calendar.Bounds.X+calendar.Bounds.W/2), int(calendar.Bounds.Y+6)
				} else if family == "tooltip" {
					x, y = 600, 400
				}
				if alpha := child.Image().RGBAAt(x, y).A; alpha != 0 {
					t.Errorf("native theme wrapper covered system material: alpha=%d", alpha)
				}
				// Backends without a material still need an opaque platform surface.
				child.SetVibrancy(false)
				if family != "tooltip" && child.Image().RGBAAt(x, y).A != 255 {
					t.Error("native platform fallback lost its background")
				}
				host.CloseAll()
				child.Frame()
				if paintedAlpha(child.Image()) != 0 || child.FramePending() || surface.closes != 1 {
					t.Error("native child flashed or kept rendering after teardown")
				}
			})
		}
	}
}
