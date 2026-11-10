package gallery

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"math"
	"net/mail"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Only application data, decoded images and public scroll/focus handles survive
// frames. Every slot is constructed with its current Card child Context.
type cardCompositionPreviewState struct {
	ready, native                   bool
	recipe, language, action, error string
	spacing                         []string
	login                           [4]cardLoginDraft
	terms                           ui.ScrollState
	termsDecision                   string
	transcribed                     bool
	landscapes                      int
	coverColor                      ui.Color
	cover                           *ui.Bitmap
}

type cardLoginDraft struct {
	email, password, error string
	submitted              bool
}

var cardCompositionRecipes = []control.Option{
	{Value: "default", Label: "Default size"},
	{Value: "small", Label: "Small size"},
	{Value: "edge", Label: "Content edge to edge"},
	{Value: "spacing", Label: "Custom spacing"},
	{Value: "header", Label: "Header with border"},
	{Value: "footer", Label: "Footer with border"},
	{Value: "header-small", Label: "Header with border (small)"},
	{Value: "footer-small", Label: "Footer with border (small)"},
	{Value: "image", Label: "With image"},
	{Value: "image-small", Label: "With image (small)"},
	{Value: "login", Label: "Login"},
	{Value: "notes", Label: "Meeting notes"},
	{Value: "docs-login", Label: "Login with header action"},
	{Value: "reports", Label: "Scheduled reports"},
	{Value: "docs-spacing", Label: "Login with custom spacing"},
	{Value: "event", Label: "Featured event"},
	{Value: "rtl", Label: "Translated RTL login"},
}

func (p *cardCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language, p.spacing = true, native, "default", "ar", []string{"16"}
}

func (p *cardCompositionPreviewState) choose(c *ui.Context, host *desktop.Host, native bool, key string, value *string, props control.SelectProps) ui.Element {
	if native {
		return desktop.NativeSelect(c, value, props)
	}
	e, err := desktop.Select(c, host, key, value, props)
	if err != nil {
		p.error = err.Error()
	}
	return e
}

func (p *cardCompositionPreviewState) view(c *ui.Context, host *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		p.native, p.error = native, ""
		if host != nil {
			host.Close("card-composition-recipe")
			host.Close("card-composition-language")
		}
	}
	display.CardWith(c, display.CardProps{Label: "Card compositions", Size: display.CardSmall}, display.CardSlots{
		Title: "Card compositions", Description: "Source layouts, real editors and independent actions.",
		Content: func(child *ui.Context) {
			p.choose(child, host, native, "card-composition-recipe", &p.recipe, control.SelectProps{Label: "Card configuration", Options: cardCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, host, native, "card-composition-language", &p.language, control.SelectProps{Label: "Card language", Options: []control.Option{{Value: "ar", Label: "العربية"}, {Value: "he", Label: "עברית"}, {Value: "en", Label: "English"}}})
			}
			if p.recipe == "spacing" || p.recipe == "docs-spacing" {
				p.spacingControl(child)
			}
			ui.Column(child).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(func() { p.card(child.Key("source-card:" + p.recipe)) })
			if p.action != "" {
				ui.Text(child, p.action).FontSize(13).Role(ui.RoleStatus)
			}
			if p.error != "" {
				ui.Text(child, p.error).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		},
	})
}

func (p *cardCompositionPreviewState) spacingControl(c *ui.Context) {
	previous := append([]string(nil), p.spacing...)
	control.ToggleGroup(c, &p.spacing, control.ToggleGroupProps{Label: "Card spacing", Outline: true, Size: control.Small,
		Items: []control.Option{{Value: "16", Label: "16px"}, {Value: "20", Label: "20px"}, {Value: "24", Label: "24px"}, {Value: "32", Label: "32px"}},
		OnValueChange: func(v []string) {
			if len(v) == 0 {
				p.spacing = previous
			}
		},
	}).AlignSelf(ui.Center)
}

