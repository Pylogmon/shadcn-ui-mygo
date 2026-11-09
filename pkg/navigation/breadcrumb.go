package navigation

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net/url"
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type BreadcrumbItem struct {
	Key, Label, Description, URL string
	Icon                         *ui.SVG
	Disabled                     bool
	// Content replaces this crumb's non-interactive visual content.
	Content func(*ui.Context)
}

type BreadcrumbProps struct {
	Label, EllipsisLabel string
	// MaxItems collapses middle ancestors, retaining the first and current
	// crumb. Values below three are raised to three when collapsing is needed.
	MaxItems                 int
	Separator                func(*ui.Context)
	OnSelect                 func(BreadcrumbItem)
	MenuWidth, MenuMaxHeight float32
}

type breadcrumbKey struct {
	kind uint8
	id   string
}
type breadcrumbState struct {
	open            bool
	menu            desktop.MenuState
	stamp, revision uint64
}

// Breadcrumb retains the simple event-returning API. URL ancestors use actual
// native Links; the current crumb is an inert current-page link.
func Breadcrumb(c *ui.Context, items []BreadcrumbItem) string {
	chosen, _ := breadcrumb(c, nil, "", items, BreadcrumbProps{}, false)
	return chosen
}

// BreadcrumbWith uses an explicit independent shadcn Flyout for overflow.
// A result/callback is delivered once in the current parent frame, after
// checking it against the current items, so stale routes cannot navigate.
func BreadcrumbWith(c *ui.Context, h *desktop.Host, key string, items []BreadcrumbItem, p BreadcrumbProps) (string, error) {
	if key == "" {
		return "", errors.New("navigation: BreadcrumbWith needs a stable key")
	}
	return breadcrumb(c, h, key, items, p, false)
}

// NativeBreadcrumbWith uses the OS menu adapter for overflow; crumb styling
// remains the same. It does not require a Host or paint a parent menu fallback.
func NativeBreadcrumbWith(c *ui.Context, key string, items []BreadcrumbItem, p BreadcrumbProps) (string, error) {
	if key == "" {
		return "", errors.New("navigation: NativeBreadcrumbWith needs a stable key")
	}
	return breadcrumb(c, nil, key, items, p, true)
}

