// Package carouselphysics ports the value-only motion primitives of the pinned
// Embla Carousel 8.5.2. It has no UI, clock, event-loop or application closures.
package carouselphysics

import "math"

type State struct {
	Location, Offset, Previous, Target float64
}

type Body struct {
	State                                          State
	baseDuration, baseFriction, duration, friction float64
	velocity, direction, raw, previousRaw          float64
}

func NewBody(initial, duration, friction float64) *Body {
	return &Body{State: State{initial, initial, initial, initial}, baseDuration: duration, baseFriction: friction, duration: duration, friction: friction, raw: initial}
}
func (b *Body) Seek() {
	displacement := b.State.Target - b.State.Location
	distance := float64(0)
	if b.duration == 0 {
		b.velocity = 0
		b.State.Previous = b.State.Target
		b.State.Location = b.State.Target
		distance = displacement
	} else {
		b.State.Previous = b.State.Location
		b.velocity += displacement / b.duration
		b.velocity *= b.friction
		b.raw += b.velocity
		b.State.Location += b.velocity
		distance = b.raw - b.previousRaw
	}
	b.direction = sign(distance)
	b.previousRaw = b.raw
}
func (b *Body) Settled() bool                   { return math.Abs(b.State.Target-b.State.Offset) < .001 }
func (b *Body) Duration() float64               { return b.duration }
func (b *Body) Friction() float64               { return b.friction }
func (b *Body) Velocity() float64               { return b.velocity }
func (b *Body) Direction() float64              { return b.direction }
func (b *Body) UseDuration(value float64) *Body { b.duration = value; return b }
func (b *Body) UseFriction(value float64) *Body { b.friction = value; return b }
func (b *Body) UseBaseDuration() *Body          { return b.UseDuration(b.baseDuration) }
func (b *Body) UseBaseFriction() *Body          { return b.UseFriction(b.baseFriction) }
func (b *Body) Interpolate(alpha float64) {
	b.State.Offset = b.State.Location*alpha + b.State.Previous*(1-alpha)
}

// Bounds intentionally consults the last rendered Offset, not Location. The
// source Engine passes offsetLocation into ScrollBounds, before each seek.
type Bounds struct {
	Limit    Limit
	View     float64
	Disabled bool
}

func (r Bounds) ShouldConstrain(b *Body) bool {
	return !r.Disabled && r.Limit.Reached(b.State.Target) && r.Limit.Reached(b.State.Offset)
}
func (r Bounds) Constrain(b *Body, pointerDown bool) {
	if !r.ShouldConstrain(b) {
		return
	}
	edge := r.Limit.Max
	if b.State.Offset < r.Limit.Min {
		edge = r.Limit.Min
	}
	diffToEdge := math.Abs(edge - b.State.Offset)
	diffToTarget := b.State.Target - b.State.Offset
	friction := clamp(diffToEdge/(r.View*.5), .1, .99)
	b.State.Target -= diffToTarget * friction
	if !pointerDown && math.Abs(diffToTarget) < r.View*.1 {
		b.State.Target = r.Limit.Constrain(b.State.Target)
		b.UseDuration(25).UseBaseFriction()
	}
}

// Wrap applies one source loop crossing to all four vectors. Raw velocity
// history stays continuous even when these finite coordinates are shifted.
func (b *Body) Wrap(content float64, limit Limit) bool {
	shift := float64(0)
	if b.direction == 1 && b.State.Offset > limit.Max+.1 {
		shift = -content
	}
	if b.direction == -1 && b.State.Offset < limit.Min+.1 {
		shift = content
	}
	if shift == 0 {
		return false
	}
	b.State.Location += shift
	b.State.Offset += shift
	b.State.Previous += shift
	b.State.Target += shift
	return true
}

// Scheduler is the source's 60Hz accumulator, including two warm-up seeks on
// its first RAF. Callbacks belong only to Advance; no callback survives a frame.
// Render can call Stop, matching a settle event ending the current animation.
type Scheduler struct {
	active            bool
	last, accumulated float64
}

const TimeStep = 1000.0 / 60

func (a *Scheduler) Start()       { a.active = true }
func (a *Scheduler) Stop()        { a.active = false; a.last = 0; a.accumulated = 0 }
func (a *Scheduler) Reset()       { a.last = 0; a.accumulated = 0 }
func (a *Scheduler) Active() bool { return a.active }
func (a *Scheduler) Advance(nowMS float64, update func(), render func(float64)) {
	if !a.active {
		return
	}
	if a.last == 0 {
		a.last = nowMS
		update()
		update()
	}
	elapsed := nowMS - a.last
	a.last = nowMS
	a.accumulated += elapsed
	for a.accumulated >= TimeStep {
		update()
		a.accumulated -= TimeStep
	}
	render(a.accumulated / TimeStep)
}

func sign(n float64) float64 {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}
func clamp(n, min, max float64) float64 {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
