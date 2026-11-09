package display

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

// ChartConfig is caller-owned metadata for independently composed native plots.
// The container snapshots it for its synchronous Children build. No CSS or
// palette is injected into raw marks: resolve ChartColor while building and
// pass that color value to Plot/marks or a Draw closure.
type ChartConfig map[string]ChartConfigItem

type ChartThemeColors struct{ Light, Dark ui.Color }

type ChartConfigItem struct {
	Label string
	// HasLabel preserves an explicitly empty config label; a non-empty Label
	// or LabelContent is present automatically. Empty differs from absent
	// when a tooltip row falls back to the payload name.
	HasLabel bool
	// LabelContent replaces Label with arbitrary native content.
	LabelContent func(*ui.Context)
	// Icon returns the actual native icon. Content helpers apply source size and
	// muted foreground to it; the callback runs only during the current build.
	Icon func(*ui.Context) ui.Element
	// Color and Theme are mutually exclusive. Pointers preserve an explicit
	// transparent color, including the all-zero ui.Color value.
	Color *ui.Color
	Theme *ChartThemeColors
}

type ChartContainerProps struct {
	Key, Label string
	Config     ChartConfig
}

type chartCompositionScopeKey struct{}
type chartCompositionScope struct {
	active bool
	config ChartConfig
}
type chartCompositionIdentity string

// Optional headings and indicators must not move native editors to another
// positional identity. These slots mirror React's retained conditional slots.
type chartCompositionPart uint8

const (
	chartTooltipHeadingPart chartCompositionPart = iota
	chartTooltipPayloadPart
	chartTooltipContentPart
	chartTooltipLabelsPart
	chartTooltipLabelPart
	chartTooltipIconPart
	chartTooltipIndicatorPart
	chartTooltipValuePart
	chartLegendIconPart
	chartLegendLabelPart
)

type chartCompositionRowKey struct{ Index int }

func chartCopyConfigItem(item ChartConfigItem) ChartConfigItem {
	if item.Color != nil {
		color := *item.Color
		item.Color = &color
	}
	if item.Theme != nil {
		colors := *item.Theme
		item.Theme = &colors
	}
	return item
}
func chartCopyConfig(config ChartConfig) ChartConfig {
	copy := make(ChartConfig, len(config))
	for key, item := range config {
		copy[key] = chartCopyConfigItem(item)
	}
	return copy
}
func chartCurrentConfig(c *ui.Context) ChartConfig {
	scope := ui.Local(c.Root(), chartCompositionScopeKey{}, func() chartCompositionScope { return chartCompositionScope{} })
	if scope == nil || !scope.active {
		panic("display: chart composition requires ChartContainer Children")
	}
	return scope.config
}

// ChartConfigOf returns a detached copy of the current build's configuration.
// It panics outside ChartContainer.Children, including in deferred Draw code.
// Do not retain Context; capture the resolved values your draw callback needs.
func ChartConfigOf(c *ui.Context) ChartConfig { return chartCopyConfig(chartCurrentConfig(c)) }

// ChartColor resolves an entry using the innermost live native theme. A missing
// entry or an entry without a color returns false; no palette is invented.
func ChartColor(c *ui.Context, key string) (ui.Color, bool) {
	item, ok := chartCurrentConfig(c)[key]
	if !ok {
		return ui.Color{}, false
	}
	if item.Theme != nil {
		if c.Theme() != nil && c.Theme().Dark {
			return item.Theme.Dark, true
		}
		return item.Theme.Light, true
	}
	if item.Color != nil {
		return *item.Color, true
	}
	return ui.Color{}, false
}

// ChartContainer scopes metadata around arbitrary native content, without
// choosing axes, marks, interaction, or data geometry. The source 16:9 ratio,
// centered row and 12-DIP type are defaults, overridable through fluent native
// layout. Native Plot/Rule/Anchor, controls and editors remain real children.
// Config helpers are valid only in this synchronous Children call, including
// nested native Children calls; nested/sibling containers restore the scope. Replacing the returned element's
// Children later does not reinstall this scope; use this callback as the build
// entry point. Late fluent styles do not re-run callbacks or color resolution.
func ChartContainer(c *ui.Context, p ChartContainerProps, children func(*ui.Context)) ui.Element {
	config := chartCopyConfig(p.Config)
	for key, item := range config {
		if item.Color != nil && item.Theme != nil {
			panic(fmt.Sprintf("display: ChartConfig[%q] has both Color and Theme", key))
		}
	}
	root := ui.Row(c).FillWidth().MinWidth(0).AspectRatio(16.0 / 9).Justify(ui.Center).FontSize(12).LineHeight(4.0 / 3)
	if p.Key != "" {
		root.Key(chartCompositionIdentity(p.Key))
	}
	if p.Label != "" {
		root.Role(ui.RoleGroup).Label(p.Label)
	}
	return root.Children(func() {
		scope := ui.Local(c.Root(), chartCompositionScopeKey{}, func() chartCompositionScope { return chartCompositionScope{} })
		previous := *scope
		*scope = chartCompositionScope{active: true, config: config}
		defer func() { *scope = previous }()
		if children != nil {
			children(c)
		}
	})
}

