package desktop

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

type focusStateBackend struct {
	fakeBackend
	active, focused bool
}
type focusStateSurface struct {
	*fakeSurface
	backend *focusStateBackend
}

func (b *focusStateBackend) ApplicationActive() bool { return b.active }
func (b *focusStateBackend) Open(spec Spec, view func(*ui.Context), closed func()) (Surface, error) {
	f, err := b.fakeBackend.Open(spec, view, closed)
	if err != nil {
		return nil, err
	}
	return &focusStateSurface{f.(*fakeSurface), b}, nil
}
func (s *focusStateSurface) IsFocused() bool { return s.backend.focused }

func TestHostFocusStateReportsRealOwnershipAndEndsWithSurface(t *testing.T) {
	b := &focusStateBackend{active: true}
	h := NewHostWithBackend(b)
	open := true
	parent := ui.NewTester(func(c *ui.Context) {
		h.BeginFrame()
		defer h.EndFrame()
		anchor := ui.Button(c, "Preview")
		if err := Popover(c, h, "preview", anchor, &open, PopoverProps{ShowInactive: true}, func(child *ui.Context) { ui.Text(child, "Details") }); err != nil {
			t.Fatal(err)
		}
	}, 400, 240)
	if active, known := h.ApplicationState(); !known || !active {
		t.Fatal("application activity was not exposed")
	}
	if focused, known := h.FocusState("preview"); !known || focused {
		t.Fatal("active application was mistaken for focus in its inactive Flyout")
	}
	b.focused = true
	if focused, known := h.FocusState("preview"); !known || !focused {
		t.Fatal("actual focus transition not reported")
	}
	b.active, b.focused = false, false
	if active, known := h.ApplicationState(); !known || active {
		t.Fatal("application blur not reported")
	}
	h.Close("preview")
	parent.Frame()
	if focused, known := h.FocusState("preview"); known || focused {
		t.Fatal("closed Flyout still reported keyboard ownership")
	}
}

func TestHostFocusAndApplicationCapabilitiesRemainUnknownWhenAbsent(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	open := true
	ui.NewTester(func(c *ui.Context) {
		h.BeginFrame()
		defer h.EndFrame()
		anchor := ui.Button(c, "Preview")
		if err := Popover(c, h, "preview", anchor, &open, PopoverProps{}, func(child *ui.Context) { ui.Text(child, "Details") }); err != nil {
			t.Fatal(err)
		}
	}, 400, 240)
	for _, host := range []*Host{nil, h} {
		if focused, known := host.FocusState("preview"); focused || known {
			t.Fatal("unknown backend invented keyboard ownership")
		}
		if active, known := host.ApplicationState(); active || known {
			t.Fatal("unknown backend invented application activity")
		}
	}
	if focused, known := h.FocusState("missing"); focused || known {
		t.Fatal("missing key reported a focus state")
	}
	native := NewHostWithBackend(&nativeBackend{})
	if active, known := native.ApplicationState(); !known || active {
		t.Fatal("native backend without a live owner reports active")
	}
}
