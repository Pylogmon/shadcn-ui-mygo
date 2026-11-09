package conversation

import (
	"math"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/shimmer"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type TraceStatus string

const (
	Pending  TraceStatus = "pending"
	Running  TraceStatus = "running"
	Complete TraceStatus = "complete"
	Error    TraceStatus = "error"
)

type TraceStep struct {
	ID, Title, Description, Detail string
	Status                         TraceStatus
	Icon                           *ui.SVG
	Children                       []TraceStep
	OnAction                       func()
	// Open controls any nested disclosure directly. Nil uses TraceState.
	Open                  *bool
	Disabled, KeepMounted bool
	URL                   string
	Content               func(*ui.Context)
	OnOpenChange          func(bool)
}
type TraceState struct {
	// Expanded retains the compatible unique top-level ID entries.
	Expanded map[string]bool
	paths    map[string]bool
}
type TraceLineProps struct {
	Status   TraceStatus
	Label    string
	Disabled bool
	hovered  bool
}
type TraceTitleProps struct{ Status TraceStatus }
type TraceIconProps struct{ Suffix bool }
type TraceCollapsibleProps struct {
	Label                 string
	Status                TraceStatus
	Disabled, KeepMounted bool
	// Trigger builds icon/title/description slots; a chevron is added.
	Trigger      func(*ui.Context)
	Content      func(*ui.Context)
	OnOpenChange func(bool)
}

type traceScopeKey struct{}
type traceColorKey struct{}
type traceScope struct {
	active  bool
	status  uint8
	base    ui.Color
	hovered bool
}
type tracePart uint8

const (
	traceLineSlot tracePart = iota
	traceChevronSlot
	tracePanelSlot
	traceInnerSlot
	traceDetailSlot
	traceNestedSlot
)

type traceNodeKey string

func traceStatusNumber(status TraceStatus) uint8 {
	switch status {
	case Running:
		return 1
	case Error:
		return 2
	case Pending:
		return 3
	}
	return 0
}
func traceStatusFromNumber(status uint8) TraceStatus {
	switch status {
	case 1:
		return Running
	case 2:
		return Error
	case 3:
		return Pending
	}
	return Complete
}
func traceColor(c *ui.Context, status TraceStatus, hovered bool) ui.Color {
	t := theme.TokensOf(c)
	if status == Error {
		return t.Destructive
	}
	if hovered {
		return t.Foreground
	}
	return t.MutedForeground
}
func withTraceScope(c *ui.Context, status TraceStatus, base ui.Color, hovered bool, content func(*ui.Context)) {
	s := ui.Local(c.Root(), traceScopeKey{}, func() traceScope { return traceScope{} })
	previous := *s
	*s = traceScope{true, traceStatusNumber(status), base, hovered}
	defer func() { *s = previous }()
	if content != nil {
		content(c)
	}
}
func traceInheritedStyle(c *ui.Context, override TraceStatus) (TraceStatus, ui.Color) {
	s := ui.Local(c.Root(), traceScopeKey{}, func() traceScope { return traceScope{} })
	if override != "" {
		return override, traceColor(c, override, s.active && s.hovered)
	}
	if s.active {
		return traceStatusFromNumber(s.status), s.base
	}
	return Complete, traceColor(c, Complete, false)
}

// This fallback estimates the default shimmer band only. Native glyphs keep
// their actual inherited family and weight, including later fluent overrides.
func traceMeasurementFont(c *ui.Context) ui.Font {
	family := c.Theme().Font
	if family == "" {
		family = "system-ui"
	}
	return ui.Font{Family: family, Size: 14, Weight: 400, LineHeight: 10.0 / 7}
}

// TraceGroup creates the source's min-width-zero vertical group, without
// automatic actions, status icons, or live announcements.
func TraceGroup(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ui.Column(c).FillWidth().MinWidth(0).Gap(8).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

// TraceLine is static. TraceAction, TraceLink and TraceCollapsible provide
// actual native interaction independently of its visual status.
func TraceLine(c *ui.Context, p TraceLineProps, content func(*ui.Context)) ui.Element {
	base := traceColor(c, p.Status, p.hovered)
	line := ui.Row(c).AlignSelf(ui.Start).Shrink(1).MaxWidthPercent(100).MinWidth(0).Gap(8).AlignItems(ui.Center).
		FontSize(14).LineHeight(10.0 / 7).TextColor(base)
	base = traceAnimatedColor(c, line, base)
	line.TextColor(base)
	if p.Status == Pending {
		line.Opacity(.6)
	}
	return line.Children(func() { withTraceScope(c, p.Status, base, p.hovered, content) })
}
func traceAnimatedColor(c *ui.Context, e ui.Element, target ui.Color) ui.Color {
	if motion.Reduced(c) {
		return target
	}
	return e.AnimateColor(traceColorKey{}, target, ui.Animation{Duration: 150 * time.Millisecond, Ease: ui.CubicBezier(.4, 0, .2, 1)})
}
func traceControl(c *ui.Context, e ui.Element, p TraceLineProps) ui.Element {
	t := theme.TokensOf(c)
	e.Row().AlignSelf(ui.Start).Shrink(1).MaxWidthPercent(100).MinWidth(0).AlignItems(ui.Center).Radius(t.RadiusMD()).
		DimDisabled(false).FocusRing(false).Cursor(ui.CursorPointer)
	if p.Label != "" {
		e.Label(p.Label)
	}
	if e.FocusVisible() {
		color, radius := t.Ring.Alpha(.5), t.RadiusMD()
		e.DrawOver(func(painter *ui.Painter, r ui.Rect) {
			painter.Stroke(ui.Rect{X: r.X - 1.5, Y: r.Y - 1.5, W: r.W + 3, H: r.H + 3}, color, radius+1.5, 3)
		})
	}
	return e
}

// TraceAction returns a native button; inspect Clicked in the live build.
func TraceAction(c *ui.Context, p TraceLineProps, content func(*ui.Context)) ui.Element {
	button := traceControl(c, ui.ButtonBase(c).Disabled(p.Disabled), p).Gap(0)
	p.hovered = button.Hovered()
	return button.Children(func() { TraceLine(c.Key(traceLineSlot), p, content) })
}
func TraceIcon(c *ui.Context, icon *ui.SVG, props ...TraceIconProps) ui.Element {
	p := TraceIconProps{}
	if len(props) > 0 {
		p = props[0]
	}
	root := ui.Box(c).Size(16, 16).Shrink(0).Center().AccessHidden(true).IgnoreInput(true)
	if p.Suffix {
		root.Margin(0, 0, 0, -4)
	}
	return root.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).Size(16, 16).AccessHidden(true).IgnoreInput(true)
		}
	})
}

