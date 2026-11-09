# Third-party notices

Command fuzzy ranking contains a Go port of cmdk 1.1.1's
[`command-score.ts`](https://github.com/dip/cmdk/blob/fb4ea04e9ec211777fbb39c6104e3c5f2ee107d2/cmdk/src/command-score.ts).
The exact source, license and SHA-256 manifest are archived in
`pkg/overlay/testdata/cmdk/`; the scoring fixtures regenerate offline.
The macOS application bundles include its license under `Contents/Resources/licenses`.

Theme/color source notices are kept alongside their archived data at
`pkg/theme/sources/LICENSE.shadcn` and `pkg/theme/sources/LICENSE.tailwind`.
Go module dependencies retain their respective licenses in their module sources.

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

Chart geometry and configuration are independently implemented using MyGo's
public native chart primitives. The pinned shadcn horizontal, mixed and negative
bar examples, base Chart documentation, MIT license and content manifest are
archived in `pkg/display/testdata/shadcn-chart/`. They use the same shadcn commit
and bundled MIT license described above; Recharts is a reference, not a Go runtime
dependency or copied implementation.

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
