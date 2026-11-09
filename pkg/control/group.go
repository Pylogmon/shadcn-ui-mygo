package control

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type ButtonGroupProps struct {
	Label              string
	Vertical, Disabled bool
	// Gap separates nested or independent groups. Zero joins their borders.
	Gap float32
}
type groupItem struct {
	element   ui.Element
	separator bool
	nested    bool
}
type groupBuild struct {
	props ButtonGroupProps
	items []groupItem
}
type groupStack struct{ active []*groupBuild }

// groupScope is owned by this window's root, never by package-global state.
// The stack is present only while synchronous Children callbacks are building.
func groupScope(c *ui.Context) *groupStack {
	return ui.Local(c.Root(), "shadcn.control.button-groups", func() groupStack { return groupStack{} })
}

// ButtonGroup joins its buttons into one shape, preserving native focus and
// activation. Nested groups are separated with an 8-DIP gap. The optional props retain
// compatibility with ButtonGroup(c, content). Use ButtonGroupItem to include
// an input, select or custom native trigger in the same joined shape.
func ButtonGroup(c *ui.Context, content func(), props ...ButtonGroupProps) ui.Element {
	p := ButtonGroupProps{}
	if len(props) > 0 {
		p = props[0]
	}
	root := ui.Row(c).AlignItems(ui.Stretch).Gap(max(0, p.Gap)).Role(ui.RoleGroup).Label(p.Label).Disabled(p.Disabled)
	orientation := ui.Horizontal
	if p.Vertical {
		root.Column()
		orientation = ui.Vertical
	}
	root.FocusGroup(orientation).Orientation(orientation)
	parentScope := groupScope(c)
	if len(parentScope.active) > 0 {
		parent := parentScope.active[len(parentScope.active)-1]
		parent.items = append(parent.items, groupItem{element: root, nested: true})
	}
	stack := groupScope(c)
	current := &groupBuild{props: p}
	stack.active = append(stack.active, current)
	defer func() { last := len(stack.active) - 1; stack.active[last] = nil; stack.active = stack.active[:last] }()
	root.Children(content)
	nested := false
	for _, item := range current.items {
		nested = nested || item.nested
	}
	if p.Gap > 0 || nested {
		if p.Gap == 0 {
			root.Gap(8)
		}
		return root
	}
	radius := theme.TokensOf(c).RadiusMD()
	for i, item := range current.items {
		e := item.element
		if item.separator {
			continue
		}
		r := [4]float32{}
		if p.Vertical {
			if i == 0 {
				r[0], r[1] = radius, radius
			}
			if i == len(current.items)-1 {
				r[2], r[3] = radius, radius
			}
			if i > 0 && !current.items[i-1].separator {
				e.Margin(-1, 0, 0, 0)
			}
		} else {
			first, last := i == 0, i == len(current.items)-1
			if c.Direction() == ui.RightToLeft {
				first, last = last, first
			}
			if first {
				r[0], r[3] = radius, radius
			}
			if last {
				r[1], r[2] = radius, radius
			}
			if i > 0 && !current.items[i-1].separator {
				e.MarginStart(-1)
			}
		}
		e.Radius(r[0], r[1], r[2], r[3])
		if e.FocusWithin() {
			e.ZIndex(1)
		}
	}
	return root
}

// ButtonGroupItem registers a custom group child. Buttons register themselves.
// A repeated registration of the same element is harmless.
func ButtonGroupItem(c *ui.Context, e ui.Element) ui.Element {
	s := groupScope(c)
	if len(s.active) == 0 {
		return e
	}
	g := s.active[len(s.active)-1]
	for _, item := range g.items {
		if item.element == e {
			return e
		}
	}
	g.items = append(g.items, groupItem{element: e})
	return e
}

func ButtonGroupSeparator(c *ui.Context) ui.Element {
	s := groupScope(c)
	vertical := false
	if len(s.active) > 0 {
		vertical = s.active[len(s.active)-1].props.Vertical
	}
	e := ui.Box(c).Background(theme.TokensOf(c).Border).Shrink(0).AccessHidden(true)
	if vertical {
		e.Height(1).AlignSelf(ui.Stretch)
	} else {
		e.Width(1).AlignSelf(ui.Stretch)
	}
	if len(s.active) > 0 {
		g := s.active[len(s.active)-1]
		g.items = append(g.items, groupItem{element: e, separator: true})
	}
	return e
}

func ButtonGroupText(c *ui.Context, text string) ui.Element {
	t := theme.TokensOf(c)
	e := ui.Row(c).MinHeight(36).PaddingX(12).Border(1, t.Input).Background(t.Muted).TextColor(t.MutedForeground).AlignItems(ui.Center).Children(func() { ui.Text(c, text).FontSize(14) })
	return ButtonGroupItem(c, e)
}