// ChartPayloadItem is the native counterpart of a Recharts payload entry.
// Non-nil Value is present automatically. HasValue distinguishes an explicit
// null (nil+true) from undefined (nil+false) for Formatter; neither paints a
// default value span. The default renderer handles numbers/strings and joins
// []any, []string, []float64 and []int like JavaScript Array.toString. Other Go
// values use fmt.Sprint; supply Formatter for another representation.
// Name/DataKey support string or numeric keys; nil means absent and "" is
// an explicit empty key. Type "none" is filtered.
// Fields holds additional top-level fields used for config indirection;
// Payload holds the original datum, including optional native ui.Color fill.
type ChartPayloadItem struct {
	Name, DataKey   any
	Type            string
	Value           any
	HasValue        bool
	Color           *ui.Color
	Fields, Payload map[string]any
}

func chartPayloadField(item ChartPayloadItem, key string) any {
	switch key {
	case "name":
		return item.Name
	case "dataKey":
		return item.DataKey
	case "type":
		if item.Type != "" {
			return item.Type
		}
		return item.Fields[key]
	case "value":
		return item.Value
	case "color":
		return item.Color
	case "payload":
		return item.Payload
	}
	return item.Fields[key]
}
func chartResolvePayload(config ChartConfig, item ChartPayloadItem, key string) (ChartConfigItem, bool) {
	resolved := key
	if value, ok := chartPayloadField(item, key).(string); ok {
		resolved = value
	} else if value, ok := item.Payload[key].(string); ok {
		resolved = value
	}
	if entry, ok := config[resolved]; ok {
		return entry, true
	}
	entry, ok := config[key]
	return entry, ok
}

// ChartPayloadConfig follows the source's top-level string -> original datum
// string -> original key resolution, with fallback to config[original key]
// when the indirect key is absent. It returns a detached entry copy.
func ChartPayloadConfig(c *ui.Context, item ChartPayloadItem, key string) (ChartConfigItem, bool) {
	entry, ok := chartResolvePayload(chartCurrentConfig(c), item, key)
	return chartCopyConfigItem(entry), ok
}

// ChartContentLabel preserves native label content for LabelFormatter. Render
// builds it now; neither a Context nor an Element is stored in this value.
type ChartContentLabel struct {
	Text    string
	Content func(*ui.Context)
	// Present distinguishes an empty label value from an absent value in
	// LabelFormatter. Non-empty Text/Content are present automatically.
	Present bool
}

func (label ChartContentLabel) Render(c *ui.Context) {
	if label.Content != nil {
		label.Content(c)
	} else if label.Text != "" {
		ui.Text(c, label.Text)
	}
}
func (label ChartContentLabel) present() bool { return label.Text != "" || label.Content != nil }
func chartEntryLabel(entry ChartConfigItem) ChartContentLabel {
	return ChartContentLabel{Text: entry.Label, Content: entry.LabelContent, Present: entry.HasLabel || entry.Label != "" || entry.LabelContent != nil}
}

type ChartIndicator uint8

const (
	ChartIndicatorDot ChartIndicator = iota
	ChartIndicatorLine
	ChartIndicatorDashed
)

type ChartTooltipFormatData struct {
	Value, Name any
	Item        ChartPayloadItem
	// Index is the filtered payload index, exactly as in the source.
	Index   int
	Payload map[string]any
}
type ChartTooltipContentProps struct {
	Active  bool
	Payload []ChartPayloadItem
	// Nil omits the label prop. Text (even "") follows the source string-label
	// lookup. A native Content label follows the source's non-string prop path:
	// the resolved first item's config supplies its heading instead.
	Label                    *ChartContentLabel
	HideLabel, HideIndicator bool
	Indicator                ChartIndicator
	NameKey, LabelKey        string
	// Non-empty keys are present automatically. These flags distinguish an
	// explicitly empty string key from the omitted key (source ?? semantics).
	HasNameKey, HasLabelKey bool
	Color                   *ui.Color
	// Formatter replaces a row only for a present value and a truthy name.
	// Null remains present when HasValue is true. Native controls are allowed.
	Formatter      func(*ui.Context, ChartTooltipFormatData)
	LabelFormatter func(*ui.Context, ChartContentLabel, []ChartPayloadItem)
}

