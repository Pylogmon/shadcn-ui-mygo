# Disclosure source reference

The two complete wrappers are unmodified originals from fixed shadcn commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The manifest records their Git blobs,
byte lengths and SHA-256 values, and reuses the archived Vega stylesheet.
The original shadcn MIT license is included unchanged.

Accordion styles define separators, 16-DIP vertical trigger padding, a fixed
16-DIP decorative icon and a 16-DIP inner content bottom padding. Collapsible's
wrapper exposes separate Root/Trigger/Content without imposing a fixed layout.
The mapped seven Accordion and five Collapsible documentation examples, plus
their two Create collections, are recorded under
`examples/disclosure-composition/source-map.json`.

Go keeps MyGo's public native CollapsibleBase for input, expanded state and
200ms reversible height progress. CollapsibleRoot separates the trigger and
content in the same current build. Content may be declared first, including in
a grid; its builder executes once at Root's end, after native trigger input.
A hidden themed reentry node supplies the declaration's Theme metadata while
children are attached to their actual plain inner container. It is not a text
ancestor, so ordinary fluent font/size/color/direction remain inherited.
Transient scope references/builders are cleared before returning. Only mount,
height and public Handle data survive frames.

The clipping root initially has zero height, where the renderer can cull inner
Draw callbacks. Public inner LayoutBounds supplies the previous measured layout
height before the next animation sample. This seed is used by both new and
legacy paths; the focused native workflow observes an intermediate first-open
height and demonstrated the legacy zero-height failure before this correction.

New content defaults to zero padding, allowing the file tree and radius grid to
match their source layouts; legacy CollapsibleWith keeps its existing 16-DIP
bottom padding. Accordion's current-build StyleItem exposes real item, trigger
and inner content styling without splitting native header navigation. Its
Content is invalid before first mounting. Both old disclosure arrows now reuse
the already pinned Lucide chevron in `internal/indicator`, rather than font
characters. No new icon asset or font is added.

Retained hidden children with unconditional explicit Focus/AutoFocus still need
application conditioning on open, as documented for the legacy entry. Native
test relationships, synthetic input and headless pixels do not establish real
screen-reader, platform, touch or IME acceptance.
