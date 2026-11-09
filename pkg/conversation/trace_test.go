package conversation

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestTraceStaticStepDoesNotCreateNativeAction(t *testing.T) {
	var state TraceState
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Trace(c, &state, []TraceStep{{ID: "static", Title: "Read configuration", Status: Complete}})
	}, 400, 180)
	for _, a := range r.AccessAll() {
		if (a.Role == ui.RoleButton || a.Role == ui.RoleDisclosure) && a.Name == "Read configuration" {
			t.Fatal("static source line became a keyboard action")
		}
	}
	if !r.HasText("Read configuration") {
		t.Fatal("static trace title disappeared")
	}
}

func traceAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing native role%d %q", role, name)
	}
	return a
}

func TestTraceCompleteAndDefaultHaveIdenticalMutedNativePixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			render := func(status TraceStatus) *ui.Tester {
				return ui.NewTesterWith(func(c *ui.Context) {
					tokens := theme.Light()
					if dark {
						tokens = theme.Dark()
					}
					theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
						TraceLine(c, TraceLineProps{Status: status}, func(child *ui.Context) { TraceTitle(child, "Completed operation"); TraceDescription(child, "14s") })
					})
				}, 360, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			}
			normal, complete := render(""), render(Complete)
			if !bytes.Equal(normal.Image().Pix, complete.Image().Pix) {
				t.Fatal("complete incorrectly changed source neutral colors or font")
			}
		})
	}
}

func TestTraceStaticGeometryTruncatesTitleAndKeepsDescription(t *testing.T) {
	title := "Edited components/settings/notification-preferences-form.tsx"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TraceGroup(c, func(group *ui.Context) {
			TraceLine(group, TraceLineProps{}, func(child *ui.Context) {
				TraceIcon(child, traceChevronIcon)
				TraceTitle(child, title)
				TraceDescription(child, "+48 −12")
				TraceIcon(child, traceChevronIcon, TraceIconProps{Suffix: true})
			})
			TraceLine(group, TraceLineProps{}, func(child *ui.Context) { TraceTitle(child, "Second row") })
		})
	}, 240, 160)
	a := traceAccess(t, r, ui.RoleText, title)
	description := traceAccess(t, r, ui.RoleText, "+48 −12")
	second := traceAccess(t, r, ui.RoleText, "Second row")
	if a.Bounds.W >= 220 || a.Bounds.H != 20 || description.Bounds.X+description.Bounds.W > 240 {
		t.Fatalf("title did not truncate around source fixed slots: title=%v description=%v", a.Bounds, description.Bounds)
	}
	if second.Bounds.Y-a.Bounds.Y-a.Bounds.H != 8 {
		t.Fatal("source trace group gap is not8DIP")
	}
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleImage || node.Role == ui.RoleButton {
			t.Fatal("decorative trace slots entered input or accessibility")
		}
	}
}

func TestTraceCollapsibleUsesSourceIndentAndRealControlledNativeRegion(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			open := true
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Details", Trigger: func(child *ui.Context) { TraceTitle(child, "Details") }, Content: func(child *ui.Context) { ui.Text(child, "Native details").Label("Native details") }})
			}, 400, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			r.SetDirection(direction)
			trigger := traceAccess(t, r, ui.RoleButton, "Details")
			panel := traceAccess(t, r, ui.RoleRegion, "Details")
			detail := traceAccess(t, r, ui.RoleText, "Native details")
			if !trigger.Expanded || len(trigger.Controls) != 1 || len(panel.DescribedBy) > 0 {
				t.Fatal("native disclosure/control relationship missing")
			}
			if math.Abs(float64(detail.Bounds.Y-trigger.Bounds.Y-trigger.Bounds.H-8)) > .1 {
				t.Fatalf("source detail mt8 missing: trigger=%v text=%v", trigger.Bounds, detail.Bounds)
			}
			if direction == ui.LeftToRight && detail.Bounds.X-panel.Bounds.X != 24 {
				t.Fatalf("source left indentation=%g want24", detail.Bounds.X-panel.Bounds.X)
			}
			if direction == ui.RightToLeft && panel.Bounds.X != 0 {
				t.Fatal("RTL changed physical panel origin")
			}
			if err := r.Click("Details"); err != nil {
				t.Fatal(err)
			}
			if open {
				t.Fatal("native disclosure did not change caller state")
			}
		})
	}
}

