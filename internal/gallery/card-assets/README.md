# Card example image sources

These unmodified downloaded files reproduce the public image URLs in shadcn/ui
revision `8ab27715521cbd4f87ae6886cd989cbb678f6de7`, captured on 2026-11-10. Image
content is separate from this repository's MIT-licensed Go implementation; the
original owners retain their rights. No additional image license is asserted by
this repository. Exact URLs, redirects, response types, lengths and SHA-256 values
are recorded in `examples/card-composition/source-map.json`.

| Local original | Source / attribution |
| --- | --- |
| `landscape.jpg` | Photo by **mymind on Unsplash**, as attributed by the fixed Create example. Original image endpoint: `https://images.unsplash.com/photo-1604076850742-4c7221f3101b` with the complete source query recorded in the map. |
| `event.png` | Generated avatar/cover from `https://avatar.vercel.sh/shadcn1`, used by the documentation's Featured event example. |
| `shadcn.jpg` | `https://github.com/shadcn.png` → `https://avatars.githubusercontent.com/u/124599?v=4`; the public endpoint returned JPEG bytes. |
| `maxleiter.png` | `https://github.com/maxleiter.png` → `https://avatars.githubusercontent.com/u/8675906?v=4`. |
| `evilrabbit.png` | `https://github.com/evilrabbit.png` → `https://avatars.githubusercontent.com/u/6880091?v=4`. |

The application embeds the originals and performs its source grayscale,
brightness and color-overlay transformations in memory. It does not make runtime
image requests, and it does not substitute generated portraits or landscape
placeholders. Avatar labels preserve the source handles for attribution and
native image semantics. The examples do not imply endorsement by these people.
