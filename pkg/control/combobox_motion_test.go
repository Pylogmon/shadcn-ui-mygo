package control

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// The fixed CSS `ease` contract is independent of the product's easing value.
var comboSourceEase = ui.CubicBezier(.25, .1, .25, 1)

// The harness observes our own value-only state, while all rendering, text
// editing, focus, input and accessibility checks use MyGo's public native tester.
type comboMotionHarness struct {
	r               *ui.Tester
	state           *comboRuntime
	value           ComboboxValue
	values          []string
	query           string
	popup, multiple bool
	chosen, inputs  int
	frameAt         time.Time
}

func newComboMotionHarness(t *testing.T, popup, multiple, top, rtl bool) *comboMotionHarness {
	t.Helper()
	h := &comboMotionHarness{popup: popup, multiple: multiple}
	h.r = ui.NewTester(func(c *ui.Context) {
		h.frameAt = c.Now()
		theme.Apply(c)
		c.Root().Background(ui.RGB(255, 255, 255))
		if rtl {
			c.Root().Direction(ui.RightToLeft)
		}
		p := ComboboxWithProps{Label: "Motion combo", PopupInput: popup, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}},
			ItemContent:   func(c *ui.Context, _ Option) { ui.Box(c).Size(28, 16).Background(ui.RGB(255, 0, 0)) },
			OnValueChange: func(ComboboxValue) { h.chosen++ }, OnValuesChange: func([]string) { h.chosen++ }, OnInputChange: func(string) { h.inputs++ }}
		var root ui.Element
		if multiple {
			root = MultiComboboxWith(c, &h.values, &h.query, p)
		} else {
			root = ComboboxWith(c, &h.value, &h.query, p)
		}
		y := float32(40)
		if top {
			y = 280
		}
		root.Absolute().Left(40).Top(y).Width(240)
		h.state = ui.Local(root, "shadcn.control.selection-combobox", func() comboRuntime { return comboRuntime{} })
		Button(c, "After", ButtonProps{}).Absolute().Left(300).Top(10)
	}, 420, 340)
	return h
}

func TestComboboxMotionFirstVisibleNativeFrameUsesObservedSide(t *testing.T) {
	for _, top := range []bool{false, true} {
		t.Run(map[bool]string{false: "bottom", true: "top"}[top], func(t *testing.T) {
			h := newComboMotionHarness(t, false, false, top, false)
			// Exactly one first frame: opacity=0 can skip Draw. The next frame
			// must use public completed geometry before sampling its entry pose.
			h.state.open = true
			h.r.Frame()
			start := h.r.Now()
			h.r.Advance(20 * time.Millisecond)
			_, x, y := comboRedMass(h.r)
			elapsed := h.frameAt.Sub(start)
			if elapsed <= 0 || elapsed >= 20*time.Millisecond {
				t.Fatalf("expected first native refresh, got %s", elapsed)
			}
			f := float64(comboSourceEase(float32(elapsed) / float32(100*time.Millisecond)))
			scale := .95 + .05*f
			a, b := h.state.motion.attachment, h.state.motion.panelBounds
			ox, oy := float64(b.X+b.W*a.x), float64(b.Y+b.H*a.y)
			h.r.Advance(100 * time.Millisecond)
			h.r.Frame()
			_, cx, cy := comboRedMass(h.r)
			dy := -8 * (1 - f)
			if top {
				dy = -dy
			}
			comboNear(t, "first visible x", x, ox+(cx-ox)*scale, .8)
			comboNear(t, "first visible side-specific y", y, oy+(cy-oy)*scale+dy, .8)
		})
	}
}

func TestComboboxPopupSearchUsesVega32DIPGroup(t *testing.T) {
	h := newComboMotionHarness(t, true, false, false, false)
	h.open(t)
	h.r.Advance(comboMotionDuration)
	h.r.Frame()
	input := selectionComboInput(t, h.r, "Motion combo")
	comboNear(t, "search native editor plus 1DIP group borders", float64(input.Bounds.H+2), 32, .01)
	comboNear(t, "search group margin top plus borders", float64(input.Bounds.Y-h.state.motion.panelBounds.Y), 6, .01)
}

func (h *comboMotionHarness) open(t *testing.T) {
	t.Helper()
	if h.popup {
		n, ok := h.r.AccessByRole(ui.RoleButton, "Motion combo")
		if !ok {
			t.Fatal("missing popup trigger")
		}
		h.r.ClickAt(n.Bounds.X+12, n.Bounds.Y+n.Bounds.H/2)
	} else {
		openSelectionCombo(t, h.r, "Motion combo")
	}
	h.r.Frame() // Resolve the first layout's observed side at the same instant.
	if !h.state.open {
		t.Fatal("combobox did not open")
	}
}

