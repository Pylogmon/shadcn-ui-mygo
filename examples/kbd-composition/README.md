# Kbd source compositions

Gallery's Text & layout page can expose **Kbd configuration** with all six docs
previews and nine Create sections from shadcn/ui
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. Two distinct guide/API snippets are
also retained, for **17 independent recipes**. [`source-map.json`](source-map.json)
records ten pinned originals, including the component and Vega style references.

| Source material | Retained content |
| --- | --- |
| Docs demo / RTL | Separately centered modifier-symbol and Ctrl + B rows, 16-DIP row gap and verbatim symbols; RTL changes direction without inventing translated key names |
| Docs Group | Use, grouped Ctrl + B / Ctrl + K keycaps and the complete command-palette sentence |
| Docs Button | Real Outline Accept button and inert return keycap, logical tail inset and physical two-DIP translation |
| Docs Input Group | Actual plain text editor, Search placeholder/icon and two trailing command-key caps |
| Docs Tooltip | Real joined Save / Print buttons and separate rich tooltip children containing S or Ctrl / P |
| Guide / API | Bare Ctrl and the distinct Ctrl / B group without a plus separator |
| Create Basic / Modifier Keys / Group / Arrow Keys | Complete literal key strings and original four/eight-DIP group spacing |
| Create Icons / Icons and Text | Static existing Material SVG equivalents, useful native names, Left and Voice Enabled text |
| Create InputGroup | Empty native editor with leading Space keycap; no docs-only placeholder |
| Create Tooltip | Real 32-DIP Outline save icon button, eight-DIP rich-content gap and source end-inset adjustment |
| Create samp | Native monospace File sample within the normal keycap box |

Every cap uses public `display.Kbd` / `KbdWith` / `KbdGroup`. Keycaps remain
inert native text and install no keyboard binding. The literal source symbols
are not rewritten using platform modifier aliases. The public `KbdShortcut`
formatter remains available separately. Standard keycaps retain 20-DIP height,
20-DIP minimum width, 12-DIP text and four-DIP visual/group gaps; fonts use the
current body family. Source wrappers explicitly center their content because
the native keycap API also permits start-aligned composition.

The Accept button records only a local activation count. Save/Print tooltip
buttons have no transport or print handler, matching their original role.
Input values are independent application strings; native editing, selection,
paste and undo stay with the real editor. Clicking an inert keycap addon
focuses that same editor. Displayed Cmd-K / Space notation does not add an
application command or intercept text input.

Configuration/language selectors follow Gallery's Flyout/Native setting.
Tooltips use genuine independent pointer-transparent, nonfocusable child
surfaces. Styled tooltips apply the source inverse keycap tint; Native uses the
child's platform theme, without claiming an OS tooltip/keycap widget. Their
window-owned content providers capture only copied text/keys and presentation
values, and always build from the current child Context. Models retain only
strings, maps and counters, never Context, Element, Scope or frame callbacks.
No font, icon font, image asset or dependency is added.

`TestKbdCompositionsSourceWorkflow` is one concentrated public native-input
workflow through all recipes, actual button activation/Tab order, real input
paste/undo/addon focus, inert shortcut notation, both tooltip routes, inverse
pixels, complete Print grouping, icon tooltip, Hebrew/narrow layout and
independent drafts across recipe/style/page changes. Set
`GALLERY_KBD_COMPOSITION_SNAPSHOT_DIR` to export four representative PNGs.
The tooltip PNG expands an injected software canvas to inspect actual child
content because the injected Backend does not run OS FitContent. It is not an
OS-fitter/window/IME/AT/font/platform acceptance claim.
