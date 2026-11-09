# Tooltip, documentation, RTL and Create source archive

The 19 unmodified TSX files in `apps/` are from shadcn commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. `manifest.json` records original
Git blob identities and SHA256 digests. The catalogue app list gives the 70
actual chart IDs; earlier Cartesian and Polar archives contain their 61 source
examples. This archive adds the nine Tooltip examples, eight Base docs examples
and Create's six-family example file.

The eight Recharts 3.8.0 published CJS files in `recharts/lib/` document actual
payload selection, `defaultIndex` coordinates, the 10px offset and boundary
flip, and reversed-axis keyboard direction. Their metadata and digests are in `recharts/manifest.json`. Recharts' MIT
license and verified npm metadata are already archived at `../polar/`.

Source Tooltip examples require running/swimming original data, config label
and icon resolution, line/no indicators, hidden/custom/formatted headings,
formatter units and the advanced sum from `item.payload`. Existing public
`ChartTooltipContent` handles their content; high-level Chart now supplies the
original metadata/row and optional source positioning.

`chart-rtl.tsx` explicitly reverses XAxis for Arabic/Hebrew. The native opt-in
mirrors category bands and axes, preserves series order and logical indices,
and maps pointer and directional keys to their physical positions. Existing
zero/default behavior is preserved.

The docs anatomy is explanatory content, not another chart family. Create's
six cards use the same public Cartesian/Polar features and inherit the host
palette. No fixed source file uses Brush or ReferenceLine; this stage does not
expand into the entire Recharts surface or extra animation APIs.
