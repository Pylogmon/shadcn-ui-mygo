package display

import (
	"bytes"
	"fmt"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type compositionLife struct{ closed *int }

func (l *compositionLife) Close() error     { *l.closed++; return nil }
func compositionColor(c ui.Color) *ui.Color { return &c }
func compositionMustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected a clear chart scope/config panic")
		}
	}()
	f()
}
func compositionTester(content func(*ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ChartContainer(c, ChartContainerProps{Config: ChartConfig{}}, func(c *ui.Context) {
			ui.Column(c).Fill().AlignItems(ui.Start).Padding(30).Gap(12).Children(func() { content(c) })
		})
	}, 600, 380)
}

func TestChartCompositionScopeErrorsAndDetachedConfig(t *testing.T) {
	red := ui.RGB(210, 30, 70)
	blue := ui.RGB(30, 70, 210)
	ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		compositionMustPanic(t, func() { ChartConfigOf(c) })
		compositionMustPanic(t, func() { ChartTooltipContent(c, ChartTooltipContentProps{}) })
		compositionMustPanic(t, func() { ChartLegendContent(c, ChartLegendContentProps{}) })
		compositionMustPanic(t, func() {
			ChartContainer(c, ChartContainerProps{Config: ChartConfig{"bad": {Color: &red, Theme: &ChartThemeColors{}}}}, nil)
		})
		ChartContainer(c, ChartContainerProps{Config: ChartConfig{"series": {Label: "Outer", Color: &red}}}, func(c *ui.Context) {
			detached := ChartConfigOf(c)
			*detached["series"].Color = blue
			detached["series"] = ChartConfigItem{Label: "Changed"}
			if got, _ := ChartColor(c, "series"); got != red {
				t.Fatal("ConfigOf mutated the live scope")
			}
			ChartContainer(c, ChartContainerProps{Key: "nested", Config: ChartConfig{"series": {Label: "Inner", Color: &blue}}}, func(c *ui.Context) {
				if got, _ := ChartColor(c, "series"); got != blue {
					t.Fatal("nested config did not replace outer")
				}
			})
			if got, _ := ChartColor(c, "series"); got != red {
				t.Fatal("nested scope did not restore")
			}
			if _, ok := ChartColor(c, "missing"); ok {
				t.Fatal("unknown key invented a color")
			}
		})
		compositionMustPanic(t, func() { ChartConfigOf(c) })
	}, 400, 300)
}

func TestChartCompositionRawNativePlotEditorAndScopedColors(t *testing.T) {
	dark := false
	disabled := false
	draft := "Initial"
	scopeText := "Outer"
	disposed := 0
	red, blue := ui.RGB(210, 30, 70), ui.RGB(30, 70, 210)
	var areaBox, barBox ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		tk := theme.Light()
		if dark {
			tk = theme.Dark()
		}
		theme.Scope(c, tk).Fill().Background(tk.Background).Children(func() {
			ChartContainer(c, ChartContainerProps{Key: "raw", Label: "Raw chart", Config: ChartConfig{"sales": {Label: scopeText, Theme: &ChartThemeColors{Light: red, Dark: blue}}}}, func(c *ui.Context) {
				color, ok := ChartColor(c, "sales")
				if !ok {
					t.Fatal("missing configured raw color")
				}
				ui.Column(c).Fill().Gap(8).Children(func() {
					chart.Plot(c, chart.Axes{Bottom: []chart.Tick{{At: .25, Label: "A"}, {At: .75, Label: "B"}}}, func(area ui.Element) {
						area.Role(ui.RoleGroup).Label("Raw plot").Draw(func(_ *ui.Painter, b ui.Rect) { areaBox = b })
						ui.Local(area, "lifecycle", func() *compositionLife { return &compositionLife{closed: &disposed} })
						chart.Rect(c, chart.Span{Lo: .15, Hi: .4}, chart.Span{Lo: 0, Hi: .75}).Background(color).Draw(func(_ *ui.Painter, b ui.Rect) { barBox = b })
						chart.RuleY(c, .85, 2).Background(ui.RGB(20, 150, 80))
						chart.Anchor(c, .65, .7, ui.Center, ui.Center, func() { ui.Text(c, "Goal") })
					}).FillWidth().Height(180).Direction(ui.LeftToRight)
					ui.TextInput(c, &draft).Label("Native notes")
				})
			}).Disabled(disabled)
		})
	}, 600, 380)
	check := func(want ui.Color) {
		pixel := r.Image().RGBAAt(int(barBox.X+barBox.W/2), int(barBox.Y+barBox.H/2))
		if pixel != (color.RGBA{R: want.R, G: want.G, B: want.B, A: 255}) {
			t.Fatalf("raw configured mark%v want%v area%v", pixel, want, areaBox)
		}
	}
	check(red)
	if !r.HasText("Goal") {
		t.Fatal("caller-owned raw Anchor did not render")
	}
	if err := r.Click("Native notes"); err != nil {
		t.Fatal(err)
	}
	r.Command("selectAll")
	r.Compose("中", 1)
	r.Type("中文 draft")
	if draft != "中文 draft" {
		t.Fatalf("native IME commit=%q", draft)
	}
	r.Command("selectAll")
	dark = true
	scopeText = "Renamed"
	r.Frame()
	r.SetDirection(ui.RightToLeft)
	r.Type("Preserved selection")
	if draft != "Preserved selection" {
		t.Fatalf("theme/scope rebuild reset native selection: %q", draft)
	}
	check(blue)
	if disposed != 0 {
		t.Fatal("stable container rebuilt/disposed its raw plot")
	}
	disabled = true
	r.Frame()
	input, _ := r.AccessByRole(ui.RoleTextField, "Native notes")
	if !input.Disabled {
		t.Fatal("native disabled ancestor did not reach the editor")
	}
	r.Type(" blocked")
	if draft != "Preserved selection" {
		t.Fatal("disabled editor accepted input")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	r.SetFocused(false)
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(time.Second)
	if r.FramePending() || !bytes.Equal(before, r.Image().Pix) {
		t.Fatal("passive composition did not idle")
	}
}

