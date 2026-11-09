package display

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestUnknownProgressAndReducedMotionDoNotLoseSemantics(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); Progress(c, -1).Width(200).Label("Download") }, 260, 120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	n, ok := r.AccessByRole(ui.RoleProgress, "Download")
	if !ok || n.Now >= n.Min {
		t.Fatalf("unknown progress became a determinate value: %+v", n)
	}
	r.Advance(3 * time.Second)
	if r.FramePending() {
		t.Fatal("reduced-motion progress still schedules frames")
	}
}
func TestReducedMotionSpinnerDoesNotScheduleIdleFrames(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); Spinner(c) }, 100, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(2 * time.Second)
	if r.FramePending() {
		t.Fatal("reduced-motion spinner still schedules frames")
	}
	n, ok := r.AccessByRole(ui.RoleProgress, "Loading")
	if !ok || n.Now >= n.Min {
		t.Fatalf("spinner lost native unknown progress semantics %+v", n)
	}
}
func TestProgressInvalidValuesRemainFinite(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		r := ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); Progress(c, value).Width(200).Label("Work") }, 260, 120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		n, _ := r.AccessByRole(ui.RoleProgress, "Work")
		if math.IsNaN(n.Now) || math.IsInf(n.Now, 0) {
			t.Fatalf("nonfinite accessible value %+v", n)
		}
	}
}
func TestTypographyHasNativeHeadingRanks(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Children(func() {
			Typography(c, "Page title", H1)
			Typography(c, "Section title", H2)
			Typography(c, "Supporting text", Paragraph)
		})
	}, 400, 220)
	for label, level := range map[string]int{"Page title": 1, "Section title": 2} {
		n, ok := r.AccessByRole(ui.RoleHeading, label)
		if !ok || n.Level != level {
			t.Fatalf("heading %q %+v", label, n)
		}
	}
	if _, ok := r.AccessByRole(ui.RoleHeading, "Supporting text"); ok {
		t.Fatal("paragraph exposed as heading")
	}
}

func TestSemanticSeparatorAndProgressValueText(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Children(func() {
			Separator(c, false)
			Separator(c, true, SeparatorProps{Semantic: true, Label: "Section boundary"}).Height(40)
			Progress(c, .65, ProgressProps{Label: "Upload", ValueText: "65 of 100 files"}).Width(200)
		})
	}, 300, 180)
	n, ok := r.AccessByRole(ui.RoleSeparator, "Section boundary")
	if !ok || !n.Vertical {
		t.Fatalf("semantic separator %+v", n)
	}
	p, ok := r.AccessByRole(ui.RoleProgress, "Upload")
	if !ok || p.Now != .65 || p.Value != "65 of 100 files" {
		t.Fatalf("progress value %+v", p)
	}
	count := 0
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleSeparator {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("decorative separator exposed, count=%d", count)
	}
}
