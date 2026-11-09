// A native Material Symbols showcase. This example deliberately imports the
// optional all package to demonstrate dynamic lookup as well as direct symbols.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/all"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/rounded"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/sharp"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type state struct {
	dark, filled bool
	style, query string
	status       string
}

func (s *state) view(c *ui.Context) {
	controller := theme.Controller{Preset: "blue", Mode: theme.LightMode}
	if s.dark {
		controller.Mode = theme.DarkMode
	}
	controller.Apply(c)
	t := theme.TokensOf(c)
	ui.Scroll(c).Fill().Background(t.Background).Children(func() {
		ui.Column(c).FillWidth().Padding(24).Gap(20).Children(func() {
			display.Typography(c, "Material Symbols", display.H1)
			ui.Text(c, "Native vectors that follow your theme.").TextColor(t.MutedForeground)
			display.Card(c, func() {
				display.Typography(c, "Individual symbols", display.H3)
				ui.Row(c).Gap(20).Wrap().AlignItems(ui.Center).Children(func() {
					material.Icon(c, outlined.Home(), material.Props{Size: 24, Label: "Outlined home"})
					material.Icon(c, rounded.Favorite(material.Filled), material.Props{Size: 32, Label: "Filled rounded favorite"}).TextColor(t.Destructive)
					material.Icon(c, sharp.Settings(), material.Props{Size: 24, Label: "Sharp settings"})
					material.Icon(c, outlined.ExpandMore(), material.Props{Size: 16, Label: "Expand"})
					if control.Button(c, "Save", control.ButtonProps{Icon: outlined.Save()}).Clicked() {
						s.status = "Saved"
					}
				})
			})
			display.Card(c, func() {
				display.Typography(c, "Explore the catalog", display.H3)
				ui.Row(c).Gap(16).Wrap().AlignItems(ui.Center).Children(func() {
					control.NativeSelect(c, &s.style, control.SelectProps{Label: "Symbol style", Options: []control.Option{
						{Value: "outlined", Label: "Outlined"}, {Value: "rounded", Label: "Rounded"}, {Value: "sharp", Label: "Sharp"},
					}}).Width(160)
					control.Switch(c, &s.filled, control.SwitchProps{Label: "Filled"})
					control.Switch(c, &s.dark, control.SwitchProps{Label: "Dark appearance"})
				})
				control.Input(c, &s.query, control.InputProps{Label: "Search symbols", Placeholder: "Search names, e.g. home, menu, arrow…"})
				variant := material.Variant{}
				if s.style == "rounded" {
					variant.Style = material.Rounded
				} else if s.style == "sharp" {
					variant.Style = material.Sharp
				}
				if s.filled {
					variant.Fill = material.Filled
				}
				matches := []string{}
				for _, name := range all.Names() {
					if strings.Contains(name, strings.ToLower(strings.TrimSpace(s.query))) {
						matches = append(matches, name)
					}
				}
				ui.Text(c, fmt.Sprintf("%d matches · %d symbols", len(matches), len(all.Names()))).FontSize(12).TextColor(t.MutedForeground)
				ui.Row(c).FillWidth().Wrap().Gap(8).Children(func() {
					for _, name := range matches[:min(48, len(matches))] {
						svg, err := all.Lookup(name, variant)
						if err != nil {
							panic(err) // names and variants above are from this pinned catalog
						}
						ui.Column(c.Key(name)).Width(104).MinHeight(76).Padding(8).Gap(8).AlignItems(ui.Center).Border(1, t.Border).Radius(t.RadiusMD()).Children(func() {
							material.Icon(c, svg, material.Props{Size: 24, Label: name})
							ui.Text(c, name).FontSize(10).MaxWidth(88).SingleLine()
						})
					}
				})
				if len(matches) == 0 {
					ui.Text(c, "No symbols found.").Role(ui.RoleStatus).TextColor(t.MutedForeground)
				}
			})
			if s.status != "" {
				ui.Text(c, s.status).Role(ui.RoleStatus)
			}
		})
	})
}

func main() {
	snapshot := flag.String("snapshot", "", "render a PNG instead of opening a window")
	dark := flag.Bool("dark", false, "use the dark theme")
	flag.Parse()
	s := &state{dark: *dark, style: "outlined", query: "home"}
	if *snapshot != "" {
		im := ui.Render(func(c *ui.Context) { c.DisableAnimations(); s.view(c) }, 900, 760, 1)
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(f, im); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "Material Symbols", Width: 900, Height: 760, Content: ui.View(s.view)})
		w.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
