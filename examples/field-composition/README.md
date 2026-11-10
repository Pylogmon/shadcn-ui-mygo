# Field compositions

Run `go run ./cmd/gallery -page controls`, then open **Field compositions**.
Search `FieldComposition`, `FieldSet` or `FieldGroup` in the Gallery command
palette to reach the Forms page. The example uses the public `control.FieldWith`
parts, `FieldSetWith`, and real native editors and choices.

**Field configuration** selects the thirteen fixed documentation examples and
the ten Create sections. **Field variant** selects each Create section's actual
configuration; the 64 configurations include label badges, either description
order, required/disabled/invalid fields, native grouped options, choice cards,
three slider thumbs, grouped OTP and horizontal layouts. The default responsive
Profile contains label and description together beside its editor at wide sizes,
and stacks them in a narrow FieldGroup. Each recipe retains its own application
drafts when switching configurations or Gallery pages.

The documentation recipes also supply payment FieldSets and a separator,
address fields, checkboxes and notification groups, a subscription radio group,
a two-option compute environment, and Arabic/Hebrew translated checkout fields.
The complete fixed source mapping and content hashes are in
[source-map.json](source-map.json). Styling follows the archived Vega rules.

Profile and payment Submit save only a local example snapshot; Cancel restores
that snapshot. No account operation, task request, payment, network submission or
browser constraint bubble is simulated. Use demonstration data in the payment
example. Manage tasks updates local example status. The explicit Native select
section always uses system menus; other selects follow the Gallery Flyout/Native
setting. Native inputs and choices retain their actual platform input semantics.

The focused application check is `TestFieldCompositionSourceWorkflow`. It edits
and submits the responsive profile, checks password helper order and native
relations, activates a whole compute card through its description, edits a
payment FieldSet, and exercises the rich label, grouped system menu, PIN, range
and horizontal switch configurations before checking page-return drafts. Set
`GALLERY_FIELD_COMPOSITION_SNAPSHOT_DIR` to export its representative native
images. This local composition workflow does not claim a complete Gallery route,
real OS input-method/reader acceptance or an all-state visual matrix.
