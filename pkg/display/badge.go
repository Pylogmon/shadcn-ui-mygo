package display

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type BadgeVariant int

const (
	BadgeDefault BadgeVariant = iota
	BadgeSecondary
	BadgeOutline
	BadgeSoft
	BadgeGhost
	BadgeLink
)

type BadgeProps struct {
	// AlignSelf positions the whole badge across its parent's axis; zero is Start.
	AlignSelf ui.Align
	Variant   BadgeVariant
	// Color overrides BadgeDefault, BadgeOutline, BadgeSoft, BadgeLink and
	// BadgeGhost colors, including hover and focus feedback. Nil uses theme
	// styles; BadgeSecondary ignores it.
	Color *ui.Color
	// Surface overrides the carrying background; nil inherits its surface scope.
	Surface *ui.Color
	// InheritForeground lets inert Outline/Ghost/Link content share its parent
	// action's state foreground, including icons and custom visual slots.
	InheritForeground  bool
	Label, Description string
	Icon, EndIcon      *ui.SVG
	// Leading and Trailing replace their icons inside 12-DIP visual slots.
	// Content replaces text. Callbacks build non-interactive content; provide
	// Label when their meaning differs from text or no text is supplied.
	Leading, Trailing, Content func(*ui.Context)
	// Button requests a native button. URL requests a native Link instead;
	// BadgeLink by itself is only a visual variant, like upstream's span.
	Button                  bool
	URL                     string
	Disabled, Invalid, Busy bool
	// Handle binds the real button/link, including the Link in a compound badge.
	Handle *ui.Handle
	// OnActivate and the returned Element.Changed are alternative action
	// notification paths. URL links also use MyGo's native URL/router action.
	OnActivate func()
}

type badgeSlotKey uint8

const (
	badgeVisual badgeSlotKey = iota
	badgeLeading
	badgeContent
	badgeTrailing
	badgeTarget
)

// Badge is the non-interactive, content-sized legacy entry point.
func Badge(c *ui.Context, text string, variant BadgeVariant) ui.Element {
	return BadgeWith(c, text, BadgeProps{Variant: variant})
}

// BadgeWith supports all Vega variants, visual slots and native button/link
// composition. Its returned element styles the whole badge; Changed reports a
// primary activation. Handle identifies the actual focus target of a URL badge.
func BadgeWith(c *ui.Context, text string, p BadgeProps) ui.Element {
	label := p.Label
	if label == "" {
		label = text
	}
	interactive := p.Button || p.OnActivate != nil || p.URL != ""
	var root, target ui.Element
	if p.URL == "" && interactive {
		root = ui.ButtonBase(c).Role(ui.RoleButton)
		target = root
	} else {
		root = ui.Row(c)
		if p.URL == "" {
			root.Role(ui.RoleText)
		} else {
			root.Role(ui.RoleNone)
		}
	}
	disabled := p.Disabled || p.Busy && interactive
	motion.ApplyTransition(c, root.AlignSelf(p.AlignSelf).Shrink(0).Height(20).Radius(999).Gap(4).Padding(2, 8).Clip().
		AlignItems(ui.Center).Justify(ui.Center).FontSize(12).LineHeight(4.0/3.0).FontWeight(500).Label(label).Description(p.Description).
		Disabled(disabled).DimDisabled(false).Busy(p.Busy).Invalid(p.Invalid).FocusRing(false))
	start, end := p.Icon != nil || p.Leading != nil, p.EndIcon != nil || p.Trailing != nil
	if start {
		root.PaddingStart(6)
	}
	if end {
		root.PaddingEnd(6)
	}
	buildContent := func() {
		if start {
			ui.Box(c.Key(badgeLeading)).Size(12, 12).Shrink(0).Center().Children(func() {
				if p.Leading != nil {
					p.Leading(c)
				} else {
					ui.Icon(c, p.Icon).Size(12, 12)
				}
			})
		}
		if p.Content != nil {
			ui.Row(c.Key(badgeContent)).MinWidth(0).Center().Children(func() { p.Content(c) })
		} else if text != "" {
			ui.Text(c.Key(badgeContent), text).SingleLine()
		}
		if end {
			ui.Box(c.Key(badgeTrailing)).Size(12, 12).Shrink(0).Center().Children(func() {
				if p.Trailing != nil {
					p.Trailing(c)
				} else {
					ui.Icon(c, p.EndIcon).Size(12, 12)
				}
			})
		}
	}
	root.Children(func() {
		if p.URL != "" {
			// MyGo Link owns native URL/router navigation and keyboard semantics.
			// Link is text-only, so visuals occupy a separate inert sibling row.
			target = ui.Link(c.Key(badgeTarget), "", p.URL).Label(label).Description(p.Description).Absolute().Left(-1).Top(-1).Right(-1).Bottom(-1).SingleLine().
				Disabled(disabled).DimDisabled(false).Busy(p.Busy).Invalid(p.Invalid).FocusRing(false)
		}
		ui.Row(c.Key(badgeVisual)).Gap(4).AlignItems(ui.Center).IgnoreInput(true).Disabled(true).DimDisabled(false).AccessHidden(true).Children(buildContent)
	})
	if p.Handle != nil && target.Valid() {
		target.Bind(p.Handle)
	}
	if target.Valid() {
		target.Cursor(ui.CursorPointer)
	}
	disabled = disabled || root.IsDisabled()
	badgeAppearance(c, root, target, p, disabled)
	if target.Valid() && target.Clicked() && !disabled {
		root.SetChanged()
		if p.OnActivate != nil {
			p.OnActivate()
		}
	}
	return root
}

