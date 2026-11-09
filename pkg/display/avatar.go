package display

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type AvatarSize uint8

const (
	AvatarDefault AvatarSize = iota // 32 DIPs
	AvatarSmall                     // 24 DIPs
	AvatarLarge                     // 40 DIPs
)

type AvatarImageStatus string

const (
	AvatarImageIdle    AvatarImageStatus = "idle"
	AvatarImageLoading AvatarImageStatus = "loading"
	AvatarImageLoaded  AvatarImageStatus = "loaded"
	AvatarImageError   AvatarImageStatus = "error"
)

type AvatarProps struct {
	Key             string
	Image           ui.ImageSource
	Fallback, Label string
	// Size explicitly sets the diameter in DIPs; Scale supplies upstream sizes.
	Size  float32
	Scale AvatarSize
	// Image is already decoded by MyGo. Loading/Error describe an app-owned
	// asynchronous load; publish its result on the UI thread with Window.Update.
	Loading bool
	Error   error
	// FallbackDelay starts at mount and passes once, like Avatar.Fallback.delay.
	// A later image change does not restart an already completed delay.
	FallbackDelay         time.Duration
	OnLoadingStatusChange func(AvatarImageStatus)
	FallbackContent       func(*ui.Context)
	Badge                 *AvatarBadgeProps
	Decorative            bool
}

type avatarState struct {
	mounted        time.Time
	delayPassed    bool
	status         AvatarImageStatus
	statusReported bool
}

func avatarDiameter(size float32, scale AvatarSize) float32 {
	if size > 0 && !math.IsNaN(float64(size)) && !math.IsInf(float64(size), 0) {
		return size
	}
	switch scale {
	case AvatarSmall:
		return 24
	case AvatarLarge:
		return 40
	default:
		return 32
	}
}

func avatarStatus(p AvatarProps) AvatarImageStatus {
	if p.Error != nil {
		return AvatarImageError
	}
	if p.Loading {
		return AvatarImageLoading
	}
	if p.Image == nil {
		return AvatarImageIdle
	}
	w, h := ui.ImageSize(p.Image)
	if w <= 0 || h <= 0 || math.IsNaN(float64(w)) || math.IsNaN(float64(h)) || math.IsInf(float64(w), 0) || math.IsInf(float64(h), 0) {
		return AvatarImageError
	}
	return AvatarImageLoaded
}

// Avatar shows a decoded image or a delayed fallback. It performs no network
// I/O; loading belongs to the application's native resource lifecycle.
func Avatar(c *ui.Context, p AvatarProps) ui.Element { return avatar(c, p, false) }

func avatar(c *ui.Context, p AvatarProps, grouped bool) ui.Element {
	if p.Key != "" {
		c = c.Key(p.Key)
	}
	t := theme.TokensOf(c)
	size := avatarDiameter(p.Size, p.Scale)
	label := p.Label
	if label == "" {
		label = p.Fallback
	}
	root := ui.Box(c).Size(size, size).Shrink(0).Radius(size / 2).Role(ui.RoleImage).Label(label).AccessHidden(p.Decorative)
	if grouped {
		root.Shadow(0, 0, 0, 2, t.Background)
	}
	if p.Badge != nil && p.Badge.Label != "" {
		root.Description(p.Badge.Label)
	}
	s := ui.Local(root, "avatar", func() avatarState { return avatarState{mounted: c.Now()} })
	status := avatarStatus(p)
	if !s.statusReported || s.status != status {
		s.status, s.statusReported = status, true
		if p.OnLoadingStatusChange != nil {
			p.OnLoadingStatusChange(status)
		}
	}
	if !s.delayPassed && (p.FallbackDelay <= 0 || c.Now().Sub(s.mounted) >= p.FallbackDelay) {
		s.delayPassed = true
	}
	if !s.delayPassed && status != AvatarImageLoaded {
		c.After(p.FallbackDelay - c.Now().Sub(s.mounted))
	}
	root.Children(func() {
		background := ui.Transparent
		if status != AvatarImageLoaded && s.delayPassed {
			background = t.Muted
		}
		ui.Box(c).Fill().Radius(size / 2).Clip().Background(background).TextColor(t.MutedForeground).Center().AccessHidden(true).Children(func() {
			if status == AvatarImageLoaded {
				ui.Image(c, p.Image).Fill().Fit(ui.Cover).AccessHidden(true)
			} else if s.delayPassed {
				if p.FallbackContent != nil {
					p.FallbackContent(c)
				} else {
					font := float32(14)
					if p.Scale == AvatarSmall {
						font = 12
					}
					if p.Size > 0 && !math.IsNaN(float64(p.Size)) && !math.IsInf(float64(p.Size), 0) {
						font = size * .35
					}
					ui.Text(c, p.Fallback).FontSize(font).FontWeight(500).SingleLine().AccessHidden(true)
				}
			}
		})
		// The ring is a separate visual layer, preserving the image's native
		// clip and keeping a corner badge outside the circular crop.
		ui.Box(c).Absolute().Fill().Radius(size/2).Border(1, t.Border).IgnoreInput(true).AccessHidden(true)
		if p.Badge != nil {
			badge := *p.Badge
			if badge.Size <= 0 {
				badge.Size = 10
				if size <= 24 {
					badge.Size = 8
				} else if size >= 40 {
					badge.Size = 12
				}
			}
			avatarBadge(c, badge, size > 24).Absolute().Right(0).Bottom(0).AccessHidden(true)
		}
	})
	return root
}

