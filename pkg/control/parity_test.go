package control

import (
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestButtonLoadingAndIconSizes(t *testing.T) {
	loading := true
	activated := 0
	tt := tester(func(c *ui.Context) {
		if Button(c, "Upload", ButtonProps{Loading: loading}).Clicked() {
			activated++
		}
		ui.Row(c).Gap(8).Children(func() {
			for _, s := range []struct {
				label string
				size  Size
			}{{"Extra small action", IconExtraSmall}, {"Small action", IconSmall}, {"Action", Icon}, {"Large action", IconLarge}} {
				Button(c, "+", ButtonProps{Label: s.label, Size: s.size})
			}
		})
	})
	button, ok := tt.AccessByRole(ui.RoleButton, "Upload")
	if !ok || !button.Busy || !button.Disabled {
		t.Fatalf("loading access = %+v", button)
	}
	click(t, tt, "Upload")
	if activated != 0 {
		t.Fatal("loading action activated")
	}
	for _, s := range []struct {
		label string
		size  float32
	}{{"Extra small action", 24}, {"Small action", 32}, {"Action", 36}, {"Large action", 40}} {
		b, ok := tt.AccessByRole(ui.RoleButton, s.label)
		if !ok || b.Bounds.W != s.size || b.Bounds.H != s.size {
			t.Fatalf("%s = %+v", s.label, b.Bounds)
		}
	}
	loading = false
	tt.Frame()
	click(t, tt, "Upload")
	if activated != 1 {
		t.Fatalf("ready button activated %d times", activated)
	}
}

func TestButtonGroupJoinedGeometryAndKeyboard(t *testing.T) {
	var pressed string
	tt := tester(func(c *ui.Context) {
		ButtonGroup(c, func() {
			for _, name := range []string{"First", "Disabled", "Last"} {
				if Button(c, name, ButtonProps{Variant: Outline, Disabled: name == "Disabled"}).Clicked() {
					pressed = name
				}
			}
		}, ButtonGroupProps{Label: "Actions"})
		ButtonGroup(c, func() {
			Button(c, "Up", ButtonProps{Variant: Outline})
			Button(c, "Down", ButtonProps{Variant: Outline})
		}, ButtonGroupProps{Label: "Vertical actions", Vertical: true})
		Button(c, "Separate", ButtonProps{Variant: Outline})
	})
	first, _ := tt.AccessByRole(ui.RoleButton, "First")
	disabled, _ := tt.AccessByRole(ui.RoleButton, "Disabled")
	last, _ := tt.AccessByRole(ui.RoleButton, "Last")
	if first.Bounds.X+first.Bounds.W-disabled.Bounds.X != 1 || disabled.Bounds.X+disabled.Bounds.W-last.Bounds.X != 1 {
		t.Fatalf("joined boundaries %+v %+v %+v", first.Bounds, disabled.Bounds, last.Bounds)
	}
	up, _ := tt.AccessByRole(ui.RoleButton, "Up")
	down, _ := tt.AccessByRole(ui.RoleButton, "Down")
	if up.Bounds.Y+up.Bounds.H-down.Bounds.Y != 1 {
		t.Fatalf("vertical boundaries %+v %+v", up.Bounds, down.Bounds)
	}
	g, _ := tt.AccessByRole(ui.RoleGroup, "Vertical actions")
	if !g.Vertical {
		t.Fatal("group orientation missing")
	}
	click(t, tt, "First")
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if pressed != "Last" {
		t.Fatalf("roving keyboard action=%s", pressed)
	}
	// Interior seams have square corners; a missing fill at the top seam would
	// expose the page background and create the old scalloped group edge.
	image := tt.Image()
	at := image.RGBAAt(int(first.Bounds.X+first.Bounds.W-3), int(first.Bounds.Y))
	if at.R > 245 && at.G > 245 && at.B > 245 {
		t.Fatalf("interior seam has rounded gap: %+v", at)
	}
}

func TestButtonGroupsScopeNestedLayoutsAndRTL(t *testing.T) {
	rtl := true
	tt := tester(func(c *ui.Context) {
		dir := ui.LeftToRight
		if rtl {
			dir = ui.RightToLeft
		}
		ui.Column(c).Direction(dir).Children(func() {
			ButtonGroup(c, func() {
				Button(c, "Start", ButtonProps{Variant: Outline})
				Button(c, "End", ButtonProps{Variant: Outline})
			}, ButtonGroupProps{Label: "Directional"})
			ButtonGroup(c, func() {
				ButtonGroup(c, func() {
					Button(c, "Nested A", ButtonProps{Variant: Outline})
					Button(c, "Nested B", ButtonProps{Variant: Outline})
				})
				Button(c, "Sibling", ButtonProps{Variant: Outline})
			})
		})
		Button(c, "Independent", ButtonProps{Variant: Outline})
	})
	a, _ := tt.AccessByRole(ui.RoleButton, "Start")
	b, _ := tt.AccessByRole(ui.RoleButton, "End")
	if a.Bounds.X <= b.Bounds.X {
		t.Fatalf("RTL order %+v %+v", a.Bounds, b.Bounds)
	}
	nestedA, _ := tt.AccessByRole(ui.RoleButton, "Nested A")
	nestedB, _ := tt.AccessByRole(ui.RoleButton, "Nested B")
	sibling, _ := tt.AccessByRole(ui.RoleButton, "Sibling")
	if math.Abs(float64(nestedA.Bounds.X-nestedB.Bounds.X-nestedB.Bounds.W+1)) > .01 {
		t.Fatalf("nested buttons were not joined: A=%+v B=%+v S=%+v", nestedA.Bounds, nestedB.Bounds, sibling.Bounds)
	}
	if math.Abs(float64(nestedB.Bounds.X-sibling.Bounds.X-sibling.Bounds.W-8)) > .01 {
		t.Fatalf("nested group gap=%v", nestedB.Bounds.X-sibling.Bounds.X-sibling.Bounds.W)
	}
	rtl = false
	tt.Frame()
	a, _ = tt.AccessByRole(ui.RoleButton, "Start")
	b, _ = tt.AccessByRole(ui.RoleButton, "End")
	if a.Bounds.X >= b.Bounds.X {
		t.Fatal("direction did not update")
	}
	independent, _ := tt.AccessByRole(ui.RoleButton, "Independent")
	if independent.Bounds.W <= 0 {
		t.Fatal("group scope leaked to following control")
	}
}

func TestFieldCompoundInputAndAddonActions(t *testing.T) {
	value := ""
	calls := 0
	tt := tester(func(c *ui.Context) {
		Field(c, FieldProps{Label: "Website", Required: true, Description: "Your public page", Error: "Enter a URL", Orientation: FieldHorizontal, LabelWidth: 80}, func() ui.Element {
			return InputGroup(c, InputGroupProps{Prefix: "https://", BlockEnd: func() {
				InputGroupButton(c, "Apply", ButtonProps{}).Label("Apply URL")
				if Button(c, "Save URL", ButtonProps{Size: ExtraSmall}).Clicked() {
					calls++
				}
			}}, func() ui.Element { return Input(c, &value, InputProps{}) })
		})
	})
	field, ok := tt.AccessByRole(ui.RoleTextField, "Website *")
	if !ok || !field.Invalid || !slices.Contains(field.ErrorMessage, "Enter a URL") || !slices.Contains(field.DescribedBy, "Your public page") {
		t.Fatalf("compound field access %+v", field)
	}
	click(t, tt, "Website *")
	tt.Type("example.com")
	if value != "example.com" {
		t.Fatalf("label did not focus compound editor: %q", value)
	}
	click(t, tt, "Save URL")
	if calls != 1 {
		t.Fatalf("addon action calls=%d", calls)
	}
	// Click on non-interactive prefix text focuses the same input, without
	// stealing activation from an embedded button.
	click(t, tt, "https://")
	tt.Type("/docs")
	if value != "example.com/docs" {
		t.Fatalf("addon text did not focus editor: %q", value)
	}
}

func TestFieldSetAndStandaloneLabel(t *testing.T) {
	checked := false
	disabled := true
	tt := tester(func(c *ui.Context) {
		FieldSet(c, FieldSetProps{Legend: "Preferences", Description: "Your notification settings", Disabled: disabled}, func() { e := Checkbox(c, &checked, "", CheckboxProps{}); LabelFor(c, "Send updates", e) })
	})
	group, ok := tt.AccessByRole(ui.RoleGroup, "Preferences")
	if !ok || group.Description != "Your notification settings" {
		t.Fatalf("fieldset access %+v", group)
	}
	click(t, tt, "Send updates")
	if checked {
		t.Fatal("disabled fieldset toggled checkbox")
	}
	disabled = false
	tt.Frame()
	click(t, tt, "Send updates")
	if !checked {
		t.Fatal("standalone label did not toggle checkbox")
	}

	peerDisabled, ancestorDisabled, peerChecked := false, false, false
	draft, changes, toggles := "draft", 0, 0
	var selection [2]int
	var selectedCaptions []ui.SelectedText
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		paragraphs := ui.Box(c).Fill().Background(ui.RGB(0, 0, 0)).Selectable().Children(func() {
			ui.Column(c).Fill().Padding(20).Gap(8).LineHeight(2).Disabled(ancestorDisabled).Children(func() {
				Label(c, "Source line height").Label("Source line height")
				ui.Text(c, "Reference line height").FontSize(14).FontWeight(500).LineHeight(1).Label("Reference line height")
				e := Checkbox(c, &peerChecked, "", CheckboxProps{Disabled: peerDisabled, OnCheckedChange: func(bool) { toggles++ }})
				LabelFor(c, "Peer caption", e).Background(ui.RGB(255, 255, 255))
				input := Input(c, &draft, InputProps{OnChange: func(string) { changes++ }})
				selection[0], selection[1] = input.TextSelection()
				LabelFor(c, "Editor caption", input).Background(ui.RGB(255, 255, 255))
			})
		})
		selectedCaptions = paragraphs.SelectedTexts()
	}, 400, 280)
	text := func(name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(ui.RoleText, name)
		if !ok {
			t.Fatalf("missing actual source Label %q", name)
		}
		return a
	}
	if text("Source line height").Bounds.H != text("Reference line height").Bounds.H {
		t.Fatal("plain Label inherited the parent's double line-height instead of source leading-none")
	}
	captionPixel := func(want uint8) {
		t.Helper()
		a := text("Peer caption")
		p := color.NRGBAModel.Convert(r.Image().At(int(a.Bounds.X+a.Bounds.W-4), int(a.Bounds.Y+a.Bounds.H/2))).(color.NRGBA)
		if p.R != want || p.G != want || p.B != want || p.A != 255 {
			t.Fatalf("peer=%v ancestor=%v caption pixel=%v want=%d", peerDisabled, ancestorDisabled, p, want)
		}
	}
	clickText := func(name string) {
		t.Helper()
		a := text(name)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	captionPixel(255)
	clickText("Editor caption")
	if !r.Focused("Editor caption") {
		t.Fatal("standalone Label failed to focus the real native editor")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("native draft")
	r.Key(ui.Cmd, ui.KeyA)
	r.Frame()
	before := changes
	clickText("Editor caption")
	if draft != "native draft" || changes != before || changes == 0 || selection != [2]int{0, len(draft)} || !r.Focused("Editor caption") {
		t.Fatal("Label focus replaced the editor, changed its value, or lost its native selection")
	}
	clickText("Peer caption")
	if !peerChecked || toggles != 1 {
		t.Fatal("plain peer Label failed one actual native checkbox activation")
	}
	peerChecked, peerDisabled = false, true
	r.Frame()
	captionPixel(128)
	clickText("Peer caption")
	r.Key(0, ui.KeySpace)
	if peerChecked || toggles != 1 {
		t.Fatal("disabled peer Label activated its native checkbox")
	}
	peerDisabled, ancestorDisabled = false, true
	r.Frame()
	captionPixel(128)
	clickText("Peer caption")
	r.Key(0, ui.KeySpace)
	if peerChecked || toggles != 1 {
		t.Fatal("ancestor-disabled Label activated its native checkbox")
	}
	ancestorDisabled = false
	r.Frame()
	captionPixel(255)
	clickText("Peer caption")
	if !peerChecked || toggles != 2 {
		t.Fatal("removing disablement failed to restore the original Label association")
	}
	clickText("Reference line height")
	r.Key(ui.Cmd, ui.KeyA)
	r.Frame()
	if len(selectedCaptions) != 1 || selectedCaptions[0].Text != "Reference line height" {
		t.Fatalf("source select-none Label joined ancestor text selection: %+v", selectedCaptions)
	}
}

func TestResponsiveFieldsUseTheirContainerWidth(t *testing.T) {
	width := float32(300)
	value := ""
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(width).Children(func() {
			FieldGroup(c, func() {
				Field(c, FieldProps{Label: "Caption", Orientation: FieldResponsive, LabelWidth: 100}, func() ui.Element { return Input(c, &value, InputProps{}) })
			})
		})
	}, 800, 300)
	label, _ := tt.AccessByRole(ui.RoleText, "Caption")
	input, _ := tt.AccessByRole(ui.RoleTextField, "Caption")
	if input.Bounds.Y < label.Bounds.Y+label.Bounds.H {
		t.Fatalf("narrow field did not stack: label=%+v input=%+v", label.Bounds, input.Bounds)
	}
	width = 600
	tt.Frame()
	label, _ = tt.AccessByRole(ui.RoleText, "Caption")
	input, _ = tt.AccessByRole(ui.RoleTextField, "Caption")
	if input.Bounds.X < label.Bounds.X+100 {
		t.Fatalf("wide field did not become horizontal: label=%+v input=%+v", label.Bounds, input.Bounds)
	}
}

