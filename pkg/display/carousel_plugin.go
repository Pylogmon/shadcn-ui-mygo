package display

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

// CarouselPlugin is an application-owned plugin identity. Construct custom
// plugins with NewCarouselPlugin; Autoplay implements the same typed contract.
// Repeated identities/names preserve slice init/destroy order. Plugins returns
// the last value for each name, as in the source.
type CarouselPlugin interface {
	Name() string
	carouselPluginDefinition() *carouselPluginDefinition
}

type CarouselPluginOptions struct {
	Active      *bool
	Breakpoints []CarouselPluginBreakpoint
}
type CarouselPluginBreakpoint struct {
	Key    string
	When   CarouselMedia
	Active *bool
}

type CarouselPluginHooks struct {
	Init    func(CarouselPluginFrame)
	Destroy func(CarouselPluginFrame)
	Events  []CarouselEventType
	Event   func(CarouselPluginFrame, CarouselEventType)
	Frame   func(CarouselPluginFrame)
}

// CarouselPluginFrame contains values and a callback-scoped action facade.
// Hooks are explicitly application-owned and must never capture UI frame objects.
// Actions expire when the hook returns, even if the caller retains this value.
// External controller/plugin commands queue; these actions run synchronously.
type CarouselPluginFrame struct {
	Now                  time.Time
	Environment          CarouselEnvironment
	Snapshot             CarouselControllerSnapshot
	Root                 ui.Handle
	Hovered, FocusWithin bool
	Actions              CarouselPluginActions
}

type CarouselPluginActions struct{ scope *carouselPluginActionScope }
type carouselPluginActionScope struct {
	frame   *carouselControllerBuild
	binding *carouselPluginBinding
	alive   bool
}

func (a CarouselPluginActions) valid() bool {
	return a.scope != nil && a.scope.alive && a.scope.frame != nil
}
func (a CarouselPluginActions) ScrollTo(index int, jump bool) bool {
	if !a.valid() {
		return false
	}
	f := a.scope.frame
	f.scroll(index, 0, jump, f.now)
	return true
}
func (a CarouselPluginActions) ScrollNext(jump bool) bool {
	if !a.valid() {
		return false
	}
	f := a.scope.frame
	f.scroll(f.engine.selected+1, -1, jump, f.now)
	return true
}
func (a CarouselPluginActions) ScrollPrev(jump bool) bool {
	if !a.valid() {
		return false
	}
	f := a.scope.frame
	f.scroll(f.engine.selected-1, 1, jump, f.now)
	return true
}
func (a CarouselPluginActions) Emit(event CarouselEventType) bool {
	if !a.valid() {
		return false
	}
	if a.scope.frame.api.Snapshot().Destroyed {
		a.scope.frame.api.Emit(event)
	} else {
		a.scope.frame.publish(event)
	}
	return true
}

// Autoplay actions use the current hook's clock and synchronous event order.
func (a CarouselPluginActions) AutoplayPlay(plugin *CarouselAutoplay, jump ...bool) bool {
	if !a.valid() || plugin == nil {
		return false
	}
	var override *bool
	if len(jump) > 0 {
		override = CarouselValue(jump[0])
	}
	return a.scope.frame.autoplayCommand(carouselCommand{plugin: &plugin.definition, pluginAction: "play", jumpOverride: override})
}
func (a CarouselPluginActions) AutoplayStop(plugin *CarouselAutoplay) bool {
	if !a.valid() || plugin == nil {
		return false
	}
	return a.scope.frame.autoplayCommand(carouselCommand{plugin: &plugin.definition, pluginAction: "stop"})
}
func (a CarouselPluginActions) AutoplayReset(plugin *CarouselAutoplay) bool {
	if !a.valid() || plugin == nil {
		return false
	}
	return a.scope.frame.autoplayCommand(carouselCommand{plugin: &plugin.definition, pluginAction: "reset"})
}

func (a CarouselPluginActions) After(delay time.Duration) bool {
	if !a.valid() {
		return false
	}
	a.scope.binding.deadline = a.scope.frame.now.Add(max(0, delay))
	return true
}

