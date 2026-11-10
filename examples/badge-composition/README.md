# Badge source compositions

Open Gallery's **Badge compositions** card and choose **Badge configuration**.
The default is the docs Demo: Badge, Secondary, Destructive and Outline.
The module preserves all seven docs examples and seven Create sections as
14 separate configurations at shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`.

| Configuration | Source content retained |
| --- | --- |
| Demo | Four centered source labels and variants |
| Variants (docs) | Five variants; the source omits the Link visual variant |
| Verified and Bookmark | Secondary with a leading verification symbol, Outline with a trailing bookmark |
| Deleting and Generating | Destructive leading spinner, Secondary trailing spinner |
| Open Link | Actual native Link, original `#link` href, trailing outward arrow |
| Custom Colors (docs) | Five pastel Blue, Green, Sky, Purple and Red badges |
| Translated Badges | Verbatim English, Arabic and Hebrew labels; Arabic default, logical icon positions and RTL order |
| Create: Six Variants | All six variants, including a static Link-styled span |
| Create: Icon Left | Six variants with the source verification symbol at logical start |
| Create: Icon Right | Six variants with a right arrow at logical end |
| Create: With Spinner | Six variants with leading spinners |
| Create: asChild Links | Four real links: Default, Secondary, Destructive and Ghost, each with the original `#` href and outward arrow |
| Create: Long Text | The complete original sentence in a Secondary badge |
| Create: Custom Colors | Four solid colors followed by the five pastel colors |

Static source spans remain non-interactive, including the visual Link variant.
Only source anchors use `display.BadgeWith`'s native `URL` Link and activation
callback. Their original hash-only hrefs are consumed by a local public MyGo
Router at `/badge-examples`, and activation records **Local badge action**.
The native Router retains path/query and drops fragments: this does not claim
HTML anchor scrolling, fragment history or an external browser request.

The body uses the existing Vega Badge API. The global Flyout/Native setting
changes the configuration/language selectors; it does not invent an OS badge
widget. Icons are on-demand Material Symbols equivalents: Verified for
BadgeCheck, Bookmark, ArrowForward and ArrowOutward. Spinner reuses the existing
12-DIP decorative SVG in the badge's visual slot, exposes the whole badge's
busy state, animates on the MyGo clock and respects reduced motion. No external
task is simulated. No font, icon font, bitmap or network loader is added.

Custom colors use the existing fixed Tailwind palettes. Solid colors retain
600 backgrounds and 50 foregrounds in both appearances. Pastel colors use
50/700 in light mode and 950/300 in dark mode. The source Badge wrapper is
`shrink-0` and `whitespace-nowrap`; despite the Long Text title, it does not
enable text wrapping. A bounded native horizontal viewport keeps the complete
sentence reachable in narrow parents.

The state holds application values, a public Router/ScrollState and reusable
resources only. Current views build every element and callback. The source
paths, original Git blobs, byte lengths, SHA-256 values and exact recipe mapping
are in [`source-map.json`](source-map.json); inventory counts are not a claim of
14 independent visual or real-platform acceptance tests.

`TestBadgeCompositionsSourceWorkflow` uses public MyGo Tester input for all
configuration choices, actual pointer/Enter-only links and local Router,
static semantics, loading motion, real palette pixels, long-text scrolling,
translated RTL and a final styled independent selector. Set
`GALLERY_BADGE_COMPOSITION_SNAPSHOT_DIR` to export four representative PNGs.
These are software Tester images, not real OS-menu pixels, assistive-technology
results or a complete font/platform/visual matrix.
