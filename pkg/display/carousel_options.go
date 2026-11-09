package display

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

type CarouselAxis uint8

const (
	CarouselHorizontal CarouselAxis = iota
	CarouselVertical
)

type CarouselDirection uint8

const (
	CarouselLTR CarouselDirection = iota
	CarouselRTL
)

type CarouselOptionErrorCode string

const (
	CarouselInvalidOption   CarouselOptionErrorCode = "invalid-option"
	CarouselRequiresPhysics CarouselOptionErrorCode = "requires-physics"
	CarouselRequiresPlugin  CarouselOptionErrorCode = "requires-plugin"
	CarouselUnresolvedNode  CarouselOptionErrorCode = "unresolved-node"
	CarouselForeignNode     CarouselOptionErrorCode = "foreign-node"
	CarouselAlreadyBound    CarouselOptionErrorCode = "already-bound"
	CarouselDestroyed       CarouselOptionErrorCode = "destroyed"
)

// CarouselOptionError is returned before a failed patch changes configuration.
// Node ownership/measurement errors additionally require a live native binding.
type CarouselOptionError struct {
	Field string
	Code  CarouselOptionErrorCode
}

func (e *CarouselOptionError) Error() string {
	return fmt.Sprintf("display: carousel %s: %s", e.Field, e.Code)
}

// CarouselValue makes explicit zero/false values convenient in partial options.
func CarouselValue[T any](value T) *T { return &value }

type CarouselAlignInput struct {
	ViewSize, SnapSize float64
	Index              int
}

// CarouselAligner is an explicitly application-owned, long-lived pure resolver.
// It may capture model values, never a Context or Element. It is sampled during
// geometry resolution, not in Draw, and its arguments/result contain no UI data.
type CarouselAligner struct {
	resolve func(CarouselAlignInput) float64
}

func NewCarouselAligner(resolve func(CarouselAlignInput) float64) *CarouselAligner {
	return &CarouselAligner{resolve: resolve}
}

type CarouselAlignment struct {
	Mode     CarouselAlign
	Resolver *CarouselAligner
}

type CarouselSlideGroup struct {
	Count int
	Auto  bool
}

// CarouselNodes selects the measurable, declared native track/items. Nil fields
// leave the previous/default selection; a non-nil empty Slides selects none.
// Configure clones Handle values. Props may refer to Handles bound during the
// same build; those are resolved after Children and never retained as pointers.
type CarouselNodes struct {
	Container *ui.Handle
	Slides    *[]ui.Handle
}

type CarouselEnvironment struct {
	Width, Height float32
	Dark, Reduced bool
	Direction     CarouselDirection
}

// CarouselMedia is the native counterpart of a media query, evaluated against
// window dimensions/preferences rather than the carousel's own viewport.
type CarouselMedia struct {
	MinWidth, MaxWidth, MinHeight, MaxHeight *float32
	Dark, Reduced                            *bool
	Direction                                *CarouselDirection
}

func (m CarouselMedia) matches(env CarouselEnvironment) bool {
	return (m.MinWidth == nil || env.Width >= *m.MinWidth) && (m.MaxWidth == nil || env.Width <= *m.MaxWidth) &&
		(m.MinHeight == nil || env.Height >= *m.MinHeight) && (m.MaxHeight == nil || env.Height <= *m.MaxHeight) &&
		(m.Dark == nil || env.Dark == *m.Dark) && (m.Reduced == nil || env.Reduced == *m.Reduced) &&
		(m.Direction == nil || env.Direction == *m.Direction)
}

// Keys preserve media-rule identity and insertion order across partial ReInit.
// Matching patches merge in order, with the last matching field taking effect.
type CarouselBreakpoint struct {
	Key     string
	When    CarouselMedia
	Options CarouselOptions
}

