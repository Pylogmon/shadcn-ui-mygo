package desktop

import (
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type comboFlowSurface struct {
	spec                Spec
	view                func(*ui.Context)
	closed              func()
	dismiss             func() bool
	tester              *ui.Tester
	invalidates, closes int
}

func (s *comboFlowSurface) SetAnchor(r ui.Rect)              { s.spec.Anchor = r }
func (s *comboFlowSurface) Invalidate()                      { s.invalidates++ }
func (s *comboFlowSurface) Close()                           { s.closes++ }
func (s *comboFlowSurface) SetDismissHandler(fn func() bool) { s.dismiss = fn }
func (s *comboFlowSurface) outside() bool {
	if s.dismiss != nil && s.dismiss() {
		return false
	}
	s.closed()
	return true
}

type comboFlowBackend struct {
	surfaces    []*comboFlowSurface
	invalidates int
}

func (b *comboFlowBackend) Open(spec Spec, view func(*ui.Context), closed func()) (Surface, error) {
	s := &comboFlowSurface{spec: spec, view: view, closed: closed}
	b.surfaces = append(b.surfaces, s)
	return s, nil
}
func (b *comboFlowBackend) InvalidateParent() { b.invalidates++ }

func comboFlowAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing %v %q", role, name)
	}
	return a
}

