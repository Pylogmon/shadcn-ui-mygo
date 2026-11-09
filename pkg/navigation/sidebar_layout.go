package navigation

import (
	"errors"
	"math"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type SidebarVariant uint8

const (
	SidebarStandard SidebarVariant = iota
	SidebarFloating
	SidebarInset
)

type SidebarCollapse uint8

const (
	SidebarOffcanvas SidebarCollapse = iota
	SidebarIcon
	SidebarFixed
)

type SidebarSide uint8

const (
	SidebarStart SidebarSide = iota
	SidebarEnd
	SidebarLeft
	SidebarRight
)

type SidebarItemSize uint8

const (
	SidebarItemDefault SidebarItemSize = iota
	SidebarItemSmall
	SidebarItemLarge
)

type SidebarItem struct {
	Key, Label, Description, Group, Tooltip, Badge string
	Icon                                           *ui.SVG
	Disabled, DefaultOpen, Outline                 bool
	Size                                           SidebarItemSize
	Children                                       []SidebarItem
	// Action is an independent sibling of the route/disclosure control.
	Action func(*ui.Context)
}

// SidebarState is caller-owned and JSON-serializable. A state belongs to one
// SidebarWith instance. Desktop collapse and mobile visibility are independent.
type SidebarState struct {
	Collapsed                 bool
	MobileOpen                bool
	Groups                    map[string]bool
	mobile                    bool
	available                 float32
	restore                   *ui.Handle
	desktopPanel, mobilePanel ui.Handle
	mainPanel                 ui.Handle
	triggers                  []*ui.Handle
	pendingTrigger            *ui.Handle
	dirty, toggleDirty        bool
	desktopVisited            bool
	mobileActive              bool
}

func (s *SidebarState) IsMobile() bool { return s != nil && s.mobile }
func (s *SidebarState) Toggle() {
	if s == nil {
		return
	}
	if s.mobile {
		s.MobileOpen = !s.MobileOpen
	} else {
		s.Collapsed = !s.Collapsed
	}
	s.dirty, s.toggleDirty = true, true
}

type SidebarLayoutProps struct {
	Items                                     []SidebarItem
	Label, ContentLabel                       string
	Variant                                   SidebarVariant
	Collapse                                  SidebarCollapse
	Side                                      SidebarSide
	Width, IconWidth, MobileWidth, Breakpoint float32
	// AvailableWidth explicitly defines the responsive container width. Zero
	// uses this layout's measured bounds, with window width only before layout.
	AvailableWidth float32
	// Mobile overrides breakpoint detection when non-nil, e.g. an embedded demo.
	Mobile                            *bool
	HideTrigger, HideRail, NoShortcut bool
	Header, Footer                    func(*ui.Context)
	OnSelect                          func(string)
	OnToggle                          func(collapsed, mobileOpen bool)
	OnGroupChange                     func(key string, open bool)
}

// SidebarTrigger toggles the instance using its current responsive mode. It is
// a real, accessible native button; any number of caller-built triggers may exist.
func SidebarTrigger(c *ui.Context, state *SidebarState) ui.Element {
	t := theme.TokensOf(c)
	button := ui.ButtonBase(c).Label("Toggle sidebar").Size(28, 28).Radius(t.RadiusMD()).Center().TextColor(t.Foreground)
	handle := ui.Local(button, "sidebar.trigger", func() ui.Handle { return ui.Handle{} })
	button.Bind(handle)
	if state != nil {
		known := false
		for _, registered := range state.triggers {
			known = known || registered == handle
		}
		if !known {
			state.triggers = append(state.triggers, handle)
		}
		if state.restore == nil {
			state.restore = handle
		}
		sidebarConfigureTrigger(c, state, button)
		if button.Clicked() {
			state.restore = handle
			if c.Resolve(state.mainPanel).Valid() {
				state.Toggle()
			} else {
				// A preceding trigger waits for this pass's responsive mode.
				state.pendingTrigger = handle
			}
			c.Invalidate()
		}
	}
	if button.Hovered() {
		button.Background(t.Accent).TextColor(t.AccentForeground)
	}
	button.Children(func() { ui.Text(c, "☰").FontSize(16).AccessHidden(true) })
	return button
}

func sidebarConfigureTrigger(c *ui.Context, state *SidebarState, button ui.Element) {
	button.Expanded(state.MobileOpen && state.mobile || !state.Collapsed && !state.mobile).HasPopup(ui.PopupNone)
	panel := c.Resolve(state.desktopPanel)
	if state.mobile {
		button.HasPopup(ui.PopupDialog)
		panel = c.Resolve(state.mobilePanel)
	}
	if panel.Valid() {
		button.Controls(panel)
	}
}

func sidebarFinite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}
func sidebarLeft(side SidebarSide, direction ui.Direction) bool {
	switch side {
	case SidebarLeft:
		return true
	case SidebarRight:
		return false
	case SidebarEnd:
		return direction == ui.RightToLeft
	default:
		return direction != ui.RightToLeft
	}
}
func sidebarValidate(items []SidebarItem, keys map[string]bool) bool {
	for _, item := range items {
		if item.Key == "" || keys[item.Key] {
			return false
		}
		keys[item.Key] = true
		if !sidebarValidate(item.Children, keys) {
			return false
		}
	}
	return true
}
func sidebarInitGroups(s *SidebarState, items []SidebarItem) {
	if s.Groups == nil {
		s.Groups = map[string]bool{}
	}
	for _, item := range items {
		if len(item.Children) > 0 {
			if _, ok := s.Groups[item.Key]; !ok {
				s.Groups[item.Key] = item.DefaultOpen
			}
			sidebarInitGroups(s, item.Children)
		}
	}
}

