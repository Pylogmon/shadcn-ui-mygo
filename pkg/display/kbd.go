package display

import (
	"runtime"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type KbdProps struct {
	// AlignSelf positions the keycap across its parent's axis; zero is Start.
	AlignSelf          ui.Align
	Label, Description string
	Icon, EndIcon      *ui.SVG
	// Content replaces the key's non-interactive visuals. Give Label for
	// custom content whose meaning cannot be read from key.
	Content func(*ui.Context)
	// InTooltip uses the source's inverse-tooltip keycap colors.
	InTooltip bool
}

type KbdGroupProps struct {
	// AlignSelf positions the whole group across its parent's axis; zero is Start.
	AlignSelf ui.Align
}

type kbdPart uint8

const (
	kbdVisual kbdPart = iota
	kbdLeading
	kbdText
	kbdTrailing
)

func Kbd(c *ui.Context, key string) ui.Element { return KbdWith(c, key, KbdProps{}) }

// KbdWith is an inert, content-sized keycap. It does not install a shortcut;
// the application's native shortcut and its accessible name remain separate.
func KbdWith(c *ui.Context, key string, p KbdProps) ui.Element {
	t := theme.TokensOf(c)
	background, foreground := t.Muted, t.MutedForeground
	if p.InTooltip {
		alpha := float32(.2)
		if t.Dark {
			alpha = .1
		}
		background, foreground = t.Background.Alpha(alpha), t.Background
	}
	label := p.Label
	if label == "" {
		label = key
	}
	family := c.Theme().Font
	if family == "" {
		family = "system-ui"
	}
	root := ui.Row(c).AlignSelf(p.AlignSelf).Shrink(0).Height(20).MinWidth(20).PaddingX(4).Gap(4).
		AlignItems(ui.Center).Justify(ui.Center).Radius(t.RadiusSM()).Background(background).TextColor(foreground).
		Font(family).FontSize(12).FontWeight(500).LineHeight(4.0 / 3.0).Role(ui.RoleText).Label(label).Description(p.Description).
		IgnoreInput(true).FocusRing(false)
	if label == "" {
		root.Role(ui.RoleNone)
	}
	root.Children(func() {
		ui.Row(c.Key(kbdVisual)).Gap(4).AlignItems(ui.Center).IgnoreInput(true).Disabled(true).DimDisabled(false).AccessHidden(true).Children(func() {
			if p.Content != nil {
				p.Content(c)
				return
			}
			if p.Icon != nil {
				ui.Icon(c.Key(kbdLeading), p.Icon).Size(12, 12)
			}
			if key != "" {
				ui.Text(c.Key(kbdText), key).SingleLine()
			}
			if p.EndIcon != nil {
				ui.Icon(c.Key(kbdTrailing), p.EndIcon).Size(12, 12)
			}
		})
	})
	return root
}

func KbdGroup(c *ui.Context, content func(*ui.Context)) ui.Element {
	return KbdGroupWith(c, KbdGroupProps{}, content)
}

// KbdGroupWith configures the outer row independently of its inert keycaps.
func KbdGroupWith(c *ui.Context, p KbdGroupProps, content func(*ui.Context)) ui.Element {
	return ui.Row(c).AlignSelf(p.AlignSelf).Gap(4).AlignItems(ui.Center).Role(ui.RoleNone).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

// KbdShortcut formats physical modifiers for this native desktop. ui.Cmd is
// already Command on macOS and Control elsewhere; key is the visible key name.
func KbdShortcut(c *ui.Context, modifiers ui.Modifiers, key string) ui.Element {
	return KbdShortcutWith(c, modifiers, key, KbdGroupProps{})
}

// KbdShortcutWith adds outer alignment to the platform-formatted shortcut.
func KbdShortcutWith(c *ui.Context, modifiers ui.Modifiers, key string, p KbdGroupProps) ui.Element {
	return KbdGroupWith(c, p, func(child *ui.Context) {
		for _, modifier := range kbdModifiers(modifiers, runtime.GOOS) {
			Kbd(child, modifier)
		}
		if key != "" {
			Kbd(child, key)
		}
	})
}

func kbdModifiers(modifiers ui.Modifiers, platform string) []string {
	keys := []struct {
		bit  ui.Modifiers
		name string
	}{{ui.Ctrl, "Ctrl"}, {ui.Alt, "Alt"}, {ui.Shift, "Shift"}, {ui.Super, "Super"}}
	if platform == "darwin" {
		keys = []struct {
			bit  ui.Modifiers
			name string
		}{{ui.Ctrl, "⌃"}, {ui.Alt, "⌥"}, {ui.Shift, "⇧"}, {ui.Super, "⌘"}}
	}
	if platform == "windows" {
		keys[3].name = "Win"
	}
	var result []string
	for _, key := range keys {
		if modifiers&key.bit != 0 {
			result = append(result, key.name)
		}
	}
	return result
}