func breadcrumb(c *ui.Context, h *desktop.Host, key string, items []BreadcrumbItem, p BreadcrumbProps, native bool) (string, error) {
	if c == nil || !c.Valid() {
		return "", errors.New("navigation: Breadcrumb needs a live context")
	}
	t := theme.TokensOf(c)
	label := p.Label
	if label == "" {
		label = "Breadcrumb"
	}
	root := ui.Row(c).MinWidth(0).Wrap().AlignItems(ui.Center).Role(ui.RoleNavigation).Label(label)
	if key != "" {
		root.Key(breadcrumbKey{0, key})
	}
	gap := float32(10)
	w, _ := c.Size()
	if w < 640 {
		gap = 6
	}
	root.Gap(gap).FontSize(14).TextColor(t.MutedForeground)
	s := ui.Local(root, "breadcrumb", func() breadcrumbState { return breadcrumbState{} })
	ids := make([]string, len(items))
	occurrences := map[string]int{}
	for i, item := range items {
		id := "index:" + strconv.Itoa(i)
		if item.Key != "" {
			ordinal := occurrences[item.Key]
			occurrences[item.Key]++
			id = "key:" + strconv.Quote(item.Key) + ":" + strconv.Itoa(ordinal)
		}
		ids[i] = id
	}
	hash := fnv.New64a()
	fmt.Fprint(hash, p.MaxItems, root.IsDisabled())
	for i, item := range items {
		fmt.Fprintf(hash, "%q/%q/%q/%q/%t/%p;", ids[i], item.Label, item.Description, item.URL, item.Disabled, item.Icon)
	}
	if stamp := hash.Sum64(); stamp != s.stamp {
		s.stamp = stamp
		s.revision++
	}
	hiddenStart, hiddenEnd := -1, -1
	if p.MaxItems > 0 && len(items) > max(3, p.MaxItems) {
		hiddenStart = 1
		hiddenEnd = len(items) - (max(3, p.MaxItems) - 2)
	}
	if hiddenStart < 0 {
		s.open = false
	}
	if root.IsDisabled() {
		s.open = false
	}
	var resultErr error
	chosen := ""
	activate := func(id string, route bool) {
		for i, item := range items {
			if ids[i] != id || i == len(items)-1 || item.Disabled || root.IsDisabled() {
				continue
			}
			chosen = item.Key
			if route && item.URL != "" {
				breadcrumbNavigate(c, item.URL)
			}
			if p.OnSelect != nil {
				p.OnSelect(item)
			}
			root.SetChanged()
			c.Invalidate()
			break
		}
	}
	root.Children(func() {
		ui.Row(c).Gap(gap).Wrap().AlignItems(ui.Center).Role(ui.RoleList).Children(func() {
			first := true
			separator := func() {
				if first {
					first = false
					return
				}
				ui.Box(c).Center().IgnoreInput(true).Disabled(true).DimDisabled(false).AccessHidden(true).Children(func() {
					if p.Separator != nil {
						p.Separator(c)
					} else {
						arrow := "›"
						if c.Direction() == ui.RightToLeft {
							arrow = "‹"
						}
						ui.Text(c, arrow).FontSize(14)
					}
				})
			}
			for i, item := range items {
				if i == hiddenStart {
					separator()
					ellipsisLabel := p.EllipsisLabel
					if ellipsisLabel == "" {
						ellipsisLabel = "Show breadcrumb ancestors"
					}
					menuItems := make([]overlay.MenuItem, 0, hiddenEnd-hiddenStart)
					for n := hiddenStart; n < hiddenEnd; n++ {
						it := items[n]
						menuItems = append(menuItems, overlay.MenuItem{Key: ids[n], Label: it.Label, Disabled: it.Disabled, Icon: it.Icon})
					}
					var menuChoice string
					ui.Box(c).Key(breadcrumbKey{1, "overflow"}).Role(ui.RoleListItem).Children(func() {
						style := func(trigger ui.Element) {
							motion.ApplyTransition(c, trigger.Label(ellipsisLabel).Size(20, 20).Radius(t.RadiusSM()).FontSize(16).TextColor(t.MutedForeground))
							if trigger.Hovered() {
								trigger.TextColor(t.Foreground)
							}
							trigger.Children(func() { ui.Text(c, "…").AccessHidden(true) })
						}
						if native {
							menuChoice, resultErr = desktop.NativeMenu(c, overlay.MenuProps{Items: menuItems}, style)
						} else {
							trigger := ui.ButtonBase(c)
							style(trigger)
							menuChoice, resultErr = desktop.Menu(c, h, "breadcrumb-overflow:"+key, trigger, &s.open, &s.menu, desktop.MenuProps{Items: menuItems, Label: ellipsisLabel, Width: p.MenuWidth, MaxHeight: p.MenuMaxHeight, Revision: s.revision})
						}
					})
					if menuChoice != "" {
						activate(menuChoice, true)
					}
				}
				if hiddenStart >= 0 && i >= hiddenStart && i < hiddenEnd {
					continue
				}
				separator()
				current := i == len(items)-1
				ui.Row(c).Key(breadcrumbKey{2, ids[i]}).Gap(6).MinWidth(0).Role(ui.RoleListItem).Children(func() {
					if current {
						ui.Row(c).Gap(6).MinWidth(0).Role(ui.RoleLink).Label(item.Label).Description(item.Description).Current(ui.CurrentPage).Disabled(true).DimDisabled(false).TextColor(t.Foreground).Children(func() { breadcrumbVisual(c, item) })
						return
					}
					region := ui.Row(c).Gap(6).MinWidth(0).AlignItems(ui.Center)
					var target ui.Element
					region.Children(func() {
						if item.URL != "" {
							target = ui.Link(c, "", item.URL)
						} else {
							target = ui.ButtonBase(c)
						}
						target.Absolute().Fill().Label(item.Label).Description(item.Description).Disabled(item.Disabled).Background(ui.Transparent).FocusRing(false).Cursor(ui.CursorPointer)
						ui.Row(c).Gap(6).MinWidth(0).AlignItems(ui.Center).IgnoreInput(true).Disabled(true).DimDisabled(false).AccessHidden(true).Children(func() { breadcrumbVisual(c, item) })
					})
					if !target.IsDisabled() && target.Hovered() {
						region.TextColor(t.Foreground)
					}
					if target.FocusVisible() {
						region.Shadow(0, 0, 0, 2, t.Ring.Alpha(.5)).Radius(t.RadiusSM())
					}
					if item.Disabled {
						region.Opacity(.5)
					}
					if target.Clicked() && !target.IsDisabled() {
						activate(ids[i], false)
					}
				})
			}
		})
	})
	return chosen, resultErr
}

