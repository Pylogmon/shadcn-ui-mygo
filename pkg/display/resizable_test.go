package display

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestResizableLegacyConstrainsKeyboardAndCapturedDrag(t *testing.T) {
	size := float32(100)
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Resizable(c, &size, ResizableProps{Min: 80, Max: 120}, func() { ui.Text(c, "Sidebar") }, func() { ui.Text(c, "Editor") }).Fill()
	}, 400, 200)
	r.Key(0, ui.KeyTab)
	for range 8 {
		r.Key(0, ui.KeyRight)
	}
	if size > 120 {
		t.Fatalf("keyboard escaped Max: %v", size)
	}
	var handle ui.AccessInfo
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleSplitter {
			handle = a
			break
		}
	}
	if handle.Bounds.H == 0 {
		t.Fatal("splitter is not accessible")
	}
	x, y := handle.Bounds.X+handle.Bounds.W/2, handle.Bounds.Y+handle.Bounds.H/2
	r.Press(x, y)
	r.Move(x-200, y)
	if size < 80 {
		t.Fatalf("captured drag escaped Min before release: %v", size)
	}
	r.Release(x-200, y)
}

func TestResizableLegacyNormalizesNonFiniteControlledSize(t *testing.T) {
	for _, initial := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		size := initial
		r := ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			Resizable(c, &size, ResizableProps{Min: 80, Max: 120}, func() {}, func() {}).Fill()
		}, 400, 240)
		if !panelFinite(float64(size)) || size < 80 || size > 120 {
			t.Fatalf("nonfinite size was not normalized: %v", size)
		}
		for _, a := range r.AccessAll() {
			if a.Role == ui.RoleSplitter && (!panelFinite(a.Now) || !panelFinite(a.Min) || !panelFinite(a.Max)) {
				t.Fatalf("nonfinite native splitter metadata: %+v", a)
			}
		}
	}
}

func panelTestAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q: %+v", role, label, r.AccessAll())
	}
	return a
}

func panelTester(layout *ResizableLayout, p *ResizablePanelGroupProps, onChange func(), reduce bool) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		if ResizablePanelGroup(c, layout, *p).Changed() && onChange != nil {
			onChange()
		}
	}, 400, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduce}})
}

func TestResizablePanelGroupPointerConstraintsPushAndCancellation(t *testing.T) {
	layout := ResizableLayout{"a": 25, "b": 50, "c": 25}
	callbacks, changes := 0, 0
	p := ResizablePanelGroupProps{Panels: []ResizablePanelProps{
		{ID: "a", Label: "Files", MinSize: 20, MaxSize: 45},
		{ID: "b", Label: "Editor", MinSize: 30, MaxSize: 60},
		{ID: "c", Label: "Inspector", MinSize: 10, MaxSize: 40},
	}, OnLayout: func(got ResizableLayout) { callbacks++; got["a"] = -999 }}
	r := panelTester(&layout, &p, func() { changes++ }, false)
	a := panelTestAccess(t, r, ui.RoleGroup, "Files")
	if math.Abs(float64(a.Bounds.W)-398*.25) > 1 {
		t.Fatalf("initial pane width=%v", a.Bounds.W)
	}
	drag := func(label string, dx float32, cancel bool) {
		h := panelTestAccess(t, r, ui.RoleSplitter, label)
		x, y := h.Bounds.X+h.Bounds.W/2, h.Bounds.Y+h.Bounds.H/2
		r.Press(x, y)
		r.Move(x+dx, y)
		if cancel {
			r.CancelPress()
		} else {
			r.Release(x+dx, y)
		}
		r.Frame()
	}
	drag("Resize Files", 220, false)
	if math.Abs(layout["a"]-45) > 1e-5 || math.Abs(layout["b"]-30) > 1e-5 || math.Abs(layout["c"]-25) > 1e-5 {
		t.Fatalf("constrained drag=%v", layout)
	}
	drag("Resize Editor", 160, false)
	if math.Abs(layout["b"]-45) > 1e-5 || math.Abs(layout["c"]-10) > 1e-5 {
		t.Fatalf("second boundary=%v", layout)
	}
	drag("Resize Files", -160, false)
	if math.Abs(layout["a"]-20) > 1e-5 || math.Abs(layout["b"]-60) > 1e-5 || math.Abs(layout["c"]-20) > 1e-5 {
		t.Fatalf("resize did not push beyond max neighbour=%v", layout)
	}
	before := panelClone(layout)
	drag("Resize Files", 60, true)
	for id, v := range before {
		if math.Abs(layout[id]-v) > 1e-5 {
			t.Fatalf("cancel did not restore press layout: %v -> %v", before, layout)
		}
	}
	if callbacks != changes || callbacks < 4 {
		t.Fatalf("callbacks=%d changes=%d", callbacks, changes)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("resizable group left idle frames")
	}
}