func TestTraceCollapsibleReversesMeasuredHeightAndClosesInputImmediately(t *testing.T) {
	open := false
	draft := ""
	var bounds ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		root := TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Output", Trigger: func(child *ui.Context) { TraceTitle(child, "Output") }, Content: func(child *ui.Context) {
			control.Input(child, &draft, control.InputProps{Label: "Detail draft"}).Width(180)
			ui.Box(child).Height(80)
		}})
		root.Draw(func(_ *ui.Painter, r ui.Rect) { bounds = r })
	}, 400, 300)
	closed := bounds.H
	if err := r.Click("Output"); err != nil {
		t.Fatal(err)
	}
	r.Advance(100 * time.Millisecond)
	middle := bounds.H
	if middle <= closed {
		t.Fatal("200ms source panel did not grow actual height")
	}
	open = false
	r.Frame()
	startClosing := bounds.H
	if _, ok := r.AccessByRole(ui.RoleTextField, "Detail draft"); ok {
		t.Fatal("closing panel remained accessible")
	}
	r.Advance(50 * time.Millisecond)
	closing := bounds.H
	if closing >= startClosing {
		t.Fatal("native close animation did not reverse smoothly")
	}
	open = true
	r.Frame()
	r.Advance(250 * time.Millisecond)
	if bounds.H <= middle || !traceAccess(t, r, ui.RoleButton, "Output").Expanded {
		t.Fatal("reversed opening did not reach natural height")
	}
	if err := r.Click("Detail draft"); err != nil {
		t.Fatal(err)
	}
	r.Type("open")
	field := traceAccess(t, r, ui.RoleTextField, "Detail draft").Bounds
	open = false
	r.Frame()
	if !r.Focused("Output") {
		t.Fatal("controlled close did not restore focus from native content")
	}
	r.ClickAt(field.X+10, field.Y+field.H/2)
	r.Type("closed")
	if draft != "open" {
		t.Fatal("closing native input accepted text")
	}
	r.Advance(250 * time.Millisecond)
	if math.Abs(float64(bounds.H-closed)) > .1 {
		t.Fatalf("source panel did not finish at collapsed height: %g want%g", bounds.H, closed)
	}
}

func TestTraceKeepMountedIsLazyAndPreservesNativeSelection(t *testing.T) {
	open := false
	draft := ""
	built := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Retained", KeepMounted: true, Trigger: func(child *ui.Context) { TraceTitle(child, "Retained") }, Content: func(child *ui.Context) {
			built++
			control.Input(child, &draft, control.InputProps{Label: "Retained draft"}).Width(180)
		}})
	}, 400, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if built != 0 {
		t.Fatal("initially closed keep-mounted content was eagerly built")
	}
	if err := r.Click("Retained"); err != nil {
		t.Fatal(err)
	}
	if err := r.Click("Retained draft"); err != nil {
		t.Fatal(err)
	}
	r.Type("selected")
	r.Key(ui.Cmd, ui.KeyA)
	open = false
	r.Frame()
	if !r.Focused("Retained") {
		t.Fatal("closing retained editor did not restore native trigger focus")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Retained draft"); ok {
		t.Fatal("retained closed editor remained accessible")
	}
	r.Key(0, ui.KeyEnter)
	r.Key(0, ui.KeyTab)
	r.Type("replacement")
	if draft != "replacement" || !r.Focused("Retained draft") {
		t.Fatalf("retained native editor lost selection/identity: %q", draft)
	}
}

func TestTraceNestedOpenPointersDuplicateAndFallbackIDsStayIndependent(t *testing.T) {
	var state TraceState
	nested := false
	steps := []TraceStep{{ID: "same", Title: "Duplicate one", Detail: "One detail"}, {ID: "same", Title: "Duplicate two", Detail: "Two detail"}, {Title: "Fallback", Detail: "Fallback detail"}, {ID: "step-2", Title: "Explicit", Detail: "Explicit detail"}, {ID: "parent", Title: "Parent", Open: new(bool), Children: []TraceStep{{ID: "nested", Title: "Nested", Open: &nested, Detail: "Nested detail"}}}}
	*steps[4].Open = true
	r := ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); Trace(c, &state, steps) }, 500, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	for _, title := range []string{"Duplicate one", "Fallback", "Nested"} {
		if err := r.Click(title); err != nil {
			t.Fatal(err)
		}
	}
	if !nested || !r.HasText("One detail") || r.HasText("Two detail") || !r.HasText("Fallback detail") || r.HasText("Explicit detail") {
		t.Fatal("duplicate/fallback/path identities linked unrelated disclosures")
	}
	if state.Expanded["same"] || state.Expanded["step-2"] {
		t.Fatal("private duplicate/fallback paths polluted compatible explicit IDs")
	}
}