func comboRedMass(r *ui.Tester) (mass, x, y float64) {
	im := r.Image()
	for py := im.Bounds().Min.Y; py < im.Bounds().Max.Y; py++ {
		for px := im.Bounds().Min.X; px < im.Bounds().Max.X; px++ {
			p := im.RGBAAt(px, py)
			w := float64(int(p.R)-max(int(p.G), int(p.B))) / 255
			if w > 0 {
				mass += w
				x += (float64(px) + .5) * w
				y += (float64(py) + .5) * w
			}
		}
	}
	if mass > 0 {
		x /= mass
		y /= mass
	}
	return
}

func comboNear(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %.6f, want %.6f ± %.6f", name, got, want, tolerance)
	}
}

func TestComboboxSourceMotionNativePixelsAndObservedPlacement(t *testing.T) {
	for _, popup := range []bool{false, true} {
		for _, top := range []bool{false, true} {
			for _, rtl := range []bool{false, true} {
				name := map[bool]string{false: "inline", true: "popup"}[popup] + map[bool]string{false: "/bottom", true: "/top"}[top] + map[bool]string{false: "/ltr", true: "/rtl"}[rtl]
				t.Run(name, func(t *testing.T) {
					h := newComboMotionHarness(t, popup, false, top, rtl)
					h.open(t)
					wantSide := comboSideBottom
					if top {
						wantSide = comboSideTop
					}
					if h.state.motion.attachment.side != wantSide {
						t.Fatalf("observed attachment %+v", h.state.motion.attachment)
					}
					start := h.r.Now()
					h.r.Advance(comboMotionDuration)
					h.r.Frame()
					full, cx, cy := comboRedMass(h.r)
					if full < 500 {
						t.Fatalf("native content did not paint: mass %.2f", full)
					}
					origin := h.state.motion.attachment
					bounds := h.state.motion.panelBounds
					ox, oy := float64(bounds.X+bounds.W*origin.x), float64(bounds.Y+bounds.H*origin.y)
					// Reopen a fully removed subtree to sample an independent full entry.
					h.state.open = false
					h.r.Frame()
					h.r.Advance(comboMotionDuration)
					h.r.Frame()
					h.state.open = true
					h.r.Frame()
					h.r.Frame()
					start = h.r.Now()
					for _, ms := range []int{0, 10, 20, 25, 50, 75, 99, 100} {
						h.r.Advance(start.Add(time.Duration(ms) * time.Millisecond).Sub(h.r.Now()))
						h.r.Frame()
						f := float64(comboSourceEase(float32(ms) / 100))
						scale := .95 + .05*f
						mass, x, y := comboRedMass(h.r)
						comboNear(t, "native alpha × scaled area", mass/full, f*scale*scale, .012)
						if ms > 0 {
							dy := -8 * (1 - f)
							if top {
								dy = -dy
							}
							comboNear(t, "native x/origin", x, ox+(cx-ox)*scale, .8)
							comboNear(t, "native y/origin/8DIP slide", y, oy+(cy-oy)*scale+dy, .8)
						}
					}
					before, _, _ := comboRedMass(h.r)
					h.state.open = false
					h.r.Frame()
					start = h.r.Now()
					atClose, _, _ := comboRedMass(h.r)
					comboNear(t, "close t0 continuous", atClose, before, .02)
					if _, ok := h.r.AccessByRole(ui.RoleOption, "Go"); ok {
						t.Fatal("closing option remained in accessibility tree")
					}
					for _, ms := range []int{20, 50, 75, 99, 100} {
						h.r.Advance(start.Add(time.Duration(ms) * time.Millisecond).Sub(h.r.Now()))
						h.r.Frame()
						f := float64(comboSourceEase(float32(ms) / 100))
						scale := 1 - .05*f
						mass, x, y := comboRedMass(h.r)
						comboNear(t, "native exit alpha × scaled area", mass/full, (1-f)*scale*scale, .012)
						if ms < 99 {
							comboNear(t, "exit has no slide", y, oy+(cy-oy)*scale, .8)
							comboNear(t, "exit x/origin", x, ox+(cx-ox)*scale, .8)
						}
					}
					if h.r.HasText("Go") || h.r.HasText("Rust") {
						t.Fatal("source exit remained mounted beyond 100ms")
					}
				})
			}
		}
	}
}

func TestComboboxMotionRapidReversalKeepsPoseAndNativeEditor(t *testing.T) {
	h := newComboMotionHarness(t, true, false, false, false)
	h.open(t)
	h.r.Advance(100 * time.Millisecond)
	h.r.Frame()
	h.r.Type("Go")
	h.r.Key(0, ui.KeyHome)
	h.r.Key(0, ui.KeyRight)
	h.r.Key(ui.Shift, ui.KeyRight)
	h.r.Frame()
	if h.state.selectionStart != 1 || h.state.selectionEnd != 2 {
		t.Fatalf("native selection before close %d:%d", h.state.selectionStart, h.state.selectionEnd)
	}
	h.state.open = false
	h.r.Frame()
	h.r.Advance(30 * time.Millisecond)
	h.r.Frame()
	before, _, _ := comboRedMass(h.r)
	if h.query != "Go" {
		t.Fatalf("closing popup editor flickered query: %q", h.query)
	}
	h.open(t)
	after, _, _ := comboRedMass(h.r)
	comboNear(t, "reversal t0", after, before, .01)
	if h.state.selectionStart != 1 || h.state.selectionEnd != 2 {
		t.Fatalf("retained native selection %d:%d", h.state.selectionStart, h.state.selectionEnd)
	}
	h.r.Type("X")
	if h.query != "GX" {
		t.Fatalf("native buffer/selection lost across reversal: %q", h.query)
	}
	h.r.Advance(100 * time.Millisecond)
	h.r.Frame()
	if h.chosen != 0 {
		t.Fatalf("native editing chose %d options", h.chosen)
	}
}

