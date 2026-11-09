package display

import (
	"errors"
	"image"
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestAvatarUsesVegaDefaultDiameterAndAccessibleImageName(t *testing.T) {
	r := displayTester(func(c *ui.Context) { Avatar(c, AvatarProps{Fallback: "SD", Label: "Sofia Davis"}) })
	box, ok := r.Find("Sofia Davis")
	if !ok || box.W != 32 || box.H != 32 {
		t.Fatalf("default avatar diameter=%+v, present=%v", box, ok)
	}
	if _, ok := r.AccessByRole(ui.RoleImage, "Sofia Davis"); !ok {
		t.Fatal("avatar has no named native image role")
	}
}

func TestAvatarInvalidDecodedImageShowsFallback(t *testing.T) {
	for _, source := range []ui.ImageSource{(*ui.Bitmap)(nil), new(ui.Bitmap), (*ui.SVG)(nil)} {
		r := displayTester(func(c *ui.Context) { Avatar(c, AvatarProps{Image: source, Fallback: "SD"}) })
		if !r.HasText("SD") {
			t.Fatal("missing decoded image hid the fallback")
		}
	}
}

func avatarPicture() *ui.Bitmap {
	im := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			value := color.RGBA{R: 220, A: 255}
			if x >= 20 {
				value = color.RGBA{B: 220, A: 255}
			}
			im.SetRGBA(x, y, value)
		}
	}
	return ui.NewBitmap(im)
}

func TestAvatarControlledLoadErrorFallbackDelayAndIdle(t *testing.T) {
	p := AvatarProps{Label: "Sofia", Fallback: "SD", Loading: true, FallbackDelay: 200 * time.Millisecond}
	var statuses []AvatarImageStatus
	p.OnLoadingStatusChange = func(status AvatarImageStatus) { statuses = append(statuses, status) }
	back := ui.RGB(220, 230, 240)
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(back).Children(func() { Avatar(c, p) })
	}, 240, 120)
	box, _ := r.Find("Sofia")
	if r.HasText("SD") {
		t.Fatal("fallback ignored its delay")
	}
	pixel := r.Image().RGBAAt(int(box.X+8), int(box.Y+16))
	if pixel != (color.RGBA{R: 220, G: 230, B: 240, A: 255}) {
		t.Fatalf("delayed fallback painted a placeholder early: %v", pixel)
	}
	r.Advance(201 * time.Millisecond)
	if !r.HasText("SD") {
		t.Fatal("fallback timer did not display initials")
	}
	if r.FramePending() {
		t.Fatal("settled fallback kept requesting frames")
	}
	p.Loading = false
	p.Error = errors.New("decode failed")
	r.Frame()
	if !r.HasText("SD") {
		t.Fatal("load failure removed the fallback")
	}
	p.Error = nil
	p.Image = avatarPicture()
	r.Frame()
	if r.HasText("SD") {
		t.Fatal("loaded image retained fallback text")
	}
	if got := r.Image().RGBAAt(int(box.X+8), int(box.Y+16)); got.R != 220 || got.B != 0 {
		t.Fatalf("loaded cover image did not paint: %v", got)
	}
	p.Image = nil
	r.Frame()
	if !r.HasText("SD") {
		t.Fatal("image removal restarted an already passed fallback delay")
	}
	r.Frame()
	want := []AvatarImageStatus{AvatarImageLoading, AvatarImageError, AvatarImageLoaded, AvatarImageIdle}
	if !slices.Equal(statuses, want) {
		t.Fatalf("status notifications=%v", statuses)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("completed image changes left idle frames")
	}
}

func TestAvatarSizesCoverPixelsCornerBadgeAndSemantics(t *testing.T) {
	for _, scale := range []AvatarSize{AvatarSmall, AvatarDefault, AvatarLarge} {
		t.Run(map[AvatarSize]string{AvatarSmall: "small", AvatarDefault: "default", AvatarLarge: "large"}[scale], func(t *testing.T) {
			content := 0
			r := displayTester(func(c *ui.Context) {
				Avatar(c, AvatarProps{Label: "Sofia", Scale: scale, Image: avatarPicture(), Badge: &AvatarBadgeProps{Label: "Online", Content: func(child *ui.Context) { content++; ui.Text(child, "+").FontSize(8) }}})
			})
			info, ok := r.AccessByRole(ui.RoleImage, "Sofia")
			if !ok || info.Description != "Online" {
				t.Fatalf("avatar status semantics=%+v", info)
			}
			want := map[AvatarSize]float32{AvatarSmall: 24, AvatarDefault: 32, AvatarLarge: 40}[scale]
			if info.Bounds.W != want || info.Bounds.H != want {
				t.Fatalf("diameter=%+v", info.Bounds)
			}
			if scale == AvatarSmall && content != 0 {
				t.Fatal("small badge rendered an icon that should be hidden")
			}
			if scale != AvatarSmall && content == 0 {
				t.Fatal("larger badge did not render its content")
			}
			im := r.Image()
			x, y := info.Bounds.X, info.Bounds.Y
			if left := im.RGBAAt(int(x+want/4), int(y+want/2)); left.R != 220 || left.B != 0 {
				t.Fatalf("cover left=%v", left)
			}
			if right := im.RGBAAt(int(x+3*want/4), int(y+want/2)); right.B != 220 || right.R != 0 {
				t.Fatalf("cover right=%v", right)
			}
			var tokens *theme.Tokens
			r2 := displayTester(func(c *ui.Context) {
				tokens = theme.TokensOf(c)
				Avatar(c, AvatarProps{Label: "Badge crop", Scale: scale, Image: avatarPicture(), Badge: &AvatarBadgeProps{Label: "Online"}})
			})
			badge, _ := r2.Find("Badge crop")
			px := r2.Image().RGBAAt(int(badge.X+want-3), int(badge.Y+want-3))
			if px.R != tokens.Primary.R || px.G != tokens.Primary.G || px.B != tokens.Primary.B {
				t.Fatalf("corner badge was clipped by the image circle: %v", px)
			}
		})
	}
}

