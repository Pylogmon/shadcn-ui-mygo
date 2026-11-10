package navigation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// TabsRootProps configures one independently composed single-selection group.
// Value is application-owned; initialize it to the desired source defaultValue.
// An unmatched value can address a panel without a trigger. Empty selects none.
type TabsRootProps struct {
	Key, Label               string
	Vertical, Manual, NoLoop bool
	OnValueChange            func(string)
}
type TabsListProps struct {
	Label   string
	Variant TabsVariant
}
type TabsTriggerProps struct {
	Key, Label          string
	Icon                *ui.SVG
	Disabled, HideLabel bool
	// Content replaces the default icon/text drawing. Label remains its native
	// accessible name; the callback is used only during the current build.
	Content func(*ui.Context)
}
type TabsContentProps struct{ Label string }

type tabsCompositionScopeKey struct{}
type tabsCompositionScope struct{ build *tabsCompositionBuild }
type tabsCompositionRootKey string
type tabsCompositionTriggerKey string
type tabsCompositionContentKey string
type tabsCompositionIndicatorKey struct{ root ui.Handle }
type tabsCompositionState struct{ root ui.Handle }
type tabsCompositionTrigger struct {
	value    string
	disabled bool
	element  ui.Element
}
type tabsCompositionPanel struct {
	value   string
	element ui.Element
}
type tabsCompositionBuild struct {
	root              ui.Element
	state             *tabsCompositionState
	value             *string
	props             TabsRootProps
	initial           string
	listBuilt, inList bool
	listVariant       TabsVariant
	triggers          []tabsCompositionTrigger
	panels            []tabsCompositionPanel
	values            map[string]bool
	panelValues       map[string]bool
	focus, step       int
	userChanged       bool
}

func currentTabs(c *ui.Context) *tabsCompositionBuild {
	scope := ui.Local(c.Root(), tabsCompositionScopeKey{}, func() tabsCompositionScope { return tabsCompositionScope{} })
	if scope.build == nil {
		panic("navigation: Tabs slots require TabsRoot Children")
	}
	return scope.build
}

// TabsRoot provides a synchronous composition scope. Put TabsList beside Input,
// Button or Menu in any header, and place TabsContent independently below or
// beside that header. Slots must run inside this callback; replacing Children
// on the returned Element does not reinstall the scope. Nested/sibling Roots
// restore their previous scope. No Context/Element survives a frame.
func TabsRoot(c *ui.Context, value *string, p TabsRootProps, children func(*ui.Context)) ui.Element {
	if value == nil {
		panic("navigation: TabsRoot requires selected value state")
	}
	root := ui.Column(c).Gap(8).MinWidth(0)
	if p.Key != "" {
		root.Key(tabsCompositionRootKey(p.Key))
	}
	if p.Vertical {
		root.Row().AlignItems(ui.Stretch)
	}
	state := ui.Local(root, "shadcn.tabs.composition", func() tabsCompositionState { return tabsCompositionState{} })
	root.Bind(&state.root)
	build := &tabsCompositionBuild{root: root, state: state, value: value, props: p, initial: *value, values: map[string]bool{}, panelValues: map[string]bool{}, focus: -1}
	root.Children(func() {
		scope := ui.Local(c.Root(), tabsCompositionScopeKey{}, func() tabsCompositionScope { return tabsCompositionScope{} })
		previous := *scope
		scope.build = build
		defer func() { *scope = previous }()
		if children != nil {
			children(c)
		}
		build.finish(c)
	})
	return root
}
func TabsList(c *ui.Context, p TabsListProps, children func(*ui.Context)) ui.Element {
	b := currentTabs(c)
	if b.listBuilt {
		panic("navigation: TabsRoot accepts one TabsList")
	}
	b.listBuilt = true
	b.listVariant = p.Variant
	t := theme.TokensOf(c)
	orientation := ui.Horizontal
	if b.props.Vertical {
		orientation = ui.Vertical
	}
	label := p.Label
	if label == "" {
		label = b.props.Label
	}
	list := ui.Row(c).Shrink(0).Padding(3).Gap(2).Radius(t.Radius).Background(t.Muted).TextColor(t.MutedForeground).Role(ui.RoleTabList).Label(label).Orientation(orientation).FocusGroup(orientation).AlignSelf(ui.Start)
	if b.props.Vertical {
		list.Column().AlignItems(ui.Stretch)
	}
	if p.Variant == TabsLine {
		list.Background(ui.Transparent).Padding(0).Gap(8).Radius(0)
	}
	list.Children(func() {
		b.inList = true
		defer func() { b.inList = false }()
		if children != nil {
			children(c)
		}
	})
	return list
}

