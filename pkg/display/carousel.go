package display

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type CarouselProps struct {
	Count                      int
	Label                      string
	Vertical, Loop, Disabled   bool
	NoDrag, HideControls, Dots bool
	// Height fixes the viewport height. Vertical carousels default to 240 DIPs;
	// horizontal carousels otherwise use the tallest slide's content height.
	Height float32
	// Duration defaults to 250ms; negative values disable programmatic motion.
	Duration time.Duration
	// DragThreshold is in DIPs, defaulting to 15% of the viewport (24..96).
	// Native draggable velocity projection also recognizes short, fast swipes.
	DragThreshold float32
	Key           func(int) any
	OnSelect      func(int)
}
type carouselState struct {
	initialized, initialLayout bool
	last, count, virtual       int
	move, target               int
	vertical, loop, rtl        bool
	extent                     float32
	drag                       ui.Point
	active                     ui.Handle
}

// Carousel preserves the original controlled, bounded carousel API.
func Carousel(c *ui.Context, selected *int, count int, slide func(int)) ui.Element {
	return CarouselWith(c, selected, CarouselProps{Count: count}, slide)
}

// CarouselWith is a native draggable pager. Slides retain their tree identity
// as the selection changes; inactive slides cannot take input or accessibility
// focus. A keyboard-editable control within a slide retains its own arrow keys.
func CarouselWith(c *ui.Context, selected *int, p CarouselProps, slide func(int)) ui.Element {
	n := max(0, p.Count)
	if n == 0 {
		*selected = 0
	} else {
		*selected = max(0, min(*selected, n-1))
	}
	t := theme.TokensOf(c)
	label := p.Label
	if label == "" {
		label = "Carousel"
	}
	root := ui.Column(c).FillWidth().MinWidth(0).Gap(12).Role(ui.RoleRegion).RoleDescription("carousel").Label(label).Focusable().Disabled(p.Disabled)
	s := ui.Local(root, "shadcn.carousel", func() carouselState { return carouselState{initialLayout: true} })
	if (s.last != *selected || n == 0) && s.active.FocusWithin(c) {
		root.Focus()
	}
	choose := func(target int) {
		if n == 0 || p.Disabled {
			return
		}
		movement := target - *selected
		if p.Loop {
			target = (target%n + n) % n
		} else {
			target = max(0, min(target, n-1))
		}
		if target == *selected {
			return
		}
		s.move, s.target = movement, target
		*selected = target
		root.SetChanged()
		c.Invalidate()
		if p.OnSelect != nil {
			p.OnSelect(target)
		}
	}
	keys := func(e ui.Element) {
		if !e.Focused() || p.Disabled {
			return
		}
		back, forth := ui.KeyLeft, ui.KeyRight
		if c.Direction() == ui.RightToLeft {
			back, forth = forth, back
		}
		if p.Vertical {
			back, forth = ui.KeyUp, ui.KeyDown
		}
		switch {
		case e.Shortcut(0, back):
			choose(*selected - 1)
		case e.Shortcut(0, forth):
			choose(*selected + 1)
		case e.Shortcut(0, ui.KeyHome):
			choose(0)
		case e.Shortcut(0, ui.KeyEnd):
			choose(n - 1)
		}
	}
	keys(root)
	if !s.initialized || s.count != n || s.vertical != p.Vertical || s.loop != p.Loop || s.rtl != (c.Direction() == ui.RightToLeft) {
		s.initialized = true
		s.initialLayout = true
		s.drag = ui.Point{}
		s.move = 0
		s.last, s.virtual = *selected, *selected
		s.count = n
		s.vertical, s.loop = p.Vertical, p.Loop
		s.rtl = c.Direction() == ui.RightToLeft
	}
	if s.last != *selected {
		delta := *selected - s.last
		if p.Loop && s.move != 0 && s.target == *selected {
			delta = s.move
		}
		s.move = 0
		s.virtual += delta
		s.last = *selected
		s.drag = ui.Point{}
	}
	duration := p.Duration
	if duration == 0 {
		duration = 250 * time.Millisecond
	}
	instant := motion.Reduced(c) || duration < 0 || s.initialLayout
	height := p.Height
	if math.IsNaN(float64(height)) || math.IsInf(float64(height), 0) || height < 0 {
		height = 0
	}
	if p.Vertical && height == 0 {
		height = 240
	}
	root.Children(func() {
		viewport := ui.Box(c).FillWidth().MinWidth(0).Clip().Role(ui.RoleGroup).Label("Slides")
		if height > 0 {
			viewport.Height(height)
		}
		viewport.Draw(func(_ *ui.Painter, r ui.Rect) {
			extent := r.W
			if p.Vertical {
				extent = r.H
			}
			if extent != s.extent {
				s.extent = extent
				s.initialLayout = true
				c.Invalidate()
			}
		})
		viewport.Children(func() {
			track := ui.Row(c).FillWidth().AlignItems(ui.Stretch)
			if p.Vertical {
				track.Column().Height(height)
			} else if height > 0 {
				track.Height(height)
			}
			pose := ui.Pose{}
			if p.Vertical {
				pose.Y = -s.extent * float32(s.virtual)
			} else {
				pose.X = -s.extent * float32(s.virtual)
				if c.Direction() == ui.RightToLeft {
					pose.X = -pose.X
				}
			}
			positionDuration := duration
			if instant {
				positionDuration = -1
			}
			x := root.AnimateWith("carousel-x", pose.X, positionDuration, motion.EaseOut)
			y := root.AnimateWith("carousel-y", pose.Y, positionDuration, motion.EaseOut)
			track.Translate(x, y)
			draggable := !p.NoDrag && !p.Disabled && n > 1 && s.extent > 0
			if draggable {
				axis := ui.Horizontal
				bounds := ui.Rect{X: -s.extent, W: s.extent * 2}
				if p.Vertical {
					axis = ui.Vertical
					bounds = ui.Rect{Y: -s.extent, H: s.extent * 2}
				}
				dragDuration := duration
				if dragDuration <= 0 {
					dragDuration = time.Nanosecond
				}
				track.Draggable(&s.drag, ui.DraggableOptions{Axis: axis, Bounds: &bounds, Animation: ui.Animation{Duration: dragDuration, Ease: motion.EaseOut}, Snap: func(projected ui.Point) ui.Point {
					if track.PressCanceled() || !c.WindowActive() {
						return ui.Point{}
					}
					threshold := p.DragThreshold
					if threshold <= 0 || math.IsNaN(float64(threshold)) || math.IsInf(float64(threshold), 0) {
						threshold = max(24, min(96, s.extent*.15))
					}
					movement := -projected.X
					if p.Vertical {
						movement = -projected.Y
					} else if c.Direction() == ui.RightToLeft {
						movement = -movement
					}
					if movement >= threshold {
						choose(*selected + 1)
					} else if movement <= -threshold {
						choose(*selected - 1)
					}
					return ui.Point{}
				}})
			}
			start := 0
			if p.Loop && s.extent > 0 {
				// Keep every currently visible page in its current occurrence.
				// Retargeting an unfinished animation must not move its source
				// page to the far side of the track. DragOffset also requests
				// builds while native draggable motion is settling.
				drag := ui.Point{}
				if draggable {
					drag = track.DragOffset()
				}
				at := -(x + drag.X) / s.extent
				if p.Vertical {
					at = -(y + drag.Y) / s.extent
				} else if c.Direction() == ui.RightToLeft {
					at = -at
				}
				start = int(math.Floor(float64(at)))
			}
			track.Children(func() {
				for i := range n {
					key := any(i)
					if p.Key != nil {
						key = p.Key(i)
					}
					item := ui.Box(c.Key(key)).Shrink(0).MinWidth(0).Role(ui.RoleGroup).RoleDescription("slide").Label(fmt.Sprintf("Slide %d of %d", i+1, n)).AccessHidden(i != *selected).IgnoreInput(i != *selected)
					if i == *selected {
						item.Bind(&s.active)
					}
					if p.Loop {
						// Move a stable slide to its next occurrence only once it
						// has left the viewport, including during rapid loop wraps.
						position := start + ((i-start)%n+n)%n
						offset := s.extent * float32(position-i)
						if p.Vertical {
							item.Translate(0, offset)
						} else {
							if c.Direction() == ui.RightToLeft {
								offset = -offset
							}
							item.Translate(offset, 0)
						}
					}
					if p.Vertical {
						item.FillWidth().Height(height)
					} else {
						item.WidthPercent(100)
						if height > 0 {
							item.Height(height)
						}
					}
					item.Children(func() {
						if slide != nil {
							slide(i)
						}
					})
				}
			})
		})
		if !p.HideControls {
			ui.Row(c).Gap(12).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
				previous := ui.ButtonBase(c).Size(32, 32).Radius(999).Border(1, t.Border).Label("Previous slide").Disabled(p.Disabled || n < 2 || !p.Loop && *selected == 0).Children(func() {
					glyph := "←"
					if p.Vertical {
						glyph = "↑"
					} else if c.Direction() == ui.RightToLeft {
						glyph = "→"
					}
					ui.Text(c, glyph)
				})
				keys(previous)
				if previous.Clicked() {
					choose(*selected - 1)
				}
				ui.Row(c).Role(ui.RoleStatus).Key("slide-status").Label(fmt.Sprintf("Slide %d of %d", min(*selected+1, n), n)).Children(func() {
					ui.Text(c, fmt.Sprintf("%d / %d", min(*selected+1, n), n)).FontSize(12).TextColor(t.MutedForeground)
				})
				next := ui.ButtonBase(c).Size(32, 32).Radius(999).Border(1, t.Border).Label("Next slide").Disabled(p.Disabled || n < 2 || !p.Loop && *selected == n-1).Children(func() {
					glyph := "→"
					if p.Vertical {
						glyph = "↓"
					} else if c.Direction() == ui.RightToLeft {
						glyph = "←"
					}
					ui.Text(c, glyph)
				})
				keys(next)
				if next.Clicked() {
					choose(*selected + 1)
				}
			})
		}
		if p.Dots {
			ui.Row(c).Gap(8).Justify(ui.Center).Children(func() {
				for i := range n {
					dot := ui.ButtonBase(c.Key(i)).Size(12, 12).Radius(999).Label(fmt.Sprintf("Go to slide %d", i+1)).Disabled(p.Disabled).Current(ui.CurrentNone)
					if i == *selected {
						dot.Background(t.Primary).Current(ui.CurrentStep)
					} else {
						dot.Background(t.MutedForeground.Alpha(.4))
					}
					if dot.Clicked() {
						choose(i)
					}
				}
			})
		}
		if p.HideControls {
			ui.Row(c).Role(ui.RoleStatus).Key("slide-status").Label(fmt.Sprintf("Slide %d of %d", min(*selected+1, n), n)).Children(func() {
				ui.Text(c, fmt.Sprintf("%d / %d", min(*selected+1, n), n)).FontSize(12).TextColor(t.MutedForeground)
			})
		}
	})
	if s.extent > 0 {
		s.initialLayout = false
	}
	return root
}