// One real two-context flow exercises typed single/multiple application values,
// current-window filters/handlers, cancellation, editor lifetime and both themes.
func TestHostedComboboxTypedIndependentWindowWorkflow(t *testing.T) {
	type country struct {
		Code, Label string
		Disabled    bool
	}
	items := []country{{"go", "Go", false}, {"gio", "Gio", false}, {"kr", "South Korea", false}, {"blocked", "Unavailable", true}, {"tr", "Istanbul", false}}
	state := control.ComboboxState[country]{}
	b := &comboFlowBackend{}
	h := NewHostWithBackend(b)
	native, multiple, rejectValue, rejectQuery, rejectClose := false, false, false, false, false
	reduced := true
	locale, customFilter := "en", true
	version, seenVersion, changes, changedReturns, opens, closes := 1, 0, 0, 0, 0, 0
	var queryEvents []string
	var completions []bool
	var search, afterHandle ui.Handle
	var anchor ui.Element
	var selection [2]int
	var providerFont string
	var popupErr error
	var parentHasChildEditor bool
	view := func(current *ui.Context) control.ComboboxRootProps[country] {
		stamp := version
		providerFont = current.Theme().Font
		var filter func(country, string) bool
		if customFilter {
			filter = func(v country, q string) bool { return strings.Contains(v.Code, strings.ToLower(q)) }
		}
		return control.ComboboxRootProps[country]{Items: items, Multiple: multiple, QueryControlled: true, AutoHighlight: control.ComboboxHighlightOnInput,
			Adapter: control.ComboboxAdapter[country]{Key: func(v country) string { return v.Code }, Label: func(v country) string { return v.Label }, Disabled: func(v country) bool { return v.Disabled }},
			Filter:  filter, Locale: locale,
			OnQueryChange: func(q string, event *control.ComboboxChange) {
				queryEvents = append(queryEvents, q)
				if rejectQuery {
					event.Cancel()
				}
			},
			OnValueChange: func(v control.ComboboxSelection[country], event *control.ComboboxChange) {
				seenVersion = stamp
				changes++
				if rejectValue {
					event.Cancel()
				}
			},
			OnValuesChange: func(values []country, event *control.ComboboxChange) {
				seenVersion = stamp
				changes++
				if rejectValue {
					event.Cancel()
				}
			},
			OnOpenChange: func(open bool, event *control.ComboboxChange) {
				if open {
					opens++
				} else {
					closes++
					if rejectClose {
						event.Cancel()
					}
				}
			},
			OnOpenChangeComplete: func(open bool) { completions = append(completions, open) },
		}
	}
	after := ""
	parent := ui.NewTester(func(c *ui.Context) {
		if reduced {
			c.DisableAnimations()
		}
		fonts := theme.FontRoles{Body: "monospace", Heading: "Georgia"}
		theme.FontScope(c, fonts).Fill().Padding(16).Gap(8).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			p := ComboboxProps{Label: "Country", Placeholder: "Choose country", SearchHandle: &search, ShowClear: true, OnTab: func(bool) { afterHandle.Focus(c) }}
			if native {
				anchor, popupErr = NativeCombobox(c, h, "country", &state, p, view)
			} else {
				anchor, popupErr = Combobox(c, h, "country", &state, p, view)
			}
			if anchor.Changed() {
				changedReturns++
			}
			control.Input(c, &after, control.InputProps{Label: "After", Handle: &afterHandle})
			parentHasChildEditor = c.Resolve(search).Valid()
		})
	}, 260, 160)
	snapshot := func(name string, current *ui.Tester) {
		t.Helper()
		if dir := os.Getenv("SHADCN_COMBOBOX_SNAPSHOTS"); dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			f, err := os.Create(filepath.Join(dir, name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(f, current.Image()); err != nil {
				f.Close()
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	open := func() *comboFlowSurface {
		t.Helper()
		if err := parent.Click("Country"); err != nil {
			t.Fatal(err)
		}
		parent.Frame()
		if popupErr != nil || !state.Open || len(b.surfaces) == 0 {
			t.Fatalf("open=%v err=%v surfaces=%d", state.Open, popupErr, len(b.surfaces))
		}
		s := b.surfaces[len(b.surfaces)-1]
		platform := *ui.LightTheme()
		platform.Font = "Georgia"
		platform.FontSize = 19
		platform.Spacing = 5
		s.tester = ui.NewTester(func(c *ui.Context) {
			c.SetTheme(&platform)
			s.view(c)
			if editor := c.Resolve(search); editor.Valid() {
				selection[0], selection[1] = editor.TextSelection()
			}
		}, s.spec.Width, s.spec.MaxHeight)
		return s
	}
	s := open()
	child := s.tester
	if opens != 1 || s.spec.SystemPopover || !s.spec.Focusable || s.spec.MaxHeight <= 160 || parent.HasText("South Korea") || providerFont != "monospace" || !child.Focused("Country search") {
		t.Fatalf("not an independent source popup: opens%d spec%+v font%q focus%v", opens, s.spec, providerFont, child.AccessAll())
	}
	snapshot("styled-country", child)
	parent.Frame()
	if parentHasChildEditor {
		t.Fatal("the child editor Handle resolved in the parent Context")
	}
	// IME preedit never executes the custom filter or chooses an option.
	child.Compose("kr", 2)
	if state.Value.Valid || len(queryEvents) != 0 || !child.HasText("Go") {
		t.Fatal("preedit was published as a query or selection")
	}
	child.Key(0, ui.KeyEnter)
	if state.Value.Valid {
		t.Fatal("candidate-confirming Enter also selected an option")
	}
	if state.Query != "kr" || !child.HasText("South Korea") || child.HasText("Gio") {
		t.Fatal("committed query did not use the current custom code filter in its own child frame")
	}
	rejectQuery = true
	child.Key(ui.Cmd, ui.KeyA)
	child.Type("deny")
	if state.Query != "kr" || comboFlowAccess(t, child, ui.RoleComboBox, "Country search").Value != "kr" {
		t.Fatal("query cancellation lost the current editor/controlled value")
	}
	rejectQuery = false
	version = 2
	child.Frame()
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	parent.Frame()
	if !state.Value.Valid || state.Value.Value.Code != "kr" || state.Open || seenVersion != 2 || changes != 1 || changedReturns != 1 || !parent.Focused("Country") || state.Query != "" {
		t.Fatalf("typed commit/current callback/focus: state%+v version%d changes%d returns%d focused%v", state, seenVersion, changes, changedReturns, parent.AccessAll())
	}
	// Metadata changes refresh the same native child. Outside cancellation must
	// keep its actual editor instance and selection, not close and rebuild it.
	s = open()
	child = s.tester
	child.Type("g")
	child.Key(ui.Shift, ui.KeyLeft)
	before := selection
	rejectClose = true
	if s.outside() {
		t.Fatal("standard dismiss interceptor allowed destruction before Cancel")
	}
	parent.Frame()
	child.Frame()
	if !state.Open || s.closes != 0 || b.surfaces[len(b.surfaces)-1] != s || selection != before {
		t.Fatal("canceled outside dismissal destroyed the child or its native selection")
	}
	items = append(items, country{"gr", "Greece", false})
	parent.Frame()
	if s.invalidates == 0 {
		t.Fatal("external collection change did not invalidate the existing child")
	}
	child.Frame()
	if !child.HasText("Greece") {
		t.Fatal("existing child retained the obsolete collection")
	}
	rejectClose = false
	s.outside()
	parent.Frame()
	if state.Open || s.closes != 1 || !parent.Focused("Country") {
		t.Fatal("accepted outside dismissal did not use the host close/focus lifecycle")
	}
	// Native is a plain platform-themed Flyout, not a static OS menu. Multiple
	// selection reuses the same typed state and current cancellation handlers.
	native, multiple = true, true
	locale, customFilter = "tr", false
	parent.Frame()
	s = open()
	child = s.tester
	if providerFont != "Georgia" || s.spec.SystemPopover || !s.spec.Focusable {
		t.Fatal("Native did not retain the child's own platform theme in a plain editable Flyout")
	}
	nativeList := comboFlowAccess(t, child, ui.RoleListBox, "Country options")
	if !nativeList.Multiselectable {
		t.Fatal("native child lost multiple-list semantics")
	}
	if nativeList.Bounds.Y+nativeList.Bounds.H > float32(s.spec.MaxHeight) {
		t.Fatalf("platform font overflowed the independent scroll viewport: %+v", nativeList.Bounds)
	}
	snapshot("native-country", child)
	child.Type("ı")
	if !child.HasText("Istanbul") {
		t.Fatal("native search lost the provider's Turkish locale")
	}
	locale = "en"
	parent.Frame()
	child.Frame()
	if child.HasText("Istanbul") || !child.HasText("No items found.") {
		t.Fatal("external locale change did not refresh the same native child")
	}
	locale, customFilter = "tr", true
	parent.Frame()
	child.Frame()
	child.Key(ui.Cmd, ui.KeyA)
	child.Type("gio")
	rejectValue = true
	child.Key(0, ui.KeyEnter)
	if len(state.Values) != 0 {
		t.Fatal("canceled typed multiple change mutated application values")
	}
	rejectValue = false
	child.Type("gio")
	child.Key(0, ui.KeyEnter)
	if !slices.ContainsFunc(state.Values, func(v country) bool { return v.Code == "gio" }) || !state.Open {
		t.Fatalf("native multi selection was not a real editable typed collection: %+v", state)
	}
	parent.Frame()
	child.Key(0, ui.KeyTab)
	parent.Frame()
	parent.Frame()
	if state.Open || !parent.Focused("After") {
		t.Fatalf("cross-window Tab did not use the explicit adjacent Handle: open%v AX%v", state.Open, parent.AccessAll())
	}
	parent.Frame()
	if popupErr != nil || changedReturns != 2 || closes < 3 {
		t.Fatalf("events were duplicated or missing: returns%d close%d err%v", changedReturns, closes, popupErr)
	}
	// The same workflow also accepts an outside close with motion enabled.
	// Repeated gestures during its exit cannot replace Host completion or emit
	// another close transaction when the retained surface is finally released.
	native, multiple, reduced = false, false, false
	parent.Frame()
	s = open()
	child = s.tester
	child.Advance(150 * time.Millisecond)
	closeBefore, completeBefore := closes, len(completions)
	child.Type("g")
	rejectQuery = true
	queryBefore := len(queryEvents)
	if s.outside() {
		t.Fatal("outside close escaped the current-handler transaction")
	}
	parent.Frame()
	if state.Open || !h.IsPresent("country") || !h.entries["country"].closing || s.closes != 0 {
		t.Fatal("accepted close skipped the Host's retained exit")
	}
	child.Frame()
	if s.outside() {
		t.Fatal("repeated outside gesture destroyed an exiting surface")
	}
	child.Advance(50 * time.Millisecond)
	if s.closes != 0 {
		t.Fatal("surface closed before the 100ms exit completed")
	}
	child.Advance(100 * time.Millisecond)
	parent.Frame()
	parent.Frame()
	if s.closes != 1 || h.IsPresent("country") || !parent.Focused("Country") || closes != closeBefore+1 || len(completions) != completeBefore+1 || completions[len(completions)-1] {
		t.Fatalf("exit completion was replaced/duplicated: closes%d/%d completion%v surface%d present%v", closes, closeBefore, completions, s.closes, h.IsPresent("country"))
	}
	if state.Query != "g" || len(queryEvents) != queryBefore+1 || queryEvents[len(queryEvents)-1] != "" {
		t.Fatal("host close cleanup lost cancellation or published the same query cleanup twice")
	}
}
