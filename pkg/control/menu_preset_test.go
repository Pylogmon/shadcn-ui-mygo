package control

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func menuRedArea(im *image.RGBA) int {
	n := 0
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			p := im.RGBAAt(x, y)
			if p.R == 255 && p.G == 0 && p.B == 0 {
				n++
			}
		}
	}
	return n
}
func TestMenuPresetComboboxAndSelectUseForceDarkAndNoPopupAnimation(t *testing.T) {
	value := ComboboxValue{}
	query, selected := "", ""
	capturedDark := false
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		theme.MenuScope(c, theme.Preset("blue", false), theme.MenuOptions{Color: theme.MenuInvertedTranslucent, Accent: theme.MenuBold, InvertedTokens: theme.Preset("blue", true)}).Fill().Padding(20).Gap(20).Children(func() {
			ComboboxWith(c, &value, &query, ComboboxWithProps{Label: "Preset combo", AutoHighlight: true, Options: []Option{{Value: "go", Label: "Go"}}, ItemContent: func(c *ui.Context, _ Option) {
				capturedDark = theme.TokensOf(c).Dark
				ui.Box(c).Size(16, 16).Background(ui.RGB(255, 0, 0))
			}})
			Select(c, &selected, SelectProps{Label: "Preset select", Options: []Option{{Value: "go", Label: "Go"}}, ItemContent: func(c *ui.Context, _ Option) {
				capturedDark = theme.TokensOf(c).Dark
				ui.Box(c).Size(16, 16).Background(ui.RGB(255, 0, 0))
			}})
		})
	}, 400, 320)
	if err := r.Click("Preset combo"); err != nil {
		t.Fatal(err)
	}
	if area := menuRedArea(r.Image()); !capturedDark || area < 250 {
		t.Fatalf("combo not fully shown in force-dark profile at t0: dark%v area%d", capturedDark, area)
	}
	r.Type("Go")
	r.Key(0, ui.KeyEnter)
	if !value.Valid || query != "Go" {
		t.Fatalf("profile broke native choice %+v/%q", value, query)
	}
	if _, ok := r.AccessByRole(ui.RoleOption, "Go"); ok {
		t.Fatal("translucent combo retained an accessible exit layer")
	}
	capturedDark = false
	if err := r.Click("Preset select"); err != nil {
		t.Fatal(err)
	}
	if area := menuRedArea(r.Image()); !capturedDark || area < 250 {
		t.Fatalf("Select not fully shown in force-dark profile at t0: dark%v area%d", capturedDark, area)
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	if selected != "go" {
		t.Fatalf("Select native selection=%q", selected)
	}
}