func TestResizablePanelGroupKeyboardRTLVerticalAndAccess(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(fmt.Sprintf("vertical=%v", vertical), func(t *testing.T) {
			layout := ResizableLayout{"one": 40, "two": 60}
			p := ResizablePanelGroupProps{Vertical: vertical, Step: 5, Panels: []ResizablePanelProps{{ID: "one", Label: "First", MinSize: 20, MaxSize: 70}, {ID: "two", Label: "Second", MinSize: 30, MaxSize: 80}}}
			r := panelTester(&layout, &p, nil, false)
			r.SetDirection(ui.RightToLeft)
			h := panelTestAccess(t, r, ui.RoleSplitter, "Resize First")
			if h.Min != 20 || h.Max != 70 || math.Abs(h.Now-40) > 1e-5 || h.Step != 5 || h.Vertical == vertical {
				t.Fatalf("splitter AX=%+v", h)
			}
			r.Key(0, ui.KeyTab)
			key := ui.KeyRight
			if vertical {
				key = ui.KeyDown
			}
			r.Key(0, key)
			want := 35.0
			if vertical {
				want = 45
			}
			if math.Abs(layout["one"]-want) > 1e-5 {
				t.Fatalf("direction key=%v", layout)
			}
			r.Key(0, ui.KeyEnd)
			if math.Abs(layout["one"]-70) > 1e-5 {
				t.Fatalf("End=%v", layout)
			}
			r.Key(0, ui.KeyHome)
			if math.Abs(layout["one"]-20) > 1e-5 {
				t.Fatalf("Home=%v", layout)
			}
			p.Disabled = true
			r.Frame()
			r.Key(0, ui.KeyEnd)
			if math.Abs(layout["one"]-20) > 1e-5 {
				t.Fatal("disabled splitter moved")
			}
		})
	}
}

func TestResizablePanelCollapsePreservesContentAndRestoresWithKeyboardAndDrag(t *testing.T) {
	layout := ResizableLayout{"nav": 30, "body": 70}
	created := 0
	text := "kept"
	p := ResizablePanelGroupProps{Panels: []ResizablePanelProps{
		{ID: "nav", Label: "Navigation", MinSize: 20, MaxSize: 60, DefaultSize: 30, Collapsible: true, Content: func(c *ui.Context) {
			e := ui.Box(c)
			ui.Local(e, "created", func() bool { created++; return true })
			e.Children(func() { ui.TextInput(c, &text).Label("Navigation filter") })
		}},
		{ID: "body", Label: "Body", MinSize: 40, MaxSize: 100},
	}}
	r := panelTester(&layout, &p, nil, false)
	h := panelTestAccess(t, r, ui.RoleSplitter, "Resize Navigation")
	r.ClickAt(h.Bounds.X+4, h.Bounds.Y+50)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if layout["nav"] != 0 || layout["body"] != 100 {
		t.Fatalf("Enter collapse=%v", layout)
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Navigation filter"); ok {
		t.Fatal("collapsed content remained accessible")
	}
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if math.Abs(layout["nav"]-30) > 1e-5 || created != 1 {
		t.Fatalf("restore lost layout/content state: %v created=%d", layout, created)
	}
	h = panelTestAccess(t, r, ui.RoleSplitter, "Resize Navigation")
	x, y := h.Bounds.X+4, h.Bounds.Y+50
	r.Press(x, y)
	r.Move(5, y)
	r.Release(5, y)
	r.Frame()
	if layout["nav"] != 0 {
		t.Fatalf("drag did not snap collapse: %v", layout)
	}
	h = panelTestAccess(t, r, ui.RoleSplitter, "Resize Navigation")
	x = h.Bounds.X + 4
	r.Press(x, y)
	r.Move(x+70, y)
	r.Release(x+70, y)
	r.Frame()
	if layout["nav"] < 20 || layout["nav"] > 60 {
		t.Fatalf("drag did not reopen above minimum: %v", layout)
	}
	if created != 1 || text != "kept" {
		t.Fatal("collapse remounted or cleared panel content")
	}
}

