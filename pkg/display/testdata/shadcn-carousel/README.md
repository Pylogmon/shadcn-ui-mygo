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

The nil-Controller/nil-Options composed API retains its 250ms ease and native
drag projection as compatibility motion. The Controller/Options profile uses
the C2 source solver described below. Autoplay/native plugin equivalents and
their Gallery examples remain separate implementation stages. Archiving
those packages does not claim they are already implemented. No unrelated
application/Music source is included.

## C1 controller/options/events reference

`controller/` contains an additional independent executable oracle and its
frozen requirements. Its generator executes the actual verified CJS `Engine`,
`EmblaCarousel`, and `EventHandler` with explicit fake observer/RAF deliveries.
It records command event snapshots, listener identity/mutation, grouping,
partial reinitialization, disabled watchers, inactive behavior, visibility,
native-selector equivalents, and large-negative source counter behavior.
Regenerate it with:

```sh
node pkg/display/testdata/shadcn-carousel/controller/generate-controller-oracle.cjs \
  pkg/display/testdata/shadcn-carousel/npm/embla-carousel/cjs/embla-carousel.cjs.js \
  "$carousel_audit_dir/controller-oracle.json"
```

The full-source traces include physics values for later stages. C1 checks source
controller/event semantics and geometry; it does not claim those physics curves.
In particular, a fresh jump emits `scroll` while the old selected index is still
observable, then `select`; it does not start a RAF or automatically emit `settle`.
An already scheduled animation can still settle after an intervening jump.

The native controller belongs to application state. Commands queue until an
owner build, retaining FIFO configuration/selection event states. `Emit` and
`Destroy` run on the UI thread; listener callbacks use explicit application-owned
identity and may not retain frame objects. Destroy terminates the engine and
clears subscriptions, while explicit later `On`/`Emit` remains possible like the
source. A same-window owner transfer advances a generation so obsolete Local
cleanup cannot destroy its replacement. Weak public Services wake the owner;
public `ui.Local`/`io.Closer` handles removal without retaining Context/Element.

Source `ScrollSnapList` is normalized; native `SnapOffsets`/the original composed
`Snapshot.Snaps` are DIPs. Progress is not clamped to 0..1. Disabled resize/slide
watchers retain the committed engine geometry while native layout/input remain
live; explicit ReInit resolves the current declared track/items. Inactive is
separate from Disabled. Watch hooks are frame-only veto callbacks. Handle
selection preserves declared physical positions/order, including gaps in a
non-contiguous subset, and rejects foreign/unresolved nodes explicitly.

Options are partial patches with pointer fields for explicit false/zero values.
Structural validation/configuration errors are atomic and available before a
build. The source `Duration`, `DragThreshold`, `DragFree`, and `SkipSnaps`
fields were reserved during C1 and are active in C2 as described below. Callback alignment uses an
explicitly application-owned pure Aligner. Breakpoint predicates use native
window dimensions/preferences and stable ordered rule keys. Native focus
observation preserves the existing offscreen input/AX exclusion; the source's
recent-Tab default is distinct from pointer/programmatic focus notifications.


## C2 physics/native input reference

`physics/oracle.cjs` executes the actual pinned CJS ScrollBody, Bounds, Looper,
Target, DragTracker, and Animations primitives. Its 43 scenarios produce 48 Go
checks. `physics/receipt.json` preserves the original **pure-core-only** receipt;
it is historical evidence rather than a claim of native integration. The three
kernel files retain those exact hashes. The repository test changes only its
fixture path. MIT attribution remains in THIRD_PARTY_NOTICES/Embla's license.

`physics/generate-engine-oracle.cjs` reuses the archived C1 rect/DOM fixtures and
executes the actual upstream Engine and DragHandler together. Its 12 native
trajectories contain 985 action snapshots, including every scroll/select/settle
and pointer event's intermediate vectors: default/40/0 Duration, retarget,
in-flight jump, loop wrap, hidden reset, snap/free/skip drags, touch velocity,
and loop dragging. Native tests compare these independent values at the source
60Hz timestamps and sample actual color pixels/visible AX groups. Additional
native checks cover RTL/vertical, single/empty snap sets, cross-viewport release,
cancellation, media/partial copies, reduced motion, and preserved child taps.
Regenerate both independent outputs with:

```sh
node pkg/display/testdata/shadcn-carousel/physics/oracle.cjs \
  pkg/display/testdata/shadcn-carousel/npm/embla-carousel/cjs/embla-carousel.cjs.js \
  "$carousel_audit_dir/physics-oracle.json"
node pkg/display/testdata/shadcn-carousel/physics/generate-engine-oracle.cjs \
  pkg/display/testdata/shadcn-carousel/npm/embla-carousel/cjs/embla-carousel.cjs.js \
  "$carousel_audit_dir/engine-oracle.json"
```

The controller profile ports source attraction/friction, bounds resistance,
periodic target selection, drag release force, and the 60Hz accumulator with two
initial seeks. Source `Duration` defaults to 25 and is an attraction parameter,
not milliseconds; zero is immediate for programmatic scroll. A drag release
uses source speed 25 (snap) or 43 (free), independently of this option. Source
`DragThreshold` defaults to 10 DIPs and records click suppression; the separate
navigation force threshold is 20% of view clamped to 50..225 DIPs. Source tracker
uses raw point timestamps over 170ms, rather than the SDK's drag velocity.
`DragFree` and `SkipSnaps` preserve their distinct source target/force behavior.

Only public SDK input is used. The viewport declares capture via `Dragged`
without automatic translation; a non-consuming `HandleInput` receives that
viewport's moves/releases even outside its bounds. The source body supplies all
track translation. `ObserveInput` supplements descendant interaction events;
coalesced raw delivery supplies actual device/pressure values omitted by the
SDK observer. The handlers retain only engine/value records, never a Context or
Element. Frame-only veto hooks continue running in the current build.

There is an explicit remaining SDK capability gap: an interactive descendant
(such as a Button) wins its own pointer. The public SDK exposes neither an
ancestor move/capture stream nor ancestor click cancellation. Ordinary child
taps remain usable, but starting a carousel drag on an arbitrary interactive
child and suppressing that child's click after the source threshold are **not
implemented**. `Engine.ClickSuppressed` reports the captured viewport gesture's
source policy; it is not a claim to cancel another native widget's action.
The SDK's current latest public handler has the same limitation. No private
flags, module-cache patch, fork, or fake ancestor capture is used. Multi-contact
touch metadata is likewise unavailable through the SDK's single-pointer events.

Native adaptation policies: a canceled/lost-window press returns to its selected
snap without committing a fling; touch movement dominated by the cross axis
cancels that capture. Reduced motion (or the preexisting negative Props Duration
override) resolves programmatic/release animation immediately and schedules no
continuing animation frames. Hidden windows pause the scheduler and reset its
accumulator; resuming performs source warm-up rather than replaying hidden wall
time. These policies do not change the source numeric options or the old
nil-Controller/nil-Options compatibility path. Autoplay remains the next stage.