type CarouselCustomPlugin struct{ definition carouselPluginDefinition }
type carouselPluginDefinition struct {
	mu         sync.Mutex
	name       string
	options    CarouselPluginOptions
	hooks      CarouselPluginHooks
	autoplay   *CarouselAutoplay
	owner      *CarouselController
	generation uint64
}

func NewCarouselPlugin(name string, options CarouselPluginOptions, hooks CarouselPluginHooks) (*CarouselCustomPlugin, error) {
	if name == "" {
		return nil, &CarouselOptionError{Field: "Plugins.Name", Code: CarouselInvalidOption}
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}
	hooks.Events = slices.Clone(hooks.Events)
	return &CarouselCustomPlugin{definition: carouselPluginDefinition{name: name, options: carouselClonePluginOptions(options), hooks: hooks}}, nil
}
func (p *CarouselCustomPlugin) Name() string {
	if p == nil {
		return ""
	}
	return p.definition.name
}
func (p *CarouselCustomPlugin) carouselPluginDefinition() *carouselPluginDefinition {
	if p == nil {
		return nil
	}
	return &p.definition
}
func (o CarouselPluginOptions) Validate() error {
	keys := map[string]bool{}
	for _, rule := range o.Breakpoints {
		if rule.Key == "" || keys[rule.Key] {
			return &CarouselOptionError{Field: "Plugins.Breakpoints.Key", Code: CarouselInvalidOption}
		}
		keys[rule.Key] = true
		if err := (CarouselOptions{Breakpoints: []CarouselBreakpoint{{Key: rule.Key, When: rule.When}}}).Validate(); err != nil {
			return err
		}
	}
	return nil
}
func carouselClonePluginOptions(o CarouselPluginOptions) CarouselPluginOptions {
	o.Active = carouselCopyValue(o.Active)
	o.Breakpoints = slices.Clone(o.Breakpoints)
	for i := range o.Breakpoints {
		r := &o.Breakpoints[i]
		r.Active = carouselCopyValue(r.Active)
		r.When = carouselCloneMedia(r.When)
	}
	return o
}
func carouselCloneMedia(m CarouselMedia) CarouselMedia {
	return carouselCloneOptions(CarouselOptions{Breakpoints: []CarouselBreakpoint{{When: m}}}).Breakpoints[0].When
}
func carouselPluginActive(o CarouselPluginOptions, env CarouselEnvironment) bool {
	active := true
	if o.Active != nil {
		active = *o.Active
	}
	for _, r := range o.Breakpoints {
		if r.When.matches(env) && r.Active != nil {
			active = *r.Active
		}
	}
	return active
}
func carouselValidatePlugins(plugins []CarouselPlugin) error {
	for i, p := range plugins {
		if p == nil || p.carouselPluginDefinition() == nil {
			return &CarouselOptionError{Field: fmt.Sprintf("Plugins[%d]", i), Code: CarouselInvalidOption}
		}
	}
	return nil
}

// SetPlugins replaces the ordered plugin list in the next owner build. A
// non-nil empty slice clears it. The engine keeps its current selected snap.
func (a *CarouselController) SetPlugins(plugins []CarouselPlugin) error {
	return a.ReInitWithPlugins(CarouselOptions{}, plugins)
}
func (a *CarouselController) ReInitWithPlugins(options CarouselOptions, plugins []CarouselPlugin) error {
	if err := options.Validate(); err != nil {
		return err
	}
	if err := carouselValidatePlugins(plugins); err != nil {
		return err
	}
	a.mu.Lock()
	if a.snapshot.Destroyed {
		a.mu.Unlock()
		return &CarouselOptionError{Field: "Controller", Code: CarouselDestroyed}
	}
	a.options = carouselMergeOptions(a.options, options)
	a.desiredPlugins = slices.Clone(plugins)
	if a.bound {
		a.commands = append(a.commands, carouselCommand{kind: carouselCommandPlugins, patch: carouselCloneOptions(options), plugins: slices.Clone(plugins)})
	}
	services := a.services
	a.mu.Unlock()
	services.Invalidate()
	return nil
}
func (a *CarouselController) Plugins() map[string]CarouselPlugin {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make(map[string]CarouselPlugin, len(a.pluginMap))
	for key, p := range a.pluginMap {
		out[key] = p
	}
	return out
}