func TestAvatarGroupOverlapCountRTLAndKeyedDelay(t *testing.T) {
	p := AvatarGroupProps{Label: "Team", Max: 2, Total: 5, Items: []AvatarProps{
		{Key: "sofia", Label: "Sofia", Fallback: "SD", FallbackDelay: 100 * time.Millisecond},
		{Key: "jackson", Label: "Jackson", Fallback: "JL"},
		{Key: "other", Label: "Other", Fallback: "OM"},
	}}
	r := displayTester(func(c *ui.Context) { AvatarGroup(c, p) })
	a, _ := r.AccessByRole(ui.RoleImage, "Sofia")
	b, _ := r.AccessByRole(ui.RoleImage, "Jackson")
	count, ok := r.AccessByRole(ui.RoleImage, "3 more people")
	if !ok || b.Bounds.X-a.Bounds.X != 24 || count.Bounds.X-b.Bounds.X != 24 {
		t.Fatalf("group spacing a=%+v b=%+v count=%+v", a.Bounds, b.Bounds, count.Bounds)
	}
	if _, ok := r.AccessByRole(ui.RoleImage, "Other"); ok {
		t.Fatal("hidden member rendered outside Max")
	}
	r.Advance(101 * time.Millisecond)
	if !r.HasText("SD") {
		t.Fatal("initial fallback delay never passed")
	}
	p.Items[0], p.Items[1] = p.Items[1], p.Items[0]
	r.Frame()
	if !r.HasText("SD") {
		t.Fatal("reordered avatar lost its keyed delay state")
	}
	r.SetDirection(ui.RightToLeft)
	a, _ = r.AccessByRole(ui.RoleImage, "Sofia")
	b, _ = r.AccessByRole(ui.RoleImage, "Jackson")
	if b.Bounds.X-a.Bounds.X != 24 {
		t.Fatalf("RTL overlap a=%+v b=%+v", a.Bounds, b.Bounds)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("group scheduled idle frames")
	}
}

func TestAvatarInvalidSizeDecorativeAndCustomFallback(t *testing.T) {
	for _, size := range []float32{float32(math.NaN()), float32(math.Inf(1)), -2} {
		r := displayTester(func(c *ui.Context) {
			Avatar(c, AvatarProps{Label: "Decorative", Decorative: true, Size: size, FallbackContent: func(child *ui.Context) { ui.Text(child, "Custom fallback") }})
		})
		if _, ok := r.AccessByRole(ui.RoleImage, "Decorative"); ok {
			t.Fatal("decorative avatar remained accessible")
		}
		box, ok := r.Find("Decorative")
		if !ok || box.W != 32 || box.H != 32 {
			t.Fatalf("invalid diameter escaped normalization: %+v", box)
		}
		if !r.HasText("Custom fallback") {
			t.Fatal("custom fallback not built")
		}
	}
}

func TestAvatarGroupUserKeyCannotCollideWithRemainingCount(t *testing.T) {
	p := AvatarGroupProps{Label: "Team", Total: 3, Items: []AvatarProps{{Key: "remaining", Label: "Alice", Fallback: "A", FallbackDelay: 100 * time.Millisecond}}}
	r := displayTester(func(c *ui.Context) { AvatarGroup(c, p) })
	if _, ok := r.AccessByRole(ui.RoleImage, "Alice"); !ok {
		t.Fatal("reserved-looking member key removed the member")
	}
	if _, ok := r.AccessByRole(ui.RoleImage, "2 more people"); !ok {
		t.Fatal("member key collided with the remaining count")
	}
	r.Advance(101 * time.Millisecond)
	if !r.HasText("A") {
		t.Fatal("count and member shared their local delay state")
	}
	p.Total = 1
	r.Frame()
	if !r.HasText("A") {
		t.Fatal("count removal remounted the keyed member")
	}
}
