# Small Gallery composition recipes

Run `go run ./cmd/gallery -page controls` and scroll to **Choose a workspace
plan** and **Choose a size**. The entire Plus / Pro / Enterprise card selects a
native radio; the size group uses a two-column layout with ordinary native
radio navigation. Their application-owned choices survive page changes and do
not replace the existing Plan chooser.

Run `go run ./cmd/gallery -page display` and scroll to **Progress, with
context**, **Portrait color**, and **Loading forms and tables**. Upload progress
shows the source's 56% label/value; the independent transfer example starts at
50 and its slider changes the actual visible and accessible progress. The
portrait switch selects color or a real grayscale bitmap. Form/table
placeholders stay non-interactive and share the existing loading animation
pause preference.

These recipes use the already delivered public `display.Progress`,
`control.Slider`, `control.RadioWith`, native `ui.RadioGroup`/Grid,
`display.Avatar`, and `display.Skeleton` APIs. They add no component state machine.
The existing plain radios, Avatar image/fallback/loading/group workflows,
spinner controls and basic skeletons remain available.

The fixed source is shadcn commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`:

Exact original sizes, Git blobs and SHA-256 values are recorded in
[source-map.json](source-map.json).

| Recipe | Actual source |
| --- | --- |
| Label and value | `apps/v4/examples/base/progress-label.tsx:9`, value 56 |
| Shared slider value | `apps/v4/examples/base/progress-controlled.tsx:9`, initial 50, range 0–100, step 1 |
| Full radio cards | `apps/v4/examples/base/radio-group-choice-card.tsx:12`, initial Plus |
| Two-column radio grid | `apps/v4/registry/bases/base/examples/radio-group-example.tsx:131`, initial Medium |
| Grayscale image | `apps/v4/examples/base/avatar-basic.tsx:9`, filter on AvatarImage |
| Form skeleton | `apps/v4/examples/base/skeleton-form.tsx:5`, two label/field pairs and a button-shaped placeholder |
| Table skeleton | `apps/v4/examples/base/skeleton-table.tsx:5`, five rows of three placeholders |

The Skeleton form/table here follow these standalone docs examples. The larger
registry collection has separate size/row variations; this recipe does not
claim those are identical. A visual grid does not add a custom two-dimensional
keyboard model, and skeletons do not act as real inputs or tables.

The portrait reuses Gallery's local authored SVG instead of fetching a GitHub
photo. `Avatar` does not expose its internal image as a fluent slot, and
`Grayscale()` on its returned Box does not filter descendants. The application
uses public `SVG.Image` and `ui.NewBitmap` with ordinary grayscale image
preprocessing, preserving alpha, then supplies that decoded image through
`AvatarProps.Image`. This is an application recipe, not a new Avatar prop or a
subtree CSS filter. Native rendering and fonts can differ from browser CSS.
