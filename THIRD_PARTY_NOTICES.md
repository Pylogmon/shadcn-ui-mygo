# Third-party notices

Command fuzzy ranking contains a Go port of cmdk 1.1.1's
[`command-score.ts`](https://github.com/dip/cmdk/blob/fb4ea04e9ec211777fbb39c6104e3c5f2ee107d2/cmdk/src/command-score.ts).
The exact source, license and SHA-256 manifest are archived in
`pkg/overlay/testdata/cmdk/`; the scoring fixtures regenerate offline.
The macOS application bundles include its license under `Contents/Resources/licenses`.

Theme/color source notices are kept alongside their archived data at
`pkg/theme/sources/LICENSE.shadcn` and `pkg/theme/sources/LICENSE.tailwind`.
Go module dependencies retain their respective licenses in their module sources.

The optional Material Symbols module includes Google's original static SVGs
from `google/material-design-icons` commit
`10ecc14790612ca8fd164e3f4ffe9c16fc50d5b7`, under Apache-2.0.
The exact license, upstream README, complete Git-blob/SHA-256 source index and
deterministic generated-asset manifest are in `pkg/icons/material/sources/`.
On-demand symbol functions retain only referenced SVG pairs; the separately
imported `material/all` package embeds the full static catalog. The Material
Symbols example bundle includes the original license as
`Contents/Resources/licenses/material-symbols.txt`.

The native text shimmer ports the shared utility in the fixed shadcn commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. Its exact CSS, documentation,
MIT license and content manifest are archived in `internal/shimmer/testdata/shadcn/`.
The macOS bundles include that license as `Contents/Resources/licenses/shadcn.txt`.

Input OTP behavior is independently implemented against input-otp 1.4.2,
commit `81ccdb48c010d800b24942aa231909f0c971b1ca`.
The exact reference source and its MIT license (Copyright (c) 2024 Guilherme Rodz)
are archived under `pkg/control/testdata/input-otp/` with a SHA-256 manifest.
The macOS application bundles include this license under
`Contents/Resources/licenses/input-otp.txt`.

Loading indicators and CodeBlock actions include Lucide SVGs, version 0.474.0,
commit `97bbe1d6b2adc6767831e52f6b12b2886b6498d5`, matching the pinned
shadcn app's direct icon dependency. The loading SVG, ISC license (including its
Feather MIT notice) and SHA-256 manifest are in `internal/indicator/`;
exact CodeBlock copy/check SVGs and their manifest are in `pkg/display/assets/`.
The shared Select chevron reuses the existing inline SVG from
`pkg/control/select.go`, with the same Lucide/Feather license notice. Its local
migration source, original literal digest and shared helper digest are recorded
in `internal/indicator/manifest.json`.
Combobox check/clear visuals reuse the already pinned exact Lucide assets,
with original asset paths and digests recorded in that shared manifest.
Toast status SVGs are archived in `pkg/overlay/assets/` with their own manifest.
Attachment file/state/action SVGs and their manifest are in `pkg/conversation/assets/`.
The Trace chevron uses the same pinned Lucide source with a separate manifest there.
The MessageScroller jump arrow uses that same source and a separate manifest.
The macOS bundles include that license as `Contents/Resources/licenses/lucide.txt`.

