# Textarea source compositions

The Gallery Forms & controls page exposes **Textarea configuration**. Its eleven
configurations preserve all six docs examples and five Create sections at
shadcn/ui revision `8ab27715521cbd4f87ae6886cd989cbb678f6de7`.
[`source-map.json`](source-map.json) records the eight original files, fixed
Git blobs, byte lengths, SHA-256 values and native adaptations.

| Configuration | Source material retained |
| --- | --- |
| Basic (docs) | Bare multiline editor, original placeholder and 320-DIP docs preview maximum |
| Field with preceding description | Message label, “Enter your message below.” before the editor, actual described-by and label-focus association |
| Disabled field (docs) | Disabled Field and actual editor, with the original Message label |
| Invalid field (docs) | Editable invalid editor and “Please enter a valid message.” following it |
| Send message | Native multiline draft and full-width native button separated by eight DIPs |
| Translated feedback | Verbatim English/Arabic/Hebrew label, placeholder and description, four rows, Arabic default and 320-DIP maximum |
| Create: Basic | Bare editor filling its available Create preview width |
| Create: Invalid editor | Bare invalid editor; no invented Field or description |
| Create: Label with six rows | Message label and the declared six-line native viewport |
| Create: Six rows and following description | Six rows, then the original “Type your message and press enter to send.” copy |
| Create: Disabled editor | Only the editor is disabled; the Field label retains its ordinary appearance |

All fields use the existing public `control.Textarea` and current-build
`FieldWith`/`FieldScope` API. The actual native multiline editor owns text,
selection, paste, undo and accessibility. Application state holds independent
strings, selection integers, change counters and submitted snapshots. It never
retains a Context, Element, FieldScope or frame callback. The configuration and
language menus follow the Gallery Flyout/Native choice.

The source Send message button has no network handler. The Gallery records a
local snapshot and leaves the draft unchanged; the status is
`Local textarea action: Send message`. The Create description is instructional
copy without a source key handler: Enter continues to insert a native newline.
No transport, account change or implicit Enter submission is claimed.

Unspecified rows use the existing native auto-growing editor with the fixed
64-DIP source minimum. Explicit rows use the public native line-count option;
line boxes and resulting pixel heights follow platform font metrics. The docs
MDX `max-w-xs` is 320 DIPs; Create's unspecified width fills the available card.
No extra resize handle, icon font, font asset or dependency is introduced.

The pinned MyGo accessibility builder returns from disabled nodes before
filling native-editor Multiline, placeholder and value metadata. The public
`Element.Multiline(true)` setter cannot bypass that early return. This example
keeps the actual disabled native Textarea and verifies that it rejects input;
it does not replace it with a fabricated accessibility control. Editable
examples explicitly assert native Multiline metadata. Real OS/assistive
technology/IME and complete font/platform acceptance remain separate.

`TestTextareaCompositionsSourceWorkflow` is one public-input flow through the
configuration selector, real multiline editing/selection/paste/undo, Field part
order and focus, disabled/invalid states, declared rows, local send, RTL/narrow
layout and independent drafts across source/style/page changes. Set
`GALLERY_TEXTAREA_COMPOSITION_SNAPSHOT_DIR` to export four representative
software PNGs. The injected selector Backend and software images are not real
operating-system window, screen-reader or input-method acceptance.
