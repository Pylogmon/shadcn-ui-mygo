# Radio Group source compositions

Open Gallery's Forms & controls page and use **Radio Group configuration**.
The default is the docs' **Spacing** group: Default, Comfortable and Compact,
with Comfortable initially selected. The module preserves seven docs examples
and six Create sections at shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7` through 13 configurations. The MDX
guide and eight example originals are pinned by URL, Git blob, byte count and
SHA-256 in [`source-map.json`](source-map.json).

| Configuration | Source content retained |
| --- | --- |
| Spacing | Default / Comfortable / Compact, medium labels, Comfortable default |
| Spacing descriptions | Original spacing descriptions and medium labels |
| Plan choice cards | Plus / Pro / Enterprise; complete punctuated descriptions; Plus default; indicator at logical end |
| Subscription plan fieldset | Label-size Subscription Plan legend, savings description and Monthly / Yearly / Lifetime prices; Monthly default |
| One disabled option | Only Disabled is disabled; Option 2 default; Option 3 remains usable |
| Invalid notification preferences | Label-size legend, description and Email only / SMS only / Both Email & SMS; Email default; all three actual inputs invalid |
| Translated spacing descriptions | Verbatim English / Arabic / Hebrew text and descriptions; Arabic default; logical direction and navigation |
| Create: Basic | Same spacing values with normal-weight Field labels |
| Create: With Descriptions | Native whole plan cards; source descriptions intentionally have no final period |
| Create: With FieldSet | Default-size Battery Level legend and description; High / Medium / Low, Medium default |
| Create: Grid Layout | Small / Medium / Large / X-Large in two columns; Medium default; whole card activation |
| Create: Disabled Group | All three Option 1 / Option 2 / Option 3 inputs disabled; Option 2 default |
| Create: Invalid | Default-size Notification Preferences legend, with source invalid options and Email default |

These are application selections. They update **Selected value** and an
independent per-configuration draft; plan labels do not charge, subscribe or
change an account. Drafts survive configuration, menu-style and page changes.
The existing Gallery Plan model is separate.

The implementation composes `control.RadioWith`, MyGo's public `RadioGroup`
and `control.FieldWith` / `FieldScope`. Radio inputs retain native checked,
disabled, invalid, focus, Space and arrow semantics. Rich Field labels make the
whole plan/grid card activate the same real Radio target; they introduce no
additional Button or Tab stop. Descriptions are associated with the native
target. Selection remains exclusive during the same input frame, before an
additional application frame can conceal a stale checked state.

Source explicit 320/384 DIP maximum widths are retained. Samples without an
explicit source width use a 384 DIP maximum native preview instead of CSS
intrinsic sizing, and text wraps with the application's native font metrics.
The grid keeps two columns. Radio content uses the current theme and existing
native controls in both global menu modes; **Native** selects system menus for
configuration selectors, while **Flyout** selects independent styled windows.
No new image, font, icon font or dependency is included.

`TestRadioCompositionsSourceWorkflow` is one public-input sequence through all
source configurations, pointer/Space/arrows, native exclusivity, disabled item
versus disabled group, Invalid, Field descriptions, whole cards, a two-column
grid, translated RTL and draft retention. Set
`GALLERY_RADIO_COMPOSITION_SNAPSHOT_DIR` to export four representative PNGs.
The test and injected popup backend provide native MyGo software-rendering
evidence, not real OS, screen-reader, IME or full font/platform/visual acceptance.

The persistent model holds application strings and maps only. Contexts,
Elements, Field scopes and their callbacks belong to the current build.