func (p *cardCompositionPreviewState) card(c *ui.Context) {
	small := p.recipe == "small" || strings.HasSuffix(p.recipe, "-small") || p.recipe == "reports"
	size, buttonSize := display.CardDefault, control.Size(0)
	if small {
		size, buttonSize = display.CardSmall, control.Small
	}
	props := display.CardProps{Label: "Source card", Size: size}
	slots := display.CardSlots{}
	maxWidth := float32(384)
	paragraph := func(text string) func(*ui.Context) {
		return func(child *ui.Context) { ui.Text(child, text).FontSize(14).LineHeight(1.5) }
	}
	action := func(text, label, result string, variant control.Variant) func(*ui.Context) {
		return func(child *ui.Context) {
			if control.Button(child, text, control.ButtonProps{Label: label, Size: buttonSize, Variant: variant}).FillWidth().Clicked() {
				p.action = result
			}
		}
	}
	switch p.recipe {
	case "default", "small":
		name, description, body := "Default Card", "This card uses the default size variant.", `The card component supports a size prop that defaults to "default" for standard spacing and sizing.`
		if small {
			name, description, body = "Small Card", "This card uses the small size variant.", `The card component supports a size prop that can be set to "sm" for a more compact appearance.`
		}
		slots.Title, slots.Description, slots.Content, slots.Footer = name, description, paragraph(body), action("Action", "Run card action", "Local card action", control.Outline)
	case "header", "header-small":
		props.HeaderBorder = true
		slots.Title, slots.Description = "Header with Border", "This is a card with a header that has a bottom border."
		if small {
			slots.Description = "This is a small card with a header that has a bottom border."
		}
		slots.Content = paragraph("The header has a border-b class applied, creating a visual separation between the header and content sections.")
	case "footer", "footer-small":
		props.FooterBorder = true
		slots.Content = paragraph("The footer has a border-t class applied, creating a visual separation between the content and footer sections.")
		slots.Footer = action("Footer with Border", "Run bordered footer action", "Local bordered footer action", control.Outline)
	case "edge":
		props.ContentBleed = true
		slots.Title, slots.Description = "Terms of Service", "Review the terms before accepting the agreement."
		slots.Content = func(child *ui.Context) {
			ui.Column(child).FillWidth().Gap(0).Children(func() {
				p.termsContent(child)
				p.termsActions(child)
			})
		}
	case "spacing":
		props.Spacing = p.spacingValue()
		slots.Title, slots.Description = "Release Health", "Track readiness across launch signals."
		slots.Content = func(child *ui.Context) {
			t := theme.TokensOf(child)
			ui.Column(child).FillWidth().Gap(8).Padding(12).Radius(t.RadiusLG()).Background(t.Muted.Alpha(.5)).Children(func() {
				for _, row := range [][2]string{{"Checks passed", "24 / 26"}, {"Open blockers", "2"}} {
					ui.Row(child).FillWidth().Justify(ui.SpaceBetween).Gap(8).Children(func() { ui.Text(child, row[0]).TextColor(t.MutedForeground); ui.Text(child, row[1]).FontWeight(500) })
				}
			})
		}
	case "image", "image-small":
		slots.Title, slots.Description = "Beautiful Landscape", "A stunning view that captures the essence of natural beauty."
		slots.Media = func(child *ui.Context) { p.landscape(child) }
		slots.Footer = func(child *ui.Context) {
			if control.Button(child, "Button", control.ButtonProps{Label: "Add landscape", Size: buttonSize, Icon: outlined.Add()}).FillWidth().Clicked() {
				p.landscapes++
				p.action = fmt.Sprintf("Local landscape actions: %d", p.landscapes)
			}
		}
	case "notes":
		slots.Title, slots.Description = "Meeting Notes", "Transcript from the meeting with the client."
		slots.Action = func(child *ui.Context) {
			if control.Button(child, "Transcribe", control.ButtonProps{Variant: control.Outline, Size: control.Small, Icon: outlined.ClosedCaption()}).Clicked() {
				p.transcribed = true
				p.action = "Local transcript preview shown"
			}
		}
		slots.Content = func(child *ui.Context) {
			ui.Text(child, "Client requested dashboard redesign with focus on mobile responsiveness.").FontSize(14).LineHeight(1.5)
			ui.Column(child).FillWidth().Gap(8).Padding(4, 0, 0).PaddingStart(24).Role(ui.RoleList).Label("Meeting agenda").Children(func() {
				for i, text := range []string{"New analytics widgets for daily/weekly metrics", "Simplified navigation menu", "Dark mode support", "Timeline: 6 weeks", "Follow-up meeting scheduled for next Tuesday"} {
					ui.Text(child, fmt.Sprintf("%d. %s", i+1, text)).FontSize(14).LineHeight(1.5).Role(ui.RoleListItem).Label(text).PositionInSet(i+1, 5)
				}
			})
		}
		slots.Footer = func(child *ui.Context) {
			display.AvatarGroup(child, display.AvatarGroupProps{Label: "Meeting participants", Max: 3, Total: 11, Items: cardMeetingParticipants()})
		}
	case "login", "docs-login", "docs-spacing", "rtl":
		index := map[string]int{"login": 0, "docs-login": 1, "docs-spacing": 2, "rtl": 3}[p.recipe]
		if p.recipe == "docs-spacing" {
			props.Spacing = p.spacingValue()
		}
		slots = p.loginSlots(&p.login[index], p.recipe != "login")
	case "reports":
		maxWidth = 320
		slots.Title, slots.Description = "Scheduled reports", "Weekly snapshots. No more manual exports."
		slots.Content = func(child *ui.Context) {
			ui.Column(child).FillWidth().Gap(8).PaddingY(8).Role(ui.RoleList).Label("Scheduled report features").Children(func() {
				for i, text := range []string{"Choose a schedule (daily, or weekly).", "Send to channels or specific teammates.", "Include charts, tables, and key metrics."} {
					ui.Row(child).FillWidth().Gap(8).AlignItems(ui.Start).Role(ui.RoleListItem).Label(text).PositionInSet(i+1, 3).Children(func() {
						ui.Column(child).Shrink(0).Padding(2, 0, 0).AccessHidden(true).Children(func() {
							material.Icon(child, outlined.ChevronRight(), material.Props{Size: 16}).TextColor(theme.TokensOf(child).MutedForeground)
						})
						ui.Text(child, text).MinWidth(0).Grow(1).FontSize(14).LineHeight(1.5).AccessHidden(true)
					})
				}
			})
		}
		slots.Footer = func(child *ui.Context) {
			ui.Column(child).FillWidth().Gap(8).Children(func() {
				action("Set up scheduled reports", "Set up scheduled reports", "Local scheduled-report setup", control.Default)(child)
				action("See what's new", "See what's new", "Local report release notes", control.Outline)(child)
			})
		}
	case "event":
		slots.Title, slots.Description = "Design systems meetup", "A practical talk on component APIs, accessibility, and shipping faster."
		slots.Action = func(child *ui.Context) { display.Badge(child, "Featured", display.BadgeSecondary) }
		slots.Media = func(child *ui.Context) {
			cover := cardEventLight
			if theme.TokensOf(child).Dark {
				cover = cardEventDark
			}
			ui.Image(child, cover).FillWidth().AspectRatio(16.0 / 9).Fit(ui.Cover).Role(ui.RoleImage).Label("Event cover")
		}
		slots.Footer = action("View Event", "View Event", "Local event preview: Design systems meetup", control.Default)
	}
	build := func(current *ui.Context) { display.CardWith(current, props, slots).MaxWidth(maxWidth) }
	if p.recipe == "rtl" {
		ui.Column(c).FillWidth().MinWidth(0).AlignItems(ui.Center).Direction(childDirection(p.language)).Children(func() { build(c) })
	} else {
		build(c)
	}
}

