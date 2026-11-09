package control

import (
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestTextareaResizeInputGroupKeepsNativeFieldTargetAndActions(t *testing.T) {
	value := "draft"
	state := TextareaResizeState{Height: 100}
	changes := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(360).Padding(20).Children(func() {
			Field(c, FieldProps{Label: "Message", Error: "Review this message"}, func() ui.Element {
				return InputGroup(c, InputGroupProps{Prefix: "Note"}, func() ui.Element {
					return TextareaWith(c, &value, TextareaProps{OnChange: func(string) { changes++ }}, TextareaResizeProps{State: &state, MaxHeight: 220, Label: "Resize message"}).Root
				})
			})
		})
	}, 500, 400)
	field := textareaResizeAccess(t, r, ui.RoleTextField, "Message")
	handle := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize message")
	if !field.Multiline || !field.Invalid || len(field.ErrorMessage) != 1 || len(handle.Controls) != 1 || handle.Controls[0] != field.Name {
		t.Fatalf("compound group lost the actual native target: field=%+v handle=%+v", field, handle)
	}
	click(t, r, "Note")
	if !r.Focused("Message") {
		t.Fatal("compound addon focused its visual group instead of the native editor")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("message")
	if value != "message" || changes != 1 {
		t.Fatal("compound editor lost native value/callback")
	}
	click(t, r, "Resize message")
	r.Key(0, ui.KeyDown)
	if state.Height <= 100 || textareaResizeAccess(t, r, ui.RoleTextField, "Message").Bounds.H != state.Height {
		t.Fatal("group wrapper broke sibling resize handle")
	}
	click(t, r, "Note")
	r.Key(ui.Cmd, ui.KeyZ)
	if value != "draft" {
		t.Fatal("compound resize lost native undo")
	}
}

func TestInputGroupAddonDoesNotFocusDisabledNativeEditor(t *testing.T) {
	outside, locked := "", "locked"
	r := tester(func(c *ui.Context) {
		Input(c, &outside, InputProps{Label: "Outside", AutoFocus: true})
		InputGroup(c, InputGroupProps{Prefix: "Locked prefix"}, func() ui.Element {
			return Input(c, &locked, InputProps{Label: "Locked editor", Disabled: true})
		})
	})
	click(t, r, "Locked prefix")
	if r.Focused("Locked editor") {
		t.Fatal("addon explicitly focused a disabled native editor")
	}
}

func TestTextareaResizeInputGroupRemovesInnerEditorBorder(t *testing.T) {
	value := ""
	var background ui.Color
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		background = theme.TokensOf(c).Background
		ui.Column(c).Width(360).Padding(20).Background(background).Children(func() {
			InputGroup(c, InputGroupProps{Prefix: "Note"}, func() ui.Element {
				return TextareaWith(c, &value, TextareaProps{Label: "Notes"}).Root
			})
		})
	}, 500, 200)
	field := textareaResizeAccess(t, r, ui.RoleTextField, "Notes")
	pixel := r.Image().RGBAAt(int(field.Bounds.X), int(field.Bounds.Y+field.Bounds.H/2))
	want := color.RGBA{R: background.R, G: background.G, B: background.B, A: 255}
	if pixel != want {
		t.Fatalf("compound group left a second border on its native editor: pixel=%v background=%v", pixel, want)
	}
}
