package gallery

import (
	"fmt"
	"image"
	"image/color"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// These are application-owned values for the small source composition recipes.
// The existing plan, volume, avatar modes and loading preview remain separate.
type recipeCompositionState struct {
	ready      bool
	progress   float64
	plan, size string
	grayscale  bool
}

func (p *recipeCompositionState) initialize() {
	if p.ready {
		return
	}
	p.ready = true
	p.progress = 50
	p.plan = "plus"
	p.size = "medium"
	p.grayscale = true
}

func (s *State) recipeChoicesPreview(c *ui.Context) {
	p := &s.recipes
	p.initialize()
	columns(c, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Choose a workspace plan", "Select the whole card or use the arrow keys.")
			ui.RadioGroup(c, func() {
				for _, option := range []control.Option{{Value: "plus", Label: "Plus", Description: "For individuals and small teams."}, {Value: "pro", Label: "Pro", Description: "For growing businesses."}, {Value: "enterprise", Label: "Enterprise", Description: "For large teams and enterprises."}} {
					e := control.RadioWith(c.Key(option.Value), &p.plan, option.Value, option.Label, control.RadioProps{Description: option.Description, Content: func(child *ui.Context) {
						ui.Column(child).Gap(4).Children(func() {
							ui.Text(child, option.Label).FontWeight(500)
							ui.Text(child, option.Description).FontSize(12).TextColor(theme.TokensOf(child).MutedForeground)
						})
					}})
					recipeRadioCard(c, e, p.plan == option.Value).Reverse().Justify(ui.SpaceBetween)
				}
			}).FillWidth().MaxWidth(384).Gap(8).Role(ui.RoleRadioGroup).Label("Workspace plan cards").FocusGroup(ui.Horizontal | ui.Vertical)
			ui.Text(c, "Selected plan: "+recipePlanLabel(p.plan)).FontSize(12).Role(ui.RoleStatus)
		})
	}, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Choose a size", "One radio group, arranged in two columns.")
			ui.RadioGroup(c, func() {
				for _, option := range []control.Option{{Value: "small", Label: "Small"}, {Value: "medium", Label: "Medium"}, {Value: "large", Label: "Large"}, {Value: "xlarge", Label: "X-Large"}} {
					e := control.RadioWith(c.Key(option.Value), &p.size, option.Value, option.Label, control.RadioProps{})
					recipeRadioCard(c, e, p.size == option.Value).Padding(14, 12)
				}
			}).Grid().Columns(2).FillWidth().MaxWidth(384).Gap(8).Role(ui.RoleRadioGroup).Label("Preview size grid").FocusGroup(ui.Horizontal | ui.Vertical)
			ui.Text(c, "Selected size: "+recipeSizeLabel(p.size)).FontSize(12).Role(ui.RoleStatus)
		})
	})
}
func recipeRadioCard(c *ui.Context, e ui.Element, selected bool) ui.Element {
	tokens := theme.TokensOf(c)
	e.FillWidth().AlignSelf(ui.Stretch).AlignItems(ui.Center).Padding(16).Radius(tokens.RadiusLG()).Border(1, tokens.Border)
	if selected {
		e.BorderColor(tokens.Primary).Background(tokens.Accent.Alpha(.35))
	}
	if e.FocusVisible() {
		e.BorderColor(tokens.Ring)
	}
	return e
}
func recipePlanLabel(value string) string {
	switch value {
	case "plus":
		return "Plus"
	case "pro":
		return "Pro"
	default:
		return "Enterprise"
	}
}
func recipeSizeLabel(value string) string {
	switch value {
	case "small":
		return "Small"
	case "medium":
		return "Medium"
	case "large":
		return "Large"
	default:
		return "X-Large"
	}
}

func (s *State) recipeProgressPreview(c *ui.Context) {
	p := &s.recipes
	p.initialize()
	display.Card(c, func() {
		s.cardHeader(c, "Progress, with context", "A label and value remain visible beside the progress bar.")
		recipeLabeledProgress(c, "Upload progress", 56)
		display.Separator(c, false)
		recipeLabeledProgress(c, "Transfer progress", p.progress)
		control.Slider(c, &p.progress, control.SliderProps{Min: 0, Max: 100, Step: 1, Label: "Adjust transfer progress"}).MaxWidth(384)
	})
}
func recipeLabeledProgress(c *ui.Context, label string, value float64) {
	ui.Column(c).FillWidth().MaxWidth(384).Gap(8).Children(func() {
		ui.Row(c).FillWidth().Justify(ui.SpaceBetween).Gap(12).Children(func() {
			ui.Text(c, label).FontSize(14).FontWeight(500)
			ui.Text(c, fmt.Sprintf("%.0f%%", value)).FontSize(14).Font("monospace").FontFeatures("tnum")
		})
		display.Progress(c, value/100, display.ProgressProps{Label: label})
	})
}