// CarouselOptions is a partial patch. Nil means omitted; pointers preserve an
// explicit false, zero, or empty list. Watch veto callbacks belong to the current
// ContainerProps.Hooks build and are never queued by Configure/ReInit.
//
// Explicit Duration, DragThreshold, DragFree, or SkipSnaps currently return
// CarouselRequiresPhysics. These retain their source meanings for the separate
// physics stage; native compatibility Duration/DragThreshold remain on Props.
type CarouselOptions struct {
	Align                                           *CarouselAlignment
	Axis                                            *CarouselAxis
	Direction                                       *CarouselDirection
	Contain                                         *CarouselContain
	SlidesToScroll                                  *CarouselSlideGroup
	StartIndex                                      *int
	Active, Loop                                    *bool
	WatchDrag, WatchResize, WatchSlides, WatchFocus *bool
	InViewThreshold                                 *[]float64
	Nodes                                           *CarouselNodes
	Breakpoints                                     []CarouselBreakpoint
	Duration, DragThreshold                         *float64
	DragFree, SkipSnaps                             *bool
}

// Validate reports all invalid or not-yet-supported fields before construction.
// A failed patch is atomic: Configure/ReInit do not enqueue any part of it.
func (o CarouselOptions) Validate() error {
	var problems []error
	add := func(field string, code CarouselOptionErrorCode) {
		problems = append(problems, &CarouselOptionError{Field: field, Code: code})
	}
	if o.Align != nil && (o.Align.Mode > CarouselAlignEnd || o.Align.Resolver != nil && o.Align.Resolver.resolve == nil) {
		add("Align", CarouselInvalidOption)
	}
	if o.Axis != nil && *o.Axis > CarouselVertical {
		add("Axis", CarouselInvalidOption)
	}
	if o.Direction != nil && *o.Direction > CarouselRTL {
		add("Direction", CarouselInvalidOption)
	}
	if o.Contain != nil && *o.Contain > CarouselUncontained {
		add("Contain", CarouselInvalidOption)
	}
	if o.SlidesToScroll != nil && (o.SlidesToScroll.Auto && o.SlidesToScroll.Count != 0 || !o.SlidesToScroll.Auto && o.SlidesToScroll.Count < 1) {
		add("SlidesToScroll", CarouselInvalidOption)
	}
	if o.InViewThreshold != nil {
		for _, threshold := range *o.InViewThreshold {
			if math.IsNaN(threshold) || math.IsInf(threshold, 0) || threshold < 0 || threshold > 1 {
				add("InViewThreshold", CarouselInvalidOption)
				break
			}
		}
	}
	if o.Duration != nil {
		add("Duration", CarouselRequiresPhysics)
	}
	if o.DragThreshold != nil {
		add("DragThreshold", CarouselRequiresPhysics)
	}
	if o.DragFree != nil {
		add("DragFree", CarouselRequiresPhysics)
	}
	if o.SkipSnaps != nil {
		add("SkipSnaps", CarouselRequiresPhysics)
	}
	keys := make(map[string]bool)
	for i, rule := range o.Breakpoints {
		field := fmt.Sprintf("Breakpoints[%d]", i)
		if rule.Key == "" || keys[rule.Key] {
			add(field+".Key", CarouselInvalidOption)
		}
		keys[rule.Key] = true
		for _, n := range []*float32{rule.When.MinWidth, rule.When.MaxWidth, rule.When.MinHeight, rule.When.MaxHeight} {
			if n != nil && (!carouselFinite(*n) || *n < 0) {
				add(field+".When", CarouselInvalidOption)
				break
			}
		}
		if rule.When.Direction != nil && *rule.When.Direction > CarouselRTL {
			add(field+".Direction", CarouselInvalidOption)
		}
		if len(rule.Options.Breakpoints) > 0 {
			add(field+".Options.Breakpoints", CarouselInvalidOption)
		} else if err := rule.Options.Validate(); err != nil {
			problems = append(problems, err)
		}
	}
	return errors.Join(problems...)
}

