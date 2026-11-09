package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
)

type previewSurface struct {
	spec desktop.Spec
	view func(*ui.Context)
}

func (*previewSurface) SetAnchor(ui.Rect) {}
func (*previewSurface) Invalidate()       {}
func (*previewSurface) Close()            {}

type previewBackend struct{ surfaces []*previewSurface }

func (b *previewBackend) Open(spec desktop.Spec, view func(*ui.Context), closed func()) (desktop.Surface, error) {
	s := &previewSurface{spec, view}
	b.surfaces = append(b.surfaces, s)
	return s, nil
}
func (*previewBackend) InvalidateParent() {}

func TestSourceExamplesUseRealIndependentContent(t *testing.T) {
	b := &previewBackend{}
	a := &application{host: desktop.NewHostWithBackend(b)}
	r := ui.NewTesterWith(a.view, 760, 1900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	for _, entry := range []struct{ trigger, content string }{
		{"Show tooltip", "Add to library"},
		{"inline-start", "Add to library"}, {"left", "Add to library"}, {"top", "Add to library"}, {"bottom", "Add to library"}, {"right", "Add to library"}, {"inline-end", "Add to library"},
		{"Information", "Additional information"},
		{"Learn how it works", "To learn more about how this works, check out the docs. If you have any questions, please reach out to us."},
		{"Unavailable", "This feature is currently unavailable"},
		{"Save changes", "Save Changes"},
		{"Tooltip documentation", "Click to read the documentation"},
		{"Status", "Last updated 2 hours ago"},
	} {
		r.Move(750, 1890)
		bounds, ok := r.Find(entry.trigger)
		if !ok {
			t.Fatalf("missing source trigger %q", entry.trigger)
		}
		before := len(b.surfaces)
		r.Move(bounds.X+bounds.W/2, bounds.Y+bounds.H/2)
		if len(b.surfaces) != before+1 {
			t.Fatalf("source trigger %q did not open a real child", entry.trigger)
		}
		surface := b.surfaces[len(b.surfaces)-1]
		child := ui.NewTesterWith(surface.view, surface.spec.MaxWidth, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		if !child.HasText(entry.content) || !surface.spec.IgnoreMouseEvents || surface.spec.Focusable {
			t.Fatalf("source trigger %q lost its inert independent content", entry.trigger)
		}
	}
	if err := r.Click("Update status"); err != nil || !a.updated || a.revision != 1 {
		t.Fatal("status button did not update application content")
	}
	bounds, _ := r.Find("Status")
	r.Move(bounds.X+bounds.W/2, bounds.Y+bounds.H/2)
	child := ui.NewTesterWith(b.surfaces[len(b.surfaces)-1].view, 334, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !child.HasText("Updated just now") {
		t.Fatal("formatted source flow retained stale content")
	}
	if _, ok := child.AccessByRole(ui.RoleTooltip, "Active. Updated just now."); !ok {
		t.Fatal("updated rich visuals retained a stale semantic description")
	}
}
