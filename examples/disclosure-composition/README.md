# Disclosure source compositions

Run `go run ./cmd/gallery` and search for **Disclosure**, **Collapsible**, or
**Accordion**. The Gallery owner connects this module to its Data page. The
`Disclosure configuration` selector offers 17 source configurations while the
existing FAQ and More details examples remain available.

The fixed reference is shadcn/ui
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The exact 14 example originals,
Git blobs, byte hashes and recipe mapping are recorded in
[source-map.json](source-map.json): seven Accordion documentation examples, five
Collapsible documentation examples and the two Create collections. The Create
file tree and Radius settings reuse the corresponding documentation recipes;
the five Create Accordion configurations retain their distinct original text,
content and geometry.

The module uses the public native `CollapsibleRoot`/Scope parts and existing
`Accordion` API. The Radius card has two always-visible native inputs and an
independent icon trigger. Its extra two inputs are declared in the same grid
**before** the trigger and appear on the next row when opened. The inputs are
never children of a button. `KeepMounted` is explicit for this recipe, preserving
the actual native editor and selection after first opening; application-owned
strings retain the four drafts through configuration and page changes. Additional
editors have the distinct accessible names `Expanded Radius X`/`Expanded Radius Y`
instead of copying the original's repeated IDs and labels.

Order details keep the Status row visible while shipping/item details are closed.
The file explorer reproduces the original hierarchy, nested folder disclosures,
zero-padded content and 20 DIP logical indent. Leaf actions update a local selected
path; they do not open repository files. Explorer/Outline updates native Tab
selection, as the source has no distinct tab panels. Product details contain a
separate native Learn More button.

Accordion recipes preserve single/multiple selection, source default opening,
disabled headers, long question wrapping, joined/separate borders and the
Subscription & Billing cards. The Create Card's annual-billing paragraph is a real
native inline Link, and View plans is an independent Button. Hash-only source
links and demo actions update local status; there is no order, subscription or
account backend. `AccordionProps.StyleItem` styles the current native parts for
the documentation's joined outer border and the Create disabled-item padding and
open background, keeping one keyboard-navigation group.

Arabic/Hebrew/English selectors reproduce both translated source examples. Native
and styled recipe selectors follow the Gallery menu preference; the styled route
uses an independent Flyout. Material Symbols provide the decorative icons. All
content uses current native contexts; application state holds only strings,
booleans and public Handles.

`TestDisclosureCompositionsSourceWorkflowAndNativeDrafts` is one concentrated
native workflow covering the input grid, real selection retention, source Tab
order, persistent Order status, nested path selection, native links/actions,
Accordion selection/disabled navigation, translations, page return and both menu
routes. Set `GALLERY_DISCLOSURE_COMPOSITION_SNAPSHOT_DIR` to export its four
representative native images. These module images and public Tester operations
do not prove full real-platform, input-method or screen-reader acceptance.
