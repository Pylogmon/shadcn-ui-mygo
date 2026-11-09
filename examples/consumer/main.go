// A small application using only the kit's public packages. Copy this file to
// your own module, require the kit, and run go mod tidy followed by go run .
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
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type appState struct {
	appearance  theme.Controller
	host        *desktop.Host
	name        string
	framework   string
	nativeMenu  bool
	status      string
	preferences string
}

func (s *appState) view(c *ui.Context) {
	s.appearance.Apply(c)
	if s.host != nil {
		s.host.BeginFrame()
		defer s.host.EndFrame()
	}
	t := theme.TokensOf(c)
	ui.Column(c).Fill().Padding(24).Gap(16).Background(t.Background).Children(func() {
		display.Typography(c, "A native Go application", display.H2)
		control.Input(c, &s.name, control.InputProps{Label: "Project name", Placeholder: "My application"})
		control.Switch(c, &s.nativeMenu, control.SwitchProps{Label: "Use system menu"})
		props := control.SelectProps{Label: "Framework", Options: []control.Option{
			{Value: "mygo", Label: "MyGo"}, {Value: "go", Label: "Go"},
		}}
		if s.nativeMenu {
			desktop.NativeSelect(c, &s.framework, props)
		} else if _, err := desktop.Select(c, s.host, "framework", &s.framework, props); err != nil {
			s.status = err.Error()
		}
		ui.Row(c).Gap(8).Wrap().Children(func() {
			if control.Button(c, "Toggle appearance", control.ButtonProps{Variant: control.Outline}).Clicked() {
				s.appearance.Toggle(c.Theme().Dark)
			}
			if control.Button(c, "Save theme", control.ButtonProps{Disabled: s.preferences == ""}).Clicked() {
				if err := s.appearance.Save(s.preferences); err != nil {
					s.status = err.Error()
				} else {
					s.status = "Theme saved"
				}
			}
		})
		local := theme.Preset("rose", t.Dark)
		theme.Scope(c, local).FillWidth().Children(func() {
			display.Card(c, func() {
				display.Typography(c, "A scoped theme", display.H3)
				control.Button(c, "Rose action", control.ButtonProps{})
			})
		})
		if s.status != "" {
			ui.Text(c, s.status).Role(ui.RoleStatus)
		}
	})
}

func main() {
	snapshot := flag.String("snapshot", "", "render a PNG instead of opening a window")
	preferences := flag.String("preferences", "", "optional application-owned theme settings file")
	verify := flag.Bool("verify-theme", false, "verify persistence before rendering the snapshot")
	flag.Parse()
	s := &appState{appearance: theme.Controller{Mode: theme.System, Preset: "blue"}, name: "Native consumer", framework: "mygo", preferences: *preferences}
	if *preferences != "" {
		if err := s.appearance.Load(*preferences); err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}
	if *verify {
		if *preferences == "" || *snapshot == "" {
			log.Fatal("verify-theme requires snapshot and preferences paths")
		}
		s.appearance.Mode = theme.DarkMode
		s.appearance.Radius = 12
		if err := s.appearance.Save(*preferences); err != nil {
			log.Fatal(err)
		}
		var restored theme.Controller
		if err := restored.Load(*preferences); err != nil {
			log.Fatal(err)
		}
		if restored.Mode != theme.DarkMode || restored.Preset != "blue" || restored.Radius != 12 || !restored.Resolve(false).Dark {
			log.Fatal("theme persistence did not survive a fresh controller")
		}
		s.appearance = restored
		s.status = "Theme persistence verified"
	}
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(f, ui.Render(s.view, 560, 520, 1)); err != nil {
			f.Close()
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Consumer snapshot rendered")
		return
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "Native consumer", Width: 560, Height: 520, Hidden: true, Content: ui.View(s.view)})
		s.host = desktop.NewHost(w)
		w.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
