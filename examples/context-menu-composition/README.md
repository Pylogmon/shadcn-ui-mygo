# ContextMenu compositions

In the Gallery overlays page, choose **Context menu configuration**, then
secondary-click the bordered target or focus it and press the Context Menu
key / Shift+F10. The target's accessible name is **Context menu target**;
side examples append `: inline-start`, `: left`, etc. **Open Dialog** opens
the actual parent Dialog for the Create dialog example.

This module maps all eleven docs examples and all eleven Create sections at
shadcn/ui revision `8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The thirteen
original files, including the guide, are recorded with pinned URLs, Git blobs,
byte lengths and SHA-256 values in [source-map.json](source-map.json).

| Configuration | Preserved source behavior |
| --- | --- |
| Basic / Create: Basic | Back, disabled Forward and Reload |
| Browser menu | Navigation shortcuts, More Tools submenu with destructive Delete, Bookmarks / Full URLs and People selection |
| Icons | Copy, Cut, Paste and destructive Delete |
| Sides | Four docs targets; all six Create physical/logical targets; Forward enabled |
| Shortcuts | Navigation hints, Save and Save As... |
| Submenu | Copy/Cut hints and all More Tools groups/actions |
| Groups | File/Edit captions, separators, eight actions and destructive Delete |
| Checkboxes | Bookmarks Bar and Developer Tools initially checked; Full URLs unchecked |
| Radio | Independent People=Pedro Duarte and Theme=Light groups |
| Destructive | Docs Edit/Share/Delete; Create additionally preserves Archive |
| Translated menu | Exact English, Arabic and Hebrew strings; Navigation icon submenu and disabled Forward; Arabic default |
| Create: In Dialog | Original heading/description, clipboard actions, More Options submenu and destructive Delete |
| Create: With Inset | Actions/Appearance/Theme captions; logical inset items, checks and radios; Bookmarks=true, Full URLs=false, Theme=System; 176-DIP panel |

The docs targets use their dashed 16:9 geometry, radius 12 and maximum width
320 DIPs. Create uses solid 4:1 targets with radius 8. Side target grids adapt
to available space. Desktop fine-pointer text is shown; source long-press
strings are retained in target accessibility descriptions and the RTL model,
without adding a touch long-press recognizer.

**Flyout** uses `desktop.ContextMenu`, with the kit's unified shadcn theme in
an independent native child window. Nested panels contribute to the child
window's measured width, and may extend beyond its parent window. Requested
logical sides resolve using the current reading direction; display collision
handling may change the final physical placement. Checkbox/radio rows keep
the styled menu open by default, as the source component does; ordinary
actions close it. The current public Menu API uses one width for every panel:
the docs Demo/RTL source requests 176-DIP subpanels under a 192-DIP root,
which this module renders as 192-DIP subpanels.

**Native** uses `desktop.NativeContextMenu` and the operating system's menu
tracking, painting and placement. Source item labels, selection/disabled
state, groups, submenus and shortcut hints are retained. Equivalent Material
Symbols outline SVGs become shared native template images. Source CSS inset,
panel widths and side requests do not restyle or reposition the system menu.
No icon font or network asset is required.

Every action is local: it updates a visible status or the configuration's
own checkbox/radio values. Names such as Copy, Delete and Save Page... do not
operate on the clipboard, browser, network or user files. Each configuration
has its own draft; values survive configuration/page changes. Models retain
application values and public menu/focus/image handles, never a frame Context
or Element. Child callbacks mutate application data only.

`TestContextMenuCompositionsSourceWorkflow` exercises all 22 real configuration
choices, secondary-click and Context key, disabled actions, submenu paths,
checkbox/radio values, real parent Dialog, translated RTL, all six styled side
requests, both routes and page-return state. Set
`GALLERY_CONTEXT_MENU_COMPOSITION_SNAPSHOT_DIR` to export representative parent
and child PNGs. The Native image shows the actual parent Dialog: a headless
Tester image does not include an operating system menu. The public injected
Backend/Tester flow is focused interaction evidence. Its submenu canvas is
explicitly enlarged because it does not run the native FitContent sizing loop;
it does not claim real platform tracking-loop or visual acceptance coverage.
