package main

import (
	"flag"
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type application struct {
	host                       *desktop.Host
	dark                       bool
	fontSize                   float32
	plan, status               string
	quantity, minimum, maximum float64
	empty, readOnly            bool
	menuOpen                   bool
	menu                       desktop.MenuState
}

func newApplication() *application {
	return &application{fontSize: 20, plan: "basic", quantity: 2, minimum: 0, maximum: 10, status: "Choose a plan or change the quantity."}
}

func (a *application) view(c *ui.Context) {
	tokens := theme.Preset("neutral", a.dark)
	tokens.FontSize = a.fontSize
	c.SetTheme(tokens.Theme())
	desktop.BindHost(c, a.host)
	a.host.BeginFrame()
	defer a.host.EndFrame()
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).FillWidth().Padding(24).Gap(20).FontSize(theme.ControlFontSize(c, 14, 0)).Children(func() {
			ui.Text(c, "Shared control capabilities").FontWeight(600)
			ui.Row(c).Wrap().Gap(8).Children(func() {
				if control.Button(c, "Change theme size", control.ButtonProps{Variant: control.Outline}).Clicked() {
					a.fontSize += 6
					if a.fontSize > 32 {
						a.fontSize = 14
					}
				}
				if control.Button(c, "Toggle appearance", control.ButtonProps{Variant: control.Outline}).Clicked() {
					a.dark = !a.dark
				}
				control.Button(c, "Local 18 DIP font", control.ButtonProps{Variant: control.Secondary, FontSize: 18})
			})
			ui.Text(c, "Plan · independent native list")
			_, err := desktop.Select(c, nil, "plan", &a.plan, control.SelectProps{
				Label: "Plan", Options: []control.Option{
					{Value: "basic", Label: "Basic"}, {Value: "team", Label: "Team"},
					{Value: "unavailable", Label: "Unavailable", Disabled: true},
				},
				OnChange: func(value string) {
					// Business identifiers and display conversion belong to the app.
					names := map[string]string{"basic": "Basic subscription", "team": "Team subscription"}
					a.status = "Selected: " + names[value]
				},
			})
			if err != nil {
				log.Print(err)
			}
			control.Field(c, control.FieldProps{Label: "Quantity", Description: "Optional, 0–10; buttons and Up/Down use a step of 0.5."}, func() ui.Element {
				return control.NumberInput(c, &a.quantity, control.NumberInputProps{
					Empty: &a.empty, Min: &a.minimum, Max: &a.maximum, Step: .5,
					ReadOnly: a.readOnly, Placeholder: "No quantity",
					Stepper: &control.NumberInputStepperProps{
						Decrement:        control.ButtonProps{Variant: control.Outline, Icon: outlined.Remove(), Label: "Decrease quantity"},
						Increment:        control.ButtonProps{Variant: control.Outline, Icon: outlined.Add(), Label: "Increase quantity"},
						DecrementTooltip: "Subtract 0.5", IncrementTooltip: "Add 0.5",
					},
				})
			})
			control.Checkbox(c, &a.readOnly, "Read only", control.CheckboxProps{})
			ui.Row(c).Gap(8).Children(func() {
				anchor := control.Button(c, "Actions", control.ButtonProps{Variant: control.Outline})
				if _, err := desktop.Menu(c, nil, "actions", anchor, &a.menuOpen, &a.menu, desktop.MenuProps{
					Items:    []overlay.MenuItem{{Key: "clear", Label: "Clear quantity"}},
					OnChoose: func(string) { a.empty = true },
				}); err != nil {
					log.Print(err)
				}
				anchor = control.Button(c, "Help", control.ButtonProps{Variant: control.Ghost})
				if err := desktop.Tooltip(c, nil, "help", anchor, "Select, Menu and Tooltip share this window's bound Host.", desktop.TooltipProps{}); err != nil {
					log.Print(err)
				}
			})
			ui.Text(c, a.status)
		})
	})
}

func main() {
	snapshot := flag.String("snapshot", "", "render the example to PNG without opening windows")
	dark := flag.Bool("dark", false, "use the dark theme")
	flag.Parse()
	a := newApplication()
	a.dark = *dark
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r := ui.NewTester(a.view, 820, 760)
		if err := png.Encode(f, r.Image()); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.OnReady(func() {
		window := mygo.NewWindow(mygo.WindowOptions{Title: "Control capabilities", Width: 820, Height: 760, Hidden: true, Content: ui.View(a.view)})
		a.host = desktop.NewHost(window)
		window.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
