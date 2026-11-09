# Cartesian chart source examples

Run `go run ./examples/chart-cartesian`, then choose one of the 30 fixed
shadcn Area, Bar or Line examples. `internal/chartcatalog/cartesian.json` preserves the source data,
series order, colors and explicit margins from commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. Original TSX files are archived in
`pkg/display/testdata/shadcn-chart`.

```sh
go run ./examples/chart-cartesian -example chart-area-gradient
go run ./examples/chart-cartesian -example chart-bar-label-custom -snapshot /tmp/bars.png
go run ./examples/chart-cartesian -example chart-line-dots-custom -dark -inspect
```

The application uses public `display.Chart`, per-series styling and the scoped
configuration/tooltip/legend helpers. The date-range Area and Desktop/Mobile
Bar/Line header controls change actual data. Focus the plot for arrows,
Home/End and Escape; native inspection always reports original values during
animation. Disable “Animate chart changes” for a completed static view.

Custom dot and active-bar callbacks style real native mark elements during the
build. Their Draw closures capture copied color/geometry only. Natural paths
use the archived official D3 algorithm and an independently generated oracle;
linear, step and monotone use public native paths. Labels use native font
metrics and appear after animation. This is native rendering, so text metrics,
icon drawing, tick thinning and gradient rasterization can differ from SVG.
The example's small trend and commit symbols are authored native paths.

The configuration builders are shared with Gallery and `examples/chart-catalog`,
which also provide the 31 Polar recipes, nine Tooltip recipes, eight docs
workflows and six Create cards. No Brush or ReferenceLine API is implied by
the fixed source collection.

Run this example from this module: its shared recipe builders live in
`internal/chartcatalog`, so the complete main.go is not a standalone external
module sample. `examples/chart-composition` remains the self-contained example
using the public packages directly.
