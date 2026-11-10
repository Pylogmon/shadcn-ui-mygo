package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

type componentPreviewKey string

// Each target identifies a currently built preview, rather than an Element
// retained across frames. Unknown/fuzzy queries keep the existing page search.
var componentSearchTargets = []struct {
	page, anchor string
	names        []string
}{
	{"overview", "card", []string{"Card", "Card compositions", "CardSpacing", "Meeting Notes"}},
	{"controls", "button", []string{"Button"}},
	{"controls", "switch", []string{"Switch"}},
	{"controls", "checkbox", []string{"Checkbox"}},
	{"controls", "radio", []string{"Radio", "RadioGroup"}},
	{"controls", "textarea", []string{"Textarea"}},
	{"controls", "input", []string{"Input"}},
	{"controls", "input-otp", []string{"InputOTP"}},
	{"controls", "slider", []string{"Slider"}},
	{"controls", "select", []string{"Select"}},
	{"controls", "field", []string{"Field", "FieldComposition", "FieldSet", "FieldGroup"}},
	{"controls", "input-group", []string{"InputGroup"}},
	{"controls", "button-group", []string{"ButtonGroup"}},
	{"display", "disclosure", []string{"Accordion", "Collapsible", "Disclosure"}},
	{"content", "alert", []string{"Alert", "AlertColors"}},
	{"content", "empty", []string{"Empty", "EmptySearch"}},
	{"content", "item", []string{"Item"}},
	{"content", "badge", []string{"Badge"}},
	{"overlays", "popover", []string{"Popover"}},
	{"overlays", "hover-card", []string{"HoverCard"}},
	{"overlays", "dropdown-menu", []string{"DropdownMenu"}},
	{"overlays", "dialog", []string{"Dialog"}},
	{"navigation", "context-menu", []string{"ContextMenu"}},
	{"navigation", "menubar", []string{"Menubar"}},
	{"navigation", "tabs", []string{"Tabs"}},
}

func componentSearchName(name string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r', '-', '_':
			return -1
		}
		return r
	}, strings.ToLower(name))
}

func componentSearchTarget(page, query string) string {
	name := componentSearchName(query)
	for _, target := range componentSearchTargets {
		if target.page != page {
			continue
		}
		for _, alias := range target.names {
			if name == componentSearchName(alias) {
				return target.anchor
			}
		}
	}
	return ""
}

func (s *State) revealComponent(id string, element ui.Element) {
	if s.pendingPreview == id {
		element.ScrollIntoView(ui.ScrollIntoViewOptions{Block: ui.ScrollStart})
		s.pendingPreview = ""
	}
}

func (s *State) componentPreview(c *ui.Context, id string, build func()) {
	section := ui.Column(c.Key(componentPreviewKey(id))).FillWidth().MinWidth(0)
	section.Children(build)
	s.revealComponent(id, section)
}
