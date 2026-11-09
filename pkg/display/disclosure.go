package display

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type AccordionItem struct {
	Key, Title, Description string
	Open                    *bool
	// KeepMounted retains the panel after its first opening. Initially closed
	// content is built lazily, so a hidden editor's AutoFocus cannot run.
	Disabled, KeepMounted bool
	Content               func()
}

type AccordionProps struct {
	Items    []AccordionItem
	Single   bool
	Disabled bool
	// Bordered separates rounded items, as the upstream bordered example.
	Bordered bool
	// HeadingLevel defaults to 3. Loop optionally wraps Up/Down at the ends.
	HeadingLevel int
	Loop         bool
	OnChange     func(openKeys []string)
}

type accordionOpenState struct {
	Seen     bool
	Open     map[string]bool
	Triggers map[string]*ui.Handle
}

func accordionItemKey(it AccordionItem, i int) string {
	if it.Key != "" {
		return it.Key
	}
	return "item:" + strconv.Itoa(i)
}

// Accordion uses native disclosure inputs with stable, caller-owned values.
// Only enabled headers take Up/Down/Home/End; Tab retains normal document order.
func Accordion(c *ui.Context, p AccordionProps) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Column(c).FillWidth().Gap(0)
	s := ui.Local(root, "shadcn.accordion", func() accordionOpenState {
		return accordionOpenState{Open: make(map[string]bool), Triggers: make(map[string]*ui.Handle)}
	})
	if p.Single {
		chosen := -1
		for i, it := range p.Items {
			if it.Open != nil && *it.Open && (chosen < 0 || s.Seen && !s.Open[accordionItemKey(it, i)]) {
				chosen = i
				if s.Seen && !s.Open[accordionItemKey(it, i)] {
					break
				}
			}
		}
		for i, it := range p.Items {
			if it.Open != nil && i != chosen {
				*it.Open = false
			}
		}
	}
	clear(s.Open)
	var enabled []string
	last := -1
	for i, it := range p.Items {
		if it.Open == nil {
			continue
		}
		key := accordionItemKey(it, i)
		s.Open[key] = *it.Open
		if s.Triggers[key] == nil {
			s.Triggers[key] = &ui.Handle{}
		}
		if !p.Disabled && !it.Disabled {
			enabled = append(enabled, key)
		}
		last = i
	}
	for key := range s.Triggers {
		if _, exists := s.Open[key]; !exists {
			delete(s.Triggers, key)
		}
	}
	s.Seen = true
	level := p.HeadingLevel
	if level < 1 || level > 6 {
		level = 3
	}
	if p.Bordered {
		root.Gap(8)
	}
	changed := false
	root.Children(func() {
		for i, it := range p.Items {
			if it.Open == nil {
				continue
			}
			key := accordionItemKey(it, i)
			item := ui.Column(c.Key(key)).FillWidth().Shrink(0)
			if p.Bordered {
				item.Border(1, t.Border).Radius(t.RadiusLG())
			} else if i != last {
				item.InsetShadow(0, -1, 0, 0, t.Border)
			}
			item.Children(func() {
				var part ui.CollapsibleParts
				ui.Row(c).Role(ui.RoleHeading).Level(level).FillWidth().Children(func() {
					part = ui.CollapsibleBase(c, it.Open)
					tr := motion.ApplyTransition(c, part.Trigger.FillWidth().Padding(16, 0).Gap(16).AlignItems(ui.Start).
						Disabled(p.Disabled || it.Disabled).Label(it.Title).Description(it.Description).
						TextColor(t.Foreground).FontSize(14).FontWeight(500).Radius(t.RadiusMD()))
					tr.Bind(s.Triggers[key])
					if p.Bordered {
						tr.PaddingX(16)
					}
					if p.Disabled || it.Disabled {
						tr.Opacity(.5)
					} else {
						accordionKeys(tr, key, enabled, s.Triggers, p.Loop)
					}
					tr.Children(func() {
						text := ui.Text(c, it.Title).Grow(1).MinWidth(0)
						if tr.Hovered() {
							text.Underline().UnderlineOffset(4)
						}
						ui.Text(c, "⌄").FontSize(16).Shrink(0).Rotate(180 * part.Progress()).TextColor(t.MutedForeground).AccessHidden(true)
					})
					if tr.Changed() {
						// Native input is delivered as the trigger constructor
						// returns. Normalize again before reporting the user edit.
						if p.Single && *it.Open {
							for _, other := range p.Items {
								if other.Open != nil && other.Open != it.Open {
									*other.Open = false
								}
							}
						}
						changed = true
					}
				})
				panel := disclosurePanel(c, part, it.Open, it.KeepMounted, it.Title, it.Content)
				if panel.Valid() && p.Bordered {
					panel.PaddingX(16)
				}
			})
		}
	})
	if changed {
		root.SetChanged()
		if p.OnChange != nil {
			keys := make([]string, 0)
			for i, it := range p.Items {
				if it.Open != nil && *it.Open {
					keys = append(keys, accordionItemKey(it, i))
				}
			}
			p.OnChange(keys)
		}
	}
	return root
}