func TestGroupedSelectCustomContentAndScrolling(t *testing.T) {
	value := ""
	options := []Option{{Value: "go", Label: "Go", Group: "Languages"}, {Separator: true}, {Value: "disabled", Label: "Unavailable", Disabled: true, Group: "Other"}, {Value: "rust", Label: "Rust", Group: "Other"}}
	tt := tester(func(c *ui.Context) {
		Select(c, &value, SelectProps{Placeholder: "Choose language", Options: options, ItemContent: func(c *ui.Context, o Option) { ui.Text(c, "★ "+o.Label) }}).Width(240)
	})
	click(t, tt, "Choose language")
	if !tt.HasText("Languages") || !tt.HasText("Other") {
		t.Fatal("select group headings missing")
	}
	if _, ok := tt.AccessByRole(ui.RoleOption, "★ Rust"); ok {
		t.Fatal("custom decoration changed accessible option label")
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(time.Second)
	if value != "rust" {
		t.Fatalf("grouped select chose %q", value)
	}
}

func TestNativeSelectUsesSystemMenuAndControlledValue(t *testing.T) {
	value := "go"
	disabled := false
	changes := 0
	tt := tester(func(c *ui.Context) {
		c.SetMenuPresentation(ui.PopoverMenus)
		if NativeSelect(c, &value, SelectProps{Label: "Native language", Disabled: disabled, Options: []Option{
			{Value: "go", Label: "Go", Group: "Languages"},
			{Value: "off", Label: "Unavailable", Disabled: true},
			{Separator: true}, {Value: "rust", Label: "Rust"},
		}}).Changed() {
			changes++
		}
		if c.MenuPresentation() != ui.PopoverMenus {
			t.Fatal("native select leaked its menu presentation policy")
		}
	})
	if _, ok := tt.AccessByRole(ui.RoleMenuButton, "Native language"); !ok {
		t.Fatal("native select is not a native menu button")
	}
	click(t, tt, "Native language")
	if !slices.Equal(tt.Menu(), []string{"Languages", "Go", "Unavailable", "-", "Rust"}) {
		t.Fatalf("native menu = %v", tt.Menu())
	}
	if tt.HasText("Rust") {
		t.Fatal("system options were painted inside the window")
	}
	if err := tt.ChooseMenuItem("Unavailable"); err == nil {
		t.Fatal("disabled native option was selectable")
	}
	if err := tt.ChooseMenuItem("Rust"); err != nil {
		t.Fatal(err)
	}
	if value != "rust" || changes == 0 {
		t.Fatalf("native choice=%q changes=%v", value, changes)
	}
	if len(tt.Menu()) != 0 {
		t.Fatal("native popup did not dismiss")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Native language")
	if len(tt.Menu()) != 0 {
		t.Fatal("disabled native select opened a menu")
	}
}

func TestMultiComboboxSelectionRemovalAndLimit(t *testing.T) {
	selected := []string{}
	query := ""
	disabled := false
	tt := tester(func(c *ui.Context) {
		MultiCombobox(c, &selected, &query, MultiComboboxProps{Label: "Languages", Placeholder: "Search languages", MaxSelected: 2, Disabled: disabled, Options: []Option{{Value: "go", Label: "Go", Group: "Languages"}, {Value: "off", Label: "Unavailable", Disabled: true}, {Value: "rust", Label: "Rust", Group: "Languages"}, {Value: "swift", Label: "Swift", Group: "Languages"}}})
	})
	input, ok := tt.AccessByRole(ui.RoleComboBox, "Languages")
	if !ok {
		t.Fatal("missing native combo box")
	}
	tt.ClickAt(input.Bounds.X+8, input.Bounds.Y+8)
	tt.Type("go")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !slices.Equal(selected, []string{"go"}) || query != "" {
		t.Fatalf("first choice %v query=%q", selected, query)
	}
	tt.Type("rust")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !slices.Equal(selected, []string{"go", "rust"}) {
		t.Fatalf("second choice %v query=%q texts=%v focused=%v", selected, query, tt.Texts(), tt.Focused("Languages"))
	}
	tt.Type("swift")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(selected) != 2 {
		t.Fatalf("max selected ignored: %v", selected)
	}
	query = ""
	tt.Frame()
	tt.Key(0, ui.KeyBackspace)
	tt.Frame()
	if !slices.Equal(selected, []string{"go"}) {
		t.Fatalf("empty backspace=%v", selected)
	}
	click(t, tt, "Remove Go")
	if len(selected) != 0 {
		t.Fatalf("chip remove=%v", selected)
	}
	selected = []string{"go"}
	disabled = true
	tt.Frame()
	click(t, tt, "Remove Go")
	if len(selected) != 1 {
		t.Fatal("disabled chip removed")
	}
	for _, node := range tt.AccessAll() {
		r := node.Bounds
		if math.IsNaN(float64(r.X + r.Y + r.W + r.H)) {
			t.Fatal("nonfinite bounds")
		}
	}
}