func badgeAppearance(c *ui.Context, root, target ui.Element, p BadgeProps, disabled bool) {
	variant, color, invalid := p.Variant, p.Color, p.Invalid
	t := theme.TokensOf(c)
	accent := t.Primary
	if color != nil {
		accent = *color
	}
	customColor := color != nil && variant != BadgeSecondary
	bg, fg, border := t.Primary, t.PrimaryForeground, ui.Transparent
	switch variant {
	case BadgeDefault:
		if customColor {
			bg = accent
		}
	case BadgeSecondary:
		bg, fg = t.Secondary, t.SecondaryForeground
	case BadgeOutline:
		bg, fg, border = ui.Transparent, t.Foreground, t.Border
		if customColor {
			fg, border = accent, accent
		}
	case BadgeSoft:
		bg, fg = accent.Alpha(.1), accent
		if t.Dark {
			bg = accent.Alpha(.2)
		}
	case BadgeGhost:
		bg, fg = ui.Transparent, t.Foreground
		if customColor {
			fg = accent
		}
	case BadgeLink:
		bg, fg = ui.Transparent, accent
	}
	hovered := !disabled && (target.Valid() && (target.Hovered() || target.Pressed()) || !target.Valid() && root.Hovered())
	if hovered {
		switch variant {
		case BadgeGhost:
			if customColor {
				bg = accent.Alpha(.1)
				if t.Dark {
					bg = accent.Alpha(.2)
				}
			} else {
				bg, fg = t.Muted, t.MutedForeground
				if t.Dark {
					bg = t.Muted.Alpha(.5)
				}
			}
		case BadgeLink:
			// MyGo offsets from the native font's baseline. Four DIPs puts
			// the line outside this 20-DIP pill's clip with system fonts;
			// one DIP keeps it visible. Inherit decoration through the whole
			// visual subtree so custom text has the same hover behavior.
			root.Underline().UnderlineOffset(1)
		default:
			if target.Valid() {
				switch variant {
				case BadgeOutline:
					if customColor {
						bg = accent.Alpha(.1)
						if t.Dark {
							bg = accent.Alpha(.2)
						}
					} else {
						bg, fg = t.Muted, t.MutedForeground
					}
				case BadgeSoft:
					bg = accent.Alpha(.2)
					if t.Dark {
						bg = accent.Alpha(.3)
					}
				case BadgeSecondary:
					bg = t.Secondary.Alpha(.8)
				default:
					bg = bg.Alpha(.8)
				}
			}
		}
	}
	surface := theme.ResolveSurface(c, p.Surface)
	if variant == BadgeSoft || variant == BadgeLink || customColor && (variant == BadgeGhost || variant == BadgeOutline) {
		quiet, active := ui.Transparent, ui.Transparent
		if variant != BadgeLink {
			tint := float32(.1)
			if t.Dark {
				tint = .2
			}
			active = accent.Alpha(tint)
			if variant == BadgeSoft {
				quiet, active = active, accent.Alpha(tint+.1)
			}
		}
		if !target.Valid() && variant != BadgeGhost {
			active = quiet
		}
		backgrounds := []ui.Color{theme.Composite(bg, surface), theme.Composite(quiet, surface), theme.Composite(active, surface)}
		fg = theme.ReadableForeground(c, accent, backgrounds...)
		for _, background := range backgrounds {
			if theme.Contrast(theme.Composite(fg, background), background) < 4.5 {
				root.Transition(ui.ElementTransition{Duration: -1, Colors: true})
			}
		}
	}
	if customColor && variant == BadgeDefault {
		// Choose readable text against the fill composited over the theme page.
		fill := theme.Composite(bg, surface)
		fg = ui.RGB(255, 255, 255)
		if theme.Contrast(ui.RGB(0, 0, 0), fill) > theme.Contrast(fg, fill) {
			fg = ui.RGB(0, 0, 0)
		}
	}
	root.Background(bg).Border(1, border)
	if !p.InheritForeground || variant != BadgeOutline && variant != BadgeGhost && variant != BadgeLink {
		root.TextColor(fg)
	}
	if invalid {
		root.Border(1, t.Destructive)
	}
	if target.Valid() && target.FocusVisible() {
		ring := t.Ring.Alpha(.5)
		border := t.Ring
		if variant == BadgeSoft || customColor || invalid {
			color := accent
			if invalid {
				color = t.Destructive
			}
			ring, border = color.Alpha(.2), color.Alpha(.4)
			if t.Dark {
				ring = color.Alpha(.4)
			}
		}
		if c.Preferences().HighContrast && !invalid {
			border, ring = t.Ring, t.Ring.Alpha(.5)
		}
		root.Border(1, border).Shadow(0, 0, 0, 3, ring)
	}
	if disabled {
		root.Opacity(.5)
		if target.Valid() {
			target.Cursor(ui.CursorNotAllowed)
		}
	}
}