func TestChartCompositionPayloadIndirectionAndCopy(t *testing.T) {
	ui.NewTester(func(c *ui.Context) {
		ChartContainer(c, ChartContainerProps{Config: ChartConfig{"selector": {Label: "Fallback"}, "outer": {Label: "Outer"}, "inner": {Label: "Inner"}, "": {Label: "Empty key"}}}, func(c *ui.Context) {
			cases := []struct {
				name      string
				item      ChartPayloadItem
				key, want string
			}{
				{"top wins", ChartPayloadItem{Fields: map[string]any{"selector": "outer"}, Payload: map[string]any{"selector": "inner"}}, "selector", "Outer"},
				{"datum", ChartPayloadItem{Fields: map[string]any{"selector": 42}, Payload: map[string]any{"selector": "inner"}}, "selector", "Inner"},
				{"unknown indirect fallback", ChartPayloadItem{Fields: map[string]any{"selector": "absent"}}, "selector", "Fallback"},
				{"empty indirect", ChartPayloadItem{Fields: map[string]any{"selector": ""}}, "selector", "Empty key"},
				{"typed name", ChartPayloadItem{Name: "outer"}, "name", "Outer"},
				{"typed datakey", ChartPayloadItem{DataKey: "inner"}, "dataKey", "Inner"},
			}
			for _, test := range cases {
				t.Run(test.name, func(t *testing.T) {
					entry, ok := ChartPayloadConfig(c, test.item, test.key)
					if !ok || entry.Label != test.want {
						t.Fatalf("resolved%+v want%s", entry, test.want)
					}
				})
			}
		})
	}, 400, 250)
}

func TestChartCompositionTooltipNativeFormatterAndPresence(t *testing.T) {
	draft := "Editable"
	calls := []ChartTooltipFormatData{}
	active := true
	value := any([]float64{2, 5})
	r := compositionTester(func(c *ui.Context) {
		ChartTooltipContent(c, ChartTooltipContentProps{Active: active, Payload: []ChartPayloadItem{{Name: "hidden", Value: 999, Type: "none"}, {Name: 7, Value: value, Payload: map[string]any{"unit": "hours"}}, {Name: "Null", HasValue: true}, {Name: "Undefined"}, {Name: 0, Value: 12}},
			Formatter: func(c *ui.Context, data ChartTooltipFormatData) {
				calls = append(calls, data)
				if data.Index == 0 {
					ui.TextInput(c, &draft).Label("Formatted editor")
				} else {
					ui.Text(c, "Explicit null")
				}
			},
		})
	})
	if len(calls) != 2 || calls[0].Index != 0 || calls[0].Name != 7 || calls[0].Payload["unit"] != "hours" || calls[1].Value != nil {
		t.Fatalf("source formatter conditions/filtered index: %+v", calls)
	}
	if !r.HasText("Explicit null") || !r.HasText("12") || r.HasText("999") {
		t.Fatal("null/zero-name/none default rendering changed")
	}
	if err := r.Click("Formatted editor"); err != nil {
		t.Fatal(err)
	}
	r.Command("selectAll")
	value = "updated"
	r.Frame()
	r.Type("Native state")
	if draft != "Native state" {
		t.Fatalf("custom native formatter editor lost focus/selection: %q", draft)
	}
	calls = nil
	active = false
	r.Frame()
	if len(calls) != 0 || r.HasText("Explicit null") {
		t.Fatal("inactive tooltip built content")
	}
}

