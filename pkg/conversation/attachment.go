package conversation

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type AttachmentProps struct {
	Name, MediaType string
	SizeBytes       int64
	// State is ready (default), uploading, or error. Progress is 0..1.
	State            string
	Progress         float64
	Disabled         bool
	Icon             *ui.SVG
	Preview          func()
	OnOpen, OnRemove func()
}

func FileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", max(0, bytes))
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}

func Attachment(c *ui.Context, p AttachmentProps) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).Gap(8).Padding(12).Radius(t.RadiusMD()).Background(t.Card).Border(1, t.Border).
		Label(p.Name).Role(ui.RoleGroup).Children(func() {
		if p.Preview != nil {
			p.Preview()
		}
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			if p.Icon != nil {
				ui.Icon(c, p.Icon).FontSize(24).TextColor(t.MutedForeground)
			}
			ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
				ui.Text(c, p.Name).FontWeight(500).SingleLine()
				meta := FileSize(p.SizeBytes)
				if p.MediaType != "" {
					meta = p.MediaType + " · " + meta
				}
				if p.State == "uploading" {
					meta = fmt.Sprintf("Uploading · %.0f%%", max(0, min(1, p.Progress))*100)
				}
				if p.State == "error" {
					meta = "Upload failed"
				}
				fg := t.MutedForeground
				if p.State == "error" {
					fg = t.Destructive
				}
				ui.Text(c, meta).FontSize(12).TextColor(fg).Role(ui.RoleStatus)
			})
			if p.OnOpen != nil && control.Button(c, "Open", control.ButtonProps{Variant: control.Ghost, Size: control.Small, Disabled: p.Disabled || p.State == "uploading"}).Clicked() {
				p.OnOpen()
			}
			if p.OnRemove != nil && control.Button(c, "Remove", control.ButtonProps{Variant: control.Ghost, Size: control.Small, Disabled: p.Disabled, Label: "Remove " + p.Name}).Clicked() {
				p.OnRemove()
			}
		})
	})
}
