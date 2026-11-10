# Popover and HoverCard compositions

Run `go run ./cmd/gallery -page overlays`, then choose **Popover configuration**
or **Hover card configuration**. The two models are independent; edits remain
local and survive configuration/page changes.

The module maps ten original files at shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7` to twelve material configurations.
Original paths, Git blobs, lengths and SHA-256 values are in
[`source-map.json`](source-map.json). Duplicate docs/Create Basic, Form and
Alignments content intentionally shares one configuration.

| Popover | Source behavior |
| --- | --- |
| Four dimensions | Width, Max. width, Height, Max. height; actual child editors, 320-DIP window, centered below trigger |
| Basic header | Visible Dimensions heading and description, start alignment, 288 DIPs |
| Dimensions form | Separate two-field Width/Height drafts, half-width labels, 256 DIPs |
| Three alignments | Start/center/end, 160 DIPs and 24-DIP trigger gaps |
| Six sides | Two source trigger rows, including inline start/end |
| Popover in Dialog | A real parent Dialog and separate editable popup Context |
| Translated dimensions | English/Arabic/Hebrew headings, physical/logical sides and source two-row layout |

| HoverCard | Source behavior |
| --- | --- |
| Next.js profile | `@nextjs`, framework description, Joined December 2021; 10-ms pointer opening and 100-ms closing |
| Four physical sides | Docs left/top/bottom/right; 100-ms opening/closing and 4-DIP body gap |
| Six physical and logical sides | Create six sides; 100-ms opening/closing and Vega 8-DIP body gap |
| Hover card in Dialog | Real parent Dialog, independent preview, 100-ms delays |
| Translated product | Headphone name and price in English/Arabic/Hebrew, two trigger rows, 10/100-ms delays |

Logical placement resolves from the current reading direction. Popovers use
the source 4-DIP gap. Native window collision handling may flip the final side;
the configuration is the requested placement, not a claim about every display.

Styled content uses the kit's theme. **Native** chooses `NativePopover`'s
platform container/theme and real `ui.TextInput` controls. Native HoverCard uses
a platform-themed inactive plain Flyout; it is not a system HoverCard widget.
Keyboard focus keeps the existing native immediate preview path, separately
from the source pointer delays.

Each popup builds its heading, description and controls using the independent
child Context. Its own native Dialog is named by the visible heading and
described by the visible description; the Host wrapper uses RoleNone to avoid
duplicate dialog semantics. Field captions name the actual child editors.
Models retain only drafts, booleans, keys and public Handles. They never retain
a parent Context, Element or per-frame handler. No business request is sent.

The Next.js original imports Avatar helpers but does not render an avatar, so
this module introduces no image, icon-font or other asset download.

`TestPopupCompositionsIndependentSourceWorkflow` is one public native UI flow:
separate parent/child Contexts, typed edits/reopen, current native platform
font, title/description associations, all placement requests, dialog nesting,
pointer delays, translated previews, style switching and page-return drafts.
Set `GALLERY_POPUP_COMPOSITION_SNAPSHOT_DIR` to export its representative child
windows. The injected Backend's native-close callback tests dismissal, not a
real operating-system tracking loop. Tester and compilation evidence do not
replace real platform, IME, screen-reader or complete visual acceptance.
