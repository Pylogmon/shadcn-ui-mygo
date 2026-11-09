package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestCreatorAvatarPreviewControlledModesAndGroup(t *testing.T) {
	for _, mode := range []theme.Mode{theme.LightMode, theme.DarkMode} {
		t.Run(string(mode), func(t *testing.T) {
			s := New()
			s.Theme.Mode = mode
			r := ui.NewTester(func(c *ui.Context) {
				s.Theme.Apply(c)
				ui.Column(c).Width(400).Padding(24).Children(func() { s.avatarsPreview(c) })
			}, 450, 500)
			if !r.HasText("Image ready") || r.HasText("SD") {
				t.Fatal("decoded sample did not paint before fallback")
			}
			if _, ok := r.AccessByRole(ui.RoleImage, "3 more people"); !ok {
				t.Fatal("preview group lost remaining members")
			}
			for _, example := range []struct{ button, status string }{{"Initials", "Showing initials"}, {"Loading", "Loading image…"}, {"Unavailable", "Image unavailable — showing initials"}} {
				advancedClick(t, r, example.button)
				if !r.HasText(example.status) || !r.HasText("SD") {
					t.Fatalf("%s failed to publish controlled state: %v", example.button, r.Texts())
				}
			}
			advancedClick(t, r, "Image")
			if r.HasText("SD") || !r.HasText("Image ready") {
				t.Fatal("image recovery retained stale fallback")
			}
		})
	}
}