func accordionKeys(trigger ui.Element, key string, enabled []string, handles map[string]*ui.Handle, loop bool) {
	at := -1
	for i, id := range enabled {
		if id == key {
			at = i
			break
		}
	}
	if at < 0 {
		return
	}
	to := -1
	switch {
	case trigger.Shortcut(0, ui.KeyDown):
		to = at + 1
	case trigger.Shortcut(0, ui.KeyUp):
		to = at - 1
	case trigger.Shortcut(0, ui.KeyHome):
		to = 0
	case trigger.Shortcut(0, ui.KeyEnd):
		to = len(enabled) - 1
	default:
		return
	}
	if loop {
		to = (to + len(enabled)) % len(enabled)
	} else {
		to = max(0, min(to, len(enabled)-1))
	}
	handles[enabled[to]].Focus()
}

type CollapsibleProps struct {
	Label string
	// KeepMounted retains content after its first opening. Explicit Focus or
	// AutoFocus in retained content must be conditional on open: MyGo's direct
	// Focus currently bypasses disabled ancestors.
	Disabled, KeepMounted bool
	// Trigger styles/builds the native disclosure, allowing compound layouts.
	// Nil uses the default text-and-chevron trigger.
	Trigger      func(ui.Element)
	Content      func()
	OnOpenChange func(bool)
}

func Collapsible(c *ui.Context, label string, open *bool, content func()) ui.Element {
	return CollapsibleWith(c, open, CollapsibleProps{Label: label, Content: content})
}

// CollapsibleWith preserves local content when KeepMounted is requested.
// Closed/closing content is immediately removed from input and accessibility,
// while its pixels complete the reversible native disclosure animation.
func CollapsibleWith(c *ui.Context, open *bool, p CollapsibleProps) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Column(c).FillWidth().Shrink(0)
	root.Children(func() {
		part := ui.CollapsibleBase(c, open)
		tr := part.Trigger.Disabled(p.Disabled).Label(p.Label)
		if p.Trigger != nil {
			p.Trigger(tr)
		} else {
			tr.Padding(8).Radius(t.RadiusMD()).FillWidth().Children(func() {
				ui.Text(c, p.Label).Grow(1)
				ui.Text(c, "⌄").Rotate(180 * part.Progress()).AccessHidden(true)
			})
		}
		if p.Disabled {
			tr.Opacity(.5)
		}
		if tr.Changed() {
			root.SetChanged()
			if p.OnOpenChange != nil {
				p.OnOpenChange(*open)
			}
		}
		disclosurePanel(c, part, open, p.KeepMounted, p.Label, p.Content)
	})
	return root
}

func disclosurePanel(c *ui.Context, part ui.CollapsibleParts, open *bool, keep bool, label string, content func()) ui.Element {
	mounted := ui.Local(part.Trigger, "has-opened", func() bool { return false })
	if *open {
		*mounted = true
	}
	panel := ui.Local(part.Trigger, "controlled-panel", func() ui.Handle { return ui.Handle{} })
	if !*open && panel.FocusWithin(c) {
		part.Trigger.Focus()
	}
	progress := part.Progress()
	if !*mounted || !*open && progress <= 0 && !keep {
		return ui.Element{}
	}
	clip := ui.Column(c).FillWidth().Shrink(0).ClipY().Role(ui.RoleRegion).Label(label).
		AccessHidden(!*open).IgnoreInput(!*open).Disabled(!*open).DimDisabled(false)
	clip.Bind(panel)
	height := ui.Local(clip, "natural-height", func() float32 { return 0 })
	var inner ui.Element
	clip.Children(func() {
		inner = ui.Column(c).FillWidth().Shrink(0).Padding(0, 0, 16, 0).FontSize(14).Children(content)
		inner.Draw(func(_ *ui.Painter, r ui.Rect) {
			if *height != r.H {
				*height = r.H
				c.Invalidate()
			}
		})
	})
	part.Trigger.Controls(clip)
	if progress < 1 {
		clip.Height(max(0, progress) * *height)
	}
	return inner
}
