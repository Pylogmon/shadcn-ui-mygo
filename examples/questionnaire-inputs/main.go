package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type appState struct {
	appearance theme.Controller
	form       conversation.QuestionnaireState
	saved      []string
	dark       bool
}

func (a *appState) view(c *ui.Context) {
	a.appearance.Mode = theme.LightMode
	if a.dark {
		a.appearance.Mode = theme.DarkMode
	}
	a.appearance.Apply(c)
	t := theme.TokensOf(c)
	questions := []conversation.Question{{
		ID: "destinations", Title: "Where should the report go?", Required: true, Multiple: true, Freeform: true,
		Choices: []control.Option{{Value: "team", Label: "Team channel"}},
		DefaultAnswer: conversation.Answer{
			Values: []string{"team"}, Mode: conversation.QuestionnaireCombinedAnswer,
			Inputs: []conversation.QuestionnaireInputValue{{Key: "primary", Text: "team", Selected: true}, {Key: "secondary", Text: "review", Selected: true}},
		},
	}}
	ui.Scroll(c).Fill().Background(t.Background).TextColor(t.Foreground).Children(func() {
		ui.Column(c).FillWidth().Padding(28).Gap(18).Children(func() {
			display.Typography(c, "Delivery destinations", display.H2)
			ui.Text(c, "Choose a channel and edit two independent destination labels.").TextColor(t.MutedForeground)
			control.Switch(c, &a.dark, control.SwitchProps{Label: "Dark appearance"})
			display.Card(c, func() {
				conversation.QuestionnaireWith(c, &a.form, conversation.QuestionnaireProps{
					Questions:      questions,
					OnAnswerChange: func(string, conversation.Answer) { a.saved = nil },
					OnReset:        func(*conversation.QuestionnaireResetEvent) { a.saved = nil },
					OnSubmit: func(map[string]conversation.Answer) {
						a.saved = slices.Clone(a.form.FormValues(questions)["destinations"])
					},
				}, func(parts *conversation.QuestionnaireParts) {
					parts.Item("destinations", func(item *conversation.QuestionnaireItemParts) {
						item.Header(func(*ui.Context) { item.Title(questions[0].Title) })
						item.Choices(func(*ui.Context) {
							primary := control.Label(c, "Primary destination")
							item.Input(conversation.QuestionnaireInputProps{Key: "primary", Label: "Primary destination"}).LabelledBy(primary)
							item.Choice("team")
							secondary := control.Label(c, "Secondary destination")
							item.Input(conversation.QuestionnaireInputProps{Key: "secondary", Label: "Secondary destination", Validate: func(value string) string {
								if strings.ContainsRune(value, '!') {
									return "Remove ! from the secondary destination."
								}
								return ""
							}}).LabelledBy(secondary)
						})
						item.Error()
					})
					parts.Actions(func(*ui.Context) {
						parts.Reset()
						parts.Submit(conversation.QuestionnaireActionProps{Label: "Save destinations"})
					})
				})
			})
			ui.Text(c, fmt.Sprintf("Current values: %q", a.form.FormValues(questions)["destinations"])).FontSize(13)
			if a.saved != nil {
				ui.Text(c, fmt.Sprintf("Saved destinations: %q", a.saved)).Role(ui.RoleStatus)
			}
		})
	})
}

func main() {
	snapshot := flag.String("snapshot", "", "write a native PNG instead of opening a window")
	dark := flag.Bool("dark", false, "use dark appearance")
	flag.Parse()
	a := &appState{appearance: theme.Controller{Preset: "blue"}, dark: *dark}
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		err = png.Encode(f, ui.Render(func(c *ui.Context) { c.DisableAnimations(); a.view(c) }, 600, 680, 1))
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			log.Fatal(err, closeErr)
		}
		return
	}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{Title: "Questionnaire inputs", Width: 600, Height: 680, Content: ui.View(a.view)}).Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