func TestResizablePanelDefaultsPersistenceDynamicIdentityAndExternalLoad(t *testing.T) {
	var layout ResizableLayout
	changes, callbacks := 0, 0
	p := ResizablePanelGroupProps{Panels: []ResizablePanelProps{{ID: "a", Label: "A", DefaultSize: 25}, {ID: "b", Label: "B", DefaultSize: 50}, {ID: "c", Label: "C", DefaultSize: 25}}, OnLayout: func(ResizableLayout) { callbacks++ }}
	r := panelTester(&layout, &p, func() { changes++ }, true)
	if math.Abs(layout["b"]-50) > 1e-5 || callbacks != 0 || changes != 0 {
		t.Fatalf("initial defaults published a user event: %v %d/%d", layout, callbacks, changes)
	}
	h := panelTestAccess(t, r, ui.RoleSplitter, "Resize B")
	r.ClickAt(h.Bounds.X+4, h.Bounds.Y+30)
	r.Key(0, ui.KeyLeft)
	data, err := json.Marshal(layout)
	if err != nil {
		t.Fatal(err)
	}
	var loaded ResizableLayout
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	layout = loaded
	r.Frame()
	if !panelTestAccess(t, r, ui.RoleSplitter, "Resize B").Focused {
		t.Fatal("loading layout replaced the stable splitter")
	}
	beforeCallbacks := callbacks
	layout = ResizableLayout{"a": 10, "b": 30, "c": 60}
	r.Frame()
	if callbacks != beforeCallbacks || changes != callbacks {
		t.Fatal("external load reported a user change")
	}
	if math.Abs(float64(panelTestAccess(t, r, ui.RoleGroup, "C").Bounds.W)-398*.6) > 1 {
		t.Fatal("loaded percentages not used for layout")
	}
	p.Panels = []ResizablePanelProps{p.Panels[1], p.Panels[0], p.Panels[2]}
	r.Frame()
	if !panelTestAccess(t, r, ui.RoleSplitter, "Resize B").Focused {
		t.Fatal("reordering lost keyed boundary focus")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("reduced-motion layout left pending frames")
	}
}

func TestResizablePanelInvalidConstraintsStayFiniteAndDoNotOverwriteLayout(t *testing.T) {
	for _, panels := range [][]ResizablePanelProps{
		{{ID: "a", MinSize: 60}, {ID: "b", MinSize: 60}},
		{{ID: "a", MaxSize: 20}, {ID: "b", MaxSize: 30}},
		{{ID: "a"}, {ID: "a"}},
		{{ID: ""}, {ID: "b"}},
		{{ID: "a", MinSize: math.NaN()}, {ID: "b"}},
	} {
		layout := ResizableLayout{"a": 30, "b": 70}
		before := panelClone(layout)
		p := ResizablePanelGroupProps{Label: "Invalid group", Panels: panels}
		r := panelTester(&layout, &p, nil, false)
		if !panelTestAccess(t, r, ui.RoleGroup, "Invalid group").Invalid {
			t.Fatal("impossible constraints not exposed as invalid")
		}
		for _, a := range r.AccessAll() {
			if a.Role == ui.RoleSplitter && (!a.Disabled || !panelFinite(a.Min) || !panelFinite(a.Max) || !panelFinite(a.Now)) {
				t.Fatalf("invalid splitter metadata: %+v", a)
			}
		}
		r.Key(0, ui.KeyTab)
		r.Key(0, ui.KeyRight)
		for key, v := range before {
			if layout[key] != v {
				t.Fatalf("invalid layout overwritten: %v", layout)
			}
		}
	}
	layout := ResizableLayout{"a": math.NaN(), "b": math.Inf(1)}
	p := ResizablePanelGroupProps{Panels: []ResizablePanelProps{{ID: "a"}, {ID: "b"}}}
	panelTester(&layout, &p, nil, false)
	if !panelFinite(layout["a"]) || !panelFinite(layout["b"]) || math.Abs(layout["a"]+layout["b"]-100) > 1e-5 {
		t.Fatalf("nonfinite external values=%v", layout)
	}
}