func TestComboboxClosingFilterQueryAndSelectionTiming(t *testing.T) {
	for _, popup := range []bool{false, true} {
		for _, multiple := range []bool{false, true} {
			t.Run(map[bool]string{false: "inline", true: "popup"}[popup]+map[bool]string{false: "/single", true: "/multiple"}[multiple], func(t *testing.T) {
				h := newComboMotionHarness(t, popup, multiple, false, false)
				h.open(t)
				h.r.Advance(comboMotionDuration)
				h.r.Frame()
				h.r.Type("Go")
				h.r.Key(0, ui.KeyEscape)
				h.r.Frame()
				want := "Go"
				if multiple && !popup {
					want = ""
				}
				if h.query != want {
					t.Fatalf("dismissal query=%q want %q", h.query, want)
				}
				if len(h.state.items) != 1 || h.state.items[0].option.Value != "go" {
					t.Fatalf("exit filter changed: %+v", h.state.items)
				}
				h.r.Advance(99 * time.Millisecond)
				h.r.Frame()
				if h.query != want {
					t.Fatalf("query reset before exit completed: %q", h.query)
				}
				h.r.Advance(time.Millisecond)
				h.r.Frame()
				if h.query != "" || h.state.closing {
					t.Fatalf("completed dismissal: query=%q closing=%v", h.query, h.state.closing)
				}
				if h.chosen != 0 || h.inputs != 1 {
					t.Fatalf("dismissal published value/input callbacks %d/%d", h.chosen, h.inputs)
				}
			})
		}
	}
}

func TestComboboxClosingPanelImmediatelyLetsInputThrough(t *testing.T) {
	h := newComboMotionHarness(t, false, false, false, false)
	h.open(t)
	h.r.Advance(100 * time.Millisecond)
	h.r.Frame()
	option, ok := h.r.AccessByRole(ui.RoleOption, "Go")
	if !ok {
		t.Fatal("missing option")
	}
	h.state.open = false
	h.r.Frame()
	h.r.Key(0, ui.KeyTab)
	h.r.Frame()
	if !h.r.Focused("After") {
		t.Fatalf("closing panel intercepted native Tab: %v", h.r.AccessAll())
	}
	h.r.ClickAt(option.Bounds.X+12, option.Bounds.Y+option.Bounds.H/2)
	h.r.Frame()
	if h.value.Valid || h.chosen != 0 {
		t.Fatal("closing option accepted pointer input")
	}
}

func TestComboboxMotionReducedAndHiddenSettleWithoutFrames(t *testing.T) {
	for _, mode := range []string{"reduced", "hidden"} {
		t.Run(mode, func(t *testing.T) {
			h := newComboMotionHarness(t, true, false, false, false)
			h.open(t)
			h.r.Advance(20 * time.Millisecond)
			h.r.Frame()
			if mode == "reduced" {
				h.r.SetPreferences(ui.Preferences{ReduceMotion: true})
			} else {
				h.r.SetWindowHidden(true)
			}
			h.state.open = false
			h.r.Frame()
			h.r.Frame()
			if h.state.motion.from != h.state.motion.to || h.state.closing {
				t.Fatal("reduced/hidden left an active exit")
			}
			if h.r.HasText("Go") {
				t.Fatal("reduced/hidden retained closing content")
			}
			if h.r.FramePending() {
				t.Fatal("settled reduced/hidden combobox requests animation frames")
			}
		})
	}
}

func TestComboboxMotionUnknownGeometryRemainsCentered(t *testing.T) {
	for _, pair := range [][2]ui.Rect{
		{{X: 20, Y: 20, W: 40, H: 20}, {X: 30, Y: 25, W: 80, H: 40}},
		{{X: 20, Y: 20, W: 40, H: 20}, {X: 80, Y: 80, W: 80, H: 40}},
		{{X: 20, Y: 20, W: 40, H: 20}, {}},
		{{X: float32(math.NaN()), Y: 20, W: 40, H: 20}, {X: 20, Y: 46, W: 80, H: 40}},
	} {
		a := comboActualAttachment(pair[0], pair[1])
		if a.side != comboSideUnknown || a.x != .5 || a.y != .5 {
			t.Fatalf("unproved placement: %+v", a)
		}
		p := comboPoseHidden(a, true)
		if p.x != 0 || p.y != 0 || p.scale != .95 || p.opacity != 0 {
			t.Fatalf("unknown pose %+v", p)
		}
	}
}
