package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type ItemVariant uint8

const (
	ItemDefault ItemVariant = iota
	ItemOutline
	ItemMuted
)

type ItemSize uint8

const (
	ItemDefaultSize ItemSize = iota
	ItemSmall
	ItemExtraSmall
)

type ItemPrimaryKind uint8

const (
	ItemPrimaryButton ItemPrimaryKind = iota
	ItemPrimaryLink
)

// ItemPrimary supplies the primary action for the media/body region. Header,
// Footer and Actions remain sibling controls and never activate this action.
// Link uses MyGo's actual Link widget, including its native URL/router policy.
type ItemPrimary struct {
	Kind       ItemPrimaryKind
	Label, URL string
	Disabled   bool
	Handle     *ui.Handle
	OnActivate func()
}

type ItemProps struct {
	// Key is stable among siblings; give every grouped/reordered item a unique key.
	Key                string
	Title, Description string
	Label              string
	Variant            ItemVariant
	Size               ItemSize
	Disabled, Invalid  bool
	// Leading/Trailing retain the original API. Media/Actions take precedence.
	Leading, Trailing func()
	Media             *ItemMediaProps
	Header, Footer    func(*ui.Context)
	Content, Actions  func(*ui.Context)
	// When Primary is set, Content and media are non-interactive visuals. Put
	// buttons/editors in Actions, Header or Footer; they keep their own focus.
	Primary *ItemPrimary
}

type itemSlot uint8

const (
	itemHeaderSlot itemSlot = iota
	itemMiddleSlot
	itemPrimarySlot
	itemTargetSlot
	itemVisualSlot
	itemMediaSlot
	itemBodySlot
	itemActionsSlot
	itemFooterSlot
)

type itemKey struct{ key string }

func itemMetrics(size ItemSize) (horizontal, vertical, gap float32) {
	switch size {
	case ItemSmall:
		return 12, 10, 10
	case ItemExtraSmall:
		return 10, 8, 8
	default:
		return 16, 14, 14
	}
}

