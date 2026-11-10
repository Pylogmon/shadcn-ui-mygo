package control

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// ComboboxHostedContentProps describes content whose independent window already
// owns placement and visibility. Native uses that window's native Theme metrics.
type ComboboxHostedContentProps struct {
	Label     string
	Native    bool
	MaxHeight float32
	OnTab     func(backward bool)
	// OpenComplete is the owning host's current entry completion, not a timer
	// retained by the control. The host owns presence and close query cleanup.
	OpenComplete bool
}

type comboHostedFrame struct{ detached, parent, native, openComplete bool }

// HostedTrigger is for a window host, not a second in-window popup. It keeps
// closed typeahead, publishes accepted open changes before the host registers,
// and leaves editor focus, query cleanup and highlighting to the child Root.
// requestClose is a queued native dismissal, applied with this build's handlers.
func (b *ComboboxScope[T]) HostedTrigger(c *ui.Context, p ComboboxTriggerProps, requestClose bool, children func()) ui.Element {
	f := b.f
	f.check()
	f.hosted.detached, f.hosted.parent = true, true
	e := f.trigger(c, p, children)
	if requestClose {
		f.s.open = false
		f.externalOpen = false
	}
	if f.s.wasOpen != f.s.open && !f.externalOpen && f.p.OnOpenChange != nil {
		f.p.OnOpenChange(f.s.open)
	}
	if f.boundOpen != nil {
		*f.boundOpen = f.s.open
		f.s.lastBoundOpen = f.s.open
	}
	f.s.wasOpen = f.s.open
	e.Expanded(f.s.open)
	return e
}

// HostedContent creates the popup body in the current window. It deliberately
// does not create an Overlay, attach to another window's Element or keep Scope.
// Call it from a fresh Root in the window host's own current Context.
func (b *ComboboxScope[T]) HostedContent(c *ui.Context, p ComboboxHostedContentProps, children func()) ui.Element {
	f := b.f
	f.check()
	f.hosted.detached, f.hosted.native = true, p.Native
	f.hosted.openComplete = p.OpenComplete
	f.p.PopupInput, f.s.popupInput = true, true
	f.labelText = p.Label
	if f.s.open {
		f.s.closing = false
	}
	wasOpen := f.s.wasOpen
	if f.justMounted {
		f.s.wasOpen = false
	}
	f.prepare()
	f.syncItems()
	f.s.wasOpen = wasOpen
	panel := ui.Column(c).FillWidth().MinWidth(0).Role(ui.RoleGroup).Label(p.Label + " content").IgnoreInput(!f.s.open).AccessHidden(!f.s.open)
	if p.Native {
		panel.Padding(c.Theme().Space(1))
	} else {
		menu := theme.MenuOf(c)
		t := &menu.Tokens
		panel.Clip().Radius(t.RadiusMD()).Border(1, t.Foreground.Alpha(.1)).Background(t.Popover).TextColor(t.PopoverForeground)
		menu.Surface(c, panel)
	}
	f.panel = panel
	f.contentHeight = p.MaxHeight
	runtime := f.s
	panel.ObserveInput(func(ev ui.InputEvent) {
		if ev.Kind == ui.InputKeyDown && ev.Key == ui.KeyTab && ev.Mods & ^ui.Shift == 0 && !runtime.composing {
			runtime.open = false
			if p.OnTab != nil {
				p.OnTab(ev.Mods&ui.Shift != 0)
			}
		}
	})

	previous := f.inContent
	f.inContent = true
	panel.Children(children)
	f.inContent = previous
	if f.s.open && (f.justMounted || !f.s.wasOpen) && f.editorElement.Valid() {
		f.editorElement.Focus()
	}
	return panel
}

// ResetHostedQuery applies close cleanup after the owning native surface has
// gone. The current Root's cancelable query callback keeps its normal contract.
func (b *ComboboxScope[T]) ResetHostedQuery() { b.f.check(); b.f.setQuery("") }

func (f *comboFrame) completionReady() bool {
	if f.hosted.detached {
		return !f.hosted.parent && f.s.open && f.hosted.openComplete
	}
	return f.s.motion.sample(f.c.Now()) == f.s.motion.to
}

func (f *comboFrame) nativeHostedInput(c *ui.Context, p ComboboxInputProps) ComboboxInputParts {
	f.labelText = p.Label
	root := ui.Row(c.Key("combobox-input")).FillWidth().MinWidth(0).Gap(c.Theme().Space(1)).AlignItems(ui.Center)
	f.bindRoot(root)
	f.inputRoot = root
	var editor ui.Element
	root.Children(func() {
		editor = f.editor(c, p, true, false).DimDisabled(true)
		if p.ShowClear && len(*f.values) > 0 {
			if ui.Button(c, "Clear").KeepFocus().TabStop(false).Disabled(!f.s.enabled).Label("Clear " + p.Label).Clicked() {
				f.change(nil)
				f.setQuery("")
				f.s.hasActive, f.s.filtering = false, false
				f.s.input.Focus()
			}
		}
	})
	handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
	editor.Bind(handle)
	return ComboboxInputParts{Root: root, Editor: editor}
}
