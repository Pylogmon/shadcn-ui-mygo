package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type CardSize uint8

const (
	CardDefault CardSize = iota
	CardSmall
)

type CardProps struct {
	Label string
	Size  CardSize
	// Spacing overrides the default 24 DIPs (16 for CardSmall).
	Spacing float32
	// HeaderBorder and FooterBorder insert full-width dividers, with the
	// corresponding inner spacing supplied by CardWith.
	HeaderBorder, FooterBorder bool
	// ContentBleed lets tables/images reach the side edges of the card.
	ContentBleed bool
	// MediaBottom places Media after Footer, flush with the bottom edge.
	MediaBottom bool
}

type CardSlots struct {
	Title, Description string
	// Header replaces the title/description stack, while Action remains at
	// the logical end. All callbacks receive the card's own child context.
	Header, Action func(*ui.Context)
	Media, Content func(*ui.Context)
	Footer         func(*ui.Context)
}

// CardWith lays out padded compound slots, actions and edge-to-edge media.
// Card(c, content) remains the compatible, uniformly padded container.
func CardWith(c *ui.Context, p CardProps, slots CardSlots) ui.Element {
	t := theme.TokensOf(c)
	spacing := p.Spacing
	if spacing <= 0 || math.IsNaN(float64(spacing)) || math.IsInf(float64(spacing), 0) {
		spacing = 24
		if p.Size == CardSmall {
			spacing = 16
		}
	}
	top, bottom := spacing, spacing
	if slots.Media != nil {
		if p.MediaBottom {
			bottom = 0
		} else {
			top = 0
		}
	}
	root := ui.Column(c).FillWidth().MinWidth(0).Gap(spacing).Padding(top, 0, bottom).
		Radius(t.RadiusXL()).Clip().Background(t.Card).TextColor(t.CardForeground).
		FontSize(14).Border(1, t.Foreground.Alpha(.1)).Shadow(0, 1, 2, 0, ui.RGB(0, 0, 0).Alpha(.05))
	if p.Label != "" {
		root.Role(ui.RoleGroup).Label(p.Label)
	}
	root.Children(func() {
		theme.WithSurface(c, t.Card, func() {
			media := func() {
				ui.Box(c.Key("media")).FillWidth().MinWidth(0).Clip().Children(func() { slots.Media(c) })
			}
			if slots.Media != nil && !p.MediaBottom {
				media()
			}
			if slots.Header != nil || slots.Title != "" || slots.Description != "" || slots.Action != nil {
				ui.Column(c.Key("header")).FillWidth().MinWidth(0).Children(func() {
					header := ui.Row(c).FillWidth().MinWidth(0).PaddingX(spacing).Gap(12).AlignItems(ui.Start)
					if p.HeaderBorder {
						header.Padding(0, spacing, spacing)
					}
					header.Children(func() {
						if slots.Header != nil || slots.Title != "" || slots.Description != "" {
							ui.Column(c.Key("heading")).Basis(0).Grow(1).MinWidth(0).Gap(4).Children(func() {
								if slots.Header != nil {
									slots.Header(c)
								} else {
									if slots.Title != "" {
										CardTitle(c, slots.Title, p.Size)
									}
									if slots.Description != "" {
										CardDescription(c, slots.Description)
									}
								}
							})
						} else {
							ui.Box(c.Key("heading")).Grow(1)
						}
						if slots.Action != nil {
							CardAction(c.Key("action"), func(child *ui.Context) { slots.Action(child) })
						}
					})
					if p.HeaderBorder {
						Separator(c, false)
					}
				})
			}
			if slots.Content != nil {
				body := ui.Column(c.Key("content")).FillWidth().MinWidth(0).Gap(12)
				if !p.ContentBleed {
					body.PaddingX(spacing)
				}
				body.Children(func() { slots.Content(c) })
			}
			if slots.Footer != nil {
				ui.Column(c.Key("footer")).FillWidth().MinWidth(0).Children(func() {
					if p.FooterBorder {
						Separator(c, false)
					}
					footer := ui.Row(c).FillWidth().MinWidth(0).PaddingX(spacing).Gap(8).AlignItems(ui.Center)
					if p.FooterBorder {
						footer.Padding(spacing, spacing, 0)
					}
					footer.Children(func() { slots.Footer(c) })
				})
			}
			if slots.Media != nil && p.MediaBottom {
				media()
			}
		})
	})
	return root
}

func CardTitle(c *ui.Context, title string, size ...CardSize) ui.Element {
	font := float32(16)
	if len(size) > 0 && size[0] == CardSmall {
		font = 14
	}
	return theme.StyleHeading(c, ui.Text(c, title).FontSize(font).LineHeight(1.5).FontWeight(500))
}

func CardDescription(c *ui.Context, description string) ui.Element {
	return ui.Text(c, description).FontSize(14).TextColor(theme.TokensOf(c).MutedForeground)
}

func CardAction(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ui.Row(c).Shrink(0).Gap(8).AlignSelf(ui.Start).Children(func() {
		if content != nil {
			content(c)
		}
	})
}
