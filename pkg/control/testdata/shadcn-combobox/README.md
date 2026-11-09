# Fixed Combobox references

The primary reference is shadcn/ui commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`, `bases/base`, with the Vega
visual profile. `manifest.json` records exact Git blob IDs and SHA-256 values.
The archive includes the registry component, documentation, the full example
collection, all twelve separate Base examples and the shadcn license.

The Vega stylesheet and shared Tailwind animation rules are referenced through
their existing, byte-verified menu archive instead of copied again. The manifest
records those repository-relative archive paths and fixed upstream Git blobs.
It also references the unchanged `tw-animate-css` 1.4.0 archive and its verified
package metadata/license for the animation keyframes.

`npm/base-ui-react` contains unmodified published files from `@base-ui/react`
1.6.0. The entire cached official tarball was checked against its metadata SHA-1
and the SHA-512 integrity in the fixed shadcn lockfile before extracting selected
members. Package metadata, package.json and the MIT license are retained. No
JavaScript package is a dependency of the Go kit.

## Behavioral reference points

| Requirement | Reference |
| --- | --- |
| Single selection and input query are separate; input displays the selected item's label | `root/AriaCombobox.js`, `value/ComboboxValue.js` |
| Default auto-highlight is false; the explicit auto-highlight example enables it; list navigation loops by default | `root/AriaCombobox.js`, `shadcn/examples/combobox-auto-highlight.tsx` |
| On completed close, an external input restores the selected label and a popup input clears its query | `root/AriaCombobox.js` `handleUnmount` |
| Multiple selection toggles items; unfiltered selection keeps the list open; filtered external-input selection closes and clears the query | `root/AriaCombobox.js` `handleSelection` / `setOpen` |
| Clear empties query and selection, clears active indices and returns input focus; disabled/read-only prevents changes | `clear/ComboboxClear.js` |
| Default input includes a trigger; a visible clear button replaces that trigger | `shadcn/base/combobox.tsx` `ComboboxInput` |
| Items have a selected check indicator and disabled behavior; groups, labels and separators are separate parts | `shadcn/base/combobox.tsx`, `item/ComboboxItem.js` |
| Chips and the input share a wrapping field; visual remove buttons can be omitted | `shadcn/base/combobox.tsx`, `shadcn/examples/combobox-example.tsx`, existing Vega stylesheet |
| Vega content enters/exits over 100ms using CSS `ease`, fade and 95% scale; entry alone translates 8 DIP from the observed side | existing Vega stylesheet, Tailwind rules and `tw-animate-css` archive |
| A closing popup remains mounted but its positioner becomes inert; its query is cleaned after the exit completes | `popup/ComboboxPopup.js`, `positioner/ComboboxPositioner.js`, `root/AriaCombobox.js` |
| Transform origin projects the anchor center through the default 5px arrow padding and includes the side offset | `positioner/ComboboxPositioner.js`, `utils/useAnchorPositioning.js` |
| Popup input group has 4px outer margins (no bottom margin), 32px height and translucent input background/border | existing Vega stylesheet |
| From input caret start, the logical previous arrow enters the last chip; empty-input Backspace deletes the last selection | `input/ComboboxInput.js` |
| Chip arrows navigate logically in LTR/RTL; deletion selects the nearest remaining chip; Enter/Space returns to input; Up/Down opens the list | `chip/ComboboxChip.js` |
| Popup trigger with internal search, custom object labels/content, InputGroup addons, invalid/disabled and RTL examples | `shadcn/examples/combobox-{popup,custom,input-group,invalid,disabled,rtl}.tsx` |

The fixed shadcn documentation and examples do not define a loading/async data
transport component. Applications can replace their item collections; loading
status and request ownership need explicit native contracts if added. The archive
is source evidence, not a claim that every behavior or platform has passed native
acceptance.
