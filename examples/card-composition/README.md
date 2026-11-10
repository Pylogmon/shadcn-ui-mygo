# Card compositions

Run `go run ./cmd/gallery -page overview`, then use **Card configuration** in
**Card compositions** on the Cards page. The original Gallery cards remain above
this module. Search `Card`, `CardSpacing` or `Meeting Notes` to reach this page.

The selector contains the fixed Create collection's 12 configurations and five
different documentation configurations: 17 native recipes in total. The six
documentation examples include the same Terms composition already present in
Create. [source-map.json](source-map.json) records all nine original source files,
their Git blob/SHA-256 values and the precise configuration mappings.

- Default/small cards, both sizes of header/footer dividers, and original
  16/20/24/32 spacing choices use the public `display.CardWith` API.
- Login recipes contain real, independently bound email/password editors,
  required validation, header/footer actions and Arabic/Hebrew translations.
  The password editor uses native secure-text mode; local status never contains
  its value. Each login example keeps its own draft through recipe/page changes.
- Terms retain the four source paragraphs in a bounded 192-DIP native viewport.
  Accept and Decline record a local decision. A zero-gap bleed column composes
  the source negative-margin footer without negative native layout geometry.
- Meeting Notes retains its ordered text, Transcribe action and actual three
  original portraits with a `+8` group count. Scheduled reports and Featured
  event retain their separate actions.
- Original images are embedded offline from the URLs used by the fixed source.
  The native image preprocessing applies grayscale, source light/dark brightness
  and the Create primary color overlay to the image pixels. Changing the palette
  updates the same landscape recipe. The source assets and attribution are listed
  in [the asset notice](../../internal/gallery/card-assets/README.md).

Actions update local preview status. Login, Google, sign-up, recovery, transcription
and scheduled reports do not contact a backend. The global menu style selects the
independent styled Flyout or system menu for recipe/language choices.

One native module workflow verifies real editing, secure semantics, local actions,
spacing geometry, bounded scrolling, image dimensions/palette updates, translated
RTL, draft retention and both selector routes. Set
`GALLERY_CARD_COMPOSITION_SNAPSHOT_DIR` while running
`TestCardCompositionsSourceWorkflowAndDrafts` to export representative module PNGs.
The test does not replace real platform/input-method/screen-reader acceptance or
the complete component/state visual matrix.
