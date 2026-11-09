package chartcatalog

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

const SourceCommit = "8ab27715521cbd4f87ae6886cd989cbb678f6de7"

type Entry struct{ ID, Title, Family, Section, Source string }

// Entries returns 70 catalogue entries, 8 documentation examples and 6 Create
// cards. It is an inventory, not a claim of 84 independent visual checks.
func Entries() []Entry {
	var out []Entry
	for _, r := range CartesianRecipes() {
		family := strings.Split(r.ID, "-")[1]
		out = append(out, Entry{r.ID, r.Title, family, "catalogue", "apps/v4/registry/new-york-v4/charts/" + r.ID + ".tsx"})
	}
	for _, r := range PolarRecipes() {
		family := strings.Split(r.ID, "-")[1]
		out = append(out, Entry{r.ID, r.Title, family, "catalogue", "apps/v4/registry/new-york-v4/charts/" + r.ID + ".tsx"})
	}
	for _, r := range AdditionalRecipes() {
		out = append(out, Entry{r.ID, r.Title, r.Family, r.Section, r.Source})
	}
	return out
}

type Model struct {
	Selected, SelectedSeries       int
	Range, Series, Month, Language string
}
type State struct {
	Collection, Family, Query, ID string
	Animate                       bool
	models                        map[string]*Model
}

func New() *State {
	return &State{Collection: "catalogue", Family: "all", ID: "chart-tooltip-advanced", Animate: true}
}
func (s *State) Initialize() {
	if s.Collection == "" {
		*s = *New()
	}
}
func (s *State) Model(id string) *Model {
	if s.models == nil {
		s.models = map[string]*Model{}
	}
	if m := s.models[id]; m != nil {
		return m
	}
	m := &Model{Selected: -1, Range: "90d", Series: "desktop", Month: "january", Language: "ar"}
	if strings.HasPrefix(id, "chart-tooltip-") {
		m.Selected = 1
	}
	s.models[id] = m
	return m
}
func (s *State) options(id string) Options {
	m := s.Model(id)
	return Options{Selected: &m.Selected, SelectedSeries: &m.SelectedSeries, Animate: s.Animate, Range: m.Range, Series: m.Series, Month: m.Month}
}
func recipeConfig(c *ui.Context, labels map[string]string, colors map[string]int) display.ChartConfig {
	config := display.ChartConfig{}
	palette := theme.TokensOf(c).Charts()
	for key, label := range labels {
		entry := display.ChartConfigItem{Label: label}
		if n, ok := colors[key]; ok {
			col := palette[n%len(palette)]
			entry.Color = &col
		}
		config[key] = entry
		config[title(key)] = entry
	}
	return config
}

