package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestGalleryNavigationAndTheme(t *testing.T) {
	s := New()
	r := ui.NewTester(s.View, 1280, 900)
	if err := r.Click("Toggle theme"); err != nil {
		t.Fatal(err)
	}
	if s.Theme.Mode != theme.DarkMode {
		t.Fatalf("appearance=%s", s.Theme.Mode)
	}
	if err := r.Click("Forms & controls"); err != nil {
		t.Fatal(err)
	}
	if s.Page != "controls" {
		t.Fatal(s.Page)
	}
	if _, ok := r.Find("Make a choice"); !ok {
		t.Fatal("controls page did not render")
	}
	if err := r.Click("Overlays"); err != nil {
		t.Fatal(err)
	}
	if err := r.Click("Edit profile"); err != nil {
		t.Fatal(err)
	}
	r.Advance(300 * time.Millisecond)
	if !s.Dialog {
		t.Fatal("dialog did not open")
	}
	r.Key(0, ui.KeyEscape)
	r.Advance(300 * time.Millisecond)
	if s.Dialog {
		t.Fatal("Escape did not close dialog")
	}
}

func TestGalleryPagesRenderInBothAppearances(t *testing.T) {
	for _, p := range pages {
		for _, dark := range []bool{false, true} {
			t.Run(p.Key+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				s := New()
				s.Page = p.Key
				if dark {
					s.Theme.Mode = theme.DarkMode
				}
				r := ui.NewTester(s.View, 1280, 900)
				if len(r.Texts()) < 10 {
					t.Fatalf("empty gallery page %q", p.Key)
				}
				// Producing pixels also exercises native painting paths and chart/code views.
				im := r.Image()
				if im.Bounds().Dx() != 1280 || im.Bounds().Dy() != 900 {
					t.Fatal(im.Bounds())
				}
			})
		}
	}
}