func chartPayloadString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case []any:
		parts := make([]string, len(v))
		for i, x := range v {
			parts[i] = chartPayloadString(x)
		}
		return strings.Join(parts, ",")
	case []string:
		return strings.Join(v, ",")
	case []float64:
		parts := make([]string, len(v))
		for i, x := range v {
			parts[i] = fmt.Sprint(x)
		}
		return strings.Join(parts, ",")
	case []int:
		parts := make([]string, len(v))
		for i, x := range v {
			parts[i] = fmt.Sprint(x)
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprint(v)
	}
}
func chartPayloadTruthy(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case string:
		return v != ""
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float64:
		return v != 0 && v == v
	case float32:
		return v != 0 && v == v
	case bool:
		return v
	}
	return true
}
func chartPayloadKey(explicit string, present bool, values ...any) string {
	if present || explicit != "" {
		return explicit
	}
	for _, value := range values {
		if value != nil {
			return chartPayloadString(value)
		}
	}
	return "value"
}
func chartDefaultPayloadValue(c *ui.Context, value any) string {
	switch value.(type) {
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		locale := c.Region().FormatLocale(c.Locale())
		return message.NewPrinter(language.Make(locale)).Sprint(number.Decimal(value, number.MaxFractionDigits(3)))
	default:
		return chartPayloadString(value)
	}
}
func chartPayloadColor(item ChartPayloadItem, override *ui.Color) ui.Color {
	if override != nil {
		return *override
	}
	switch color := item.Payload["fill"].(type) {
	case ui.Color:
		return color
	case *ui.Color:
		if color != nil {
			return *color
		}
	}
	if item.Color != nil {
		return *item.Color
	}
	return ui.Color{}
}
func chartCompositionEmpty(c *ui.Context) ui.Element {
	return ui.Box(c).Size(0, 0).Hidden(true).IgnoreInput(true).AccessHidden(true)
}

// ChartTooltipContent builds the fixed source's composable tooltip content.
// The caller owns visibility/position and supplies the inspected payload; this
// helper does not invent a hover manager or bind to the high-level Chart.
func ChartTooltipContent(c *ui.Context, p ChartTooltipContentProps) ui.Element {
	config := chartCurrentConfig(c) // Source useChart runs even when inactive.
	if !p.Active || len(p.Payload) == 0 {
		return chartCompositionEmpty(c)
	}
	t := theme.TokensOf(c)
	var heading ChartContentLabel
	if !p.HideLabel {
		first := p.Payload[0]
		key := chartPayloadKey(p.LabelKey, p.HasLabelKey, first.DataKey, first.Name)
		entry, _ := chartResolvePayload(config, first, key)
		heading = chartEntryLabel(entry)
		if p.LabelKey == "" && p.Label != nil && p.Label.Content == nil {
			if entry, ok := config[p.Label.Text]; ok && chartEntryLabel(entry).Present {
				heading = chartEntryLabel(entry)
			} else {
				heading = *p.Label
				heading.Present = true
			}
		}
	}
	showHeading := !p.HideLabel && (heading.present() || p.LabelFormatter != nil)
	renderHeading := func(child *ui.Context) {
		ui.Column(child).Key(chartTooltipHeadingPart).FontWeight(500).Children(func() {
			if p.LabelFormatter != nil {
				p.LabelFormatter(child, heading, p.Payload)
			} else {
				heading.Render(child)
			}
		})
	}
	nest := len(p.Payload) == 1 && p.Indicator != ChartIndicatorDot
	return ui.Column(c).MinWidth(128).Gap(6).Padding(6, 10).Radius(t.RadiusLG()).FontSize(12).LineHeight(4.0/3).Background(t.Background).Border(1, t.Border.Alpha(.5)).
		Shadow(0, 20, 25, -5, ui.RGBA(0, 0, 0, .1)).Shadow(0, 8, 10, -6, ui.RGBA(0, 0, 0, .1)).Children(func() {
		if !nest && showHeading {
			renderHeading(c)
		}
		ui.Column(c).Key(chartTooltipPayloadPart).Gap(6).Children(func() {
			index := 0
			for _, item := range p.Payload {
				if item.Type == "none" {
					continue
				}
				entry, _ := chartResolvePayload(config, item, chartPayloadKey(p.NameKey, p.HasNameKey, item.Name, item.DataKey))
				rowIndex := index
				index++
				row := ui.Row(c).Key(chartCompositionRowKey{Index: rowIndex}).FillWidth().Wrap().Gap(8).AlignItems(ui.Stretch)
				if p.Indicator == ChartIndicatorDot {
					row.AlignItems(ui.Center)
				}
				row.Children(func() {
					if p.Formatter != nil && (item.HasValue || item.Value != nil) && chartPayloadTruthy(item.Name) {
						p.Formatter(c, ChartTooltipFormatData{Value: item.Value, Name: item.Name, Item: item, Index: rowIndex, Payload: item.Payload})
						return
					}
					if entry.Icon != nil {
						entry.Icon(c.Key(chartTooltipIconPart)).Size(10, 10).Shrink(0).TextColor(t.MutedForeground)
					} else if !p.HideIndicator {
						color := chartPayloadColor(item, p.Color)
						switch p.Indicator {
						case ChartIndicatorLine:
							ui.Box(c).Key(chartTooltipIndicatorPart).Width(4).Shrink(0).Radius(2).Background(color).AccessHidden(true)
						case ChartIndicatorDashed:
							line := ui.Box(c).Key(chartTooltipIndicatorPart).Width(3).Shrink(0).Radius(2).Border(1.5, color).BorderStyle(ui.BorderDashed).AccessHidden(true)
							if nest {
								line.MarginY(2)
							}
						default:
							ui.Box(c).Key(chartTooltipIndicatorPart).Size(10, 10).Shrink(0).Radius(2).Background(color).AccessHidden(true)
						}
					}
					content := ui.Row(c).Key(chartTooltipContentPart).Grow(1).MinWidth(0).Justify(ui.SpaceBetween).AlignItems(ui.Center).LineHeight(1)
					if nest {
						content.AlignItems(ui.End)
					}
					content.Children(func() {
						ui.Column(c).Key(chartTooltipLabelsPart).Gap(6).Children(func() {
							if nest && showHeading {
								renderHeading(c)
							}
							label := chartEntryLabel(entry)
							ui.Column(c).Key(chartTooltipLabelPart).TextColor(t.MutedForeground).Children(func() {
								if label.Present {
									label.Render(c)
								} else if item.Name != nil {
									ui.Text(c, chartPayloadString(item.Name))
								}
							})
						})
						if item.Value != nil {
							ui.Text(c, chartDefaultPayloadValue(c, item.Value)).Key(chartTooltipValuePart).Font("monospace").FontWeight(500).FontFeatures("tnum").TextColor(t.Foreground)
						}
					})
				})
			}
		})
	})
}