func TestTraceActionsLinksDialogsAndCodeBlockUseActualNativeCapabilities(t *testing.T) {
	var state TraceState
	dialog, sheet := false, false
	calls := 0
	var code display.CodeBlockState
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		Trace(c, &state, []TraceStep{{ID: "dialog", Title: "Search results", OnAction: func() { calls++; dialog = true }}, {ID: "sheet", Title: "Explore files", OnAction: func() { sheet = true }}, {ID: "url", Title: "Open docs", URL: "https://example.com/docs"}, {ID: "code", Title: "Command output", Content: func(child *ui.Context) {
			display.CodeBlock(child, display.CodeBlockProps{Label: "Native output", Code: "alpha\nbeta", Height: 80, State: &code, HideHeader: true})
		}}})
		overlay.Dialog(c, &dialog, overlay.DialogProps{Title: "Results"}, func() { ui.Text(c, "Three sources") })
		overlay.Sheet(c, &sheet, overlay.SheetProps{Title: "Files", Side: overlay.Right}, func() { ui.Text(c, "Read four files") })
	}, 600, 420, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Search results"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("native action repeated")
	}
	traceAccess(t, r, ui.RoleDialog, "Results")
	r.Key(0, ui.KeyEscape)
	if !r.Focused("Search results") {
		t.Fatal("native dialog did not restore trace action focus")
	}
	if err := r.Click("Explore files"); err != nil {
		t.Fatal(err)
	}
	traceAccess(t, r, ui.RoleDialog, "Files")
	r.Key(0, ui.KeyEscape)
	if err := r.Click("Open docs"); err != nil {
		t.Fatal(err)
	}
	if urls := r.OpenedURLs(); len(urls) != 1 || urls[0] != "https://example.com/docs" {
		t.Fatalf("actual native URL not opened: %v", urls)
	}
	if err := r.Click("Command output"); err != nil {
		t.Fatal(err)
	}
	if err := r.Click("Native output"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "alpha\nbeta" {
		t.Fatal("composed CodeBlock did not expose actual native selection/copy")
	}
}

func TestTraceRunningTitleInheritsShimmerAndStopsWithoutPendingFrames(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			status := Running
			r := ui.NewTester(func(c *ui.Context) {
				tokens := theme.Light()
				if dark {
					tokens = theme.Dark()
				}
				theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
					TraceLine(c, TraceLineProps{Status: status}, func(child *ui.Context) { TraceTitle(child, "Running native 字形 text") })
				})
			}, 360, 100)
			before := append([]byte(nil), r.Image().Pix...)
			r.Advance(time.Second)
			if bytes.Equal(before, r.Image().Pix) || !r.FramePending() {
				t.Fatal("default title did not inherit source running shimmer")
			}
			traceAccess(t, r, ui.RoleText, "Running native 字形 text")
			status = Complete
			r.Frame()
			r.Advance(250 * time.Millisecond)
			idle := append([]byte(nil), r.Image().Pix...)
			r.Advance(3 * time.Second)
			if !bytes.Equal(idle, r.Image().Pix) || r.FramePending() {
				t.Fatal("completed trace title retained shimmer/timer")
			}
		})
	}
}

func TestTracePendingAppliesSourceOpacityToWholeLine(t *testing.T) {
	render := func(status TraceStatus, reference bool) *ui.Tester {
		return ui.NewTesterWith(func(c *ui.Context) {
			tokens := theme.Light()
			tokens.Background = ui.RGB(20, 40, 60)
			theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
				parent := ui.Column(c)
				if reference {
					parent.Opacity(.6)
				}
				parent.Children(func() {
					TraceLine(c, TraceLineProps{Status: status}, func(child *ui.Context) {
						TraceIcon(child, traceChevronIcon)
						TraceTitle(child, "Pending operation")
						TraceDescription(child, "14s")
					})
				})
			})
		}, 360, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	}
	actual, reference := render(Pending, false), render(Complete, true)
	if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
		t.Fatal("pending source opacity was not0.6 on the entire line")
	}
}

