# Independent disclosure composition

`display.CollapsibleRoot` separates a native disclosure's trigger from its
content. Both live in the current window; no popup or separate window is needed.
The caller owns the open bool and application drafts.

```go
display.CollapsibleRoot(c, &model.Expanded, display.CollapsibleRootProps{
    Label: "Radius settings", KeepMounted: true,
}, func(s *display.CollapsibleScope) {
    ui.Row(c).Gap(8).AlignItems(ui.Start).Children(func() {
        ui.Grid(c).Grow(1).Columns(2).Gap(8).Children(func() {
            control.Input(c.Key("x"), &model.X, control.InputProps{Label: "Radius X"})
            control.Input(c.Key("y"), &model.Y, control.InputProps{Label: "Radius Y"})
            parts := s.Content(c, display.CollapsibleContentProps{}, func(child *ui.Context) {
                control.Input(child, &model.DetailX, control.InputProps{Label: "Detail X"})
                control.Input(child, &model.DetailY, control.InputProps{Label: "Detail Y"})
            })
            parts.Root.ColumnSpan(2)
            parts.Content.Grid().Columns(2).Gap(8)
        })
        s.Trigger(c, display.CollapsibleTriggerProps{
            Label: "Expand radius settings",
            Style: func(tr ui.Element) {
                control.StyleButton(c, tr, "", control.ButtonProps{
                    Variant: control.Outline, Size: control.IconSmall,
                })
            },
        }, func(child *ui.Context) {
            material.Icon(child, outlined.UnfoldMore(), material.Props{Size: 16})
        })
    })
})
```

The two ordinary editors are outside the disclosure button. Content can be
declared before Trigger and inside a different native grid or column. It returns
the actual clipping `Root` and inner `Content`: apply grid span/width/alignment
to Root and padding/gap to Content, whose natural height drives the animation.
New content has no imposed padding. The legacy `CollapsibleWith` retains its
existing trigger-then-panel layout and 16-DIP bottom padding.

Each Root declares one Trigger and at most one Content. Scope methods are valid
only during that Root callback; retaining one and calling it later panics. The
Root clears its temporary Context, Element, native parts and builder references
before returning. Content executes once at the end of the same current build,
after native trigger input. It does not queue a handler for a later frame.

The actual children keep the declaration's scoped native Theme, font roles, menu
profile and direction while inheriting ordinary ancestor font, size and color.
The hidden theme reentry helper occupies no layout/AX space and is not a text
ancestor of those children. Content must use the Context passed to its builder.
Input binding, keyboard activation, expanded state and 200ms reversible progress
come from public MyGo CollapsibleBase. `Open` and `Progress` can build decorative
icons; the default chevron is a fixed 16-DIP SVG shared with Select.

Initially closed content builds lazily. `KeepMounted` retains it after the first
opening, preserving editor selection and local state. Closing immediately
excludes content from input and accessibility, restores trigger focus when the
panel owned it, and retains pixels until height exit completes. Fully closed
content uses native Hidden, so a closed grid item leaves no row or gap. Reduced
motion converges immediately. Explicit Focus/AutoFocus in retained children must
be conditional on open; the SDK's direct focus operations are an existing
boundary, not a promise of browser inert enforcement.

The initial zero-height clip can prevent a Draw callback from publishing natural
height. Both the new Root and legacy Accordion/Collapsible panel now seed that
measurement from the inner element's public previous-frame LayoutBounds, so a
first lazy opening has an actual intermediate height rather than appearing only
after the 200ms animation completes. Draw continues tracking later size changes.

`AccordionProps.StyleItem` is a synchronous current-build style callback receiving
`AccordionItemParts{Key,Open,Root,Trigger,Content}`. Content is the actual inner
panel, invalid while not mounted. It allows a single outer border, explicit
trigger/panel padding or per-item open color while preserving the same native
Up/Down/Home/End header navigation. Do not retain Parts or callbacks in a model.
Existing Accordion defaults, controlled values and callbacks remain.

The Gallery source recipes are documented in
[Disclosure examples](../examples/disclosure-composition/README.md). Original
wrappers, license and Vega reference are pinned in
[`testdata/shadcn-disclosure`](../pkg/display/testdata/shadcn-disclosure/README.md).
The native layout/state bridge is a Go adaptation rather than a DOM alias.
Headless native tests do not replace real platform, screen-reader or input-method
acceptance.
