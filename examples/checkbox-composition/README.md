# Checkbox source compositions

The Gallery Forms & controls page exposes **Checkbox configuration**. Its 15
configurations preserve all eight docs examples and seven Create sections at
shadcn/ui revision `8ab27715521cbd4f87ae6886cd989cbb678f6de7`.
[`source-map.json`](source-map.json) records all ten original files, exact
Git blobs, bytes, SHA-256 values, URLs and native adaptations. The default is
the complete docs Demo.

| Configuration | Source content and state retained |
| --- | --- |
| Demo | Two independent Accept terms controls, the second checked with its shorter description; disabled notification; unchecked whole-card notification |
| Basic | An unchecked terms control in the original 224-DIP group |
| With Description | Checked terms with the full “terms and conditions” sentence in a 288-DIP group |
| Disabled | Disabled notification and its disabled field-label styling |
| Desktop Preferences (docs) | Label-sized legend, helper text, four normal-weight labels; Hard disks and External disks initially checked |
| Invalid | Invalid native terms checkbox, still interactive |
| In Table (docs) | Four original people/emails/roles, Sarah initially selected, independent row selection and boolean Select all |
| Translated Checkboxes | Verbatim English/Arabic/Hebrew labels and descriptions, Arabic default; the same distinct defaults as Demo |
| Create: Basic | Unchecked terms in the Create wrapper |
| Create: With Description | Checked terms with the original full description |
| Create: Invalid | Invalid terms, still interactive |
| Create: Disabled | Disabled notification without introducing the docs-only disabled field marker |
| Create: With Title | Checked notification card and a separate disabled unchecked card; title and description retained |
| Create: In Table | Its own original Sarah-only selection, separate from the docs table |
| Create: Desktop Preferences | All four unchecked, without the docs legend-description combination |

The module uses the public `control.Checkbox`, `FieldWith` and table composition
APIs. Every target is an actual native checkbox: clicking its associated label,
whole choice-card description or indicator updates the same bool once; Space
uses the focused target. Invalid does not mean disabled. Disabled fields and
cards cannot change their models. Native/Flyout affects the configuration and
language selectors, not the checkbox input contract.

Each recipe owns its draft and the table owns its actual four row values.
Source table `checked` is **all-selected boolean**, not indeterminate: partial
selection keeps Select all unchecked. Activating Select all selects all four
rows, or clears all four when already selected. A native horizontal viewport
keeps the original complete table reachable in a narrow window. It uses a
600-DIP preferred minimum as a native layout adaptation and adds explicit
nonvisual names to the otherwise visually unlabeled checkbox targets.

The fixed Vega styles use 28-DIP FieldGroup gaps, a 1-DIP top offset beside rich
content and 12-DIP inner card padding. Source disabled title cards dim the native
indicator and title while preserving their description and boundary opacity;
local fluent styles retain those differences and the native disabled behavior.
Docs' explicit 224/288/384-DIP bounds are retained. Create's unspecified widths
and the docs unbounded group are bounded to 384 DIP in this Gallery.

Changes publish `Local checkbox change: <source label>` and increment the
recipe's local count; no system preference or external record is modified.
Persistent state is application data and public `ScrollState`, never a Context,
Element or Field scope. The module adds no fonts, assets or dependencies.

`TestCheckboxCompositionsSourceWorkflow` is one public-input flow through all
source selectors, associated labels, Space, whole-card activation, disabled and
invalid states, independent source defaults, real table selection and narrow
horizontal input, Hebrew content, page hiding and the independent Flyout
selector. Set `GALLERY_CHECKBOX_COMPOSITION_SNAPSHOT_DIR` to export its four
representative software PNGs. Its injected backend and software images do not
establish real operating-system, screen-reader, IME, multi-display or
platform-font acceptance.