// TabsTrigger returns the real native tab button; Bind may expose its Handle.
// HideLabel hides visual text only. It preserves Label for icon-only native AX,
// improving on the fixed source example's unlabeled icon-only buttons.
func TabsTrigger(c *ui.Context, value string, p TabsTriggerProps) ui.Element {
	b := currentTabs(c)
	if !b.inList {
		panic("navigation: TabsTrigger requires TabsList Children")
	}
	if value == "" || b.values[value] {
		panic("navigation: TabsTrigger requires a unique nonempty value")
	}
	b.values[value] = true
	label := p.Label
	if label == "" {
		label = value
	}
	key := value
	if p.Key != "" {
		key = p.Key
	}
	t := theme.TokensOf(c)
	tab := motion.ApplyTransition(c, ui.ButtonBase(c.Key(tabsCompositionTriggerKey(key))).Role(ui.RoleTab).Label(label).Disabled(p.Disabled).Checked(value == *b.value).Height(30).PaddingX(12).Gap(6).Radius(t.RadiusMD()).FontSize(14).FontWeight(500).TextColor(t.MutedForeground))
	if b.props.Vertical {
		tab.Justify(ui.Start)
	}
	index := len(b.triggers)
	b.triggers = append(b.triggers, tabsCompositionTrigger{value: value, disabled: p.Disabled, element: tab})
	if !p.Disabled {
		back, forth := ui.KeyLeft, ui.KeyRight
		if c.Direction() == ui.RightToLeft {
			back, forth = forth, back
		}
		if b.props.Vertical {
			back, forth = ui.KeyUp, ui.KeyDown
		}
		switch {
		case tab.Shortcut(0, back):
			b.focus, b.step = index, -1
		case tab.Shortcut(0, forth):
			b.focus, b.step = index, 1
		case tab.Shortcut(0, ui.KeyHome):
			b.focus, b.step = -2, 1
		case tab.Shortcut(0, ui.KeyEnd):
			b.focus, b.step = -3, -1
		}
		if tab.Clicked() {
			*b.value = value
			b.userChanged = true
		}
	} else {
		tab.DimDisabled(false).Opacity(.5)
	}
	tab.Children(func() {
		if value == *b.value {
			indicator := ui.Box(c).Absolute().ZIndex(-1).IgnoreInput(true)
			if !motion.Reduced(c) {
				transition := motion.Transition(c)
				transition.Match = tabsCompositionIndicatorKey{b.state.root}
				indicator.Transition(transition)
			}
			if b.listVariant == TabsLine {
				if b.props.Vertical {
					indicator.Right(0).Top(0).Bottom(0).Width(2)
				} else {
					indicator.Left(0).Right(0).Bottom(0).Height(2)
				}
				indicator.Background(t.Foreground)
			} else {
				indicator.Top(0).Bottom(0).Left(0).Right(0).Radius(t.RadiusMD()).Background(t.Background).Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, .06))
				if t.Dark {
					indicator.Background(t.Input.Alpha(.3)).Border(1, t.Input)
				}
			}
			tab.TextColor(t.Foreground)
		}
		if p.Content != nil {
			p.Content(c)
		} else {
			if p.Icon != nil {
				ui.Icon(c, p.Icon).Size(16, 16)
			}
			if !p.HideLabel {
				ui.Text(c, label).SingleLine()
			}
		}
	})
	return tab
}

// TabsContent builds only the active panel. Application models retain drafts;
// inactive child closures are not called, so hidden controls cannot steal focus.
// The real panel and trigger are linked after all slots have been collected,
// including when a panel is composed before its list. Inactive returns zero.
func TabsContent(c *ui.Context, value string, p TabsContentProps, children func(*ui.Context)) ui.Element {
	b := currentTabs(c)
	if b.inList {
		panic("navigation: TabsContent belongs outside TabsList")
	}
	if b.panelValues[value] {
		panic("navigation: duplicate TabsContent value")
	}
	b.panelValues[value] = true
	if value == "" || value != *b.value {
		return ui.Element{}
	}
	label := p.Label
	if label == "" {
		label = value
	}
	region := ui.Column(c.Key(tabsCompositionContentKey(value))).Grow(1).MinWidth(0).Role(ui.RoleRegion).RoleDescription("tab panel").Label(label).Focusable()
	b.panels = append(b.panels, tabsCompositionPanel{value: value, element: region})
	region.Children(func() {
		if children != nil {
			children(c)
		}
	})
	return region
}
func (b *tabsCompositionBuild) finish(c *ui.Context) {
	if b.step != 0 {
		items := make([]TabItem, len(b.triggers))
		for i, t := range b.triggers {
			items[i].Disabled = t.disabled
		}
		from := b.focus
		loop := !b.props.NoLoop
		if from == -2 {
			from = -1
			loop = false
		}
		if from == -3 {
			from = len(items)
			loop = false
		}
		target := nextTab(items, from, b.step, loop)
		if target >= 0 && target < len(b.triggers) {
			b.triggers[target].element.Focus()
			if !b.props.Manual {
				*b.value = b.triggers[target].value
				b.userChanged = true
			}
		}
	}
	for _, trigger := range b.triggers {
		// SDK LabelledBy also activates a caption click. A tab is the keyboard
		// target itself, so retain its native focus after publishing that link.
		retainFocus := trigger.element.Focused() || trigger.element.Clicked()
		trigger.element.Checked(trigger.value == *b.value)
		for _, panel := range b.panels {
			if panel.value == trigger.value {
				trigger.element.Controls(panel.element)
				panel.element.LabelledBy(trigger.element)
			}
		}
		if retainFocus && !trigger.disabled {
			trigger.element.Focus()
		}
	}
	if *b.value != b.initial {
		c.Invalidate()
		if b.userChanged {
			b.root.SetChanged()
			if b.props.OnValueChange != nil {
				b.props.OnValueChange(*b.value)
			}
		}
	}
}

// The source's TabsMultiple example is four ordinary single-selection triggers.
// This composition intentionally does not add a multi-select tab state.
