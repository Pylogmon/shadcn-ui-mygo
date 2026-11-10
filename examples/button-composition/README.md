# Button source compositions

Open the Gallery Forms & controls page and select **Button configuration**. The
default is the docs Demo: an outline **Button** and icon-only **Submit**.

All 14 docs examples and 6 Create sections at shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7` appear as 20 configurations. Original
paths, Git blobs, SHA-256 values and the MDX guide are recorded in
[`source-map.json`](source-map.json). No source configuration is replaced by a
single generic approximation.

The docs provide Default/Outline/Secondary/Ghost/Destructive/Link variants,
four paired text/icon sizes, New Branch/Fork icons, circle-up icon, rounded
buttons, disabled Generating/Downloading spinners, a Login anchor and verbatim
Arabic/Hebrew/English content. The translated Submit arrow mirrors explicitly;
source Add keeps its accessible name. The Link visual variant remains a Button.

Create exposes Variants & Sizes, Icon Right, Icon Left, Icon Only, Invalid States
and Examples. Each of the first five configurations contains all four sizes and
six variants. Invalid buttons remain usable. Examples preserves Cancel/Submit,
Delete/icon action groups and the default anchor. Custom bodies retain source
inline-icon padding and the Link variant's hover underline. Two unusual source
markers are intentional: Small Outline in Icon Right has no `data-icon`, while
Downloading's spinner follows its text but still has `inline-start`.

Every source Button uses a real native MyGo target. The Login and Create Link
anchors use a native Link with its real RoleLink and Enter-only activation. A
public Router held by SDK `ui.Local` consumes their original `href="#"` as the
same local page, avoiding an external browser. Browser fragment/address-bar and
hash-scroll behavior are not claimed. A transparent, inert, AX-hidden visual
row paints above the real Link target; hover, pressed and focus appearance come
from styling that target, with no secondary ButtonBase or focus stop.

Actions record **Local button action**, and anchors record **Local button link**.
These are local example intents, without authentication, document submission or
filesystem effects. Each configuration retains its own last action/count when
changing configuration or leaving the page. Matrix accessible names include
size context; originally unnamed icon buttons receive descriptive names.

Icons use on-demand Material Symbols equivalents and the existing shared SVG
Spinner. No icon font, font family, bitmap or runtime download is introduced.
Menu style switches only the configuration/language pickers between Flyout and
system menus; the source button targets use native MyGo input with kit styling.

`TestButtonCompositionsSourceWorkflow` follows actual public pointer/keyboard
input through all source configurations, four sizes, marked/unmarked inline
spacing, invalid and disabled/busy controls, native anchors, Hebrew RTL,
configuration changes and leave/return. A high-contrast pixel witness verifies
one 50% disabled-opacity layer while preserving loading-only and inherited
disabled behavior; the old explicit-disabled SDK path produced 25%. The same
flow checks visible Link hover underline and that the Link label paints above
its target background. Set `GALLERY_BUTTON_COMPOSITION_SNAPSHOT_DIR` to export
four representative images. They are software Tester evidence, not a complete
native-window, screen-reader, IME, font or platform visual acceptance matrix.
