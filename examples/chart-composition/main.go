// This application composes raw MyGo chart marks with the kit's chart surfaces.
// The range filter and bar labels are application code, not Recharts Brush or
// LabelList implementations. No display.Chart is involved.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var months = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

// This simple geometric icon is authored here for the example.
var targetIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 12 12"><path d="M1 6h10" fill="none" stroke="currentColor" stroke-width="2" stroke-dasharray="2 2"/></svg>`))

type appState struct {
	appearance theme.Controller
	dark       bool
	low, high  float64
	revenue    []float64
	selected   int // Absolute month index; retained when the pointer leaves.
	indicator  string
	hide       bool
	edit, note string
	editError  string
}

func newState(dark, inspect bool) *appState {
	s := &appState{
		appearance: theme.Controller{Preset: "blue"}, dark: dark,
		low: 0, high: 5, selected: -1, indicator: "dot",
		revenue: []float64{186, 305, 237, 173, 209, 214, 287, 319, 246, 278, 330, 352},
		note:    "Compare the quieter spring months with the monthly target.",
	}
	if inspect {
		s.inspect(2)
	}
	return s
}

func (s *appState) inspect(index int) {
	if index == s.selected {
		return
	}
	s.selected, s.editError = index, ""
	s.edit = ""
	if index >= 0 {
		s.edit = strconv.FormatFloat(s.revenue[index], 'f', -1, 64)
	}
}

func (s *appState) applyEdit() {
	if s.selected < 0 {
		return
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s.edit), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 400 {
		s.editError = "Enter a number from 0 to 400."
		return
	}
	s.revenue[s.selected], s.editError = v, ""
}

func (s *appState) view(c *ui.Context) {
	s.appearance.Mode = theme.LightMode
	if s.dark {
		s.appearance.Mode = theme.DarkMode
	}
	s.appearance.Apply(c)
	t := theme.TokensOf(c)
	ui.Scroll(c).Fill().Background(t.Background).Children(func() {
		ui.Column(c).FillWidth().MaxWidth(1000).MarginX(ui.Auto).Padding(24).Gap(18).Children(func() {
			display.Typography(c, "Monthly revenue", display.H2)
			ui.Text(c, "Choose a month to inspect its revenue, then edit a value or narrow the date range.").TextColor(t.MutedForeground)
			ui.Row(c).FillWidth().Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
				control.RadioGroup(c, &s.indicator, control.RadioGroupProps{Label: "Tooltip indicator", Horizontal: true, Options: []control.Option{
					{Value: "dot", Label: "Dot"}, {Value: "line", Label: "Line"}, {Value: "dashed", Label: "Dashed"},
				}})
				control.Switch(c, &s.hide, control.SwitchProps{Label: "Hide indicator"})
				if control.Switch(c, &s.dark, control.SwitchProps{Label: "Dark appearance"}).Changed() {
					c.Invalidate()
				}
			})
			s.composition(c)
		})
	})
}

