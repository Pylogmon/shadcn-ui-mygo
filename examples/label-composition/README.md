# Label source compositions

Gallery's Forms & controls page can expose **Label configuration** with both
Label docs previews, four Create sections, two distinct guide snippets and the
documentation's referenced complete Field demo: **nine independent recipes**.
All originals use shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. Seven original files, including
the component/related Field/Vega style references, are recorded with pinned
Git blobs, SHA-256 and byte lengths in [`source-map.json`](source-map.json).

| Source material | Retained content |
| --- | --- |
| Docs checkbox / RTL | Separate plain Label and actual unchecked native Checkbox, eight-DIP gap; original English/Arabic/Hebrew words and direction |
| Guide standalone | Your email address caption without inventing an editor absent from that snippet |
| Guide Label in Field | Actual FieldLabel and native email-text target, preserving their association |
| Create With Checkbox | Horizontal Field layout with separate plain Label and Checkbox |
| Create With Input | Username caption before the real editor and original placeholder |
| Create Disabled | Half-opacity caption and actual disabled editor, with the original Disabled placeholder |
| Create With Textarea | Message caption associated with the actual native multiline editor |
| Related Field demo | Complete Payment Method / Billing Address / Comments composition, original month/year collections and shipping default, native inputs/selects/checkbox/textarea and Submit / Cancel |

Plain captions use public `control.Label` / `LabelFor`; before-target captions
link the actual editor back with the public native `LabelledBy` relationship.
No fake focus or accessibility target replaces a label or editor. The related
Field example uses current-build FieldScope parts and independent strings;
required inputs retain their required metadata without adding visible source
asterisks. Source Fieldset/Legend/description, three-column month/year/CVV row,
28-DIP FieldGroup gaps and 448-DIP full-form maximum are preserved. Unspecified
Create widths use a 384-DIP native upper bound. Native font line boxes may be
taller than CSS's 14-DIP leading-none minimum.

The Label API now explicitly resets inherited line-height to source
leading-none and excludes captions from an ancestor Selectable container.
`LabelFor` dims a disabled peer once and marks its cursor; when the label
already inherits ordinary native disablement, it keeps that ancestor's dim
instead of multiplying it. Modern FieldLabel keeps its own independent
Field typography/disabled styling. A custom container that suppresses native
disabled dim remains responsible for its visuals; FieldLabel is the matching
part for a disabled Field. Actual native label clicking focuses editors or
toggles checkboxes, with disabled targets refusing input.

The referenced form's Submit captures only an application-memory snapshot
after its source-required values are nonempty. It does not invent a card-number
length/format validator from description text, add an Enter handler or perform
a transaction. Cancel records a local intent and leaves current drafts intact;
it is not a reset. Security/encryption wording is original sample copy, not an
implemented payment, encryption or storage service. Nullable source month/year
placeholders use empty native selection strings. Configuration/language and
form menus follow Flyout/Native choice. Models contain only data and counters,
never Context, Element, FieldScope or frame callbacks.

The existing `TestFieldSetAndStandaloneLabel` native workflow now also checks
line-height isolation, actual disabled-peer/ancestor pixels, real focus and
selection retention, one checkbox callback per activation, disabled refusal,
restoration and exclusion from ancestor text selection.
`TestLabelCompositionsSourceWorkflow` is one source workflow through all nine
recipes, caption/target order, real editing/selection, disabled and multiline
behavior, full form values, translated narrow layout and independent drafts.
Set `GALLERY_LABEL_COMPOSITION_SNAPSHOT_DIR` to export four representative
software PNGs. These checks do not establish real OS/AT/IME, every font/scale,
multi-display or full platform/visual acceptance. No font, asset or dependency
is added; SDK caches are unchanged.
