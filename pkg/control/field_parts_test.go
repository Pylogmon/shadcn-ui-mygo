package control_test

import (
	"math"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// The same live native workflow changes composition rather than mounting a
// matrix of synthetic fields. Every target remains a genuine editor/choice.
func TestFieldPartsComposeOrderedDescriptionsAndNativeChoiceCards(t *testing.T) {
	phase, width := "profile", float32(600)
	name, password, environment, cardNumber := "", "", "kubernetes", ""
	errors := []string{"Use eight characters", "Use a number", "Use eight characters"}
	disabled, switched, legendOnly, invalidOnly := false, false, false, false
	sliderValue := float64(30)
	var inputHandle ui.Handle
	var expired *control.FieldScope
	var labelBounds, descriptionBounds, inputBounds, cardBounds, legendBounds ui.Rect
	var anchor, caret int
	r := ui.NewTester(func(c *ui.Context) {
		c.DisableAnimations()
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(20).Children(func() {
			switch phase {
			case "profile":
				control.FieldGroup(c, func() {
					control.FieldWith(c, control.FieldProps{Orientation: control.FieldResponsive, Required: true}, func(f *control.FieldScope) {
						expired = f
						f.Content(c, func() {
							labelBounds = f.Label(c, control.FieldLabelProps{Text: "Name"}, nil).LayoutBounds()
							descriptionBounds = f.Description(c, "Provide your full name for identification").LayoutBounds()
						})
						inputBounds = f.Control(c, func() ui.Element {
							return control.Input(c, &name, control.InputProps{Handle: &inputHandle}).Grow(1)
						}).LayoutBounds()
					})
				}).Width(width)
			case "password":
				control.FieldWith(c, control.FieldProps{Errors: errors, Disabled: disabled, Invalid: invalidOnly}, func(f *control.FieldScope) {
					f.Label(c, control.FieldLabelProps{Label: "Password"}, func() {
						control.FieldTitle(c, "Password")
						display.Badge(c, "Recommended", display.BadgeSecondary)
					})
					descriptionBounds = f.Description(c, "Must be at least 8 characters long.").LayoutBounds()
					inputBounds = f.Control(c, func() ui.Element {
						return control.InputGroup(c, control.InputGroupProps{Prefix: "#"}, func() ui.Element {
							return control.Input(c, &password, control.InputProps{Password: true, Handle: &inputHandle})
						})
					}).LayoutBounds()
					f.Error(c)
					f.DescriptionWith(c, "Review account security", func() {
						ui.Link(c, "Review account security", "https://example.com/security")
					})
				})
			case "choice":
				ui.RadioGroup(c, func() {
					for _, option := range []struct{ value, title, description string }{
						{"kubernetes", "Kubernetes", "Run GPU workloads on a K8s cluster."},
						{"vm", "Virtual Machine", "Access a cluster to run GPU workloads."},
					} {
						control.FieldWith(c.Key(option.value), control.FieldProps{Disabled: disabled}, func(f *control.FieldScope) {
							card := f.Label(c, control.FieldLabelProps{Label: option.title, Choice: true, Selected: environment == option.value}, func() {
								ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Start).Children(func() {
									f.Content(c, func() {
										control.FieldTitle(c, option.title)
										f.Description(c, option.description)
									})
									f.Control(c, func() ui.Element { return control.Radio(c, &environment, option.value, "", false) })
								})
							})
							if option.value == "vm" {
								cardBounds = card.LayoutBounds()
							}
						})
					}
				}).Label("Compute Environment").Gap(12).FocusGroup(ui.Vertical)
				control.FieldWith(c, control.FieldProps{Orientation: control.FieldHorizontal}, func(f *control.FieldScope) {
					f.Content(c, func() {
						f.Label(c, control.FieldLabelProps{Text: "Multi-factor authentication"}, nil)
						f.Description(c, "Verify your identity with another method")
					})
					f.Control(c, func() ui.Element { return control.Switch(c, &switched, control.SwitchProps{}) })
				})
			case "slider":
				control.FieldWith(c, control.FieldProps{Invalid: invalidOnly}, func(f *control.FieldScope) {
					f.Label(c, control.FieldLabelProps{Text: "Volume"}, nil)
					f.Control(c, func() ui.Element {
						return control.Slider(c, &sliderValue, control.SliderProps{Label: "Volume", Min: 0, Max: 100, Step: 1})
					})
				})
			case "payment":
				control.FieldSetWith(c, control.FieldSetProps{}, control.FieldSetSlots{
					Legend: func(child *ui.Context) ui.Element {
						e := control.FieldLegend(child, "Payment Method", control.FieldLegendLabel)
						legendBounds = e.LayoutBounds()
						return e
					},
					Description: func(child *ui.Context) ui.Element {
						if legendOnly {
							return ui.Element{}
						}
						return control.FieldDescription(child, "All transactions are secure and encrypted")
					},
					Content: func(child *ui.Context) {
						control.FieldWith(child, control.FieldProps{}, func(f *control.FieldScope) {
							f.Label(child, control.FieldLabelProps{Text: "Card number"}, nil)
							inputBounds = f.Control(child, func() ui.Element {
								return control.Input(child, &cardNumber, control.InputProps{})
							}).LayoutBounds()
						})
						control.FieldSeparator(child, "Or continue with")
						control.Button(child, "Pay locally", control.ButtonProps{})
					},
				})
			}
			anchor, caret = c.Resolve(inputHandle).TextSelection()
		})
	}, 760, 520)
	r.Frame()
	if inputBounds.X <= labelBounds.X+labelBounds.W || descriptionBounds.Y <= labelBounds.Y || descriptionBounds.X != labelBounds.X {
		t.Fatalf("wide responsive content did not group its label/description beside the editor: label=%+v description=%+v editor=%+v", labelBounds, descriptionBounds, inputBounds)
	}
	if err := r.Click("Name *"); err != nil {
		t.Fatal(err)
	}
	r.Type("Evil Rabbit")
	r.Key(0, ui.KeyLeft)
	width = 320
	r.Frame()
	r.Frame()
	if inputBounds.Y <= descriptionBounds.Y || inputBounds.X != labelBounds.X || !r.Focused("Name *") || name != "Evil Rabbit" || anchor != 10 || caret != 10 {
		t.Fatalf("responsive stacking lost editor: label=%+v description=%+v input=%+v focused=%v name=%q selection=%d/%d", labelBounds, descriptionBounds, inputBounds, r.Focused("Name *"), name, anchor, caret)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expired FieldScope remained usable")
			}
		}()
		expired.Description(nil, "Invalid")
	}()
	phase = "password"
	r.Frame()
	r.Frame()
	a := errorAccess(t, r, ui.RoleTextField, "Password")
	if !a.Password || !a.Invalid || !slices.Contains(a.DescribedBy, "Must be at least 8 characters long.") || !slices.Contains(a.DescribedBy, "Review account security") || !slices.Equal(a.ErrorMessage, []string{"Use eight characters\nUse a number"}) || descriptionBounds.Y >= inputBounds.Y {
		t.Fatalf("rich label/pre-control description/errors did not reach the real password editor: %+v", a)
	}
	if err := r.Click("Password"); err != nil {
		t.Fatal(err)
	}
	r.Type("secret1")
	disabled = true
	r.Frame()
	if !errorAccess(t, r, ui.RoleTextField, "Password").Disabled {
		t.Fatal("field disabled state missed the editor inside InputGroup")
	}
	disabled, errors = false, nil
	r.Frame()
	if errorAccess(t, r, ui.RoleTextField, "Password").Invalid || password != "secret1" {
		t.Fatal("clearing errors retained stale invalid metadata or replaced the password")
	}
	invalidOnly = true
	r.Frame()
	if !errorAccess(t, r, ui.RoleTextField, "Password").Invalid {
		t.Fatal("validation-only Invalid failed to reach the editor")
	}
	if _, ok := r.AccessByRole(ui.RoleAlert, ""); ok {
		t.Fatal("validation-only Invalid manufactured error text")
	}
	invalidOnly = false
	phase = "choice"
	r.Frame()
	r.Frame()
	// Activate padding far below the small radio; the whole label is the hit
	// target, and native label forwarding must activate exactly one real radio.
	r.ClickAt(cardBounds.X+24, cardBounds.Y+cardBounds.H-6)
	if environment != "vm" || !errorAccess(t, r, ui.RoleRadio, "Virtual Machine").Checked {
		t.Fatal("choice-card padding failed to activate its actual native radio")
	}
	r.Key(0, ui.KeyUp)
	if environment != "kubernetes" {
		t.Fatal("choice labels intercepted native radio-group keyboard navigation")
	}
	if err := r.Click("Multi-factor authentication"); err != nil {
		t.Fatal(err)
	}
	if !switched || !errorAccess(t, r, ui.RoleSwitch, "Multi-factor authentication").Checked {
		t.Fatal("horizontal Content label did not activate the real switch")
	}
	for _, node := range r.AccessAll() {
		if (node.Name == "Virtual Machine" || node.Name == "Kubernetes") && node.Focusable && node.Role != ui.RoleRadio {
			t.Fatal("choice label introduced another semantic choice or tab stop")
		}
	}
	disabled = true
	r.Frame()
	r.ClickAt(cardBounds.X+24, cardBounds.Y+cardBounds.H-6)
	if environment != "kubernetes" {
		t.Fatal("disabled choice-card label activated its radio")
	}
	phase, disabled = "payment", false
	r.Frame()
	r.Frame()
	set := errorAccess(t, r, ui.RoleGroup, "Payment Method")
	if !slices.Contains(set.DescribedBy, "All transactions are secure and encrypted") || !r.HasText("Or continue with") {
		t.Fatal("composed fieldset lost its actual legend/description or separator text")
	}
	legendOnly = true
	r.Frame()
	r.Frame()
	label, ok := r.Find("Card number")
	if !ok || math.Abs(float64(label.Y-(legendBounds.Y+legendBounds.H)-12)) > .5 {
		t.Fatalf("lone legend accumulated a fieldset flex gap: legend=%+v next=%+v", legendBounds, label)
	}
	if err := r.Click("Card number"); err != nil {
		t.Fatal(err)
	}
	r.Type("4242")
	if cardNumber != "4242" {
		t.Fatal("composed payment label did not focus its real editor")
	}
	phase, invalidOnly = "slider", true
	r.Frame()
	a = errorAccess(t, r, ui.RoleSlider, "Volume")
	if !a.Invalid || a.Disabled {
		t.Fatal("Field Invalid did not reach the real slider thumb independently of numeric constraints")
	}
	if err := r.Click("Volume"); err != nil {
		t.Fatal(err)
	}
	before := sliderValue
	r.Key(0, ui.KeyRight)
	if sliderValue <= before {
		t.Fatal("application validation incorrectly disabled a valid slider domain")
	}
	invalidOnly = false
	r.Frame()
	if errorAccess(t, r, ui.RoleSlider, "Volume").Invalid {
		t.Fatal("cleared Field Invalid retained stale thumb metadata")
	}
}
