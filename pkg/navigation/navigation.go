// Package navigation provides native shadcn navigation components.
package navigation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func navButton(c *ui.Context, b ui.Element, label string) ui.Element {
	t := theme.TokensOf(c)
	motion.ApplyTransition(c, b.Height(36).PaddingX(12).Radius(t.Radius).Gap(8).FontSize(14).FontWeight(500).TextColor(t.Foreground))
	if b.Hovered() {
		b.Background(t.Accent).TextColor(t.AccentForeground)
	}
	b.Children(func() { ui.Text(c, label).SingleLine() })
	return b
}

// NavItem is a route entry. Key is returned on activation; applications decide
// whether that changes an in-app route or opens an external URL.
type NavItem struct {
	Key, Label, Description, Group string
	Icon                           *ui.SVG
	Disabled                       bool
	Children                       []NavItem
}
