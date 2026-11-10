package theme

import (
	"fmt"
	"weak"

	"github.com/egoist/mygo/ui"
)

// MenuColor matches Create's menuColor. Inverted forces the popup's dark
// palette; it does not invert an already-dark palette back to light.
type MenuColor string

const (
	MenuDefault             MenuColor = "default"
	MenuInverted            MenuColor = "inverted"
	MenuDefaultTranslucent  MenuColor = "default-translucent"
	MenuInvertedTranslucent MenuColor = "inverted-translucent"
)

type MenuAccent string

const (
	MenuSubtle MenuAccent = "subtle"
	MenuBold   MenuAccent = "bold"
)

// MenuOptions configures styled menu surfaces in this theme scope. Inverted
// tokens are a complete, explicitly supplied dark palette for custom scopes.
// Controller supplies the matching palette automatically. Values are copied;
// neither the options nor their token pointer is retained.
type MenuOptions struct {
	Color          MenuColor
	Accent         MenuAccent
	InvertedTokens *Tokens
}

func (p MenuOptions) Normalized() MenuOptions {
	if p.Color == "" {
		p.Color = MenuDefault
	}
	if p.Accent == "" {
		p.Accent = MenuSubtle
	}
	if (p.Color == MenuDefaultTranslucent || p.Color == MenuInvertedTranslucent) && (p.Accent == MenuSubtle || p.Accent == MenuBold) {
		p.Accent = MenuSubtle
	}
	return p
}
func (p MenuOptions) ColorIsTranslucent() bool {
	return p.Color == MenuDefaultTranslucent || p.Color == MenuInvertedTranslucent
}
func (p MenuOptions) Validate() error {
	switch p.Color {
	case "", MenuDefault, MenuInverted, MenuDefaultTranslucent, MenuInvertedTranslucent:
	default:
		return fmt.Errorf("unknown menu color %q", p.Color)
	}
	switch p.Accent {
	case "", MenuSubtle, MenuBold:
	default:
		return fmt.Errorf("unknown menu accent %q", p.Accent)
	}
	return nil
}

// MenuSnapshot is an immutable value snapshot, safe to copy to a separate
// native window. It contains no contexts, elements or application callbacks.
type MenuSnapshot struct {
	Color                     MenuColor
	Accent                    MenuAccent
	Fonts                     FontRoles
	scopeTokens               Tokens
	Tokens                    Tokens
	inverted                  Tokens
	hasInverted, nativeWindow bool
}

func (m MenuSnapshot) Translucent() bool {
	return m.Color == MenuDefaultTranslucent || m.Color == MenuInvertedTranslucent
}
func (m MenuSnapshot) Inverted() bool {
	return m.Color == MenuInverted || m.Color == MenuInvertedTranslucent
}
func (m MenuSnapshot) Animated() bool { return !m.Translucent() }

// ForNativeWindow selects the operating system's cross-window material path.
// Its framebuffer cannot sample pixels in the parent window or on the desktop.
func (m MenuSnapshot) ForNativeWindow() MenuSnapshot { m.nativeWindow = true; return m }
func (m MenuSnapshot) ForScene() MenuSnapshot        { m.nativeWindow = false; return m }
func (m MenuSnapshot) Theme() *ui.Theme              { return m.Tokens.ThemeWithMenus(m) }
func (t *Tokens) ThemeWithMenus(m MenuSnapshot) *ui.Theme {
	native := t.Theme()
	m.Fonts = m.Fonts.Normalized()
	native.Font = m.Fonts.Body
	m.scopeTokens = *t
	metadata.Store(weak.Make(native), themeMetadata{tokens: *t, menu: m, fonts: m.Fonts})
	return native
}
func MenuOf(c *ui.Context) MenuSnapshot {
	if native := c.Theme(); native != nil {
		if value, ok := metadata.Load(weak.Make(native)); ok {
			m := value.(themeMetadata).menu
			m.Fonts = FontsOf(c)
			return m
		}
	}
	return MenuSnapshot{Color: MenuDefault, Accent: MenuSubtle, Tokens: *TokensOf(c), Fonts: FontsOf(c)}
}
func menuTheme(c *ui.Context, t *Tokens, options []MenuOptions) *ui.Theme {
	return menuThemeFrom(t, MenuOf(c), options)
}

// MenuTheme makes a complete scoped theme with an explicit copied menu profile.
func (t *Tokens) MenuTheme(options MenuOptions) *ui.Theme {
	return menuThemeFrom(t, MenuSnapshot{Color: MenuDefault, Accent: MenuSubtle}, []MenuOptions{options})
}

