package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type colorPickerPreviewState struct {
	ready            bool
	value            ui.Color
	shape            string
	custom, disabled bool
}

func (p *colorPickerPreviewState) view(c *ui.Context) {
	if !p.ready {
		p.ready, p.custom, p.shape = true, true, "rounded"
		p.value = control.DefaultColorOptions()[0].Color
	}
	t := theme.TokensOf(c)
	display.CardWith(c, display.CardProps{Label: tr(c, "Color picker example")}, display.CardSlots{
		Title: tr(c, "Color Picker"), Description: tr(c, "Choose a preset or open the rainbow well to pick any color."),
		Content: func(child *ui.Context) {
			ui.Column(child).FillWidth().Gap(24).Children(func() {
				control.RadioGroup(child, &p.shape, control.RadioGroupProps{
					Label: tr(child, "Swatch shape"), Horizontal: true, Orientation: ui.Horizontal,
					Options: localizedOptions(child, []control.Option{{Value: "rounded", Label: "Rounded square"}, {Value: "square", Label: "Square"}, {Value: "circle", Label: "Circle"}}),
				}).Wrap()
				ui.Row(child).Wrap().Gap(20).Children(func() {
					control.Switch(child, &p.custom, control.SwitchProps{Label: tr(child, "Custom color well")})
					control.Switch(child, &p.disabled, control.SwitchProps{Label: tr(child, "Disable color picker")})
				})
				shape := control.ColorSwatchRounded
				switch p.shape {
				case "square":
					shape = control.ColorSwatchSquare
				case "circle":
					shape = control.ColorSwatchCircle
				}
				ui.Column(child).FillWidth().Padding(12).Children(func() {
					control.ColorPicker(child, &p.value, control.ColorPickerProps{Label: tr(child, "Accent color"), Shape: shape, Custom: p.custom, Disabled: p.disabled})
				})
				ui.Row(child).Gap(10).AlignItems(ui.Center).Children(func() {
					ui.Box(child).Size(18, 18).Radius(4).Background(p.value).Border(1, t.Border).AccessHidden(true)
					ui.Text(child, fmt.Sprintf(tr(child, "Selected: #%02x%02x%02x · alpha %d%%"), p.value.R, p.value.G, p.value.B, int(p.value.A)*100/255)).
						FontSize(14).TextColor(t.MutedForeground).Role(ui.RoleStatus)
				})
			})
		},
	})
}
