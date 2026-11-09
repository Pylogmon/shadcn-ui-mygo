package control

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func tester(view func(*ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Gap(12).Children(func() { view(c) })
	}, 400, 360)
}

func click(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	tt.Frame()
}

func TestCheckboxSwitchAndDisabled(t *testing.T) {
	var checked, disabled, on bool
	tt := tester(func(c *ui.Context) {
		Checkbox(c, &checked, "Accept terms", CheckboxProps{})
		Checkbox(c, &disabled, "Unavailable", CheckboxProps{Disabled: true})
		Switch(c, &on, SwitchProps{Label: "Airplane mode"})
	})
	click(t, tt, "Accept terms")
	if !checked {
		t.Fatal("checkbox label did not change controlled state")
	}
	box, ok := tt.AccessByRole(ui.RoleCheckBox, "Accept terms")
	if !ok || !box.Checked {
		t.Fatal("checkbox did not report checked accessibility state")
	}
	click(t, tt, "Unavailable")
	if disabled {
		t.Fatal("disabled checkbox changed")
	}
	if !tt.HasText("Airplane mode") {
		t.Fatal("switch label is not visible")
	}
	sw, ok := tt.AccessByRole(ui.RoleSwitch, "Airplane mode")
	if !ok {
		t.Fatal("switch missing accessible label")
	}
	tt.ClickAt(sw.Bounds.X+sw.Bounds.W/2, sw.Bounds.Y+sw.Bounds.H/2)
	tt.Frame()
	if !on {
		t.Fatal("switch did not toggle")
	}
}

func TestTextInputTypingAndReadOnly(t *testing.T) {
	value := ""
	readOnly := "Locked"
	tt := tester(func(c *ui.Context) {
		Input(c, &value, InputProps{Label: "Name", Placeholder: "Your name"}).Width(240)
		Input(c, &readOnly, InputProps{Label: "Read only", ReadOnly: true}).Width(240)
	})
	field, ok := tt.AccessByRole(ui.RoleTextField, "Name")
	if !ok {
		t.Fatal("missing text field")
	}
	tt.ClickAt(field.Bounds.X+8, field.Bounds.Y+field.Bounds.H/2)
	tt.Type("Ada")
	if value != "Ada" {
		t.Fatalf("typed value = %q", value)
	}
	fixed, _ := tt.AccessByRole(ui.RoleTextField, "Read only")
	tt.ClickAt(fixed.Bounds.X+8, fixed.Bounds.Y+fixed.Bounds.H/2)
	tt.Type(" changed")
	if readOnly != "Locked" {
		t.Fatalf("read-only field changed: %q", readOnly)
	}
}

func TestSelectKeyboardSkipsDisabled(t *testing.T) {
	value := ""
	tt := tester(func(c *ui.Context) {
		Select(c, &value, SelectProps{Placeholder: "Pick framework", Options: []Option{
			{Value: "go", Label: "Go"}, {Value: "off", Label: "Unavailable", Disabled: true}, {Value: "rust", Label: "Rust"},
		}}).Width(240)
	})
	click(t, tt, "Pick framework")
	if !tt.HasText("Rust") {
		t.Fatal("select popup missing options")
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(time.Second)
	tt.Frame()
	if value != "rust" {
		t.Fatalf("keyboard selected %q, want rust (skip disabled)", value)
	}
	if tt.HasText("Go") {
		t.Fatal("select popup did not close")
	}
}

func TestRadioAndSliderKeyboard(t *testing.T) {
	selected := "go"
	value := 20.0
	tt := tester(func(c *ui.Context) {
		RadioGroup(c, &selected, RadioGroupProps{Label: "Language", Options: []Option{
			{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"},
		}})
		Slider(c, &value, SliderProps{Min: 0, Max: 100, Step: 10, Label: "Volume"}).Width(240)
	})
	click(t, tt, "Go")
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if selected != "rust" {
		t.Fatalf("radio arrows selected %q", selected)
	}
	slider, ok := tt.AccessByRole(ui.RoleSlider, "Volume")
	if !ok {
		t.Fatal("missing slider")
	}
	tt.ClickAt(slider.Bounds.X+slider.Bounds.W/2, slider.Bounds.Y+slider.Bounds.H/2)
	tt.Key(0, ui.KeyEnd)
	if value != 100 {
		t.Fatalf("slider End value = %v", value)
	}
	tt.Key(0, ui.KeyLeft)
	if value != 90 {
		t.Fatalf("slider Left step value = %v", value)
	}
}

func TestToggleGroupControlledSelection(t *testing.T) {
	selected := []string{}
	tt := tester(func(c *ui.Context) {
		ToggleGroup(c, &selected, ToggleGroupProps{Items: []Option{
			{Value: "bold", Label: "Bold"}, {Value: "italic", Label: "Italic"},
		}})
	})
	click(t, tt, "Bold")
	if !slices.Equal(selected, []string{"bold"}) {
		t.Fatalf("first selection %v", selected)
	}
	click(t, tt, "Italic")
	if !slices.Equal(selected, []string{"italic"}) {
		t.Fatalf("single selection %v", selected)
	}
	click(t, tt, "Italic")
	if len(selected) != 0 {
		t.Fatalf("toggle off selection %v", selected)
	}
}

func TestOTPFiltersAndCompletesOnce(t *testing.T) {
	value, completed := "", ""
	calls := 0
	tt := tester(func(c *ui.Context) {
		InputOTP(c, &value, InputOTPProps{Length: 4, OnComplete: func(s string) { completed = s; calls++ }})
	})
	field, ok := tt.AccessByRole(ui.RoleTextField, "One-time code")
	if !ok {
		t.Fatal("OTP missing native text editor")
	}
	tt.ClickAt(field.Bounds.X+8, field.Bounds.Y+field.Bounds.H/2)
	tt.Type("1a2")
	tt.Frame()
	if value != "12" {
		t.Fatalf("OTP filtered to %q", value)
	}
	tt.Type("34")
	tt.Frame()
	tt.Frame()
	if completed != "1234" || calls != 1 {
		t.Fatalf("OTP complete = %q (%d calls)", completed, calls)
	}
	// A pasted code may include separators; filter before applying the length limit.
	value = ""
	tt.Frame()
	tt.Type("12-34 56")
	tt.Frame()
	if value != "1234" {
		t.Fatalf("OTP with separators filtered to %q, want 1234", value)
	}
}
