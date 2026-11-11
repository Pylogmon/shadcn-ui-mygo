package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Application values and public native state only. All content and callbacks
// are constructed by the current view; no Context or Element survives a frame.
type badgeCompositionPreviewState struct {
	ready, native            bool
	recipe, language, action string
	popupError               string
	router                   *ui.Router
	longScroll               ui.ScrollState
	activations              int
	lastVariant              display.BadgeVariant
}

var badgeCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Demo"},
	{Value: "variants", Label: "Variants (docs)"},
	{Value: "icon", Label: "Verified and Bookmark"},
	{Value: "spinner", Label: "Deleting and Generating"},
	{Value: "link", Label: "Open Link"},
	{Value: "colors", Label: "Custom Colors (docs)"},
	{Value: "rtl", Label: "Translated Badges"},
	{Value: "create-variants", Label: "Create: Six Variants"},
	{Value: "create-icon-left", Label: "Create: Icon Left"},
	{Value: "create-icon-right", Label: "Create: Icon Right"},
	{Value: "create-spinner", Label: "Create: With Spinner"},
	{Value: "create-link", Label: "Create: asChild Links"},
	{Value: "create-long-text", Label: "Create: Long Text"},
	{Value: "create-colors", Label: "Create: Custom Colors"},
}

func (p *badgeCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	// The original source links are hash-only targets. A native Router consumes
	// their real hrefs locally rather than passing a browser fragment to the OS.
	p.router = ui.NewRouter("/badge-examples")
	p.router.Transition = ui.TransitionNone
}

func (p *badgeCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *badgeCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-badge-config")
			h.Close("source-badge-language")
		}
		p.native, p.popupError = native, ""
	}
	p.router.View(c, func(_ *ui.Route) {
		display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: tr(c, "Badge compositions")}, display.CardSlots{
			Title: tr(c, "Badge compositions"), Description: tr(c, "Source variants, icons, loading cues and native links."),
			Content: func(child *ui.Context) {
				p.choose(child, h, native, "source-badge-config", &p.recipe, control.SelectProps{Label: tr(child, "Badge configuration"), Options: localizedOptions(child, badgeCompositionRecipes)})
				if p.recipe == "rtl" {
					p.choose(child, h, native, "source-badge-language", &p.language, control.SelectProps{Label: tr(child, "Badge language"), Options: localizedOptions(child, statusLanguages())})
				}
				p.draw(child.Key(p.recipe))
				if p.action != "" {
					ui.Text(child, tr(child, p.action)).FontSize(13).Role(ui.RoleStatus).Label(tr(child, p.action))
				}
				if p.popupError != "" {
					ui.Text(child, p.popupError).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
				}
			},
		})
	})
}

type badgeSourceSpec struct {
	text                     string
	variant                  display.BadgeVariant
	icon, endIcon            *ui.SVG
	spinnerStart, spinnerEnd bool
	url, palette             string
	solid                    bool
}

var badgeSourceVariants = []struct {
	text    string
	variant display.BadgeVariant
}{
	{"Default", display.BadgeDefault}, {"Secondary", display.BadgeSecondary},
	{"Destructive", display.BadgeSoft}, {"Outline", display.BadgeOutline},
	{"Ghost", display.BadgeGhost}, {"Link", display.BadgeLink},
}

func badgeSourceWords(language string) [6]string {
	switch language {
	case "ar":
		return [6]string{"شارة", "ثانوي", "مدمر", "مخطط", "متحقق", "إشارة مرجعية"}
	case "he":
		return [6]string{"תג", "משני", "הרסני", "קווי מתאר", "מאומת", "סימנייה"}
	default:
		return [6]string{"Badge", "Secondary", "Destructive", "Outline", "Verified", "Bookmark"}
	}
}

