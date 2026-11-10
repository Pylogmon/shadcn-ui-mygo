# Gallery font selection

Run `go run ./cmd/gallery`. The Font and Heading controls expose all 26
families from the fixed Create picker, using their readable names. The default
Vega pairing is Inter with inherited headings. A family is explicitly registered
through `pkg/fonts` before a changed preset becomes active; repeated builds use
the font package's idempotent registration.

The preview's body font applies to native text and editors. Heading applies to
the source Card, Dialog/AlertDialog/Drawer/Sheet, Empty and Questionnaire titles,
and the Gallery's own headings. Styled Flyouts receive both roles, including a
font-only update to an existing child. Native menus/popovers keep their platform
font and appearance. The dark configuration panel keeps its own interface font.

Choose **Inherit** to let headings follow subsequent body changes. Choosing the
same named catalogue family for body and heading is saved as `inherit`, following
the fixed source. Native `system`, `serif` and `mono` choices stay explicit even
when both roles match. **System** is an explicit system font choice for either role.
Existing Version 1 `system`, `serif`, `mono` and heading `same` presets still
import; `same` normalizes to `inherit`. Unknown IDs fail before configuration or
history changes. Export/import, undo/redo, reset and shuffle use the same
normalized state and keep application drafts independent of the design.

All assets are embedded in the Gallery because it imports `pkg/fonts/all` for
arbitrary picker selection. Registration loads only the current body/heading;
it does not eagerly register every family. Other applications can call an
individual `fonts.Inter()`-style function for linker selection instead. Font
assets and original notices are owned by the font package.

The concentrated native flow edits a real InputGroup textarea and retains its
selection across font/history changes, compares actual Gallery heading pixels
with independent family references, checks JSON/legacy normalization, updates
an existing styled popup and verifies the platform route. It also uses keyboard
navigation to reveal and choose the final named family in a bounded styled
picker. Its Tester adapter uses the declared menu height cap because Tester
has no operating-system FitContent negotiation; it is not physical window,
assistive-technology or cross-platform runtime acceptance.

Fixed source commit and mapping: [source-map.json](source-map.json). The original
Create and component files are archived under
[shadcn-font-roles](../../pkg/theme/testdata/shadcn-font-roles/README.md).