func TestResizableNestedGroupsHaveIndependentLayoutsAndDisabledKeepsContentUsable(t *testing.T) {
	outer := ResizableLayout{"left": 50, "right": 50}
	inner := ResizableLayout{"top": 25, "bottom": 75}
	value := false
	ip := ResizablePanelGroupProps{Vertical: true, Panels: []ResizablePanelProps{{ID: "top", Label: "Top"}, {ID: "bottom", Label: "Bottom"}}}
	op := ResizablePanelGroupProps{Disabled: true, Panels: []ResizablePanelProps{{ID: "left", Label: "Left", Content: func(c *ui.Context) { ui.Checkbox(c, &value, "Keep editing") }}, {ID: "right", Label: "Right", Content: func(c *ui.Context) { ResizablePanelGroup(c, &inner, ip) }}}}
	r := panelTester(&outer, &op, nil, false)
	if err := r.Click("Keep editing"); err != nil {
		t.Fatal(err)
	}
	if !value {
		t.Fatal("disabling resize disabled panel content")
	}
	h := panelTestAccess(t, r, ui.RoleSplitter, "Resize Top")
	r.ClickAt(h.Bounds.X+h.Bounds.W/2, h.Bounds.Y+4)
	r.Key(0, ui.KeyDown)
	r.Frame()
	if inner["top"] <= 25 || !slices.Equal([]float64{outer["left"], outer["right"]}, []float64{50, 50}) {
		t.Fatalf("nested layouts coupled: outer=%v inner=%v", outer, inner)
	}
}

func TestResizableNativeAccessibilityKeyBridgeAndControlledFocusRecovery(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(fmt.Sprintf("vertical=%v", vertical), func(t *testing.T) {
			layout := ResizableLayout{"a": 40, "b": 60}
			text := "value"
			p := ResizablePanelGroupProps{Vertical: vertical, Panels: []ResizablePanelProps{{ID: "a", Label: "Editable", MinSize: 20, Collapsible: true, Content: func(c *ui.Context) { ui.TextInput(c, &text).Label("Edit pane") }}, {ID: "b", Label: "Other"}}}
			r := panelTester(&layout, &p, nil, false)
			r.SetDirection(ui.RightToLeft)
			if err := r.Click("Edit pane"); err != nil {
				t.Fatal(err)
			}
			layout = ResizableLayout{"a": 0, "b": 100}
			r.Frame()
			if !panelTestAccess(t, r, ui.RoleSplitter, "Resize Editable").Focused {
				t.Fatal("controlled collapse stranded focus in hidden content")
			}
			// Pinned MyGo AccessIncrement synthesizes Up for RTL; the vertical
			// separator explicitly uses LTR and therefore receives Right.
			increment := ui.KeyUp
			if vertical {
				increment = ui.KeyRight
			}
			for range 6 {
				r.Key(0, increment)
			}
			if layout["a"] < 20 {
				t.Fatalf("native increment alias did not reopen pane: %v", layout)
			}
		})
	}
}

