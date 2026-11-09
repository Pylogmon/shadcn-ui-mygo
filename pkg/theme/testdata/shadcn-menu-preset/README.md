# Fixed Create menu presets

These MIT-licensed shadcn files are copied from commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The manifest records each
verified Git blob and SHA-256, plus references to the existing menu/Combobox
archives and the fixed Tailwind theme. No source was obtained from Music.

`registry/config.ts:698–741` merges the theme and maps the theme-wide accent
pair to the primary pair for `menuAccent=bold`; SidebarAccent is unchanged.
`search-params.ts:151–207` and `menu-picker.tsx:79–108` normalize translucent
menus to subtle and restore the last solid accent when switching back.
The existing `transform-menu.ts` maps the four menuColor values, removes the
markers for default, forces a `dark` class for inverted, and inlines the
translucent utilities. The fixed Vega default is default/subtle. The source
has six effective color/accent combinations, not eight.

The source markers cover Combobox, Select, DropdownMenu, ContextMenu and
Menubar content and subcontent. They do not apply the translucent/no-animation
preset to Popover, Tooltip, HoverCard, Dialog or Command. The profile therefore
travels as an immutable theme snapshot independently of generic popover motion.
Native system menus and native popovers keep their platform renderer.

The native scene material samples its own framebuffer through public
`ui/shader`, with a 40DIP Gaussian, CSS saturate(1.5), then a 70% popover tint.
It does not blur text or sample another window. A separate Flyout requests
public `mygo.VibrancyMenu`, updates it through `Window.SetVibrancy`, and uses
`Context.Vibrancy` plus transparency/contrast preferences to select a system
material or an opaque fallback. The OS kernel is not claimed to equal CSS
40px/1.5, and a child framebuffer is not claimed to contain its parent or the
desktop. A live macOS probe observed the actual material capability and an
accelerated parent scene material; Windows/Linux runtime material behavior
remains platform-specific.

Standard preset scopes resolve their own corresponding dark palette. An
ordinary completely custom, unpaired Scope keeps its exact supplied colors
and inherits the dark flag, like fixed CSS variables; HasInvertedPalette is
false and no unrelated ancestor palette is borrowed. Explicit MenuScope or
Tokens.MenuTheme requires a recognizable preset or explicit InvertedTokens
when requesting inversion of a custom light palette. Scope retains its original
two-argument function type. All configurations and token pointers are copied.
