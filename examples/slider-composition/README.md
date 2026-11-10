# Slider source compositions

Gallery's Forms & controls page offers **Slider configuration** with all seven
docs examples and six Create sections from shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The thirteen configurations retain
separate application drafts; [`source-map.json`](source-map.json) records the
nine original files, SHA-256 values, Git blobs, byte lengths and fixed URLs.

| Configuration | Source values and composition |
| --- | --- |
| Basic (docs) | One thumb at75, maximum100, step1 and320-DIP preview maximum |
| Range (docs) | Two independent endpoints25/50 and step5 |
| Multiple thumbs (docs) | Three native thumbs10/20/70 and step10 |
| Vertical pair (docs) | Independent50/25 values,160-DIP tracks and24-DIP gap, centered row |
| Controlled temperature (docs) | Visible Temperature label, joined0.3/0.7 readout,0–1 domain and0.1 step |
| Disabled (docs) | Actual disabled thumb at50; pointer and keyboard cannot change it |
| RTL slider | Value75, native reversed horizontal geometry and source Arabic/Hebrew/English directions |
| Create: Basic | Value50 rather than the docs'75; available native card width |
| Create: Range | Independent25/50 endpoints and step5 |
| Create: Multiple thumbs | Independent10/20/70 values and step10 |
| Create: Vertical pair | Independent50/25 values,160-DIP height and24-DIP gap |
| Create: Controlled temperature | Its own0.3/0.7 values, label, readout and fractional domain |
| Create: Disabled | Its own disabled value50 |

These are existing public `control.Slider`, `RangeSlider` and `SliderValues`
controls. The actual native focusable thumbs expose numeric ranges, steps,
orientation and disabled state; they handle keyboard and captured pointer
input. Range values remain ordered. The example does not add a submit, save or
network action where the source only edits local numbers.

Disabled styling applies one half-opacity layer to both track and thumb. The
fixed source Control uses `data-disabled:opacity-50`; Base UI1.6.0 renders the
visible Thumb as a div, so its `disabled:opacity-50` class does not add a second
layer. The real disabled native range input is hidden inside it. The native kit
keeps disabled input while avoiding additional default dimming on the root or
thumb. One existing Slider test now checks actual high-contrast pixels255→128
for explicit and inherited disabled states, along with rejected native input.
The source style and fixed Base UI behavior references are also recorded in the
source map.

The source's controlled readout joins decimal values with a comma and space.
Its display rounds native binary step noise to the declared tenth, without
modifying application numeric storage. Temperature retains its visible label
and group LabelledBy relation. The bare source examples supply no visible
labels; native-only accessible names identify Slider and the two vertical
sliders. The RTL source translations have no visible strings, so its language
selector changes direction without fabricating translations.

Docs explicitly use `max-w-xs`, represented by320 DIPs and shrinking with the
parent. Create supplies no explicit CSS width and fills its available native
card. Its vertical pair retains intrinsic width and is centered by the preview;
the docs row also centers the pair in its320-DIP maximum. Fixed height `h-40`,
gap6, controlled gap3 and label/readout gap2 map to160,24,12 and8 DIPs. Platform
font metrics and normal theme tokens remain native. No font, icon font, image
asset, dependency or runtime network loader is introduced.

Application state holds only values, strings, booleans and change counters.
Native Elements, Contexts and frame callbacks are constructed for each current
view and never retained. Selector menus follow the Gallery Flyout/Native
setting. Source drafts survive configuration/style changes and page hide/return.

`TestSliderCompositionsSourceWorkflow` is one real native-input flow covering
pointer and keyboard edits, ordered range limits, three-thumb identity,
disabled input, vertical height/gap and edits, controlled readout, RTL/narrow
geometry and independent drafts. Set
`GALLERY_SLIDER_COMPOSITION_SNAPSHOT_DIR` to export four representative software
PNGs. The injected selector Backend and these images do not establish actual
OS menu tracking, screen readers, IME, multi-display behavior, all fonts or the
complete platform/visual matrix.