func TestTraceErrorLinkHoverKeepsInheritedDestructiveColor(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Light()
		tokens.Destructive = ui.RGB(220, 10, 20)
		tokens.Foreground = ui.RGB(10, 120, 220)
		theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
			TraceLine(c, TraceLineProps{Status: Error}, func(child *ui.Context) { TraceLink(child, "Error documentation", "https://example.com/error") })
		})
	}, 400, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	link := traceAccess(t, r, ui.RoleLink, "Error documentation")
	redPixels := func() int {
		count := 0
		image := r.Image()
		for y := int(link.Bounds.Y); y < int(link.Bounds.Y+link.Bounds.H-5); y++ {
			for x := int(link.Bounds.X); x < int(link.Bounds.X+link.Bounds.W); x++ {
				p := image.RGBAAt(x, y)
				if int(p.R) > int(p.G)*2 && int(p.R) > int(p.B)*2 {
					count++
				}
			}
		}
		return count
	}
	before := redPixels()
	r.Move(link.Bounds.X+link.Bounds.W/2, link.Bounds.Y+link.Bounds.H/2)
	after := redPixels()
	if before == 0 || after != before {
		t.Fatalf("hover replaced inherited error color: red=%d->%d", before, after)
	}
}

func TestTraceControlledCallbackReversalDoesNotReplayOrExposeContent(t *testing.T) {
	open := false
	calls := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Controlled", Trigger: func(child *ui.Context) { TraceTitle(child, "Controlled") }, OnOpenChange: func(next bool) { calls++; open = false }, Content: func(child *ui.Context) {
			control.Input(child, new(string), control.InputProps{Label: "Must stay closed"})
		}})
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Controlled"); err != nil {
		t.Fatal(err)
	}
	if open || calls != 1 || traceAccess(t, r, ui.RoleButton, "Controlled").Expanded {
		t.Fatal("callback reversal was overwritten or left native expanded state stale")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Must stay closed"); ok {
		t.Fatal("rejected controlled opening exposed content")
	}
	open = true
	r.Frame()
	if calls != 1 {
		t.Fatal("programmatic controlled opening replayed onOpenChange")
	}
}

func TestTraceStructuredCallbackCanReplaceCompatibleMap(t *testing.T) {
	state := TraceState{}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		Trace(c, &state, []TraceStep{{ID: "controlled", Title: "Controlled map", Detail: "Hidden result", OnOpenChange: func(bool) { state.Expanded = map[string]bool{"controlled": false} }}})
	}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Controlled map"); err != nil {
		t.Fatal(err)
	}
	if traceAccess(t, r, ui.RoleButton, "Controlled map").Expanded || r.HasText("Hidden result") {
		t.Fatal("structured callback replaced map but the stale local value won")
	}
}

func TestTraceKeepMountedAutoFocusIsLazyAndCannotReclaimClosedFocus(t *testing.T) {
	open := false
	draft := ""
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Button(c, "Outside")
		TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Autofocus", KeepMounted: true, Trigger: func(child *ui.Context) { TraceTitle(child, "Autofocus") }, Content: func(child *ui.Context) {
			control.Input(child, &draft, control.InputProps{Label: "Auto draft"}).AutoFocus()
		}})
	}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if r.Focused("Auto draft") {
		t.Fatal("closed keep-mounted autofocus stole focus")
	}
	if err := r.Click("Autofocus"); err != nil {
		t.Fatal(err)
	}
	if !r.Focused("Auto draft") {
		t.Fatal("initial open failed to autofocus actual native content")
	}
	open = false
	r.Frame()
	r.Frame()
	if !r.Focused("Autofocus") {
		t.Fatal("retained closed autofocus reclaimed focus")
	}
}

