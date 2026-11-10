# Item source compositions

Run `go run ./cmd/gallery -page content`, then use **Item configuration** on the
Content & feedback page. The default **Basic Item** composition has a real,
independent Action button; the other recipes cover the source's documentation
examples and Create collections. Selection and local preview state survive page
and configuration changes.

The fixed source contains 11 documentation examples and 24 Create function
groups, exposed as 35 main configurations. [source-map.json](source-map.json)
records the 12 original files, exact Git blobs, lengths and SHA-256 values at
shadcn revision `8ab27715521cbd4f87ae6886cd989cbb678f6de7`. Each Create group
retains its title/media/description/action combinations; a single selected group
is shown rather than rendering every variant collection at once.

- Default, outline and muted surfaces cover default/small/extra-small sizes,
  title-only, description, media and multiple independent actions.
- Link items use actual native Link targets for their primary region. Header,
  actions and footer remain separate siblings and do not activate the primary
  target. Source fragment placeholders report local preview intents.
  Decorative trailing chevrons/external-link icons remain inside that primary
  target, so clicking the icon activates the same Link; real buttons stay separate.
- ItemGroup examples preserve native list positions, separators, header/footer
  text, grayscale images, model image headers and music artwork with an
  independent duration column.
- The member dropdown shows rich avatar/name/email Item rows in a styled
  independent Flyout. Native mode uses an operating-system text menu; it does
  not pretend the OS provides this custom rich row layout.
  Its original 192DIP width and p-2 content are retained. `ItemHeight: 60`
  reserves space for the 26DIP avatar and wrapped email; the accessible/native
  menu label contains both username and email. Ordinary styled menus retain 32DIP rows.
- Documentation recipes retain avatar groups, translated Arabic/Hebrew content,
  and source responsive visibility. Their actual bitmaps are supplied offline.

Nine original images are embedded from
[`internal/gallery/item-assets`](../../internal/gallery/item-assets/README.md).
The three original GitHub portraits are reused unchanged from the committed
[Card assets](../../internal/gallery/card-assets/README.md). Exact source URLs,
redirects, response types and byte hashes are in the map. Image transformations
are applied to decoded application bitmaps; no runtime network fetch or arbitrary
placeholder replaces the source imagery.

The module composes the public `display.Item`, ItemGroup/parts, Avatar,
Button/Link and desktop menu APIs. Every rich slot is built with its current
child Context. The application stores values and public menu state, not
Contexts, Elements or frame callbacks. Actions record local choices; the
example does not invite accounts remotely, play media, download user files or
claim successful backend changes.

The concentrated native workflow and representative screenshots document the
exercised behavior. Source counts and hashes are provenance bookkeeping; they
do not establish every platform, screen-reader, scale or visual state.