// TraceTitle inherits its direct line's status and hover color. Complete can
// explicitly override a running title to neutral. Native truncation keeps AX.
func TraceTitle(c *ui.Context, text string, props ...TraceTitleProps) ui.Element {
	p := TraceTitleProps{}
	if len(props) > 0 {
		p = props[0]
	}
	status, base := traceInheritedStyle(c, p.Status)
	measure := traceMeasurementFont(c)
	title := shimmer.Text(c, text, shimmer.Props{Active: status == Running && text != "", Base: base, MeasureFont: &measure}).
		MinWidth(0).Shrink(1).SingleLine()
	if text == "" {
		title.Size(0, 0)
	}
	return title
}
func TraceDescription(c *ui.Context, text string) ui.Element {
	description := ui.Text(c, text).
		TextColor(theme.TokensOf(c).MutedForeground.Alpha(.7)).FontFeatures("tnum").Shrink(0).SingleLine()
	if text == "" {
		description.Size(0, 0)
	}
	return description
}

// TraceLink keeps the actual native URL action. Native Link has no compound
// LinkBase and owns its hover underline; decorative siblings belong in a Line.
func TraceLink(c *ui.Context, text, url string, props ...TraceLineProps) ui.Element {
	p := TraceLineProps{}
	if len(props) > 0 {
		p = props[0]
	}
	status, base := traceInheritedStyle(c, p.Status)
	link := ui.Link(c, text, url).FontSize(14).LineHeight(10.0 / 7).TextColor(base).
		MinWidth(0).Shrink(1).SingleLine().Disabled(p.Disabled).DimDisabled(false)
	if p.Label != "" {
		link.Label(p.Label)
	}
	if p.Status == Pending {
		link.Opacity(.6)
	}
	if link.Hovered() {
		base = traceColor(c, status, true)
	}
	link.TextColor(traceAnimatedColor(c, link, base))
	return link
}