func TestChartCompositionTooltipDefaultValuesAndNativeLabels(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ChartContainer(c, ChartContainerProps{Config: ChartConfig{"series": {Label: "Configured", LabelContent: func(c *ui.Context) { ui.Text(c, "Native config label") }}, "empty": {HasLabel: true}}}, func(c *ui.Context) {
			ui.Column(c).Fill().AlignItems(ui.Start).Padding(20).Children(func() {
				ChartTooltipContent(c, ChartTooltipContentProps{Active: true, Label: &ChartContentLabel{Text: "Heading"}, Payload: []ChartPayloadItem{{Name: "series", Value: 12345.6789}, {Name: "Array", Value: []any{2, "five", nil}}, {Name: "String", Value: "Ready"}, {Name: "empty", Value: 1}, {Name: "No value"}}})
			})
		})
	}, 600, 400)
	r.SetLocale("en-US")
	for _, text := range []string{"Heading", "Native config label", "12,345.679", "2,five,", "Ready", "No value"} {
		if !r.HasText(text) {
			t.Fatalf("missing source payload text %q: %v", text, r.Texts())
		}
	}
	if r.HasText("empty") {
		t.Fatal("explicit empty config label fell back to payload name")
	}
	r.SetLocale("de-DE")
	if !r.HasText("12.345,679") {
		t.Fatalf("numeric formatting did not follow native locale: %v", r.Texts())
	}
}

func TestChartCompositionTooltipIndicatorGeometryAndNestedHeading(t *testing.T) {
	for _, indicator := range []ChartIndicator{ChartIndicatorDot, ChartIndicatorLine, ChartIndicatorDashed} {
		t.Run(fmt.Sprint(indicator), func(t *testing.T) {
			red := ui.RGB(210, 30, 40)
			var bounds ui.Rect
			r := compositionTester(func(c *ui.Context) {
				ChartTooltipContent(c, ChartTooltipContentProps{Active: true, Indicator: indicator, Color: &red, Label: &ChartContentLabel{Text: "Heading"}, Payload: []ChartPayloadItem{{Name: "Desktop", Value: 42, Color: compositionColor(ui.RGB(10, 180, 60))}}}).Draw(func(_ *ui.Painter, b ui.Rect) { bounds = b })
			})
			heading, _ := r.Find("Heading")
			name, _ := r.Find("Desktop")
			value, _ := r.Find("42")
			if heading.Y >= name.Y || name.Y != value.Y {
				t.Fatalf("source heading/value alignment: heading%v name%v value%v", heading, name, value)
			}
			minX, minY, maxX, maxY := int(bounds.X+bounds.W), int(bounds.Y+bounds.H), 0, 0
			painted := 0
			image := r.Image()
			for y := int(bounds.Y); y < int(bounds.Y+bounds.H); y++ {
				for x := int(bounds.X); x < int(bounds.X+bounds.W); x++ {
					pixel := image.RGBAAt(x, y)
					if pixel.R > 180 && pixel.G < 60 && pixel.B < 70 {
						painted++
						minX = min(minX, x)
						maxX = max(maxX, x)
						minY = min(minY, y)
						maxY = max(maxY, y)
					}
				}
			}
			if painted == 0 {
				t.Fatal("source indicator painted no configured pixels")
			}
			width, height := maxX-minX+1, maxY-minY+1
			switch indicator {
			case ChartIndicatorDot:
				if width > 10 || height > 10 || height < 8 {
					t.Fatalf("dot%dx%d", width, height)
				}
			case ChartIndicatorLine:
				if width != 4 || height < 25 {
					t.Fatalf("line%dx%d did not stretch across nested heading", width, height)
				}
			case ChartIndicatorDashed:
				if width > 3 || height < 20 || painted >= width*height {
					t.Fatalf("dashed %dx%d painted%d is not a broken vertical indicator", width, height, painted)
				}
			}
		})
	}
}

