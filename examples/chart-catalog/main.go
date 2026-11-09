// Fixed shadcn chart catalogue, documentation workflows and Create preview.
package main

import (
	"flag"
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/chartcatalog"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type app struct {
	appearance theme.Controller
	catalogue  *chartcatalog.State
	dark       bool
}

func (a *app) view(c *ui.Context) {
	a.appearance.Mode = theme.LightMode
	if a.dark {
		a.appearance.Mode = theme.DarkMode
	}
	a.appearance.Apply(c)
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).FillWidth().Padding(28).Gap(16).Children(func() { control.Switch(c, &a.dark, control.SwitchProps{Label: "Dark appearance"}); a.catalogue.View(c) })
	})
}
func main() {
	example := flag.String("example", "chart-tooltip-advanced", "fixed source ID")
	collection := flag.String("collection", "catalogue", "catalogue, docs or create")
	export := flag.String("export", "", "render a native PNG without opening a window")
	exportHeight := flag.Int("export-height", 1000, "native PNG height (use 4000 for all six Create cards)")
	flag.Parse()
	catalogue := chartcatalog.New()
	catalogue.ID = *example
	catalogue.Collection = *collection
	a := &app{appearance: theme.Controller{Preset: "blue"}, catalogue: catalogue}
	if *export != "" {
		catalogue.Animate = false
		r := ui.NewTesterWith(a.view, 1000, max(300, *exportHeight), ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		f, err := os.Create(*export)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err = png.Encode(f, r.Image()); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "Chart catalogue", Width: 1040, Height: 860, Content: ui.View(a.view)})
		w.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
