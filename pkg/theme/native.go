package theme

import (
	"weak"

	"github.com/egoist/mygo/ui"
)

// FromNative maps the current platform UI theme onto semantic color roles. It
// never loads shadcn palette values or mutates the supplied native theme.
// These aliases let a constrained Go-rendered component retain native colors
// when the OS does not expose an equivalent widget (for example a range grid).
func FromNative(n *ui.Theme) *Tokens {
	if n == nil {
		n = ui.LightTheme()
	}
	font, spacing := n.FontSize, n.Spacing
	if font <= 0 {
		font = 14
	}
	if spacing <= 0 {
		spacing = 4
	}
	dangerText := ui.RGB(255, 255, 255)
	if Contrast(n.Danger, ui.RGB(0, 0, 0)) > Contrast(n.Danger, dangerText) {
		dangerText = ui.RGB(0, 0, 0)
	}
	return &Tokens{
		Dark: n.Dark, Background: n.Background, Foreground: n.Text,
		Card: n.Surface, CardForeground: n.Text, Popover: n.Surface, PopoverForeground: n.Text,
		Primary: n.Accent, PrimaryForeground: n.AccentText,
		Secondary: n.Surface, SecondaryForeground: n.Text, Muted: n.Surface, MutedForeground: n.TextMuted,
		Accent: n.SurfaceHover, AccentForeground: n.Text, Destructive: n.Danger, DestructiveForeground: dangerText,
		Border: n.Border, Input: n.Border, Ring: n.Focus,
		Chart1: n.Accent, Chart2: n.Success, Chart3: n.Warning, Chart4: n.Danger, Chart5: n.TextMuted,
		Sidebar: n.Surface, SidebarForeground: n.Text, SidebarPrimary: n.Accent, SidebarPrimaryForeground: n.AccentText,
		SidebarAccent: n.SurfaceHover, SidebarAccentForeground: n.Text, SidebarBorder: n.Border, SidebarRing: n.Focus,
		Radius: n.Radius, FontSize: font, Spacing: spacing,
	}
}

// NativeScope keeps every native Theme property intact, including font,
// tooltip delay and focus geometry, while supplying semantic aliases for
// Go-rendered rich content. It does not change the context outside Children.
func NativeScope(c *ui.Context) ui.Element {
	native := c.Theme()
	if native == nil {
		native = ui.LightTheme()
	}
	copy := *native
	remember(&copy, FromNative(native))
	key := weak.Make(&copy)
	value, _ := metadata.Load(key)
	m := value.(themeMetadata)
	m.menu.textScale = TextScale(c)
	metadata.Store(key, m)
	return ui.Themed(c, &copy)
}
