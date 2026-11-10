package gallery

import (
	"math"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Only application values survive a build. Native thumbs own pointer capture,
// keyboard focus and accessibility; no frame Context or Element is retained.
type sliderCompositionPreviewState struct {
	ready, native                bool
	recipe, language, popupError string
	drafts                       map[string]*sliderCompositionDraft
}

type sliderCompositionDraft struct {
	values   []float64
	vertical [2]float64
	changes  int
}

var sliderCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Basic (docs)"},
	{Value: "range", Label: "Range (docs)"},
	{Value: "multiple", Label: "Multiple thumbs (docs)"},
	{Value: "vertical", Label: "Vertical pair (docs)"},
	{Value: "controlled", Label: "Controlled temperature (docs)"},
	{Value: "disabled", Label: "Disabled (docs)"},
	{Value: "rtl", Label: "RTL slider"},
	{Value: "create-basic", Label: "Create: Basic"},
	{Value: "create-range", Label: "Create: Range"},
	{Value: "create-multiple", Label: "Create: Multiple thumbs"},
	{Value: "create-vertical", Label: "Create: Vertical pair"},
	{Value: "create-controlled", Label: "Create: Controlled temperature"},
	{Value: "create-disabled", Label: "Create: Disabled"},
}

func (p *sliderCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.drafts = map[string]*sliderCompositionDraft{}
}

func (p *sliderCompositionPreviewState) draft() *sliderCompositionDraft {
	if draft := p.drafts[p.recipe]; draft != nil {
		return draft
	}
	draft := &sliderCompositionDraft{values: []float64{75}, vertical: [2]float64{50, 25}}
	switch strings.TrimPrefix(p.recipe, "create-") {
	case "basic", "disabled":
		draft.values = []float64{50}
	case "range":
		draft.values = []float64{25, 50}
	case "multiple":
		draft.values = []float64{10, 20, 70}
	case "controlled":
		draft.values = []float64{.3, .7}
	}
	p.drafts[p.recipe] = draft
	return draft
}

func (p *sliderCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *sliderCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-slider-config")
			h.Close("source-slider-language")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Slider compositions"}, display.CardSlots{
		Title: "Slider compositions", Description: "Source single, range and multiple native thumbs, vertical pairs and controlled temperature values.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-slider-config", &p.recipe, control.SelectProps{Label: "Slider configuration", Options: sliderCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-slider-language", &p.language, control.SelectProps{Label: "Slider language", Options: statusLanguages()})
			}
			preview := ui.Column(child.Key("slider-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
			if p.recipe == "rtl" {
				preview.Direction(childDirection(p.language))
			}
			preview.Children(func() { p.draw(child) })
			if p.popupError != "" {
				ui.Text(child, p.popupError).Role(ui.RoleAlert).TextColor(theme.TokensOf(child).Destructive)
			}
		},
	})
}

func (p *sliderCompositionPreviewState) draw(c *ui.Context) {
	draft := p.draft()
	kind := strings.TrimPrefix(p.recipe, "create-")
	container := ui.Column(c).FillWidth().MinWidth(0).Gap(12)
	if !strings.HasPrefix(p.recipe, "create-") {
		container.MaxWidth(320)
	}
	container.Children(func() {
		if kind == "vertical" {
			row := ui.Row(c).Gap(24).AlignItems(ui.Center).AlignSelf(ui.Center)
			if p.recipe == "vertical" {
				row.FillWidth().Justify(ui.Center)
			}
			row.Children(func() {
				for i, label := range []string{"Vertical slider 1", "Vertical slider 2"} {
					if control.Slider(c.Key(i), &draft.vertical[i], control.SliderProps{Label: label, Max: 100, Step: 1, Vertical: true}).Height(160).Changed() {
						draft.changes++
					}
				}
			})
			return
		}
		props := control.SliderProps{Label: "Slider", Max: 100, Step: 1, Disabled: kind == "disabled"}
		var label ui.Element
		if kind == "controlled" {
			props.Label, props.Max, props.Step = "Temperature", 1, .1
			ui.Row(c).FillWidth().MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
				label = control.Label(c, "Temperature")
				ui.Box(c).Grow(1)
				ui.Text(c, sliderSourceTemperature(draft.values)).FontSize(14).TextColor(theme.TokensOf(c).MutedForeground)
			})
		} else if kind == "range" {
			props.Step = 5
		} else if kind == "multiple" {
			props.Step = 10
		}
		var slider ui.Element
		switch len(draft.values) {
		case 1:
			slider = control.Slider(c, &draft.values[0], props)
		case 2:
			slider = control.RangeSlider(c, &draft.values[0], &draft.values[1], props)
		default:
			slider = control.SliderValues(c, &draft.values, props)
		}
		if label.Valid() {
			slider.LabelledBy(label)
		}
		if slider.Changed() {
			draft.changes++
		}
	})
}

func sliderSourceTemperature(values []float64) string {
	parts := make([]string, len(values))
	for i, value := range values {
		// Source values are tenths; avoid displaying binary roundoff from the
		// native numeric step calculation as extra decimal digits.
		parts[i] = strconv.FormatFloat(math.Round(value*10)/10, 'f', -1, 64)
	}
	return strings.Join(parts, ", ")
}
