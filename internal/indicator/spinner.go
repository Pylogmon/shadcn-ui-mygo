// Package indicator contains the shared, non-interactive loading visual used
// by display primitives, busy buttons and notifications.
package indicator

import (
	_ "embed"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
)

//go:embed loader-circle.svg
var loaderCircleSource []byte

var loaderCircle = ui.MustParseSVG(loaderCircleSource)

type Props struct {
	Size               float32
	Label, Description string
	Static, Decorative bool
}

type spinnerIconKey struct{}
type spinnerLoopKey struct{}

// Spinner retains native unknown-progress semantics and lets the SVG inherit
// the actual current text color, including parent controls and fluent styles.
func Spinner(c *ui.Context, p Props) ui.Element {
	size, invalid := p.Size, false
	if size == 0 {
		size = 16
	} else if size < 0 || math.IsNaN(float64(size)) || math.IsInf(float64(size), 0) {
		size, invalid = 16, true
	}
	label := p.Label
	if label == "" {
		label = "Loading"
	}
	root := ui.Box(c).Size(size, size).Shrink(0).Role(ui.RoleProgress).Range(0, 1, -1).
		Label(label).Description(p.Description).Invalid(invalid).AccessHidden(p.Decorative)
	angle := float32(0)
	if !p.Static && !motion.Reduced(c) {
		angle = root.Loop(spinnerLoopKey{}, time.Second, ui.Linear) * 360
	}
	return root.Children(func() {
		ui.Icon(c.Key(spinnerIconKey{}), loaderCircle).Absolute().Fill().Rotate(angle).IgnoreInput(true).AccessHidden(true)
	})
}