func (p *badgeCompositionPreviewState) specs() []badgeSourceSpec {
	switch p.recipe {
	case "demo", "rtl":
		words := badgeSourceWords("en")
		if p.recipe == "rtl" {
			words = badgeSourceWords(p.language)
		}
		result := []badgeSourceSpec{
			{text: words[0]}, {text: words[1], variant: display.BadgeSecondary},
			{text: words[2], variant: display.BadgeSoft}, {text: words[3], variant: display.BadgeOutline},
		}
		if p.recipe == "rtl" {
			result = append(result, badgeSourceSpec{text: words[4], variant: display.BadgeSecondary, icon: outlined.Verified()}, badgeSourceSpec{text: words[5], variant: display.BadgeOutline, endIcon: outlined.Bookmark()})
		}
		return result
	case "icon":
		return []badgeSourceSpec{{text: "Verified", variant: display.BadgeSecondary, icon: outlined.Verified()}, {text: "Bookmark", variant: display.BadgeOutline, endIcon: outlined.Bookmark()}}
	case "spinner":
		return []badgeSourceSpec{{text: "Deleting", variant: display.BadgeSoft, spinnerStart: true}, {text: "Generating", variant: display.BadgeSecondary, spinnerEnd: true}}
	case "link":
		return []badgeSourceSpec{{text: "Open Link", url: "#link", endIcon: outlined.ArrowOutward()}}
	case "colors", "create-colors":
		result := []badgeSourceSpec{}
		if p.recipe == "create-colors" {
			for i, name := range []string{"blue", "green", "sky", "purple"} {
				result = append(result, badgeSourceSpec{text: []string{"Blue", "Green", "Sky", "Purple"}[i], palette: name, solid: true})
			}
		}
		for i, name := range []string{"blue", "green", "sky", "purple", "red"} {
			result = append(result, badgeSourceSpec{text: []string{"Blue", "Green", "Sky", "Purple", "Red"}[i], palette: name})
		}
		return result
	case "create-link":
		result := []badgeSourceSpec{}
		for _, variant := range []display.BadgeVariant{display.BadgeDefault, display.BadgeSecondary, display.BadgeSoft, display.BadgeGhost} {
			result = append(result, badgeSourceSpec{text: "Link", variant: variant, url: "#", endIcon: outlined.ArrowOutward()})
		}
		return result
	case "create-long-text":
		return []badgeSourceSpec{{text: "A badge with a lot of text to see how it wraps", variant: display.BadgeSecondary}}
	default:
		result := []badgeSourceSpec{}
		for i, variant := range badgeSourceVariants {
			if p.recipe == "variants" && i == 5 {
				break
			}
			spec := badgeSourceSpec{text: variant.text, variant: variant.variant}
			switch p.recipe {
			case "create-icon-left":
				spec.icon = outlined.Verified()
			case "create-icon-right":
				spec.endIcon = outlined.ArrowForward()
			case "create-spinner":
				spec.spinnerStart = true
			}
			result = append(result, spec)
		}
		return result
	}
}

func badgeSourceColors(palette string, solid, dark bool) (ui.Color, ui.Color) {
	background, foreground := 50, 700
	if solid {
		background, foreground = 600, 50
	} else if dark {
		background, foreground = 950, 300
	}
	bg, _ := theme.Palette(palette, background)
	fg, _ := theme.Palette(palette, foreground)
	return bg, fg
}

func (p *badgeCompositionPreviewState) badge(c *ui.Context, spec badgeSourceSpec) {
	props := display.BadgeProps{Variant: spec.variant, Icon: spec.icon, EndIcon: spec.endIcon, URL: spec.url, Busy: spec.spinnerStart || spec.spinnerEnd}
	if spec.variant == display.BadgeSoft {
		// Fixed source recipes retain their original Destructive appearance.
		props.Color = &theme.TokensOf(c).Destructive
	}
	spinner := func(child *ui.Context) { display.Spinner(child, display.SpinnerProps{Size: 12, Decorative: true}) }
	if spec.spinnerStart {
		props.Leading = spinner
	}
	if spec.spinnerEnd {
		props.Trailing = spinner
	}
	if spec.url != "" {
		props.OnActivate = func() {
			p.action = "Local badge action: " + spec.text
			p.activations++
			p.lastVariant = spec.variant
		}
	}
	b := display.BadgeWith(c, tr(c, spec.text), props)
	if spec.palette != "" {
		bg, fg := badgeSourceColors(spec.palette, spec.solid, theme.TokensOf(c).Dark)
		b.Background(bg).TextColor(fg)
	}
}

func (p *badgeCompositionPreviewState) draw(c *ui.Context) {
	if p.recipe == "create-long-text" {
		// Source Badge is shrink-0/whitespace-nowrap. Keep that complete string
		// reachable in a narrow native parent rather than clipping its label.
		ui.ScrollHorizontal(c).FillWidth().MinWidth(0).Height(28).TrackScroll(&p.longScroll).Role(ui.RoleRegion).Label(tr(c, "Long badge viewport")).Children(func() {
			p.badge(c, p.specs()[0])
		})
		return
	}
	row := ui.Row(c).FillWidth().MinWidth(0).Wrap().Gap(8)
	if p.recipe == "demo" || p.recipe == "rtl" {
		row.Justify(ui.Center)
	}
	if p.recipe == "rtl" {
		row.Direction(childDirection(p.language))
	}
	row.Children(func() {
		for i, spec := range p.specs() {
			p.badge(c.Key(i), spec)
		}
	})
}
