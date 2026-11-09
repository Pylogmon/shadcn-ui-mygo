package gallery

import (
	"fmt"
	"math"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type chartPreviewState struct {
	ready                        bool
	orientation, mode            string
	selected                     int
	legend, values, grid, change bool
	updated                      bool
	animate, storyUpdated        bool
	family                       string
	familySelected               int
}

func (p *chartPreviewState) initialize() {
	if p.ready {
		return
	}
	p.ready = true
	p.orientation, p.mode, p.selected = "horizontal", "single", -1
	p.family, p.familySelected = "line", -1
	p.animate = true
}

func (p *chartPreviewState) props(c *ui.Context) display.ChartProps {
	radius := float32(5)
	props := display.ChartProps{
		Kind: display.BarChart, Label: "Visitor comparison", Height: 260,
		Labels: []string{"January", "February", "March", "April", "May", "June"},
		Series: []display.ChartSeries{
			{Name: "Desktop", Values: []float64{186, 305, 237, 73, 209, 214}},
		},
		Selected: &p.selected, Legend: p.legend,
		Animation: display.ChartAnimationProps{Disabled: !p.animate},
		Cartesian: display.ChartCartesianProps{
			HideValueAxis: !p.values, HideGrid: !p.grid, HideCursor: true,
			BarRadius: &radius,
			FormatCategory: func(label string) string {
				letters := []rune(label)
				return string(letters[:min(3, len(letters))])
			},
		},
	}
	if p.orientation == "horizontal" {
		props.Orientation = display.ChartHorizontal
	}
	if p.change {
		props.Series[0] = display.ChartSeries{Name: "Desktop change", Values: []float64{186, 205, -207, 173, -209, 214}}
	}
	if p.updated {
		props.Series[0].Values = []float64{220, 328, 252, 105, 231, 246}
		if p.change {
			props.Series[0].Values = []float64{220, 198, -180, 205, -164, 246}
		}
	}
	if p.mode == "grouped" || p.mode == "stacked" || p.mode == "normalized" {
		name := "Mobile"
		values := []float64{80, 200, 120, 190, 130, 140}
		if p.change {
			name, values = "Mobile change", []float64{80, 160, -90, 120, -70, 140}
		}
		props.Series = append(props.Series, display.ChartSeries{Name: name, Values: values})
		props.Stacked = p.mode != "grouped"
		props.Normalized = p.mode == "normalized"
		if props.Normalized {
			props.Cartesian.FormatValue = func(value float64) string { return fmt.Sprintf("%.0f%%", value*100) }
		}
	}
	if p.mode == "browsers" {
		props.Labels = []string{"Chrome", "Safari", "Firefox", "Edge", "Other"}
		props.Series = []display.ChartSeries{{Name: "Visitors", Values: []float64{275, 200, 187, 173, 90}}}
		colors := theme.TokensOf(c).Charts()
		props.CategoryColors = colors[:]
		props.Cartesian.FormatCategory = nil
		if p.updated {
			props.Series[0].Values = []float64{302, 218, 192, 185, 96}
		}
	}
	return props
}

func (s *State) chartCompositionPreview(c *ui.Context) {
	p := &s.chartPreview
	p.initialize()
	ui.Column(c).FillWidth().MinWidth(0).Shrink(0).Gap(20).Children(func() {
		control.Switch(c, &p.animate, control.SwitchProps{Label: "Animate chart changes", Size: control.Small})
		s.chartBarPreview(c)
		s.chartFamilyPreview(c)
	})
}

func (s *State) chartBarPreview(c *ui.Context) {
	p := &s.chartPreview
	display.Card(c, func() {
		s.cardHeader(c, "A different view of the numbers", "Compare visitors, explore a change and inspect the values that matter.")
		control.RadioGroup(c, &p.orientation, control.RadioGroupProps{
			Label: "Bar orientation", Horizontal: true,
			Options: []control.Option{{Value: "horizontal", Label: "Horizontal bars"}, {Value: "vertical", Label: "Vertical bars"}},
		}).Wrap().MinWidth(0)
		control.RadioGroup(c, &p.mode, control.RadioGroupProps{
			Label: "Visitor view", Horizontal: true,
			Options: []control.Option{
				{Value: "single", Label: "Desktop only"}, {Value: "grouped", Label: "Side by side"},
				{Value: "stacked", Label: "Stacked totals"}, {Value: "normalized", Label: "Share of total"},
				{Value: "browsers", Label: "Browsers"},
			},
		}).Wrap().MinWidth(0)
		ui.Row(c).Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
			control.Switch(c, &p.values, control.SwitchProps{Label: "Show axis values", Size: control.Small})
			control.Switch(c, &p.grid, control.SwitchProps{Label: "Show grid", Size: control.Small})
			control.Switch(c, &p.legend, control.SwitchProps{Label: "Show chart legend", Size: control.Small})
			control.Switch(c, &p.change, control.SwitchProps{Label: "Month-to-month change", Size: control.Small, Disabled: p.mode == "browsers"})
		})
		props := p.props(c)
		display.Chart(c, props).Shrink(0)
		ui.Row(c).Wrap().Gap(12).AlignItems(ui.Center).Children(func() {
			if control.Button(c, "Update visitor counts", control.ButtonProps{Variant: control.Outline, Size: control.Small, Disabled: p.updated}).Clicked() {
				p.updated = true
				c.Invalidate()
			}
			if control.Button(c, "Restore visitor counts", control.ButtonProps{Variant: control.Ghost, Size: control.Small, Disabled: !p.updated}).Clicked() {
				p.updated = false
				c.Invalidate()
			}
		})
		text := "Choose a category in the chart to compare its values."
		if p.selected >= 0 && p.selected < len(props.Labels) {
			values := []string{}
			for _, series := range props.Series {
				if p.selected < len(series.Values) && !math.IsNaN(series.Values[p.selected]) && !math.IsInf(series.Values[p.selected], 0) {
					values = append(values, fmt.Sprintf("%s: %g", series.Name, series.Values[p.selected]))
				}
			}
			text = props.Labels[p.selected] + " · " + strings.Join(values, " · ")
		}
		ui.Text(c, text).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
	}).Label("Visitor explorer")
}

