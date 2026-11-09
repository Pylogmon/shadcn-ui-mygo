package desktop

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type sizeBackend struct {
	fakeBackend
	sizes [][2]int
}
type sizeSurface struct {
	*fakeSurface
	backend *sizeBackend
}

func (b *sizeBackend) Open(spec Spec, view func(*ui.Context), closed func()) (Surface, error) {
	s, err := b.fakeBackend.Open(spec, view, closed)
	if err != nil {
		return nil, err
	}
	return &sizeSurface{s.(*fakeSurface), b}, nil
}
func (s *sizeSurface) SetSize(width, height int) {
	s.spec.Width, s.spec.Height = width, height
	s.backend.sizes = append(s.backend.sizes, [2]int{width, height})
}
func TestPopoverSizeCapabilityPreservesIndependentTree(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "shadcn", true: "native"}[native], func(t *testing.T) {
			b := &sizeBackend{}
			h := NewHostWithBackend(b)
			open := true
			width, height := 240, 180
			created := 0
			parent := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				h.BeginFrame()
				defer h.EndFrame()
				anchor := ui.Button(c, "Resize")
				props := PopoverProps{Width: width, Height: height}
				content := func(child *ui.Context) {
					body := ui.Column(child)
					value := ui.Local(body, "value", func() int { created++; return 0 })
					body.Children(func() {
						if ui.Button(child, "Increment").Clicked() {
							*value++
						}
						ui.Textf(child, "Value %d", *value)
					})
				}
				var err error
				if native {
					err = NativePopover(c, h, "resizable", anchor, &open, props, content)
				} else {
					err = Popover(c, h, "resizable", anchor, &open, props, content)
				}
				if err != nil {
					t.Fatal(err)
				}
			}, 500, 300)
			child := ui.NewTester(b.surfaces[0].view, width, height)
			child.Advance(300 * time.Millisecond)
			baseline := created
			child.Click("Increment")
			width, height = 420, 280
			parent.Frame()
			child.SetSize(width, height)
			if len(b.surfaces) != 1 || b.surfaces[0].closes != 0 || len(b.sizes) != 1 || b.sizes[0] != [2]int{420, 280} {
				t.Fatalf("resized by recreation: surfaces=%d sizes=%v", len(b.surfaces), b.sizes)
			}
			if created != baseline || !child.HasText("Value 1") {
				t.Fatalf("resize lost child state created=%d baseline=%d texts=%v", created, baseline, child.Texts())
			}
			parent.Frame()
			if len(b.sizes) != 1 {
				t.Fatal("idle frame resized again")
			}
		})
	}
}
func TestPopoverSizeCapabilityStillRecreatesImmutableNativeOptions(t *testing.T) {
	b := &sizeBackend{}
	h := NewHostWithBackend(b)
	open := true
	focusable := true
	parent := ui.NewTester(func(c *ui.Context) {
		h.BeginFrame()
		defer h.EndFrame()
		anchor := ui.Button(c, "Trigger")
		if err := Popover(c, h, "options", anchor, &open, PopoverProps{Width: 240, Height: 180, NonFocusable: !focusable}, func(child *ui.Context) { ui.Text(child, "Content") }); err != nil {
			t.Fatal(err)
		}
	}, 500, 300)
	focusable = false
	parent.Frame()
	if len(b.surfaces) != 2 || b.surfaces[0].closes != 1 || len(b.sizes) != 0 {
		t.Fatal("immutable changes did not replace native surface")
	}
}
func TestPopoverParentDirectionPropagatesAndRefreshesOpenChild(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "shadcn", true: "native"}[native], func(t *testing.T) {
			b := &fakeBackend{}
			h := NewHostWithBackend(b)
			open := true
			direction := ui.RightToLeft
			childDirection := ui.LeftToRight
			parent := ui.NewTester(func(c *ui.Context) {
				h.BeginFrame()
				defer h.EndFrame()
				ui.Column(c).Direction(direction).Children(func() {
					anchor := ui.Button(c, "Direction")
					fn := Popover
					if native {
						fn = NativePopover
					}
					if err := fn(c, h, "direction", anchor, &open, PopoverProps{}, func(child *ui.Context) {
						childDirection = child.Direction()
						ui.Row(child).Children(func() { ui.Text(child, "First"); ui.Text(child, "Second") })
					}); err != nil {
						t.Fatal(err)
					}
				})
			}, 500, 300)
			child := ui.NewTester(b.surfaces[0].view, 280, 180)
			child.Advance(300 * time.Millisecond)
			if childDirection != ui.RightToLeft {
				t.Fatal("parent RTL did not cross native Context boundary")
			}
			first, _ := child.Find("First")
			second, _ := child.Find("Second")
			if first.X <= second.X {
				t.Fatal("native child did not lay out right to left")
			}
			parent.Frame()
			parent.Frame()
			b.surfaces[0].invalidates = 0
			direction = ui.LeftToRight
			parent.Frame()
			if b.surfaces[0].invalidates != 1 || len(b.surfaces) != 1 {
				t.Fatal("live direction change did not invalidate existing child once")
			}
			child.Frame()
			if childDirection != ui.LeftToRight {
				t.Fatal("open native child retained old direction")
			}
			parent.Frame()
			if b.surfaces[0].invalidates != 1 {
				t.Fatal("stable direction kept invalidating child")
			}
		})
	}
}

func TestPopoverWithoutSizeCapabilityKeepsReplacementBehavior(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	open := true
	width := 240
	parent := ui.NewTester(func(c *ui.Context) {
		h.BeginFrame()
		defer h.EndFrame()
		if err := Popover(c, h, "fallback", ui.Button(c, "Resize"), &open, PopoverProps{Width: width, Height: 180}, func(child *ui.Context) { ui.Text(child, "Content") }); err != nil {
			t.Fatal(err)
		}
	}, 500, 300)
	width = 420
	parent.Frame()
	if len(b.surfaces) != 2 || b.surfaces[0].closes != 1 {
		t.Fatal("legacy surface without SetSize did not recreate")
	}
}