// Avatar's image is intentionally private to that component, and Grayscale on
// its returned Box does not filter descendants. This recipe supplies a genuinely
// gray decoded image through the public ImageSource field instead. The local
// SVG is decoded at 2x its 80-DIP original; alpha survives image preprocessing.
var recipeGrayPortrait = recipeGrayImage(avatarPortrait)

func recipeGrayImage(source *ui.SVG) *ui.Bitmap {
	original := source.Image(160, 160, nil)
	gray := image.NewNRGBA(original.Bounds())
	for y := gray.Rect.Min.Y; y < gray.Rect.Max.Y; y++ {
		for x := gray.Rect.Min.X; x < gray.Rect.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(original.At(x, y)).(color.NRGBA)
			luma := uint8((2126*uint32(pixel.R) + 7152*uint32(pixel.G) + 722*uint32(pixel.B) + 5000) / 10000)
			gray.SetNRGBA(x, y, color.NRGBA{R: luma, G: luma, B: luma, A: pixel.A})
		}
	}
	return ui.NewBitmap(gray)
}
func (s *State) recipeAvatarPreview(c *ui.Context) {
	p := &s.recipes
	p.initialize()
	display.Card(c, func() {
		s.cardHeader(c, "Portrait color", "The same local portrait, in color or grayscale.")
		control.Switch(c, &p.grayscale, control.SwitchProps{Label: "Grayscale portrait", Size: control.Small})
		var portrait ui.ImageSource = avatarPortrait
		caption := "Color portrait"
		if p.grayscale {
			portrait = recipeGrayPortrait
			caption = "Grayscale portrait"
		}
		ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
			display.Avatar(c, display.AvatarProps{Key: "recipe-portrait", Image: portrait, Fallback: "SD", Label: "Sofia Davis portrait", Scale: display.AvatarLarge})
			ui.Text(c, caption).FontSize(14)
		})
	})
}

func (s *State) recipeSkeletonPreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "Loading forms and tables", "Placeholders keep a familiar shape while content loads.")
		control.Switch(c, &s.loadingStatic, control.SwitchProps{Label: "Pause skeleton animations", Size: control.Small})
		ui.Row(c).FillWidth().Wrap().Gap(32).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Width(320).MaxWidthPercent(100).Gap(16).Children(func() {
				ui.Text(c, "Form").FontSize(14).FontWeight(500)
				ui.Column(c).FillWidth().Gap(28).Role(ui.RoleGroup).Label("Form loading preview").Busy(true).Children(func() {
					for _, labelWidth := range []float32{80, 96} {
						ui.Column(c).FillWidth().Gap(12).Children(func() {
							display.Skeleton(c, display.SkeletonProps{Width: labelWidth, Height: 16, Static: s.loadingStatic}).AccessHidden(true)
							display.Skeleton(c, display.SkeletonProps{Height: 32, Static: s.loadingStatic}).AccessHidden(true)
						})
					}
					display.Skeleton(c, display.SkeletonProps{Width: 96, Height: 32, Static: s.loadingStatic}).AccessHidden(true)
				})
			})
			ui.Column(c).Width(384).MaxWidthPercent(100).Gap(16).Children(func() {
				ui.Text(c, "Table").FontSize(14).FontWeight(500)
				ui.Column(c).FillWidth().Gap(8).Role(ui.RoleGroup).Label("Table loading preview").Busy(true).Children(func() {
					for i := 0; i < 5; i++ {
						ui.Row(c).FillWidth().Gap(16).Children(func() {
							display.Skeleton(c, display.SkeletonProps{Height: 16, Static: s.loadingStatic}).Basis(0).Grow(1).MinWidth(0).AccessHidden(true)
							display.Skeleton(c, display.SkeletonProps{Width: 96, Height: 16, Static: s.loadingStatic}).Shrink(0).AccessHidden(true)
							display.Skeleton(c, display.SkeletonProps{Width: 80, Height: 16, Static: s.loadingStatic}).Shrink(0).AccessHidden(true)
						})
					}
				})
			})
		})
	})
}