func (s *State) chartFamilyPreview(c *ui.Context) {
	p := &s.chartPreview
	display.Card(c, func() {
		s.cardHeader(c, "Find the shape of the story", "Follow a trend, compare a share or see how the months fit together.")
		control.RadioGroup(c, &p.family, control.RadioGroupProps{
			Label: "More chart styles", Horizontal: true,
			Options: []control.Option{
				{Value: "line", Label: "Line"}, {Value: "area", Label: "Area"},
				{Value: "pie", Label: "Pie"}, {Value: "donut", Label: "Donut"},
				{Value: "radar", Label: "Radar"}, {Value: "radial", Label: "Radial"},
			},
		}).Wrap().MinWidth(0)
		props := display.ChartProps{
			Kind: display.LineChart, Label: "Visitor story", Height: 240, Legend: true,
			Animation: display.ChartAnimationProps{Disabled: !p.animate},
			Selected:  &p.familySelected,
			Labels:    []string{"January", "February", "March", "April", "May", "June"},
			Series: []display.ChartSeries{
				{Name: "Desktop", Values: []float64{186, 305, 237, 73, 209, 214}},
				{Name: "Mobile", Values: []float64{80, 200, 120, 190, 130, 140}},
			},
			Cartesian: display.ChartCartesianProps{FormatCategory: func(label string) string {
				letters := []rune(label)
				return string(letters[:min(3, len(letters))])
			}},
		}
		if p.storyUpdated {
			props.Series[0].Values = []float64{286, 145, 312, 173, 126, 280}
			props.Series[1].Values = []float64{140, 110, 240, 86, 210, 90}
		}
		switch p.family {
		case "area":
			props.Kind, props.Stacked = display.AreaChart, true
		case "pie", "donut", "radial":
			props.Kind = display.PieChart
			if p.family == "donut" {
				props.Kind = display.DonutChart
			} else if p.family == "radial" {
				props.Kind = display.RadialChart
			}
			props.Labels = []string{"Chrome", "Safari", "Firefox", "Edge", "Other"}
			props.Series = []display.ChartSeries{{Name: "Visitors", Values: []float64{275, 200, 187, 173, 90}}}
			if p.storyUpdated {
				props.Series[0].Values = []float64{160, 310, 125, 240, 185}
			}
		case "radar":
			props.Kind = display.RadarChart
		}
		display.Chart(c, props).Shrink(0)
		ui.Row(c).Wrap().Gap(12).AlignItems(ui.Center).Children(func() {
			if control.Button(c, "Update story counts", control.ButtonProps{Variant: control.Outline, Size: control.Small, Disabled: p.storyUpdated}).Clicked() {
				p.storyUpdated = true
				c.Invalidate()
			}
			if control.Button(c, "Restore story counts", control.ButtonProps{Variant: control.Ghost, Size: control.Small, Disabled: !p.storyUpdated}).Clicked() {
				p.storyUpdated = false
				c.Invalidate()
			}
		})
	}).Label("Chart style explorer")
}
