package gallery

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type carouselPreviewState struct {
	initialized                           bool
	layout                                string
	compact, loop, short                  bool
	horizontal, vertical, rtl             int
	order                                 []int
	notes                                 [5]string
	horizontalView, verticalView, rtlView display.CarouselSnapshot
}

var carouselProjectNames = [...]string{"Discover", "Create", "Review", "Publish", "Share"}

func (s *State) carouselPreview(c *ui.Context) {
	p := &s.carousels
	if !p.initialized {
		p.initialized, p.layout = true, "half"
		p.order = []int{0, 1, 2, 3, 4}
	}
	ui.Column(c).FillWidth().MinWidth(0).Shrink(0).Gap(20).Children(func() {
		display.Card(c.Key("carousel-explorer"), func() {
			s.cardHeader(c, "Carousel", "Browse projects, change the layout and keep a note with each card.")
			control.RadioGroup(c, &p.layout, control.RadioGroupProps{Label: "Carousel layout", Horizontal: true, Options: []control.Option{
				{Value: "full", Label: "One card"}, {Value: "half", Label: "Two cards"}, {Value: "third", Label: "Three cards"},
				{Value: "responsive", Label: "Responsive cards"}, {Value: "fixed", Label: "Fixed card width"},
			}}).Wrap().MinWidth(0)
			ui.Row(c).FillWidth().Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
				control.Switch(c, &p.compact, control.SwitchProps{Label: "Compact card spacing", Size: control.Small})
				control.Switch(c, &p.loop, control.SwitchProps{Label: "Loop project cards", Size: control.Small})
				control.Switch(c, &p.short, control.SwitchProps{Label: "Show only two projects", Size: control.Small})
				if control.Button(c, "Reverse projects", control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
					slices.Reverse(p.order)
					c.Invalidate()
				}
			})
			spacing := float32(16)
			if p.compact {
				spacing = 4
			}
			basis := display.CarouselFraction(.5)
			switch p.layout {
			case "full":
				basis = display.CarouselBasis{}
			case "third":
				basis = display.CarouselFraction(1.0 / 3)
			case "fixed":
				basis = display.CarouselFixed(220)
			case "responsive":
				width, _ := c.Size()
				basis = display.CarouselBasis{}
				if width >= 1024 {
					basis = display.CarouselFraction(1.0 / 3)
				} else if width >= 640 {
					basis = display.CarouselFraction(.5)
				}
			}
			count := len(p.order)
			if p.short {
				count = 2
			}
			ui.Box(c).FillWidth().MinWidth(0).Padding(8, 48).Children(func() {
				display.CarouselContainer(c, &p.horizontal, display.CarouselContainerProps{
					Key: "project-cards", Label: "Project carousel", Count: count, Height: 180,
					Align: display.CarouselAlignStart, Loop: p.loop,
				}, func(c *ui.Context) {
					display.CarouselContent(c, display.CarouselContentProps{Spacing: &spacing}, func(c *ui.Context) {
						for index, id := range p.order[:count] {
							display.CarouselItem(c, index, display.CarouselItemProps{Key: id, Basis: basis}, func(c *ui.Context) {
								s.carouselProjectCard(c, id, true)
							})
						}
					})
					display.CarouselPrevious(c, display.CarouselControlProps{Label: "Previous project"})
					display.CarouselNext(c, display.CarouselControlProps{Label: "Next project"})
					p.horizontalView = display.CarouselSnapshotOf(c)
				})
			})
			carouselPreviewPosition(c, p.horizontalView)
			if p.loop && !p.horizontalView.Loop {
				ui.Text(c, "This layout needs more projects to loop.").FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
			}
		}).Label("Carousel explorer").Shrink(0)
		columns(c, func() {
			display.Card(c.Key("carousel-vertical"), func() {
				s.cardHeader(c, "Vertical carousel", "Use the arrows above and below to browse the collection.")
				spacing := float32(4)
				ui.Box(c).FillWidth().MinWidth(0).Padding(52, 24).Children(func() {
					display.CarouselContainer(c, &p.vertical, display.CarouselContainerProps{
						Key: "vertical-projects", Label: "Vertical project carousel", Count: 5, Height: 270,
						Vertical: true, Align: display.CarouselAlignStart,
					}, func(c *ui.Context) {
						display.CarouselContent(c, display.CarouselContentProps{Spacing: &spacing}, func(c *ui.Context) {
							for id := range 5 {
								display.CarouselItem(c, id, display.CarouselItemProps{Key: id, Basis: display.CarouselFraction(.5)}, func(c *ui.Context) { s.carouselProjectCard(c, id, false) })
							}
						})
						display.CarouselPrevious(c, display.CarouselControlProps{Label: "Previous vertical project"})
						display.CarouselNext(c, display.CarouselControlProps{Label: "Next vertical project"})
						p.verticalView = display.CarouselSnapshotOf(c)
					})
				})
				carouselPreviewPosition(c, p.verticalView)
			}).Label("Vertical carousel example").Shrink(0)
		}, func() {
			display.Card(c.Key("carousel-rtl"), func() {
				s.cardHeader(c, "Right-to-left carousel", "The next card follows the reading direction.")
				ui.Box(c).FillWidth().MinWidth(0).Padding(8, 48).Direction(ui.RightToLeft).Children(func() {
					display.CarouselContainer(c, &p.rtl, display.CarouselContainerProps{
						Key: "rtl-projects", Label: "RTL project carousel", Count: 5, Height: 180,
						Align: display.CarouselAlignStart, Loop: true,
					}, func(c *ui.Context) {
						display.CarouselContent(c, display.CarouselContentProps{}, func(c *ui.Context) {
							for id := range 5 {
								display.CarouselItem(c, id, display.CarouselItemProps{Key: id}, func(c *ui.Context) { s.carouselProjectCard(c, id, false) })
							}
						})
						display.CarouselPrevious(c, display.CarouselControlProps{Label: "Previous RTL project"})
						display.CarouselNext(c, display.CarouselControlProps{Label: "Next RTL project"})
						p.rtlView = display.CarouselSnapshotOf(c)
					})
				})
				carouselPreviewPosition(c, p.rtlView)
			}).Label("RTL carousel example").Shrink(0)
		})
	})
}

func (s *State) carouselProjectCard(c *ui.Context, id int, editable bool) {
	t := theme.TokensOf(c)
	ui.Column(c).Fill().MinWidth(0).Padding(16).Gap(10).Radius(t.RadiusMD()).Border(1, t.Border).Background(t.Card).Children(func() {
		ui.Text(c, fmt.Sprintf("%02d", id+1)).FontSize(24).FontWeight(600)
		ui.Text(c, carouselProjectNames[id]).FontWeight(600).SingleLine()
		if editable {
			control.Input(c, &s.carousels.notes[id], control.InputProps{Label: "Notes for " + carouselProjectNames[id], Placeholder: "Add a note…", Size: control.Small}).MinWidth(0)
		}
	})
}

func carouselPreviewPosition(c *ui.Context, snapshot display.CarouselSnapshot) {
	ui.Text(c, fmt.Sprintf("Position %d of %d · %d projects", snapshot.SelectedSnap+1, len(snapshot.Snaps), snapshot.SlideCount)).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
}
