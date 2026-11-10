package main

import (
	"flag"
	"image/png"
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type appState struct {
	code, result          string
	checked, custom, dark bool
}

func projectIssues(code string) []string {
	var issues []string
	if utf8.RuneCountInString(code) < 4 {
		issues = append(issues, "Use at least four characters.")
	}
	if strings.IndexFunc(code, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) >= 0 {
		issues = append(issues, "Use only letters and digits.")
	}
	return issues
}

func (a *appState) view(c *ui.Context) {
	mode := theme.LightMode
	if a.dark {
		mode = theme.DarkMode
	}
	appearance := theme.Controller{Mode: mode, Preset: "blue"}
	appearance.Apply(c)
	t := theme.TokensOf(c)
	ui.Column(c).Fill().Padding(28).Gap(18).Background(t.Background).TextColor(t.Foreground).Children(func() {
		display.Typography(c, "Name your project", display.H2)
		ui.Text(c, "Choose a code for this local project.").TextColor(t.MutedForeground)
		control.Switch(c, &a.dark, control.SwitchProps{Label: "Dark appearance"})
		control.Switch(c, &a.custom, control.SwitchProps{Label: "Custom error content"})
		if control.Button(c, "Check project", control.ButtonProps{}).Clicked() {
			a.checked = true
			a.result = ""
			if len(projectIssues(a.code)) == 0 {
				a.result = "Project code ready: " + a.code
			}
		}
		var issues []string
		if a.checked {
			issues = projectIssues(a.code)
		}
		props := control.FieldProps{Label: "Project code", Description: "A short identifier for your project.", Errors: issues}
		if a.custom && len(issues) != 0 {
			props.ErrorLabel = strings.Join(issues, "\n")
			props.ErrorContent = func(child *ui.Context) {
				ui.Text(child, "Project code requirements").FontWeight(500)
				for _, issue := range issues {
					ui.Text(child, issue)
				}
			}
		}
		control.Field(c, props, func() ui.Element {
			return control.InputGroup(c, control.InputGroupProps{Prefix: "project/"}, func() ui.Element {
				return control.Input(c, &a.code, control.InputProps{OnChange: func(string) { a.result = "" }})
			})
		})
		if a.result != "" {
			ui.Text(c, a.result).Role(ui.RoleStatus)
		}
	})
}

func main() {
	snapshot := flag.String("snapshot", "", "write a native PNG instead of opening a window")
	dark := flag.Bool("dark", false, "use dark appearance")
	custom := flag.Bool("custom", false, "use custom error content")
	flag.Parse()
	a := &appState{code: "ab!", checked: true, dark: *dark, custom: *custom}
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		err = png.Encode(f, ui.Render(func(c *ui.Context) { c.DisableAnimations(); a.view(c) }, 600, 520, 1))
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			log.Fatal(err, closeErr)
		}
		return
	}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{Title: "Field errors", Width: 600, Height: 520, Content: ui.View(a.view)}).Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
