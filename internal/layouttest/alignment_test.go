package layouttest_test

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type alignedComponent struct {
	name       string
	build      func(*ui.Context, ui.Align) ui.Element
	legacy     func(*ui.Context) ui.Element
	scope      func(*ui.Context, func(*ui.Context))
	fixedWidth bool
	editor     bool
}

func alignmentComponents() []alignedComponent {
	keys := func(c *ui.Context) { display.Kbd(c, "Ctrl"); display.Kbd(c, "K") }
	items := []control.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
	date := time.Date(2026, time.November, 11, 0, 0, 0, 0, time.UTC)
	return []alignedComponent{
		{name: "Checkbox", build: func(c *ui.Context, a ui.Align) ui.Element {
			return control.Checkbox(c, new(bool), "Accept terms", control.CheckboxProps{AlignSelf: a, Description: "Read the terms\nbefore continuing"})
		}},
		{name: "CheckboxState", build: func(c *ui.Context, a ui.Align) ui.Element {
			return control.CheckboxState(c, new(ui.CheckState), "Select all", control.CheckboxProps{AlignSelf: a})
		}},
		{name: "Radio", build: func(c *ui.Context, a ui.Align) ui.Element {
			return control.RadioWith(c, new(string), "a", "Alpha", control.RadioProps{AlignSelf: a, Description: "First option"})
		}, legacy: func(c *ui.Context) ui.Element { return control.Radio(c, new(string), "a", "Alpha", false) }},
		{name: "Toggle", build: func(c *ui.Context, a ui.Align) ui.Element {
			return control.Toggle(c, new(bool), "Bold", control.ToggleProps{AlignSelf: a})
		}},
		{name: "ToggleGroup", build: func(c *ui.Context, a ui.Align) ui.Element {
			return control.ToggleGroup(c, new([]string), control.ToggleGroupProps{AlignSelf: a, Items: items})
		}},
		{name: "ToggleGroup vertical", build: func(c *ui.Context, a ui.Align) ui.Element {
			gap := float32(0)
			return control.ToggleGroup(c, new([]string), control.ToggleGroupProps{AlignSelf: a, Items: items, Vertical: true, Spacing: &gap})
		}},
		{name: "Badge", build: func(c *ui.Context, a ui.Align) ui.Element {
			return display.BadgeWith(c, "Ready", display.BadgeProps{AlignSelf: a})
		}, legacy: func(c *ui.Context) ui.Element { return display.Badge(c, "Ready", display.BadgeDefault) }},
		{name: "Badge button", build: func(c *ui.Context, a ui.Align) ui.Element {
			return display.BadgeWith(c, "Run", display.BadgeProps{AlignSelf: a, Button: true})
		}},
		{name: "Badge compound link", build: func(c *ui.Context, a ui.Align) ui.Element {
			return display.BadgeWith(c, "Guide", display.BadgeProps{AlignSelf: a, URL: "https://example.com", Leading: func(c *ui.Context) { ui.Box(c).Size(12, 12) }})
		}},
		{name: "Kbd", build: func(c *ui.Context, a ui.Align) ui.Element {
			return display.KbdWith(c, "K", display.KbdProps{AlignSelf: a})
		}, legacy: func(c *ui.Context) ui.Element { return display.Kbd(c, "K") }},
		{name: "KbdGroup", build: func(c *ui.Context, a ui.Align) ui.Element {
			return display.KbdGroupWith(c, display.KbdGroupProps{AlignSelf: a}, keys)
		}, legacy: func(c *ui.Context) ui.Element { return display.KbdGroup(c, keys) }},
		{name: "KbdShortcut", build: func(c *ui.Context, a ui.Align) ui.Element {
			return display.KbdShortcutWith(c, ui.Cmd|ui.Shift, "K", display.KbdGroupProps{AlignSelf: a})
		}, legacy: func(c *ui.Context) ui.Element { return display.KbdShortcut(c, ui.Cmd|ui.Shift, "K") }},
		{name: "InputOTP", fixedWidth: true, editor: true, build: func(c *ui.Context, a ui.Align) ui.Element {
			return control.InputOTP(c, new(string), control.InputOTPProps{AlignSelf: a, Groups: []int{3, 3}})
		}},
		{name: "CalendarWith", build: func(c *ui.Context, a ui.Align) ui.Element {
			value := date
			return display.CalendarWith(c, &value, display.CalendarProps{AlignSelf: a})
		}},
		{name: "CalendarMultiple", build: func(c *ui.Context, a ui.Align) ui.Element {
			value := []time.Time{date}
			return display.CalendarMultiple(c, &value, display.CalendarProps{AlignSelf: a})
		}},
		{name: "CalendarRange", build: func(c *ui.Context, a ui.Align) ui.Element {
			value := display.DateRange{From: date}
			return display.CalendarRange(c, &value, display.CalendarProps{AlignSelf: a})
		}},
		{name: "InlineCode", build: func(c *ui.Context, a ui.Align) ui.Element {
			return display.TypographyWith(c, "ui.Context", display.InlineCode, display.TypographyProps{AlignSelf: a})
		}, legacy: func(c *ui.Context) ui.Element { return display.Typography(c, "ui.Context", display.InlineCode) }},
		{name: "TypographyCode", build: func(c *ui.Context, a ui.Align) ui.Element {
			return display.TypographyCodeWith(c, "ui.Context", display.TypographyProps{AlignSelf: a})
		}, legacy: func(c *ui.Context) ui.Element { return display.TypographyCode(c, "ui.Context") }},
		{name: "Attachment", build: func(c *ui.Context, a ui.Align) ui.Element {
			return conversation.Attachment(c, conversation.AttachmentProps{AlignSelf: a, Name: "report.pdf", MediaType: "application/pdf", SizeBytes: 2048})
		}},
		{name: "AttachmentWith vertical", fixedWidth: true, build: func(c *ui.Context, a ui.Align) ui.Element {
			return conversation.AttachmentWith(c, conversation.AttachmentProps{AlignSelf: a, Name: "report.pdf", Orientation: ui.Vertical}, func(parts *conversation.AttachmentParts) {
				parts.Media(conversation.AttachmentIconMedia, nil)
				parts.Content(func(c *ui.Context) { ui.Text(c, "Report") })
			}).Root
		}},
		{name: "TabsList", scope: func(c *ui.Context, content func(*ui.Context)) {
			navigation.TabsRoot(c, new(string), navigation.TabsRootProps{}, content).Fill()
		}, build: func(c *ui.Context, a ui.Align) ui.Element {
			return navigation.TabsList(c, navigation.TabsListProps{AlignSelf: a}, func(c *ui.Context) {
				navigation.TabsTrigger(c, "a", navigation.TabsTriggerProps{Label: "Alpha"})
				navigation.TabsTrigger(c, "b", navigation.TabsTriggerProps{Label: "Beta"})
			})
		}},
		{name: "TabsList vertical line", scope: func(c *ui.Context, content func(*ui.Context)) {
			navigation.TabsRoot(c, new(string), navigation.TabsRootProps{Vertical: true}, content).Fill()
		}, build: func(c *ui.Context, a ui.Align) ui.Element {
			return navigation.TabsList(c, navigation.TabsListProps{AlignSelf: a, Variant: navigation.TabsLine}, func(c *ui.Context) {
				navigation.TabsTrigger(c, "a", navigation.TabsTriggerProps{Label: "Alpha"})
				navigation.TabsTrigger(c, "b", navigation.TabsTriggerProps{Label: "Beta"})
			})
		}},
	}
}

