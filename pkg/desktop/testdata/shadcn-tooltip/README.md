# Fixed tooltip references

The UI wrapper and all eight example configurations are exact originals from
shadcn/ui commit `8ab27715521cbd4f87ae6886cd989cbb678f6de7`. `manifest.json`
records their original Git blobs, SHA256 hashes and runnable native mappings.
Vega CSS is reused byte-for-byte from the existing menu archive, and the shared
shadcn MIT license is included here and in the native bundles.

The native styled bubble uses 12-DIP horizontal/6-DIP vertical padding,
6-DIP spacing, a 6-DIP corner radius and 12-DIP text. Its directional arrow is
the source's 10-DIP square with 2-DIP rounded corners rotated by 45 degrees.
Native windows reserve a 7-DIP transparent gutter to keep those pixels inside
the window. Gap describes the distance to this outer native window; its bubble
therefore has the additional gutter inset. Alignment follows actual geometry,
including OS flips and shifts. Unknown/ambiguous geometry omits the arrow.

NativeTooltipWith retains platform styling without the styled arrow. Callers
use KbdWith's explicit InTooltip flag for inverse keycap colors; CSS descendant
selectors are not injected into arbitrary Go children. Rich content uses native
font metrics/layout and preserves a separate semantic description.