// Item is a composed surface. Its returned outer Element reports Changed when
// the optional Primary activates; use Primary.OnActivate or Changed once, rather
// than treating the outer non-interactive container as a clickable button.
func Item(c *ui.Context, p ItemProps) ui.Element {
	t := theme.TokensOf(c)
	px, py, gap := itemMetrics(p.Size)
	root := motion.ApplyTransition(c, ui.Column(c).FillWidth().MinWidth(0).Gap(gap).Padding(py, px).Radius(t.RadiusMD()).
		Border(1, ui.Transparent).TextColor(t.Foreground).FontSize(14).Disabled(p.Disabled).Invalid(p.Invalid).DimDisabled(false))
	if p.Key != "" {
		root.Key(itemKey{p.Key})
	}
	if p.Label != "" {
		root.Role(ui.RoleGroup).Label(p.Label)
	}
	switch p.Variant {
	case ItemOutline:
		root.Border(1, t.Border)
	case ItemMuted:
		root.Background(t.Muted.Alpha(.5))
	}
	if p.Disabled {
		root.Opacity(.5)
	}
	var primary ui.Element
	root.Children(func() {
		if p.Header != nil {
			ItemHeader(c.Key(itemHeaderSlot), p.Header)
		}
		ui.Row(c).Key(itemMiddleSlot).FillWidth().MinWidth(0).Gap(gap).AlignItems(ui.Center).Wrap().Children(func() {
			region := ui.Box(c).Key(itemPrimarySlot).Basis(72).Grow(1).MinWidth(0)
			region.Children(func() {
				if p.Primary != nil {
					label := p.Primary.Label
					if label == "" {
						label = p.Label
					}
					if label == "" {
						label = p.Title
					}
					if label == "" {
						label = "Item"
					}
					if p.Primary.Kind == ItemPrimaryLink {
						primary = ui.Link(c.Key(itemTargetSlot), "", p.Primary.URL)
					} else {
						primary = ui.ButtonBase(c.Key(itemTargetSlot))
					}
					primary.Absolute().Fill().Label(label).Description(p.Description).Disabled(p.Primary.Disabled).
						Background(ui.Transparent).TextColor(ui.Transparent).Border(0, ui.Transparent).FocusRing(false).Cursor(ui.CursorPointer).Invalid(p.Invalid)
					if p.Primary.Handle != nil {
						primary.Bind(p.Primary.Handle)
					}
				}
				visual := ui.Row(c).Key(itemVisualSlot).FillWidth().MinWidth(0).Gap(gap).AlignItems(ui.Center)
				if p.Primary != nil {
					// The real native target occupies this region. Decorative
					// children must not replace it as a pointer/AX target.
					visual.IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false)
					if p.Primary.Disabled {
						visual.Opacity(.5)
					}
				}
				visual.Children(func() {
					if p.Media != nil || p.Leading != nil {
						media := ui.Box(c).Key(itemMediaSlot).Shrink(0)
						if p.Description != "" {
							media.AlignSelf(ui.Start).Translate(0, 2)
						}
						media.Children(func() {
							if p.Media != nil {
								ItemMedia(c, *p.Media, p.Size)
							} else {
								p.Leading()
							}
						})
					}
					ItemContent(c.Key(itemBodySlot), func(child *ui.Context) {
						if p.Content != nil {
							p.Content(child)
						} else {
							if p.Title != "" {
								ItemTitle(child, p.Title)
							}
							if p.Description != "" {
								ItemDescription(child, p.Description, p.Size)
							}
						}
					}, p.Size)
				})
			})
			if p.Actions != nil || p.Trailing != nil {
				ItemActions(c.Key(itemActionsSlot), func(child *ui.Context) {
					if p.Actions != nil {
						p.Actions(child)
					} else {
						p.Trailing()
					}
				})
			}
		})
		if p.Footer != nil {
			ItemFooter(c.Key(itemFooterSlot), p.Footer)
		}
	})
	if p.Invalid {
		root.Border(1, t.Destructive)
	}
	if primary.Valid() {
		if !primary.IsDisabled() && (primary.Hovered() || primary.Pressed()) {
			root.Background(t.Muted)
		}
		if primary.FocusVisible() {
			root.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
		}
		if primary.Clicked() && !primary.IsDisabled() {
			root.SetChanged()
			if p.Primary.OnActivate != nil {
				p.Primary.OnActivate()
			}
		}
	}
	return root
}

type ItemMediaVariant uint8

const (
	ItemMediaDefault ItemMediaVariant = iota
	ItemMediaIcon
	ItemMediaImage
)

type ItemMediaProps struct {
	Variant    ItemMediaVariant
	Image      ui.ImageSource
	Icon       *ui.SVG
	IconSize   float32
	Label      string
	Decorative bool
	// Custom content keeps its own native accessibility/input; Label names its
	// group. Simple images/icons use an image role. Decorative makes it inert.
	Content func(*ui.Context)
}

// ItemMedia accepts already decoded images; loading/fetching belongs to the app.
func ItemMedia(c *ui.Context, p ItemMediaProps, size ...ItemSize) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Box(c).Shrink(0).Gap(8).Center()
	if p.Label != "" && !p.Decorative {
		role := ui.RoleImage
		if p.Content != nil {
			role = ui.RoleGroup
		}
		root.Role(role).Label(p.Label)
	} else if p.Decorative || p.Content == nil {
		root.IgnoreInput(true).AccessHidden(true)
	}
	if p.Variant == ItemMediaImage {
		diameter := float32(40)
		if len(size) > 0 {
			if size[0] == ItemSmall {
				diameter = 32
			} else if size[0] == ItemExtraSmall {
				diameter = 24
			}
		}
		root.Size(diameter, diameter).Radius(t.RadiusSM()).Clip()
	}
	root.Children(func() {
		visual := ui.Box(c).Center()
		if p.Variant == ItemMediaImage {
			visual.Fill()
		}
		if p.Decorative || p.Content == nil {
			visual.IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false)
		}
		visual.Children(func() {
			if p.Content != nil {
				p.Content(c)
			} else if p.Image != nil && p.Variant == ItemMediaImage {
				w, h := ui.ImageSize(p.Image)
				if w > 0 && h > 0 && !math.IsNaN(float64(w)) && !math.IsNaN(float64(h)) && !math.IsInf(float64(w), 0) && !math.IsInf(float64(h), 0) {
					ui.Image(c, p.Image).Fill().Fit(ui.Cover).AccessHidden(true)
				}
			} else if p.Icon != nil {
				diameter := p.IconSize
				if diameter <= 0 || math.IsNaN(float64(diameter)) || math.IsInf(float64(diameter), 0) {
					diameter = 16
				}
				ui.Icon(c, p.Icon).Size(diameter, diameter).AccessHidden(true)
			}
		})
	})
	return root
}

