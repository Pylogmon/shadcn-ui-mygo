package desktop

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestSelectLargeFontsCrossNativeContextAndRefreshWithoutRemount(t *testing.T) {
	b := &sizeBackend{}
	h := NewHostWithBackend(b)
	value := "one"
	font, override := float32(28), float32(0)
	var triggerText, optionText, reference ui.Rect
	const sample = "MMMM"
	const scale = float32(1.5)
	parent := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Light()
		tokens.FontSize = font
		c.SetTheme(tokens.Theme())
		BindHost(c, h)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Gap(8).Children(func() {
			_, err := Select(c, nil, "large", &value, control.SelectProps{Label: "Large choice", FontSize: override, Options: []control.Option{{Value: "one", Label: sample, Group: "Choices"}, {Value: "two", Label: "Two", Group: "Choices"}},
				TriggerContent: func(c *ui.Context, o control.Option) {
					ui.Text(c, o.Label).DrawOver(func(_ *ui.Painter, r ui.Rect) { triggerText = r })
				},
				ItemContent: func(c *ui.Context, o control.Option) {
					text := ui.Text(c, o.Label).AlignSelf(ui.Start)
					if o.Value == "one" {
						text.DrawOver(func(_ *ui.Painter, r ui.Rect) { optionText = r })
					}
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			expected := font * scale
			if override > 0 {
				expected = override * scale
			}
			ui.Text(c, sample).AlignSelf(ui.Start).FontSize(expected).FontWeight(500).DrawOver(func(_ *ui.Painter, r ui.Rect) { reference = r })
		})
	}, 640, 260, ui.TesterOptions{Preferences: ui.Preferences{TextScale: scale, ReduceMotion: true}})
	if err := parent.Click("Large choice"); err != nil {
		t.Fatal(err)
	}
	// The backend's independent tester deliberately has default preferences;
	// the menu must use the parent's resolved typography snapshot.
	child := ui.NewTesterWith(b.surfaces[0].view, 640, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	assertFont := func(expected float32) {
		t.Helper()
		// Menu text is regular weight; compare against its native measurement.
		var expectedWidth float32
		ui.NewTester(func(c *ui.Context) {
			expectedWidth, _ = c.MeasureText(1000, ui.Span{Text: sample, Size: expected})
		}, 640, 200)
		if math.Abs(float64(optionText.W-expectedWidth)) > .1 || math.Abs(float64(triggerText.W-reference.W)) > .1 {
			t.Fatalf("native popup/trigger lost resolved font: trigger=%v option=%v reference=%v expected option width=%g", triggerText, optionText, reference, expectedWidth)
		}
		row, ok := child.AccessByRole(ui.RoleOption, sample)
		if !ok || row.Bounds.H < expected*(20.0/14)+12-.1 || row.Bounds.H < optionText.H {
			t.Fatalf("native option row clipped large text: row=%+v text=%v", row, optionText)
		}
	}
	assertFont(42)
	font = 35
	parent.Frame()
	child.Frame()
	assertFont(52.5)
	override = 20
	parent.Frame()
	child.Frame()
	assertFont(30)
	if len(b.surfaces) != 1 || b.surfaces[0].closes != 0 || len(b.sizes) == 0 {
		t.Fatal("font update remounted an existing resize-capable native popup")
	}
	parent.Frame()
	child.Frame()
	b.surfaces[0].invalidates = 0
	parent.Frame()
	if b.surfaces[0].invalidates != 0 || parent.FramePending() || child.FramePending() {
		t.Fatal("stable typography kept scheduling native frames")
	}
}
