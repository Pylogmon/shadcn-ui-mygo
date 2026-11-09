package display

import (
	"bytes"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestKbdUsesThemeBodyFamilyInsideSpecializedFontScope(t *testing.T) {
	native := theme.Light().Theme()
	native.Font = "serif"
	view := func(parent, override string) func(*ui.Context) {
		return func(c *ui.Context) {
			ui.Themed(c, native).Fill().Background(native.Background).Children(func() {
				ui.Column(c).Padding(20).Font(parent).Children(func() {
					key := Kbd(c, "Command").Label("Key cap")
					if override != "" {
						key.Font(override)
					}
				})
			})
		}
	}
	monospaceParent := ui.NewTester(view("monospace", ""), 250, 100)
	reference := ui.NewTester(view("serif", ""), 250, 100)
	if !bytes.Equal(monospaceParent.Image().Pix, reference.Image().Pix) {
		t.Fatal("keycap inherited a specialized parent font instead of the configured body family")
	}
	custom := ui.NewTester(view("serif", "monospace"), 250, 100)
	if bytes.Equal(custom.Image().Pix, reference.Image().Pix) {
		t.Fatal("keycap's explicit fluent font override did not apply")
	}
}
