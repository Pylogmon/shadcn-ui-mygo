# Polar chart source examples

`go run ./examples/chart-polar` provides all 31 fixed shadcn Pie, Radar and
Radial examples using public `display.Chart`, polar composition, scoped config
and tooltip/legend content helpers. `recipes.json` preserves their actual data,
series order, colors and margins from shadcn commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`.

```sh
go run ./examples/chart-polar -example chart-pie-interactive
go run ./examples/chart-polar -example chart-radial-label -dark
go run ./examples/chart-polar -example chart-radar-grid-circle -snapshot /tmp/radar.png
```

The optional `ChartProps.Polar` selects the source profile. Its nil value keeps
the existing public polar defaults. Angles are degrees: 0 points right and
positive turns counterclockwise. `ChartRadiusDIP`, `ChartRadiusPercent` and
`Add` express fixed, responsive and relative sector radii. Shape resolvers return
pure value layers; native center, tick, label and dot callbacks run only during
the build. Draw and geometry animation retain copied geometry/colors and public
native glyphs, never application callbacks or Context.

Pie pointer inspection uses the actual current sector layers, including the active
outer ring and its empty gap. Radar and Radial use their source angular/radial
axis regions, so an unfilled part of a Radial track still reports its category.
A stacked Pie reports the currently inspected
ring; Page Up and Page Down change rings while arrows/Home/End inspect its
categories. The active-month source picker owns decorative selection separately
from transient inspection, keeping the center and expanded shape consistent.
Tooltip and accessible values always report original data during animation.

Radar dots move with the animated polygon; active dots use target points, and
sector labels appear after animation. Inside-start radial labels use native
shaped glyphs along the arc. Font metrics, text layout, icon appearance and edge
rasterization can differ from browser SVG. The source picker uses the native
styled Select label presentation; per-option browser color swatches are not
added to the existing Select API by this module.

Original 31 TSX examples, actual Recharts 3.8 CJS geometry-function output,
Sector/axis/grid source and licenses are archived in
`pkg/display/testdata/shadcn-chart`. This example is separate from the nine
Tooltip catalog recipes and the final Gallery integration stage.