func TestTraceAncestorFontAndSemanticMetadataSurviveCompoundScopes(t *testing.T) {
	render := func(trace bool) *ui.Tester {
		return ui.NewTesterWith(func(c *ui.Context) {
			tokens := theme.Light()
			tokens.MutedForeground = ui.RGB(30, 80, 140)
			native := tokens.Theme()
			native.Font = "monospace"
			ui.Themed(c, native).Fill().Background(tokens.Background).Font("serif").Children(func() {
				if trace {
					TraceLine(c, TraceLineProps{}, func(child *ui.Context) { TraceTitle(child, "iii WWW 014") })
				} else {
					ui.Text(c, "iii WWW 014").FontSize(14).LineHeight(10.0 / 7).TextColor(tokens.MutedForeground).SingleLine()
				}
			})
		}, 400, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	}
	actual, reference := render(true), render(false)
	if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
		t.Fatal("Trace reset inherited native family or lost semantic token metadata")
	}
}

func TestTraceCompoundSlotsPreserveAncestorFont(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Font("monospace").Children(func() {
			TraceLine(c, TraceLineProps{}, func(c *ui.Context) { TraceTitle(c, "iii WWW 014").Label("Trace sample") })
			ui.Text(c, "iii WWW 014").FontSize(14).LineHeight(10.0 / 7).AlignSelf(ui.Start).SingleLine().Label("Inherited sample")
		})
	}, 400, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	actual, _ := r.Find("Trace sample")
	want, _ := r.Find("Inherited sample")
	if math.Abs(float64(actual.W-want.W)) > .1 {
		t.Fatalf("source font inheritance reset to window theme: trace=%v inherited=%v", actual, want)
	}
}

func TestTraceCompoundFamilyAndWeightMatchInheritedNativePixels(t *testing.T) {
	for _, family := range []string{"serif", "monospace"} {
		for _, weight := range []int{400, 700} {
			for _, fluent := range []bool{false, true} {
				for _, slot := range []string{"line", "link", "content", "action"} {
					t.Run(fmt.Sprintf("%s/%d/fluent%v/%s", family, weight, fluent, slot), func(t *testing.T) {
						render := func(actual bool) *ui.Tester {
							return ui.NewTesterWith(func(c *ui.Context) {
								theme.Apply(c)
								tokens := theme.TokensOf(c)
								parent := ui.Column(c).Fill().Font("system-ui").FontWeight(400).Background(tokens.Background)
								if !fluent {
									parent.Font(family).FontWeight(weight)
								}
								parent.Children(func() {
									var element ui.Element
									sourceLine := func(c *ui.Context) ui.Element {
										return ui.Row(c).AlignSelf(ui.Start).Shrink(1).MaxWidthPercent(100).MinWidth(0).Gap(8).AlignItems(ui.Center).
											FontSize(14).LineHeight(10.0 / 7).TextColor(tokens.MutedForeground).Children(func() {
											ui.Text(c, "iii WWW 014").MinWidth(0).Shrink(1).SingleLine()
											ui.Text(c, "+48 −12").FontSize(14).LineHeight(10.0 / 7).TextColor(tokens.MutedForeground.Alpha(.7)).FontFeatures("tnum").Shrink(0).SingleLine()
										})
									}
									lineSlots := func(c *ui.Context) { TraceTitle(c, "iii WWW 014"); TraceDescription(c, "+48 −12") }
									switch slot {
									case "line":
										if actual {
											element = TraceLine(c, TraceLineProps{}, lineSlots)
										} else {
											element = sourceLine(c)
										}
									case "link":
										if actual {
											element = TraceLink(c, "iii WWW 014", "https://example.com")
										} else {
											element = ui.Link(c, "iii WWW 014", "https://example.com").FontSize(14).LineHeight(10.0 / 7).TextColor(tokens.MutedForeground).MinWidth(0).Shrink(1).SingleLine()
										}
									case "content":
										if actual {
											element = TraceCollapsibleContent(c, func(c *ui.Context) { ui.Text(c, "iii WWW 014") })
										} else {
											element = ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Shrink(0).Margin(8, 0, 4, 7).BorderWidth(0, 0, 0, 1).BorderColor(tokens.Border).
												Padding(0, 4, 0, 16).LineHeight(1.625).TextColor(tokens.MutedForeground).Children(func() { ui.Text(c, "iii WWW 014") })
										}
									case "action":
										if actual {
											element = TraceAction(c, TraceLineProps{Label: "Native action"}, lineSlots)
										} else {
											element = ui.ButtonBase(c).Row().AlignSelf(ui.Start).Shrink(1).MaxWidthPercent(100).MinWidth(0).AlignItems(ui.Center).Radius(tokens.RadiusMD()).
												DimDisabled(false).FocusRing(false).Cursor(ui.CursorPointer).Label("Native action").Gap(0).Children(func() { sourceLine(c) })
										}
									}
									if fluent {
										element.Font(family).FontWeight(weight)
									}
								})
							}, 400, 120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
						}
						actual, reference := render(true), render(false)
						if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
							t.Fatal("compound slot reset source native family or weight inheritance")
						}
						role, label := ui.RoleText, "iii WWW 014"
						if slot == "action" {
							role, label = ui.RoleButton, "Native action"
						} else if slot == "link" {
							role = ui.RoleLink
						}
						if _, ok := actual.AccessByRole(role, label); !ok {
							t.Fatal("inherited glyph styling lost native accessibility")
						}
						if actual.FramePending() {
							t.Fatal("static inherited text retained animation work")
						}
					})
				}
			}
		}
	}
}

