package display

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type AccordionItem struct {
	Key, Title string
	Open       *bool
	Disabled   bool
	Content    func()
}
type AccordionProps struct {
	Items  []AccordionItem
	Single bool
}
type accordionOpenState struct {
	Seen bool
	Open map[*bool]bool
}

func Accordion(c *ui.Context, p AccordionProps) ui.Element {
	t := theme.TokensOf(c)
	acc := ui.AccordionBase(c)
	previous := ui.Local(acc.Root, "open-state", func() accordionOpenState { return accordionOpenState{Open: make(map[*bool]bool)} })
	if p.Single {
		var chosen *bool
		for _, item := range p.Items {
			if item.Open != nil && *item.Open {
				chosen = item.Open
				break
			}
		}
		// Prefer a newly opened section; otherwise keep the first open one.
		// Normalize before rendering so earlier panels close in the same frame.
		if previous.Seen {
			for _, item := range p.Items {
				if item.Open != nil && *item.Open && !previous.Open[item.Open] {
					chosen = item.Open
					break
				}
			}
		}
		for _, item := range p.Items {
			if item.Open != nil && item.Open != chosen {
				*item.Open = false
			}
		}
	}
	clear(previous.Open)
	for _, item := range p.Items {
		if item.Open != nil {
			previous.Open[item.Open] = *item.Open
		}
	}
	previous.Seen = true
	return acc.Root.FillWidth().Children(func() {
		for _, it := range p.Items {
			if it.Open == nil {
				continue
			}
			ui.Column(c.Key(it.Key)).FillWidth().Border(1, ui.Transparent).InsetShadow(0, -1, 0, 0, t.Border).Children(func() {
				part := acc.Item(it.Open)
				tr := part.Trigger.FillWidth().Padding(16, 0).Gap(16).Disabled(it.Disabled).TextColor(t.Foreground).FontSize(14).FontWeight(500).Transition(motion.Transition(c))
				if tr.Hovered() {
					tr.TextColor(t.MutedForeground)
				}
				tr.Children(func() {
					ui.Text(c, it.Title).Grow(1)
					ui.Text(c, "⌄").Rotate(180 * part.Progress()).TextColor(t.MutedForeground)
				})
				part.Panel(func() {
					if it.Content != nil {
						it.Content()
					}
				}).Padding(0, 0, 16, 0).FontSize(14)
			})
		}
	})
}
func Collapsible(c *ui.Context, label string, open *bool, content func()) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).FillWidth().Children(func() {
		p := ui.CollapsibleBase(c, open)
		p.Trigger.Padding(8).Radius(t.Radius).FillWidth().Children(func() { ui.Text(c, label).Grow(1); ui.Text(c, "⌄").Rotate(180 * p.Progress()) })
		p.Panel(content).Padding(8).FontSize(14)
	})
}

// Carousel keeps the selected slide in caller-owned state. Previous and next
// remain native, keyboard-operable buttons with descriptive names.
func Carousel(c *ui.Context, selected *int, count int, slide func(int)) ui.Element {
	t := theme.TokensOf(c)
	n := max(count, 0)
	if n > 0 {
		*selected = max(0, min(*selected, n-1))
	} else {
		*selected = 0
	}
	return ui.Column(c).FillWidth().Gap(12).Children(func() {
		ui.Box(c.Key(*selected)).FillWidth().Children(func() {
			if n > 0 && slide != nil {
				slide(*selected)
			}
		})
		ui.Row(c).Gap(12).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			b := ui.ButtonBase(c).Size(32, 32).Radius(999).Border(1, t.Border).Label("Previous slide").Disabled(*selected <= 0).Children(func() { ui.Text(c, "←") })
			if b.Clicked() {
				*selected--
			}
			ui.Text(c, fmt.Sprintf("%d / %d", min(*selected+1, n), n)).FontSize(12).TextColor(t.MutedForeground)
			b = ui.ButtonBase(c).Size(32, 32).Radius(999).Border(1, t.Border).Label("Next slide").Disabled(n == 0 || *selected >= n-1).Children(func() { ui.Text(c, "→") })
			if b.Clicked() {
				*selected++
			}
		})
	})
}
func ScrollArea(c *ui.Context, content func()) ui.Element {
	return ui.Scroll(c).Fill().Children(content)
}

type ResizableProps struct {
	Vertical bool
	Min, Max float32
}

func Resizable(c *ui.Context, size *float32, p ResizableProps, first, second func()) ui.Element {
	if p.Min > 0 {
		*size = max(p.Min, *size)
	}
	if p.Max > 0 {
		*size = min(p.Max, *size)
	}
	if p.Vertical {
		return ui.SplitVertical(c, size, first, second)
	}
	return ui.Split(c, size, first, second)
}
func Direction(c *ui.Context, dir ui.Direction, content func()) ui.Element {
	return ui.Column(c).Direction(dir).Children(content)
}