func breadcrumbVisual(c *ui.Context, item BreadcrumbItem) {
	if item.Content != nil {
		item.Content(c)
		return
	}
	if item.Icon != nil {
		ui.Icon(c, item.Icon).Size(14, 14).AccessHidden(true)
	}
	ui.Text(c, item.Label).FontSize(14).SingleLine()
}

func breadcrumbNavigate(c *ui.Context, target string) {
	u, err := url.Parse(target)
	if err == nil && !u.IsAbs() && u.Host == "" && c.Router() != nil {
		c.Router().Push(target)
		return
	}
	c.OpenURL(target)
}

// BreadcrumbNav and the compound helpers support custom native compositions.
// Content callbacks build in their actual container; empty content never leaves
// a pending Context.Key for a following sibling.
func BreadcrumbNav(c *ui.Context, label string, content func(*ui.Context)) ui.Element {
	if label == "" {
		label = "Breadcrumb"
	}
	return ui.Row(c).MinWidth(0).Role(ui.RoleNavigation).Label(label).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

func BreadcrumbList(c *ui.Context, content func(*ui.Context)) ui.Element {
	gap := float32(10)
	width, _ := c.Size()
	if width < 640 {
		gap = 6
	}
	return ui.Row(c).MinWidth(0).Wrap().Gap(gap).AlignItems(ui.Center).FontSize(14).TextColor(theme.TokensOf(c).MutedForeground).Role(ui.RoleList).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

// BreadcrumbLink returns the visual container; an optional Handle binds its
// actual native Link. MyGo's text Link always underlines on hover, so keeping
// its empty target separate preserves shadcn's color-only hover appearance.
func BreadcrumbLink(c *ui.Context, label, url string, handle ...*ui.Handle) ui.Element {
	t := theme.TokensOf(c)
	root := motion.ApplyTransition(c, ui.Row(c).MinWidth(0).FontSize(14).TextColor(t.MutedForeground).Role(ui.RoleNone))
	var target ui.Element
	root.Children(func() {
		target = ui.Link(c, "", url).Label(label).Absolute().Fill().FocusRing(false)
		if len(handle) > 0 && handle[0] != nil {
			target.Bind(handle[0])
		}
		ui.Text(c, label).SingleLine().IgnoreInput(true).AccessHidden(true)
	})
	if target.Hovered() {
		root.TextColor(t.Foreground)
	}
	if target.FocusVisible() {
		root.Radius(t.RadiusSM()).Shadow(0, 0, 0, 2, t.Ring.Alpha(.5))
	}
	if target.Clicked() {
		root.SetChanged()
	}
	return root
}

func BreadcrumbPage(c *ui.Context, label string) ui.Element {
	return ui.Text(c, label).FontSize(14).FontWeight(400).TextColor(theme.TokensOf(c).Foreground).Role(ui.RoleLink).Label(label).Current(ui.CurrentPage).Disabled(true).DimDisabled(false)
}

func BreadcrumbSeparator(c *ui.Context, content ...func(*ui.Context)) ui.Element {
	return ui.Box(c).Center().IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false).Children(func() {
		if len(content) > 0 && content[0] != nil {
			content[0](c)
			return
		}
		arrow := "›"
		if c.Direction() == ui.RightToLeft {
			arrow = "‹"
		}
		ui.Text(c, arrow).FontSize(14)
	})
}

func BreadcrumbEllipsis(c *ui.Context) ui.Element {
	return ui.Box(c).Size(20, 20).Center().IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false).Children(func() { ui.Text(c, "…").FontSize(16) })
}
