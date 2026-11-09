package desktop

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// TooltipSide includes the fixed source's physical and logical placements.
type TooltipSide string

const (
	TooltipTop         TooltipSide = "top"
	TooltipBottom      TooltipSide = "bottom"
	TooltipLeft        TooltipSide = "left"
	TooltipRight       TooltipSide = "right"
	TooltipInlineStart TooltipSide = "inline-start"
	TooltipInlineEnd   TooltipSide = "inline-end"
)

func tooltipPlacement(p TooltipProps, direction ui.Direction) (mygo.Placement, error) {
	placement := p.Placement
	if placement == "" {
		placement = mygo.PlacementTop
	}
	if p.Side == "" {
		return placement, nil
	}
	side := string(p.Side)
	switch p.Side {
	case TooltipTop, TooltipBottom, TooltipLeft, TooltipRight:
	case TooltipInlineStart, TooltipInlineEnd:
		left := p.Side == TooltipInlineStart
		if direction == ui.RightToLeft {
			left = !left
		}
		side = "right"
		if left {
			side = "left"
		}
	default:
		return "", fmt.Errorf("desktop: unknown tooltip side %q", p.Side)
	}
	for _, suffix := range []string{"-start", "-end"} {
		if strings.HasSuffix(string(placement), suffix) {
			side += suffix
			break
		}
	}
	return mygo.Placement(side), nil
}

// A 10-DIP square with 2-DIP rounded corners, rotated by 45 degrees, fits
// inside this transparent gutter. No pixels are drawn beyond a native window.
const tooltipArrowInset = 7

func tooltipContent(c *ui.Context, h *Host, key string, maxWidth int, background, foreground ui.Color, arrow bool, content func(*ui.Context)) {
	outer := ui.Column(c).MinWidth(0).MaxWidth(float32(maxWidth + 2*tooltipArrowInset))
	if arrow {
		outer.Padding(tooltipArrowInset)
		attachment := menuAttachment{}
		if entry := h.entries[key]; entry != nil {
			if geometry, ok := entry.surface.(GeometrySurface); ok {
				if anchor, window, known := geometry.Geometry(); known {
					attachment = menuActualAttachment(anchor, window, tooltipArrowInset+6)
				}
			}
		}
		// Only copied colors/geometry enter Draw. Unknown native geometry
		// omits the arrow instead of pointing toward an unverified side.
		outer.Draw(func(p *ui.Painter, r ui.Rect) {
			if !attachment.originKnown {
				return
			}
			x, y := r.X+r.W*attachment.x, r.Y+r.H*attachment.y
			switch attachment.side {
			case menuSideTop:
				y = r.Y + r.H - tooltipArrowInset
			case menuSideBottom:
				y = r.Y + tooltipArrowInset
			case menuSideLeft:
				x = r.X + r.W - tooltipArrowInset
			case menuSideRight:
				x = r.X + tooltipArrowInset
			}
			p.Transform(ui.Rotation(45), x, y, func() {
				p.Fill(ui.Rect{X: x - 5, Y: y - 5, W: 10, H: 10}, background, 2)
			})
		})
	}
	outer.Children(func() {
		ui.Row(c).MinWidth(0).MaxWidth(float32(maxWidth)).Wrap().AlignItems(ui.Center).Gap(6).
			Padding(6, 12).Radius(6).Background(background).TextColor(foreground).FontSize(12).
			Children(func() { content(c) })
	})
}
