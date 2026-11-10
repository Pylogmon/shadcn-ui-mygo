# Item example image sources

These unmodified files reproduce public image URLs used by shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`, captured on 2026-11-10. Image
content is separate from this repository's MIT-licensed Go implementation. The
original owners retain their rights; this repository asserts no additional
image license. Full URLs, redirects, response types, lengths and SHA-256 values
are recorded in `examples/item-composition/source-map.json`.

| Local original | Source / attribution |
| --- | --- |
| `model-small.jpg` | Photo by **Valeria Reverdo on Unsplash**; source model `v0-1.5-sm`, endpoint `photo-1650804068570-7fb2e3dbf888`. |
| `model-large.jpg` | Photo by **Michael Oeser on Unsplash**; source model `v0-1.5-lg`, endpoint `photo-1610280777472-54133d004c8c`. |
| `model-mini.jpg` | Photo by **Cherry Laithang on Unsplash**; source model `v0-2.0-mini`, endpoint `photo-1602146057681-08560aee8cde`. |
| `song-midnight.png` | Generated artwork from `https://avatar.vercel.sh/Midnight%20City%20Lights`. |
| `song-coffee.png` | Generated artwork from `https://avatar.vercel.sh/Coffee%20Shop%20Conversations`. |
| `song-rain.png` | Generated artwork from `https://avatar.vercel.sh/Digital%20Rain`. |
| `project.png` | Generated source media from `https://avatar.vercel.sh/Project`. |
| `document.png` | Generated source media from `https://avatar.vercel.sh/Document`. |
| `file.png` | Generated source media from `https://avatar.vercel.sh/File`. |

The original `shadcn`, `maxleiter` and `evilrabbit` profile images are reused from
the neighboring `card-assets` directory, whose README records their provenance.
Their original byte hashes and Item use sites are repeated in the Item map.

The application decodes the originals offline and applies the source grayscale
transformation in memory. It makes no runtime image requests and does not imply
endorsement by the photographers, profile owners or source services.
