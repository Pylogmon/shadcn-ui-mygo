package control_test

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func errorAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing %v %q", role, name)
	}
	return a
}

func TestFieldErrorsLinkDeduplicatedListAndPreserveTheEditor(t *testing.T) {
	value, other := "", ""
	messages := []string{"Choose a different name", "", "Use at least four characters", "Choose a different name"}
	original := slices.Clone(messages)
	var handle ui.Handle
	var anchor, caret int
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(16).Children(func() {
			control.Field(c, control.FieldProps{Label: "Account name", Required: true, Description: "Your public name", Error: "Choose a different name", Errors: messages}, func() ui.Element {
				return control.InputGroup(c, control.InputGroupProps{Prefix: "@"}, func() ui.Element {
					return control.Input(c, &value, control.InputProps{Handle: &handle})
				})
			})
			control.Input(c, &other, control.InputProps{Label: "Unrelated"})
		})
		anchor, caret = c.Resolve(handle).TextSelection()
	}, 460, 360, ui.TesterOptions{})
	const summary = "Choose a different name\nUse at least four characters"
	a := errorAccess(t, r, ui.RoleTextField, "Account name *")
	if !a.Invalid || !slices.Equal(a.ErrorMessage, []string{summary}) || !slices.Contains(a.DescribedBy, "Your public name") || !slices.Contains(a.DescribedBy, summary) {
		t.Fatalf("multiple errors are not linked to the real group editor: %+v", a)
	}
	errorAccess(t, r, ui.RoleAlert, summary)
	errorAccess(t, r, ui.RoleList, "Validation errors")
	for index, message := range []string{"Choose a different name", "Use at least four characters"} {
		item := errorAccess(t, r, ui.RoleListItem, message)
		if item.PosInSet != index+1 || item.SetSize != 2 {
			t.Fatalf("list order/count lost: %+v", item)
		}
	}
	if !slices.Equal(messages, original) || errorAccess(t, r, ui.RoleTextField, "Unrelated").Invalid {
		t.Fatal("error rendering mutated validator output or leaked field appearance")
	}
	if err := r.Click("Account name *"); err != nil {
		t.Fatal(err)
	}
	r.Type("alex")
	r.Key(0, ui.KeyLeft)
	messages = []string{"Choose a different name"}
	r.Frame()
	if !r.Focused("Account name *") || value != "alex" || anchor != 3 || caret != 3 {
		t.Fatal("changing list content replaced the editor or its selection")
	}
	if _, ok := r.AccessByRole(ui.RoleList, "Validation errors"); ok {
		t.Fatal("one deduplicated message retained a list")
	}
	// The legacy Error stays valid independently of the optional error slice.
	messages = nil
	r.Frame()
	if !errorAccess(t, r, ui.RoleTextField, "Account name *").Invalid {
		t.Fatal("an empty slice suppressed the legacy single error")
	}
}

func TestFieldErrorCustomContentAndClearingUseCurrentNativeRelations(t *testing.T) {
	value := "draft"
	errors := []string{"Default message"}
	custom, generation, chosen := true, 1, 0
	standalone := []string{""}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(16).Children(func() {
			control.FieldError(c, control.FieldErrorProps{Errors: standalone})
			props := control.FieldProps{Label: "Profile notes", Description: "Keep your draft", Errors: errors}
			if custom {
				current := generation
				props.ErrorLabel = "Review the profile requirements"
				props.ErrorContent = func(child *ui.Context) {
					ui.Text(child, "Profile requirements").FontWeight(500)
					if control.Button(child, "Show requirements", control.ButtonProps{Variant: control.Link}).Clicked() {
						chosen = current
					}
				}
			}
			control.Field(c, props, func() ui.Element {
				return control.InputGroup(c, control.InputGroupProps{BlockEnd: func() { ui.Text(c, "Local draft") }}, func() ui.Element {
					return control.Textarea(c, &value, control.TextareaProps{Rows: 3})
				})
			})
		})
	}, 460, 360, ui.TesterOptions{})
	a := errorAccess(t, r, ui.RoleTextField, "Profile notes")
	if !a.Invalid || !a.Multiline || !slices.Equal(a.ErrorMessage, []string{"Review the profile requirements"}) || r.HasText("Default message") {
		t.Fatal("custom content did not replace the default message with an accessible summary")
	}
	generation = 2
	errors = nil
	r.Frame()
	if err := r.Click("Show requirements"); err != nil {
		t.Fatal(err)
	}
	if chosen != 2 {
		t.Fatal("the error action used an obsolete content builder")
	}
	custom = false
	errors = []string{"Default message"}
	r.Frame()
	errorAccess(t, r, ui.RoleAlert, "Default message")
	if r.HasText("Profile requirements") || !slices.Equal(errorAccess(t, r, ui.RoleTextField, "Profile notes").ErrorMessage, []string{"Default message"}) {
		t.Fatal("custom content or its old error relation survived replacement")
	}
	errors = []string{"", ""}
	standalone = nil
	r.Frame()
	a = errorAccess(t, r, ui.RoleTextField, "Profile notes")
	if a.Invalid || len(a.ErrorMessage) != 0 || !slices.Equal(a.DescribedBy, []string{"Keep your draft"}) || value != "draft" {
		t.Fatalf("cleared errors left stale metadata or lost the draft: %+v", a)
	}
	if _, ok := r.AccessByRole(ui.RoleAlert, ""); ok {
		t.Fatal("empty error arrays exposed an empty native alert")
	}
	standalone = []string{"Standalone issue", "Standalone issue"}
	r.Frame()
	errorAccess(t, r, ui.RoleAlert, "Standalone issue")
	if _, ok := r.AccessByRole(ui.RoleList, "Validation errors"); ok {
		t.Fatal("a standalone single issue was incorrectly rendered as a list")
	}
}