func TestChartCompositionTooltipRawPayloadLengthControlsNesting(t *testing.T) {
	spans := []int{}
	red := ui.RGB(210, 30, 40)
	for _, payload := range [][]ChartPayloadItem{{{Name: "Desktop", Value: 42}}, {{Name: "Ignored", Type: "none"}, {Name: "Desktop", Value: 42}}} {
		r := compositionTester(func(c *ui.Context) {
			ChartTooltipContent(c, ChartTooltipContentProps{Active: true, Indicator: ChartIndicatorLine, Color: &red, Label: &ChartContentLabel{Text: "Heading"}, Payload: payload})
		})
		image := r.Image()
		low, high := image.Bounds().Max.Y, 0
		for y := 0; y < image.Bounds().Max.Y; y++ {
			for x := 0; x < image.Bounds().Max.X; x++ {
				pixel := image.RGBAAt(x, y)
				if pixel.R == red.R && pixel.G == red.G && pixel.B == red.B {
					low = min(low, y)
					high = max(high, y)
				}
			}
		}
		spans = append(spans, high-low+1)
	}
	// A single raw payload nests its heading inside the row, so the line
	// stretches alongside both heading and name. A preceding none payload
	// still counts for nesting and leaves the line alongside the name only.
	if spans[0] <= spans[1] || spans[1] < 12 {
		t.Fatalf("raw payload len must decide nesting before filtering: line spans%v", spans)
	}
}

