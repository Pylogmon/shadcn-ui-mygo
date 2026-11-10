# Menubar compositions

Open the Gallery navigation page and use **Menubar configuration**. The default
is the docs' complete File/Edit/View/Profiles bar. Each chosen leaf updates
**Last command** locally. File, document and account labels do not perform
filesystem, editing, sign-out or deletion requests; shortcut captions are
display metadata.

The module maps all six docs examples and twelve Create sections at shadcn/ui
revision `8ab27715521cbd4f87ae6886cd989cbb678f6de7`. Original paths, Git blobs,
byte lengths and SHA-256 values, including the Menubar MDX guide, are recorded in
[`source-map.json`](source-map.json). Three identical docs/Create compositions
(Checkbox, Radio and Submenu) share configurations. The two Icons variants
retain the docs' semantic icons and Create's generic placeholders separately.

| Configuration | Source differences retained |
| --- | --- |
| File, Edit, View and Profiles | Complete docs sets; Share/Find submenus; disabled Incognito/Force Reload; Bookmarks off, Full URLs on, Benoit selected; inset actions |
| Basic | File/Edit commands, shortcuts, separators and disabled Incognito |
| With Submenu | Share and three-item Find; File/Edit sibling switching |
| Six sides | Six independent bars in a centered wrapping row; physical and logical Flyout placements; Incognito enabled as in this source |
| With Checkboxes | View/Format; Full URLs and Strikethrough initially checked; Code and Superscript off |
| With Radio | Separate Profiles and Theme groups; Benoit and System defaults |
| With Icons (docs/Create) | File/Folder/Save plus semantic Settings/Help/Trash or generic placeholders; destructive Delete |
| With Shortcuts | Source File/Edit captions including Cut/Copy/Paste shortcuts |
| Format | Bold/Italic/Underline icons; Strikethrough on, Code off; Ruler off, Grid on; inset Zoom actions |
| Insert | Media submenu with Image/Video/Audio; Link/Table; Find & Replace/Spell Check |
| Destructive | File/Account; Delete File, Sign out and Delete, with source icons and captions |
| In Dialog | Real Menubar Example Dialog; Copy/Cut/Paste; More Options submenu and destructive Delete; Edit commands |
| With Inset | Actions/Appearance/Theme captions; independently inset labels, actions, checkbox/radio and submenu; Bookmarks on, URLs off, System selected |
| Translated File, Edit, View and Profiles | Verbatim English/Arabic/Hebrew labels, Arabic default, logical RTL alignment and nested content |

**Menu** uses `desktop.Menubar` and one independent Flyout. Sibling arrow
navigation reuses that window. Checkbox/radio selections keep the Flyout open
with refreshed checked state; ordinary commands close it and return focus.
Inset padding follows the logical start. Six-side placement resolves logical
sides from the direction selector; the platform may flip the final placement
to fit the work area.

`overlay.MenubarItem.Width` gives each top-level styled menu its own panel
width. A positive finite value overrides `desktop.MenubarProps.Width`;
nonpositive values inherit that fallback, whose default remains 144 DIPs.
Styled desktop menus reject non-finite widths on any candidate. The source's
explicit widths now map directly:

| Source menu | Source and Gallery width |
| --- | --- |
| Docs Demo / RTL: View | 176 DIPs |
| Docs / Create Checkboxes: View | 256 DIPs |
| Create Insert: Tools | 176 DIPs |
| Create Destructive: File | 160 DIPs |
| Create With Inset: View | 176 DIPs |

Checkboxes' Format has no explicit source width and retains the 224-DIP Gallery
fallback. Other unspecified widths also retain 224, and side previews retain 192.
The fixed Vega source uses a 144-DIP minimum with natural CSS content sizing;
these Gallery fallbacks are native fixed-width adaptations, not source pixel
measurements. Intrinsic content sizing and the complete font/platform visual
matrix remain acceptance work.

The child reads the current sibling's width before the parent's next reanchor
frame. The standard native adapter implements `desktop.SizeSurface`, so a
width-only change resizes the same Flyout and retains its editor/focus. A custom
backend without this optional capability retains Host's replacement policy.
Legacy `InWindowMenubar` forwards finite item widths as its existing menu minimum;
non-finite legacy values fall back safely. Native popup width stays under OS
control and ignores this visual setting.

**Native** uses `desktop.NativeMenubar`, a group of system popup triggers.
It is distinct from an application's global operating-system menu bar.
The system owns menu styling, popup positioning, tracking and dismissal. Custom
Flyout inset/side positioning and sibling switching during OS tracking are not
claimed for Native. Both routes use the same local checkbox/radio drafts,
retained when changing configuration, route or page.

Icons use on-demand Material Symbols SVG. Native template images are cached
from the same SVG; no icon font or runtime network loader is introduced.
Create's generic CircleDashed placeholder maps to the generic ProgressActivity
symbol. The model stores drafts, public Menubar states/Handles and reusable
icon resources, and does not retain a Context, Element or frame callback.

`TestMenubarCompositionsSourceWorkflow` is one public-input flow through the
Native commands and the same model's Flyout sibling navigation, focus,
checkbox/radio retention, inset groups, six placement requests, RTL, actual
Dialog nesting and page cleanup. `TestMenubarCompositionsMaterialSourceDifferences`
checks source-specific defaults and icon/inset/disabled distinctions. Set
`GALLERY_MENUBAR_COMPOSITION_SNAPSHOT_DIR` to export representative PNGs.
The injected Backend sizes its test surface to measured public menu bounds to
stand in for the platform's FitContent sizing. Tester images show the parent
and Flyout content, not an OS-owned menu's pixels. These checks do not replace
real-platform tracking, assistive-technology or full visual acceptance.