type AvatarBadgeProps struct {
	Label   string
	Icon    *ui.SVG
	Size    float32
	Content func(*ui.Context)
}

// AvatarBadge is a non-interactive status mark. AvatarProps.Badge positions it
// on the image and includes its label in the avatar's accessible description.
func AvatarBadge(c *ui.Context, p AvatarBadgeProps) ui.Element { return avatarBadge(c, p, true) }

func avatarBadge(c *ui.Context, p AvatarBadgeProps, showIcon bool) ui.Element {
	t := theme.TokensOf(c)
	size := p.Size
	if size <= 0 || math.IsNaN(float64(size)) || math.IsInf(float64(size), 0) {
		size = 10
	}
	return ui.Box(c).Size(size, size).Shrink(0).Radius(size/2).Background(t.Primary).TextColor(t.PrimaryForeground).
		Shadow(0, 0, 0, 2, t.Background).Center().Role(ui.RoleImage).Label(p.Label).Children(func() {
		if !showIcon {
			return
		}
		if p.Content != nil {
			p.Content(c)
		} else if p.Icon != nil {
			ui.Icon(c, p.Icon).Size(8, 8).AccessHidden(true)
		}
	})
}

type AvatarGroupProps struct {
	Items []AvatarProps
	Label string
	// Max limits visible avatars; Total includes members not in Items.
	// Zero Max shows all supplied items; zero Total uses len(Items).
	Max, Total int
	Scale      AvatarSize
	Size       float32
	// Overlap defaults to 8 DIPs, as the upstream -space-x-2 group.
	Overlap      float32
	CountLabel   func(int) string
	CountContent func(*ui.Context, int)
}

type avatarGroupCountKey struct{}

// AvatarGroup overlaps keyed avatars and optionally adds a remaining-count
// member. Give each AvatarProps.Key a stable value when reordering a group.
func AvatarGroup(c *ui.Context, p AvatarGroupProps) ui.Element {
	visible := len(p.Items)
	if p.Max > 0 {
		visible = min(visible, p.Max)
	}
	total := max(len(p.Items), p.Total)
	more := total - visible
	overlap := p.Overlap
	if overlap <= 0 || math.IsNaN(float64(overlap)) || math.IsInf(float64(overlap), 0) {
		overlap = 8
	}
	root := ui.Row(c).AlignItems(ui.Center).Gap(-overlap).Role(ui.RoleGroup).Label(p.Label)
	root.Children(func() {
		for i, item := range p.Items[:visible] {
			if p.Scale != AvatarDefault {
				item.Scale = p.Scale
			}
			if p.Size > 0 {
				item.Size = p.Size
			}
			key := any(i)
			if item.Key != "" {
				key = item.Key
			}
			avatar(c.Key(key), item, true)
		}
		if more > 0 {
			label := fmt.Sprintf("%d more people", more)
			if p.CountLabel != nil {
				label = p.CountLabel(more)
			}
			scale, size := p.Scale, p.Size
			if visible > 0 && scale == AvatarDefault && size <= 0 {
				scale, size = p.Items[0].Scale, p.Items[0].Size
			}
			AvatarGroupCount(c.Key(avatarGroupCountKey{}), more, AvatarGroupCountProps{Label: label, Scale: scale, Size: size, Content: p.CountContent})
		}
	})
	return root
}

type AvatarGroupCountProps struct {
	Label   string
	Size    float32
	Scale   AvatarSize
	Content func(*ui.Context, int)
}

func AvatarGroupCount(c *ui.Context, count int, p AvatarGroupCountProps) ui.Element {
	t := theme.TokensOf(c)
	size := avatarDiameter(p.Size, p.Scale)
	label := p.Label
	if label == "" {
		label = fmt.Sprintf("%d more people", max(0, count))
	}
	return ui.Box(c).Size(size, size).Shrink(0).Radius(size/2).Background(t.Muted).TextColor(t.MutedForeground).
		Shadow(0, 0, 0, 2, t.Background).Center().Role(ui.RoleImage).Label(label).Children(func() {
		if p.Content != nil {
			p.Content(c, count)
		} else {
			font := float32(14)
			if size <= 24 {
				font = 12
			}
			ui.Textf(c, "+%d", max(0, count)).FontSize(font).AccessHidden(true)
		}
	})
}
