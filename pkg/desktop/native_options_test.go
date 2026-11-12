package desktop

import (
	"runtime"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestNativeFlyoutDecorationsFollowSurfaceStyle(t *testing.T) {
	for _, tc := range []struct {
		name   string
		props  PopoverProps
		styled bool
	}{
		{"styled-popover", PopoverProps{Plain: true}, true},
		{"styled-tooltip", PopoverProps{Plain: true, NonFocusable: true, IgnoreMouseEvents: true, ShowInactive: true}, true},
		{"native-popover", PopoverProps{SystemStyle: true}, false},
		{"native-tooltip", PopoverProps{Plain: true, SystemStyle: true, NonFocusable: true}, false},
		{"native-hover-card", PopoverProps{Plain: true, SystemStyle: true, ShowInactive: true}, false},
		{"system-container", PopoverProps{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := normalize(tc.props, ui.Rect{X: 20, Y: 40, W: 80, H: 30}, 6)
			if err != nil {
				t.Fatal(err)
			}
			options := nativeFlyoutOptions(nil, spec, func(*ui.Context) {})
			wantShadow := runtime.GOOS != "windows" || !tc.styled
			if options.Shadow != wantShadow {
				t.Fatalf("native shadow/outline = %v, want %v on %s", options.Shadow, wantShadow, runtime.GOOS)
			}
			if !options.Hidden || options.Focusable == tc.props.NonFocusable {
				t.Fatal("decoration policy changed initial visibility or keyboard ownership")
			}
		})
	}
}
