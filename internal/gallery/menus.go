package gallery

import (
	"errors"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

var errNoDesktopWindow = errors.New("this preview has no desktop window")

type creatorMenu struct {
	Open bool
	Menu desktop.MenuState
}

func (s *State) previewMenu(c *ui.Context, key, label string, p overlay.MenuProps) {
	trigger := func(b ui.Element) {
		control.StyleButton(c, b, label, control.ButtonProps{Variant: control.Outline, Size: control.Small})
	}
	if s.Config.NativeMenus {
		chosen, err := desktop.NativeMenu(c, p, trigger)
		if err != nil {
			s.popupError(c, err)
		}
		if chosen != "" {
			s.pendingMenuAction = chosen
		}
		return
	}
	anchor := control.Button(c, label, control.ButtonProps{Variant: control.Outline, Size: control.Small})
	if s.Popovers == nil {
		if anchor.Clicked() {
			s.popupError(c, errNoDesktopWindow)
		}
		return
	}
	if s.menuStates == nil {
		s.menuStates = make(map[string]*creatorMenu)
	}
	state := s.menuStates[key]
	if state == nil {
		state = &creatorMenu{}
		s.menuStates[key] = state
	}
	_, err := desktop.Menu(c, s.Popovers, key, anchor, &state.Open, &state.Menu, desktop.MenuProps{Items: p.Items, Width: p.Width, Label: label, OnChoose: func(chosen string) { s.pendingMenuAction = chosen }})
	if err != nil {
		s.popupError(c, err)
	}
}

func (s *State) previewPopover(c *ui.Context, anchor ui.Element) {
	if s.Popovers == nil {
		if anchor.Clicked() {
			s.popupError(c, errNoDesktopWindow)
		}
		return
	}
	p := desktop.PopoverProps{Label: "Dimensions", Placement: mygo.PlacementBottomStart, Width: 320, Height: 220, FitContent: true}
	var err error
	if s.Config.NativeMenus {
		err = desktop.NativePopover(c, s.Popovers, "preview-popover", anchor, &s.Popover, p, func(child *ui.Context) {
			ui.Column(child).Gap(10).Children(func() {
				ui.Text(child, "Dimensions").FontSize(16).FontWeight(600)
				ui.Text(child, "Set the dimensions for the layer.")
				ui.TextInput(child, &s.Name).Label("Layer name")
			})
		})
	} else {
		err = desktop.Popover(c, s.Popovers, "preview-popover", anchor, &s.Popover, p, func(child *ui.Context) {
			ui.Column(child).Gap(12).Children(func() {
				s.cardHeader(child, "Dimensions", "Set the dimensions for the layer.")
				control.Input(child, &s.Name, control.InputProps{Label: "Layer name"})
			})
		})
	}
	if err != nil {
		s.popupError(c, err)
	}
}
