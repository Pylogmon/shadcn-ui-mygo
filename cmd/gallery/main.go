package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/gallery"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func main() {
	snapshot := flag.String("snapshot", "", "write a PNG and exit")
	dark := flag.Bool("dark", false, "start in dark mode")
	page := flag.String("page", "overview", "overview, controls, display, content, text-layout, overlays, conversation, navigation or themes")
	preset := flag.String("preset", "", "load a creator preset JSON file")
	flag.Parse()
	state := gallery.New()
	state.Page = *page
	if *preset != "" {
		data, err := os.ReadFile(*preset)
		if err != nil {
			log.Fatal(err)
		}
		if err := state.ImportPreset(string(data)); err != nil {
			log.Fatal(err)
		}
	}
	if *dark {
		state.Theme.Mode = theme.DarkMode
	}
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err = png.Encode(f, renderSnapshot(state)); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		window := mygo.NewWindow(mygo.WindowOptions{Title: "shadcn/ui — Native Go", Width: 1280, Height: 900, Hidden: true, Content: ui.View(state.View)})
		state.Window = window
		state.Popovers = desktop.NewHost(window)
		window.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func renderSnapshot(state *gallery.State) *image.RGBA {
	return ui.Render(func(c *ui.Context) {
		c.DisableAnimations()
		state.View(c)
	}, 1280, 900, 1)
}
