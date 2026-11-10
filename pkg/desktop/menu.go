package desktop

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

type MenuProps struct {
	Items            []overlay.MenuItem
	Label            string
	Width, MaxHeight float32
	// ItemHeight is the full height of each themed row, including ItemContent.
	// Non-positive or non-finite values keep the standard 32 DIP row. NativeMenu
	// uses the operating system's own metrics and has no custom row-height path.
	ItemHeight       float32
	Placement        mygo.Placement
	Gap              int
	Revision         uint64
	KeepOpenOnSelect bool
	ItemContent      func(*ui.Context, overlay.MenuItem)
	OnChoose         func(string)
	OnDismiss        func()
}

type menuItemLayoutKey struct{ key string }
type menuItemLayout struct {
	height   float32
	stamp    uint64
	revision uint64
}

// Menu opens a shadcn-styled, measured menu in its own native Flyout. The Host
// is required when opening, and there is no parent-bounded overlay fallback.
// OnChoose runs in the child UI event; the return reports that key once on the
// next parent build. Use either to dispatch an action, not both.
func Menu(c *ui.Context, h *Host, key string, anchor ui.Element, open *bool, state *MenuState, p MenuProps) (string, error) {
	return menu(c, h, key, anchor, open, state, p, false)
}
func DropdownMenu(c *ui.Context, h *Host, key string, anchor ui.Element, open *bool, state *MenuState, p MenuProps) (string, error) {
	return Menu(c, h, key, anchor, open, state, p)
}
func menu(c *ui.Context, h *Host, key string, anchor ui.Element, open *bool, state *MenuState, p MenuProps, manual bool) (string, error) {
	if c == nil || !c.Valid() || !anchor.Valid() {
		return "", errors.New("desktop: Menu needs a live parent context and anchor")
	}
	if open == nil || state == nil {
		return "", errors.New("desktop: Menu needs caller-owned open state and MenuState")
	}
	result := state.Chosen
	state.Chosen = ""
	if !menuDimensionsFinite(p.Width, p.MaxHeight, p.Items) {
		*open, state.wasOpen = false, false
		if h != nil {
			h.Close(key)
		}
		return result, errors.New("desktop: menu dimensions must be finite")
	}
	if !manual {
		anchor.Bind(&state.contextFocus)
	}
	willOpen := *open
	if !manual && anchor.Clicked() {
		willOpen = !willOpen
	}
	fromUp := !manual && anchor.Shortcut(0, ui.KeyUp)
	fromDown := !manual && anchor.Shortcut(0, ui.KeyDown)
	if fromDown || fromUp {
		willOpen = true
		*open = true
	}
	if willOpen && !state.wasOpen {
		focus, x, y := state.contextFocus, state.contextX, state.contextY
		state.Reset()
		state.contextFocus, state.contextX, state.contextY = focus, x, y
		if fromUp {
			for i := len(p.Items) - 1; i >= 0; i-- {
				if menuEnabled(p.Items[i]) {
					state.Highlighted = []string{menuItemKey(p.Items[i], i)}
					break
				}
			}
		}
	}
	state.wasOpen = willOpen
	if h == nil {
		if willOpen {
			*open = false
			state.wasOpen = false
			return result, errors.New("desktop: opening a themed Menu requires a parent-bound Host; create desktop.NewHost(parentWindow)")
		}
		anchor.HasPopup(ui.PopupMenu).Expanded(false)
		return result, nil
	}
	width, maxHeight := p.Width, p.MaxHeight
	if width <= 0 {
		width = 192
	}
	if maxHeight <= 0 {
		maxHeight = 480
	}
	itemHeight := menuItemHeight(p.ItemHeight)
	stamp := menuItemsStamp(p.Items)
	layout := ui.Local(anchor, menuItemLayoutKey{key}, func() menuItemLayout { return menuItemLayout{height: itemHeight, stamp: stamp} })
	if layout.height != itemHeight || layout.stamp != stamp {
		layout.height = itemHeight
		layout.stamp = stamp
		layout.revision++
	}
	zero := float32(0)
	props := PopoverProps{menuMotion: menuMotionPopup, menuContext: manual, Label: p.Label, Placement: p.Placement, Width: int(math.Ceil(float64(width))), Height: int(math.Ceil(float64(menuPanelHeight(p.Items, itemHeight, maxHeight)))), Gap: p.Gap, FitContent: true, MaxHeight: int(math.Ceil(float64(maxHeight))), Plain: true, ManualTrigger: manual, Revision: p.Revision + layout.revision, ContentPadding: &zero, ContentRole: ui.RoleGroup,
		OnDismiss: func() {
			state.wasOpen = false
			if manual {
				state.contextFocus.Focus()
			}
			if p.OnDismiss != nil {
				p.OnDismiss()
			}
		}}
	chose := false
	closeMenu := func() {
		*open = false
		state.wasOpen = false
		h.Close(key)
		state.contextFocus.Focus()
		if !chose && p.OnDismiss != nil {
			p.OnDismiss()
		}
	}
	err := Popover(c, h, key, anchor, open, props, func(child *ui.Context) {
		// An independently scheduled child may render changed item data before
		// the parent can report its normal dimensions error.
		if !menuDimensionsFinite(width, maxHeight, p.Items) {
			closeMenu()
			return
		}
		MenuContent(child, state, p.Items, MenuContentProps{Label: p.Label, Width: width, MaxHeight: maxHeight, ItemHeight: itemHeight, KeepOpenOnSelect: p.KeepOpenOnSelect, ItemContent: p.ItemContent,
			OnChoose: func(value string) {
				chose = true
				if p.OnChoose != nil {
					p.OnChoose(value)
				}
				h.backend.InvalidateParent()
			}, OnClose: closeMenu})
	})
	anchor.HasPopup(ui.PopupMenu).Expanded(*open)
	if err != nil {
		state.wasOpen = false
	}
	return result, err
}