func TestTracePrimitiveTypographyInheritsNativeSizeAndLineHeight(t *testing.T) {
	for _, slot := range []string{"description", "content", "fluent line"} {
		t.Run(slot, func(t *testing.T) {
			render := func(actual bool) *ui.Tester {
				return ui.NewTesterWith(func(c *ui.Context) {
					theme.Apply(c)
					tokens := theme.TokensOf(c)
					ui.Column(c).Fill().Font("monospace").FontSize(22).FontWeight(700).LineHeight(2).Background(tokens.Background).Children(func() {
						switch slot {
						case "description":
							if actual {
								TraceDescription(c, "003.25 seconds")
							} else {
								ui.Text(c, "003.25 seconds").TextColor(tokens.MutedForeground.Alpha(.7)).FontFeatures("tnum").Shrink(0).SingleLine()
							}
						case "content":
							if actual {
								TraceCollapsibleContent(c, func(c *ui.Context) { ui.Text(c, "Inherited native panel") })
							} else {
								ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Shrink(0).Margin(8, 0, 4, 7).BorderWidth(0, 0, 0, 1).BorderColor(tokens.Border).
									Padding(0, 4, 0, 16).LineHeight(1.625).TextColor(tokens.MutedForeground).Children(func() { ui.Text(c, "Inherited native panel") })
							}
						case "fluent line":
							if actual {
								TraceLine(c, TraceLineProps{}, func(c *ui.Context) { TraceTitle(c, "Glyphs"); TraceDescription(c, "14s") }).FontSize(22)
							} else {
								ui.Row(c).AlignSelf(ui.Start).Shrink(1).MaxWidthPercent(100).MinWidth(0).Gap(8).AlignItems(ui.Center).FontSize(22).LineHeight(10.0 / 7).TextColor(tokens.MutedForeground).Children(func() {
									ui.Text(c, "Glyphs").MinWidth(0).Shrink(1).SingleLine()
									ui.Text(c, "14s").TextColor(tokens.MutedForeground.Alpha(.7)).FontFeatures("tnum").Shrink(0).SingleLine()
								})
							}
						}
					})
				}, 500, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			}
			actual, reference := render(true), render(false)
			if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
				t.Fatal("primitive reset inherited22DIP size, native line height or bold glyphs")
			}
			if actual.FramePending() {
				t.Fatal("inherited static primitive retained animation work")
			}
		})
	}
}

func TestTraceDisabledNativeTriggerAndActionIgnorePointerAndKeys(t *testing.T) {
	open := false
	calls := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Disabled disclosure", Disabled: true, Trigger: func(child *ui.Context) { TraceTitle(child, "Disabled disclosure") }, Content: func(child *ui.Context) { ui.Text(child, "Hidden") }, OnOpenChange: func(bool) { calls++ }})
		if TraceAction(c, TraceLineProps{Label: "Disabled action", Disabled: true}, func(child *ui.Context) { TraceTitle(child, "Disabled action") }).Clicked() {
			calls++
		}
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	for _, name := range []string{"Disabled disclosure", "Disabled action"} {
		a := traceAccess(t, r, ui.RoleButton, name)
		if !a.Disabled {
			t.Fatal("native disabled state missing")
		}
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Key(0, ui.KeyEnter)
		r.Key(0, ui.KeySpace)
	}
	if open || calls != 0 || r.HasText("Hidden") {
		t.Fatal("disabled native controls accepted input")
	}
}