func childDirection(language string) ui.Direction {
	if language == "ar" || language == "he" {
		return ui.RightToLeft
	}
	return ui.LeftToRight
}
func (p *cardCompositionPreviewState) spacingValue() float32 {
	if len(p.spacing) > 0 {
		v, _ := strconv.Atoi(p.spacing[0])
		if v == 16 || v == 20 || v == 24 || v == 32 {
			return float32(v)
		}
	}
	return 16
}

var cardTerms = []string{
	"These terms govern your use of the workspace, including access to shared documents, project files, and collaboration tools.",
	"You are responsible for the content you upload and for ensuring that your team has the appropriate permissions to view or edit it.",
	"We may update features or limits as the service evolves. When those changes materially affect your workflow, we will notify your workspace administrators.",
	"By continuing, you agree to keep your account credentials secure and to follow your organization's acceptable use policies.",
}

func (p *cardCompositionPreviewState) termsActions(child *ui.Context) {
	ui.Row(child).FillWidth().PaddingX(24).Justify(ui.End).Gap(8).Children(func() {
		for _, x := range []struct {
			text    string
			variant control.Variant
		}{{"Decline", control.Outline}, {"Accept", control.Default}} {
			if control.Button(child, x.text, control.ButtonProps{Variant: x.variant}).Clicked() {
				p.termsDecision = x.text
				p.action = "Local agreement decision: " + x.text
			}
		}
	})
}

