package control

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestControlsUseThemeSizeRatiosAndLocalFontOverrides(t *testing.T) {
	for _, size := range []Size{ExtraSmall, Small, DefaultSize, Large} {
		for _, override := range []float32{0, 24} {
			t.Run(fmt.Sprintf("size=%d/override=%g", size, override), func(t *testing.T) {
				n, selected := 123., "one"
				font := float32(28)
				const scale = float32(1.5)
				buttonRef, numberRef := float32(14), float32(14)
				if size == ExtraSmall {
					buttonRef, numberRef = 12, 12
				} else if size == Large {
					numberRef = 16
				}
				buttonFont, numberFont := font*buttonRef/14*scale, font*numberRef/14*scale
				if override > 0 {
					buttonFont, numberFont = override*scale, override*scale
				}
				var triggerText, optionText, referenceText, buttonReference ui.Rect
				sample := "MMMM"
				r := ui.NewTesterWith(func(c *ui.Context) {
					tokens := theme.Light()
					tokens.FontSize = font
					c.SetTheme(tokens.Theme())
					ui.Column(c).Padding(12).Gap(12).Children(func() {
						Button(c, sample, ButtonProps{Label: "Action", Size: size, FontSize: override}).AlignSelf(ui.Start)
						Select(c, &selected, SelectProps{Label: "Choice", Size: size, FontSize: override, Content: SelectContentProps{Width: 600}, Options: []Option{{Value: "one", Label: sample}, {Value: "two", Label: "Two"}},
							TriggerContent: func(c *ui.Context, _ Option) {
								ui.Text(c, sample).DrawOver(func(_ *ui.Painter, r ui.Rect) { triggerText = r })
							},
							ItemContent: func(c *ui.Context, o Option) {
								if o.Value == "one" {
									ui.Text(c, sample).DrawOver(func(_ *ui.Painter, r ui.Rect) { optionText = r })
								} else {
									ui.Text(c, o.Label)
								}
							},
						})
						NumberInput(c, &n, NumberInputProps{Label: "Number", Size: size, FontSize: override, AutoFocus: true})
						ui.Text(c, sample).AlignSelf(ui.Start).FontSize(buttonFont).DrawOver(func(_ *ui.Painter, r ui.Rect) { referenceText = r })
						ui.Text(c, sample).AlignSelf(ui.Start).FontSize(buttonFont).FontWeight(500).DrawOver(func(_ *ui.Painter, r ui.Rect) { buttonReference = r })
					})
				}, 900, 650, ui.TesterOptions{Preferences: ui.Preferences{TextScale: scale, ReduceMotion: true}})
				if math.Abs(float64(triggerText.W-referenceText.W)) > .1 {
					t.Fatalf("trigger did not inherit actual theme text: trigger=%v reference=%v", triggerText, referenceText)
				}
				action, _ := r.AccessByRole(ui.RoleButton, "Action")
				_, padding := metrics(size)
				if math.Abs(float64(action.Bounds.W-buttonReference.W-2*padding-2)) > .1 {
					t.Fatalf("button text ignored font: button=%v reference=%v", action.Bounds, buttonReference)
				}
				if action.Bounds.H < buttonFont*(20.0/14)+4 {
					t.Fatalf("button clipped large text: %v font=%g", action.Bounds, buttonFont)
				}
				caret, active := r.TextCaret()
				editor, _ := r.AccessByRole(ui.RoleTextField, "Number")
				if !active || caret.H < numberFont*.9 || caret.Y < editor.Bounds.Y+1 || caret.Y+caret.H > editor.Bounds.Y+editor.Bounds.H-1 {
					t.Fatalf("real editor ignored font/height: font=%g editor=%v caret=%v", numberFont, editor.Bounds, caret)
				}
				if err := r.Click("Choice"); err != nil {
					t.Fatal(err)
				}
				r.Frame()
				if math.Abs(float64(optionText.W-referenceText.W)) > .1 {
					t.Fatalf("styled option text ignored trigger font: option=%v reference=%v", optionText, referenceText)
				}
			})
		}
	}
}

func TestNumberInputFontOverrideSurvivesFieldAndInputGroup(t *testing.T) {
	n := 3.
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Field(c, FieldProps{Label: "Amount", Description: "Units", Invalid: true}, func() ui.Element {
			return InputGroup(c, InputGroupProps{Prefix: "$"}, func() ui.Element {
				return NumberInput(c, &n, NumberInputProps{FontSize: 36, AutoFocus: true, Stepper: &NumberInputStepperProps{}})
			})
		})
	}, 600, 240)
	caret, _ := r.TextCaret()
	editor, ok := r.AccessByRole(ui.RoleTextField, "Amount")
	if !ok || caret.H < 32 || !editor.Invalid || len(editor.DescribedBy) == 0 || caret.Y+caret.H > editor.Bounds.Y+editor.Bounds.H-1 {
		t.Fatalf("compound input lost the actual editor or font: editor=%+v caret=%v", editor, caret)
	}
}