func TestResizablePanelGroupResizesProportionallyAndCancelsStaleDrag(t *testing.T) {
	layout := ResizableLayout{"a": 30, "b": 70}
	p := ResizablePanelGroupProps{Panels: []ResizablePanelProps{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}}
	r := panelTester(&layout, &p, nil, false)
	r.SetSize(800, 240)
	if math.Abs(float64(panelTestAccess(t, r, ui.RoleGroup, "A").Bounds.W)-799*.3) > 1 {
		t.Fatal("window resize lost percentage layout")
	}
	h := panelTestAccess(t, r, ui.RoleSplitter, "Resize A")
	x, y := h.Bounds.X+4, h.Bounds.Y+30
	r.Press(x, y)
	r.Move(x+60, y)
	layout = ResizableLayout{"a": 60, "b": 40}
	r.Frame()
	r.Release(x+100, y)
	if math.Abs(layout["a"]-60) > 1e-5 {
		t.Fatalf("stale drag overwrote an external layout: %v", layout)
	}
	layout = ResizableLayout{"a": 1e308, "b": 1e308}
	r.Frame()
	if math.Abs(layout["a"]+layout["b"]-100) > 1e-5 {
		t.Fatalf("huge finite values escaped normalization: %v", layout)
	}
}

func TestResizablePanelRedistributionKeepsTotalAndEveryConstraint(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for n := 2; n <= 6; n++ {
		panels := make([]ResizablePanelProps, n)
		values := make([]float64, n)
		for i := range panels {
			panels[i] = ResizablePanelProps{ID: fmt.Sprint(i), MinSize: float64(5 + (i%3)*5), MaxSize: 80, Collapsible: i%2 == 0, CollapsedSize: 1}
			values[i] = 100 / float64(n)
		}
		limits, _, valid := panelLimits(panels)
		if !valid {
			t.Fatal("invalid test constraints")
		}
		for step := 0; step < 300; step++ {
			before := slices.Clone(values)
			boundary, delta := rng.Intn(n-1), rng.Float64()*400-200
			values = panelResize(values, limits, boundary, delta)
			if math.Abs(panelSum(values)-100) > 1e-6 {
				t.Fatalf("n=%d move=%d boundary=%d delta=%v total changed: %v -> %v", n, step, boundary, delta, before, values)
			}
			for i, v := range values {
				l := limits[i]
				if !panelFinite(v) || v > l.max+1e-6 || v < l.min-1e-6 && (!l.collapsible || math.Abs(v-l.collapsed) > 1e-6) {
					t.Fatalf("n=%d move=%d constraint broken: %v limits=%+v", n, step, values, limits)
				}
			}
		}
	}
}

func TestResizableCollapsedModesNormalizeToAFeasibleCombination(t *testing.T) {
	for _, test := range []struct {
		sizes  []float64
		limits []panelLimit
	}{
		{[]float64{0, 0, 0}, []panelLimit{{min: 20, max: 100, collapsible: true}, {min: 20, max: 100, collapsible: true}, {min: 20, max: 100, collapsible: true}}},
		{[]float64{33, 33, 34}, []panelLimit{{min: 40, max: 100, collapsible: true}, {min: 40, max: 100, collapsible: true}, {min: 40, max: 100, collapsible: true}}},
		{[]float64{33, 33, 34}, []panelLimit{{min: 70, max: 90, collapsible: true}, {min: 40, max: 60, collapsible: true}, {min: 10, max: 20, collapsible: true}}},
	} {
		got, valid := panelNormalize(test.sizes, test.limits)
		if !valid || math.Abs(panelSum(got)-100) > 1e-6 {
			t.Fatalf("satisfiable collapsed modes became invalid: %v -> %v", test.sizes, got)
		}
		for i, v := range got {
			l := test.limits[i]
			if v > l.max+1e-6 || v < l.min-1e-6 && math.Abs(v-l.collapsed) > 1e-6 {
				t.Fatalf("normalized value outside allowed modes: %v", got)
			}
		}
	}
	if _, valid := panelNormalize([]float64{50, 50}, []panelLimit{{min: 60, max: 70, collapsible: true}, {min: 60, max: 70, collapsible: true}}); valid {
		t.Fatal("discrete collapsed constraints with no possible total were accepted")
	}
}