type ChartLegendAlign uint8

const (
	ChartLegendBottom ChartLegendAlign = iota
	ChartLegendTop
)

type ChartLegendContentProps struct {
	Payload  []ChartPayloadItem
	HideIcon bool
	NameKey  string
	// HasNameKey explicitly selects the empty string key.
	HasNameKey    bool
	VerticalAlign ChartLegendAlign
}

// ChartLegendContent renders config labels/icons and payload colors. Missing
// labels remain absent, as in the source. Native item wrapping is retained so
// a narrow legend keeps whole items readable instead of overflowing its plot.
func ChartLegendContent(c *ui.Context, p ChartLegendContentProps) ui.Element {
	config := chartCurrentConfig(c)
	if len(p.Payload) == 0 {
		return chartCompositionEmpty(c)
	}
	t := theme.TokensOf(c)
	root := ui.Row(c).FillWidth().Wrap().Gap(16).AlignItems(ui.Center).Justify(ui.Center).FontSize(12).LineHeight(4.0 / 3).Role(ui.RoleGroup).Label("Chart legend")
	if p.VerticalAlign == ChartLegendTop {
		root.Padding(0, 0, 12, 0)
	} else {
		root.Padding(12, 0, 0, 0)
	}
	return root.Children(func() {
		index := 0
		for _, item := range p.Payload {
			if item.Type == "none" {
				continue
			}
			entry, _ := chartResolvePayload(config, item, chartPayloadKey(p.NameKey, p.HasNameKey, item.DataKey))
			rowIndex := index
			index++
			ui.Row(c).Key(chartCompositionRowKey{Index: rowIndex}).Gap(6).AlignItems(ui.Center).Children(func() {
				if entry.Icon != nil && !p.HideIcon {
					entry.Icon(c.Key(chartLegendIconPart)).Size(12, 12).Shrink(0).TextColor(t.MutedForeground)
				} else {
					color := ui.Color{}
					if item.Color != nil {
						color = *item.Color
					}
					ui.Box(c).Size(8, 8).Shrink(0).Radius(2).Background(color).AccessHidden(true)
				}
				ui.Row(c).Key(chartLegendLabelPart).MinWidth(0).Children(func() { chartEntryLabel(entry).Render(c) })
			})
		}
	})
}
