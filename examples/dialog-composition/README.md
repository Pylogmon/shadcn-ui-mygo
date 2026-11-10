# Dialog source compositions

Gallery's Overlays page exposes **Dialog configuration** with all six docs
previews, the distinct inline guide Usage and five Create sections: twelve
independent recipes at shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`.
[`source-map.json`](source-map.json) records the eight original example/guide
files and two dialog/style references with byte lengths, SHA-256 and Git blobs.

| Recipe | Original material retained |
| --- | --- |
| Edit profile (docs) | Open Dialog, 384-DIP maximum, original Name/Username defaults, description, Cancel and Save changes |
| Basic dialog (guide) | Open, “Are you absolutely sure?” and the original destructive-action description, without inventing confirmation actions |
| Custom close / share link | Share link, real readonly installation URL and a start-aligned native Close action |
| No close button (docs) | No corner or footer close; ordinary Escape/backdrop dismissal remains enabled |
| Sticky footer (docs) | Original distinct title/description, ten full paragraphs, 50%-of-window body cap and fixed header/footer |
| Scrollable content (docs) | Original shorter description, same ten paragraphs and 50% cap, without a footer |
| Translated profile | Original English/Arabic/Hebrew strings, Arabic default, real profile editors and logical-end close |
| Create: With Form | Edit Profile, 448-DIP default maximum, original extended description and independent profile draft |
| Create: Scrollable Content | Original text, ten paragraphs and 70%-of-window body cap |
| Create: With Sticky Footer | Original “Scrollable Content” title/description, 70% cap and Close footer |
| Create: No Close Button | No corner close, with the Create-only footer Close action |
| Create: Chat Settings | Complete four-section General/Notifications/Personalization/Security settings, native controls and responsive section selection |

Every popup uses the real public `overlay.Dialog` and its native modal scope,
focus trap, Escape/backdrop policies, bounded layout and retained exit. Public
`HideHeader`/`NoScroll` allow current source header/body composition; the title
uses Vega's inherited 14-DIP font instead of the convenience header's 18-DIP
default. Native font metrics can make a 14-DIP line box taller than fourteen
pixels. The actual returned current modal panel receives a real Ghost/IconSmall
button with the existing on-demand Material Close SVG. Its 32-DIP target sits
at the logical end, reserving the title's end gutter in RTL. No frame Element
or Context is retained. Narrow footers preserve reversed source order and
stretch their actual native buttons; wide share footers stay start-aligned.

Scrollable variants retain ten complete original paragraphs. Their real native
scroll viewport hides its scrollbar as the source requests and leaves the
header/footer stationary. Native window constraints can further reduce the
declared 50%/70% body cap. The standard 448-DIP Create width, explicit 384-DIP
docs profile width and 32-DIP total window margin follow the pinned Vega style.
The Chat source's minimum width is constrained by the actual native window
when necessary; its settings body retains a 550-DIP content minimum inside a
bounded native scroll viewport.

Chat Settings includes all eighteen spoken-language and eight voice choices
with the original defaults. Desktop windows use the real Tabs slots; below
768 DIPs a real OS NativeSelect controls the same application-owned section.
General retains Theme, Accent Color, Spoken Language and Voice. Notifications
keeps the checked disabled Responses checkbox, independent editable Tasks
checkboxes and real Manage tasks link. Personalization keeps the actual
Nickname InputGroup, rich independent tooltip with Kbd N, 120-DIP-minimum
multiline editor and whole-caption customization switch. Security retains its
multi-factor switch, complete descriptions, Log Out and Log Out All actions.
Controls do not introduce duplicate visible switch captions. Config/language
selectors, General choice menus and the rich help tooltip use Gallery's
Flyout/Native setting. The rich Native tooltip is a platform-themed independent
surface, not a fabricated operating-system tooltip widget. Styled choice
menus use the existing ordinary desktop selection facade rather than claiming
the full current-window Select's item-aligned popup geometry.

Source Save buttons have no application transport handler. They record only a
local profile snapshot, keep the draft and leave the dialog open. Cancel and
actual close buttons close normally. No implicit Enter submit, server save,
theme application, email notification, customization service, authentication
change or logout is performed. Source settings values are local demonstration
state. Original descriptive copy remains sample text. Manage tasks uses a
real native hash link in a local public Router; its path/query behavior drops
the fragment and does not implement HTML anchor scrolling/history or open an
OS browser.

`TestDialogCompositionsSourceWorkflow` is one public-input flow covering all
twelve configurations, actual profile editing/read-only input, no implicit
Enter action, save/cancel/focus return, both no-close contracts, Escape and
backdrop, full paragraph scrolling with fixed footer, all four settings panels,
real menu selections, switches/checkboxes, native rich tooltip, responsive
section selection, narrow Hebrew, and independent drafts across source/style
and page changes. Set `GALLERY_DIALOG_COMPOSITION_SNAPSHOT_DIR` to export four
representative software PNGs. The injected menu/tooltip backend and software
images do not establish real OS tracking, assistive technology, IME, display
scale, multi-display, Windows/Linux runtime or the full font/visual matrix.

Application state holds only strings, booleans, maps, counters, submitted values
and public Handles. Scope, Context, Element and frame callbacks never survive a
view. No SDK, font, icon font, image asset, dependency or module-cache edit is
introduced by this source module.
