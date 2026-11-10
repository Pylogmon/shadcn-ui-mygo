package theme

import (
	"weak"

	"github.com/egoist/mygo/ui"
)

const (
	// FontInherit makes headings follow the body role. It is not a body family.
	FontInherit = "inherit"
	// FontSystem explicitly selects the native interface font. An empty Body
	// also selects it; an empty Heading instead means FontInherit.
	FontSystem = "system"
)

// FontRoles contains native family names or comma-separated fallback lists,
// not asset/catalog identifiers. The zero value keeps native defaults: system
// body and inherited headings. Values are copied across scopes and windows.
// Registration of optional font files remains an application/asset concern.
type FontRoles struct {
	Body, Heading string
}

func (f FontRoles) Normalized() FontRoles {
	if f.Body == FontSystem {
		f.Body = ""
	}
	if f.Heading == "" {
		f.Heading = FontInherit
	}
	return f
}

// HeadingFamily resolves the heading role to a native family list. StyleHeading
// additionally preserves ordinary ancestor inheritance for FontInherit.
func (f FontRoles) HeadingFamily() string {
	f = f.Normalized()
	if f.Heading == FontInherit {
		return f.Body
	}
	if f.Heading == FontSystem {
		return ""
	}
	return f.Heading
}

// FontsOf reads the innermost role snapshot. Body always reflects the actual
// native Theme.Font, including applications that set that public field directly.
// MyGo has no public computed-font getter for arbitrary ancestor .Font calls;
// use FontScope or Theme.Font when a family must cross a window boundary.
func FontsOf(c *ui.Context) FontRoles {
	f := FontRoles{}.Normalized()
	if native := c.Theme(); native != nil {
		if value, ok := metadata.Load(weak.Make(native)); ok {
			f = value.(themeMetadata).fonts.Normalized()
		}
		f.Body = native.Font
	}
	return f
}

// FontScope changes only font roles, retaining all native theme properties and
// menu/color metadata. Body is explicit (empty means system); to change only
// headings, copy FontsOf(c), change Heading, and pass the result here. Children
// inherit the body; leaving the scope restores its parent's roles.
func FontScope(c *ui.Context, fonts FontRoles) ui.Element {
	fonts = fonts.Normalized()
	base := c.Theme()
	if base == nil {
		base = ui.LightTheme()
	}
	native := *base
	native.Font = fonts.Body
	tokens := TokensOf(c)
	remember(&native, tokens)
	menu := MenuOf(c)
	menu.Fonts = fonts
	metadata.Store(weak.Make(&native), themeMetadata{tokens: *tokens, menu: menu, fonts: fonts})
	return ui.Themed(c, &native).Font(fonts.Body)
}

// StyleHeading applies the source cn-font-heading role to an existing title.
// Inherit leaves its native ancestor font untouched, preserving old defaults
// and local .Font composition; an explicit role also works for FontSystem.
func StyleHeading(c *ui.Context, e ui.Element) ui.Element {
	fonts := FontsOf(c)
	if fonts.Heading != FontInherit {
		e.Font(fonts.HeadingFamily())
	}
	return e
}
