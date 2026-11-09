package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type application struct {
	host       *desktop.Host
	appearance theme.Controller
	native     bool
	updated    bool
	revision   uint64
}

func (a *application) tip(c *ui.Context, key string, anchor ui.Element, description string, side desktop.TooltipSide, content func(*ui.Context)) {
	build := desktop.TooltipWith
	if a.native {
		build = desktop.NativeTooltipWith
	}
	if err := build(c, a.host, key, anchor, desktop.TooltipProps{Description: description, Side: side, Revision: a.revision}, content); err != nil {
		log.Print(err)
	}
}

func (a *application) view(c *ui.Context) {
	a.appearance.Apply(c)
	if a.host != nil {
		a.host.BeginFrame()
		defer a.host.EndFrame()
	}
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).FillWidth().Padding(24).Gap(16).Background(c.Theme().Background).Children(func() {
			display.Typography(c, "Tooltips", display.H2)
			ui.Text(c, "Hover or focus a trigger. Escape dismisses the hint without moving focus.")
			ui.Row(c).Wrap().Gap(12).Children(func() {
				control.Switch(c, &a.native, control.SwitchProps{Label: "Platform theme"})
				if control.Button(c, "Toggle appearance", control.ButtonProps{Variant: control.Outline}).Clicked() {
					a.appearance.Toggle(c.Theme().Dark)
				}
				if control.Button(c, "Update status", control.ButtonProps{Variant: control.Outline}).Clicked() {
					a.updated = !a.updated
					a.revision++
				}
			})
			card := func(title string, build func()) {
				display.Card(c, func() { ui.Text(c, title).FontWeight(600); build() })
			}
			card("Basic", func() {
				anchor := control.Button(c, "Show tooltip", control.ButtonProps{Variant: control.Outline})
				a.tip(c, "basic", anchor, "Add to library", "", func(child *ui.Context) { ui.Text(child, "Add to library") })
			})
			card("Sides", func() {
				ui.Row(c).Wrap().Gap(8).Children(func() {
					for _, side := range []desktop.TooltipSide{desktop.TooltipInlineStart, desktop.TooltipLeft, desktop.TooltipTop, desktop.TooltipBottom, desktop.TooltipRight, desktop.TooltipInlineEnd} {
						anchor := control.Button(c.Key(side), string(side), control.ButtonProps{Variant: control.Outline})
						a.tip(c, "side-"+string(side), anchor, "Add to library", side, func(child *ui.Context) { ui.Text(child, "Add to library") })
					}
				})
			})
			card("With icon", func() {
				anchor := control.Button(c, "", control.ButtonProps{Variant: control.Ghost, Size: control.IconSmall}).Label("Information")
				anchor.Children(func() { ui.Icon(c, outlined.Info()).Size(16, 16) })
				a.tip(c, "icon", anchor, "Additional information", "", func(child *ui.Context) { ui.Text(child, "Additional information") })
			})
			card("Long content", func() {
				anchor := control.Button(c, "Learn how it works", control.ButtonProps{Variant: control.Outline})
				description := "To learn more about how this works, check out the docs. If you have any questions, please reach out to us."
				a.tip(c, "long", anchor, description, "", func(child *ui.Context) { ui.Text(child, description).MaxWidth(296) })
			})
			card("Disabled trigger", func() {
				wrapper := ui.Box(c).AlignSelf(ui.Start)
				wrapper.Children(func() {
					control.Button(c, "Unavailable", control.ButtonProps{Variant: control.Outline, Disabled: true})
				})
				a.tip(c, "disabled", wrapper, "This feature is currently unavailable", "", func(child *ui.Context) { ui.Text(child, "This feature is currently unavailable") })
			})
			card("With keyboard shortcut", func() {
				anchor := control.Button(c, "Save changes", control.ButtonProps{Variant: control.Outline})
				a.tip(c, "keyboard", anchor, "Save changes with S", "", func(child *ui.Context) {
					ui.Text(child, "Save Changes")
					if a.native {
						ui.Text(child, "[S]")
					} else {
						display.KbdWith(child, "S", display.KbdProps{InTooltip: true})
					}
				})
			})
			card("On link", func() {
				anchor := ui.Link(c, "Tooltip documentation", "https://ui.shadcn.com/docs/components/tooltip")
				a.tip(c, "link", anchor, "Click to read the documentation", "", func(child *ui.Context) { ui.Text(child, "Click to read the documentation") })
			})
			card("Formatted content", func() {
				anchor := control.Button(c, "Status", control.ButtonProps{Variant: control.Outline})
				description := "Active. Last updated two hours ago."
				if a.updated {
					description = "Active. Updated just now."
				}
				a.tip(c, "formatted", anchor, description, "", func(child *ui.Context) {
					ui.Column(child).Gap(4).Children(func() {
						ui.Text(child, "Active").FontWeight(600)
						text := "Last updated 2 hours ago"
						if a.updated {
							text = "Updated just now"
						}
						ui.Text(child, text).FontSize(12).Opacity(.8)
					})
				})
			})
		})
	})
}

func main() {
	a := &application{}
	mygo.App.OnReady(func() {
		window := mygo.NewWindow(mygo.WindowOptions{Title: "Shadcn Tooltips", Width: 760, Height: 720, Hidden: true, Content: ui.View(a.view)})
		a.host = desktop.NewHost(window)
		window.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
