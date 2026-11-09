# Chart catalogue

`go run ./examples/chart-catalog` opens the fixed shadcn catalogue: 70 chart
examples, eight documentation examples and the six chart cards used in Create.
Gallery's Data & display page uses the same `internal/chartcatalog` builders.
The Cartesian and Polar apps retain their original commands and controls.

```sh
go run ./examples/chart-catalog -example chart-tooltip-advanced
go run ./examples/chart-catalog -collection docs -example docs-chart-rtl
go run ./examples/chart-catalog -collection create
go run ./examples/chart-catalog -collection create -export-height 4000 -export /tmp/create.png
go run ./examples/chart-catalog -example chart-tooltip-icons -export /tmp/tooltip.png
```

Filter the directory by family or search its titles/source IDs. The Area date
range, Bar/Line and docs Desktop/Mobile totals, Pie active month and RTL language
controls change the real plot. Create shows all six cards and inherits its
host's live palette and appearance. Inspection remains transient; Escape does
not reapply a Tooltip example's initial `defaultIndex=1`.

`ChartProps.Data` and optional `ChartSeries.Data` carry original rows into
`ChartTooltipData.Payload`. `DataKey` and `TooltipName` preserve config lookup
metadata independently of human-readable series names. JSON-shaped `map[string]any`, `[]any`, `[]map[string]any` and
`[]string`/`[]float64`/`[]int` values are detached for formatters; marks and inspection always read current
`Series.Values`, including during geometry animation. Opaque application
objects are borrowed; source rows must be acyclic JSON-shaped values to be
recursively detached.

`TooltipFollowInspection` opts into the fixed source's 10-DIP offset and edge
flip, using actual native content dimensions. Its zero value retains the
existing fixed tooltip corner. `Cartesian.ReversedCategory` opts into the RTL
documentation example's reversed category axis while retaining logical indices
and series order. The default physical category order stays unchanged.

The tooltip anatomy uses the archived source arrow paths with native layout.
The icons example uses on-demand Material Symbols Footprint and Waves, keeping
the kit's icon integration. Native fonts, shaping, rasterization and responsive
wrapping can differ from browser SVG/CSS. The catalogue inventory/build test
is one coverage check; it is not 84 independent visual acceptance tests.
