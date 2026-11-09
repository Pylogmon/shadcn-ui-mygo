// Run with go run ./examples/native-popups. The small parent window makes
// the difference between a native Flyout and a parent-bounded overlay visible.
package main

import (
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func main() {
	var host *desktop.Host
	var appearance theme.Controller
	var menuOpen, popoverOpen, nativePopoverOpen bool
	var state desktop.MenuState
	name := "Sofia Davis"
	lastName := name
	var nameRevision uint64
	date := time.Date(2026, time.November, 9, 0, 0, 0, 0, time.Local)
	var dates display.DateRange
	items := []overlay.MenuItem{
		{Key: "profile", Label: "Profile"},
		{Key: "export", Label: "Export", Children: []overlay.MenuItem{{Key: "go", Label: "Go source"}, {Key: "json", Label: "Preset JSON"}}},
		{Separator: true},
		{Key: "remove", Label: "Remove", Destructive: true},
	}
	view := func(c *ui.Context) {
		appearance.Apply(c)
		if host != nil {
			host.BeginFrame()
			defer host.EndFrame()
		}
		ui.Column(c).Fill().Padding(24).Gap(16).Background(c.Theme().Background).Children(func() {
			display.Typography(c, "One design. Two menu styles.", display.H3)
			display.Typography(c, "Menu uses a shadcn Flyout. NativeMenu uses the operating system.", display.Muted)
			control.NativeSelect(c, &name, control.SelectProps{Label: "NativeSelect", Options: []control.Option{{Value: "Sofia Davis", Label: "Sofia Davis"}, {Value: "Jackson Lee", Label: "Jackson Lee"}}})
			if name != lastName {
				lastName = name
				nameRevision++
			}
			ui.Row(c).Gap(8).FillWidth().Wrap().Children(func() {
				if host == nil {
					return
				}
				props := desktop.DatePickerProps{Label: "DatePicker", Calendar: display.CalendarProps{DefaultMonth: date, Browse: true}}
				_, err := desktop.DatePicker(c, host, "date", &date, props)
				report(err)
				props.Label = "NativeDatePicker"
				_, err = desktop.NativeDatePicker(c, host, "native-date", &date, props)
				report(err)
				rangeProps := desktop.DatePickerProps{Label: "DateRangePicker", Calendar: display.CalendarProps{DefaultMonth: date, NumberOfMonths: 2, Browse: true}}
				_, err = desktop.DateRangePicker(c, host, "dates", &dates, rangeProps)
				report(err)
				rangeProps.Label = "NativeDateRangePicker"
				_, err = desktop.NativeDateRangePicker(c, host, "native-dates", &dates, rangeProps)
				report(err)
			})
			ui.Row(c).Gap(8).FillWidth().Wrap().Children(func() {
				tip := control.Button(c, "Tooltip", control.ButtonProps{Variant: control.Outline})
				if host != nil {
					report(desktop.Tooltip(c, host, "tip", tip, "Shadcn tooltip outside the parent.", desktop.TooltipProps{}))
				}
				tip = control.Button(c, "NativeTooltip", control.ButtonProps{Variant: control.Outline})
				if host != nil {
					report(desktop.NativeTooltip(c, host, "native-tip", tip, "Platform-themed tooltip.", desktop.TooltipProps{}))
				}
				card := control.Button(c, "HoverCard", control.ButtonProps{Variant: control.Outline})
				if host != nil {
					report(desktop.HoverCard(c, host, "hover", card, desktop.HoverCardProps{Label: "Author", Width: 280, Revision: nameRevision}, func(child *ui.Context) {
						ui.Text(child, "Sofia Davis").FontSize(18).FontWeight(600)
						control.Input(child, &name, control.InputProps{Label: "Name"})
					}))
				}
				card = control.Button(c, "NativeHoverCard", control.ButtonProps{Variant: control.Outline})
				if host != nil {
					report(desktop.NativeHoverCard(c, host, "native-hover", card, desktop.HoverCardProps{Label: "Native author", Width: 280, Revision: nameRevision}, func(child *ui.Context) {
						ui.Text(child, "Sofia Davis").FontWeight(600)
						ui.TextInput(child, &name).Label("Name")
					}))
				}
			})
			ui.Box(c).Grow(1)
			ui.Row(c).FillWidth().Gap(8).Wrap().Justify(ui.End).Children(func() {
				anchor := control.Button(c, "Menu", control.ButtonProps{Variant: control.Outline})
				if host != nil {
					_, err := desktop.Menu(c, host, "menu", anchor, &menuOpen, &state, desktop.MenuProps{Items: items, Width: 240, Label: "Project actions"})
					if err != nil {
						log.Print(err)
					}
				}
				if _, err := desktop.NativeMenu(c, overlay.MenuProps{Items: items}, func(b ui.Element) {
					control.StyleButton(c, b, "NativeMenu", control.ButtonProps{Variant: control.Outline})
				}); err != nil {
					log.Print(err)
				}
				anchor = control.Button(c, "Popover", control.ButtonProps{})
				if host != nil {
					err := desktop.Popover(c, host, "popover", anchor, &popoverOpen, desktop.PopoverProps{Width: 320, Height: 220, Label: "Shadcn popover", Revision: nameRevision}, func(child *ui.Context) {
						ui.Text(child, "Shadcn popover").FontSize(18).FontWeight(600)
						control.Input(child, &name, control.InputProps{Label: "Name"})
						if control.Button(child, "Done", control.ButtonProps{}).Clicked() {
							popoverOpen = false
						}
					})
					if err != nil {
						log.Print(err)
					}
				}
				anchor = control.Button(c, "NativePopover", control.ButtonProps{Variant: control.Outline})
				if host != nil {
					err := desktop.NativePopover(c, host, "native-popover", anchor, &nativePopoverOpen, desktop.PopoverProps{Width: 300, Height: 200, Label: "Native popover", Revision: nameRevision}, func(child *ui.Context) {
						ui.Text(child, "Native popover").FontWeight(600)
						ui.TextInput(child, &name).Label("Name")
						if ui.Button(child, "Done").Clicked() {
							nativePopoverOpen = false
						}
					})
					if err != nil {
						log.Print(err)
					}
				}
			})
		})
	}
	mygo.App.WhenReady(func() {
		window := mygo.NewWindow(mygo.WindowOptions{Title: "Shadcn popup styles", Width: 640, Height: 460, Hidden: true, Content: ui.View(view)})
		host = desktop.NewHost(window)
		window.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func report(err error) {
	if err != nil {
		log.Print(err)
	}
}