func (s *appState) composition(c *ui.Context) {
	config := display.ChartConfig{
		"revenue": {Label: "Revenue", Theme: &display.ChartThemeColors{Light: ui.RGB(37, 99, 235), Dark: ui.RGB(147, 197, 253)}},
		"target": {Label: "Monthly target", Theme: &display.ChartThemeColors{Light: ui.RGB(234, 88, 12), Dark: ui.RGB(251, 146, 60)},
			Icon: func(child *ui.Context) ui.Element { return ui.Icon(child, targetIcon).AccessHidden(true) }},
	}
	display.ChartContainer(c, display.ChartContainerProps{Key: "revenue-composition", Label: "Revenue workspace", Config: config}, func(c *ui.Context) {
		// Resolve colors now. Draw receives copied colors/ticks, never Context.
		barColor, _ := display.ChartColor(c, "revenue")
		targetColor, _ := display.ChartColor(c, "target")
		t := theme.TokensOf(c)
		ui.Column(c).FillWidth().Gap(14).Children(func() {
			ui.Text(c, "Visible months").FontWeight(500)
			c.Key("month-range")
			control.RangeSlider(c, &s.low, &s.high, control.SliderProps{
				Min: 0, Max: 11, Step: 1, MinGap: 1, Label: "Month range",
				ThumbIDs: []string{"from", "through"}, ThumbLabels: []string{"First month", "Last month"},
				FormatValue: func(_ int, v float64) string { return months[int(v)] },
			})
			if s.selected < int(s.low) || s.selected > int(s.high) {
				s.inspect(-1)
			}
			lo, hi := int(s.low), int(s.high)
			ui.Text(c, fmt.Sprintf("%s – %s", months[lo], months[hi])).TextColor(t.MutedForeground)
			labels := months[lo : hi+1]
			x := chart.NewBand(labels...).Padding(.34, .24)
			y := chart.NewLinear(0, 400)
			y.Format = func(v float64) string { return fmt.Sprintf("$%.0fk", v) }
			ticks := y.Ticks(4)
			xs := make([]float32, len(labels))
			for i := range xs {
				xs[i] = x.At(i)
			}
			axes := chart.Axes{Left: ticks, Bottom: x.Ticks(), Label: func(side chart.Side, tick chart.Tick) {
				label := tick.Label
				if side == chart.Bottom {
					label = label[:3]
				}
				ui.Text(c, label).NoWrap().TextColor(t.MutedForeground)
			}}
			chart.Plot(c, axes, func(area ui.Element) {
				area.Key("monthly-plot").Role(ui.RoleGroup).Label("Monthly revenue plot").Description("Revenue in thousands of dollars. Left and Right inspect months; Escape clears the inspection.")
				gridColor := t.Border.Alpha(.5)
				area.Draw(func(p *ui.Painter, r ui.Rect) {
					for _, tick := range ticks {
						p.Line(r.X, chart.Y(r, tick.At), r.X+r.W, chart.Y(r, tick.At), 1, gridColor)
					}
				})
				current := -1
				if s.selected >= lo && s.selected <= hi {
					current = s.selected - lo
				}
				cleared := area.Shortcut(0, ui.KeyEscape)
				if cleared {
					s.inspect(-1)
					current = -1
				}
				chart.Select(c, area, &current, xs, func(i int) string {
					return fmt.Sprintf("%s: $%.0fk revenue", labels[i], s.revenue[lo+i])
				})
				if current >= 0 && !cleared {
					s.inspect(lo + current)
				}
				// Select clears its transient pointer choice on leave. The app
				// retains an inspection for editing, so keep its accessible value
				// aligned with that same app-owned state on the leave frame too.
				area.Value("")
				if s.selected >= 0 {
					area.Value(fmt.Sprintf("%s: $%.0fk revenue", months[s.selected], s.revenue[s.selected]))
				}
				for i, month := range labels {
					absolute, value := lo+i, s.revenue[lo+i]
					color := barColor
					if s.selected >= 0 && s.selected != absolute {
						color = color.Alpha(.5)
					}
					chart.Rect(c, x.Band(i), y.Span(0, value)).Key(month).Background(color).Radius(4).
						Role(ui.RoleImage).Label(fmt.Sprintf("%s revenue %.0f", month, value))
				}
				chart.RuleY(c, y.At(250), 2).Background(targetColor).AccessHidden(true)
				for i := range labels {
					value := s.revenue[lo+i]
					chart.Anchor(c, x.At(i), y.At(value), ui.Center, ui.End, func() {
						ui.Text(c, fmt.Sprintf("%.0f", value)).NoWrap().Background(t.Background).PaddingX(2).Margin(0, 0, 4, 0).AccessHidden(true)
					})
				}
				chart.Anchor(c, .99, y.At(250), ui.End, ui.End, func() {
					ui.Text(c, "Target $250k").NoWrap().Background(t.Background).TextColor(targetColor).PaddingX(4).Margin(0, 0, 4, 0)
				})
			}).Key("revenue-axes").FillWidth().Height(290).Shrink(0).Gap(12).Padding(12, 0, 0, 0).
				// Native Select uses physical X and Bounds().W. This application
				// deliberately keeps its timeline LTR and untransformed, including
				// in an RTL window. Do not scale/rotate this plot without adapting
				// pointer selection yourself.
				Direction(ui.LeftToRight)
			display.ChartLegendContent(c, display.ChartLegendContentProps{Payload: []display.ChartPayloadItem{
				{DataKey: "revenue", Color: &barColor}, {DataKey: "target", Color: &targetColor},
			}})
			ui.Row(c).FillWidth().Wrap().Gap(18).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Width(220).Gap(8).Children(func() {
					if s.selected < 0 {
						ui.Text(c, "Point at a bar or focus the plot and use the arrow keys.").TextColor(t.MutedForeground)
						return
					}
					indicator := display.ChartIndicatorDot
					if s.indicator == "line" {
						indicator = display.ChartIndicatorLine
					} else if s.indicator == "dashed" {
						indicator = display.ChartIndicatorDashed
					}
					display.ChartTooltipContent(c, display.ChartTooltipContentProps{
						Active: true, Label: &display.ChartContentLabel{Text: months[s.selected]},
						Indicator: indicator, HideIndicator: s.hide,
						Payload: []display.ChartPayloadItem{{Name: "revenue", DataKey: "revenue", Value: s.revenue[s.selected], Color: &barColor}},
					}).Key("inspection").Role(ui.RoleGroup).Label("Inspected month")
				})
				ui.Column(c).Grow(1).Basis(260).MinWidth(180).Gap(8).Children(func() {
					c.Key("revenue-editor")
					control.Input(c, &s.edit, control.InputProps{Label: "Inspected revenue", Placeholder: "Choose a month", Disabled: s.selected < 0, Invalid: s.editError != "", OnSubmit: func(string) { s.applyEdit() }})
					if control.Button(c, "Apply revenue", control.ButtonProps{Variant: control.Outline, Disabled: s.selected < 0}).Clicked() {
						s.applyEdit()
						c.Invalidate()
					}
					if s.editError != "" {
						ui.Text(c, s.editError).TextColor(t.Destructive).Role(ui.RoleAlert)
					}
				})
			})
			c.Key("chart-note")
			control.Textarea(c, &s.note, control.TextareaProps{Label: "Chart note", Rows: 2})
		})
	}).AspectRatio(0)
}

func main() {
	snapshot := flag.String("snapshot", "", "render a PNG instead of opening a window")
	dark := flag.Bool("dark", false, "use the dark chart theme")
	inspect := flag.Bool("inspect", false, "initially inspect March, useful for a snapshot")
	flag.Parse()
	s := newState(*dark, *inspect)
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(f, ui.Render(s.view, 1000, 960, 1)); err != nil {
			f.Close()
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "Chart composition", Width: 1000, Height: 960, Content: ui.View(s.view)})
		w.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
