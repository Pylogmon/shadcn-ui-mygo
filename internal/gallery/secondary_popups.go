package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

type popupValueRevision[T comparable] struct {
	value    T
	revision uint64
}

// Track shared values so parent-side changes redraw the independent child.
func popupRevision[T comparable](anchor ui.Element, value T) uint64 {
	s := ui.Local(anchor, "gallery.popup-content", func() popupValueRevision[T] { return popupValueRevision[T]{value: value} })
	if s.value != value {
		s.value = value
		s.revision++
	}
	return s.revision
}

func (s *State) previewDatePicker(c *ui.Context) {
	p := desktop.DatePickerProps{Label: "Project date", Calendar: display.CalendarProps{Browse: true}}
	var err error
	if s.Config.NativeMenus {
		_, err = desktop.NativeDatePicker(c, s.Popovers, "preview-project-date", &s.Date, p)
	} else {
		_, err = desktop.DatePicker(c, s.Popovers, "preview-project-date", &s.Date, p)
	}
	if err != nil {
		s.popupError(c, err)
	}
}

func (s *State) previewDateRangePicker(c *ui.Context) {
	p := desktop.DatePickerProps{Label: "Booking dates", Calendar: display.CalendarProps{NumberOfMonths: 2, Min: 1, Browse: true}}
	var err error
	if s.Config.NativeMenus {
		_, err = desktop.NativeDateRangePicker(c, s.Popovers, "preview-booking-dates", &s.DateRange, p)
	} else {
		_, err = desktop.DateRangePicker(c, s.Popovers, "preview-booking-dates", &s.DateRange, p)
	}
	if err != nil {
		s.popupError(c, err)
	}
}

func (s *State) previewTooltip(c *ui.Context, anchor ui.Element) {
	var err error
	if s.Config.NativeMenus {
		err = desktop.NativeTooltip(c, s.Popovers, "preview-tooltip", anchor, "Small details make a difference.", desktop.TooltipProps{})
	} else {
		err = desktop.Tooltip(c, s.Popovers, "preview-tooltip", anchor, "Small details make a difference.", desktop.TooltipProps{})
	}
	if err != nil {
		s.popupError(c, err)
	}
}

func (s *State) previewHoverCard(c *ui.Context, anchor ui.Element) {
	native := s.Config.NativeMenus
	content := func(child *ui.Context) {
		if native {
			ui.Text(child, "Sofia Davis").FontSize(18).FontWeight(600)
			ui.Text(child, "Building native experiences with Go.")
			ui.Checkbox(child, &s.Marketing, "Follow project updates")
		} else {
			display.Avatar(child, display.AvatarProps{Fallback: "SD", Size: 40})
			s.cardHeader(child, "Sofia Davis", "Building native experiences with Go.")
			control.Checkbox(child, &s.Marketing, "Follow project updates", control.CheckboxProps{})
		}
	}
	p := desktop.HoverCardProps{Label: "Author preview", Description: "About Sofia Davis", Width: 320, Revision: popupRevision(anchor, s.Marketing)}
	var err error
	if native {
		err = desktop.NativeHoverCard(c, s.Popovers, "preview-author", anchor, p, content)
	} else {
		err = desktop.HoverCard(c, s.Popovers, "preview-author", anchor, p, content)
	}
	if err != nil {
		s.popupError(c, err)
	}
}