// Props builds the same public API configurations used by the two standalone
// apps. The caller installs returned config with ChartContainer.Children.
func (s *State) Props(c *ui.Context, id string) (display.ChartProps, display.ChartConfig) {
	o := s.options(id)
	for _, r := range CartesianRecipes() {
		if r.ID == id {
			return CartesianProps(c, r, o), recipeConfig(c, r.Labels, r.Colors)
		}
	}
	for _, r := range PolarRecipes() {
		if r.ID == id {
			return PolarProps(c, r, o), recipeConfig(c, r.Labels, r.Colors)
		}
	}
	for _, r := range AdditionalRecipes() {
		if r.ID == id {
			switch r.Section {
			case "catalogue":
				return TooltipProps(c, r, o), tooltipConfig(c, r)
			case "docs":
				return DocsProps(c, r, o, s.Model(id).Language)
			case "create":
				return CreateProps(c, r, o)
			}
		}
	}
	panic("chartcatalog: unknown source ID " + id)
}
func (s *State) matches() []Entry {
	var out []Entry
	query := strings.ToLower(strings.TrimSpace(s.Query))
	for _, e := range Entries() {
		if e.Section != s.Collection {
			continue
		}
		if s.Family != "all" && s.Collection == "catalogue" && e.Family != s.Family {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(e.ID+" "+e.Title), query) {
			continue
		}
		out = append(out, e)
	}
	return out
}
func (s *State) View(c *ui.Context) {
	s.Initialize()
	ui.Column(c).FillWidth().MinWidth(0).Gap(16).Children(func() {
		ui.Text(c, "Chart source catalogue").FontSize(20).FontWeight(600)
		ui.Text(c, "70 chart examples · 8 documentation examples · 6 Create cards").FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
		control.Select(c, &s.Collection, control.SelectProps{Label: "Chart collection", Options: []control.Option{{Value: "catalogue", Label: "Chart examples (70)"}, {Value: "docs", Label: "Documentation (8)"}, {Value: "create", Label: "Create chart preview (6)"}}})
		if s.Collection == "catalogue" {
			opts := []control.Option{{Value: "all", Label: "All chart families"}}
			for _, f := range []string{"area", "bar", "line", "pie", "radar", "radial", "tooltip"} {
				opts = append(opts, control.Option{Value: f, Label: title(f)})
			}
			control.Select(c, &s.Family, control.SelectProps{Label: "Chart family", Options: opts})
		}
		control.Input(c, &s.Query, control.InputProps{Label: "Find chart example", Placeholder: "Search title or source ID", Search: true})
		control.Switch(c, &s.Animate, control.SwitchProps{Label: "Animate source charts", Size: control.Small})
		matches := s.matches()
		if len(matches) == 0 {
			ui.Text(c, "No chart examples match this search.")
			return
		}
		found := false
		opts := make([]control.Option, len(matches))
		for i, e := range matches {
			opts[i] = control.Option{Value: e.ID, Label: e.Title}
			found = found || e.ID == s.ID
		}
		if !found {
			s.ID = matches[0].ID
		}
		if s.Collection != "create" {
			control.Select(c, &s.ID, control.SelectProps{Label: "Source chart example", Options: opts})
			for _, e := range matches {
				if e.ID == s.ID {
					s.Render(c, e)
				}
			}
		} else {
			for _, e := range matches {
				s.Render(c, e)
			}
		}
	})
}
func (s *State) controls(c *ui.Context, e Entry) {
	m := s.Model(e.ID)
	if e.ID == "chart-area-interactive" {
		control.Select(c, &m.Range, control.SelectProps{Label: "Date range", Options: []control.Option{{Value: "90d", Label: "Last 3 months"}, {Value: "30d", Label: "Last 30 days"}, {Value: "7d", Label: "Last 7 days"}}})
	}
	if e.ID == "chart-bar-interactive" || e.ID == "chart-line-interactive" || e.ID == "docs-chart-demo" {
		var rows []map[string]any
		for _, r := range CartesianRecipes() {
			if r.ID == e.ID {
				rows = r.Data
			}
		}
		for _, r := range AdditionalRecipes() {
			if r.ID == e.ID {
				rows = r.Data
			}
		}
		ui.Row(c).Wrap().Gap(8).Children(func() {
			for _, key := range []string{"desktop", "mobile"} {
				total := 0.0
				for _, row := range rows {
					v, _ := row[key].(float64)
					total += v
				}
				if control.Button(c, fmt.Sprintf("%s · %.0f", title(key), total), control.ButtonProps{Variant: control.Outline}).Clicked() {
					m.Series = key
					c.Invalidate()
				}
			}
		})
	}
	if e.ID == "chart-pie-interactive" {
		for _, r := range PolarRecipes() {
			if r.ID == e.ID {
				opts := []control.Option{}
				for _, row := range r.rows() {
					key := row[r.category()].(string)
					opts = append(opts, control.Option{Value: key, Label: r.Labels[key]})
				}
				control.Select(c, &m.Month, control.SelectProps{Label: "Active month", Options: opts})
			}
		}
	}
	if e.ID == "docs-chart-rtl" {
		control.Select(c, &m.Language, control.SelectProps{Label: "Chart language", Options: []control.Option{{Value: "en", Label: "English"}, {Value: "ar", Label: "العربية"}, {Value: "he", Label: "עברית"}}})
	}
}
func (s *State) Render(c *ui.Context, e Entry) {
	// Key the entire recipe, including its plot and controls, so changing an ID
	// mounts new animation geometry without resetting caller-owned inspection.
	display.Card(c, func() {
		ui.Text(c, e.Title).FontSize(18).FontWeight(600)
		for _, r := range AdditionalRecipes() {
			if r.ID == e.ID && r.Description != "" {
				ui.Text(c, r.Description).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
			}
		}
		s.controls(c, e)
		if e.ID == "docs-chart-tooltip" {
			Anatomy(c)
			return
		}
		props, config := s.Props(c, e.ID)
		ui.Column(c).FillWidth().MinWidth(0).Direction(ui.LeftToRight).Children(func() {
			if e.ID == "docs-chart-rtl" && s.Model(e.ID).Language != "en" { // direction scopes text and native tooltip layout as source DirectionProvider does
				ui.Column(c).FillWidth().Direction(ui.RightToLeft).Children(func() { s.plot(c, e, props, config) })
			} else {
				s.plot(c, e, props, config)
			}
		})
		if e.Section == "create" {
			ui.Text(c, "Trending up by 5.2% this month ↗").FontSize(14).FontWeight(500)
			caption := "Showing total visitors for the last 6 months"
			if e.Family == "area" || e.Family == "bar" || e.Family == "line" {
				caption = "January - June 2024"
			}
			ui.Text(c, caption).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
		}
		ui.Text(c, "Inspect with the pointer or arrow keys; Escape clears inspection.").FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
	}).Key("source-card-" + e.ID)
}
func (s *State) plot(c *ui.Context, e Entry, p display.ChartProps, config display.ChartConfig) {
	if e.ID == "chart-area-icons" || e.ID == "chart-radar-icons" {
		for _, key := range []string{"desktop", "mobile"} {
			item := config[key]
			arrow := "↗"
			if key == "desktop" {
				arrow = "↘"
			}
			item.Icon = func(c *ui.Context) ui.Element { return ui.Text(c, arrow).AccessHidden(true) }
			config[key] = item
		}
		p.Legend = false
	}
	if e.ID == "docs-chart-example-legend" {
		p.Legend = false
	}
	display.ChartContainer(c, display.ChartContainerProps{Key: e.ID, Config: config}, func(c *ui.Context) {
		ui.Column(c).FillWidth().MinWidth(0).Children(func() {
			display.Chart(c, p)
			switch e.ID {
			case "chart-area-icons", "chart-radar-icons", "docs-chart-example-legend", "docs-chart-rtl":
				payload := []display.ChartPayloadItem{}
				keys := []string{"desktop", "mobile"}
				if e.ID == "chart-area-icons" {
					keys = []string{"mobile", "desktop"}
				}
				for _, key := range keys {
					item := config[key]
					payload = append(payload, display.ChartPayloadItem{DataKey: key, Name: key, Color: item.Color})
				}
				display.ChartLegendContent(c, display.ChartLegendContentProps{Payload: payload})
			case "chart-pie-legend":
				payload := []display.ChartPayloadItem{}
				for i, label := range p.Labels {
					color := p.CategoryColors[i]
					payload = append(payload, display.ChartPayloadItem{Name: label, DataKey: "visitors", Color: &color, Payload: p.Data[i]})
				}
				display.ChartLegendContent(c, display.ChartLegendContentProps{Payload: payload, NameKey: "browser"})
			}
		})
	}).AspectRatio(0)
}
