package overlay

import "github.com/egoist/mygo/ui"

// DropdownMenu is the first-version window-bounded menu.
// Deprecated: use desktop.DropdownMenu for a shadcn Flyout, NativeDropdownMenu
// for a system menu, or InWindowDropdownMenu for an explicit embedded surface.
func DropdownMenu(c *ui.Context, open *bool, p MenuProps, trigger func(ui.Element)) string {
	return InWindowDropdownMenu(c, open, p, trigger)
}

// ContextMenu is the first-version window-bounded context menu.
// Deprecated: use desktop.ContextMenu, NativeContextMenu or InWindowContextMenu.
func ContextMenu(c *ui.Context, target ui.Element, p MenuProps) string {
	return InWindowContextMenu(c, target, p)
}

// Menubar is the first-version window-bounded menu bar.
// Deprecated: use NativeMenubar for platform menus or InWindowMenubar explicitly.
func Menubar(c *ui.Context, items []MenubarItem) string { return InWindowMenubar(c, items) }

// Popover is the first-version window-bounded popover.
// Deprecated: use desktop.Popover, desktop.NativePopover or InWindowPopover.
// Independent windows require a callback receiving their own *ui.Context.
func Popover(c *ui.Context, anchor ui.Element, open *bool, p PopoverProps, content func()) ui.Element {
	return InWindowPopover(c, anchor, open, p, content)
}

// Tooltip is the first-version window-bounded tooltip.
// Deprecated: use InWindowTooltip to request this surface explicitly.
func Tooltip(c *ui.Context, anchor ui.Element, text string, p TooltipProps) ui.Element {
	return InWindowTooltip(c, anchor, text, p)
}

// HoverCard is the first-version window-bounded hover card.
// Deprecated: use InWindowHoverCard to request this surface explicitly.
func HoverCard(c *ui.Context, anchor ui.Element, p HoverCardProps, content func()) ui.Element {
	return InWindowHoverCard(c, anchor, p, content)
}
