package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

func (s *State) notificationProps() overlay.ToasterProps {
	returnFocus := s.toastReturnFocus
	if s.Page != "overlays" {
		returnFocus = &s.toastPageReturn
	}
	return overlay.ToasterProps{
		State: &s.toastPreview, Position: overlay.ToastPosition(s.toastPosition),
		RichColors: s.toastRich, CloseButton: s.toastClose, Expand: s.toastExpand,
		ReturnFocus: returnFocus,
	}
}

func (s *State) notificationViewport(c *ui.Context) ui.Element {
	return overlay.ToasterWith(c, s.notificationProps())
}

func (s *State) notificationsPreview(c *ui.Context) {
	if s.toastKind == "" {
		s.toastKind = "normal"
	}
	if s.toastPosition == "" {
		s.toastPosition = string(overlay.ToastBottomRight)
	}
	display.Card(c, func() {
		s.cardHeader(c, "Notifications", "Stack a few updates, choose a position and try an action.")
		control.RadioGroup(c, &s.toastKind, control.RadioGroupProps{Label: "Notification type", Horizontal: true,
			Options: []control.Option{{Value: "normal", Label: "Normal"}, {Value: "success", Label: "Success"}, {Value: "info", Label: "Info"}, {Value: "warning", Label: "Warning"}, {Value: "error", Label: "Error"}, {Value: "loading", Label: "Loading"}},
		}).Wrap()
		control.RadioGroup(c, &s.toastPosition, control.RadioGroupProps{Label: "Notification position", Horizontal: true,
			Options: []control.Option{{Value: string(overlay.ToastTopLeft), Label: "Top left"}, {Value: string(overlay.ToastTopCenter), Label: "Top center"}, {Value: string(overlay.ToastTopRight), Label: "Top right"}, {Value: string(overlay.ToastBottomLeft), Label: "Bottom left"}, {Value: string(overlay.ToastBottomCenter), Label: "Bottom center"}, {Value: string(overlay.ToastBottomRight), Label: "Bottom right"}},
		}).Wrap()
		ui.Row(c).Wrap().Gap(16).Children(func() {
			control.Switch(c, &s.toastRich, control.SwitchProps{Label: "Rich notification colors", Size: control.Small})
			control.Switch(c, &s.toastClose, control.SwitchProps{Label: "Show notification close button", Size: control.Small})
			control.Switch(c, &s.toastExpand, control.SwitchProps{Label: "Expand notifications", Size: control.Small})
		})
		control.Checkbox(c, &s.toastKeepOpen, "Keep open after action", control.CheckboxProps{})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			if control.Button(c, "Show notification", control.ButtonProps{Size: control.Small}).Bind(&s.toastShow).Clicked() {
				s.toastReturnFocus = &s.toastShow
				s.toastPreview.Show(ui.Toast{Title: notificationTitle(s.toastKind), Description: "Your workspace has a new update.", Type: s.toastKind, Action: "Review"}, overlay.ToastOptions{
					Cancel: "Later", OnCancel: func() { s.toastResult = "Notification postponed" },
					Action: func(event *overlay.ToastActionEvent) {
						s.toastResult = "Notification action selected"
						if s.toastKeepOpen {
							event.PreventDefault()
						}
					},
				})
			}
			if control.Button(c, "Stack three updates", control.ButtonProps{Variant: control.Outline, Size: control.Small}).Bind(&s.toastStack).Clicked() {
				s.toastReturnFocus = &s.toastStack
				for i := 1; i <= 3; i++ {
					s.toastPreview.Show(ui.Toast{Title: fmt.Sprintf("Workspace update %d", i), Description: "A little progress, one step at a time.", Type: "success"})
				}
			}
			if control.Button(c, "Start upload", control.ButtonProps{Variant: control.Outline, Size: control.Small}).Bind(&s.toastStart).Clicked() {
				s.toastReturnFocus = &s.toastStart
				s.toastUploadID = s.toastPreview.Show(ui.Toast{ID: "gallery-upload", Title: "Uploading your file", Description: "You can keep working while this finishes.", Type: "loading"})
			}
			if control.Button(c, "Complete upload", control.ButtonProps{Variant: control.Outline, Size: control.Small, Disabled: s.toastUploadID == ""}).Bind(&s.toastComplete).Clicked() {
				s.toastReturnFocus = &s.toastStart
				if s.toastPreview.Update(s.toastUploadID, ui.Toast{Title: "Upload complete", Description: "Your file is ready to share.", Type: "success"}) {
					s.toastResult = "Upload notification updated"
				}
				s.toastUploadID = ""
				if s.toastComplete.Focused(c) {
					s.toastStart.Focus(c)
				}
			}
			if control.Button(c, "Clear preview notifications", control.ButtonProps{Variant: control.Ghost, Size: control.Small}).Clicked() {
				s.toastPreview.Dismiss("")
				c.CloseToast("")
				s.toastUploadID = ""
			}
		})
		if s.toastResult != "" {
			ui.Text(c, s.toastResult).Role(ui.RoleStatus).FontSize(14)
		}
	})
}

func notificationTitle(kind string) string {
	switch kind {
	case "success":
		return "Workspace saved"
	case "info":
		return "A new update is available"
	case "warning":
		return "Storage is almost full"
	case "error":
		return "Could not save your workspace"
	case "loading":
		return "Syncing your workspace"
	default:
		return "Event created"
	}
}
