package theme

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestControlFontsApplyDesktopScaleExactlyOnce(t *testing.T) {
	for _, route := range []string{"apply", "tokens", "controller", "font-scope", "menu", "native-scope"} {
		t.Run(route, func(t *testing.T) {
			controller := Controller{}
			var actual, small, override float32
			read := func(c *ui.Context) {
				actual = ControlFontSize(c, 14, 0)
				small = ControlFontSize(c, 12, 0)
				override = ControlFontSize(c, 14, 18)
			}
			r := ui.NewTesterWith(func(c *ui.Context) {
				switch route {
				case "apply":
					Apply(c)
					read(c)
				case "tokens":
					c.SetTheme(Light().Theme())
					read(c)
				case "native-scope":
					// NativeScope retains the platform theme, which already follows preferences.
					NativeScope(c).Children(func() { read(c) })
				default:
					controller.Apply(c)
					if route == "font-scope" {
						FontScope(c, FontRoles{Body: "monospace"}).Children(func() { read(c) })
					} else if route == "menu" {
						ui.Themed(c, MenuOf(c).Theme()).Children(func() { read(c) })
					} else {
						read(c)
					}
				}
			}, 200, 100, ui.TesterOptions{Preferences: ui.Preferences{TextScale: 2}})
			// The platform's default font can differ from the shadcn baseline.
			if route != "native-scope" && (actual != 28 || small != 24) || override != 36 {
				t.Fatalf("font scaled zero/two times: actual=%g small=%g override=%g", actual, small, override)
			}
			r.SetPreferences(ui.Preferences{TextScale: 1.5})
			if route != "native-scope" && (actual != 21 || small != 18) || override != 27 {
				t.Fatalf("live text-scale update was stale: actual=%g small=%g override=%g", actual, small, override)
			}
		})
	}
}

func TestControlFontsRespectNativeThemeMutationAndRejectInvalidOverrides(t *testing.T) {
	native := Light().Theme()
	for _, invalid := range []float32{0, -1, float32(math.NaN()), float32(math.Inf(1))} {
		ui.NewTester(func(c *ui.Context) {
			c.SetTheme(native)
			native.FontSize = 28
			if got := ControlFontSize(c, 12, invalid); got != 24 {
				t.Fatalf("native FontSize or invalid override ignored: %g", got)
			}
		}, 200, 100)
	}
}