func carouselCopyValue[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func carouselCloneOptions(o CarouselOptions) CarouselOptions {
	o.Align = carouselCopyValue(o.Align)
	o.Axis = carouselCopyValue(o.Axis)
	o.Direction = carouselCopyValue(o.Direction)
	o.Contain = carouselCopyValue(o.Contain)
	o.SlidesToScroll = carouselCopyValue(o.SlidesToScroll)
	o.StartIndex = carouselCopyValue(o.StartIndex)
	o.Active, o.Loop = carouselCopyValue(o.Active), carouselCopyValue(o.Loop)
	o.WatchDrag, o.WatchResize = carouselCopyValue(o.WatchDrag), carouselCopyValue(o.WatchResize)
	o.WatchSlides, o.WatchFocus = carouselCopyValue(o.WatchSlides), carouselCopyValue(o.WatchFocus)
	if o.InViewThreshold != nil {
		values := slices.Clone(*o.InViewThreshold)
		o.InViewThreshold = &values
	}
	if o.Nodes != nil {
		nodes := *o.Nodes
		nodes.Container = carouselCopyValue(nodes.Container)
		if nodes.Slides != nil {
			values := slices.Clone(*nodes.Slides)
			nodes.Slides = &values
		}
		o.Nodes = &nodes
	}
	if o.Breakpoints != nil {
		o.Breakpoints = slices.Clone(o.Breakpoints)
		for i := range o.Breakpoints {
			rule := &o.Breakpoints[i]
			rule.When.MinWidth, rule.When.MaxWidth = carouselCopyValue(rule.When.MinWidth), carouselCopyValue(rule.When.MaxWidth)
			rule.When.MinHeight, rule.When.MaxHeight = carouselCopyValue(rule.When.MinHeight), carouselCopyValue(rule.When.MaxHeight)
			rule.When.Dark, rule.When.Reduced = carouselCopyValue(rule.When.Dark), carouselCopyValue(rule.When.Reduced)
			rule.When.Direction = carouselCopyValue(rule.When.Direction)
			rule.Options = carouselCloneOptions(rule.Options)
		}
	}
	return o
}

func carouselMergeOptions(base, patch CarouselOptions) CarouselOptions {
	patch = carouselCloneOptions(patch)
	if patch.Align != nil {
		base.Align = patch.Align
	}
	if patch.Axis != nil {
		base.Axis = patch.Axis
	}
	if patch.Direction != nil {
		base.Direction = patch.Direction
	}
	if patch.Contain != nil {
		base.Contain = patch.Contain
	}
	if patch.SlidesToScroll != nil {
		base.SlidesToScroll = patch.SlidesToScroll
	}
	if patch.StartIndex != nil {
		base.StartIndex = patch.StartIndex
	}
	if patch.Active != nil {
		base.Active = patch.Active
	}
	if patch.Loop != nil {
		base.Loop = patch.Loop
	}
	if patch.WatchDrag != nil {
		base.WatchDrag = patch.WatchDrag
	}
	if patch.WatchResize != nil {
		base.WatchResize = patch.WatchResize
	}
	if patch.WatchSlides != nil {
		base.WatchSlides = patch.WatchSlides
	}
	if patch.WatchFocus != nil {
		base.WatchFocus = patch.WatchFocus
	}
	if patch.InViewThreshold != nil {
		base.InViewThreshold = patch.InViewThreshold
	}
	if patch.Nodes != nil {
		if base.Nodes == nil {
			base.Nodes = &CarouselNodes{}
		} else {
			base.Nodes = carouselCopyValue(base.Nodes)
		}
		if patch.Nodes.Container != nil {
			base.Nodes.Container = patch.Nodes.Container
		}
		if patch.Nodes.Slides != nil {
			base.Nodes.Slides = patch.Nodes.Slides
		}
	}
	for _, rule := range patch.Breakpoints {
		index := slices.IndexFunc(base.Breakpoints, func(old CarouselBreakpoint) bool { return old.Key == rule.Key })
		if index < 0 {
			base.Breakpoints = append(slices.Clone(base.Breakpoints), rule)
		} else {
			base.Breakpoints = slices.Clone(base.Breakpoints)
			base.Breakpoints[index].When = rule.When
			base.Breakpoints[index].Options = carouselMergeOptions(base.Breakpoints[index].Options, rule.Options)
		}
	}
	return base
}

// Resolved source defaults are values. They contain only an explicitly owned
// pure Aligner and public Handles, never frame hooks/Context/Element.
type CarouselResolvedOptions struct {
	Align                                           CarouselAlignment
	Axis                                            CarouselAxis
	Direction                                       CarouselDirection
	Contain                                         CarouselContain
	SlidesToScroll                                  CarouselSlideGroup
	StartIndex                                      int
	Active, Loop                                    bool
	WatchDrag, WatchResize, WatchSlides, WatchFocus bool
	InViewThreshold                                 []float64
	Container                                       ui.Handle
	HasContainer, HasSlides                         bool
	Slides                                          []ui.Handle
}

func carouselResolveOptions(o CarouselOptions, env CarouselEnvironment) CarouselResolvedOptions {
	resolved := CarouselResolvedOptions{Active: true, WatchDrag: true, WatchResize: true, WatchSlides: true, WatchFocus: true, SlidesToScroll: CarouselSlideGroup{Count: 1}, InViewThreshold: []float64{0}}
	apply := func(p CarouselOptions) {
		if p.Align != nil {
			resolved.Align = *p.Align
		}
		if p.Axis != nil {
			resolved.Axis = *p.Axis
		}
		if p.Direction != nil {
			resolved.Direction = *p.Direction
		}
		if p.Contain != nil {
			resolved.Contain = *p.Contain
		}
		if p.SlidesToScroll != nil {
			resolved.SlidesToScroll = *p.SlidesToScroll
		}
		if p.StartIndex != nil {
			resolved.StartIndex = *p.StartIndex
		}
		if p.Active != nil {
			resolved.Active = *p.Active
		}
		if p.Loop != nil {
			resolved.Loop = *p.Loop
		}
		if p.WatchDrag != nil {
			resolved.WatchDrag = *p.WatchDrag
		}
		if p.WatchResize != nil {
			resolved.WatchResize = *p.WatchResize
		}
		if p.WatchSlides != nil {
			resolved.WatchSlides = *p.WatchSlides
		}
		if p.WatchFocus != nil {
			resolved.WatchFocus = *p.WatchFocus
		}
		if p.InViewThreshold != nil {
			resolved.InViewThreshold = slices.Clone(*p.InViewThreshold)
			if len(resolved.InViewThreshold) == 0 {
				resolved.InViewThreshold = []float64{0}
			}
			slices.Sort(resolved.InViewThreshold)
			resolved.InViewThreshold = slices.Compact(resolved.InViewThreshold)
		}
		if p.Nodes != nil {
			if p.Nodes.Container != nil {
				resolved.Container, resolved.HasContainer = *p.Nodes.Container, true
			}
			if p.Nodes.Slides != nil {
				resolved.Slides, resolved.HasSlides = slices.Clone(*p.Nodes.Slides), true
			}
		}
	}
	apply(o)
	for _, rule := range o.Breakpoints {
		if rule.When.matches(env) {
			apply(rule.Options)
		}
	}
	return resolved
}

type CarouselPointerEvent struct {
	Position       ui.Point
	Kind           ui.PointerType
	Cancelled      bool
	Button, Clicks int
	Modifiers      ui.Modifiers
	Time           time.Time
	Pressure       float32
}
type CarouselResizeEvent struct{ Previous, Current []ui.Rect }
type CarouselSlidesEvent struct{ Previous, Current []ui.Handle }
type CarouselFocusEvent struct {
	Slide  int
	Handle ui.Handle
	ViaTab bool
	Time   time.Time
}

// Hooks are sampled only from the current synchronous Container build. A false
// result vetoes the default source action; it does not disable native layout.
type CarouselWatchHooks struct {
	BeforeDrag   func(*CarouselController, CarouselPointerEvent) bool
	BeforeResize func(*CarouselController, CarouselResizeEvent) bool
	BeforeSlides func(*CarouselController, CarouselSlidesEvent) bool
	BeforeFocus  func(*CarouselController, CarouselFocusEvent) bool
}
