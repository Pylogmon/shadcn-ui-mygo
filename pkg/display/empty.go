package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type EmptyMediaVariant uint8

const (
	EmptyMediaDefault EmptyMediaVariant = iota
	EmptyMediaIcon
)

type EmptyProps struct {
	Title, Description string
	// Action retains the original callback API, following typed Content.
	Action          func()
	Label           string
	Icon            *ui.SVG
	MediaVariant    EmptyMediaVariant
	Bordered, Muted bool
	// Padding accepts an explicit zero; nil supplies the upstream 48 DIPs.
	Padding *float32
	// ContentWidth bounds the header/content, default 384 DIPs.
	ContentWidth float32
}

type EmptySlots struct {
	// Header replaces the whole media/title/description header when non-nil.
	// The other matching callbacks replace their props. Content follows header.
	Header, Media, Title, Description, Content func(*ui.Context)
}

type emptySlotKey uint8

const (
	emptyHeaderKey emptySlotKey = iota
	emptyMediaKey
	emptyTitleKey
	emptyDescriptionKey
	emptyContentKey
)

func emptyFinite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }

func Empty(c *ui.Context, p EmptyProps) ui.Element { return EmptyWith(c, p, EmptySlots{}) }

// EmptyWith lays out a centered, bounded media/header and arbitrary native
// content. An empty state is not automatically a live region; callers can
// intentionally choose Status on the returned element when it announces news.
func EmptyWith(c *ui.Context, p EmptyProps, slots EmptySlots) ui.Element {
	t := theme.TokensOf(c)
	padding := float32(48)
	if p.Padding != nil && emptyFinite(*p.Padding) {
		padding = max(0, *p.Padding)
	}
	width := p.ContentWidth
	if width <= 0 || !emptyFinite(width) {
		width = 384
	}
	root := ui.Column(c).FillWidth().Grow(1).MinWidth(0).AlignItems(ui.Center).Justify(ui.Center).Gap(16).Padding(padding).
		Radius(t.RadiusLG()).BorderColor(t.Border).BorderStyle(ui.BorderDashed).TextColor(t.Foreground).FontSize(14).TextAlign(ui.Center)
	if p.Bordered {
		root.Border(1, t.Border)
	}
	if p.Muted {
		root.Background(t.Muted)
	}
	if p.Label != "" {
		root.Role(ui.RoleGroup).Label(p.Label)
	}
	root.Children(func() {
		if slots.Header != nil || slots.Media != nil || p.Icon != nil || slots.Title != nil || p.Title != "" || slots.Description != nil || p.Description != "" {
			EmptyHeader(c.Key(emptyHeaderKey), func(child *ui.Context) {
				if slots.Header != nil {
					slots.Header(child)
					return
				}
				if slots.Media != nil || p.Icon != nil {
					EmptyMedia(child.Key(emptyMediaKey), EmptyMediaProps{Variant: p.MediaVariant, Icon: p.Icon, Content: slots.Media})
				}
				if slots.Title != nil || p.Title != "" {
					ui.Column(child.Key(emptyTitleKey)).FillWidth().MinWidth(0).AlignItems(ui.Center).FontSize(18).FontWeight(500).Children(func() {
						if slots.Title != nil {
							slots.Title(child)
						} else {
							EmptyTitle(child, p.Title)
						}
					})
				}
				if slots.Description != nil || p.Description != "" {
					ui.Column(child.Key(emptyDescriptionKey)).FillWidth().MinWidth(0).AlignItems(ui.Center).TextColor(t.MutedForeground).LineHeight(1.625).Children(func() {
						if slots.Description != nil {
							slots.Description(child)
						} else {
							EmptyDescription(child, p.Description)
						}
					})
				}
			}).MaxWidth(width)
		}
		if slots.Content != nil || p.Action != nil {
			EmptyContent(c.Key(emptyContentKey), func(child *ui.Context) {
				if slots.Content != nil {
					slots.Content(child)
				}
				if p.Action != nil {
					p.Action()
				}
			}).MaxWidth(width)
		}
	})
	return root
}

func EmptyHeader(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ui.Column(c).FillWidth().MaxWidth(384).MinWidth(0).AlignItems(ui.Center).Gap(8).TextAlign(ui.Center).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

type EmptyMediaProps struct {
	Variant EmptyMediaVariant
	Icon    *ui.SVG
	// Label names an icon or a group of custom media. Custom content keeps its
	// accessible children; an unnamed simple Icon is decorative by default.
	Label      string
	Decorative bool
	Content    func(*ui.Context)
}

func EmptyMedia(c *ui.Context, p EmptyMediaProps) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Box(c).Shrink(0).Center().Margin(0, 0, 8, 0)
	if p.Variant == EmptyMediaIcon {
		root.Size(40, 40).Radius(t.RadiusLG()).Background(t.Muted).TextColor(t.Foreground)
	}
	decorative := p.Decorative || p.Label == "" && p.Icon != nil && p.Content == nil
	if decorative {
		root.IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false)
	} else if p.Label != "" {
		role := ui.RoleImage
		if p.Content != nil {
			role = ui.RoleGroup
		}
		root.Role(role).Label(p.Label)
	}
	root.Children(func() {
		ui.Box(c).Center().Children(func() {
			if p.Content != nil {
				p.Content(c)
			} else if p.Icon != nil {
				ui.Icon(c, p.Icon).Size(24, 24).AccessHidden(true)
			}
		})
	})
	return root
}

func EmptyTitle(c *ui.Context, title string) ui.Element {
	return ui.Text(c, title).FontSize(18).FontWeight(500).TextAlign(ui.Center)
}

func EmptyDescription(c *ui.Context, description string) ui.Element {
	return ui.Text(c, description).FontSize(14).LineHeight(1.625).TextColor(theme.TokensOf(c).MutedForeground).TextAlign(ui.Center)
}

func EmptyContent(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ui.Column(c).FillWidth().MaxWidth(384).MinWidth(0).AlignItems(ui.Center).Gap(16).FontSize(14).TextAlign(ui.Center).Children(func() {
		if content != nil {
			content(c)
		}
	})
}
