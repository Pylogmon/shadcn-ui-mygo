# Carousel reference and geometry fixtures

This directory is an audit/test archive. Production Go code does not load or
execute the JavaScript, npm packages, or JSON fixtures.

The shadcn reference is `shadcn-ui/ui` commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. `manifest.json` records each
original path, archived path, Git blob ID, and SHA256. The archive includes the
Base primitive, Base documentation, all eight independent examples, the Base
example collection, Vega styles, the original app package/lock files, and MIT
license. These bytes were checked against the fixed Git tree.

All four Embla packages are version **8.5.2**, pinned by that original lock.
Each complete npm tarball was checked against both the registry metadata's
`dist.integrity` and the lock's SHA512 value. The manifest records tarball
SHA256 and extracted source/type file SHA256 values. The formal Go integrity
test verifies these checks and matches extracted bytes to their tar members.

The npm tarballs do **not** contain a LICENSE file. `npm/LICENSE.embla` was
downloaded separately from the official repository at the packages' published
`gitHead`, `5ee3c4ed0251dc685198d0ac21dc3e754b865476`:

<https://raw.githubusercontent.com/davidjerleke/embla-carousel/5ee3c4ed0251dc685198d0ac21dc3e754b865476/LICENSE>

Its SHA256 is
`992490064b90ec94c302fe5f29dcc14cb44145857064e46f9ffe6c8f511c1093`.
The original MIT copyright and permission notice must accompany distributions
containing the source or substantial native translations of its algorithms.

## Independent upstream oracle

`generate-oracle.cjs` executes the archived, verified official CJS functions
through Node's VM: Axis, Alignment, SlideSizes, SlidesToScroll, ScrollSnaps,
ScrollContain, SlideRegistry, and SlideLooper. It does not import Go product
code or translate these functions into another reference implementation.
Inputs are synthetic measured rectangles derived from the source's compensated
track/leading-padding layout; this is a geometry oracle, not a browser render.

The complete generator produces **1,296 cases**:

- Six item configurations: full, half, third, mixed fixed sizes, mixed
  fractions, and two half-size items with insufficient loop content.
- Spacing 0, 4, and 16.
- Horizontal LTR/RTL and vertical LTR/RTL.
- Start, center, and end alignment.
- Trim snaps, keep snaps, and no containment.
- Loop requested on/off, with upstream loop fallback resolved independently.

`geometry-oracle.json` contains the **324 LTR cases selected unchanged** from
that complete result. Logical sizes, snap positions, containment, and loop
resolution are independent of the physical axis/direction. Each fixture also
retains the upstream loop shifts, outer/content rectangles, and slide registry
at every snap. The formal Go geometry test compares all 324 cases, including
the visible outer-item geometry. Native tester checks separately exercise
actual horizontal/vertical/RTL layout, clipping, pixels, and input/AX behavior.
Very small fixed/fractional bases use at least the leading padding as their
outer used size, matching CSS border-box and native padded-box layout. Separate
native tests measure this floor directly; the source-function fixtures start
from measured rectangles and do not replace that layout check.

Regenerate and verify the committed selection without writing generated files
into the repository:

```sh
carousel_audit_dir=$(mktemp -d)
node pkg/display/testdata/shadcn-carousel/generate-oracle.cjs \
  pkg/display/testdata/shadcn-carousel/npm/embla-carousel/cjs/embla-carousel.cjs.js \
  "$carousel_audit_dir/full.json"
python3 - pkg/display/testdata/shadcn-carousel/geometry-oracle.json \
  "$carousel_audit_dir/full.json" <<'PY'
import json, sys
committed = json.load(open(sys.argv[1]))
generated = json.load(open(sys.argv[2]))
assert generated['caseCount'] == 1296
assert committed['sourceSHA256'] == generated['sourceSHA256']
assert committed['cases'] == [c for c in generated['cases'] if c['mode'] == 'ltr']
print('324 committed cases match the complete upstream execution')
PY
go test ./pkg/display -run 'TestCarousel(PinnedSource|Geometry)'
```

`fullOracleSHA256` identifies the original independent capture. A regenerated
full file can have a different file hash because its header contains the CJS
source path; compare source SHA256 and decoded cases as above.

## Phase scope

The current native composition stage implements basis, compensated spacing,
center/start/end alignment, containment, snap-vs-slide selection, loop/fallback,
independent content/items/controls, and visible-item input/accessibility.
The old `Carousel`/`CarouselWith` no-gap pager API remains compatible.

Snapshots may be read before or after Content in the synchronous Container
build. They use the last resolved item/spacing layout and sample the current
programmatic animation; Content resolves new declarations and refreshes the
native drag sample. Geometry changes converge through native invalidation.
Root/control shortcuts require that target's own focus, preserving editors'
arrow keys. Vertical Up/Down, horizontal RTL, and Home/End are deliberate native
adaptations of the web primitive's captured Left/Right handler.

The composed API's current 250ms ease and native drag projection are compatibility
motion, **not** Embla physics. Complete controller/options/events, source physics,
Autoplay/native plugin equivalents, and their Gallery examples remain separate
implementation stages. Archiving those upstream packages does not claim they
are already implemented. No unrelated application/Music source is included.