func TestChartCompositionIconsHideSemanticsAndLegendAlignment(t *testing.T) {
	solid := ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><path fill="currentColor" d="M0 0h10v10H0z"/></svg>`))
	for _, align := range []ChartLegendAlign{ChartLegendBottom, ChartLegendTop} {
		for _, hideIcon := range []bool{false, true} {
			t.Run(fmt.Sprint(align, hideIcon), func(t *testing.T) {
				calls := 0
				var iconBounds, legendBounds ui.Rect
				payloadColor := ui.RGB(210, 30, 40)
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					ChartContainer(c, ChartContainerProps{Config: ChartConfig{"desktop": {Label: "Desktop", Icon: func(c *ui.Context) ui.Element {
						calls++
						return ui.Icon(c, solid).Draw(func(_ *ui.Painter, b ui.Rect) { iconBounds = b })
					}}}}, func(c *ui.Context) {
						ui.Column(c).Fill().Padding(30).AlignItems(ui.Start).Gap(10).Children(func() {
							ChartTooltipContent(c, ChartTooltipContentProps{Active: true, HideIndicator: true, Payload: []ChartPayloadItem{{Name: "desktop", Value: 1, Color: &payloadColor}}})
							ChartLegendContent(c, ChartLegendContentProps{VerticalAlign: align, HideIcon: hideIcon, Payload: []ChartPayloadItem{{DataKey: "desktop", Color: &payloadColor}, {DataKey: "ignored", Type: "none"}}}).Draw(func(_ *ui.Painter, b ui.Rect) { legendBounds = b })
						})
					})
				}, 400, 300)
				if hideIcon {
					if calls != 1 || iconBounds.W != 10 || iconBounds.H != 10 {
						t.Fatalf("hideIndicator must retain tooltip icon, hideIcon removes only legend icon: calls%d box%v", calls, iconBounds)
					}
				} else {
					if calls != 2 || iconBounds.W != 12 || iconBounds.H != 12 {
						t.Fatalf("legend source icon size: calls%d box%v", calls, iconBounds)
					}
				}
				if pixel := r.Image().RGBAAt(int(iconBounds.X+iconBounds.W/2), int(iconBounds.Y+iconBounds.H/2)); pixel.R != theme.Light().MutedForeground.R {
					t.Fatalf("icon did not inherit muted foreground: %v", pixel)
				}
				for _, node := range r.AccessAll() {
					if node.Role == ui.RoleText && node.Name == "Desktop" && node.Bounds.Y >= legendBounds.Y {
						wantY := legendBounds.Y
						if align == ChartLegendBottom {
							wantY += 12
						}
						if node.Bounds.Y != wantY {
							t.Fatalf("legend source side padding: text%v root%v", node.Bounds, legendBounds)
						}
					}
				}
			})
		}
	}
}

func TestChartCompositionLabelFormatterAndIndependentKeys(t *testing.T) {
	var received ChartContentLabel
	r := ui.NewTester(func(c *ui.Context) {
		ChartContainer(c, ChartContainerProps{Config: ChartConfig{"month": {Label: "June"}, "browser": {Label: "Chrome"}}}, func(c *ui.Context) {
			ChartTooltipContent(c, ChartTooltipContentProps{Active: true, NameKey: "browserKey", LabelKey: "monthKey", Payload: []ChartPayloadItem{{Name: "Fallback", Value: 18, Payload: map[string]any{"browserKey": "browser", "monthKey": "month"}}}, LabelFormatter: func(c *ui.Context, label ChartContentLabel, payload []ChartPayloadItem) {
				received = label
				ui.Text(c, label.Text+" report")
			}})
		})
	}, 400, 280)
	if received.Text != "June" || !received.Present || !r.HasText("June report") || !r.HasText("Chrome") {
		t.Fatalf("independent label/name keys: %+v %v", received, r.Texts())
	}
}

func TestChartCompositionOptionalPayloadKeysAndEmptyLabels(t *testing.T) {
	for _, test := range []struct {
		name string
		item ChartPayloadItem
		want string
	}{
		{"name only", ChartPayloadItem{Name: "desktop", Value: 1}, "Desktop label"},
		{"no keys", ChartPayloadItem{Value: 1}, "Default value"},
		{"explicit empty datakey", ChartPayloadItem{Name: "desktop", DataKey: "", Value: 1}, "Empty key"},
		{"named datakey", ChartPayloadItem{Name: "desktop", DataKey: "visitors", Value: 1}, "Visitor label"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var heading ChartContentLabel
			r := ui.NewTester(func(c *ui.Context) {
				ChartContainer(c, ChartContainerProps{Config: ChartConfig{"desktop": {Label: "Desktop label"}, "value": {Label: "Default value"}, "": {Label: "Empty key"}, "visitors": {Label: "Visitor label"}}}, func(c *ui.Context) {
					ChartTooltipContent(c, ChartTooltipContentProps{Active: true, Payload: []ChartPayloadItem{test.item}, LabelFormatter: func(c *ui.Context, label ChartContentLabel, _ []ChartPayloadItem) { heading = label; label.Render(c) }})
				})
			}, 400, 280)
			if heading.Text != test.want || !r.HasText(test.want) {
				t.Fatalf("heading lookup=%+v want%s", heading, test.want)
			}
		})
	}
	ui.NewTester(func(c *ui.Context) {
		ChartContainer(c, ChartContainerProps{Config: ChartConfig{"inner": {Label: "Resolved datum"}}}, func(c *ui.Context) {
			for _, key := range []string{"dataKey", "type"} {
				entry, ok := ChartPayloadConfig(c, ChartPayloadItem{Payload: map[string]any{key: "inner"}}, key)
				if !ok || entry.Label != "Resolved datum" {
					t.Fatalf("omitted top-level %s blocked original datum: %+v", key, entry)
				}
			}
		})
	}, 300, 200)
}

func TestChartCompositionContainerDefaultsAndInheritedNativeContent(t *testing.T) {
	var bounds ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Font("serif").FontWeight(700).TextColor(ui.RGB(180, 30, 80)).Children(func() {
			ChartContainer(c, ChartContainerProps{Label: "Unstyled composition"}, func(c *ui.Context) { ui.Text(c, "Native inheritance") }).Draw(func(_ *ui.Painter, b ui.Rect) { bounds = b })
		})
	}, 320, 260)
	if bounds.W != 320 || bounds.H != 180 {
		t.Fatalf("source aspect-video default=%v", bounds)
	}
	label, _ := r.Find("Native inheritance")
	if label.H != 16 {
		t.Fatalf("source text-xs/line-height=%v", label)
	}
	reference := ui.NewTester(func(c *ui.Context) {
		ui.Text(c, "Native inheritance").Font("serif").FontSize(12).FontWeight(700).LineHeight(4.0/3).TextColor(ui.RGB(180, 30, 80)).Absolute().Left(label.X).Top(label.Y).Size(label.W, label.H)
	}, 320, 260)
	actualImage, refImage := r.Image(), reference.Image()
	for y := int(label.Y); y < int(label.Y+label.H); y++ {
		for x := int(label.X); x < int(label.X+label.W); x++ {
			if refImage.RGBAAt(x, y).A != 0 && refImage.RGBAAt(x, y) != actualImage.RGBAAt(x, y) {
				t.Fatalf("container changed inherited native foreground/family/weight at%d,%d", x, y)
			}
		}
	}
}

func TestChartCompositionLabelVisibilityAndAllNonePayload(t *testing.T) {
	for _, hide := range []bool{false, true} {
		t.Run(fmt.Sprint(hide), func(t *testing.T) {
			calls := 0
			var size ui.Rect
			r := compositionTester(func(c *ui.Context) {
				ChartTooltipContent(c, ChartTooltipContentProps{Active: true, HideLabel: hide, Payload: []ChartPayloadItem{{Name: "Removed", Type: "none", Value: 1}}, LabelFormatter: func(c *ui.Context, label ChartContentLabel, _ []ChartPayloadItem) {
					calls++
					if label.Present {
						t.Fatal("absent label became present")
					}
					ui.Text(c, "Only heading")
				}}).Draw(func(_ *ui.Painter, b ui.Rect) { size = b })
			})
			if hide {
				if calls != 0 || r.HasText("Only heading") {
					t.Fatal("HideLabel invoked formatter")
				}
			} else if calls != 1 || !r.HasText("Only heading") {
				t.Fatal("all-none payload incorrectly treated as an empty input payload")
			}
			if size.W != 128 {
				t.Fatalf("nonempty original payload lost source min-width: %v", size)
			}
		})
	}
	var tooltip, legend ui.Rect
	compositionTester(func(c *ui.Context) {
		ChartTooltipContent(c, ChartTooltipContentProps{Active: true}).Draw(func(_ *ui.Painter, b ui.Rect) { tooltip = b })
		ChartLegendContent(c, ChartLegendContentProps{}).Draw(func(_ *ui.Painter, b ui.Rect) { legend = b })
	})
	if tooltip.W != 0 || tooltip.H != 0 || legend.W != 0 || legend.H != 0 {
		t.Fatalf("empty payload reserves layout: tooltip%v legend%v", tooltip, legend)
	}
}

func TestChartCompositionNativeFormatterSurvivesHeaderAndIndicatorLayout(t *testing.T) {
	draft := "Native draft"
	hide := false
	indicator := ChartIndicatorDot
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ChartContainer(c, ChartContainerProps{}, func(c *ui.Context) {
			ChartTooltipContent(c, ChartTooltipContentProps{Active: true, Label: &ChartContentLabel{Text: "Heading"}, HideLabel: hide, Indicator: indicator, Payload: []ChartPayloadItem{{Name: "Metric", Value: 1}}, Formatter: func(c *ui.Context, _ ChartTooltipFormatData) {
				ui.TextInputBase(c, &draft).Size(180, 30).Label("Native formatter editor")
			}})
		})
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Native formatter editor"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	hide = true
	r.Frame()
	if !r.Focused("Native formatter editor") {
		t.Fatal("HideLabel discarded the real formatter editor focus")
	}
	r.Type("After hidden heading")
	if draft != "After hidden heading" {
		t.Fatalf("heading layout reset real native selection: %q", draft)
	}
	r.Key(ui.Cmd, ui.KeyA)
	hide = false
	indicator = ChartIndicatorLine
	r.Frame()
	if !r.Focused("Native formatter editor") {
		t.Fatal("indicator nesting discarded the real formatter editor focus")
	}
	r.Type("After nesting")
	if draft != "After nesting" {
		t.Fatalf("indicator layout reset real native selection: %q", draft)
	}
}

func TestChartCompositionNativeConfigLabelSurvivesIndicatorVisibility(t *testing.T) {
	draft := "Config editor"
	hide := false
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ChartContainer(c, ChartContainerProps{Config: ChartConfig{"metric": {LabelContent: func(c *ui.Context) { ui.TextInputBase(c, &draft).Size(160, 30).Label("Native config editor") }}}}, func(c *ui.Context) {
			ChartTooltipContent(c, ChartTooltipContentProps{Active: true, HideLabel: true, HideIndicator: hide, Payload: []ChartPayloadItem{{Name: "metric", Value: 7}}})
		})
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Native config editor"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	hide = true
	r.Frame()
	if !r.Focused("Native config editor") {
		t.Fatal("HideIndicator discarded native config label editor focus")
	}
	r.Type("After indicator hidden")
	if draft != "After indicator hidden" {
		t.Fatalf("indicator visibility reset native editor selection: %q", draft)
	}
}

func TestChartCompositionNativeLegendLabelSurvivesEmptyIconAndAlignment(t *testing.T) {
	draft := "Legend draft"
	hideIcon := false
	align := ChartLegendBottom
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ChartContainer(c, ChartContainerProps{Config: ChartConfig{"metric": {Icon: func(*ui.Context) ui.Element { return ui.Element{} }, LabelContent: func(c *ui.Context) { ui.TextInputBase(c, &draft).Size(160, 30).Label("Legend editor") }}}}, func(c *ui.Context) {
			ChartLegendContent(c, ChartLegendContentProps{HideIcon: hideIcon, VerticalAlign: align, Payload: []ChartPayloadItem{{DataKey: "metric", Color: compositionColor(ui.RGB(30, 120, 180))}}})
		})
	}, 500, 300)
	if err := r.Click("Legend editor"); err != nil {
		t.Fatal(err)
	}
	r.Command("selectAll")
	hideIcon = true
	align = ChartLegendTop
	r.Frame()
	if !r.Focused("Legend editor") {
		t.Fatal("source empty-icon slot or side padding reset native label focus")
	}
	r.Type("Kept selection")
	if draft != "Kept selection" {
		t.Fatalf("native legend editor selection reset: %q", draft)
	}
}

func TestChartCompositionExplicitEmptySelectorKeys(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprint(explicit), func(t *testing.T) {
			var heading ChartContentLabel
			r := ui.NewTester(func(c *ui.Context) {
				config := ChartConfig{"ordinary": {Label: "Default heading"}, "name": {Label: "Default name"}, "outer": {Label: "Outer selector"}, "inner": {Label: "Inner selector"}}
				ChartContainer(c, ChartContainerProps{Config: config}, func(c *ui.Context) {
					ui.Column(c).Fill().AlignItems(ui.Start).Children(func() {
						payload := []ChartPayloadItem{{Name: "name", DataKey: "ordinary", Value: 1, Fields: map[string]any{"": "outer"}, Payload: map[string]any{"": "inner"}}}
						ChartTooltipContent(c, ChartTooltipContentProps{Active: true, Payload: payload, HasNameKey: explicit, HasLabelKey: explicit, LabelFormatter: func(c *ui.Context, label ChartContentLabel, _ []ChartPayloadItem) { heading = label; label.Render(c) }})
						ChartLegendContent(c, ChartLegendContentProps{Payload: payload, HasNameKey: explicit})
					})
				})
			}, 400, 300)
			want := "Default heading"
			if explicit {
				want = "Outer selector"
			}
			if heading.Text != want || !r.HasText(want) {
				t.Fatalf("explicit-empty %v heading%+v want%s", explicit, heading, want)
			}
			if explicit && r.HasText("Default name") {
				t.Fatal("explicit empty NameKey was treated as omitted")
			}
			if !explicit && !r.HasText("Default name") {
				t.Fatal("omitted NameKey did not use name fallback")
			}
		})
	}
}

func TestChartCompositionEmptyContentAddsNoNativeLayoutGap(t *testing.T) {
	for _, legend := range []bool{false, true} {
		name := "tooltip"
		if legend {
			name = "legend"
		}
		t.Run(name, func(t *testing.T) {
			r := ui.NewTester(func(c *ui.Context) {
				ChartContainer(c, ChartContainerProps{}, func(c *ui.Context) {
					ui.Column(c).Gap(10).Children(func() {
						ui.Box(c).Size(20, 20).Label("Before empty")
						if legend {
							ChartLegendContent(c, ChartLegendContentProps{})
						} else {
							ChartTooltipContent(c, ChartTooltipContentProps{Active: false})
						}
						ui.Box(c).Size(20, 20).Label("After empty")
					})
				})
			}, 400, 300)
			before, _ := r.Find("Before empty")
			after, _ := r.Find("After empty")
			if after.Y-before.Y != 30 {
				t.Fatalf("source null content introduced a real native flex gap: before%v after%v", before, after)
			}
		})
	}
}
