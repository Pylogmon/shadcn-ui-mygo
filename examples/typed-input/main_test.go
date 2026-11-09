package main

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
)

func exampleInput(s *appState) *ui.Tester {
	return ui.NewTesterWith(s.view, 880, 720, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func clickRole(t *testing.T, r *ui.Tester, role ui.Role, name string) {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing %v %q", role, name)
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
}
func enter(t *testing.T, r *ui.Tester, name, value string) {
	t.Helper()
	clickRole(t, r, ui.RoleTextField, name)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type(value)
	r.Frame()
}
func await(t *testing.T, r *ui.Tester, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !ready() && time.Now().Before(deadline) {
		r.Frame()
		time.Sleep(time.Millisecond)
	}
	if !ready() {
		t.Fatal("file result not applied")
	}
}
func TestFixedInputExamplesAndFieldStates(t *testing.T) {
	for _, section := range []string{"types", "states", "composition", "form"} {
		t.Run(section, func(t *testing.T) {
			r := exampleInput(newState(false, section))
			if len(r.AccessAll()) == 0 {
				t.Fatal("empty example")
			}
			switch section {
			case "types":
				for _, name := range []string{"Password", "Phone", "URL", "Search", "Number"} {
					if _, ok := r.AccessByRole(ui.RoleTextField, name); !ok {
						t.Fatal("missing input type", name)
					}
				}
				secure, _ := r.AccessByRole(ui.RoleTextField, "Password")
				search, _ := r.AccessByRole(ui.RoleTextField, "Search")
				if !secure.Password || !search.Search {
					t.Fatal("password/search native semantics lost")
				}
				for _, item := range []struct {
					role ui.Role
					name string
				}{{ui.RolePopUpButton, "Date"}, {ui.RoleGroup, "Time"}, {ui.RoleButton, "File"}} {
					if _, ok := r.AccessByRole(item.role, item.name); !ok {
						t.Fatal("missing specialized type", item.name)
					}
				}
			case "states":
				a, _ := r.AccessByRole(ui.RoleTextField, "Invalid input")
				b, _ := r.AccessByRole(ui.RoleTextField, "Disabled email")
				if !a.Invalid || !b.Disabled {
					t.Fatal("source states lost")
				}
			case "composition":
				a, _ := r.AccessByRole(ui.RoleTextField, "Amount")
				if a.Bounds.W <= 0 {
					t.Fatal("compound number missing actual editor")
				}
			case "form":
				if _, ok := r.AccessByRole(ui.RoleButton, "Submit"); !ok {
					t.Fatal("form has no action")
				}
			}
		})
	}
}
func TestInputTypesNativeWorkflowAndReview(t *testing.T) {
	s := newState(false, "types")
	s.picker = func(mygo.OpenDialogOptions) ([]string, error) { return []string{"/tmp/example.txt"}, nil }
	r := exampleInput(s)
	r.SetRegion(i18n.Region{Locale: "en-US", HourCycle: 24, FirstWeekday: time.Monday})
	enter(t, r, "Number", "42")
	r.Key(0, ui.KeyEnter)
	clickRole(t, r, ui.RolePopUpButton, "Date")
	if err := r.Click("November 12, 2026"); err != nil {
		t.Fatal(err)
	}
	clickRole(t, r, ui.RoleStepper, "Time hours")
	r.Key(0, ui.KeyUp)
	clickRole(t, r, ui.RoleButton, "File")
	await(t, r, func() bool { return len(s.paths) == 1 })
	clickRole(t, r, ui.RoleButton, "Review values")
	if s.number != 42 || s.date.Day() != 12 || s.clock.Hour() != 16 || !strings.Contains(s.status, "Number 42 · 2026-11-12 · 16:30 · 1 attachment(s)") {
		t.Fatalf("actual native review %s", s.status)
	}
}
func TestInputCompositionActionsAndApplicationValidation(t *testing.T) {
	s := newState(false, "composition")
	r := exampleInput(s)
	if err := r.Click("$"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("25.5")
	if s.number != 25.5 {
		t.Fatal("InputGroup prefix did not focus real numeric editor")
	}
	enter(t, r, "Search text", "native form")
	clickRole(t, r, ui.RoleButton, "Search")
	if s.status != "Search: native form" {
		t.Fatal("source search button is inert")
	}
	s.section = "types"
	r.Frame()
	enter(t, r, "URL", "broken")
	a, _ := r.AccessByRole(ui.RoleTextField, "URL")
	if !a.Invalid {
		t.Fatal("URL application validation not linked")
	}
	enter(t, r, "URL", "https://example.com")
	a, _ = r.AccessByRole(ui.RoleTextField, "URL")
	if a.Invalid {
		t.Fatal("valid URL did not clear error")
	}
}
func TestInputFormValidationSubmitAndCancel(t *testing.T) {
	s := newState(false, "form")
	r := exampleInput(s)
	clickRole(t, r, ui.RoleButton, "Submit")
	name, _ := r.AccessByRole(ui.RoleTextField, "Name")
	email, _ := r.AccessByRole(ui.RoleTextField, "Email")
	if !name.Invalid || !email.Invalid {
		t.Fatal("blank form did not show field errors")
	}
	enter(t, r, "Name", "Ada")
	enter(t, r, "Email", "ada@example.com")
	enter(t, r, "Phone", "+1 (555) 123-4567")
	enter(t, r, "Address", "123 Main St")
	clickRole(t, r, ui.RoleButton, "Submit")
	if s.status != "Submitted for Ada · ada@example.com" {
		t.Fatal("form did not use current values", s.status)
	}
	clickRole(t, r, ui.RoleButton, "Cancel")
	if s.name != "" || s.email != "" || s.address != "" || s.attempted || s.status != "Form cleared." {
		t.Fatal("cancel did not reset form")
	}
}
