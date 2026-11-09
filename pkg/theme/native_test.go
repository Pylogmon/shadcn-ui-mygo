package theme

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestNativeScopePreservesPlatformThemeAndRestoresParent(t *testing.T) {
	native := *ui.DarkTheme()
	native.Accent = ui.RGB(12, 86, 171)
	native.Surface = ui.RGB(31, 47, 63)
	native.SurfaceHover = ui.RGB(50, 66, 80)
	native.Font = "monospace"
	native.FontSize = 17
	native.FocusRingWidth = 5
	native.FocusRingOffset = -1
	native.TooltipDelay = 123 * time.Millisecond
	var captured ui.Theme
	var tokens Tokens
	var parent *ui.Theme
	r := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(&native)
		NativeScope(c).Children(func() { captured = *c.Theme(); tokens = *TokensOf(c); ui.Button(c, "Native control") })
		parent = c.Theme()
	}, 260, 120)
	if captured != native {
		t.Fatalf("native properties were reconstructed: got %+v want %+v", captured, native)
	}
	if parent != &native {
		t.Fatal("native scope leaked to parent")
	}
	if tokens.Primary != native.Accent || tokens.Popover != native.Surface || tokens.Accent != native.SurfaceHover || tokens.Ring != native.Focus {
		t.Fatalf("semantic aliases lost platform colors %+v", tokens)
	}
	if _, ok := r.Find("Native control"); !ok {
		t.Fatal("native content not built")
	}
	native.Accent = ui.RGB(200, 100, 30)
	r.Frame()
	if tokens.Primary != native.Accent || captured != native {
		t.Fatal("platform theme changes did not reach new scope")
	}
}
