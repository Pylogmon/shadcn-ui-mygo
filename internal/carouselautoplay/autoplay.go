// Package carouselautoplay ports the value-only state machine of Embla Autoplay 8.5.2.
// It stores no UI objects, clocks, callbacks or application event listeners.
package carouselautoplay

type Options struct {
	Delays                                             []float64
	Jump, PlayOnInit, StopOnFocusIn, StopOnInteraction bool
	StopOnMouseEnter, StopOnLastSnap, WatchDrag        bool
}

func Defaults() Options {
	return Options{PlayOnInit: true, StopOnFocusIn: true, StopOnInteraction: true, WatchDrag: true}
}

type Effect struct {
	Kind, Event string
	Index       int
	Jump        bool
}

type Core struct {
	options                                                   Options
	count, selected                                           int
	loop, hidden, ready, destroyed, active, mouseOver, resume bool
	timer                                                     bool
	started, deadline                                         float64
	jump                                                      bool
}

// emit is only used during the call. Synchronous delivery preserves the
// upstream intermediate isPlaying/timeUntilNext values inside event callbacks.
func (m *Core) Init(p Options, count int, loop, hidden bool, now float64, emit func(Effect)) {
	m.options = p
	m.options.Delays = append([]float64(nil), p.Delays...)
	m.count, m.loop, m.hidden = count, loop, hidden
	m.ready = count > 1
	m.destroyed = false
	if !m.ready {
		return
	}
	m.jump = p.Jump
	if p.PlayOnInit {
		m.start(now, emit)
	}
}
func (m *Core) Selected() int             { return m.selected }
func (m *Core) SetSelected(index int)     { m.selected = index }
func (m *Core) IsPlaying() bool           { return m.active }
func (m *Core) Deadline() (float64, bool) { return m.deadline, m.timer }
func (m *Core) TimeUntilNext(now float64) *float64 {
	if m.started == 0 {
		return nil
	}
	v := m.delay() - now + m.started
	return &v
}
func (m *Core) delay() float64 {
	if m.selected < 0 || m.selected >= len(m.options.Delays) {
		return 0
	}
	return m.options.Delays[m.selected]
}
func fire(emit func(Effect), e Effect) {
	if emit != nil {
		emit(e)
	}
}
func (m *Core) event(name string, emit func(Effect)) { fire(emit, Effect{Kind: "event", Event: name}) }
func (m *Core) setTimer(now float64, emit func(Effect)) {
	m.timer = true
	m.started = now
	m.deadline = now + m.delay()
	m.event("autoplay:timerset", emit)
}
func (m *Core) clearTimer(emit func(Effect)) {
	m.timer = false
	m.started = 0
	m.deadline = 0
	m.event("autoplay:timerstopped", emit)
}
func (m *Core) start(now float64, emit func(Effect)) {
	if m.destroyed || !m.ready {
		return
	}
	if m.hidden {
		m.resume = true
		return
	}
	if !m.active {
		m.event("autoplay:play", emit)
	}
	m.setTimer(now, emit)
	m.active = true
}
func (m *Core) stop(emit func(Effect)) {
	if m.destroyed {
		return
	}
	if m.active {
		m.event("autoplay:stop", emit)
	}
	m.clearTimer(emit)
	m.active = false
}
func (m *Core) Play(now float64, jump *bool, emit func(Effect)) {
	if jump != nil {
		m.jump = *jump
	}
	m.start(now, emit)
}
func (m *Core) Stop(emit func(Effect)) {
	if m.active {
		m.stop(emit)
	}
}
func (m *Core) Reset(now float64, emit func(Effect)) {
	if m.active {
		m.start(now, emit)
	}
}
func (m *Core) Destroy(emit func(Effect)) { m.stop(emit); m.destroyed = true; m.active = false }
func (m *Core) Visibility(hidden bool, now float64, emit func(Effect)) {
	m.hidden = hidden
	if hidden {
		m.resume = m.active
		m.stop(emit)
		return
	}
	if m.resume {
		m.start(now, emit)
	}
}
func (m *Core) PointerDown(emit func(Effect)) {
	if m.options.WatchDrag && !m.mouseOver {
		m.stop(emit)
	}
}
func (m *Core) PointerUp(now float64, emit func(Effect)) {
	if m.options.WatchDrag && !m.options.StopOnInteraction && !m.mouseOver {
		m.start(now, emit)
	}
}
func (m *Core) MouseEnter(emit func(Effect)) {
	if m.options.StopOnMouseEnter {
		m.mouseOver = true
		m.stop(emit)
	}
}
func (m *Core) MouseLeave(now float64, emit func(Effect)) {
	if m.options.StopOnMouseEnter && !m.options.StopOnInteraction {
		m.mouseOver = false
		m.start(now, emit)
	}
}
func (m *Core) FocusStart(emit func(Effect)) {
	if m.options.StopOnFocusIn {
		m.stop(emit)
	}
}
func (m *Core) FocusOut(now float64, emit func(Effect)) {
	if m.options.StopOnFocusIn && !m.options.StopOnInteraction {
		m.start(now, emit)
	}
}

// Advance delivers one overdue timer at the current UI time. It does not replay
// timers as if a blocked/hidden UI had run each missing frame in the meantime.
func (m *Core) Advance(now float64, emit func(Effect)) {
	if !m.timer || now < m.deadline || m.destroyed {
		return
	}
	m.timer = false
	next := m.selected + 1
	if m.loop {
		next %= m.count
	} else {
		next = min(next, m.count-1)
	}
	kill := m.options.StopOnLastSnap && next == m.count-1
	if m.loop || m.selected < m.count-1 {
		m.selected = next
		fire(emit, Effect{Kind: "scrollNext", Index: m.selected, Jump: m.jump})
	} else {
		m.selected = 0
		fire(emit, Effect{Kind: "scrollTo", Index: 0, Jump: m.jump})
	}
	m.event("autoplay:select", emit)
	if kill {
		m.stop(emit)
		return
	}
	m.start(now, emit)
}

// Carry contains the interaction flags that the source Autoplay object keeps
// across destroy/init. Timers, active state and owner lifecycle are excluded.
type Carry struct{ MouseOver, Resume bool }

func (m *Core) Carry() Carry         { return Carry{m.mouseOver, m.resume} }
func (m *Core) RestoreCarry(v Carry) { m.mouseOver, m.resume = v.MouseOver, v.Resume }

// Started is a value query for the native detached timer snapshot.
func (m *Core) Started() float64 { return m.started }