func requireNear(t *testing.T, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > .1 {
		t.Fatalf("got %g, want %g", got, want)
	}
}

func findBounds(t *testing.T, r *ui.Tester, label string) ui.Rect {
	t.Helper()
	bounds, ok := r.Find(label)
	if !ok || bounds.W <= 0 || bounds.H <= 0 {
		t.Fatalf("missing layout bounds for %q: %+v", label, bounds)
	}
	return bounds
}

// Measure real native layout rather than just checking a style setter. In
// particular, InputOTP's returned editor must move with its outer slot box.
func TestComponentOuterAlignment(t *testing.T) {
	for _, component := range alignmentComponents() {
		t.Run(component.name, func(t *testing.T) {
			for _, layout := range []struct {
				name      string
				column    bool
				direction ui.Direction
			}{
				{"row", false, ui.LeftToRight}, {"column", true, ui.LeftToRight},
				{"row RTL", false, ui.RightToLeft}, {"column RTL", true, ui.RightToLeft},
			} {
				t.Run(layout.name, func(t *testing.T) {
					align, fluent, legacy, natural := ui.Start, false, false, false
					view := func(c *ui.Context) {
						parent := ui.Row(c).Width(900).Gap(12).AlignItems(ui.Center).Direction(layout.direction).Label("Parent")
						if !natural {
							parent.Height(500)
						}
						if layout.column {
							parent.Column()
						}
						parent.Children(func() {
							ui.Text(c, "First line\nSecond line\nThird line\nFourth line").Width(180).FontSize(16).LineHeight(1.5).Label("Multiline peer")
							var element ui.Element
							if legacy {
								element = component.legacy(c)
							} else {
								element = component.build(c, align)
							}
							if fluent {
								element.AlignSelf(ui.Center)
							}
							element.Label("Measured component")
						})
					}
					r := ui.NewTesterWith(func(c *ui.Context) {
						theme.Apply(c)
						ui.Column(c).Fill().Padding(24).Children(func() {
							if component.scope != nil {
								component.scope(c, view)
							} else {
								view(c)
							}
						})
					}, 1000, 600, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
					initial := findBounds(t, r, "Measured component")
					if initial.W >= findBounds(t, r, "Parent").W {
						t.Fatalf("component should keep its content or slot width: %+v", initial)
					}
					check := func(t *testing.T, want ui.Align) {
						t.Helper()
						bounds, parent := findBounds(t, r, "Measured component"), findBounds(t, r, "Parent")
						position, size, origin, available := bounds.Y, bounds.H, parent.Y, parent.H
						if layout.column {
							position, size, origin, available = bounds.X, bounds.W, parent.X, parent.W
							if layout.direction == ui.RightToLeft && want != ui.Center {
								if want == ui.Start {
									want = ui.End
								} else {
									want = ui.Start
								}
							}
						}
						offset := float32(0)
						if want == ui.Center {
							offset = (available - size) / 2
						} else if want == ui.End {
							offset = available - size
						}
						requireNear(t, position, origin+offset)
					}
					t.Run("default remains Start", func(t *testing.T) { check(t, ui.Start) })
					for _, setting := range []struct {
						name  string
						align ui.Align
					}{{"Center", ui.Center}, {"End", ui.End}} {
						t.Run(setting.name, func(t *testing.T) {
							align = setting.align
							r.Frame()
							check(t, align)
							bounds := findBounds(t, r, "Measured component")
							requireNear(t, bounds.W, initial.W)
							requireNear(t, bounds.H, initial.H)
						})
					}
					if component.legacy != nil {
						t.Run("legacy remains Start", func(t *testing.T) {
							legacy = true
							r.Frame()
							check(t, ui.Start)
						})
					}
					if !component.editor {
						t.Run("fluent override", func(t *testing.T) {
							fluent = true
							r.Frame()
							check(t, ui.Center)
						})
					}
					fluent, legacy, align = false, false, ui.Center
					if !layout.column {
						t.Run("natural height beside multiline text", func(t *testing.T) {
							natural = true
							r.Frame()
							bounds, peer := findBounds(t, r, "Measured component"), findBounds(t, r, "Multiline peer")
							requireNear(t, bounds.Y+bounds.H/2, peer.Y+peer.H/2)
							requireNear(t, bounds.W, initial.W)
						})
					} else {
						t.Run("Stretch respects fixed widths", func(t *testing.T) {
							align = ui.Stretch
							r.Frame()
							want := findBounds(t, r, "Parent").W
							if component.fixedWidth {
								want = initial.W
							}
							requireNear(t, findBounds(t, r, "Measured component").W, want)
						})
					}
				})
			}
		})
	}
}