// TraceCollapsibleTrigger styles an existing native trigger. Its content is
// line slots; the source chevron has its own four-DIP gap and physical rotation.
func TraceCollapsibleTrigger(c *ui.Context, trigger ui.Element, status TraceStatus, progress float32, content func(*ui.Context)) ui.Element {
	if math.IsNaN(float64(progress)) || math.IsInf(float64(progress), 0) {
		progress = 0
	}
	progress = max(0, min(1, progress))
	p := TraceLineProps{Status: status}
	traceControl(c, trigger, p).Gap(4)
	p.hovered = trigger.Hovered()
	trigger.TextColor(traceAnimatedColor(c, trigger, traceColor(c, status, p.hovered)))
	return trigger.Children(func() {
		TraceLine(c.Key(traceLineSlot), p, content)
		ui.Box(c.Key(traceChevronSlot)).Size(16, 16).Shrink(0).Center().AccessHidden(true).IgnoreInput(true).Children(func() {
			ui.Icon(c, traceChevronIcon).Size(16, 16).Rotate(90 * progress).AccessHidden(true).IgnoreInput(true)
		})
	})
}

// TraceCollapsibleContent styles the source inner panel. Its indentation and
// border remain physically left in RTL. The main component owns animated clip.
func TraceCollapsibleContent(c *ui.Context, content func(*ui.Context)) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Shrink(0).Margin(8, 0, 4, 7).BorderWidth(0, 0, 0, 1).BorderColor(t.Border).
		Padding(0, 4, 0, 16).LineHeight(1.625).TextColor(t.MutedForeground).
		Children(func() {
			if content != nil {
				content(c)
			}
		})
}

// TraceCollapsible is controlled by caller state. KeepMounted is lazy until
// its first opening. Closed content leaves input/AX immediately while native
// measured height finishes the reversible 200ms disclosure animation.
func TraceCollapsible(c *ui.Context, open *bool, p TraceCollapsibleProps) ui.Element {
	if open == nil {
		panic("conversation: TraceCollapsible requires caller-owned open state")
	}
	root := ui.Column(c).FillWidth().MinWidth(0).Shrink(0).Gap(0)
	return root.Children(func() {
		part := ui.CollapsibleBase(c, open)
		trigger := part.Trigger.Role(ui.RoleButton).Disabled(p.Disabled).DimDisabled(false).Label(p.Label)
		TraceCollapsibleTrigger(c, trigger, p.Status, part.Progress(), p.Trigger)
		if trigger.Changed() {
			root.SetChanged()
			if p.OnOpenChange != nil {
				p.OnOpenChange(*open)
			}
		}
		mounted := ui.Local(trigger, "trace.has-opened", func() bool { return false })
		if *open {
			*mounted = true
		}
		panelHandle := ui.Local(trigger, "trace.panel-handle", func() ui.Handle { return ui.Handle{} })
		restore := !*open && panelHandle.FocusWithin(c)
		progress := part.Progress()
		if !*mounted || !*open && progress <= 0 && !p.KeepMounted {
			if restore {
				trigger.Focus()
			}
			return
		}
		panel := ui.Column(c.Key(tracePanelSlot)).FillWidth().MinWidth(0).Shrink(0).ClipY().Role(ui.RoleRegion).Label(p.Label).
			AccessHidden(!*open).IgnoreInput(!*open).Disabled(!*open).DimDisabled(false).Bind(panelHandle)
		height := ui.Local(panel, "trace.natural-height", func() float32 { return 0 })
		services := c.Services()
		panel.Children(func() {
			inner := TraceCollapsibleContent(c.Key(traceInnerSlot), p.Content)
			// A zero-height clip does not paint its child, so its Draw cannot
			// seed the first opening. Bounds reads the preceding native layout
			// while this Element is still live; only the number survives.
			if measured := inner.LayoutBounds().H; measured >= 0 && math.Abs(float64(*height-measured-12)) > .1 {
				*height = measured + 12
			}
			inner.Draw(func(_ *ui.Painter, r ui.Rect) {
				// Bounds exclude margins; mt8+mb4 belong in animated panel height.
				natural := r.H + 12
				if math.Abs(float64(*height-natural)) > .1 {
					*height = natural
					services.Invalidate()
				}
			})
		})
		trigger.Controls(panel)
		panel.LabelledBy(trigger)
		if progress < 1 {
			panel.Height(max(0, progress) * *height)
		}
		if restore {
			trigger.Focus()
		}
	})
}