func (p *cardCompositionPreviewState) termsContent(c *ui.Context) {
	t := theme.TokensOf(c)
	ui.Scroll(c).FillWidth().MaxHeight(192).TrackScroll(&p.terms).Scrollbars(true).Background(t.Muted.Alpha(.5)).Border(1, t.Border).BorderWidth(1, 0, 0).Children(func() {
		ui.Column(c).FillWidth().Padding(16, 24).Gap(16).Children(func() {
			for _, text := range cardTerms {
				ui.Text(c, text).FontSize(14).LineHeight(1.625)
			}
		})
	})
}

type cardLoginWords struct{ title, description, signup, email, password, forgot, login, google string }

func (p *cardCompositionPreviewState) loginWords() cardLoginWords {
	if p.recipe == "rtl" {
		switch p.language {
		case "ar":
			return cardLoginWords{"تسجيل الدخول إلى حسابك", "أدخل بريدك الإلكتروني أدناه لتسجيل الدخول إلى حسابك", "إنشاء حساب", "البريد الإلكتروني", "كلمة المرور", "نسيت كلمة المرور؟", "تسجيل الدخول", "تسجيل الدخول باستخدام Google"}
		case "he":
			return cardLoginWords{"התחבר לחשבון שלך", "הזן את האימייל שלך למטה כדי להתחבר לחשבון שלך", "הירשם", "אימייל", "סיסמה", "שכחת את הסיסמה?", "התחבר", "התחבר עם Google"}
		}
	}
	return cardLoginWords{"Login to your account", "Enter your email below to login to your account", "Sign Up", "Email", "Password", "Forgot your password?", "Login", "Login with Google"}
}
func (p *cardCompositionPreviewState) loginSlots(d *cardLoginDraft, headerAction bool) display.CardSlots {
	w := p.loginWords()
	slots := display.CardSlots{Title: w.title, Description: w.description}
	if headerAction {
		slots.Action = func(c *ui.Context) {
			if control.Button(c, w.signup, control.ButtonProps{Variant: control.Link}).Clicked() {
				p.action = "Local sign-up preview"
			}
		}
	}
	slots.Content = func(c *ui.Context) {
		ui.Column(c.Key("login-fields")).FillWidth().Gap(24).Children(func() {
			control.Field(c, control.FieldProps{Label: w.email, Required: true, Error: d.error}, func() ui.Element {
				return control.Input(c, &d.email, control.InputProps{Placeholder: "m@example.com", OnSubmit: func(string) { p.submitLogin(d) }})
			})
			ui.Column(c).FillWidth().Gap(8).Children(func() {
				var password, label ui.Element
				row := ui.Row(c).FillWidth().Justify(ui.SpaceBetween).Gap(8).AlignItems(ui.Center)
				row.Children(func() {
					label = control.Label(c, w.password)
					if control.Button(c, w.forgot, control.ButtonProps{Variant: control.Link, Size: control.Small}).Height(20).Padding(0).Clicked() {
						p.action = "Local password-recovery preview"
					}
				})
				password = control.Input(c, &d.password, control.InputProps{Label: w.password, Password: true, OnSubmit: func(string) { p.submitLogin(d) }})
				password.LabelledBy(label)
			})
		})
	}
	slots.Footer = func(c *ui.Context) {
		ui.Column(c).FillWidth().Gap(8).Children(func() {
			if control.Button(c, w.login, control.ButtonProps{}).FillWidth().Clicked() {
				p.submitLogin(d)
			}
			if control.Button(c, w.google, control.ButtonProps{Variant: control.Outline}).FillWidth().Clicked() {
				p.action = "Local Google sign-in preview"
			}
			if !headerAction {
				ui.Row(c).FillWidth().Justify(ui.Center).Gap(4).Padding(16, 0, 0).Wrap().Children(func() {
					ui.Text(c, "Don't have an account?").FontSize(14)
					if control.Button(c, "Sign up", control.ButtonProps{Variant: control.Link, Size: control.Small}).Height(20).Padding(0).Clicked() {
						p.action = "Local sign-up preview"
					}
				})
			}
		})
	}
	return slots
}
func (p *cardCompositionPreviewState) submitLogin(d *cardLoginDraft) {
	address, err := mail.ParseAddress(strings.TrimSpace(d.email))
	if err != nil || address.Address != strings.TrimSpace(d.email) || d.password == "" {
		d.error = "Enter an email address and password to preview login."
		d.submitted = false
		return
	}
	d.error = ""
	d.submitted = true
	p.action = "Local login preview: " + address.Address
}