func menuDimensionsFinite(width, height float32, items []overlay.MenuItem) bool {
	if math.IsNaN(float64(width)) || math.IsInf(float64(width), 0) || math.IsNaN(float64(height)) || math.IsInf(float64(height), 0) {
		return false
	}
	for _, item := range items {
		if !menuDimensionsFinite(item.SubmenuWidth, 0, item.Children) {
			return false
		}
	}
	return true
}

// Only item data participates in automatic refresh. Custom content and callback
// changes continue to use the caller's explicit Revision contract.
func menuItemsStamp(items []overlay.MenuItem) uint64 {
	hash := fnv.New64a()
	var visit func([]overlay.MenuItem)
	visit = func(items []overlay.MenuItem) {
		fmt.Fprint(hash, "[")
		for _, item := range items {
			closeOnSelect, checked := "default", "none"
			if item.CloseOnSelect != nil {
				closeOnSelect = fmt.Sprint(*item.CloseOnSelect)
			}
			if item.Checked != nil {
				checked = fmt.Sprint(*item.Checked)
			}
			fmt.Fprintf(hash, "%q/%q/%q/%q/%t/%t/%t/%t/%t/%t/%t/%s/%s/%p/%g;", item.Key, item.Label, item.Shortcut, item.Group,
				item.Disabled, item.Destructive, item.Separator, item.Radio, item.Selected, item.Inset, item.GroupInset, checked, closeOnSelect, item.Icon, item.SubmenuWidth)
			visit(item.Children)
		}
		fmt.Fprint(hash, "]")
	}
	visit(items)
	return hash.Sum64()
}

// ContextMenu opens the themed Flyout at the secondary-click position, or
// below the target for Shift+F10. A stable native handle restores target focus.
func ContextMenu(c *ui.Context, h *Host, key string, target ui.Element, open *bool, state *MenuState, p MenuProps) (string, error) {
	if state == nil || open == nil {
		return "", errors.New("desktop: ContextMenu needs caller-owned open state and MenuState")
	}
	if c == nil || !c.Valid() || !target.Valid() {
		return "", errors.New("desktop: ContextMenu needs a live target")
	}
	target.Bind(&state.contextFocus)
	keyboard := target.Shortcut(ui.Shift, ui.KeyF10) || target.Shortcut(0, ui.KeyContextMenu)
	controlClick := target.Clicked() && target.ClickModifiers()&ui.Ctrl != 0
	if target.RightClicked() || keyboard || controlClick {
		r := target.Bounds()
		state.contextX, state.contextY = r.X, r.Y+r.H
		if !keyboard {
			if x, y, over := target.PointerPosition(); over {
				state.contextX, state.contextY = r.X+x, r.Y+y
			}
		}
		*open = true
		state.wasOpen = false
	}
	anchor := target
	if *open {
		ui.Overlay(c.Key("context-menu-anchor:"+key), func() {
			anchor = ui.Box(c).Absolute().Left(state.contextX).Top(state.contextY).Size(1, 1).IgnoreInput(true).AccessHidden(true)
		})
	}
	value, err := menu(c, h, key, anchor, open, state, p, true)
	target.HasPopup(ui.PopupMenu).Expanded(*open)
	return value, err
}

// NativeMenu always registers MyGo's OS menu adapter, independently of the
// surrounding preview policy. The previous policy is restored in this frame.
func NativeMenu(c *ui.Context, p overlay.MenuProps, trigger func(ui.Element)) (string, error) {
	if c == nil || !c.Valid() {
		return "", errors.New("desktop: NativeMenu needs a live context")
	}
	policy := c.MenuPresentation()
	c.SetMenuPresentation(ui.SystemMenus)
	defer c.SetMenuPresentation(policy)
	return overlay.NativeDropdownMenu(c, p, trigger), nil
}
func NativeDropdownMenu(c *ui.Context, p overlay.MenuProps, trigger func(ui.Element)) (string, error) {
	return NativeMenu(c, p, trigger)
}
func NativeContextMenu(c *ui.Context, target ui.Element, p overlay.MenuProps) (string, error) {
	if c == nil || !c.Valid() || !target.Valid() {
		return "", errors.New("desktop: NativeContextMenu needs a live context and target")
	}
	policy := c.MenuPresentation()
	c.SetMenuPresentation(ui.SystemMenus)
	defer c.SetMenuPresentation(policy)
	return overlay.NativeContextMenu(c, target, p), nil
}
