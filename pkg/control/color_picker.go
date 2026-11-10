package control

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// ColorSwatchShape controls both preset swatches and the optional custom well.
type ColorSwatchShape uint8

const (
	ColorSwatchRounded ColorSwatchShape = iota
	ColorSwatchSquare
	ColorSwatchCircle
)

// ColorOption is a preset color. An empty Label uses its hexadecimal value.
type ColorOption struct {
	Color    ui.Color
	Label    string
	Disabled bool
}

type ColorPickerProps struct {
	Label, Description string
	// Nil uses DefaultColorOptions; an empty slice shows no presets.
	Colors []ColorOption
	// Custom adds an equally sized rainbow well before the presets. It opens
	// MyGo's ColorPicker in a window-local popover, including alpha and hex input.
	Custom      bool
	CustomLabel string
	Shape       ColorSwatchShape
	// SwatchSize is in DIPs; zero uses 48. Spacing defaults to 24; a pointer
	// permits an explicit zero. Invalid dimensions fall back to the defaults.
	SwatchSize        float32
	Spacing           *float32
	Disabled, Invalid bool
	OnValueChange     func(ui.Color)
}

// DefaultColorOptions returns an independent copy of the reference palette.
func DefaultColorOptions() []ColorOption {
	return []ColorOption{
		{Color: ui.Hex("#193cb8"), Label: "Blue"},
		{Color: ui.Hex("#6e11b0"), Label: "Purple"},
		{Color: ui.Hex("#8a0194"), Label: "Magenta"},
		{Color: ui.Hex("#a50036"), Label: "Red"},
		{Color: ui.Hex("#973c00"), Label: "Orange"},
		{Color: ui.Hex("#016630"), Label: "Green"},
		{Color: ui.Hex("#005f78"), Label: "Teal"},
		{Color: ui.Hex("#e7e5e4"), Label: "White"},
	}
}

type colorPickerOptionKey struct {
	color      ui.Color
	occurrence int
}

type colorPickerPart uint8

const (
	colorPickerCustom colorPickerPart = iota
	colorPickerPresets
	colorPickerPanel
)

type colorPickerSwatch struct {
	element ui.Element
	color   ui.Color
	custom  bool
}

// ColorPicker selects a caller-owned RGBA color. Presets are native radios:
// arrows/Home/End select, disabled options are skipped, and Tab enters at the
// selected preset. Custom is a separate native keyboard target. Selection is
// never cleared by clicking it again. Changed and OnValueChange report edits
// once in the input frame, including edits made in MyGo's picker.
func ColorPicker(c *ui.Context, value *ui.Color, p ColorPickerProps) ui.Element {
	if value == nil {
		panic("control: ColorPicker requires color state")
	}
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	size, spacing := float32(48), float32(24)
	if p.SwatchSize > 0 && finiteNumber(float64(p.SwatchSize)) {
		size = p.SwatchSize
	}
	if p.Spacing != nil && *p.Spacing >= 0 && finiteNumber(float64(*p.Spacing)) {
		spacing = *p.Spacing
	}
	colors := p.Colors
	if colors == nil {
		colors = DefaultColorOptions()
	}
	label := p.Label
	if label == "" {
		label = "Color"
	}
	root := ui.Row(c).AlignSelf(ui.Start).MinWidth(0).MaxWidthPercent(100).Gap(spacing).
		AlignItems(ui.Start).Role(ui.RoleGroup).Label(label).Description(p.Description).
		Disabled(p.Disabled).DimDisabled(false).Invalid(p.Invalid)
	open := ui.Local(root, "color-picker-open", func() bool { return false })
	if !p.Custom || root.IsDisabled() {
		*open = false
	}
	before := *value
	swatches := make([]colorPickerSwatch, 0, len(colors)+1)
	root.Children(func() {
		if p.Custom {
			name := p.CustomLabel
			if name == "" {
				name = "Custom color"
			}
			well := ui.ButtonBase(c.Key(colorPickerCustom)).Role(ui.RoleColorWell).
				Label(name).Description(colorPickerHex(*value)).Value(colorPickerHex(*value))
			if well.Clicked() && !root.IsDisabled() {
				*open = !*open
			}
			ui.PopoverBase(c.Key(colorPickerPanel), well, open, func(panel ui.Element) {
				t := theme.TokensOf(c)
				panel.Column().Padding(12).Margin(8, 0, 0, 0).Radius(t.RadiusLG()).
					Background(t.Popover).TextColor(t.PopoverForeground).Border(1, t.Border).
					Shadow(0, 6, 20, 0, ui.RGBA(0, 0, 0, .18)).Label(name)
				ui.ColorPicker(c, value).Changed()
			})
			well.Expanded(*open).Value(colorPickerHex(*value)).Description(colorPickerHex(*value))
			swatches = append(swatches, colorPickerSwatch{element: well, custom: true})
		}
		if len(colors) > 0 {
			occurrences := map[ui.Color]int{}
			ui.RadioGroup(c.Key(colorPickerPresets), func() {
				for _, option := range colors {
					key := colorPickerOptionKey{option.Color, occurrences[option.Color]}
					occurrences[option.Color]++
					name := option.Label
					if name == "" {
						name = colorPickerHex(option.Color)
					}
					radio := ui.RadioBase(c.Key(key), value, option.Color).Label(name).
						Description(colorPickerHex(option.Color)).Disabled(option.Disabled)
					radio.Changed()
					swatches = append(swatches, colorPickerSwatch{element: radio, color: option.Color})
				}
			}).Row().Wrap().MinWidth(0).MaxWidthPercent(100).Shrink(1).Gap(spacing).
				Label(label).FocusGroup(ui.Horizontal).Orientation(ui.Horizontal)
		}
	})
	if before != *value {
		root.SetChanged()
		c.Invalidate()
		if p.OnValueChange != nil {
			p.OnValueChange(*value)
		}
	}
	customSelected := true
	for _, option := range colors {
		if option.Color == *value {
			customSelected = false
			break
		}
	}
	for _, swatch := range swatches {
		selected := swatch.color == *value
		if swatch.custom {
			selected = customSelected
			swatch.color = *value
			swatch.element.Value(colorPickerHex(*value)).Description(colorPickerHex(*value))
		} else {
			swatch.element.Checked(selected)
		}
		styleColorSwatch(c, swatch, selected, size, p.Shape, p.Invalid)
	}
	return root
}

