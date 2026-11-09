package display_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestPublicShimmerRetainsNativeTextAndStopsOnPreferenceChange(t *testing.T) {
	base := ui.RGB(30, 70, 180)
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		display.ShimmerText(c, "正在处理 Native text", display.ShimmerTextProps{
			Base: &base, Font: ui.Font{Size: 28},
		}).Width(220).SingleLine().Ellipsis("…").Role(ui.RoleText).Description("Upload title")
	}, 300, 100)
	n, ok := r.AccessByRole(ui.RoleText, "正在处理 Native text")
	if !ok || n.Bounds.W != 220 || n.Description != "Upload title" {
		t.Fatalf("public utility lost native text composition: %+v", n)
	}
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(500 * time.Millisecond)
	if bytes.Equal(before, r.Image().Pix) || !r.FramePending() {
		t.Fatal("public utility did not paint a moving sweep")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	before = append([]byte(nil), r.Image().Pix...)
	r.Advance(time.Second)
	if !bytes.Equal(before, r.Image().Pix) || r.FramePending() {
		t.Fatal("preference change did not stop the public utility")
	}
}