// SidebarWith composes the sidebar and its main/inset area. On small layouts it
// replaces the desktop pane with an in-window modal Sheet. Header, Footer and
// main build in the current Context; tooltips use an explicit independent Host.
func SidebarWith(c *ui.Context, host *desktop.Host, key string, selected *string, state *SidebarState, p SidebarLayoutProps, main func(*ui.Context)) (ui.Element, error) {
	if c == nil || !c.Valid() || key == "" || selected == nil || state == nil {
		return ui.Element{}, errors.New("navigation: SidebarWith needs a live context, stable key, route and state")
	}
	root := ui.Row(c.Key("sidebar-layout:" + key)).Fill().MinWidth(0).MinHeight(0).Clip()
	for _, n := range []float32{p.Width, p.IconWidth, p.MobileWidth, p.Breakpoint, p.AvailableWidth} {
		if !sidebarFinite(n) {
			return root.Invalid(true), errors.New("navigation: sidebar dimensions must be finite")
		}
	}
	if !sidebarValidate(p.Items, map[string]bool{}) {
		return root.Invalid(true), errors.New("navigation: sidebar items require unique nonempty keys")
	}
	sidebarInitGroups(state, p.Items)
	beforeMobile := state.MobileOpen
	width, iconWidth, mobileWidth, breakpoint := p.Width, p.IconWidth, p.MobileWidth, p.Breakpoint
	if width <= 0 {
		width = 256
	}
	if iconWidth <= 0 {
		iconWidth = 48
	}
	if mobileWidth <= 0 {
		mobileWidth = 288
	}
	if breakpoint <= 0 {
		breakpoint = 768
	}
	iconWidth = min(width, iconWidth)
	available := p.AvailableWidth
	if available <= 0 {
		available = state.available
	}
	if available <= 0 {
		available, _ = c.Size()
	}
	mobile := available < breakpoint && p.Collapse != SidebarFixed
	if p.Mobile != nil && p.Collapse != SidebarFixed {
		mobile = *p.Mobile
	}
	if !mobile && state.mobile {
		state.MobileOpen = false
	}
	state.mobile = mobile
	if state.pendingTrigger != nil {
		state.restore = state.pendingTrigger
		state.pendingTrigger = nil
		state.Toggle()
	}
	dir := c.Direction()
	left := sidebarLeft(p.Side, dir)
	t := theme.TokensOf(c)
	label := p.Label
	if label == "" {
		label = "Sidebar"
	}
	contentLabel := p.ContentLabel
	if contentLabel == "" {
		contentLabel = "Content"
	}
	root.Direction(ui.LeftToRight)
	if p.Variant == SidebarInset {
		root.Background(t.Sidebar)
	}
	if !p.NoShortcut && p.Collapse != SidebarFixed && (root.Shortcut(ui.Cmd, ui.KeyB) || root.Shortcut(ui.Ctrl, ui.KeyB)) {
		state.Toggle()
		c.Invalidate()
	}
	collapsed := state.Collapsed && p.Collapse != SidebarFixed
	icon := collapsed && p.Collapse == SidebarIcon && !mobile
	off := collapsed && p.Collapse == SidebarOffcanvas && !mobile
	padding := float32(0)
	if p.Variant != SidebarStandard {
		padding = min(8, width/4)
	}
	iconWidth = min(iconWidth, max(1, width-padding*2))
	gapWidth := width
	if icon {
		gapWidth = iconWidth + padding*2
	}
	if off || mobile {
		gapWidth = 0
	}
	duration := 200 * time.Millisecond
	if motion.Reduced(c) {
		duration = -1
	}
	animate := func(key string, target float32) float32 {
		if duration < 0 {
			return root.Jump(key, target)
		}
		return root.AnimateWith(key, target, duration, ui.Linear)
	}
	gapWidth = animate("sidebar-gap", gapWidth)
	panelWidth := max(1, width-padding*2)
	if icon {
		panelWidth = iconWidth
	}
	panelWidth = animate("sidebar-width", panelWidth)
	translation := float32(0)
	if off {
		translation = width
		if left {
			translation = -translation
		}
	}
	translation = animate("sidebar-offset", translation)
	insetNear := float32(0)
	if collapsed {
		insetNear = 8
	}
	insetNear = animate("sidebar-inset-near", insetNear)
	restoreDesktop := off && state.desktopPanel.FocusWithin(c)
	var popupErr error
	restoreMobile := false
	buildPane := func() {
		if mobile {
			return
		}
		slot := ui.Box(c.Key("desktop-slot")).Width(gapWidth).FillHeight().Shrink(0)
		slot.Children(func() {
			pane := ui.Column(c.Key("desktop-panel")).Absolute().Left(padding).Top(padding).Bottom(padding).Width(panelWidth).Clip().
				Direction(dir).Background(t.Sidebar).TextColor(t.SidebarForeground).Role(ui.RoleNavigation).Label(label).
				AccessHidden(off).IgnoreInput(off).Disabled(off).DimDisabled(false).Translate(translation, 0).Bind(&state.desktopPanel)
			if p.Variant == SidebarFloating {
				pane.Border(1, t.SidebarBorder).Radius(t.RadiusLG()).Shadow(0, 1, 3, 0, ui.RGBA(0, 0, 0, .1))
			}
			if p.Variant == SidebarStandard {
				if left {
					pane.BorderWidth(0, 1, 0, 0)
				} else {
					pane.BorderWidth(0, 0, 0, 1)
				}
				pane.BorderColor(t.SidebarBorder)
			}
			if !off {
				state.desktopVisited = true
			}
			pane.Children(func() {
				if state.desktopVisited {
					sidebarBody(c, host, key, state, selected, p, icon, left, &popupErr)
				}
			})
		})
	}
	buildMain := func() {
		body := ui.Column(c.Key("main")).Grow(1).MinWidth(0).MinHeight(0).FillHeight().Direction(dir).Background(t.Background).TextColor(t.Foreground).Role(ui.RoleRegion).Label(contentLabel).Focusable().TabStop(false).Bind(&state.mainPanel)
		if p.Variant == SidebarInset && !mobile {
			body.Margin(8).Radius(t.RadiusXL()).Shadow(0, 1, 3, 0, ui.RGBA(0, 0, 0, .1))
			if left {
				body.Margin(8, 8, 8, insetNear)
			} else {
				body.Margin(8, insetNear, 8, 8)
			}
		}
		body.Children(func() {
			if !p.HideTrigger && p.Collapse != SidebarFixed {
				ui.Row(c).Height(48).PaddingX(16).AlignItems(ui.Center).Children(func() { SidebarTrigger(c, state) })
			}
			ui.Column(c).Grow(1).MinHeight(0).FillWidth().Children(func() {
				if main != nil {
					main(c)
				}
			})
		})
	}
	root.Children(func() {
		if left {
			buildPane()
			buildMain()
		} else {
			buildMain()
			buildPane()
		}
		// Draw the rail after main, so its hover line remains above the inset
		// surface even when offcanvas collapse has reduced the sidebar gap to 0.
		if !mobile && !p.HideRail && p.Collapse != SidebarFixed {
			rail := ui.ButtonBase(c.Key("rail")).Label("Toggle sidebar rail").TabStop(false).Absolute().Top(0).Bottom(0).Width(8).Cursor(ui.CursorResizeEW)
			if left {
				rail.Left(gapWidth - 4)
			} else {
				rail.Right(gapWidth - 4)
			}
			if rail.Hovered() {
				rail.Background(t.SidebarBorder)
			}
			if rail.Clicked() {
				state.Toggle()
				c.Invalidate()
			}
		}
		ui.Box(c.Key("measure")).Absolute().Fill().IgnoreInput(true).AccessHidden(true).Draw(func(_ *ui.Painter, r ui.Rect) {
			// Resolve after the whole view is built: external triggers can come
			// before or after this layout and can disappear conditionally.
			restoreValid := false
			if state.restore != nil {
				e := c.Resolve(*state.restore)
				restoreValid = e.Valid() && !e.IsDisabled()
			}
			if !restoreValid || state.restore == &state.mainPanel {
				previous := state.restore
				state.restore = &state.mainPanel
				for _, trigger := range state.triggers {
					if e := c.Resolve(*trigger); e.Valid() && !e.IsDisabled() {
						state.restore = trigger
						break
					}
				}
				if previous != nil && previous != state.restore {
					previous.CancelFocus(c)
				}
			}
			if restoreDesktop || restoreMobile {
				state.restore.Focus(c)
				// Draw follows MyGo's focus-request commit. Schedule one build
				// to apply this stable-handle request, then the flags settle.
				c.Invalidate()
			}
			if state.available != r.W {
				state.available = r.W
				if p.AvailableWidth <= 0 && p.Mobile == nil {
					c.Invalidate()
				}
			}
		})
	})
	if !mobile {
		state.MobileOpen = false
	}
	zero := float32(0)
	modal := overlay.Sheet(c.Key("sidebar-mobile:"+key), &state.MobileOpen, overlay.SheetProps{Title: label, Description: "Mobile navigation", HideHeader: true,
		Side: map[bool]overlay.Placement{true: overlay.Left, false: overlay.Right}[left], Size: mobileWidth, HideClose: false,
		ContentPadding: &zero, ContentGap: &zero}, func() {
		ui.Column(c.Key("mobile-panel")).Fill().Clip().Direction(dir).Background(t.Sidebar).TextColor(t.SidebarForeground).Role(ui.RoleNavigation).Label(label).Bind(&state.mobilePanel).
			Children(func() { sidebarBody(c, host, key, state, selected, p, false, left, &popupErr) })
	})
	if modal.Valid() {
		state.mobileActive = true
		modal.Padding(0).Gap(0).Label(label).Description("Mobile navigation").Background(t.Sidebar).TextColor(t.SidebarForeground)
	} else if state.mobileActive {
		state.mobileActive = false
		restoreMobile = true
	}
	currentTriggers := state.triggers[:0]
	for _, handle := range state.triggers {
		if button := c.Resolve(*handle); button.Valid() {
			currentTriggers = append(currentTriggers, handle)
			sidebarConfigureTrigger(c, state, button)
		}
	}
	state.triggers = currentTriggers
	if beforeMobile != state.MobileOpen {
		state.dirty, state.toggleDirty = true, true
	}
	if state.dirty {
		root.SetChanged()
		state.dirty = false
		c.Invalidate()
	}
	if state.toggleDirty {
		state.toggleDirty = false
		if p.OnToggle != nil {
			p.OnToggle(state.Collapsed, state.MobileOpen)
		}
	}
	return root, popupErr
}