func TestTraceTitleExplicitStatusAndScopeRestorationDoNotShimmer(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TraceLine(c, TraceLineProps{Status: Running}, func(child *ui.Context) { TraceTitle(child, "Explicit normal", TraceTitleProps{Status: Complete}) })
		TraceTitle(c, "Outside line")
	}, 400, 180)
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(3 * time.Second)
	if !bytes.Equal(before, r.Image().Pix) || r.FramePending() {
		t.Fatal("explicit status or numeric scope restoration leaked running shimmer")
	}
}

func TestTraceEmptySlotsDoNotInjectTextLineHeight(t *testing.T) {
	var h float32
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TraceLine(c, TraceLineProps{}, func(child *ui.Context) { TraceTitle(child, ""); TraceDescription(child, "") }).Draw(func(_ *ui.Painter, bounds ui.Rect) { h = bounds.H })
	}, 400, 180)
	if h != 0 {
		t.Fatalf("empty source spans injected line height: %g", h)
	}
	if r.FramePending() {
		t.Fatal("empty slots left animation work")
	}
}

func TestTraceContentPhysicalBorderAndPaddingPixelsInRTL(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			open := true
			tokens := theme.Light()
			tokens.Border = ui.RGB(220, 20, 70)
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
					TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Panel", Trigger: func(child *ui.Context) { TraceTitle(child, "Panel") }, Content: func(child *ui.Context) { ui.Box(child).FillWidth().Height(30).Background(ui.RGB(20, 80, 180)) }})
				})
			}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			r.SetDirection(direction)
			trigger := traceAccess(t, r, ui.RoleButton, "Panel")
			y := int(trigger.Bounds.Y + trigger.Bounds.H + 10)
			image := r.Image()
			border := image.RGBAAt(7, y)
			inside := image.RGBAAt(24, y)
			right := image.RGBAAt(399, y)
			if border.R != 220 || border.G != 20 || border.B != 70 || inside.B != 180 || right.R != 255 || right.G != 255 || right.B != 255 {
				t.Fatalf("physical source content spacing lost dir%v: border=%v inside=%v right=%v", direction, border, inside, right)
			}
		})
	}
}

func TestTraceNativeActionTitleHoverInterpolatesSourceColor(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		tokens := theme.Light()
		tokens.MutedForeground = ui.RGB(30, 70, 120)
		tokens.Foreground = ui.RGB(220, 80, 20)
		theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
			TraceAction(c, TraceLineProps{Label: "Hover action"}, func(child *ui.Context) { TraceTitle(child, "MMMM native glyphs") })
		})
	}, 400, 160)
	before := append([]byte(nil), r.Image().Pix...)
	action := traceAccess(t, r, ui.RoleButton, "Hover action")
	r.Move(action.Bounds.X+action.Bounds.W/2, action.Bounds.Y+action.Bounds.H/2)
	r.Advance(75 * time.Millisecond)
	middle := append([]byte(nil), r.Image().Pix...)
	r.Advance(100 * time.Millisecond)
	after := r.Image().Pix
	if bytes.Equal(middle, before) || bytes.Equal(middle, after) {
		t.Fatal("source hover transition did not interpolate actual title glyphs")
	}
}

func TestTraceDynamicOffscreenContentUsesFreshNaturalHeight(t *testing.T) {
	open := true
	height := float32(40)
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).FillWidth().Children(func() {
			ui.Box(c).Height(500).Shrink(0)
			TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Offscreen", KeepMounted: true, Trigger: func(child *ui.Context) { TraceTitle(child, "Offscreen") }, Content: func(child *ui.Context) { ui.Box(child).Height(height).Shrink(0) }}).Role(ui.RoleGroup).Label("Offscreen trace")
		})
	}, 400, 240)
	height = 240
	r.Frame()
	full := traceAccess(t, r, ui.RoleGroup, "Offscreen trace").Bounds.H
	if math.Abs(float64(full-272)) > .1 {
		t.Fatalf("offscreen open natural height=%g want272", full)
	}
	open = false
	r.Frame()
	r.Advance(100 * time.Millisecond)
	middle := traceAccess(t, r, ui.RoleGroup, "Offscreen trace").Bounds.H
	if middle <= 20 || middle >= full {
		t.Fatalf("offscreen closing did not use newly measured height: %g", middle)
	}
	open = true
	r.Frame()
	height = 320
	r.Frame()
	r.Advance(250 * time.Millisecond)
	if got := traceAccess(t, r, ui.RoleGroup, "Offscreen trace").Bounds.H; math.Abs(float64(got-352)) > .1 {
		t.Fatalf("dynamic offscreen reversal ended at stale height: %g", got)
	}
}