func ItemContent(c *ui.Context, content func(*ui.Context), size ...ItemSize) ui.Element {
	gap := float32(4)
	if len(size) > 0 && size[0] == ItemExtraSmall {
		gap = 0
	}
	return ui.Column(c).Grow(1).MinWidth(0).Gap(gap).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

func ItemTitle(c *ui.Context, title string) ui.Element {
	return ui.Text(c, title).FontSize(14).FontWeight(500).LineHeight(1.375).MaxLines(1)
}

func ItemDescription(c *ui.Context, description string, size ...ItemSize) ui.Element {
	font := float32(14)
	if len(size) > 0 && size[0] == ItemExtraSmall {
		font = 12
	}
	return ui.Text(c, description).FontSize(font).LineHeight(1.5).MaxLines(2).TextColor(theme.TokensOf(c).MutedForeground)
}

func ItemActions(c *ui.Context, content func(*ui.Context)) ui.Element {
	// Native fixed-size custom controls cannot be shrunk by a parent. Only
	// genuinely overflowing actions scroll; ordinary strips stay content-sized.
	return ui.ScrollHorizontal(c).Shrink(0).MinWidth(0).MaxWidthPercent(100).Gap(8).AlignItems(ui.Center).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

func ItemHeader(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ui.Row(c).FillWidth().MinWidth(0).Gap(8).AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

func ItemFooter(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ItemHeader(c, content)
}

type ItemGroupProps struct {
	Items      []ItemProps
	Label      string
	Separators bool
	// Gap explicitly overrides auto spacing derived from the smallest item.
	Gap *float32
}

type itemGroupIndexKey struct{ index int }
type itemGroupSeparatorKey struct {
	before string
	index  int
}

// ItemGroup uses native list/list-item semantics. Explicit unique item Keys keep
// local state, caret and focus stable when data is reordered.
func ItemGroup(c *ui.Context, p ItemGroupProps) ui.Element {
	gap := float32(16)
	for _, item := range p.Items {
		if item.Size == ItemSmall {
			gap = min(gap, 10)
		} else if item.Size == ItemExtraSmall {
			gap = 8
		}
	}
	if p.Gap != nil && *p.Gap >= 0 && !math.IsNaN(float64(*p.Gap)) && !math.IsInf(float64(*p.Gap), 0) {
		gap = *p.Gap
	}
	root := ui.Column(c).FillWidth().MinWidth(0).Gap(gap).Role(ui.RoleList).Label(p.Label)
	root.Children(func() {
		for i, item := range p.Items {
			if p.Separators && i > 0 {
				ItemSeparator(c).Key(itemGroupSeparatorKey{before: item.Key, index: i})
			}
			label := item.Label
			if label == "" {
				label = item.Title
			}
			if item.Key == "" {
				ui.Box(c).Key(itemGroupIndexKey{i}).FillWidth().Children(func() {
					Item(c, item).Role(ui.RoleListItem).Label(label).PositionInSet(i+1, len(p.Items))
				})
			} else {
				Item(c, item).Role(ui.RoleListItem).Label(label).PositionInSet(i+1, len(p.Items))
			}
		}
	})
	return root
}

func ItemSeparator(c *ui.Context) ui.Element { return Separator(c, false).MarginY(8) }