func colorPickerHex(color ui.Color) string {
	if color.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", color.R, color.G, color.B, color.A)
}

func styleColorSwatch(c *ui.Context, swatch colorPickerSwatch, selected bool, size float32, shape ColorSwatchShape, invalid bool) {
	t := theme.TokensOf(c)
	radius := size * 7 / 24
	switch shape {
	case ColorSwatchSquare:
		radius = 0
	case ColorSwatchCircle:
		radius = size / 2
	}
	e := swatch.element.Size(size, size).Shrink(0).Padding(0).Radius(radius).
		Background(ui.Transparent).Border(0, ui.Transparent).FocusRing(false).
		DimDisabled(false).Invalid(invalid).Cursor(ui.CursorPointer)
	if e.IsDisabled() {
		e.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	focused := e.FocusVisible()
	// Composite alpha before choosing the ring and check contrast. This also
	// keeps transparent and near-background colors visibly selectable.
	color := theme.Mix(t.Background, ui.RGB(swatch.color.R, swatch.color.G, swatch.color.B), float32(swatch.color.A)/255)
	ring := color
	if theme.Contrast(ring, t.Background) < 1.5 {
		ring = t.Foreground
	}
	if !swatch.custom && swatch.color.A < 255 {
		// Alpha presets are drawn over white/gray checks, regardless of theme.
		color = theme.Mix(ui.RGB(230, 230, 230), ui.RGB(swatch.color.R, swatch.color.G, swatch.color.B), float32(swatch.color.A)/255)
	}
	foreground := ui.RGB(255, 255, 255)
	if theme.Contrast(ui.RGB(0, 0, 0), color) > theme.Contrast(foreground, color) {
		foreground = ui.RGB(0, 0, 0)
	}
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		outline := func(amount, width float32, color ui.Color) {
			rad := radius
			if rad > 0 {
				rad += amount
			}
			p.Stroke(ui.Rect{X: r.X - amount, Y: r.Y - amount, W: r.W + 2*amount, H: r.H + 2*amount}, color, rad, width)
		}
		if selected {
			outline(7, 4, ring)
		}
		if focused {
			outline(11, 2, t.Ring)
		}
		if invalid {
			outline(7, 2, t.Destructive)
		}
		if swatch.custom {
			p.FillConicGradient(r, ui.ConicGradient{Stops: []ui.GradientStop{
				{Color: ui.Hex("#ef4444"), At: 0}, {Color: ui.Hex("#facc15"), At: 1.0 / 6},
				{Color: ui.Hex("#22c55e"), At: 2.0 / 6}, {Color: ui.Hex("#06b6d4"), At: 3.0 / 6},
				{Color: ui.Hex("#3b82f6"), At: 4.0 / 6}, {Color: ui.Hex("#a855f7"), At: 5.0 / 6},
				{Color: ui.Hex("#ef4444"), At: 1},
			}}, radius)
			p.FillRadialGradient(r, ui.RadialGradient{Stops: []ui.GradientStop{
				{Color: ui.RGBA(255, 255, 255, .9), At: 0}, {Color: ui.Transparent, At: 1},
			}}, radius)
			if selected {
				d := size * .6
				p.Fill(ui.Rect{X: r.X + (size-d)/2, Y: r.Y + (size-d)/2, W: d, H: d}, color, d/2)
			}
		} else {
			p.Fill(r, t.Background, radius)
			if swatch.color.A < 255 {
				// Clip edge cells with the same rounded bounds as the swatch.
				// A checker gradient stays within those bounds without extra nodes.
				p.FillConicGradient(r, ui.ConicGradient{Stops: []ui.GradientStop{
					{Color: ui.RGB(255, 255, 255), At: 0}, {Color: ui.RGB(255, 255, 255), At: .25},
					{Color: ui.RGB(204, 204, 204), At: .25}, {Color: ui.RGB(204, 204, 204), At: .5},
					{Color: ui.RGB(255, 255, 255), At: .5}, {Color: ui.RGB(255, 255, 255), At: .75},
					{Color: ui.RGB(204, 204, 204), At: .75}, {Color: ui.RGB(204, 204, 204), At: 1},
				}}, radius)
			}
			p.Fill(r, swatch.color, radius)
		}
		if selected {
			d := min(float32(24), size*.5)
			p.Icon(check, ui.Rect{X: r.X + (size-d)/2, Y: r.Y + (size-d)/2, W: d, H: d}, foreground)
		}
	})
}
