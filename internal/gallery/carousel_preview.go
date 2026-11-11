package gallery

import (
	"fmt"
	"slices"
	"time"

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
	controllers                           [4]*display.CarouselController
	autoplay                              *display.CarouselAutoplay
	autoplayRoot                          ui.Handle
	autoplayHovered                       bool
	autoplaySelected                      int
	autoplayView                          display.CarouselSnapshot
	alignment                             string
	attraction                            string
	free, skip                            bool
	selectionEvents                       int
}

var carouselProjectNames = [...]string{"Discover", "Create", "Review", "Publish", "Share"}

func (s *State) carouselPreview(c *ui.Context) {
	p := &s.carousels
	if !p.initialized {
		p.initialized, p.layout = true, "half"
		p.order = []int{0, 1, 2, 3, 4}
		p.attraction = "25"
		p.alignment = "start"
	}
	ensure := func(index, selected int) *display.CarouselController {
		if p.controllers[index] == nil || p.controllers[index].Snapshot().Destroyed {
			p.controllers[index], _ = display.NewCarouselController(display.CarouselOptions{StartIndex: display.CarouselValue(selected)})
			if index == 0 {
				p.controllers[index].On(display.CarouselSelect, display.NewCarouselListener(func(*display.CarouselController, display.CarouselEventType) { p.selectionEvents++ }))
			}
		}
		return p.controllers[index]
	}
	horizontal := ensure(0, p.horizontal)
	vertical := ensure(1, p.vertical)
	rtl := ensure(2, p.rtl)
	autoplayController := ensure(3, p.autoplaySelected)
	if p.autoplay == nil {
		p.autoplay, _ = display.NewCarouselAutoplay(display.CarouselAutoplayOptions{Delay: display.CarouselValue(2 * time.Second), StopOnInteraction: display.CarouselValue(true)})
	}
	attraction := 25.
	switch p.attraction {
	case "0":
		attraction = 0
	case "40":
		attraction = 40
	}
	align := display.CarouselAlignStart
	if p.alignment == "center" {
		align = display.CarouselAlignCenter
	}
	horizontalOptions := display.CarouselOptions{Align: &display.CarouselAlignment{Mode: align}, Loop: &p.loop, Duration: &attraction, DragFree: &p.free, SkipSnaps: &p.skip}
	ui.Column(c).FillWidth().MinWidth(0).Shrink(0).Gap(20).Children(func() {
		display.Card(c.Key("carousel-explorer"), func() {
			s.cardHeader(c, "Carousel", "Browse projects, change the layout and keep a note with each card.")
			control.RadioGroup(c, &p.layout, control.RadioGroupProps{Label: tr(c, "Carousel layout"), Horizontal: true, Options: localizedOptions(c, []control.Option{
				{Value: "full", Label: "One card"}, {Value: "half", Label: "Two cards"}, {Value: "third", Label: "Three cards"},
				{Value: "responsive", Label: "Responsive cards"}, {Value: "size", Label: "Responsive item size"}, {Value: "fixed", Label: "Fixed card width"},
			})}).Wrap().MinWidth(0)
			ui.Row(c).FillWidth().Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
				control.Switch(c, &p.compact, control.SwitchProps{Label: tr(c, "Compact card spacing"), Size: control.Small})
				control.Switch(c, &p.loop, control.SwitchProps{Label: tr(c, "Loop project cards"), Size: control.Small})
				control.Switch(c, &p.short, control.SwitchProps{Label: tr(c, "Show only two projects"), Size: control.Small})
				if control.Button(c, tr(c, "Reverse projects"), control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
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
			case "size":
				width, _ := c.Size()
				basis = display.CarouselFraction(.5)
				if width >= 1024 {
					basis = display.CarouselFraction(1.0 / 3)
				}
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
					Key: "project-cards", Label: tr(c, "Project carousel"), Count: count, Height: 180,
					Align: align, Controller: horizontal, Options: &horizontalOptions,
				}, func(c *ui.Context) {
					display.CarouselContent(c, display.CarouselContentProps{Spacing: &spacing}, func(c *ui.Context) {
						for index, id := range p.order[:count] {
							display.CarouselItem(c, index, display.CarouselItemProps{Key: id, Basis: basis}, func(c *ui.Context) {
								s.carouselProjectCard(c, id, true)
							})
						}
					})
					display.CarouselPrevious(c, display.CarouselControlProps{Label: tr(c, "Previous project")})
					display.CarouselNext(c, display.CarouselControlProps{Label: tr(c, "Next project")})
					p.horizontalView = display.CarouselSnapshotOf(c)
				})
			})
			carouselPreviewPosition(c, p.horizontalView)
			ui.Row(c).FillWidth().Wrap().Gap(12).AlignItems(ui.Center).Children(func() {
				control.RadioGroup(c, &p.attraction, control.RadioGroupProps{Label: tr(c, "Carousel motion"), Horizontal: true, Options: localizedOptions(c, []control.Option{{Value: "25", Label: "Default motion"}, {Value: "40", Label: "Gentler motion"}, {Value: "0", Label: "Instant motion"}})}).Wrap()
				control.RadioGroup(c, &p.alignment, control.RadioGroupProps{Label: tr(c, "Card alignment"), Horizontal: true, Options: localizedOptions(c, []control.Option{{Value: "start", Label: "Start alignment"}, {Value: "center", Label: "Centered alignment"}})})
				control.Switch(c, &p.free, control.SwitchProps{Label: tr(c, "Free dragging"), Size: control.Small})
				control.Switch(c, &p.skip, control.SwitchProps{Label: tr(c, "Skip snap positions"), Size: control.Small})
			})
			ui.Row(c).Wrap().Gap(8).Children(func() {
				if control.Button(c, tr(c, "Jump to first"), control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
					horizontal.ScrollTo(0, true)
				}
				if control.Button(c, tr(c, "Jump to last"), control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
					horizontal.ScrollTo(len(horizontal.ScrollSnapList())-1, true)
				}
				ui.Text(c, fmt.Sprintf(tr(c, "%d selections · %.0f%% through"), p.selectionEvents, horizontal.ScrollProgress()*100)).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
			})

			if p.loop && !p.horizontalView.Loop {
				ui.Text(c, tr(c, "This layout needs more projects to loop.")).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
			}
		}).Label(tr(c, "Carousel explorer")).Shrink(0)

		display.Card(c.Key("carousel-autoplay"), func() {
			s.cardHeader(c, "Autoplay carousel", "Cards advance every two seconds. Hover or interact to pause, then press Play to resume.")
			outer := ui.Box(c).FillWidth().MinWidth(0).Padding(8, 48).Bind(&p.autoplayRoot)
			hovered := outer.Hovered()
			if hovered != p.autoplayHovered {
				p.autoplayHovered = hovered
				if hovered {
					p.autoplay.Stop()
				} else {
					p.autoplay.Reset()
				}
			}
			outer.Children(func() {
				display.CarouselContainer(c, &p.autoplaySelected, display.CarouselContainerProps{Key: "autoplay-projects", Label: tr(c, "Autoplay project carousel"), Count: 5, Height: 160, Controller: autoplayController, Plugins: []display.CarouselPlugin{p.autoplay}, Align: display.CarouselAlignStart}, func(c *ui.Context) {
					display.CarouselContent(c, display.CarouselContentProps{}, func(c *ui.Context) {
						for id := range 5 {
							display.CarouselItem(c, id, display.CarouselItemProps{Key: id}, func(c *ui.Context) { s.carouselProjectCard(c, id, false) })
						}
					})
					display.CarouselPrevious(c, display.CarouselControlProps{Label: tr(c, "Previous autoplay project")})
					display.CarouselNext(c, display.CarouselControlProps{Label: tr(c, "Next autoplay project")})
					p.autoplayView = display.CarouselSnapshotOf(c)
				})
			})
			ui.Row(c).Wrap().Gap(8).Children(func() {
				label := "Play autoplay"
				if p.autoplay.IsPlaying() {
					label = "Pause autoplay"
				}
				if control.Button(c, tr(c, label), control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
					if p.autoplay.IsPlaying() {
						p.autoplay.Stop()
					} else {
						p.autoplay.Play()
					}
				}
				if control.Button(c, tr(c, "Restart autoplay timer"), control.ButtonProps{Variant: control.Ghost, Size: control.Small}).Clicked() {
					p.autoplay.Reset()
				}
				if control.Button(c, tr(c, "Play with instant transitions"), control.ButtonProps{Variant: control.Ghost, Size: control.Small}).Clicked() {
					p.autoplay.Play(true)
				}
			})
			carouselPreviewPosition(c, p.autoplayView)
		}).Label(tr(c, "Autoplay carousel example")).Shrink(0)
		columns(c, func() {
			display.Card(c.Key("carousel-vertical"), func() {
				s.cardHeader(c, "Vertical carousel", "Use the arrows above and below to browse the collection.")
				spacing := float32(4)
				ui.Box(c).FillWidth().MinWidth(0).Padding(52, 24).Children(func() {
					display.CarouselContainer(c, &p.vertical, display.CarouselContainerProps{
						Key: "vertical-projects", Label: tr(c, "Vertical project carousel"), Count: 5, Height: 270,
						Vertical: true, Align: display.CarouselAlignStart, Controller: vertical,
					}, func(c *ui.Context) {
						display.CarouselContent(c, display.CarouselContentProps{Spacing: &spacing}, func(c *ui.Context) {
							for id := range 5 {
								display.CarouselItem(c, id, display.CarouselItemProps{Key: id, Basis: display.CarouselFraction(.5)}, func(c *ui.Context) { s.carouselProjectCard(c, id, false) })
							}
						})
						display.CarouselPrevious(c, display.CarouselControlProps{Label: tr(c, "Previous vertical project")})
						display.CarouselNext(c, display.CarouselControlProps{Label: tr(c, "Next vertical project")})
						p.verticalView = display.CarouselSnapshotOf(c)
					})
				})
				carouselPreviewPosition(c, p.verticalView)
			}).Label(tr(c, "Vertical carousel example")).Shrink(0)
		}, func() {
			display.Card(c.Key("carousel-rtl"), func() {
				s.cardHeader(c, "Right-to-left carousel", "The next card follows the reading direction.")
				ui.Box(c).FillWidth().MinWidth(0).Padding(8, 48).Direction(ui.RightToLeft).Children(func() {
					display.CarouselContainer(c, &p.rtl, display.CarouselContainerProps{
						Key: "rtl-projects", Label: tr(c, "RTL project carousel"), Count: 5, Height: 180,
						Align: display.CarouselAlignStart, Controller: rtl,
					}, func(c *ui.Context) {
						display.CarouselContent(c, display.CarouselContentProps{}, func(c *ui.Context) {
							for id := range 5 {
								display.CarouselItem(c, id, display.CarouselItemProps{Key: id}, func(c *ui.Context) { s.carouselProjectCard(c, id, false) })
							}
						})
						display.CarouselPrevious(c, display.CarouselControlProps{Label: tr(c, "Previous RTL project")})
						display.CarouselNext(c, display.CarouselControlProps{Label: tr(c, "Next RTL project")})
						p.rtlView = display.CarouselSnapshotOf(c)
					})
				})
				carouselPreviewPosition(c, p.rtlView)
			}).Label(tr(c, "RTL carousel example")).Shrink(0)
		})
	})
}

func (s *State) carouselProjectCard(c *ui.Context, id int, editable bool) {
	t := theme.TokensOf(c)
	ui.Column(c).Fill().MinWidth(0).Padding(16).Gap(10).Radius(t.RadiusMD()).Border(1, t.Border).Background(t.Card).Children(func() {
		ui.Text(c, fmt.Sprintf("%02d", id+1)).FontSize(24).FontWeight(600)
		ui.Text(c, carouselProjectNames[id]).FontWeight(600).SingleLine()
		if editable {
			control.Input(c, &s.carousels.notes[id], control.InputProps{Label: tr(c, "Notes for ") + carouselProjectNames[id], Placeholder: tr(c, "Add a note…"), Size: control.Small}).MinWidth(0)
		}
	})
}

func carouselPreviewPosition(c *ui.Context, snapshot display.CarouselSnapshot) {
	ui.Text(c, fmt.Sprintf(tr(c, "Position %d of %d · %d projects"), snapshot.SelectedSnap+1, len(snapshot.Snaps), snapshot.SlideCount)).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
}
