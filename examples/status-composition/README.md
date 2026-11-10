# Alert and Empty compositions

Run `go run ./cmd/gallery -page content`, then use **Alert configuration** and
**Empty configuration** on the Content & feedback page. Both sections keep their
own recipe, language and local preview state across configuration/page changes.

There are 20 Alert recipes and 13 Empty recipes. The map covers all six Alert
documentation examples, seven Empty documentation examples, and the fixed Create
collections' four Alert and six Empty sections. [source-map.json](source-map.json)
records the 15 exact source files, Git blobs, lengths, SHA-256 values and recipe
mapping at shadcn commit `8ab27715521cbd4f87ae6886cd989cbb678f6de7`.

- Alert recipes preserve title-only/description-only combinations, icon and long
  text cases, separate inline links, the destructive billing-information list,
  Undo and Badge actions, custom amber light/dark colors, and translated RTL.
- Empty recipes preserve project/no-result states, actual grayscale avatar and
  avatar-group media, notifications/Refresh, dashed cloud storage, solid/muted
  404 search fields, New Post, Card nesting, and Arabic/Hebrew translations.
- Search fields are real native editors inside InputGroup, with separate visual
  icons, slash Kbd and actual support links. The keycap alone installs no global
  shortcut. Links use native link semantics; source fragment placeholders are
  handled as local preview actions.
- The three original portraits are reused unchanged from the already committed
  [Card assets](../../internal/gallery/card-assets/README.md). Source URLs,
  redirects, original byte hashes and use sites are recorded in the map. Native
  grayscale is applied in memory; the application fetches no images at runtime.

The module uses the existing public `AlertWith`, `EmptyWith`, InputGroup, Avatar,
Card and native Link APIs. Slot callbacks receive their current child Context;
the application model retains only values and public handles, not frame builders,
Contexts or Elements. The global menu preference chooses styled independent
Flyout or system menus for recipe and language selectors.

Actions record local preview decisions or draft changes. They do not process a
payment, save an account remotely, send a message, upload a file or fetch real
notifications. Static source example copy is retained to show the composition.

The concentrated native workflow and exported PNGs cover the exercised source
configurations. Representative light/dark/RTL module images do not establish
every component state, real operating-system window behavior, actual input
methods or screen-reader acceptance.