// Original bytes are downloaded from the exact URLs in the fixed examples.
// They are decoded once, with no runtime network request. The source map keeps
// URLs, response types, byte hashes and the photographer attribution.
//
//go:embed card-assets/landscape.jpg
var cardLandscapeBytes []byte

//go:embed card-assets/event.png
var cardEventBytes []byte

//go:embed card-assets/shadcn.jpg
var cardShadcnBytes []byte

//go:embed card-assets/maxleiter.png
var cardMaxleiterBytes []byte

//go:embed card-assets/evilrabbit.png
var cardEvilrabbitBytes []byte

func cardDecodeImage(data []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return img
}
func cardDecodeBitmap(data []byte) *ui.Bitmap { return ui.NewBitmap(cardDecodeImage(data)) }

var cardLandscapeImage = cardDecodeImage(cardLandscapeBytes)
var cardEventImage = cardDecodeImage(cardEventBytes)
var cardEventLight = cardFilteredImage(cardEventImage, .6*.65, nil)
var cardEventDark = cardFilteredImage(cardEventImage, .4*.65, nil)
var cardMeetingImages = []*ui.Bitmap{cardDecodeBitmap(cardShadcnBytes), cardDecodeBitmap(cardMaxleiterBytes), cardDecodeBitmap(cardEvilrabbitBytes)}

func cardMeetingParticipants() []display.AvatarProps {
	return []display.AvatarProps{{Key: "shadcn", Label: "@shadcn", Fallback: "CN", Image: cardMeetingImages[0]}, {Key: "maxleiter", Label: "@maxleiter", Fallback: "LR", Image: cardMeetingImages[1]}, {Key: "evilrabbit", Label: "@evilrabbit", Fallback: "ER", Image: cardMeetingImages[2]}}
}

func (p *cardCompositionPreviewState) landscape(c *ui.Context) {
	primary := theme.TokensOf(c).Primary
	if p.cover == nil || p.coverColor != primary {
		p.coverColor = primary
		p.cover = cardFilteredImage(cardLandscapeImage, .6, &primary)
	}
	ui.Image(c, p.cover).FillWidth().AspectRatio(16.0 / 9).Fit(ui.Cover).Role(ui.RoleImage).Label("Photo by mymind on Unsplash")
}

// CSS grayscale + brightness are applied to the actual image pixels. The
// optional overlay implements 50% mix-blend-color using SetLum/ClipColor,
// preserving backdrop luminosity before opacity compositing. A parent's GPU
// filter would not reliably affect its image child on all native renderers.
func cardFilteredImage(source image.Image, brightness float64, overlay *ui.Color) *ui.Bitmap {
	out := image.NewNRGBA(source.Bounds())
	for y := out.Rect.Min.Y; y < out.Rect.Max.Y; y++ {
		for x := out.Rect.Min.X; x < out.Rect.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
			lum := (.2126*float64(pixel.R) + .7152*float64(pixel.G) + .0722*float64(pixel.B)) / 255 * brightness
			rgb := [3]float64{lum, lum, lum}
			if overlay != nil {
				blend := cardSetLum([3]float64{float64(overlay.R) / 255, float64(overlay.G) / 255, float64(overlay.B) / 255}, lum)
				a := .5 * float64(overlay.A) / 255
				for i := range rgb {
					rgb[i] = rgb[i]*(1-a) + blend[i]*a
				}
			}
			out.SetNRGBA(x, y, color.NRGBA{R: cardByte(rgb[0]), G: cardByte(rgb[1]), B: cardByte(rgb[2]), A: pixel.A})
		}
	}
	return ui.NewBitmap(out)
}
func cardByte(v float64) uint8 { return uint8(math.Round(max(0, min(1, v)) * 255)) }
func cardSetLum(rgb [3]float64, lum float64) [3]float64 {
	d := lum - (.3*rgb[0] + .59*rgb[1] + .11*rgb[2])
	for i := range rgb {
		rgb[i] += d
	}
	lo, hi := min(rgb[0], rgb[1], rgb[2]), max(rgb[0], rgb[1], rgb[2])
	if lo < 0 {
		for i := range rgb {
			rgb[i] = lum + (rgb[i]-lum)*lum/(lum-lo)
		}
	}
	if hi > 1 {
		for i := range rgb {
			rgb[i] = lum + (rgb[i]-lum)*(1-lum)/(hi-lum)
		}
	}
	return rgb
}