func sidebarBody(c *ui.Context, host *desktop.Host, key string, s *SidebarState, selected *string, p SidebarLayoutProps, icon, left bool, popupErr *error) {
	if p.Header != nil {
		ui.Column(c.Key("header")).Padding(8).Gap(8).Children(func() { p.Header(c) })
	}
	ui.Scroll(c.Key("items")).Grow(1).MinHeight(0).Gap(4).Padding(8).Children(func() { sidebarTree(c, host, key, s, selected, p, p.Items, icon, left, false, 0, popupErr) })
	if p.Footer != nil {
		ui.Column(c.Key("footer")).Padding(8).Gap(8).Children(func() { p.Footer(c) })
	}
}

func sidebarTree(c *ui.Context, host *desktop.Host, key string, s *SidebarState, selected *string, p SidebarLayoutProps, items []SidebarItem, icon, left, disabled bool, depth int, popupErr *error) {
	t := theme.TokensOf(c)
	group := ""
	for _, item := range items {
		if item.Group != "" && item.Group != group && !icon {
			group = item.Group
			ui.Text(c, group).Height(32).PaddingX(8).FontSize(12).FontWeight(500).TextColor(t.SidebarForeground.Alpha(.7))
		}
		itemDisabled := disabled || item.Disabled
		container := ui.Column(c.Key(item.Key)).FillWidth().Shrink(0).Gap(2)
		container.Children(func() {
			var part ui.CollapsibleParts
			var trigger ui.Element
			var open *bool
			ui.Row(c).FillWidth().Gap(4).Children(func() {
				if len(item.Children) > 0 {
					open = ui.Local(container, "expanded", func() bool { return s.Groups[item.Key] })
					*open = s.Groups[item.Key]
					part = ui.CollapsibleBase(c, open)
					trigger = part.Trigger
				} else {
					trigger = ui.ButtonBase(c).Role(ui.RoleLink)
				}
				motion.ApplyTransition(c, trigger.Label(item.Label).Description(item.Description).Disabled(itemDisabled).FillWidth().Grow(1).MinWidth(0).Height(32).Padding(8).Gap(8).Radius(t.RadiusMD()).
					TextColor(t.SidebarForeground).FocusRingColor(t.SidebarRing))
				if item.Size == SidebarItemSmall {
					trigger.Height(28).FontSize(12)
				} else {
					trigger.FontSize(14)
				}
				if item.Size == SidebarItemLarge {
					trigger.Height(48)
				}
				if item.Outline {
					trigger.Background(t.Background).Border(1, t.SidebarBorder)
				}
				if depth > 0 && !icon {
					trigger.MarginStart(8)
				}
				if icon {
					trigger.Size(32, 32).Grow(0).Padding(8)
				}
				active := item.Key == *selected
				if active {
					trigger.Current(ui.CurrentPage).FontWeight(500)
				}
				if trigger.Hovered() || active {
					trigger.Background(t.SidebarAccent).TextColor(t.SidebarAccentForeground)
				}
				trigger.Children(func() {
					if item.Icon != nil {
						ui.Icon(c, item.Icon).Size(16, 16).Shrink(0).AccessHidden(true)
					} else if icon {
						ui.Text(c, initial(item.Label)).Size(16, 16).AccessHidden(true)
					}
					if !icon {
						ui.Text(c, item.Label).Grow(1).MinWidth(0).SingleLine()
						if item.Badge != "" {
							ui.Text(c, item.Badge).FontSize(12)
						}
						if len(item.Children) > 0 {
							angle := 90 * part.Progress()
							if c.Direction() == ui.RightToLeft {
								angle = 180 - angle
							}
							ui.Text(c, "›").Rotate(angle).AccessHidden(true)
						}
					}
				})
				if !itemDisabled {
					if len(item.Children) > 0 {
						expand, collapse := ui.KeyRight, ui.KeyLeft
						if c.Direction() == ui.RightToLeft {
							expand, collapse = collapse, expand
						}
						if trigger.Shortcut(0, expand) {
							*open = true
						}
						if trigger.Shortcut(0, collapse) {
							*open = false
						}
						if trigger.Changed() || *open != s.Groups[item.Key] {
							s.Groups[item.Key] = *open
							s.dirty = true
							c.Invalidate()
							if p.OnGroupChange != nil {
								p.OnGroupChange(item.Key, *open)
							}
							if icon {
								s.Collapsed = false
								s.toggleDirty = true
							}
						}
					} else if trigger.Clicked() {
						if *selected != item.Key {
							*selected = item.Key
							s.dirty = true
							if p.OnSelect != nil {
								p.OnSelect(item.Key)
							}
						}
						if s.mobile {
							s.MobileOpen = false
							s.toggleDirty = true
							s.dirty = true
						}
					}
				}
				if icon {
					text := item.Tooltip
					if text == "" {
						text = item.Label
					}
					placement := mygo.PlacementRight
					if !left {
						placement = mygo.PlacementLeft
					}
					if err := desktop.Tooltip(c, host, key+":tooltip:"+item.Key, trigger, text, desktop.TooltipProps{Placement: placement, Gap: 8, Disabled: itemDisabled}); err != nil && *popupErr == nil {
						*popupErr = err
					}
				}
				if !icon && item.Action != nil {
					ui.Box(c).Disabled(itemDisabled).Children(func() { item.Action(c) })
				}
			})
			if len(item.Children) > 0 && !icon {
				panel := part.Panel(func() {
					sidebarTree(c, host, key, s, selected, p, item.Children, false, left, itemDisabled, depth+1, popupErr)
				})
				if panel.Valid() {
					panel.Gap(4).MarginStart(14).PaddingStart(10).BorderColor(t.SidebarBorder).Role(ui.RoleGroup).Label(item.Label + " routes").AccessHidden(!*open).IgnoreInput(!*open).Disabled(!*open || itemDisabled).DimDisabled(false)
					if c.Direction() == ui.RightToLeft {
						panel.BorderWidth(0, 1, 0, 0)
					} else {
						panel.BorderWidth(0, 0, 0, 1)
					}
					trigger.Controls(panel)
				}
			}
		})
	}
}