func TestTraceLateSecondDisclosureDoesNotPingPongIdleFrames(t *testing.T) {
	first, second := true, true
	late := false
	builds := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		builds++
		theme.Apply(c)
		services := c.Services()
		TraceGroup(c, func(group *ui.Context) {
			TraceCollapsible(group.Key("first"), &first, TraceCollapsibleProps{Label: "First", Trigger: func(child *ui.Context) { TraceTitle(child, "First") }, Content: func(child *ui.Context) { ui.Box(child).Height(40) }}).Draw(func(_ *ui.Painter, _ ui.Rect) {
				if !late {
					late = true
					services.Invalidate()
				}
			})
			if late {
				TraceCollapsible(group.Key("second"), &second, TraceCollapsibleProps{Label: "Second", Trigger: func(child *ui.Context) { TraceTitle(child, "Second") }, Content: func(child *ui.Context) { ui.Box(child).Height(70) }})
			}
		})
	}, 400, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if builds >= 20 || r.FramePending() {
		t.Fatalf("independent disclosures ping-ponged correction frames: builds=%d pending=%v", builds, r.FramePending())
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("static multi-Trace view did not remain idle")
	}
}

func TestTraceEmptyContentMarginsStillAnimate(t *testing.T) {
	open := false
	var height float32
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Empty details", Trigger: func(child *ui.Context) { TraceTitle(child, "Empty details") }, Content: func(*ui.Context) {}}).Draw(func(_ *ui.Painter, b ui.Rect) { height = b.H })
	}, 400, 180)
	if err := r.Click("Empty details"); err != nil {
		t.Fatal(err)
	}
	r.Advance(100 * time.Millisecond)
	if height <= 20 || height >= 32 {
		t.Fatalf("empty source detail margins did not animate: %g", height)
	}
	r.Advance(150 * time.Millisecond)
	if height != 32 {
		t.Fatalf("empty source detail natural margin height=%g want32", height)
	}
}

func TestTraceNestedDuplicateIDsKeepIndependentOpenState(t *testing.T) {
	state := TraceState{Expanded: map[string]bool{"first": true, "second": true}}
	steps := []TraceStep{
		{ID: "first", Title: "First branch", Children: []TraceStep{{ID: "shared", Title: "First child", Detail: "First details"}}},
		{ID: "second", Title: "Second branch", Children: []TraceStep{{ID: "shared", Title: "Second child", Detail: "Second details"}}},
	}
	r := ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); Trace(c, &state, steps) }, 500, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("First child"); err != nil {
		t.Fatal(err)
	}
	if !r.HasText("First details") || r.HasText("Second details") {
		t.Fatal("same child ID linked independent branch disclosures")
	}
}

func TestTraceScaledDisclosureUsesLayoutHeightAndSettles(t *testing.T) {
	open := false
	var actual ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TraceCollapsible(c, &open, TraceCollapsibleProps{Label: "Scaled details", KeepMounted: true,
			Trigger: func(c *ui.Context) { TraceTitle(c, "Scaled details") },
			Content: func(c *ui.Context) { ui.Box(c).Height(80).Shrink(0) },
		}).Scale(.75, .75).Draw(func(_ *ui.Painter, bounds ui.Rect) { actual = bounds })
	}, 400, 250)
	open = true
	r.Frame()
	r.Advance(100 * time.Millisecond)
	if actual.H <= 20 || actual.H >= 112 {
		t.Fatalf("scaled opening did not animate true layout height: %v", actual)
	}
	r.Advance(150 * time.Millisecond)
	if math.Abs(float64(actual.H-112)) > .1 || r.FramePending() {
		t.Fatalf("scaled open disclosure did not settle: height=%v frame=%v", actual.H, r.FramePending())
	}
	open = false
	r.Frame()
	r.Advance(250 * time.Millisecond)
	if math.Abs(float64(actual.H-20)) > .1 || r.FramePending() {
		t.Fatalf("scaled close did not settle: height=%v frame=%v", actual.H, r.FramePending())
	}
}