func menuThemeFrom(t *Tokens, m MenuSnapshot, options []MenuOptions) *ui.Theme {
	root := *t
	if len(options) > 1 {
		panic("theme: at most one MenuOptions is allowed")
	}
	explicitPair := false
	if len(options) == 1 {
		p := options[0]
		if err := p.Validate(); err != nil {
			panic(err)
		}
		p = p.Normalized()
		m.Color, m.Accent = p.Color, p.Accent
		if p.InvertedTokens != nil {
			m.inverted = *p.InvertedTokens
			m.inverted.Dark = true
			m.hasInverted = true
			explicitPair = true
		}
	}
	if m.Color == "" {
		m.Color = MenuDefault
	}
	if m.Accent == "" {
		m.Accent = MenuSubtle
	}
	if m.Translucent() {
		m.Accent = MenuSubtle
	}
	if !explicitPair {
		if root.Dark {
			m.inverted = root
			m.hasInverted = true
		} else if pair, ok := matchingMenuPreset(root, m.Accent); ok {
			m.inverted = pair
			m.hasInverted = true
		} else if menuPalette(root) != menuPalette(m.scopeTokens) {
			m.hasInverted = false
		}
	}
	if m.Accent == MenuBold {
		root.Accent, root.AccentForeground = root.Primary, root.PrimaryForeground
	}
	m.Tokens = root
	if m.Inverted() {
		if m.hasInverted {
			m.Tokens = m.inverted
			m.Tokens.Radius, m.Tokens.FontSize, m.Tokens.Spacing = root.Radius, root.FontSize, root.Spacing
		} else if !root.Dark && len(options) > 0 {
			panic("theme: inverted menu color needs matching InvertedTokens in a custom light scope")
		}
		// An ordinary Scope with a completely custom, unpaired palette keeps
		// that palette exact, like unconditional CSS variables. It inherits the
		// dark class flag but never borrows an unrelated ancestor's dark colors.
		// Explicit MenuScope/MenuTheme asks for a paired inverted palette above.
		m.Tokens.Dark = true
		if m.Accent == MenuBold {
			m.Tokens.Accent, m.Tokens.AccentForeground = m.Tokens.Primary, m.Tokens.PrimaryForeground
		}
	}
	return root.ThemeWithMenus(m)
}

// Highlight and Separator implement the source's shared menu descendant rules.
func (m MenuSnapshot) Highlight(destructive bool) (ui.Color, ui.Color) {
	t := m.Tokens
	if m.Translucent() {
		return t.Foreground.Alpha(.1), t.AccentForeground
	}
	if destructive {
		alpha := float32(.1)
		if t.Dark {
			alpha = .2
		}
		return t.Destructive.Alpha(alpha), t.Destructive
	}
	return t.Accent, t.AccentForeground
}
func (m MenuSnapshot) Separator() ui.Color {
	if m.Translucent() {
		return m.Tokens.Foreground.Alpha(.05)
	}
	return m.Tokens.Border
}

// HasInvertedPalette reports whether the scope has a supplied or recognized
// matching dark palette. Unpaired custom ordinary Scopes retain their exact
// colors while inheriting the dark flag; MenuScope can provide a matching pair.
func (m MenuSnapshot) HasInvertedPalette() bool { return m.hasInverted }

type menuPaletteColors struct {
	dark   bool
	colors [20]ui.Color
}

func menuPalette(t Tokens) menuPaletteColors {
	return menuPaletteColors{t.Dark, [20]ui.Color{t.Background, t.Foreground, t.Card, t.CardForeground, t.Popover, t.PopoverForeground, t.Primary, t.PrimaryForeground, t.Secondary, t.SecondaryForeground, t.Muted, t.MutedForeground, t.Accent, t.AccentForeground, t.Destructive, t.DestructiveForeground, t.Border, t.Input, t.Ring, t.Sidebar}}
}
func matchingMenuPreset(t Tokens, accent MenuAccent) (Tokens, bool) {
	for _, name := range Presets() {
		base := Preset(name, t.Dark)
		if menuPalette(t) == menuPalette(*base) {
			return *Preset(name, true), true
		}
		if accent == MenuBold {
			base.Accent, base.AccentForeground = base.Primary, base.PrimaryForeground
			if menuPalette(t) == menuPalette(*base) {
				return *Preset(name, true), true
			}
		}
	}
	return Tokens{}, false
}