func tracePath(parent, id string, index, ordinal int) string {
	if id == "" {
		return parent + "/i" + strconv.Itoa(index)
	}
	return parent + "/k" + strconv.Itoa(len(id)) + ":" + id + ":" + strconv.Itoa(ordinal)
}
func Trace(c *ui.Context, state *TraceState, steps []TraceStep) ui.Element {
	if state == nil {
		panic("conversation: Trace requires caller-owned state")
	}
	if state.Expanded == nil {
		state.Expanded = map[string]bool{}
	}
	if state.paths == nil {
		state.paths = map[string]bool{}
	}
	return traceSteps(c, state, steps, "").Role(ui.RoleGroup).Label("Trace")
}
func traceSteps(c *ui.Context, state *TraceState, steps []TraceStep, parent string) ui.Element {
	counts := map[string]int{}
	for _, step := range steps {
		counts[step.ID]++
	}
	seen := map[string]int{}
	return TraceGroup(c, func(child *ui.Context) {
		for i, step := range steps {
			ordinal := seen[step.ID]
			seen[step.ID]++
			path := tracePath(parent, step.ID, i, ordinal)
			ctx := child.Key(traceNodeKey(path))
			line := func(lineContext *ui.Context) {
				if step.Icon != nil {
					TraceIcon(lineContext, step.Icon)
				}
				TraceTitle(lineContext, step.Title)
				if step.Description != "" {
					TraceDescription(lineContext, step.Description)
				}
			}
			expandable := step.Detail != "" || step.Content != nil || len(step.Children) > 0
			if expandable {
				values, key := state.paths, path
				if parent == "" && step.ID != "" && counts[step.ID] == 1 {
					values, key = state.Expanded, step.ID
				}
				value := values[key]
				open := step.Open
				if open == nil {
					open = &value
				}
				TraceCollapsible(ctx, open, TraceCollapsibleProps{Label: step.Title, Status: step.Status, Disabled: step.Disabled, KeepMounted: step.KeepMounted, Trigger: line,
					OnOpenChange: func(next bool) {
						values[key] = next
						if step.OnOpenChange != nil {
							step.OnOpenChange(next)
						}
						if step.OnAction != nil {
							step.OnAction()
						}
						if step.Open == nil {
							*open = values[key]
						}
					},
					Content: func(detail *ui.Context) {
						if step.Detail != "" {
							ui.Text(detail.Key(traceDetailSlot), step.Detail).Selectable()
						}
						if step.Content != nil {
							step.Content(detail)
						}
						if len(step.Children) > 0 {
							traceSteps(detail.Key(traceNestedSlot), state, step.Children, path)
						}
					},
				})
				values[key] = *open
			} else if step.URL != "" {
				TraceLine(ctx, TraceLineProps{Status: step.Status}, func(linkContext *ui.Context) {
					if step.Icon != nil {
						TraceIcon(linkContext, step.Icon)
					}
					TraceLink(linkContext, step.Title, step.URL, TraceLineProps{Disabled: step.Disabled})
					if step.Description != "" {
						TraceDescription(linkContext, step.Description)
					}
				})
			} else if step.OnAction != nil {
				if TraceAction(ctx, TraceLineProps{Status: step.Status, Label: step.Title, Disabled: step.Disabled}, line).Clicked() {
					step.OnAction()
				}
			} else {
				TraceLine(ctx, TraceLineProps{Status: step.Status}, line)
			}
		}
	})
}
