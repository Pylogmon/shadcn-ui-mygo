package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type appState struct {
	appearance       theme.Controller
	colors           [3]ui.Color
	custom, disabled bool
}

func newState(dark bool) *appState {
	s := &appState{custom: true, appearance: theme.Controller{Mode: theme.LightMode}}
	if dark {
		s.appearance.Mode = theme.DarkMode
	}
	for i := range s.colors {
		s.colors[i] = control.DefaultColorOptions()[0].Color
	}
	return s
}

func (s *appState) view(c *ui.Context) {
	s.appearance.Apply(c)
	t := theme.TokensOf(c)
	ui.Column(c).Fill().Padding(28).Gap(20).Background(t.Background).TextColor(t.Foreground).Children(func() {
		ui.Text(c, "Color Picker").FontSize(28).FontWeight(600)
		ui.Text(c, "Eight presets, three shapes, and any color you can imagine.").TextColor(t.MutedForeground)
		ui.Row(c).Wrap().Gap(24).Children(func() {
			control.Switch(c, &s.custom, control.SwitchProps{Label: "Custom color well"})
			control.Switch(c, &s.disabled, control.SwitchProps{Label: "Disabled"})
			dark := s.appearance.Mode == theme.DarkMode
			if control.Switch(c, &dark, control.SwitchProps{Label: "Dark appearance"}).Changed() {
				s.appearance.Mode = theme.LightMode
				if dark {
					s.appearance.Mode = theme.DarkMode
				}
			}
		})
		ui.Scroll(c).Fill().Children(func() {
			ui.Column(c).FillWidth().Gap(28).Children(func() {
				for i, item := range []struct {
					name  string
					shape control.ColorSwatchShape
				}{{"Rounded square", control.ColorSwatchRounded}, {"Square", control.ColorSwatchSquare}, {"Circle", control.ColorSwatchCircle}} {
					ui.Column(c.Key(item.name)).FillWidth().Gap(12).Children(func() {
						ui.Row(c).FillWidth().Wrap().Gap(12).Children(func() {
							ui.Text(c, item.name).FontSize(14).FontWeight(500)
							color := s.colors[i]
							ui.Text(c, fmt.Sprintf("#%02x%02x%02x · alpha %d%%", color.R, color.G, color.B, int(color.A)*100/255)).FontSize(13).TextColor(t.MutedForeground)
						})
						ui.Column(c).FillWidth().Padding(12).Children(func() {
							control.ColorPicker(c, &s.colors[i], control.ColorPickerProps{Label: item.name + " colors", CustomLabel: item.name + " custom color", Custom: s.custom, Disabled: s.disabled, Shape: item.shape})
						})
						ui.Box(c).FillWidth().Height(1).Background(t.Border)
					})
				}
			})
		})
	})
}

func main() {
	snapshot := flag.String("snapshot", "", "write a PNG and exit")
	dark := flag.Bool("dark", false, "dark appearance")
	flag.Parse()
	s := newState(*dark)
	if *snapshot != "" {
		file, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(file, ui.Render(s.view, 760, 620, 1)); err != nil {
			log.Fatal(err)
		}
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		window := mygo.NewWindow(mygo.WindowOptions{Title: "Color Picker", Width: 760, Height: 620, Content: ui.View(s.view)})
		window.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
