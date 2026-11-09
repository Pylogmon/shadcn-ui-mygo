package conversation

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/internal/shimmer"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type AttachmentSize uint8

const (
	AttachmentDefault AttachmentSize = iota
	AttachmentSmall
	AttachmentExtraSmall
)
const (
	AttachmentIdle       = "idle"
	AttachmentUploading  = "uploading"
	AttachmentProcessing = "processing"
	AttachmentError      = "error"
	AttachmentDone       = "done"
)

type AttachmentMediaVariant uint8

const (
	AttachmentIconMedia AttachmentMediaVariant = iota
	AttachmentImageMedia
)

type AttachmentProps struct {
	Name, MediaType string
	SizeBytes       int64
	// ID supplies stable identity when attachments move within a group.
	ID string
	// Empty and the legacy "ready" state mean done. Progress is 0..1;
	// it is display data and never starts an upload or changes itself.
	State       string
	Progress    float64
	Size        AttachmentSize
	Orientation ui.Orientation
	Disabled    bool
	Description string
	Icon        *ui.SVG
	Image       ui.ImageSource
	ImageAlt    string
	HideMedia   bool
	// Preview is the compatible custom media slot. It runs during this build.
	Preview func()
	// OnOpen retains the compatible separate Open action. Use the compound
	// Trigger for a whole-card action while child controls keep their input.
	OpenLabel                     string
	OnOpen, OnRemove              func()
	OnCancel, OnRetry, OnDownload func()
	// Native file drops are an opt-in composition extension. The application
	// receives paths and owns validation, transport and durable state.
	OnDropFiles func([]string)
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

type attachmentIdentity struct{ id string }

// AttachmentParts contains only this build's slots. Keep durable editor,
// transfer and dialog state in the application, rather than retaining parts.
type AttachmentParts struct {
	Root, MediaRoot, ContentRoot, ActionsRoot, TriggerRoot ui.Element
	c                                                      *ui.Context
	props                                                  AttachmentProps
	state                                                  string
	hasMedia, hasContent, triggered                        bool
}

func attachmentState(value string) (string, bool) {
	switch value {
	case "", "ready", AttachmentDone:
		return AttachmentDone, true
	case AttachmentIdle, AttachmentUploading, AttachmentProcessing, AttachmentError:
		return value, true
	default:
		return AttachmentDone, false
	}
}
func attachmentMetrics(size AttachmentSize) (media, gap, font, px, py float32) {
	switch size {
	case AttachmentSmall:
		return 32, 10, 12, 8, 6
	case AttachmentExtraSmall:
		return 28, 6, 12, 6, 4
	default:
		return 40, 8, 14, 10, 8
	}
}

// AttachmentWith builds the source's compound slots without implicit media,
// content or actions. Trigger makes the surface a real native button while
// child controls keep their own hit testing, keys and accessibility identities.
func AttachmentWith(c *ui.Context, p AttachmentProps, build func(*AttachmentParts)) AttachmentParts {
	state, valid := attachmentState(p.State)
	if p.Size > AttachmentExtraSmall {
		p.Size, valid = AttachmentDefault, false
	}
	if p.Orientation != 0 && p.Orientation != ui.Horizontal && p.Orientation != ui.Vertical {
		p.Orientation, valid = ui.Horizontal, false
	}
	if math.IsNaN(p.Progress) || math.IsInf(p.Progress, 0) {
		p.Progress, valid = 0, false
	}
	p.Progress = max(0, min(1, p.Progress))
	t := theme.TokensOf(c)
	_, gap, font, _, _ := attachmentMetrics(p.Size)
	radius := t.RadiusXL()
	if p.Size == AttachmentExtraSmall {
		radius = t.RadiusLG()
	}
	root := ui.Row(c).MinWidth(0).MaxWidthPercent(100).Shrink(0).AlignSelf(ui.Start).Wrap().Gap(gap).
		Radius(radius).Border(1, t.Border).Background(t.Card).TextColor(t.CardForeground).
		FontSize(font).Role(ui.RoleGroup).Label(p.Name).Disabled(p.Disabled).DimDisabled(false).
		Invalid(!valid).FocusRing(false).SnapAlignX(ui.Start).
		Busy(state == AttachmentUploading || state == AttachmentProcessing)
	if p.ID != "" {
		root.Key(attachmentIdentity{p.ID})
	}
	if p.Orientation == ui.Vertical {
		root.Column().Width(96)
	} else {
		root.MinWidth(160).AlignItems(ui.Center)
	}
	if state == AttachmentIdle {
		root.BorderStyle(ui.BorderDashed)
	}
	if state == AttachmentError {
		root.BorderColor(t.Destructive.Alpha(.3))
	}
	a := AttachmentParts{Root: root, c: c, props: p, state: state}
	root.Children(func() {
		if build != nil {
			build(&a)
		}
	})
	if a.triggered && !p.Disabled && root.Hovered() {
		root.Background(t.Muted.Alpha(.5))
	}
	if root.FocusWithin() {
		root.Shadow(0, 0, 0, 1, t.Ring.Alpha(.5))
	}
	if p.OnDropFiles != nil && !p.Disabled {
		if paths := root.DroppedFiles(); len(paths) > 0 {
			p.OnDropFiles(append([]string(nil), paths...))
		}
	}
	motion.ApplyTransition(c, root)
	return a
}
func (a *AttachmentParts) padding() {
	_, _, _, px, py := attachmentMetrics(a.props.Size)
	if a.hasMedia {
		a.Root.Padding(py)
	} else if a.hasContent {
		a.Root.Padding(py, px)
	}
	if a.props.Orientation == ui.Vertical && a.hasContent {
		a.Root.Width(120)
	}
}
func (a *AttachmentParts) Media(variant AttachmentMediaVariant, content func(*ui.Context)) ui.Element {
	t := theme.TokensOf(a.c)
	size, _, _, _, _ := attachmentMetrics(a.props.Size)
	radius := t.RadiusLG()
	if a.props.Size == AttachmentExtraSmall {
		radius = t.RadiusMD()
	}
	media := ui.Box(a.c).Shrink(0).Center().Clip().Radius(radius).Background(t.Muted).TextColor(t.Foreground)
	if a.props.Orientation == ui.Vertical {
		media.WidthPercent(100).AspectRatio(1)
	} else {
		media.Size(size, size)
	}
	if a.state == AttachmentError {
		media.Background(t.Destructive.Alpha(.1)).TextColor(t.Destructive)
	}
	if variant == AttachmentImageMedia && a.state != AttachmentIdle && a.state != AttachmentDone {
		media.Opacity(.6)
	}
	a.hasMedia, a.MediaRoot = true, media
	a.padding()
	return media.Children(func() {
		if content != nil {
			content(a.c)
		}
	})
}
func (a *AttachmentParts) Content(content func(*ui.Context)) ui.Element {
	box := ui.Column(a.c).Grow(1).MinWidth(0).MaxWidthPercent(100).LineHeight(1.25)
	if a.props.Orientation == ui.Vertical {
		box.FillWidth().PaddingX(4)
	} else {
		box.Basis(0)
	}
	a.hasContent, a.ContentRoot = true, box
	a.padding()
	return box.Children(func() {
		if content != nil {
			content(a.c)
		}
	})
}
func (a *AttachmentParts) Title(text string) ui.Element {
	_, _, size, _, _ := attachmentMetrics(a.props.Size)
	measure := ui.Font{Family: a.c.Theme().Font, Size: size, Weight: 500, LineHeight: 1.25}
	return shimmer.Text(a.c, text, shimmer.Props{
		Active:      a.state == AttachmentUploading || a.state == AttachmentProcessing,
		Base:        theme.TokensOf(a.c).CardForeground,
		Font:        ui.Font{Weight: 500},
		MeasureFont: &measure,
	}).MinWidth(0).MaxWidthPercent(100).SingleLine().Ellipsis("…")
}
func (a *AttachmentParts) Description(text string) ui.Element {
	fg := theme.TokensOf(a.c).MutedForeground
	if a.state == AttachmentError {
		fg = theme.TokensOf(a.c).Destructive.Alpha(.8)
	}
	return ui.Text(a.c, text).MinWidth(0).MaxWidthPercent(100).Margin(2, 0, 0, 0).FontSize(12).SingleLine().Ellipsis("…").TextColor(fg)
}
func (a *AttachmentParts) Actions(content func(*ui.Context)) ui.Element {
	box := ui.Row(a.c).Shrink(0).AlignItems(ui.Center).ZIndex(20)
	if a.props.Orientation == ui.Vertical {
		box.Absolute().Top(12).Right(12).Gap(4)
	}
	a.ActionsRoot = box
	return box.Children(func() {
		if content != nil {
			content(a.c)
		}
	})
}

// Action returns an icon-sized ghost Button. The caller may replace its
// children or fluent styles and uses Clicked for an application action.
func (a *AttachmentParts) Action(label string, icon *ui.SVG) ui.Element {
	return control.Button(a.c, "", control.ButtonProps{Variant: control.Ghost, Size: control.IconExtraSmall, Icon: icon, Label: label})
}
func (a *AttachmentParts) Trigger(label string) ui.Element {
	a.triggered, a.TriggerRoot = true, a.Root
	a.Root.Role(ui.RoleButton).Label(label).Focusable().Cursor(ui.CursorPointer)
	a.Root.Clicked() // Register native pointer and Enter/Space activation.
	return a.Root
}

// Attachment is the compatible ready-made file card. The compound form also
// permits content-only cards, custom editors and whole-card triggers.
func Attachment(c *ui.Context, p AttachmentProps) ui.Element {
	a := AttachmentWith(c, p, func(a *AttachmentParts) {
		if !p.HideMedia {
			variant := AttachmentIconMedia
			if p.Image != nil || p.Preview != nil {
				variant = AttachmentImageMedia
			}
			a.Media(variant, func(c *ui.Context) {
				if p.Image != nil {
					image := ui.Image(c, p.Image).Fill().Fit(ui.Cover).IgnoreInput(true)
					if p.ImageAlt != "" {
						image.Label(p.ImageAlt)
					} else {
						image.AccessHidden(true)
					}
				} else if p.Preview != nil {
					p.Preview()
				} else if a.state == AttachmentUploading {
					indicator.Spinner(c, indicator.Props{Size: a.iconSize(), Decorative: true})
				} else {
					icon := p.Icon
					if icon == nil {
						kind := attachmentFileIcon(p.Name, p.MediaType)
						switch a.state {
						case AttachmentIdle:
							kind = "clock"
						case AttachmentError:
							kind = "file-warning"
						case AttachmentDone:
							if p.State == AttachmentDone {
								kind = "check"
							}
						}
						icon = attachmentDefaultIcon(kind)
					}
					ui.Icon(c, icon).Size(a.iconSize(), a.iconSize()).IgnoreInput(true).AccessHidden(true)
				}
			})
		}
		a.Content(func(*ui.Context) {
			a.Title(p.Name)
			if description := attachmentDescription(a.props, a.state); description != "" {
				a.Description(description).Role(ui.RoleStatus)
			}
		})
		if p.OnOpen != nil || p.OnRemove != nil || p.OnCancel != nil || p.OnRetry != nil || p.OnDownload != nil {
			a.Actions(func(c *ui.Context) {
				if p.OnOpen != nil {
					label := p.OpenLabel
					if label == "" {
						label = "Open"
					}
					if control.Button(c, label, control.ButtonProps{Variant: control.Ghost, Size: control.ExtraSmall, Disabled: a.state == AttachmentUploading}).Clicked() {
						p.OnOpen()
					}
				}
				if p.OnDownload != nil && a.Action("Download "+p.Name, attachmentDefaultIcon("download")).Clicked() {
					p.OnDownload()
				}
				if p.OnRetry != nil && a.state == AttachmentError && a.Action("Retry "+p.Name, attachmentDefaultIcon("refresh-cw")).Clicked() {
					p.OnRetry()
				}
				if p.OnCancel != nil && (a.state == AttachmentUploading || a.state == AttachmentProcessing) && a.Action("Cancel "+p.Name, attachmentDefaultIcon("x")).Clicked() {
					p.OnCancel()
				}
				if p.OnRemove != nil && a.Action("Remove "+p.Name, attachmentDefaultIcon("x")).Clicked() {
					p.OnRemove()
				}
			})
		}
	})
	return a.Root
}
func (a *AttachmentParts) iconSize() float32 {
	if a.props.Orientation == ui.Vertical {
		return 24
	}
	if a.props.Size == AttachmentExtraSmall {
		return 14
	}
	return 16
}

func attachmentFileIcon(name, mediaType string) string {
	extension, mediaType := strings.ToLower(filepath.Ext(name)), strings.ToLower(mediaType)
	switch extension {
	case ".csv", ".tsv", ".xls", ".xlsx":
		return "table"
	case ".zip", ".tar", ".gz", ".bz2", ".7z", ".rar":
		return "file-archive"
	case ".ppt", ".pptx", ".key":
		return "presentation"
	case ".go", ".rs", ".py", ".js", ".jsx", ".ts", ".tsx", ".json", ".html", ".css", ".yaml", ".yml", ".toml":
		return "file-code"
	}
	if strings.Contains(mediaType, "spreadsheet") || strings.Contains(mediaType, "csv") {
		return "table"
	}
	if strings.Contains(mediaType, "json") || strings.Contains(mediaType, "javascript") {
		return "file-code"
	}
	return "file-text"
}
func attachmentDescription(p AttachmentProps, state string) string {
	if p.Description != "" {
		return p.Description
	}
	switch state {
	case AttachmentIdle:
		return "Ready to upload"
	case AttachmentUploading:
		return fmt.Sprintf("Uploading · %.0f%%", p.Progress*100)
	case AttachmentProcessing:
		return "Processing document"
	case AttachmentError:
		return "Upload failed. Try again."
	default:
		if p.MediaType != "" {
			return p.MediaType + " · " + FileSize(p.SizeBytes)
		}
		return FileSize(p.SizeBytes)
	}
}
