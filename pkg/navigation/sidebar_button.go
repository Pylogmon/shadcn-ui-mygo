package navigation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// SidebarMenuButtonProps configures an independently composable native action.
// Content layout (including fixed icon columns/centered labels) belongs to the
// caller. Nil Color follows the official Sidebar semantic tokens.
type SidebarMenuButtonProps struct {
	Label, Description string
	Active, Disabled   bool
	Size               SidebarItemSize
	FontSize           float32
	Icon               *ui.SVG
	Outline            bool
	// Color is a native extension: colored foreground in every state, a soft
	// fill when Active, and a stronger tint on hover/press. Nil uses SidebarAccent.
	Color *ui.Color
	// Muted uses MutedForeground at rest, Foreground on hover/focus/press.
	// Color takes precedence. Useful for a free-license/account footer.
	Muted bool
	// Surface overrides the carrying background; nil uses the surface scope,
	// falling back to Sidebar for a standalone menu button.
	Surface *ui.Color
}

// SidebarMenuButton owns focus, keyboard activation, disabled/current-page
// semantics and standard motion. Content builds inert children inheriting the
// state foreground; nil renders Icon and Label in a normal start-aligned row.
// Use the returned Element's OnClick/Clicked for application navigation.
func SidebarMenuButton(c *ui.Context, p SidebarMenuButtonProps, content func(*ui.Context)) ui.Element {
	return StyleSidebarMenuButton(c, ui.ButtonBase(c), p, content)
}

// StyleSidebarMenuButton styles an existing native trigger (e.g. Collapsible)
// without replacing its input or accessibility behavior.
func StyleSidebarMenuButton(c *ui.Context, button ui.Element, p SidebarMenuButtonProps, content func(*ui.Context)) ui.Element {
	t := theme.TokensOf(c)
	height, reference := float32(32), float32(14)
	if p.Size == SidebarItemSmall {
		height, reference = 28, 12
	} else if p.Size == SidebarItemLarge {
		height = 48
	}
	font := theme.ControlFontSize(c, reference, p.FontSize)
	motion.ApplyTransition(c, button.Height(max(height, font*20/14+8)).Padding(4, 8).Gap(8).
		AlignItems(ui.Center).Justify(ui.Start).Radius(t.RadiusMD()).FontSize(font).LineHeight(20.0/14).
		Label(p.Label).Description(p.Description).Disabled(p.Disabled).DimDisabled(false).
		FocusRingColor(t.SidebarRing).Cursor(ui.CursorPointer))
	if p.Active {
		button.Current(ui.CurrentPage).FontWeight(500)
	}
	disabled := button.IsDisabled()
	interacting := !disabled && (button.Hovered() || button.Pressed())
	bg, fg := ui.Transparent, t.SidebarForeground
	if p.Outline {
		bg = t.Background
		button.Border(1, t.SidebarBorder)
	}
	if p.Active || interacting {
		bg, fg = t.SidebarAccent, t.SidebarAccentForeground
	}
	if p.Muted {
		fg = t.MutedForeground
		if interacting || !disabled && button.Focused() || p.Active {
			fg = t.Foreground
		}
	}
	// The sidebar fallback applies only outside a known enclosing surface.
	surface := theme.SurfaceOr(c, t.Sidebar)
	if p.Surface != nil {
		surface = theme.Composite(*p.Surface, surface)
	}
	if p.Color != nil {
		bg = ui.Transparent
		tint := float32(.1)
		if t.Dark {
			tint = .2
		}
		if p.Active {
			bg = p.Color.Alpha(tint)
			if interacting {
				bg = p.Color.Alpha(tint + .1)
			}
		} else if interacting {
			bg = p.Color.Alpha(tint)
		}
		quiet, active := ui.Transparent, p.Color.Alpha(tint)
		if p.Active {
			quiet, active = active, p.Color.Alpha(tint+.1)
		}
		backgrounds := []ui.Color{theme.Composite(bg, surface), surface, theme.Composite(quiet, surface), theme.Composite(active, surface)}
		fg = theme.ReadableForeground(c, *p.Color, backgrounds...)
		for _, background := range backgrounds {
			if theme.Contrast(theme.Composite(fg, background), background) < 4.5 {
				button.Transition(ui.ElementTransition{Duration: -1, Colors: true})
			}
		}
	}
	button.Background(bg).TextColor(fg)
	if disabled {
		button.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	button.Children(func() {
		theme.WithSurface(c, theme.Composite(bg, surface), func() {
			if content != nil {
				content(c)
				return
			}
			if p.Icon != nil {
				ui.Icon(c, p.Icon).Size(16, 16).Shrink(0).AccessHidden(true)
			}
			ui.Text(c, p.Label).MinWidth(0).SingleLine()
		})
	})
	return button
}
