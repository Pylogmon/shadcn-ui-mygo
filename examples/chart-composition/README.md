# Native chart composition

This application uses the kit's public `ChartContainer`, configuration and
Tooltip/Legend content helpers around raw MyGo `chart.Plot` primitives. It does
not call the high-level `display.Chart` API.

```sh
go run ./examples/chart-composition
make charts-app # macOS: build/Shadcn Charts.app
go run ./examples/chart-composition -inspect -snapshot /tmp/chart-light.png
go run ./examples/chart-composition -dark -inspect -snapshot /tmp/chart-dark.png
```

Hover a month or focus the plot and use Left/Right, Home/End or Escape. Inspection
is application-owned and remains available after the pointer leaves. Change
the month range with the two native slider thumbs, edit an inspected revenue
value, or keep a note while changing the range and theme. Tooltip controls switch
dot/line/dashed indicators and hide their default mark.

The target line, labels above bars and range filter are application compositions
using native RuleY, Anchor and RangeSlider. They are not implementations of
Recharts Brush, ReferenceLine or LabelList APIs. The plot deliberately keeps a
physical LTR timeline in RTL windows and uses no scale/rotation transform, since
raw MyGo Select normalizes local pointer positions using Bounds width. Native
axis labels and grid colors are styled explicitly.