MessageScroller behavior is independently ported against the shadcn primitive at
commit `8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The complete reference
implementation, geometry/controller/command sources, tests, MIT license and
content manifest are archived in `pkg/conversation/testdata/message-scroller/`.
Its license is the same shadcn MIT license already bundled as
`Contents/Resources/licenses/shadcn.txt`.

Questionnaire behavior is independently ported against that same shadcn commit.
Its complete reference implementation, collection/hooks, tests, MIT license and
content manifest are archived in `pkg/conversation/testdata/questionnaire/`.
The shared shadcn MIT license is included in the macOS application bundles.

Flyout menu motion is compared against the fixed shadcn Base UI wrappers and
Vega utilities, Base UI 1.6.0 / utilities 0.3.1, and tw-animate-css 1.4.0.
The corresponding Radix Menu 2.1.16 / Presence 1.1.5 wrappers are comparison
references; the Base Menubar composition is the native target. Exact reference
files, metadata, MIT licenses and lockfile SHA-512 evidence are archived in
`pkg/desktop/testdata/shadcn-menu/`. Native application bundles include
`base-ui.txt` (Material-UI SAS), `radix-ui.txt` (WorkOS) and
`tw-animate-css.txt` (Wombosvideo) under `Contents/Resources/licenses/`.
No React, Base UI, Radix or CSS runtime is linked into the native application.

Combobox selection, input, chips and popup behavior are independently compared
against the same fixed shadcn Base wrappers, Vega rules and Base UI1.6.0.
Exact registry/docs/examples, selected official published members, metadata,
MIT licenses and Git-blob/SHA-512 evidence are archived under
`pkg/control/testdata/shadcn-combobox/`. Its shadcn and Base UI licenses match
the copies already included in the native application bundles.

Chart geometry and configuration are independently implemented using MyGo's
public native chart primitives. The pinned shadcn horizontal, mixed and negative
bar examples, base Chart implementation and documentation, Vega style source,
MIT license and content manifest are
archived in `pkg/display/testdata/shadcn-chart/`. They use the same shadcn commit
and bundled MIT license described above. No React or Recharts runtime is linked
into the native application.

Cartesian natural curves adapt D3-shape 3.2.0's ISC-licensed natural spline
algorithm to public MyGo paths; the full notice is retained in
`pkg/display/chart_curves.go`. Exact D3-shape/D3-path distribution members,
licenses and an oracle executing the official factories are archived in
`pkg/display/testdata/shadcn-chart/cartesian/`. Native bundles include the ISC
notices as `licenses/d3-shape.txt` and `licenses/d3-path.txt`. JavaScript
packages are reference/oracle inputs and are not linked as a runtime.

Chart data animation defaults and geometric rules are compared against Recharts
3.8.0, commit `c04a79039ebc4d5057de163298d2e9917a5e5a66`, locked by that
shadcn version. Fixed source, selected official npm distribution members,
metadata, MIT license and integrity manifest are archived in
`pkg/display/testdata/recharts-animation/`. The npm archive's SHA-512 matches the
lockfile; the eleven TypeScript reference files also match the official Git tree
blobs. Native apps include its license as
`Contents/Resources/licenses/recharts.txt`.

Chart shadow parameters are additionally verified against the official
Tailwind CSS 4.3.0 package locked by the pinned shadcn reference. Its exact
theme CSS, MIT license and npm integrity evidence are archived with the chart
references. The macOS application bundles include the matching Tailwind MIT
license as `Contents/Resources/licenses/tailwind.txt`.

Toast behavior and styles are independently ported against Sonner 2.0.7,
commit `3ba7aa17ab7e8101b9cf4893936f873b0d4769b3`, the version locked by
the fixed shadcn reference. The exact source, MIT license (Copyright (c) 2023
Emil Kowalski) and SHA-256 manifest are archived in `pkg/overlay/testdata/sonner/`.
The Toast close SVG preserves Sonner's original geometry. The macOS bundles
include its license as `Contents/Resources/licenses/sonner.txt`.

## Embla Carousel

The composed Carousel's layout, snap geometry, controller, motion, plugin lifecycle and Autoplay primitives are native Go translations
of Embla Carousel 8.5.2, commit `5ee3c4ed0251dc685198d0ac21dc3e754b865476`,
under the MIT license, Copyright (c) David Jerleke.
The exact source, license, four lockfile-verified npm packages and source-function
geometry/controller/physics/plugin/Autoplay fixtures are archived in `pkg/display/testdata/shadcn-carousel/`.
The macOS application bundles include the full license as
`Contents/Resources/licenses/embla.txt`. No JavaScript package is a runtime
dependency of the Go kit.

## cmdk

MIT License

Copyright (c) 2022 Paco Coursey

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
