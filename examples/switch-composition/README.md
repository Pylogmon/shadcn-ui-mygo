# Switch source compositions

Open **Forms & controls** in Gallery and choose **Switch configuration**. The
module retains all seven docs examples and all five sections of the Create
collection at shadcn/ui `8ab27715521cbd4f87ae6886cd989cbb678f6de7`, through twelve
configurations. [`source-map.json`](source-map.json) records the nine original
files, URLs, Git blobs, SHA-256 values and section mapping.

| Configuration | Source differences retained |
| --- | --- |
| Airplane Mode (docs) | Unchecked native switch with the original plain label and 8-DIP row gap |
| Description (docs) | Label and full description before the unchecked switch in a horizontal Field; the description itself does not activate it |
| Choice Cards | Two complete clickable FieldLabel cards; Share off, Enable notifications on; independent values |
| Disabled (docs) | One unchecked native disabled switch in a disabled horizontal Field |
| Invalid Terms | Original terms label and helper before an enabled, invalid switch |
| Sizes (docs) | Small/default tracks in horizontal Fields, 160-DIP maximum group width and 28-DIP group gap |
| Translated Switch | Verbatim English, Arabic and Hebrew text, Arabic default and logical RTL ordering |
| Create: Basic | Airplane Mode in a horizontal Field with its 12-DIP row gap |
| Create: Description Choice Card | One whole clickable Share card, distinct from the docs' plain description layout |
| Create: Bluetooth Label | Bluetooth with the original checked default and plain label |
| Create: Disabled States | Both Disabled (Unchecked) and Disabled (Checked), with a 48-DIP group gap |
| Create: Sizes | Small/default plain label rows with the source's 48-DIP group gap |

Each configuration owns its application booleans. Selection, menu style, language
and page changes retain those values; changing a switch records a local action.
The examples do not change device focus mode, Bluetooth, notifications or any
external settings.

The implementation uses `control.Switch`, `FieldWith` and the current
`FieldScope` label/content/control/description parts. Whole cards associate one
native `SwitchBase` with their enclosing label. Clicking the caption, card
description or actual track reaches that target once; there is no additional
button, manual duplicate toggle or second Tab stop. Plain descriptions remain
non-interactive. Plain Label rows use the source's leading-none and medium
font weight; Field labels retain their independent Field typography.
Disabled and invalid semantics belong to the actual native
switch, and invalid switches remain editable.

Small tracks are 24×14 DIPs and default tracks are 32×18.4, following the pinned
kit's Vega styling. The source's 384-DIP rich-layout bound, two distinct label
gaps and two distinct vertical group gaps are preserved. Native font line boxes
can differ from CSS. RTL content and switch motion use the current direction.

Configuration and language selectors follow the global **Flyout / Native**
menu backend. The switch itself is a themed native control in both routes.
Models contain only application values; no Context, Element, FieldScope or
frame callback survives a view build. No font, icon font, asset or dependency
is introduced.

`TestSwitchCompositionsSourceWorkflow` is one public-input flow through the
source selectors and actual pointer/Space input, thumb motion, label and
whole-card association, default/disabled/invalid state, source sizes and gaps,
Hebrew/narrow layout and state across route/hide changes. Set
`GALLERY_SWITCH_COMPOSITION_SNAPSHOT_DIR` to export four representative PNGs.
The injected selector Backend and software Tester images do not establish real
OS window placement/tracking, assistive technology, IME, font or full platform
visual acceptance.
